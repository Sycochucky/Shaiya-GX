package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

// Panel capture is limited to pointer activity inside its visible rectangle.
// Idle hover does not consume hotkeys or held movement keys.
func panelPointer(in *platform.Input, x, y, w, h int) bool {
	return in.MouseX >= x && in.MouseX < x+w && in.MouseY >= y && in.MouseY < y+h &&
		(in.Pressed || in.Clicked || in.MouseDown || in.RightPressed || in.RightDown || in.Wheel != 0)
}

func (a *App) checkPendingActions() bool {
	if a.personalMarketTimeout() {
		return false
	}
	var sitPending bool
	var sitDeadline float64
	if a.wv != nil {
		sitPending, sitDeadline = a.wv.sitPending, a.wv.sitDeadline
	}
	for _, p := range []struct {
		pending  bool
		deadline float64
		name     string
	}{
		{a.inventory.pending, a.inventory.deadline, a.tr(12266)},
		{a.vehicle.pending, a.vehicle.deadline, a.tr(7034)},
		{a.exchange.mutation != 0, a.exchange.deadline, a.tr(5089)},
		{a.duel.mutation != 0, a.duel.deadline, a.tr(12267)},
		{a.guild.createPending, a.guild.createDeadline, a.tr(12268)},
		{a.guild.lifecyclePending, a.guild.lifecycleDeadline, a.tr(12269)},
		{a.shop.pending, a.shop.deadline, a.tr(12270)},
		{a.skills.pending, a.skills.deadline, a.tr(12271)},
		{a.status.pending, a.status.deadline, a.tr(12272)},
		{sitPending, sitDeadline, a.tr(12273)},
	} {
		if p.pending && a.clock >= p.deadline {
			a.backToLogin(a.msg.Format(12278, map[string]string{"b": p.name}))
			return false
		}
	}
	return true
}

