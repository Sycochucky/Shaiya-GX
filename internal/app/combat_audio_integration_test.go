//go:build integration

package app

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/audio"
	"github.com/Spelunkern/Shaiya-GX/internal/data/archive"
	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
	"github.com/Spelunkern/Shaiya-GX/internal/testworld"
)

func TestCombatCuesAgainstRapture(t *testing.T) {
	if os.Getenv("SHAIYA_TEST_COMBAT") != "1" || os.Getenv("SHAIYA_AUDIO_TEST") != "1" {
		t.Skip("requires disposable combat fixture, GL and audio device (muted)")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	win, err := platform.Open(platform.Config{Title: "weapon audio probe", Width: 1024, Height: 734, Hidden: true, Debug: true})
	if err != nil {
		t.Fatal(err)
	}
	defer win.Close()
	if err := render.Init(true); err != nil {
		t.Fatal(err)
	}
	arc, err := archive.Open(os.Getenv("SHAIYA_DATA"))
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()
	st, world, send, wait := testworld.Enter(t)
	mob := testworld.LivingMob(t, st, send, wait)
	a, err := New(Config{}, arc)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	a.sound, err = audio.New(arc.ReadFile, 0)
	if err != nil {
		t.Fatal(err)
	}
	a.soundErrors = map[string]bool{}
	a.state, a.world, a.screen = st, world, screenWorld
	a.enterWorld()
	if a.wv.self == nil {
		t.Fatal("fixture player model did not load")
	}
	st.Target = mob.ID
	if !a.attackTarget() {
		t.Fatal("could not start approaching the target")
	}
	const frame = time.Second / 60
	deadline := time.Now().Add(12 * time.Second)
	input := &platform.Input{}
	for len(a.wv.attackSounds) < 2 {
		a.pumpWorld()
		if a.state == nil || time.Now().After(deadline) {
			t.Fatal("confirmed weapon hit did not queue its swing/contact", st.LastAttack)
		}
		if len(a.wv.attackSounds) >= 2 {
			break
		}
		a.clock += frame.Seconds()
		a.updateWorld(input, frame.Seconds())
		a.updateAudio()
		time.Sleep(frame)
	}
	v := a.wv
	position, action, recovery := v.pos, v.self.ActionID(), v.attackMoveUntil
	a.updateWorld(&platform.Input{Down: "S"}, 1.0/60)
	if v.attackSent || v.engagedID != 0 || v.approachTarget != 0 {
		t.Fatal("backpedal did not stop the ongoing attack order")
	}
	if !v.self.ActionActive() || v.self.ActionID() != action || v.attackMoveUntil != recovery ||
		v.pos[0] != position[0] || v.pos[2] != position[2] {
		t.Fatal("backpedal cut the server-confirmed swing or bypassed its recovery")
	}
	result, pools := st.LastAttack, st.Pools
	if result.Source != st.SelfID || result.Target != mob.ID || result.HP == 0 || result.Result > 1 {
		t.Fatal("audio was not tied to confirmed positive damage", result)
	}
	swing, hit := a.wv.attackSounds[0], a.wv.attackSounds[1]
	if swing.entity != st.SelfID || hit.entity != mob.ID || swing.time <= 0 || hit.time <= 0 || swing.actor != a.wv.self || hit.actor != a.wv.self {
		t.Fatal("weapon audio lost its actor/contact anchor", swing, hit)
	}
	for i := 0; i < 120 && len(a.wv.attackSounds) > 0; i++ {
		a.wv.self.Update(1.0 / 60)
		a.updateAudio()
	}
	if len(a.wv.attackSounds) != 0 || st.Pools != pools {
		t.Fatal("weapon cues repeated or changed gameplay pools")
	}
	t.Logf("Rapture hit queued %s at %.3fs and %s at ALT contact %.3fs; both drained muted, pools unchanged", swing.file, swing.time, hit.file, hit.time)
	checkPhysicalSkillContact(t, a, send, wait)
	send(proto.LogoutPacket())
	wait(func() bool { return st.Phase == game.PhaseCharSelect })
}

func TestShadowAttackWeaponContactGL(t *testing.T) {
	if os.Getenv("SHAIYA_TEST_RENDER") != "1" {
		t.Skip("requires GL and original assets")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	win, err := platform.Open(platform.Config{Width: 64, Height: 64, Hidden: true})
	if err != nil {
		t.Fatal(err)
	}
	defer win.Close()
	if err := render.Init(false); err != nil {
		t.Fatal(err)
	}
	arc, err := archive.OpenDirectory(os.Getenv("SHAIYA_DATA"))
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()
	a := inputApp()
	a.arc, a.assets = arc, scene.NewAssets(arc)
	a.skillDefs = nil
	a.loadSkills()
	a.state.SelfID = 1
	const weapon = 10
	a.state.Inventory[[2]uint8{0, 5}] = proto.InventoryItem{Type: weapon}
	actor, err := scene.BuildCharacter(a.assets, scene.Appearance{Race: 1, Job: 2, Height: 2, Helmet: -1, WeaponType: weapon}, "000_normal")
	if err != nil {
		t.Fatal(err)
	}
	a.wv.self = actor
	victim := &game.Entity{ID: 2, Kind: game.KindMob, TypeID: 2009, X: 1}
	a.state.Entities[victim.ID] = victim
	mob := a.trackMob(victim)
	a.loadMobActor(mob)
	if mob.actor == nil {
		t.Fatal("could not load the target's native animations")
	}
	const skill = 309
	def := a.skillDefs[sdata.SkillKey{ID: skill, Level: 1}]
	if !actor.PlaySkillAction(skill, scene.SkillAction(def.Animation, def.Attack, def.Effect, weapon, 0)) {
		t.Fatal("could not play Shadow Attack")
	}
	first, second := actor.ContactTimes()
	if first <= 0 {
		t.Fatal("Shadow Attack has no native contact marker in the fixture")
	}
	result := proto.SkillResult{ID: skill, Level: 1, Source: 1, Target: victim.ID, HP: 11}
	miss := result
	miss.Result, miss.HP = 3, 0
	if a.deliverMeleeSkillHit(miss, def, false) || mob.incomingHits != 0 {
		t.Fatal("a miss queued a weapon impact")
	}
	a.state.LastSkill = result
	a.confirmedSkillFeedback(proto.SkillUse)
	if mob.incomingHits != 1 || len(a.combatNumbers) != 0 {
		t.Fatal("the confirmed hit was presented before contact")
	}
	mob.actor.Time = .1
	actor.Time = first
	a.updatePendingNumbers()
	if mob.incomingHits != 0 || mob.actor.Time >= .1 || len(a.combatNumbers) != 1 {
		t.Fatal("first contact did not present the hit and monster reaction")
	}
	if second > first {
		actor.Time = second
		a.updatePendingNumbers()
		if len(a.combatNumbers) != 2 {
			t.Fatal("second contact did not present the remaining damage")
		}
	}

	if !actor.PlaySkillAction(skill, scene.SkillAction(def.Animation, def.Attack, def.Effect, weapon, 0)) {
		t.Fatal("could not replay Shadow Attack")
	}
	a.confirmedSkillFeedback(proto.SkillUse)
	victim.Dead = true
	a.updatePendingNumbers()
	if mob.deathFrozen {
		t.Fatal("the monster died visually before the lethal strike")
	}
	actor.Time = first
	a.updatePendingNumbers()
	if !mob.deathFrozen || mob.incomingHits != 0 {
		t.Fatal("the first contact failed to release the monster's death")
	}
	t.Logf("Shadow Attack native contact %.3fs, second %.3fs", first, second)
}
