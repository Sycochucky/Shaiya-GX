package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
	"math"
	"strings"
)

type vehicleView struct {
	effectBudget              int
	inviteSerial, replySerial uint64
	expires, cooldownEnd      float64
	castState                 uint8
	peerName                  string
	states                    map[uint32]proto.VehicleState
	pending                   bool
	deadline                  float64
}

func (a *App) requestVehicle() {
	if a.state == nil || a.world == nil || a.state.Self.Dead || a.vehicle.pending || a.inventory.pending || a.shop.pending {
		return
	}
	if !a.state.Self.Mounted && a.state.Self.Vehicle.State != proto.VehicleCasting {
		if a.clock < a.vehicle.cooldownEnd {
			return
		}
		item, ok := a.state.Inventory[[2]uint8{0, 13}]
		if !ok || item.Type != 42 {
			a.questMessage(a.chatEntityName(a.state.SelfID) + " " + a.msg.Get(5245))
			return
		}
	}
	if err := a.world.Send(proto.UseVehiclePacket()); err != nil {
		a.backToLoginID(12202)
		return
	}
	a.vehicle.pending, a.vehicle.deadline = true, a.clock+10
}
func (a *App) vehicleCommand(text string) bool {
	if !strings.EqualFold(strings.TrimSpace(text), "/mount") {
		return false
	}
	a.requestVehicle()
	return true
}
func (a *App) confirmedVehicle(p []byte) {
	if proto.Op(p) == proto.UserLeave {
		if len(p) >= 6 {
			delete(a.vehicle.states, proto.NewReader(p).U32())
		}
		return
	}
	if proto.Op(p) == proto.GXVehicleState {
		v, err := proto.DecodeVehicleState(p)
		if err != nil {
			return
		}
		if a.vehicle.states == nil {
			a.vehicle.states = map[uint32]proto.VehicleState{}
		}
		previous := a.vehicle.states[v.ID]
		if v.State == proto.VehicleOff {
			delete(a.vehicle.states, v.ID)
		} else {
			a.vehicle.states[v.ID] = v
		}
		if a.wv != nil {
			if v.State == proto.VehicleCasting && previous.State != v.State {
				if actor := a.playerActor(v.ID); actor != nil {
					actor.PlayCast(85, float32(v.CastMS)/1000)
				}
				a.startMountCastEffects(v.ID, float64(v.CastMS)/1000)
				if pos, ok := a.effectPosition(v.ID); ok {
					a.playSound("ef_in_atrbu01.wav", pos[0], pos[2])
				}
			}
			if previous.State == proto.VehicleCasting && v.State != previous.State {
				a.stopCastEffects(v.ID)
				if actor := a.playerActor(v.ID); actor != nil {
					actor.CancelCast()
				}
				if v.State == proto.VehicleDriver {
					a.startMountAppearance(v)
				}
			}
			if v.State == proto.VehicleOff && previous.State >= proto.VehicleDriver {
				a.startMountAppearance(previous)
			}
		}
		if v.ID != a.state.SelfID {
			return
		}
		if v.State >= proto.VehicleDriver && previous.State < proto.VehicleDriver {
			a.cancelSkillIntentOnMount()
		}
		a.vehicle.cooldownEnd = a.clock + float64(v.CooldownMS)/1000
		if v.State != proto.VehicleCasting {
			a.vehicle.pending = false
		}
		if v.State >= proto.VehicleDriver && a.vehicle.castState < proto.VehicleDriver {
			a.questMessage(a.chatEntityName(v.ID) + " " + a.msg.Get(5243))
		}
		if v.State == proto.VehicleOff && a.vehicle.castState >= proto.VehicleDriver {
			a.questMessage(a.chatEntityName(v.ID) + " " + a.msg.Get(5244))
		}
		a.vehicle.castState = v.State
		return
	}
	if proto.Op(p) != proto.UseVehicle {
		return
	}
	r, err := proto.DecodeVehicleResult(p)
	if err != nil {
		return
	}
	a.vehicle.pending = false
	if r.Result == 0 {
		a.questMessage(a.chatEntityName(a.state.SelfID) + " " + a.msg.Get(5245))
	}
}

func (a *App) cancelSkillIntentOnMount() {
	a.skills.queuedSkill = nil
	a.revolver.cast = revolverCast{}
	v := a.wv
	if v == nil {
		return
	}
	if len(v.approachSkill) > 0 {
		v.approachSkill, v.approachTarget, v.travel = nil, 0, false
	}
	v.requestedSkill, v.requestedSkills = nil, nil
	v.attackMoveUntil = 0
	if v.self.ActionActive() {
		v.self.ReleaseAction()
	}
	a.stopCastEffects(a.state.SelfID)
}

