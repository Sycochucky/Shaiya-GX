package world

import (
	"fmt"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"math"
	"strings"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

const positionSaveInterval = 5 * time.Second

func (s *Session) usablePortal(x, y, z float32, index *uint8) (MapPortal, bool) {
	ch := s.char
	var portal MapPortal
	var ok bool
	if index != nil {
		portal, ok = FindPortalByIndex(ch.MapID, *index)
	} else {
		portal, ok = FindPortalAt(ch.MapID, x, y, z)
	}
	if !ok {
		return MapPortal{}, false
	}
	switch {
	case !portal.IsOpen():
		s.log.Printf("GX Server World rejected closed portal index=%d map=%d", portal.Index, ch.MapID)
	case realMap(ch.MapID) != 27 && !portal.AllowsFaction(s.faction):
		s.log.Printf("GX Server World rejected portal faction index=%d map=%d faction=%d", portal.Index, ch.MapID, s.faction)
	case !s.canEnterMap(portal.TargetMapID, ch.Level) || (!s.isAdmin && !portal.AllowsLevel(ch.Level)):
		s.log.Printf("GX Server World rejected portal level index=%d map=%d level=%d", portal.Index, ch.MapID, ch.Level)
	default:
		return portal, true
	}
	return MapPortal{}, false
}

// savePosition persists the character (usp_Save_Char_Info_E).
func (s *Session) savePosition() {
	if s.char == nil {
		return
	}
	if err := db.SaveCharacter(s.char); err != nil {
		s.logErr("save character position", err)
		return
	}
	s.positionDirty = false
	s.lastPositionSave = time.Now()
}

func (s *Session) syncPlayer() {
	if s.player != nil && s.char != nil {
		s.player.UpdateCharacter(*s.char)
	}
}

// applyTeleport is the single path for every teleport (portals, town,
// resurrection): moves the character, persists it, acks the client,
// introduces it to viewers at the destination and refreshes visibility.
func (s *Session) applyTeleport(mapID uint16, x, y, z float32, label string, gmDirect ...bool) bool {
	if s.char == nil || !s.canEnterMap(mapID, s.char.Level) {
		return false
	}
	s.cancelTravelCast()
	s.endMyShop()
	s.closeMyShopView()
	if !s.keepsMountAcrossTeleport(mapID, label) {
		s.dismount()
	}
	s.dash = nil
	s.pendingCombo = nil
	s.pendingSkill = nil
	// /cmove may inspect the base map without joining a party/guild instance.
	direct := (s.isAdmin || s.gmForced) && len(gmDirect) > 0 && gmDirect[0]
	ch := s.char
	// CWorld::GetZoneByMapID: a guild ranking map resolves to the guild's
	// instance while the battle is open; without one the move fails.
	if !direct && mapCreateType(mapID) == grbCreateType {
		zone, ok := s.enterGuildRanking(mapID)
		if !ok {
			s.log.Printf("GX Server World guild ranking map %d refused (closed or no guild)", mapID)
			return false
		}
		mapID = zone
	}
	// GetZoneByMapID's party case: the party's own instance.
	if !direct && !isInstanceZone(mapID) && mapCreateType(mapID) == partyCreateType {
		zone, ok := s.enterPartyInstance(mapID)
		if !ok {
			s.log.Printf("GX Server World party map %d refused (no party)", mapID)
			return false
		}
		mapID = zone
	}
	g, err := s.playerGeometry(mapID)
	if err != nil {
		s.log.Printf("GX Server teleport refused: %v", err)
		return false
	}
	destination, valid := g.Spawn([3]float32{x, y, z})
	if !valid || !g.ValidPlayerHeight(destination[1]) {
		s.log.Printf("GX Server teleport refused: map %d has no physical floor at destination", mapID)
		return false
	}
	x, y, z = destination[0], destination[1], destination[2]
	// LeaveUser / EnterUser: the old view goes (on the same map the client
	// keeps its entities, so they get leave packets), then the new one.
	if mapID != ch.MapID && s.player != nil {
		duelLeave(s.player, 0, false) // PvPChkLeaveByZoneLeave
		guildDuelLeave(s.player, nil, 1)
	}
	if s.auctionContact {
		s.useNpc(0)
	}
	if mapID != ch.MapID {
		s.endGMGuildRanking(false)
	}
	s.clearView(mapID == ch.MapID)
	previousMap := ch.MapID
	changedMap := mapID != previousMap
	ch.MapID, ch.PosX, ch.PosY, ch.PosZ = mapID, x, y, z
	s.movement = playerMovement{geometry: g, mapID: mapID, ladder: s.movement.ladder}
	if err := s.initPlayerMovement(time.Now()); err != nil {
		return false
	}
	if s.player != nil {
		s.player.SetSitting(false)
		s.player.SetMap(mapID)
		s.player.UpdateCharacter(*ch)
		s.player.UpdateMotion(0, 0)
	}
	if changedMap {
		s.noteDungeonExit(previousMap)
		s.refreshMapCurses(true)
	}
	s.savePosition()
	s.partyMove(changedMap)
	s.send(buildCharacterMapPosition(ch), "0x020B "+label)
	s.send(buildCharacterMove(ch, 0, 0), "0x0501 "+label+" move")
	if changedMap {
		s.awaitingMapLoad = true
		s.sendRelicList()
	}
	s.placeView()
	return true
}

// handleCharacterMove (0x0501): motion:u8, direction:u16, x, y, z:f32.
func (s *Session) handleCharacterMove(p []byte) {
	if s.dash != nil {
		return
	}
	if len(p) < 17 || s.char == nil {
		return
	}
	if s.player != nil && s.player.vehicleState().State == proto.VehiclePassenger {
		s.correctMovement("passenger cannot move independently")
		return
	}
	motion := p[2]
	direction, _ := packet.U16(p[3:])
	x, _ := packet.F32(p[5:])
	y, _ := packet.F32(p[9:])
	z, _ := packet.F32(p[13:])
	ch := s.char
	oldX, oldZ := ch.PosX, ch.PosZ
	if x != oldX || z != oldZ {
		s.pendingCombo = nil
		s.pendingSkill = nil
	}
	if x != ch.PosX || z != ch.PosZ {
		s.cancelMove()
		s.cancelVehicleCast()
	}
	ok, reason := s.validatePlayerMovement(x, y, z, motion, direction, time.Now())
	if !ok {
		s.log.Printf("GX Server movement rejected: %s", reason)
		s.correctMovement(reason)
		return
	}
	if reason == heightRecovery {
		return
	}
	// GX also sends idle and jump-height snapshots. Horizontal movement can
	// interrupt preparation even with the jump flag set.
	moved := ch.PosX != oldX || ch.PosZ != oldZ
	if proto.CastInterruptingMotion(motion) || moved {
		s.cancelTravelCast()
		if s.cast != nil {
			s.log.Printf("GX Server World cast interrupted by locomotion skill_number=%d motion=0x%02X from=(%.3f,%.3f) to=(%.3f,%.3f)", s.cast.number, motion, oldX, oldZ, ch.PosX, ch.PosZ)
			s.cast = nil
			s.syncPlayerSpeed(false)
		}
	}
	if moved {
		s.autoAttack = false
		s.endMyShop() // moving cancels the personal market
	}
	x, y, z = ch.PosX, ch.PosY, ch.PosZ
	motion = s.movement.motion
	if reason != "" {
		s.correctMovement(reason)
	}
	s.positionDirty = true
	if s.player != nil {
		s.player.UpdateCharacter(*ch)
		s.player.UpdateMotion(motion, direction)
	}
	s.partyMove()
	s.duelCheckMove()
	// A gated map (0x020A refusals) is left to the explicit portal request.
	if portal, ok := s.usablePortal(x, y, z, nil); ok && s.portalGate(portal.TargetMapID) == 0 {
		s.applyTeleport(portal.TargetMapID, portal.TargetX, portal.TargetY, portal.TargetZ, "portal teleport")
		return
	}
	// CZone::MoveUser: a cell change updates the view first, then the move
	// goes to the users that show this one.
	if s.player != nil && s.player.moveViewCell(x, z) {
		s.refreshView()
	}
	s.sendToViewers(buildCharacterMove(ch, motion, direction), "0x0501 character move (broadcast)")
	if time.Since(s.lastPositionSave) >= positionSaveInterval {
		s.savePosition()
	}
}

func (s *Session) handleEnteredPortal(index uint8) {
	if s.char == nil {
		return
	}
	if portal, ok := s.usablePortal(s.char.PosX, s.char.PosY, s.char.PosZ, &index); ok {
		// The request only counts while the character is in the portal: it can
		// not teleport from across the map, and a stale request sent just
		// before an automatic teleport finds no matching portal here.
		if !portal.Contains(s.char.PosX, s.char.PosY, s.char.PosZ) {
			return
		}
		if r := s.portalGate(portal.TargetMapID); r != 0 {
			s.send(buildPortalResult(index, r), "0x020A portal refused")
			return
		}
		s.applyTeleport(portal.TargetMapID, portal.TargetX, portal.TargetY, portal.TargetZ, "explicit portal teleport")
	}
}

func isTownCommand(message string) bool {
	fields := strings.Fields(message)
	return len(fields) == 1 && (strings.EqualFold(fields[0], "/town") || strings.EqualFold(fields[0], "town"))
}

// executeTown moves the character to the nearest faction spawn area on
// its map (also the escape hatch while dead).
func (s *Session) executeTown() {
	ch := s.char
	if area, ok := FindNearestSpawnArea(ch.MapID, s.faction, ch.PosX, ch.PosZ); ok {
		x, y, z := area.RandomPoint()
		s.applyTeleport(ch.MapID, x, y, z, "town teleport")
		s.send(buildChatNormal(ch.ID, fmt.Sprintf("Returned to nearest town spawn (%.2f, %.2f, %.2f).", ch.PosX, ch.PosY, ch.PosZ)),
			"0x1101 town command result")
		return
	}
	// Dungeons usually have no spawn area of their own: their configured
	// rebirth point lives on another map.
	if p, ok := rebirthPosition(ch.MapID, s.faction); ok && mapLevelAllowed(p.MapID, ch.Level, s.isAdmin) {
		if size, ok := mapSize(p.MapID); ok && p.X >= 0 && p.X < size && p.Z >= 0 && p.Z < size {
			s.applyTeleport(p.MapID, p.X, p.Y+0.9, p.Z, "town teleport")
			s.send(buildChatNormal(ch.ID, fmt.Sprintf("Returned to town (map %d: %.2f, %.2f, %.2f).", ch.MapID, ch.PosX, ch.PosY, ch.PosZ)),
				"0x1101 town command result")
			return
		}
	}
	s.log.Printf("GX Server World town command found no spawn area or rebirth point map=%d faction=%d", ch.MapID, s.faction)
	s.send(buildChatNormal(ch.ID, "No town spawn area was found for this map/faction."), "0x1101 town command result")
}

// deathExpPercent is CUser::Death's drop without protection (deathExpLoss).
const deathExpPercent = 5

// expDrop is CUser::ExpDrop (0x0045ad50): above level 10, percent of the
// level's exp span, never below the level's start.
func expDrop(ch *db.Character, percent int, faction uint8) uint32 {
	if ch.Level <= 10 || percent <= 0 {
		return 0
	}
	prev, next := expBounds(ch)
	loss := int64(math.RoundToEven(float64(float32(percent)) / 100 * float64(float32(int32(next-prev)))))
	// The country tier's type 25 cuts the loss by that percent.
	if cut := kcWorldType(faction, 25); cut != 0 {
		loss -= int64(math.RoundToEven(float64(float32(cut)) / 100 * float64(float32(int32(loss)))))
	}
	room := int64(ch.Exp) - int64(prev)
	if room < 1 {
		return 0
	}
	loss = min(loss, room)
	ch.Exp -= uint32(loss)
	return uint32(loss)
}

// returnToTownDelay is 0x0556's tick + 10000 (CUser::PacketPC).
const returnToTownDelay = 10 * time.Second

// handleReturnToTown (0x0556): alive and with no move pending, the user
// is moved to its bind point 10 s later (UpdateResetPosition, type 0).
func (s *Session) handleReturnToTown() {
	if s.char == nil || s.dead || !s.returnAt.IsZero() {
		return
	}
	s.returnAt = time.Now().Add(returnToTownDelay)
}

// cancelMove is CUser::CancleMove for the type-0 move: moving, attacking,
// casting or being hit drops it without a reply.
func (s *Session) cancelMove() { s.returnAt = time.Time{} }

func (s *Session) tickReturnToTown() {
	if s.returnAt.IsZero() || time.Now().Before(s.returnAt) || s.char == nil {
		return
	}
	s.returnAt = time.Time{}
	if s.dead {
		return
	}
	m, x, y, z := s.respawnDestination(s.char)
	s.applyTeleport(m, x, y, z, "return to town")
}
