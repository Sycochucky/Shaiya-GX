package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

const contextRowHeight = ui.PlayerMenuRowHeight

type playerContext struct {
	id            uint32
	x, y, held    int
	width, height int
	heldLabel     string
	open          bool
	checkRange    bool
	mode, slot    int // contextClassic, contextPortrait (own portrait) or contextUnion (a raid slot)
}
type contextAction struct {
	label  string
	packet []byte
	slot   int    // Move group: the picked slot
	name   string // union whisper: the member's name
}

func (a *App) contextActions(id uint32) []contextAction {
	if id == 0 || a.state == nil {
		return nil
	}
	if id != a.state.SelfID && a.state.Party.Member(id) == nil {
		if e := a.state.Entities[id]; e == nil || e.Kind != game.KindUser {
			return nil
		}
	}
	member := a.state.Party.Member(id)
	leader := a.state.Party.Leader == a.state.SelfID
	sameFaction := true
	if e := a.state.Entities[id]; e != nil && e.Shape != nil {
		if own, ok := a.selfSlot(); ok {
			sameFaction = own.Race/2 == e.Shape.Race/2
		}
	}
	if !sameFaction && id != a.state.SelfID {
		if a.canViewPersonalMarket(id) {
			return []contextAction{{label: "View market"}}
		}
		return nil
	}
	var out []contextAction
	if a.canViewPersonalMarket(id) {
		out = append(out, contextAction{label: "View market"})
	}
	if sameFaction && id != a.state.SelfID {
		if e := a.state.Entities[id]; e != nil {
			own, other := a.state.Self.Vehicle, e.Vehicle
			if own.State == proto.VehicleDriver && own.Seats == 2 && own.Passenger == 0 && other.State == proto.VehicleOff || other.State == proto.VehicleDriver && other.Seats == 2 && other.Passenger == 0 && own.State == proto.VehicleOff {
				out = append(out, contextAction{label: "Go together", packet: proto.VehicleInvitePacket(id)})
			}
		}
	}
	room := a.state.Party.Raid || len(a.state.Party.Members) < 7
	if sameFaction && room && member == nil && id != a.state.SelfID && (len(a.state.Party.Members) == 0 || leader || a.state.Party.Raid && a.state.Party.RaidSubLeader == a.state.SelfID) {
		if a.state.Party.Raid {
			out = append(out, contextAction{label: "Invite to raid", packet: proto.RaidInvitePacket(id)})
		} else {
			out = append(out, contextAction{label: "Invite to party", packet: proto.PartyInvitePacket(id)})
		}
	}
	if member != nil && leader && id != a.state.SelfID {
		if a.state.Party.Raid {
			out = append(out, contextAction{label: "Remove from raid", packet: proto.RaidExpelPacket(id)}, contextAction{label: "Make leader", packet: proto.RaidLeaderPacket(id)}, contextAction{label: "Make assistant", packet: proto.RaidSubLeaderPacket(id)})
		} else {
			out = append(out, contextAction{label: "Remove from party", packet: proto.PartyExpelPacket(id)}, contextAction{label: "Make leader", packet: proto.PartyLeaderPacket(id)})
		}
	}
	if member != nil && id == a.state.SelfID {
		label := "Leave party"
		if a.state.Party.Raid {
			label = "Leave raid"
		}
		out = append(out, contextAction{label: label, packet: proto.PartyLeavePacket()})
	}
	if id != a.state.SelfID && !a.state.Exchange.Active && a.state.Duel.Phase == game.DuelNone && a.state.GuildDuel.Phase == game.DuelNone {
		out = append(out, contextAction{label: "Trade", packet: proto.ExchangeRequestPacket(id)}, contextAction{label: "Duel", packet: proto.DuelRequestPacket(id)})
	}
	if id != a.state.SelfID && a.state.Guild.ID != 0 && a.state.Guild.Rank > 0 && a.state.Guild.Rank <= 4 && a.state.GuildDuel.Phase == game.DuelNone && a.state.Duel.Phase == game.DuelNone && !a.state.Exchange.Active {
		if e := a.state.Entities[id]; e != nil && e.Kind == game.KindUser && e.GuildID != 0 && e.GuildID != a.state.Guild.ID {
			out = append(out, contextAction{label: "Guild duel", packet: proto.GuildDuelRequestPacket(id)})
		}
	}
	if id != a.state.SelfID && a.chatEntityName(id) != "" {
		out = append(out, contextAction{label: "Whisper", packet: nil})
	}
	if a.inspectUI != nil && a.inspectTarget(id) != nil {
		out = append(out, contextAction{label: "Inspect"})
	}
	return out
}

