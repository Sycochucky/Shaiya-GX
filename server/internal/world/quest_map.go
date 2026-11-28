package world

import (
	"maps"
	"slices"

	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

type questNPCCatalogCache struct {
	ready    bool
	revision uint64
	rows     []proto.QuestNPC
}
type questNPCCatalogState struct {
	sent     bool
	mapID    uint16
	revision uint64
	rows     []proto.QuestNPC
}

// Shared immutable map metadata. Patrols use their authored first waypoint;
// ordinary movement neither invalidates this cache nor streams global positions.
func (sh *mapShard) questNPCSnapshot() (uint64, []proto.QuestNPC) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	revision := sh.npcRevision.Load()
	c := &sh.questNPCs
	if !c.ready || c.revision != revision {
		var rows []proto.QuestNPC
		for _, n := range sh.entities.NPCs {
			def := findNpcDef(n.Type, n.TypeID)
			if def == nil || len(def.StartQuests)+len(def.EndQuests) == 0 {
				continue
			}
			x, z := n.PosX, n.PosZ
			if len(n.patrol) > 0 {
				x, z = n.patrol[0].X, n.patrol[0].Z
			}
			rows = append(rows, proto.QuestNPC{Type: n.Type, TypeID: n.TypeID, X: x, Z: z})
		}
		if !c.ready || !slices.Equal(c.rows, rows) {
			c.rows = rows
		}
		c.ready, c.revision = true, revision
	}
	return revision, c.rows
}

func questNPCFrames(mapID uint16, rows []proto.QuestNPC) [][]byte {
	// length/opcode 4 + map/flags/count 5 + 11 bytes per NPC.
	const batch = proto.QuestNPCChunkLimit
	var frames [][]byte
	for start := 0; ; {
		end := min(start+batch, len(rows))
		flags := uint8(0)
		if start == 0 {
			flags |= proto.QuestNPCBegin
		}
		if end == len(rows) {
			flags |= proto.QuestNPCEnd
		}
		w := packet.New(packet.Opcode(proto.GXQuestNPCs))
		w.U16(mapID)
		w.U8(flags)
		w.U16(uint16(end - start))
		for _, n := range rows[start:end] {
			w.U8(n.Type)
			w.U16(n.TypeID)
			w.F32(n.X)
			w.F32(n.Z)
		}
		frames = append(frames, w.MustFinish())
		if end == len(rows) {
			return frames
		}
		start = end
	}
}

func (s *Session) syncQuestNPCCatalog(force bool) {
	if s.char == nil || s.player == nil {
		return
	}
	sh := shard(s.char.MapID)
	c := &s.questNPCCatalog
	if !force && c.sent && c.mapID == s.char.MapID && c.revision == sh.npcRevision.Load() {
		return
	}
	revision, rows := sh.questNPCSnapshot()
	changed := force || !c.sent || c.mapID != s.char.MapID || !slices.Equal(c.rows, rows)
	c.sent, c.mapID, c.revision, c.rows = true, s.char.MapID, revision, rows
	if changed {
		for _, frame := range questNPCFrames(realMap(s.char.MapID), rows) {
			s.send(frame, "GX map quest NPC catalog")
		}
	}
}

// Quest map markers describe spawn sites. Ordinary movement and deaths leave
// the catalog intact so the full map stays useful while monsters respawn.
func (sh *mapShard) invalidateQuestMobCatalog() {
	sh.questMobCatalog = nil
	sh.questMobRevision.Add(1)
}

// setQuestEventSpawn is called with the shard locked.
func (sh *mapShard) setQuestEventSpawn(m *MapMob, present bool) {
	if present {
		row := proto.QuestMobSpawn{MobID: m.MobID, X: m.homeX, Z: m.homeZ}
		if sh.questEventSpawns[m.ID] == row {
			return
		}
		if sh.questEventSpawns == nil {
			sh.questEventSpawns = make(map[uint32]proto.QuestMobSpawn)
		}
		sh.questEventSpawns[m.ID] = row
	} else {
		if _, ok := sh.questEventSpawns[m.ID]; !ok {
			return
		}
		delete(sh.questEventSpawns, m.ID)
	}
	sh.invalidateQuestMobCatalog()
}

func (sh *mapShard) questMobFrames(mapID uint16) (uint64, [][]byte) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	revision := sh.questMobRevision.Load()
	if sh.questMobCatalog != nil {
		return revision, sh.questMobCatalog
	}
	catalog := sh.questMobSpawns
	if len(sh.questEventSpawns) > 0 {
		catalog = slices.Clone(catalog)
		for _, id := range slices.Sorted(maps.Keys(sh.questEventSpawns)) {
			catalog = append(catalog, sh.questEventSpawns[id])
		}
	}
	for start := 0; ; {
		end := min(start+proto.QuestMobChunkLimit, len(catalog))
		var flags uint8
		if start == 0 {
			flags |= proto.QuestNPCBegin
		}
		if end == len(catalog) {
			flags |= proto.QuestNPCEnd
		}
		w := packet.New(packet.Opcode(proto.GXQuestMobs))
		w.U16(mapID)
		w.U8(flags)
		w.U16(uint16(end - start))
		for _, mob := range catalog[start:end] {
			w.U16(mob.MobID)
			w.F32(mob.X)
			w.F32(mob.Z)
		}
		sh.questMobCatalog = append(sh.questMobCatalog, w.MustFinish())
		if end == len(catalog) {
			return revision, sh.questMobCatalog
		}
		start = end
	}
}

type questMobCatalogState struct {
	mapID    uint16
	revision uint64
}

func (s *Session) syncQuestMobCatalog(force bool) {
	if s.char == nil || s.player == nil || s.awaitingMapLoad && !force {
		return
	}
	mapID := s.char.MapID
	sh := shard(mapID)
	if !force && s.questMobCatalog == (questMobCatalogState{mapID: mapID, revision: sh.questMobRevision.Load()}) {
		return
	}
	revision, frames := sh.questMobFrames(realMap(mapID))
	s.questMobCatalog = questMobCatalogState{mapID: mapID, revision: revision}
	for _, frame := range frames {
		s.send(frame, "GX map quest mob spawn catalog")
	}
}
