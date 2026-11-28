package scene

import (
	"math"
	"slices"

	"github.com/Spelunkern/Shaiya-GX/internal/render"
)

const MinCrowdActors = 24
const maxCrowdPaletteMatrices = 1 << 20 // 64 MiB, below the OpenGL 4.5 SSBO limit.

type rigidMaterial struct {
	mesh    *render.Mesh
	texture *render.Texture
	ref     float32
}

type rigidGroup struct {
	material rigidMaterial
	worlds   []render.M4
}

type crowdActorDraw struct {
	instance                  render.SkinInstance
	serial, visible, occluded bool
}

type crowdWeaponEffect struct {
	actor      *Character
	attachment *attachment
	world      render.M4
}

type crowdPaletteKey struct {
	bind  *render.M4
	clip  *Clip
	time  float32
	count int
	used  [render.MaxBones / 64]uint64
}

type CrowdDrawStats struct {
	Skin         int `json:"skin"`
	SkinGroups   int `json:"skin_groups"`
	PaletteBytes int `json:"skin_palette_bytes"`
	Rigid        int `json:"weapons_shields"`
	Cloth        int `json:"cloth"`
	Fallback     int `json:"serial"`
	ClothBytes   int `json:"cloth_position_bytes"`
	ClothUploads int `json:"cloth_position_uploads"`
	WeaponFX     int `json:"weapon_effects"`
}

// CharacterBatch gathers already visible crowd actors for one camera pass.
type CharacterBatch struct {
	keys        map[render.SkinGroup]int
	prepared    []crowdActorDraw
	groups      []render.SkinGroup
	buckets     [][]render.SkinInstance
	instances   []render.SkinInstance
	palettes    render.CrowdPalettes
	partOffsets map[*skinPart][2]int32
	sharedPoses map[crowdPaletteKey]int32
	frameActive bool
	offsets     [][4]int32
	rigidKeys   map[rigidMaterial]int
	rigidGroups []rigidGroup
	cloaks      []*Character
	cloth       render.ClothCrowd
	skin        render.SkinCrowd
	weaponFX    []crowdWeaponEffect
	Stats       CrowdDrawStats
}

func (b *CharacterBatch) Draw(k *render.Skinner, s *render.Scene3D, actors []*Character) {
	b.Stats = CrowdDrawStats{}
	if len(actors) < MinCrowdActors || s.Shadow || !k.CrowdInstancing() {
		before := render.DrawCalls()
		for _, c := range actors {
			c.Draw(k, s)
		}
		b.Stats.Fallback = render.DrawCalls() - before
		return
	}
	if b.keys == nil {
		b.keys = make(map[render.SkinGroup]int)
		b.rigidKeys = make(map[rigidMaterial]int)
		b.partOffsets = make(map[*skinPart][2]int32)
		b.sharedPoses = make(map[crowdPaletteKey]int32)
	}
	if !b.frameActive {
		b.resetPalettes()
	}
	b.cloth.Begin(len(actors))
	b.prepared = slices.Grow(b.prepared[:0], len(actors))[:len(actors)]
	occlusion := s.ActiveOcclusion()
	forEachCrowdCharacter(actors, func(i int, c *Character) {
		draw := &b.prepared[i]
		*draw = crowdActorDraw{serial: c.Concealed || c.hasBlendedAttachment()}
		if draw.serial {
			return
		}
		_, draw.visible = c.prepareDrawPose(false)
		if !draw.visible {
			return
		}
		if occlusion != nil {
			if point, valid := c.occlusionWitness(); !valid || occlusion.PointMayBeHidden(point) {
				lo, hi, valid := c.drawOcclusionBounds()
				draw.occluded = valid && occlusion.BoxHidden(lo, hi)
			}
		}
		if !draw.occluded {
			draw.instance = s.CrowdSkinInstance(c.World)
		}
	})
	for i, c := range actors {
		if c == nil {
			continue
		}
		draw := &b.prepared[i]
		if draw.serial {
			b.flush(k, s)
			before := render.DrawCalls()
			c.Draw(k, s)
			b.Stats.Fallback += render.DrawCalls() - before
			continue
		}
		if draw.occluded {
			occlusion.Culled++
		} else if draw.visible {
			if len(b.palettes.Matrices)+len(c.parts)*render.MaxBones > maxCrowdPaletteMatrices {
				b.flush(k, s)
				b.resetPalettes()
			}
			b.add(s, c, draw.instance, c.lastFinals)
		}
	}
	b.flush(k, s)
	b.Stats.ClothBytes, b.Stats.ClothUploads = b.cloth.Uploaded, b.cloth.UploadOps
}

