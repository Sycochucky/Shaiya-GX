package world

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/worlddrops"
	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

var scopedDropRules atomic.Pointer[[]config.ScopedDropRule]
var dropAnnouncementMu sync.Mutex
var nextDropAnnouncement time.Time

func SaveScopedDrop(req operatorwire.Request, log *logx.Logger) error {
	if req.ScopedDrop == nil {
		return fmt.Errorf("missing drop rule settings")
	}
	edit := *req.ScopedDrop
	rewardSettingsMu.Lock()
	defer rewardSettingsMu.Unlock()
	if edit.Rule != nil {
		if err := validateScopedDrop(*edit.Rule); err != nil {
			return err
		}
		if edit.Rule.DailyLimit > 0 {
			if err := db.EnsureEventDropLimits(); err != nil {
				return fmt.Errorf("prepare event drop caps: %w", err)
			}
		}
	}
	rules, err := config.SaveScopedDrop(".", edit)
	if err != nil {
		return err
	}
	scopedDropRules.Store(&rules)
	log.Event("Updated event drop rules", "activity", "staff.settings")
	return nil
}

func validateScopedDrop(rule config.ScopedDropRule) error {
	if err := rule.Validate(); err != nil {
		return err
	}
	for _, mapID := range rule.Maps {
		if _, exists := mapConfig.Zones[mapID]; !exists {
			return fmt.Errorf("map %d is unavailable", mapID)
		}
	}
	if rule.ItemType != 0 {
		item := db.MustDefs().Item(rule.ItemType, rule.ItemID)
		if item == nil || item.Grade == 0 || item.Grade > 1024 || item.TypeID == 0 || isQuestItemType(item.Type) {
			return fmt.Errorf("item %d/%d is unavailable in the drop catalog", rule.ItemType, rule.ItemID)
		}
	}
	return nil
}

func rollScopedDrops(zone uint16, mob *MapMob, ownerID uint32, percent uint32, place func(*db.ItemDef, uint32), log logPrinter) {
	rules := scopedDropRules.Load()
	if rules == nil || len(*rules) == 0 || percent == 0 {
		return
	}
	id := realMap(zone)
	now := time.Now()
	for _, rule := range *rules {
		if !rule.Eligible(id, mob.Level) || !rollScaledDrop(rule.Chance, worlddrops.RollRange, percent, rand.Uint64N) {
			continue
		}
		var item *db.ItemDef
		if rule.Grade != 0 {
			item = pickGradeItem(rule.Grade)
		} else {
			if def := db.MustDefs().Item(rule.ItemType, rule.ItemID); def != nil {
				item = pickExactDrop(def)
			}
		}
		if item == nil || !allowedScopedDrop(rule, ownerID, now, log) {
			continue
		}
		place(item, 0)
		logActivity(log, "loot.drop", fmt.Sprintf("%s dropped (event: %s)", item.Name, rule.Name),
			"map_id", id, "character_id", ownerID, "mob_id", mob.MobID, "source", "scoped_event",
			"rule", rule.ID, "item_type", item.Type, "item_id", item.TypeID)
		if rule.Announce {
			announceScopedDrop(id, item.Name, now)
		}
	}
}

func allowedScopedDrop(rule config.ScopedDropRule, ownerID uint32, now time.Time, log logPrinter) bool {
	if rule.DailyLimit == 0 {
		return true
	}
	player := FindPlayerByID(ownerID)
	if player == nil || player.owner == nil {
		return false
	}
	allowed, err := db.ClaimEventDrop(rule.ID, player.owner.accountID, rule.DailyLimit, now)
	if err != nil {
		log.Printf("event drop rule %d daily cap could not be checked: %v", rule.ID, err)
	}
	return err == nil && allowed
}

func announceScopedDrop(mapID uint16, name string, now time.Time) {
	dropAnnouncementMu.Lock()
	if now.Before(nextDropAnnouncement) {
		dropAnnouncementMu.Unlock()
		return
	}
	nextDropAnnouncement = now.Add(30 * time.Second)
	dropAnnouncementMu.Unlock()
	sendRewardNotice(config.RewardPreset{Maps: []uint16{mapID}}, name+" dropped in this map.")
}
