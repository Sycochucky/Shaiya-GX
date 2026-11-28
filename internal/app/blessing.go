package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

func (a *App) confirmedBlessing(op proto.Opcode) {
	if op == proto.Logout {
		a.blessingBannerUntil = 0
		a.blessingFeedbackActive = false
		return
	}
	if op != proto.GXBlessingSnapshot || a.state == nil {
		return
	}
	b := &a.state.Blessing
	if b.Country > 1 || b.Country != a.state.Faction.Faction || b.Maximum == 0 {
		return
	}
	maximum := b.Count == b.Maximum && b.RemainingMS > 0
	if b.Entry {
		a.blessingFeedbackActive = maximum
		return
	}
	if !maximum {
		a.blessingFeedbackActive = false
		return
	}
	// TCP normally delivers one activation. A repeated maximum notification
	// must not replay the burst, including after reentry.
	if a.blessingFeedbackActive {
		return
	}
	a.blessingFeedbackActive = true
	a.blessingBannerUntil = a.clock + 5
	if a.sound != nil {
		name := a.audioPath("ui_etainbless01.wav", false)
		a.audioError(name, a.sound.Play(name, 1))
	}
	a.startBlessingEffect(b.Country)
}

// Original 0055fc90 attaches type 15, sequence country, to the local actor.
// Every winning-country client does this for itself; it is never replicated
// onto remote rigs. Allow this rare event over the ordinary 32-impact budget
// so a busy PvP frame cannot swallow the faction activation.
func (a *App) startBlessingEffect(country uint8) {
	if a.wv == nil || a.assets == nil || a.state.SelfID == 0 {
		return
	}
	if x := a.loadSkillEffect("Effect/bless.EFT", int(country)); x != nil {
		if m, ok := a.effectTargetMatrix(a.state.SelfID); ok {
			x.SetTransform(m)
		}
		a.wv.impacts = append(a.wv.impacts, skillImpact{target: a.state.SelfID, instance: x, authoredLifetime: true})
	}
}

// Native 00484be0 lists the nonzero cumulative types in numeric order using
// SysMsg 5127..5153. It does not print amounts or an invented title.
func (a *App) blessingLines() []string {
	if a.state == nil || !a.state.Blessing.Known {
		return nil
	}
	var lines []string
	for i, value := range a.state.Blessing.Bonuses {
		if value != 0 {
			lines = append(lines, a.msg.Get(5127+i))
		}
	}
	return lines
}

func (a *App) drawBlessingTooltip(w, h int) {
	if a.bottom == nil || !a.bottom.BlessingAt(a.lastMouseX, a.lastMouseY, w, h) || a.menuOpen || a.state == nil || a.state.Self.Dead {
		return
	}
	for _, dialog := range []*ui.Message{a.accountMessage, a.partyMessage, a.vehicleMessage, a.guildConfirm, a.unionConfirm, a.auctionConfirm, a.guildDuelMessage, a.duelMessage, a.gemConfirm, a.gateConfirm, a.questConfirm, a.exchangeInvite} {
		if dialog != nil && dialog.Visible {
			return
		}
	}
	if a.appearanceEdit != nil || a.shop.quantity != nil || a.exchange.quantity != nil || a.exchange.moneyOpen || a.duel.quantity != nil || a.duel.moneyOpen || a.blacksmith.quantity {
		return
	}
	if a.barCovered(&platform.Input{MouseX: a.lastMouseX, MouseY: a.lastMouseY}, w, h) {
		return
	}
	lines := a.blessingLines()
	if !a.state.Blessing.Known || a.state.Blessing.Country > 1 || a.state.Blessing.Maximum == 0 {
		return
	}
	// Native 00484be0 has a 400px wide variant. Fit translated text too.
	f := a.tooltipFont
	if f == nil {
		f = a.font
	}
	width := 400
	for _, line := range lines {
		width = max(width, f.Width(line)+20)
	}
	width = min(width, w)
	height := (len(lines) + 1) * 15
	left := max(0, min(a.lastMouseX-10, w-width))
	top := max(0, a.lastMouseY-height)
	a.sprites.Draw(a.white, float32(left), float32(top), float32(width), float32(height), 0, 0, 1, 1, [4]float32{29.0 / 255, 31.0 / 255, 36.0 / 255, 205.0 / 255})

	for i, line := range lines {
		f.Draw(a.sprites, line, left+10, top+15+i*15, white, 0)
	}
}

func (a *App) drawBlessingBanner(w, h int) {
	if a.assets == nil || a.clock >= a.blessingBannerUntil {
		return
	}
	if tex := a.assets.Texture("interface/worldname/event.png"); tex != nil {
		remaining := a.blessingBannerUntil - a.clock
		alpha := float32(min(1, remaining, 5-remaining))
		a.sprites.Draw(tex, float32(w/2-128), float32(h/3-64), 256, 64, 0, 0, 1, 1, [4]float32{1, 1, 1, alpha})
	}
}