func (b *CharacterBatch) Delete() {
	b.cloth.Delete()
	b.skin.Delete()
	b.palettes.Delete()
}

// BeginFrame shares only bone matrices across views. Group membership,
// camera transforms and culling are still evaluated separately for each view.
func (b *CharacterBatch) BeginFrame() {
	b.resetPalettes()
	b.frameActive = true
}

func (b *CharacterBatch) EndFrame() {
	b.frameActive = false
	clear(b.partOffsets)
	clear(b.sharedPoses)
}

func (b *CharacterBatch) resetPalettes() {
	b.palettes.Reset()
	clear(b.partOffsets)
	clear(b.sharedPoses)
}

func (b *CharacterBatch) paletteOffset(c *Character, part *skinPart, count int, used []int) [4]int32 {
	if count == 0 {
		return [4]int32{}
	}
	if offset, ok := b.partOffsets[part]; ok {
		return [4]int32{offset[0], offset[1]}
	}
	matrices := part.pose[:count]
	offset := int32(len(b.palettes.Matrices))
	if c.blendLeft <= 0 && !c.oneShot {
		// Cloned parts share immutable inverse binds. Equal phase and
		// weighted-bone masks produce identical matrices at every used index.
		key := crowdPaletteKey{bind: &part.bones[0], clip: c.clip, time: c.Time, count: count}
		for _, bone := range used {
			key.used[bone/64] |= uint64(1) << (bone % 64)
		}
		if previous, ok := b.sharedPoses[key]; ok {
			b.partOffsets[part] = [2]int32{previous, int32(count)}
			return [4]int32{previous, int32(count)}
		} else {
			b.sharedPoses[key] = offset
		}
	}
	b.palettes.Matrices = append(b.palettes.Matrices, matrices...)
	b.partOffsets[part] = [2]int32{offset, int32(count)}
	return [4]int32{offset, int32(count)}
}

func (c *Character) hasBlendedAttachment() bool {
	if !c.Mounted {
		for _, a := range c.attachments {
			if a.part.Alpha {
				return true
			}
		}
	}
	return false
}

func (b *CharacterBatch) add(s *render.Scene3D, c *Character, instance render.SkinInstance, finals []render.M4) {
	b.offsets = b.offsets[:0]
	for i, part := range c.parts {
		source := c.poseSources[i]
		if source == i {
			count := min(len(part.pose), render.MaxBones)
			used := c.poseBones[source]
			if len(used) > 0 {
				count = used[len(used)-1] + 1
			}
			b.offsets = append(b.offsets, b.paletteOffset(c, part, count, used))
		} else {
			b.offsets = append(b.offsets, b.offsets[source])
		}
		if part.mesh.Count == 0 {
			continue
		}
		ref := float32(-1)
		if part.cut {
			ref = .5
		}
		group := render.SkinGroup{Mesh: part.mesh, Texture: part.tex, AlphaRef: ref, Unlit: part.unlit, TwoSided: part.twoSided}
		index, ok := b.keys[group]
		if !ok {
			index = len(b.groups)
			b.keys[group] = index
			b.groups = append(b.groups, group)
			if index == len(b.buckets) {
				b.buckets = append(b.buckets, nil)
			}
		}
		instance.Palette = b.offsets[i]
		instance.Palette[2] = int32(math.Float32bits(c.CorpseFade))
		b.buckets[index] = append(b.buckets[index], instance)
	}
	if c.cloak != nil {
		b.cloaks = append(b.cloaks, c)
	}
	if c.Mounted {
		return
	}
	for i := range c.attachments {
		a := &c.attachments[i]
		if a.bone < 0 || a.bone >= len(finals) || i == 0 && c.hiddenWeaponAction != 0 && c.hiddenWeaponAction == c.ActionID() {
			continue
		}
		world := a.extra.Mul(finals[a.bone]).Mul(c.World)
		if !s.HideWeaponEffects && (a.enchant != nil || a.glow != nil) {
			b.weaponFX = append(b.weaponFX, crowdWeaponEffect{c, a, world})
		}
		if !s.RigidInstanceVisible(a.part.Mesh, world) {
			continue
		}
		material := rigidMaterial{mesh: a.part.Mesh, texture: a.part.Tex, ref: -1}
		if a.cut {
			material.ref = .5
		}
		index, ok := b.rigidKeys[material]
		if !ok {
			index = len(b.rigidKeys)
			b.rigidKeys[material] = index
			if index == len(b.rigidGroups) {
				b.rigidGroups = append(b.rigidGroups, rigidGroup{})
			}
			b.rigidGroups[index].material = material
		}
		group := &b.rigidGroups[index]
		group.worlds = append(group.worlds, world)
	}
}

