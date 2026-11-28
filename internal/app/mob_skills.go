package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/data/effect"
	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
)

// Native 0x57add0/0x579b80/0x57a2c0 look up level one and use impact
// sequence two for monster recipients. Wire level 100 is a marker, not rank 100.
func (a *App) startMobSkillEffects(r proto.SkillResult, startsCast bool) {
	if a.wv == nil || r.Result > 4 {
		return
	}
	if startsCast {
		if mob := a.trackMob(a.state.Entities[r.Source]); mob != nil && !mob.entity.Dead {
			a.loadMobActor(mob)
			if mob.actor != nil {
				a.playMobAttack(mob, r.Target, mob.actor.AttackAction(r.AttackSlot))
			}
		}
	}
	d, ok := a.skillDefinition(r.ID, 100)
	if !ok || d.Effect == 0 || r.Result > 1 {
		return
	}
	target := r.Target
	if d.Effect == 2 || d.Effect == 4 {
		target = r.Source
	}
	impact, sound := mobSkillImpact(d, startsCast)
	a.playEntitySound(target, sound)
	if impact != "" {
		a.startSkillImpact(target, impact, 2)
	}
	if d.Effect != 2 && d.Effect != 3 && d.Effect != 4 {
		visual := r
		visual.Level = 1
		visual.ID = d.VisualCode
		if file, sequence := skillSecondaryLibrary(visual, d); file != "" && file != impact {
			a.startSkillImpact(target, file, sequence)
		}
	}
}

// 0x060b/0x060c play the target selector. The 0x060d recipient handler
// (0x57a2c0) tries the area selector only when the target recipe is absent
// and the effect is not a debuff. Weather showers take their separate route.
func mobSkillImpact(d sdata.Skill, startsCast bool) (library, sound string) {
	id := d.VisualCode
	library = effect.ImpactLibrary(id)
	if library != "" || startsCast || d.Effect == 3 {
		return library, effect.TargetSound(id)
	}
	switch id {
	case 53, 68, 71:
		return "", ""
	}
	return effect.AreaLibrary(id), effect.AreaSound(id)
}