func (a *App) contextLabels(actions []contextAction) []string {
	labels := make([]string, len(actions))
	for i, action := range actions {
		id := 0
		switch action.label {
		case "Invite to party", "Invite to raid":
			id = 5002
		case "Remove from party":
			id = 5003
		case "Remove from raid":
			id = 6509
		case "Make leader":
			id = 5004
			if a.state.Party.Raid {
				id = 6507
			}
		case "Make assistant":
			id = 6508
		case "Leave party", "Leave raid":
			id = 5005
		case "Go together":
			id = 5366
		case "Trade":
			id = 5091
		case "View market":
			id = 5090
		case "Duel":
			id = 5008
		case "Guild duel":
			id = 5339
		case "Whisper":
			id = 5007
		case "Inspect":
			id = 5421
		case "Create union":
			id = 6500
		case "Leave union":
			id = 6501
		case "Dismantle union":
			id = 6502
		case "Config union":
			id = 6503
		case "Move group":
			id = 6506
		case "Delegate leader":
			id = 6507
		case "Union whisper":
			id = 6510
		}
		labels[i] = a.tr(id)
	}
	return labels
}

func (a *App) performContextAction(id uint32, action contextAction) {
	switch action.label {
	case "Inspect":
		a.requestInspect(id)
		return
	case "View market":
		a.viewPersonalMarket(id)
		return
	case "Create union":
		a.openUnionCreate()
		return
	case "Config union":
		a.openUnionConfig()
		return
	case "Dismantle union":
		a.askDismantle()
		return
	case "Move group":
		a.party.moving, a.party.moveSlot = true, action.slot
		return
	}
	if action.label == "Whisper" || action.label == "Union whisper" {
		name := a.chatEntityName(id)
		if action.label == "Union whisper" {
			name = action.name
		}
		if name == "" || a.chatUI == nil {
			return
		}
		a.chatUI.Channel, a.chatUI.Recipient = 5, name
		a.chatUI.ModeOpen, a.chatUI.Typing = false, true
		a.openChatComposer()
		a.chatUI.RecipientFocus = false
		return
	}
	if action.label == "Go together" {
		a.vehicle.peerName = a.chatEntityName(id)
	}
	if action.label == "Trade" && a.state.Exchange.Active {
		a.questMessageID(12169)
	} else if a.world != nil && action.packet != nil {
		if err := a.world.Send(action.packet); err != nil {
			a.backToLoginID(12168)
			return
		}
		if action.label == "Invite to party" || action.label == "Invite to raid" {
			a.invitedNotice(a.chatEntityName(id))
		}
	}
}

// EP5.4 0x47e29b checks 3D distance against the 10.0 value at 0x6bc9c8.
func (a *App) playerContextInRange(id uint32) bool {
	e := a.state.Entities[id]
	if e == nil {
		return false
	}
	pos := a.wv.visualPosition()
	dx, dy, dz := pos[0]-e.X, pos[1]-e.Y, pos[2]-e.Z
	return dx*dx+dy*dy+dz*dz <= 10*10
}