func (b *CharacterBatch) flush(k *render.Skinner, s *render.Scene3D) {
	view, occlusion := s.View, s.ActiveOcclusion()
	parallelEffects(len(b.weaponFX), func(i int) {
		job := b.weaponFX[i]
		if enchant := job.attachment.enchant; enchant != nil {
			job.actor.prepareWeaponEnchant(view, enchant, job.world, occlusion)
		}
		if glow := job.attachment.glow; glow != nil {
			job.actor.updateWeaponGlow(job.attachment, job.world)
			glow.prepareDraw(view, occlusion)
		}
	})
	for i := range b.groups {
		instances := b.buckets[i]
		b.groups[i].First, b.groups[i].Count = len(b.instances), len(instances)
		b.instances = append(b.instances, instances...)
		b.buckets[i] = instances[:0]
	}
	ambient := s.CharacterAmbient
	if ambient <= 0 {
		ambient = .8
	}
	before := render.DrawCalls()
	indirect, written := k.DrawCrowdIndirect(s, &b.skin, b.groups, b.instances, &b.palettes, ambient)
	if !indirect {
		written = k.DrawCrowd(s, b.groups, b.instances, &b.palettes, ambient)
	}
	b.Stats.Skin += render.DrawCalls() - before
	b.Stats.SkinGroups += len(b.groups)
	b.Stats.PaletteBytes += written
	before = render.DrawCalls()
	if len(b.cloaks) != 0 || len(b.rigidKeys) != 0 {
		s.Begin()
	}
	for _, c := range b.cloaks {
		cloak := c.cloak
		if !s.ModelBoundsOccluded(cloak.min, cloak.max, c.World) {
			if !b.cloth.Queue(s, cloak.geometry.mesh, cloak.geometry.texture, cloak.id, cloak.version, cloak.vertices, c.World) {
				cloak.drawInPass(s, c.World, false)
			}
		}
	}
	b.cloth.Draw(s)
	b.Stats.Cloth += render.DrawCalls() - before
	before = render.DrawCalls()
	if len(b.rigidKeys) != 0 {
		for i := range len(b.rigidKeys) {
			group := &b.rigidGroups[i]
			m := group.material
			s.SetInstances(group.worlds)
			s.DrawEquipmentInstanced(m.mesh, m.texture, m.ref)
			group.material = rigidMaterial{}
			group.worlds = group.worlds[:0]
		}
	}
	b.Stats.Rigid += render.DrawCalls() - before
	if len(b.weaponFX) > 0 {
		// Opaque depths are complete. Keep the original actor/attachment/layer
		// order among effects, with depth writes disabled throughout their pass.
		before = render.DrawCalls()
		s.BeginEffects()
		for _, job := range b.weaponFX {
			if enchant := job.attachment.enchant; enchant != nil {
				job.actor.submitWeaponEnchant(s, enchant)
			}
			if glow := job.attachment.glow; glow != nil {
				glow.submitDraw(s)
			}
		}
		s.EndEffects()
		b.Stats.WeaponFX += render.DrawCalls() - before
		clear(b.weaponFX)
		b.weaponFX = b.weaponFX[:0]
	}
	clear(b.keys)
	clear(b.rigidKeys)
	clear(b.groups)
	clear(b.cloaks)
	b.groups, b.instances, b.cloaks = b.groups[:0], b.instances[:0], b.cloaks[:0]
}
