package app

import (
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
	"github.com/Spelunkern/Shaiya-GX/internal/session"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

func TestMountedSkillInputsLeaveRiderAndTargetAlone(t *testing.T) {
	for name, mountState := range map[string]uint8{"driver": proto.VehicleDriver, "passenger": proto.VehiclePassenger} {
		t.Run(name, func(t *testing.T) {
			a := inputApp()
			a.state.SelfID, a.state.Target = 1, 1
			a.state.Entities[2] = &game.Entity{ID: 2, Kind: game.KindMob, X: 20}
			a.clock = 10
			a.wv.self = &scene.Character{}
			k := proto.LearnedSkill{ID: 77, Level: 1, Number: 1}
			a.state.Skills[k.Number] = k
			a.skillDefs[sdata.SkillKey{ID: k.ID, Level: k.Level}] = sdata.Skill{Effect: 1, Target: 3, Show: 1}
			a.state.QuickSlots = []proto.QuickSlot{{Bag: 100, Number: uint16(k.Number)}}
			a.revolver.profile = a.localActionsKey()
			a.revolver.Slots[0][0] = revolverSlot{Skill: k.ID}
			a.applyWorldPacket(proto.VehicleStatePacket(proto.VehicleState{ID: 1, State: mountState, Speed: 4, Seats: 2, Driver: 3}))
			if !a.state.Self.Mounted {
				t.Fatal("mount fixture was not accepted")
			}
			// No live socket. Attempting to send a skill would drop this session.
			world := &session.World{}
			a.world = world
			a.wv.autoRun = true
			a.useAction(0)
			a.useRevolver()
			a.skills.cooldowns = map[uint16]float64{k.ID: 20}
			a.useRevolver()
			delete(a.skills.cooldowns, k.ID)

			a.skills.open = true
			const w, h = 1024, 768
			x, y := a.skillBook.Origin(w, h)
			cx, cy, _ := ui.SkillBookRow(0)
			a.updateSkills(&platform.Input{MouseX: x + cx + 1, MouseY: y + cy + 1, Pressed: true, MouseDown: true}, w, h)
			a.updateSkills(&platform.Input{MouseX: x + cx + 1, MouseY: y + cy + 1, Clicked: true}, w, h)
			for _, user := range []bool{false, true} {
				packet := proto.UseSkillPacket(k.Number, 2, user)
				if a.dispatchShortcut(packet) {
					t.Fatal("mounted rider dispatched a skill packet")
				}
				a.requestedSkillPresentation(packet)
			}
			if a.world != world || a.state.Target != 1 || !a.wv.autoRun || a.wv.approachTarget != 0 || a.wv.attackMoveUntil != 0 || a.skills.queuedSkill != nil || a.attackVariant != 0 || a.revolver.selected != 0 {
				t.Fatal("blocked spell changed targeting, movement, animation or Revolver state")
			}
			a.applyWorldPacket(proto.VehicleStatePacket(proto.VehicleState{ID: 1, State: proto.VehicleOff}))
			a.state.Target = 2
			if got := proto.Op(a.actionPacket(0)); got != proto.SkillUse {
				t.Fatalf("action bar skill did not return after dismount: %#x", got)
			}
		})
	}
}

func TestMountConfirmationDropsPendingSkillAndIgnoresLateAnimation(t *testing.T) {
	a := inputApp()
	a.state.SelfID = 1
	a.wv.self = &scene.Character{}
	key := sdata.SkillKey{ID: 99, Level: 1}
	packet := proto.UseSkillPacket(1, 2, false)
	a.wv.approachTarget, a.wv.approachSkill, a.wv.travel = 2, packet, true
	a.skills.queuedSkill = &queuedSkill{packet: packet}
	a.revolver.cast = revolverCast{skill: key}
	a.wv.requestedSkill = &requestedSkillPresentation{key: key, actor: a.wv.self}
	a.wv.requestedSkills = []*requestedSkillPresentation{a.wv.requestedSkill}
	a.wv.attackMoveUntil = 11
	a.applyWorldPacket(proto.VehicleStatePacket(proto.VehicleState{ID: 1, State: proto.VehiclePassenger, Speed: 4, Seats: 2, Driver: 3}))
	if a.skills.queuedSkill != nil || a.revolver.cast.skill.ID != 0 || a.wv.approachTarget != 0 || len(a.wv.approachSkill) != 0 || a.wv.travel || a.wv.attackMoveUntil != 0 || a.wv.requestedSkill != nil || len(a.wv.requestedSkills) != 0 {
		t.Fatal("joining a mount left a pending skill or movement lock")
	}
	cast := proto.NewWriter(proto.SkillCastStarted)
	cast.U32(1)
	cast.U32(2)
	cast.U16(key.ID)
	cast.U8(key.Level)
	a.applyWorldPacket(cast.Bytes())
	if a.state.LastCast.ID != key.ID {
		t.Fatal("cast-start fixture did not reach client state")
	}
	a.state.LastSkill = proto.SkillResult{Source: 1, Target: 2, ID: key.ID, Level: key.Level}
	a.confirmedAnimation(proto.NewWriter(proto.SkillUse).Bytes())
	if a.wv.self.ActionActive() || a.wv.attackMoveUntil != 0 {
		t.Fatal("late skill confirmation affected the rider's presentation")
	}
}

func TestVehicleCanonicalCancellationClearsPendingRequest(t *testing.T) {
	a := inputApp()
	a.state.SelfID = 42
	a.vehicle.pending = true
	a.confirmedVehicle(proto.VehicleStatePacket(proto.VehicleState{ID: 42, State: proto.VehicleOff}))
	if a.vehicle.pending {
		t.Fatal("asynchronous cancellation left mount request pending")
	}
}

func TestVehicleCooldownUsesSecondsIndependentlyOfPotions(t *testing.T) {
	a := inputApp()
	a.clock = 10
	a.inventory.cooldowns = map[uint8]float64{10: 610}
	it := proto.InventoryItem{Type: 42, TypeID: 5}
	if a.itemCooldown(it) != 0 {
		t.Fatal("potion cooldown blocked mount")
	}
	a.vehicle.cooldownEnd = 20
	if a.itemCooldown(it) != 10 || itemCooldownDuration(&sdata.Item{Type: 42, ReqIg: 10}) != 10 {
		t.Fatal("mount ReqIg was treated as potion group")
	}
}

func TestMountAppearanceSoundUsesNativeRaceAndBoardRules(t *testing.T) {
	want := []string{"run_horse-roar01.wav", "run_pumav-roar01.wav", "run_pumav-roar01.wav", "run_baslk-roar01.wav"}
	for race, name := range want {
		if got := mountAppearanceSound(uint8(race), 1); got != name {
			t.Fatal(race, got)
		}
		if mountAppearanceSound(uint8(race), 6) != "" {
			t.Fatal("board played animal roar")
		}
	}
}
