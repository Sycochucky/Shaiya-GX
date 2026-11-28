package world

import "github.com/Spelunkern/Shaiya-GX/server/internal/db"

func (s *Session) canEnterMap(mapID, level uint16) bool {
	if s.char != nil && realMap(s.char.MapID) == realMap(mapID) {
		return mapLevelAllowed(mapID, level, s.isAdmin)
	}
	return s.mapAdmissionAllowed(mapID, level)
}

func mapLevelAllowed(mapID, level uint16, gm bool) bool {
	zone := mapConfig.Zones[realMap(mapID)]
	return gm || (zone.MinLevel == 0 || level >= zone.MinLevel) &&
		(zone.MaxLevel == 0 || level <= zone.MaxLevel)
}

func birthPosition(ch *db.Character) MapPos {
	return mapConfig.Start[min(int(ch.Race), len(mapConfig.Start)-1)]
}

func (s *Session) respawnDestination(ch *db.Character) (uint16, float32, float32, float32) {
	return characterRespawn(ch, s.faction, s.isAdmin)
}

func characterRespawn(ch *db.Character, faction uint8, gm bool) (uint16, float32, float32, float32) {
	if m, x, y, z, ok := bindPositionFound(ch.MapID, faction, ch.PosX, ch.PosY, ch.PosZ); ok && mapLevelAllowed(m, ch.Level, gm) {
		return m, x, y, z
	}
	start := birthPosition(ch)
	return start.MapID, start.X, start.Y, start.Z
}

func (s *Session) enforceMapLevel() {
	if s.char == nil || s.canEnterMap(s.char.MapID, s.char.Level) {
		return
	}
	p := birthPosition(s.char)
	s.applyTeleport(p.MapID, p.X, p.Y, p.Z, "map level limit")
}
