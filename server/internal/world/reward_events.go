package world

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

var currentRewardEvent atomic.Pointer[config.TimedRewardEvent]

func SaveRewardEvent(req operatorwire.Request, log *logx.Logger) error {
	if req.RewardEvent == nil {
		return errors.New("missing reward event settings")
	}
	rewardSettingsMu.Lock()
	defer rewardSettingsMu.Unlock()
	state, err := config.SaveRewardEvent(".", *req.RewardEvent, time.Now(), func(id uint16) bool {
		_, ok := mapConfig.Zones[id]
		return ok
	})
	if err != nil {
		return err
	}
	if req.RewardEvent.Preset == nil {
		currentRewardEvent.Store(state.Scheduled)
	}
	log.Event("Updated reward event settings", "activity", "staff.settings")
	return nil
}

// The schedule persists across process starts. Rates are calculated from the
// time range on each reward, independent of the notice polling interval.
func pollRewardEvent(current **config.TimedRewardEvent, now time.Time, log *logx.Logger) {
	next := currentRewardEvent.Load()
	if !next.Running(now) {
		next = nil
	}
	if *current == next {
		return
	}
	if *current != nil {
		notifyRewardEvent((*current).Preset, " ended.", log)
	}
	if next != nil {
		notifyRewardEvent(next.Preset, " started.", log)
	}
	*current = next
}

func notifyRewardEvent(preset config.RewardPreset, suffix string, log *logx.Logger) {
	message := "Reward event " + preset.Name + suffix
	log.Event(message, "activity", "staff.settings")
	if preset.Announce {
		sendRewardNotice(preset, message)
	}
}

func sendRewardNotice(preset config.RewardPreset, message string) {
	if !validNoticeMessage(message) {
		return
	}
	w := packet.New(packet.Opcode(proto.GXServerNotice))
	writeChatText(w, message)
	frame := w.MustFinish()
	for _, player := range AllPlayers() {
		if preset.Affects(realMap(player.MapID())) {
			player.Send(frame, "reward event notice")
		}
	}
}
