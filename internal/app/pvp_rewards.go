package app

import (
	"fmt"

	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/pvprank"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

type pvpRewardView struct {
	confirm  int
	pending  proto.Opcode
	deadline float64
}

func (a *App) openPvPRewards() {
	a.pvpReward = pvpRewardView{}
	if a.pvpRewardUI != nil {
		a.pvpRewardUI.Reset()
	}
	if a.pvpRewardMessage != nil {
		a.pvpRewardMessage.Visible = false
	}
	c, _ := a.selfSlot()
	if c.Level < 16 {
		a.questMessage(a.msg.Get(5347))
		a.shop.npcID = 0
		return
	}
	a.state.PvPRewards = proto.PvPRewards{}
	a.sendPvPReward(proto.PvPRewardStatus)
}

func (a *App) sendPvPReward(op proto.Opcode) {
	if a.world == nil {
		return
	}
	if err := a.world.Send(proto.NewWriter(op).Bytes()); err != nil {
		a.backToLoginID(12172)
		return
	}
	a.pvpReward.pending, a.pvpReward.deadline = op, a.clock+10
}

func (a *App) confirmedPvPReward(op proto.Opcode) {
	if op != a.pvpReward.pending {
		return
	}
	a.pvpReward.pending = 0
	if op == proto.PvPRewardStatus {
		return
	}
	v := a.state.LastPvPReward
	if v.OK && op == proto.PvPRewardKill {
		a.questMessage(a.msg.Format(5349, map[string]string{"b": fmt.Sprint(v.Value)}))
	} else if v.OK && op == proto.PvPRewardDeath {
		a.questMessage(a.msg.Format(53, map[string]string{"b": fmt.Sprint(v.Value)}))
	} else if !v.OK {
		a.questMessage(a.msg.Get(5350))
	}
}

func (a *App) pvpRewardState() ui.PvPRewardState {
	r := a.state.PvPRewards
	threshold, _, _, killOK := pvprank.Reward(r.KillStage)
	deathThreshold, _, _, deathOK := pvprank.Reward(r.DeathStage)
	state := ui.PvPRewardState{Description: a.msg.Lookup(12375), Busy: a.pvpReward.pending != 0,
		CanKill:  r.Known && killOK && a.state.Details.Kills >= threshold,
		CanDeath: r.Known && deathOK && a.state.Details.Deaths >= deathThreshold}
	if e := a.state.Entities[a.shop.npcID]; e != nil {
		def := a.npcDefs[sdata.NPCKey{Type: e.NpcType, ID: e.TypeID}]
		state.Name, state.Portrait = def.Name, a.npcPortrait(def)
	}
	if !r.Known {
		return state
	}
	killStage, deathStage := min(r.KillStage, pvprank.RewardCount), min(r.DeathStage, pvprank.RewardCount)
	threshold, points, _, _ := pvprank.Reward(killStage)
	deathThreshold, _, gold, _ := pvprank.Reward(deathStage)
	state.Reward = a.msg.Format(5353, map[string]string{"v": fmt.Sprint(killStage), "v2": fmt.Sprint(points)})
	state.Kills = fmt.Sprintf("%d / %d", a.state.Details.Kills, threshold)
	state.Deaths = fmt.Sprintf("%d / %d", a.state.Details.Deaths, deathThreshold)
	state.DeathLevel = fmt.Sprintf("%d %s / %d %s", deathStage, a.msg.Get(5354), pvprank.RewardCount, a.msg.Get(5354))
	state.Gold = fmt.Sprintf("%s %s", ui.GoldText(gold), a.msg.Get(5355))
	if r.KillStage == proto.PvPRewardFinishedStage {
		state.KillStatus = a.msg.Lookup(12376)
	}
	if r.DeathStage == proto.PvPRewardFinishedStage {
		state.DeathStatus = a.msg.Lookup(12376)
	}
	if threshold > 0 {
		state.KillProgress = float32(a.state.Details.Kills) / float32(threshold)
	}
	// Native 0x0051a09a multiplies the claim level by float32(1/31).
	state.DeathProgress = float32(deathStage) * float32(1.0/31)
	return state
}

func (a *App) updatePvPRewards(in *platform.Input, w, h int) bool {
	if a.pvpReward.pending != 0 && a.clock >= a.pvpReward.deadline {
		a.pvpReward.pending = 0
	}
	if a.pvpRewardMessage != nil && a.pvpRewardMessage.Visible {
		if result := a.pvpRewardMessage.Update(in, w, h); result != ui.MessageNone {
			claim := a.pvpReward.confirm
			a.pvpReward.confirm = 0
			if result == ui.MessageOK && a.pvpReward.pending == 0 {
				op := proto.PvPRewardKill
				if claim == 2 {
					op = proto.PvPRewardDeath
				}
				a.sendPvPReward(op)
			}
		}
		return true
	}
	claim, closed := a.pvpRewardUI.Update(in, w, h, a.pvpRewardState())
	if closed {
		if a.world != nil {
			if err := a.world.Send(proto.NpcTalkPacket(0)); err != nil {
				a.backToLoginID(12011)
			}
		}
		a.shop.npcID = 0
		return true
	}
	if claim != 0 && a.pvpReward.pending == 0 && a.pvpRewardMessage != nil {
		r := a.state.PvPRewards
		stage := r.KillStage
		if claim == 2 {
			stage = r.DeathStage
		}
		count, points, _, ok := pvprank.Reward(stage)
		if ok {
			id := 5351
			if claim == 2 {
				id = 5352
			}
			a.pvpReward.confirm = claim
			a.pvpRewardMessage.OpenFull(a.msg.Format(id, map[string]string{"p": fmt.Sprint(count), "b": fmt.Sprint(points)}), true, a.font)
		}
	}
	x, y := a.pvpRewardUI.Origin(w, h)
	return panelPointer(in, x, y, ui.PvPRewardWidth, ui.PvPRewardHeight)
}

func (a *App) drawPvPRewards(w, h int) {
	x, y := a.pvpRewardUI.Origin(w, h)
	a.clearTooltipUnderPanel(x, y, ui.PvPRewardWidth, ui.PvPRewardHeight)
	a.pvpRewardUI.Draw(a.sprites, a.font, w, h, a.pvpRewardState())
}