// Independent panels can stay open together. Captured events never reach a
// world target or logout; held movement keys remain available.
func (a *App) worldControls(in *platform.Input, w, h int) platform.Input {
	a.validateInspect()
	vInventory := &a.inventory
	defer func() {
		if !in.MouseDown || in.Clicked || !vInventory.open {
			vInventory.cancelItemPointer()
		}
	}()
	opened := a.personalPanelMask()
	defer func() { a.placeNewPersonalPanels(opened, w, h) }()
	controls := *in
	v := a.wv
	if a.socialInviteUI != nil && a.socialInviteUI.Visible {
		a.updateSocialInvitation(in, w, h)
		return platform.Input{MouseX: in.MouseX, MouseY: in.MouseY, SuppressAutoRun: true}
	}
	if a.menuOpen {
		var captured bool
		controls, captured = a.gameMenuControls(in, w, h)
		if captured {
			return controls
		}
		in = &controls
	}
	if a.personalMarket.modal() && a.updatePersonalMarketModal(in, w, h) {
		return platform.Input{MouseX: in.MouseX, MouseY: in.MouseY, SuppressAutoRun: true}
	}
	if a.gateConfirm != nil && a.gateConfirm.Visible {
		a.updateGateConfirm(in, w, h)
		return platform.Input{MouseX: in.MouseX, MouseY: in.MouseY}
	}
	if a.pvpRewardMessage != nil && a.pvpRewardMessage.Visible {
		if a.state.Self.Dead || a.state.Entities[a.shop.npcID] == nil {
			a.pvpRewardMessage.Visible = false
			a.pvpReward.confirm = 0
		} else {
			a.updatePvPRewards(in, w, h)
		}
		return platform.Input{MouseX: in.MouseX, MouseY: in.MouseY, SuppressAutoRun: true}
	}
	if a.blacksmith.quantity {
		a.updateEnchantQuantity(in)
		return platform.Input{MouseX: in.MouseX, MouseY: in.MouseY}
	}
	if a.gemConfirm != nil && a.gemConfirm.Visible {
		a.updateGemConfirm(in, w, h)
		return platform.Input{MouseX: in.MouseX, MouseY: in.MouseY}
	}
	if a.duel.quantity != nil || a.duel.moneyOpen {
		a.updateDuelModal(in)
		return platform.Input{MouseX: in.MouseX, MouseY: in.MouseY}
	}
	if a.exchange.quantity != nil || a.exchange.moneyOpen {
		a.updateExchangeModal(in)
		return platform.Input{MouseX: in.MouseX, MouseY: in.MouseY}
	}
	if a.shop.quantity != nil {
		if in.Pressed {
			v.uiLeft = 5
		}
		if in.Clicked || !in.MouseDown {
			v.uiLeft = 0
		}
		a.updateTradeQuantity(in)
		return platform.Input{MouseX: in.MouseX, MouseY: in.MouseY}
	}
	// A right drag belongs to its press owner until release. Virtual cursor
	// movement or a panel opened by a hotkey cannot transfer it to the UI.
	cameraDrag := v.mouselook && in.RightDown && !in.RightPressed && !v.uiRight
	if cameraDrag {
		pointer := *in
		pointer.RightDown = false
		pointer.Wheel = 0
		in = &pointer
	}
	vInventory.advanceItemDrag(in)
	if a.auction.npc != 0 && a.auction.modal == "setup" && in.Clicked && vInventory.dragging {
		dialog := auctionDialogBox("setup", w, h)
		if dialog.source().contains(in.MouseX, in.MouseY) && a.auctionSource() {
			v.uiLeft = 0
			return platform.Input{MouseX: in.MouseX, MouseY: in.MouseY, SuppressAutoRun: true}
		}
	}
	if a.auction.npc != 0 && a.auction.focus >= 0 && !in.Pressed && (len(in.Chars) > 0 || len(in.Keys) > 0) && a.updateAuction(in, w, h) {
		return platform.Input{MouseX: in.MouseX, MouseY: in.MouseY, SuppressAutoRun: true}
	}

	captured := false
	if a.teleportStone.open && a.recordedPoints != nil {
		const recordedPointOwner = -4
		panel := a.recordedPoints
		x, y := panel.Origin(w, h)
		over := pointInside(in, x, y, ui.RecordedPointsWidth, ui.RecordedPointsHeight)
		if a.windowDrag.position == &panel.Position || a.windowDrag.position == nil && over && in.Pressed {
			captured = a.updateWindowDrag(in, w, h)
		}
		if !captured && a.windowDrag.position == nil {
			input := *in
			if in.Clicked && !in.Pressed && v.uiLeft != recordedPointOwner {
				input.Clicked = false
			}
			captured = a.updateTeleportStone(&input, w, h)
			if captured && in.Pressed {
				v.uiLeft = recordedPointOwner
			}
		}
	}
	if !captured && a.chatUI != nil && (a.chatUI.ModeOpen || a.chatUI.OptionsOpen()) {
		captured = a.updateChatPointer(in, w, h)
		if captured && in.Pressed {
			v.uiLeft = -3
		}
	}
	if !captured {
		captured = a.updateQuickPotionDrop(in, w, h)
		if !captured {
			captured = a.updateShortcutDrag(in, w, h)
		}
	}
	if a.state == nil || a.wv == nil {
		return platform.Input{SuppressAutoRun: true}
	}
	underMall := false
	if a.mall.open && a.mallUI != nil {
		x, y := a.mallUI.Origin(w, h)
		underMall = pointInside(in, x, y, ui.MallWidth, ui.MallHeight)
		if a.mallUI.GiftOpen {
			x, y = a.mallUI.GiftOrigin(w, h)
			underMall = underMall || pointInside(in, x, y, ui.GiftBoxWidth, ui.GiftBoxHeight)
		}
	}
	if !underMall && !captured {
		captured = a.updatePlayerContext(in, w, h)
	}
	if !underMall && !captured {
		captured = a.dropPersonalMarketItem(in, w, h)
	}
	if !captured {
		captured = a.updateDuelDrag(in, w, h)
	}
	if !captured {
		captured = a.updateExchangeDrag(in, w, h)
	}
	if !captured {
		captured = a.updateLocalActionDrag(in, w, h)
	}
	if !captured {
		captured = a.updateWindowDrag(in, w, h)
	}
	if !captured {
		captured = a.updateBlacksmithDrag(in, w, h)
	}
	if !captured {
		captured = a.updateEmblemDrag(in, w, h)
	}
	if !captured {
		captured = a.updateWarehouseDrag(in, w, h)
	}
	if !captured {
		captured = a.updateCommerceDrag(in, w, h)
	}
	if a.wv.mapView.full && !captured {
		captured = a.updateMinimap(in, w, h)
		if captured && in.Pressed {
			v.uiLeft = -1 // The topmost full map owns release, even outside its frame.
		}
	}

	for index, update := range []func(*platform.Input, int, int) bool{a.updateDropDialog, a.updateMall, a.updatePersonalMarket, a.updateGuild, a.updateEvolutionBar, a.updateQuickPotions, a.updateRevolver, a.updateBarEditing, a.updateQuestLog, a.updateStatus, a.updateSubActions, a.updateSkills, a.updateInspect, a.updateInventory, a.updateDuel, a.updateExchange, a.updateShop, a.updateQuestTracker, a.updateBottom, a.updateHUDPointer, a.updateChatPointer, a.updateBuffs, a.updateMinimap, a.updatePartyPointer} {
		if captured {
			break
		}
		input := *in
		if in.Clicked && !in.Pressed && v.uiLeft != index+1 {
			// Keep the release as a hit-test event, but only its original
			// owner may activate a control. This also blocks click-through.
			input.Clicked = false
			input.MouseDown = true
		}
		if update(&input, w, h) {
			captured = true
			if in.Pressed {
				v.uiLeft = index + 1
			}
			break
		}
	}
	a.dropReleased(in, w, h, captured)
	if in.Pressed && !captured {
		v.uiLeft = 0
	}
	if in.RightPressed {
		v.uiRight = captured
	}
	pointerOwned := v.uiLeft != 0 || v.uiRight
	if in.Clicked || !in.MouseDown {
		v.uiLeft = 0
	}
	if !in.RightDown {
		v.uiRight = false
	}
	if captured {
		controls.Chars = nil
		controls.Actions = nil
		controls.Keys = nil
		controls.Released = ""
	}
	if captured || pointerOwned {
		controls.Pressed = false
		controls.Clicked = false
		controls.MouseDown = false
		controls.RightPressed = false
		controls.RightDown = false
		controls.Wheel = 0
		controls.MiddlePressed = false
	}

	// Native 0x43bfa0 picks and acts when the left button goes down, not when
	// it is released. A press that no panel owns reaches the world as a click
	// and its later release is swallowed. Both buttons down is mouse-run.
	holding := a.inventory.dragging || a.shop.dragging || a.bank.dragging || a.skills.drag.active || a.wv.localActionDrag.active
	if cameraDrag {
		controls.RightDown = true
	}
	v.worldLeftPress(&controls, in, !captured && !pointerOwned && !holding)
	if controls.HasKey(platform.KeyShowIDs) {
		a.showCatalogIDs = !a.showCatalogIDs
	}
	if controls.HasKey(platform.KeyEscape) || hasChar(&controls, 'o') {
		a.openGameMenu()
		controls.Keys, controls.Chars = nil, nil
	}
	return controls
}

func (v *worldView) worldLeftPress(controls, pointer *platform.Input, free bool) {
	if !free {
		controls.Pressed, controls.Clicked = false, false
	}
	if controls.Pressed {
		if controls.RightDown {
			controls.Clicked = false
		} else {
			controls.Clicked = true
		}
		v.worldPressActed = !pointer.Clicked
	} else if v.worldPressActed && (pointer.Clicked || !pointer.MouseDown) {
		controls.Clicked = false
		v.worldPressActed = false
	}
}
