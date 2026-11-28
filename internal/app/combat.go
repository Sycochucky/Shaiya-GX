package app

import (
	"fmt"
	"log"
	"math"

	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

// Rapture's 0x0205 failure payload contains an explicit reason. Successful
// item and money packets retain their normal inventory handling.
func (a *App) confirmedPickup(p []byte) {
	if proto.Op(p) != proto.AddItem || len(p) != 4 || p[2] != 0 {
		return
	}
	message := a.tr(12238)
	switch p[3] {
	case 0:
		message = a.tr(12239)
	case 1:
		message = a.tr(12240)
	case 2:
		message = a.tr(12241)
	case 3:
		message = a.tr(12242)
	case 4:
		message = a.tr(12243)
	}
	ids := map[byte]int{0: 202, 1: 238, 3: 239, 4: 233}
	if id := ids[p[3]]; id != 0 && a.msg.Get(id) != "" {
		a.systemMessage(1, id, nil)
	} else {
		a.questMessage(message)
	}
}

// Select on the first click and engage on a subsequent click on the same target.
func (a *App) updateCombat(in *platform.Input, w, h int) {
	if a.wv == nil {
		return
	}
	if a.state.Self.Dead {
		for _, r := range in.Chars {
			if (r == 'r' || r == 'R') && a.clock >= a.wv.reviveAt {
				if err := a.world.Send(proto.RevivePacket()); err != nil {
					a.backToLoginID(12030)
					return
				}
				a.wv.reviveAt = a.clock + 3
			}
		}
		return
	}
	for key := platform.KeyParty1; key <= platform.KeyParty8; key++ {
		if in.HasKey(key) {
			id := a.partyShortcutTarget(int(key - platform.KeyParty1))
			if id == a.state.SelfID && id != 0 {
				a.selectSelf()
			} else if id != 0 {
				a.selectGroupTarget(id)
			}
			return
		}
	}
	if in.HasKey(platform.KeyAssist) {
		a.requestPartyAssist()
		return
	}
	id := a.pickGroundModel(in, w, h)
	if id != 0 {
		a.approachLoot(id)
		in.Clicked = false // This click issued the approach; do not cancel it in movement.
		return
	}
	if in.Clicked || in.HasKey(platform.KeyTarget) || in.HasKey(platform.KeyPreviousTarget) {
		selected := a.state.Target
		target := a.pickMob(in, w, h)
		if in.Clicked && target == 0 {
			if npc := a.pickNPC(in, w, h); npc != 0 {
				if selected == npc {
					a.openNPC(npc)
				} else {
					a.selectNPC(npc)
				}
				in.Clicked = false
				return
			}
			a.wv.travel = false
			a.moveToGround(in, w, h)
			return
		}
		if in.HasKey(platform.KeyPreviousTarget) {
			target = a.state.PreviousTarget
			if e := a.state.Entities[target]; target != a.state.SelfID && (e == nil || e.Dead && e.Kind != game.KindUser) {
				return
			}
		}
		if target != a.state.Target && a.wv.engagedID != target {
			// Selecting something else leaves the previous attack action.
			if !a.abandonAttack() {
				return
			}
		}
		if e := a.state.Entities[target]; e != nil && e.Kind == game.KindUser && target != a.state.Target {
			if old := a.state.Entities[a.state.Target]; old != nil && old.Kind == game.KindMob {
				if err := a.world.Send(proto.StopAttackPacket()); err != nil {
					a.backToLoginID(12031)
					return
				}
			}
		}
		if e := a.state.Entities[target]; target != a.state.Target && e != nil && e.Kind == game.KindMob {
			if err := a.world.Send(proto.TargetMobPacket(target)); err != nil {
				a.backToLoginID(12003)
				return
			}
		}
		if target != a.state.Target {
			a.state.PreviousTarget = a.state.Target
		}
		a.state.Target = target
		if in.Clicked {
			a.wv.travel = false
		}
		// Native 0x43bfa0: a click selects; a click on the already selected
		// target engages it. There is no time window between the two clicks.
		if in.Clicked && target != 0 && target == selected && target != a.state.SelfID {
			in.Clicked = false
			if !a.attackTarget() {
				return
			}
		}
	}
}

func (a *App) pickupNearest() bool {
	id := uint32(0)
	// Native 0x43bfa0 nearest-item action searches within 30 units.
	best := float32(30 * 30)
	for k, it := range a.state.GroundItems {
		dx, dz := it.X-a.wv.pos[0], it.Z-a.wv.pos[2]
		d := dx*dx + dz*dz
		if d < best || d == best && (id == 0 || k < id) {
			id, best = k, d
		}
	}
	if id != 0 {
		return a.approachLoot(id)
	}
	return true
}

func (a *App) attackTarget() bool {
	if a.state.Self.Dead {
		a.wv.sitAttack = 0
		return true
	}
	if a.state.Target == a.state.SelfID && a.state.SelfID != 0 {
		// With yourself selected the attack goes for the nearest enemy, like an
		// offensive skill does.
		if id := a.nearestHostileMob(); id != 0 {
			if a.world != nil {
				if err := a.world.Send(proto.TargetMobPacket(id)); err != nil {
					a.backToLoginID(12003)
					return false
				}
			}
			a.state.PreviousTarget = a.state.SelfID
			a.state.Target = id
		}
	}
	e := a.state.Entities[a.state.Target]
	if e == nil || e.Dead || e.Kind != game.KindMob && (e.Kind != game.KindUser || e.ID == a.state.SelfID) {
		return true
	}
	if a.friendlyMob(e) {
		a.wv.approachTarget = 0
		a.wv.approachSkill = nil
		a.systemMessage(17, 317, nil)
		return true
	}
	follow := a.followOnly(e)
	if !follow && a.deferAirAction(e.ID, nil) {
		return true
	}
	if a.state.Self.Sitting || a.wv.sitPending || a.wv.self != nil && a.wv.self.PlayingAction(11) {
		a.wv.sitAttack = e.ID
		if a.state.Self.Sitting && !a.wv.sitPending {
			return a.requestPosture(proto.MotionStandUp)
		}
		return true
	}
	a.wv.approachTarget, a.wv.approachSkill = 0, nil
	a.wv.approachNPC = false
	a.wv.travel, a.wv.autoRun = false, false
	if follow {
		return a.followPlayer(e)
	}
	if !a.faceRequestedTarget(e.ID) {
		return false
	}
	goal := render.V3{e.X, e.Y, e.Z}
	if e.Kind == game.KindMob {
		goal = a.visualEntityPosition(e)
	}
	delta := goal.Sub(a.wv.pos)
	delta[1] = 0
	if delta.Len() > a.attackReachTo(e) || a.clock < a.wv.attackMoveUntil {
		a.wv.autoRun = false
		a.wv.approachTarget, a.wv.approachLoot = e.ID, false
		a.wv.approachSkill = nil
		return a.updateActionApproach(&platform.Input{})
	}
	if !a.stopForAttack() {
		return false
	}
	// Rapture decides whether a player is hostile (faction, PvP map, duel,
	// guild duel). A client-side duel-only check would block normal PvP.
	if e != nil && e.Kind == game.KindUser && e.ID != a.state.SelfID && !e.Dead {
		if err := a.world.Send(proto.UserAttackPacket(e.ID)); err != nil {
			a.backToLoginID(12032)
			return false
		}
		a.keepChasing(e.ID)
		return true
	}
	if e != nil && e.Kind == game.KindMob && !e.Dead {
		if err := a.world.Send(proto.MobAttackPacket(e.ID)); err != nil {
			a.backToLoginID(12033)
			return false
		}
		a.keepChasing(e.ID)
	}
	return true
}

// keepChasing leaves the approach tracking a target after its attack request:
// in the original the attack state persists, following the target and swinging
// again whenever it leaves and re-enters reach.
func (a *App) keepChasing(id uint32) {
	v := a.wv
	v.approachTarget, v.approachLoot, v.approachNPC = id, false, false
	v.approachSkill = nil
	v.engagedID, v.engagedAt = id, a.clock
	v.attackSent = true
}

// abandonAttack stops repeat attacks and pursuit. A confirmed swing still
// plays through the server's movement recovery.
func (a *App) abandonAttack() bool {
	v := a.wv
	if !v.attackSent && v.engagedID == 0 {
		return true
	}
	if v.attackSent && a.world != nil {
		if err := a.world.Send(proto.StopAttackPacket()); err != nil {
			a.backToLoginID(12031)
			return false
		}
	}
	v.attackSent, v.engagedID = false, 0
	if v.approachTarget != 0 && !v.approachLoot && !v.approachNPC {
		v.approachTarget, v.approachSkill = 0, nil
	}
	return true
}

func (a *App) drawCombat(w, h int) {
	a.drawCombatNumbers(w, h)
	s := a.state
	if a.clock < a.combatNoticeUntil {
		a.font.Draw(a.sprites, a.combatNotice, w/2-45, h/2-80, white, 0)
	}
	view := ui.HUDState{ViewportW: w, ViewportH: h,
		Portrait: a.hudPortrait(w, h),
		HP:       ui.Resource{Current: s.Pools.HP, Max: s.Details.MaxHP},
		MP:       ui.Resource{Current: s.Pools.MP, Max: s.Details.MaxMP},
		SP:       ui.Resource{Current: s.Pools.SP, Max: s.Details.MaxSP},
		Exp:      s.Details.Exp, PrevExp: s.Details.PrevExp, NextExp: s.Details.NextExp,
	}
	if s.Blessing.Known && s.Blessing.Country < 2 && s.Blessing.Maximum > 0 {
		view.BlessCount, view.BlessMax, view.BlessKnown = s.Blessing.Count, s.Blessing.Maximum, true
		view.BlessRemainingMS, view.BlessCountry, view.Clock = s.Blessing.RemainingMS, s.Blessing.Country, a.clock
	}
	for _, slot := range s.Slots {
		if slot.ID == s.SelfID {
			view.Name, view.Level = slot.Name, slot.Level
			break
		}
	}
	if s.Target == s.SelfID && s.SelfID != 0 {
		view.TargetName, view.TargetHP, view.TargetKnown = view.Name, view.HP, true
	} else if e := s.Entities[s.Target]; e != nil && e.Kind == game.KindMob {
		view.TargetName = a.msg.Format(12297, map[string]string{"v": fmt.Sprint(e.TypeID)})
		if def, err := a.assets.MonsterDefinition(e.TypeID); err == nil {
			view.TargetName, view.TargetHP.Max = def.Name, def.HP
			view.TargetColor = ui.RGBColor(ui.MobNameRGB(def.Level, view.Level))
			view.TargetElement = def.Element
		} else if !a.assets.Missing[err.Error()] {
			log.Printf("target definition: %v", err)
			a.assets.Missing[err.Error()] = true
		}
		view.TargetHP.Current = uint32(max(e.HP, 0))
		view.TargetKnown = e.HealthKnown
	} else if e != nil && e.Kind == game.KindUser && e.Shape != nil {
		view.TargetName = e.Shape.Name
		if look := s.PlayerForms[e.ID]; look.Kind == proto.FormDisguise {
			view.TargetName = look.Shape.Name
		}
		if member := s.Party.Member(e.ID); member != nil {
			view.TargetHP = ui.Resource{Current: member.HP, Max: member.MaxHP}
			view.TargetKnown = true
		}
		if h, ok := s.UserHealth[e.ID]; ok && !view.TargetKnown {
			view.TargetHP = ui.Resource{Current: h.HP, Max: h.Max}
			view.TargetKnown = true
		}
	}
	a.hud.Draw(a.sprites, a.font, view)
	a.drawTargetEffects(w, h)
	a.drawCastAndOxygen(w, h)
	a.bottom.QuestOpen, a.bottom.GuildOpen = a.questLog.open, a.guild.open
	a.bottom.Draw(a.sprites, a.font, a.white, w, h, view, a.inventory.open, a.skills.open, a.status.open, a.subActionsOpen)
	a.drawBlessingTooltip(w, h)
	a.drawBlessingBanner(w, h)
}

// selectNPC makes an NPC the current target. Its target frame is never drawn:
// the original has no target UI for NPCs.
func (a *App) selectNPC(id uint32) {
	if a.state.Target != id {
		if old := a.state.Entities[a.state.Target]; old != nil && old.Kind == game.KindMob && a.world != nil {
			if err := a.world.Send(proto.StopAttackPacket()); err != nil {
				a.backToLoginID(12031)
				return
			}
		}
		a.state.PreviousTarget = a.state.Target
	}
	a.state.Target = id
	if a.wv != nil {
		a.wv.approachSkill = nil
	}
}

func (a *App) selectSelf() {
	if a.state.Target != a.state.SelfID {
		if old := a.state.Entities[a.state.Target]; old != nil && old.Kind == game.KindMob && a.world != nil {
			if err := a.world.Send(proto.StopAttackPacket()); err != nil {
				a.backToLoginID(12031)
				return
			}
		}
		a.state.PreviousTarget = a.state.Target
	}
	a.state.Target = a.state.SelfID
	if a.wv != nil {
		a.wv.approachTarget = 0
		a.wv.approachSkill = nil
	}
}

// castBarOrigin follows 0x004c02d0 and the native FISTP rounding of screen pixels.
func castBarOrigin(pos render.V3, view, proj render.M4, w, h int) (int, int, bool) {
	// The native actor origin is .9 above GX feet. Native camera-right
	// comes from the horizontal camera direction, independent of pitch.
	right := (render.V3{view[0], 0, view[8]}).Norm()
	p := pos.Add(render.V3{0, .9, 0}).Add(right.Scale(.5))
	vp := view.Mul(proj)
	cw := p[0]*vp[3] + p[1]*vp[7] + p[2]*vp[11] + vp[15]
	z := p[0]*vp[2] + p[1]*vp[6] + p[2]*vp[10] + vp[14]
	if cw <= 0 || z < 0 || z > cw {
		return 0, 0, false
	}
	x := (1 + (p[0]*vp[0]+p[1]*vp[4]+p[2]*vp[8]+vp[12])/cw) * float32(w) * .5
	y := (1 - (p[0]*vp[1]+p[1]*vp[5]+p[2]*vp[9]+vp[13])/cw) * float32(h) * .5
	return int(math.RoundToEven(float64(x))), int(math.RoundToEven(float64(y))), true
}

// nearestHostileMob is the closest living enemy monster (the lowest id on a tie).
func (a *App) nearestHostileMob() uint32 {
	if a.wv == nil {
		return 0
	}
	id, best := uint32(0), float32(math.MaxFloat32)
	for k, e := range a.state.Entities {
		if e.Kind != game.KindMob || e.Dead || a.friendlyMob(e) {
			continue
		}
		pos := a.visualEntityPosition(e)
		dx, dz := pos[0]-a.wv.pos[0], pos[2]-a.wv.pos[2]
		if d := dx*dx + dz*dz; d < best || d == best && k < id {
			id, best = k, d
		}
	}
	return id
}

// followOnly is the native team test of 0x43bfa0: a player of your own
// faction is followed, never attacked, unless they are your duel or rival
// guild-duel opponent. Players without shape data keep the server's verdict.
func (a *App) followOnly(e *game.Entity) bool {
	if e == nil || e.Kind != game.KindUser || e.ID == a.state.SelfID || e.Shape == nil {
		return false
	}
	if e.Shape.Faction != a.state.Faction.Faction {
		return false
	}
	if d := a.state.Duel; d.Phase == game.DuelFighting && d.Opponent == e.ID {
		return false
	}
	if g := a.state.GuildDuel; g.Active && e.Shape.Guild != a.state.Guild.Name {
		return false
	}
	return true
}

// followPlayer walks to 1.5 units from a friendly player and keeps following
// them. Purely client side: no attack, target or heading packet is sent.
func (a *App) followPlayer(e *game.Entity) bool {
	v := a.wv
	v.approachTarget, v.approachLoot, v.approachNPC, v.approachSkill = e.ID, false, false, nil
	delta := render.V3{e.X - v.pos[0], 0, e.Z - v.pos[2]}
	if delta.Len() <= a.attackReachTo(e) {
		v.engagedID, v.engagedAt = e.ID, a.clock
		return true
	}
	v.engagedID = 0
	return a.updateActionApproach(&platform.Input{})
}

// targetFrameShown is whether the target frame is on screen: the original draws
// it for monsters, players and yourself, never for an NPC.
func (a *App) targetFrameShown() bool {
	if a.state == nil || a.state.Target == 0 {
		return false
	}
	if a.state.Target == a.state.SelfID {
		return true
	}
	e := a.state.Entities[a.state.Target]
	return e != nil && e.Kind != game.KindNpc
}
