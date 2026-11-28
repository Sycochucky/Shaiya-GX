package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/data/effect"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
)

// placeField shows a field left on the ground (Fire Bush's curtain of fire):
// its area library, as authored, once, at the spot (DA_area_fire's sequence
// lasts about eleven seconds, the spell's own time). Nothing is repeated or
// changed.
func (a *App) placeField(f proto.Field) {
	if a.wv == nil {
		return
	}
	a.playSound(effect.AreaSound(f.Skill), f.X, f.Z)
	if len(a.wv.impacts) >= 32 {
		return
	}
	file := effect.AreaLibrary(f.Skill)
	if file == "" {
		return
	}
	x := a.loadSkillEffect(file, int(f.Level)-1)
	if x == nil {
		return
	}
	m := render.Identity()
	m[12], m[13], m[14] = f.X, f.Y, f.Z
	if a.wv.world != nil && a.wv.world.HasTerrain() {
		m[13] = a.wv.world.Height(f.X, f.Z)
	}
	x.SetTransform(m)
	a.wv.impacts = append(a.wv.impacts, skillImpact{instance: x, authoredLifetime: true, fixedWorld: &m})
}