// Native 0x4178a0 creates sequence 1 twice, following the casting hands.
func (a *App) startMountCastEffects(id uint32, seconds float64) {
	a.stopCastEffects(id)
	for hand := 0; hand < 2; hand++ {
		if x := a.loadSkillEffect("Effect/IN_mc_yellow.EFT", 1); x != nil {
			if len(a.wv.castEffects) >= 16 {
				a.wv.castEffects = a.wv.castEffects[1:]
			}
			a.wv.castEffects = append(a.wv.castEffects, castEffect{id, hand, a.clock + seconds, x})
		}
	}
}

// Native 0x417cc0: appearance sound is selected by race, not MON Sounds.
func mountAppearanceSound(race, model uint8) string {
	if scene.VehicleAlternate(model) {
		return ""
	}
	switch race {
	case 0:
		return "run_horse-roar01.wav"
	case 1, 2:
		return "run_pumav-roar01.wav"
	case 3:
		return "run_baslk-roar01.wav"
	}
	return ""
}
func (a *App) startMountAppearance(v proto.VehicleState) {
	pos, ok := a.effectPosition(v.ID)
	if !ok {
		return
	}
	a.playSound(mountAppearanceSound(v.Race, v.Model), pos[0], pos[2])
	a.playSound("ef_dmg_transf01.wav", pos[0], pos[2])
	if a.assets == nil || len(a.wv.impacts) >= 32 {
		return
	}
	if x := a.loadSkillEffect("Effect/DA_target_transformation.EFT", 0); x != nil {
		matrix := render.Identity()
		matrix[12], matrix[13], matrix[14] = pos[0], pos[1], pos[2]
		x.SetTransform(matrix)
		a.wv.impacts = append(a.wv.impacts, skillImpact{target: v.ID, instance: x, authoredLifetime: true, positionOnly: true})
	}
}

func (a *App) updateVehicleInvite(in *platform.Input, w, h int) bool {
	if a.vehicleMessage == nil || a.state == nil {
		return false
	}
	if a.vehicle.inviteSerial != a.state.VehicleInviteSerial {
		a.vehicle.inviteSerial = a.state.VehicleInviteSerial
		a.vehicle.expires = a.clock + 30
		a.vehicle.peerName = a.chatEntityName(a.state.VehicleInviteID)
		a.vehicleMessage.Open(a.chatEntityName(a.state.VehicleInviteID)+" "+a.msg.Get(5367), true)
	}
	if a.vehicle.replySerial != a.state.VehicleReplySerial {
		a.vehicle.replySerial = a.state.VehicleReplySerial
		id := 5370
		if a.state.VehicleReply == 0 {
			id = 5369
		} else if a.state.VehicleReply == 1 {
			id = 5368
		}
		a.questMessage(a.vehicle.peerName + " " + a.msg.Get(id))
		a.vehicleMessage.Visible = false
	}
	if !a.vehicleMessage.Visible {
		return false
	}
	if a.clock >= a.vehicle.expires {
		a.vehicleMessage.Visible = false
		return false
	}
	answer := a.vehicleMessage.Update(in, w, h)
	if answer != ui.MessageNone && a.world != nil {
		if err := a.world.Send(proto.VehicleAnswerPacket(answer == ui.MessageOK)); err != nil {
			a.backToLoginID(12203)
		}
	}
	return true
}
func (a *App) driverMount(id uint32) *scene.Vehicle {
	if id == a.state.SelfID {
		return a.wv.selfMount
	}
	if c := a.wv.actors[id]; c != nil {
		return c.mount
	}
	return nil
}
func (a *App) attachVehiclePassengers() {
	attach := func(v proto.VehicleState, c *scene.Character) {
		if v.State != proto.VehiclePassenger || c == nil {
			return
		}
		if m := a.driverMount(v.Driver); m != nil {
			world := c.World
			height := float32(math.Sqrt(float64(world[4]*world[4] + world[5]*world[5] + world[6]*world[6])))
			pos := m.RiderSeat(true)
			for i := 0; i < 12; i++ {
				world[i] = m.World[i]
			}
			world[4] *= height
			world[5] *= height
			world[6] *= height
			world[12], world[13], world[14] = pos[0], pos[1], pos[2]
			c.World = world
		}
	}
	attach(a.state.Self.Vehicle, a.wv.self)
	for id, e := range a.state.Entities {
		if c := a.wv.actors[id]; c != nil {
			attach(e.Vehicle, c.character)
		}
	}
}
