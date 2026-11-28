package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/data/effect"
	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
)

// Recipes follow 0x577390 and the target attachments in 0x57bdc0/0x57e320.
// Melee impacts follow the first ALT contact. Area announcements create one
// centre effect even if the server reports no individual recipients.
func skillEffectLibraries(r proto.SkillResult, d sdata.Skill) (flight, impact string) {
	if r.Level == 0 || r.Level == 100 || d.Effect == 0 || r.Target == 0 && d.Target != 6 {
		return
	}
	if r.Result == 4 && (d.Target == 6 || d.Target == 7) {
		if d.Target == 7 {
			if file := effect.AreaLibrary(r.ID); file != "" {
				if areaShowerSkill(r.ID) {
					return "", file
				}
				return "", ""
			}
		}
		return "", effect.ImpactLibrary(r.ID)
	}
	if d.Attack == 3 && (d.Effect == 1 || d.Effect == 3 || d.Effect == 5) {
		if r.Result == 0 || r.Result == 1 || r.Result == 3 {
			flight = effect.ProjectileLibrary(r.ID)
			if flight != "" {
				return flight, effect.ImpactLibrary(r.ID)
			}
		}
	}
	// A resisted or missed spell (result 3) still plays its effect.
	if r.Result <= 3 && d.Target >= 1 && d.Target <= 7 {
		impact = effect.ImpactLibrary(r.ID)
	}
	return
}

func (a *App) startSkillResultEffects(r proto.SkillResult) {
	if a.evolutionSkillEffects(r, true) {
		return
	}
	a.loadSkills()
	d, ok := a.skillDefs[sdata.SkillKey{ID: r.ID, Level: r.Level}]
	if !ok {
		return
	}
	if d.IsSoulBurn() {
		if r.Result == 0 {
			a.startSoulBurnImpact(r)
		}
		return
	}
	switch r.ID { // 0x577020: Earthquake shakes the camera for two seconds, Earth Spikes for half
	case 69:
		a.cameraShake(2)
	case 83:
		a.cameraShake(.5)
	}
	a.startSkillReleaseEffect(r, d)
	if a.wv.physicalSkillQueued {
		return
	}
	flight, impact := skillEffectLibraries(r, d)
	sound := skillResultEffectSound(r, d)
	if d.Target == 6 && (r.Target == 0 || r.Result == 4) {
		r.Target = r.Source
	}
	if flight == "" {
		a.playEntitySound(r.Target, sound)
	}
	if r.Result == 4 && areaShowerSkill(r.ID) && d.Target == 7 && impact != "" {
		a.startSkillAreaShower(r, d, impact)
		return
	}
	if flight != "" {
		a.startSkillProjectile(r, flight, impact, sound)
	} else if impact != "" {
		elapsed := float32(0)
		if actor := a.playerActor(r.Source); actor != nil && actor.ActionActive() {
			elapsed = actor.Time
		}
		if d.Attack == 1 && (d.Effect == 1 || d.Effect == 3 || d.Effect == 5) {
			actor := a.playerActor(r.Source)
			slot := scene.SkillAction(d.Animation, d.Attack, d.Effect, a.playerWeapon(r.Source), a.attackVariant)
			if actor != nil && actor.PlayingAction(slot) && actor.ContactTime() > 0 {
				if actor.Time >= actor.ContactTime() {
					a.startSkillImpactAt(r.Target, impact, int(r.Level)-1, actor.Time-actor.ContactTime())
					return
				}
				if len(a.wv.impactCues) < 32 {
					a.wv.impactCues = append(a.wv.impactCues, skillImpactCue{actionClock{actor, actor.ActionID(), actor.ContactTime()}, r.Source, r.Target, impact, int(r.Level) - 1})
				}
				return
			}
		}
		a.startSkillImpactAt(r.Target, impact, int(r.Level)-1, elapsed)
	}
}

// Target-centered area announcements use mode 0. A caster-centered result
// uses the target selector on the caster. Secondary recipient FX have no
// spell cue in the original 0x5783d0/0x579390 handlers.
func skillResultEffectSound(r proto.SkillResult, d sdata.Skill) string {
	if r.Level == 0 || r.Level == 100 || d.Effect == 0 || d.Target < 1 || d.Target > 7 {
		return ""
	}
	if r.Result <= 3 || r.Result == 4 && d.Target == 6 {
		return effect.TargetSound(r.ID)
	}
	if r.Result == 4 && d.Target == 7 {
		return effect.AreaSound(r.ID)
	}
	return ""
}

