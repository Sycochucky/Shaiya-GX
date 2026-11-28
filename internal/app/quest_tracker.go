package app

import (
	"fmt"
	"math"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/npcquest"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

// questTrackerView mirrors EP5.4 QUESTMINI. Like the per-quest byte at +8
// of the original active-quest record, expansion outlives the window and is
// dropped only when the quest ends or the world is released.
type questTrackerView struct {
	open, inventoryReady             bool
	expanded                         map[uint16]bool
	cachedState                      *game.State
	questRevision, inventoryRevision uint64
	refreshAt                        time.Time
	entries                          []ui.QuestTrackerEntry
}

// showQuestTracker follows 0x4ffb00: a journal row press creates the window
// only when it is absent, anchors it to the journal and toggles that quest
// (0x4dc340). An open window is left untouched.
func (a *App) showQuestTracker(id uint16, w, h int) {
	v := &a.questTracker
	if a.questTrackerUI == nil || v.open {
		return
	}
	v.open = true
	if v.expanded == nil {
		v.expanded = map[uint16]bool{}
	}
	v.expanded[id] = !v.expanded[id]
	v.cachedState = nil
	if a.questUI != nil {
		jx, jy := a.questUI.Origin(w, h)
		a.questTrackerUI.Anchor(jx, jy, w)
	}
	a.refreshQuestTracker()
}

func (a *App) closeQuestTracker() {
	a.questTracker.open = false
	a.questTracker.cachedState = nil
	if a.questTrackerUI != nil {
		a.questTrackerUI.Reset()
	}
}

func (a *App) resetQuestTracker() {
	a.closeQuestTracker()
	a.questTracker = questTrackerView{}
}

// questMinutesLeft is the "%5d" value of 0x4dc9b0 for limited quests.
func (a *App) questMinutesLeft(def sdata.Quest, q proto.ActiveQuest) (int, bool) {
	limit := npcquest.Deadline(def.MaxMinutes, def.TickKeep)
	if limit == 0 {
		return 0, false
	}
	var seconds float64
	if at, ok := a.questStartedAt[q.ID]; ok {
		seconds = limit.Seconds() - time.Since(at).Seconds()
	} else if at, ok := a.questSnapshotAt[q.ID]; ok {
		seconds = float64(uint32(q.MinutesLeft)+1)*60 - time.Since(at).Seconds()
	} else {
		return int(q.MinutesLeft), true
	}
	return max(0, int(math.Ceil(seconds/60))), true
}

// questTrackerLines reproduces the objective rows of 0x4dc9b0.
func (a *App) questTrackerLines(def sdata.Quest, q proto.ActiveQuest) []string {
	var lines []string
	showGoal := def.ResultType != 0 || def.UserSelect != 0 // client quest bytes +0x64/+0x65
	count := func(label string, have, need uint8) string {
		if !showGoal {
			return fmt.Sprintf("%s : %d", label, have)
		}
		return fmt.Sprintf("%s : %d/%d", label, min(have, need), need)
	}
	if def.EndType == 3 {
		if def.PVPKills != 0 {
			lines = append(lines, count(ui.Tr(438), q.Counters[0], def.PVPKills))
		}
	} else {
		for i, id := range def.MobIDs {
			if def.MobCounts[i] == 0 {
				continue
			}
			// 0x4dc9b0 still advances the row when 0x460900 finds no monster.
			line := ""
			if a.assets != nil {
				if m, err := a.assets.MonsterDefinition(id); err == nil && m.Name != "" {
					line = count(m.Name+" "+ui.Tr(439), q.Counters[i], def.MobCounts[i])
				}
			}
			lines = append(lines, line)
		}
	}
	for _, item := range def.FarmItems {
		if item.Count == 0 {
			continue
		}
		d := a.itemsItem(item.Type, item.ID)
		if d == nil {
			continue // 0x4607b0 returned no item: no row
		}
		have := 0
		for key, it := range a.state.Inventory {
			if key[0] >= 1 && key[0] <= 5 && it.Type == item.Type && it.TypeID == item.ID {
				have += int(it.Count)
			}
		}
		lines = append(lines, fmt.Sprintf("%s %d/%d", d.Name, have, item.Count))
	}
	return lines
}

func (a *App) refreshQuestTracker() {
	v, s := &a.questTracker, a.state
	now := time.Now()
	if v.cachedState == s && v.questRevision == s.QuestMapRevision && v.inventoryRevision == s.QuestInventoryRevision &&
		v.inventoryReady == s.InventoryReady && (v.refreshAt.IsZero() || now.Before(v.refreshAt)) {
		return
	}
	v.cachedState, v.questRevision, v.inventoryRevision = s, s.QuestMapRevision, s.QuestInventoryRevision
	v.inventoryReady = s.InventoryReady
	v.refreshAt = time.Time{}
	v.entries = v.entries[:0]
	for _, id := range a.questIDs() {
		if id == 0 || int(id) > len(a.questDefs) {
			continue
		}
		def, q := a.questDefs[id-1], s.Quests[id]
		entry := ui.QuestTrackerEntry{ID: id, Title: a.questName(id), Expanded: v.expanded[id], Ready: a.questObjectivesReady(def, q)}
		if minutes, ok := a.questMinutesLeft(def, q); ok {
			entry.Time = fmt.Sprintf("%5d%s", minutes, ui.Tr(440))
			v.refreshAt = now.Add(time.Second)
		}
		if entry.Expanded {
			entry.Lines = a.questTrackerLines(def, q)
		}
		v.entries = append(v.entries, entry)
	}
	a.questTrackerUI.SetEntries(v.entries)
}

func (a *App) updateQuestTracker(in *platform.Input, w, h int) bool {
	q := a.questTrackerUI
	if q == nil || !a.questTracker.open {
		return false
	}
	a.refreshQuestTracker()
	id, close, captured := q.Update(in, w, h)
	if close {
		a.closeQuestTracker()
		a.playUISound("ui_window.ogg")
		return true
	}
	if id != 0 {
		v := &a.questTracker
		v.expanded[id] = !v.expanded[id]
		v.cachedState = nil
		a.refreshQuestTracker()
		a.playUI() // 0x4dc540 plays sound 0x1e, registered as ui_button001 in 0x40b2d0
		return true
	}
	x, y := q.Origin(w, h)
	return captured || panelPointer(in, x, y, ui.QuestTrackerWidth, ui.QuestTrackerHeight)
}

func (a *App) drawQuestTracker(w, h int) {
	q := a.questTrackerUI
	if q == nil || !a.questTracker.open {
		return
	}
	a.refreshQuestTracker()
	x, y := q.Origin(w, h)
	a.clearTooltipUnderPanel(x, y, ui.QuestTrackerWidth, ui.QuestTrackerHeight)
	q.Draw(a.sprites, a.font, w, h)
}
