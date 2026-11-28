package world

import (
	"fmt"

	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
)

func logActivity(log logPrinter, kind, message string, fields ...any) {
	if logger, ok := log.(interface{ Event(string, ...any) }); ok {
		logger.Event(message, append([]any{"activity", kind}, fields...)...)
	}
}

func (s *Session) activity(kind, message string, fields ...any) {
	if s.log == nil || s.char == nil {
		return
	}
	ch := s.char
	if kind == "kill.pvp" {
		activityForMap(realMap(ch.MapID)).pvp.Add(1)
	}
	fields = append([]any{"account_id", s.accountID, "character_id", ch.ID, "character", ch.Name, "map_id", realMap(ch.MapID)}, fields...)
	logActivity(s.log, kind, message, fields...)
}

func (s *Session) goldActivity(before uint32, source string) {
	if gained := int64(s.char.Money) - int64(before); gained > 0 {
		activityForMap(realMap(s.char.MapID)).gold.Add(uint64(gained))
		s.activity("loot.gold", fmt.Sprintf("%s gained %d gold (%s)", s.char.Name, gained, source),
			"gold", gained, "balance", s.char.Money, "source", source)
	}
}

func itemActivityName(typ, id uint8) string {
	if def := db.MustDefs().Item(typ, id); def != nil {
		return def.Name
	}
	return fmt.Sprintf("item %d/%d", typ, id)
}