// Soul Burn's result arrives while the corpse is still in the view. Freeze
// the impact at its feet so it stays there through the subsequent rebirth.
func (a *App) startSoulBurnImpact(r proto.SkillResult) {
	if r.Target == 0 || len(a.wv.impacts) >= 32 {
		return
	}
	pos, ok := a.effectPosition(r.Target)
	if !ok {
		return
	}
	file := effect.ImpactLibrary(r.ID)
	x := a.loadSkillEffect(file, int(r.Level)-1)
	if x == nil {
		return
	}
	m := render.Identity()
	m[12], m[13], m[14] = pos[0], pos[1], pos[2]
	x.SetTransform(m)
	a.wv.impacts = append(a.wv.impacts, skillImpact{instance: x, fixedWorld: &m, authoredLifetime: true})
}

// Area, party and additional hits have already announced the cast; render an
// impact for each successful target without restarting the caster's animation.
func (a *App) startSkillSecondaryEffect(r proto.SkillResult) {
	if a.evolutionSkillEffects(r, false) {
		return
	}
	if a.wv != nil && a.wv.physicalSkillQueued {
		return
	}
	if a.wv == nil || r.Level == 0 || r.Level == 100 || r.Target == 0 || r.Result > 3 {
		return
	}
	a.loadSkills()
	d, ok := a.skillDefs[sdata.SkillKey{ID: r.ID, Level: r.Level}]
	if !ok || d.Effect == 0 {
		return
	}
	if file, sequence := skillSecondaryLibrary(r, d); file != "" {
		a.startSkillImpact(r.Target, file, sequence)
	}
}

// 0x579390/0x5783d0 select the recipient effect independently of the
// central announcement. A DA_area_* filename does not suppress that hit.
func skillSecondaryLibrary(r proto.SkillResult, d sdata.Skill) (string, int) {
	if r.Level == 0 || r.Level == 100 || r.Target == 0 || r.Result > 3 || d.Effect == 0 {
		return "", 0
	}
	if d.Target == 7 && !areaShowerSkill(r.ID) {
		if file := effect.AreaLibrary(r.ID); file != "" {
			return file, int(r.Level) - 1
		}
	}
	if d.Effect == 2 || d.Effect == 3 || d.Effect == 4 {
		return effect.ImpactLibrary(r.ID), int(r.Level) - 1
	}
	// Native 0x577b40 maps AttribType to the five Adamage sequences.
	sequence := effect.AttributeDamageSequence(d.Attribute)
	return "Effect/DA_target_Adamage.EFT", sequence
}

type skillImpactCue struct {
	actionClock
	source, target uint32
	file           string
	sequence       int
}

func (a *App) updateSkillImpactCues() {
	kept := a.wv.impactCues[:0]
	for _, cue := range a.wv.impactCues {
		due, alive := cue.ready(a.playerActor(cue.source))
		if !alive {
			continue
		}
		if due {
			a.startSkillImpactAt(cue.target, cue.file, cue.sequence, max(0, cue.actor.Time-cue.time))
		} else {
			kept = append(kept, cue)
		}
	}
	clear(a.wv.impactCues[len(kept):])
	a.wv.impactCues = kept
}

func (a *App) startSkillImpact(target uint32, file string, sequence int) {
	a.startSkillImpactAt(target, file, sequence, 0)
}

func (a *App) startSkillImpactAt(target uint32, file string, sequence int, elapsed float32) {
	if len(a.wv.impacts) >= 32 {
		return
	}
	if _, ok := a.effectPosition(target); !ok {
		return
	}
	if x := a.loadSkillEffect(file, sequence); x != nil {
		// Position before catching up; otherwise births use the world origin.
		if m, ok := a.effectTargetMatrix(target); ok {
			x.SetTransform(m)
		}
		elapsed = min(max(0, elapsed), 2)
		advanceSkillEffect(x, elapsed)
		a.wv.impacts = append(a.wv.impacts, skillImpact{target: target, instance: x, age: elapsed})
	}
}
