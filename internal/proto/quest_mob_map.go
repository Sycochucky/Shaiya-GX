package proto

import (
	"fmt"
	"math"
)

// GXQuestMobs supplies monster spawn locations without creating visible entities.
const GXQuestMobs Opcode = 0x7F2C

// Keep frames below the transport's uint16 length limit.
const QuestMobChunkLimit = (65535 - 9) / 10

type QuestMobSpawn struct {
	MobID uint16
	X, Z  float32
}

func DecodeQuestMobs(p []byte) (mapID uint16, flags uint8, rows []QuestMobSpawn, err error) {
	r := NewReader(p)
	mapID, flags = r.U16(), r.U8()
	count := int(r.U16())
	if r.Err() != nil || count > QuestMobChunkLimit || flags & ^uint8(QuestNPCBegin|QuestNPCEnd) != 0 || r.Remaining() != count*10 {
		return 0, 0, nil, fmt.Errorf("invalid quest mob catalog")
	}
	rows = make([]QuestMobSpawn, count)
	for i := range rows {
		row := QuestMobSpawn{MobID: r.U16(), X: r.F32(), Z: r.F32()}
		if row.MobID == 0 || math.IsNaN(float64(row.X)) || math.IsNaN(float64(row.Z)) || math.IsInf(float64(row.X), 0) || math.IsInf(float64(row.Z), 0) {
			return 0, 0, nil, fmt.Errorf("invalid quest mob position")
		}
		rows[i] = row
	}
	return mapID, flags, rows, r.Err()
}