func (a *App) updatePlayerContext(in *platform.Input, w, h int) bool {
	c := &a.playerContext
	if c.open {
		actions := a.menuActions(c)
		if len(actions) == 0 || c.checkRange && !a.playerContextInRange(c.id) || c.mode == contextClassic && a.state.Entities[c.id] == nil && a.state.Party.Member(c.id) == nil {
			c.open = false
			return false
		}
		c.width, c.height = ui.PlayerMenuSize(a.font, a.contextLabels(actions))
		c.x, c.y = max(0, min(c.x, w-c.width)), max(0, min(c.y, h-c.height))
		if in.RightPressed || in.HasKey(platform.KeyEscape) {
			c.open = false
			return true
		}
		row := ui.PlayerMenuRowAt(in.MouseX, in.MouseY, c.x, c.y, c.width, len(actions))
		if in.Pressed {
			c.held = row
			c.heldLabel = ""
			if row >= 0 {
				c.heldLabel = actions[row].label
			}
			if !pointInside(in, c.x, c.y, c.width, c.height) {
				c.open = false
			}
			return true
		}
		if in.Clicked {
			if row >= 0 && row == c.held && actions[row].label == c.heldLabel {
				a.performContextAction(c.id, actions[row])
			}
			c.open = false
			return true
		}
		return panelPointer(in, c.x, c.y, c.width, c.height)
	}
	if !in.RightPressed || a.wv == nil || a.wv.mapView.full {
		return false
	}
	if a.barCovered(in, w, h) || a.chatUI != nil && a.chatUI.Contains(in.MouseX, in.MouseY, h) {
		return false
	}
	if a.actionBar != nil && a.actionBar.SlotAt(in.MouseX, in.MouseY, w, h) >= 0 {
		return false
	}
	if a.minimap != nil && !a.wv.mapView.compactHidden {
		x, y := a.minimap.Origin(w, h)
		if pointInside(in, x, y, 196, 197) {
			return false
		}
	}
	id := uint32(0)
	hx, hy := 8, 8
	if a.hud != nil {
		hx, hy = a.hud.Origin(w, h)
	}
	mode, slot := contextClassic, 0
	checkRange := false
	if pointInside(in, hx, hy, 213, 114) && a.state.Party.Member(a.state.SelfID) != nil {
		id, mode = a.state.SelfID, contextPortrait
	}
	if s, over := a.unionSlotAt(in, w, h); over && id == 0 {
		if s >= 0 {
			id, mode, slot = 1, contextUnion, s
		} else {
			return false
		}
	}
	tx, ty := 256, 0
	if a.hud != nil {
		tx, ty = a.hud.TargetOrigin(w, h)
	}
	if id == 0 && pointInside(in, tx, ty, 256, 64) {
		if a.state.Target == a.state.SelfID {
			id = a.state.SelfID
		} else if e := a.state.Entities[a.state.Target]; e != nil && e.Kind == game.KindUser {
			id, checkRange = e.ID, true
		}
	}
	if id == 0 && a.partyUI != nil && len(a.state.Party.Members) > 0 {
		if !a.state.Party.Raid {
			x, y := a.partyUI.Origin(w, h, len(a.state.Party.Members))
			if in.MouseX >= x && in.MouseX < x+ui.PartyWidth && in.MouseY >= y && in.MouseY < y+len(a.state.Party.Members)*ui.PartyRowHeight {
				id = a.state.Party.Members[(in.MouseY-y)/ui.PartyRowHeight].ID
			}
		}
	}
	if id == 0 && !a.barCovered(in, w, h) && in.MouseY < h-ui.BottomBarHeight(w, h) {
		// Use the same model/depth picking as left-click selection. World
		// right-click must not target a player through another actor or wall.
		pointer := *in
		pointer.Clicked = true
		picked := a.pickMob(&pointer, w, h)
		if e := a.state.Entities[picked]; picked == a.state.Target && e != nil && e.Kind == game.KindUser {
			id, checkRange = picked, true
		}
	}
	if checkRange && !a.playerContextInRange(id) {
		return false
	}
	next := playerContext{id: id, mode: mode, slot: slot, checkRange: checkRange}
	actions := a.menuActions(&next)
	if len(actions) == 0 {
		return false
	}
	width, height := ui.PlayerMenuSize(a.font, a.contextLabels(actions))
	next.x, next.y, next.width, next.height, next.held, next.open = max(0, min(in.MouseX, w-width)), max(0, min(in.MouseY, h-height)), width, height, -1, true
	*c = next
	return true
}
func (a *App) drawPlayerContext(w, h int) {
	c := &a.playerContext
	if !c.open || a.state == nil || a.playerMenuUI == nil {
		return
	}
	actions := a.menuActions(c)
	labels := a.contextLabels(actions)
	c.width, c.height = ui.PlayerMenuSize(a.font, labels)
	c.x, c.y = max(0, min(c.x, w-c.width)), max(0, min(c.y, h-c.height))
	hover := ui.PlayerMenuRowAt(a.lastMouseX, a.lastMouseY, c.x, c.y, c.width, len(actions))
	a.clearTooltipUnderPanel(c.x, c.y, c.width, c.height)
	a.playerMenuUI.Draw(a.sprites, a.font, a.white, c.x, c.y, c.width, labels, hover)
}
