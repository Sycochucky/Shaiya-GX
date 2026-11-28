package app

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
)

// Definitions are loaded on first quest interaction, not for every NPC draw.
func (a *App) loadQuests() error {
	if a.questDataLoaded {
		if a.questDefs == nil {
			return fmt.Errorf("quest definitions unavailable")
		}
		return nil
	}
	a.questDataLoaded = true
	raw, err := a.arc.ReadFile("Npc/NpcQuest.SData")
	if err != nil {
		return err
	}
	npcs, quests, links, err := sdata.ParseQuestCatalog(raw)
	if err != nil {
		return err
	}
	a.npcDefs, a.questDefs, a.questItemLinks = npcs, quests, links
	return nil
}

// questMessageID adds the SysMsg text to the chat as a plain line.
func (a *App) questMessageID(id int) { a.questMessage(a.tr(id)) }

func (a *App) questMessage(message string) {
	a.state.AddChat(game.ChatLine{Text: message})
}
func (a *App) questName(id uint16) string {
	if id == 0 || int(id) > len(a.questDefs) {
		return a.msg.Format(12314, map[string]string{"v": fmt.Sprint(id)})
	}
	return a.questDefs[id-1].Name
}

func (a *App) confirmedQuestProgress(id uint16, slot uint8) {
	active, ok := a.state.Quests[id]
	if !ok || int(slot) >= len(active.Counters) {
		return
	}
	if !a.questDataLoaded && a.arc != nil {
		_ = a.loadQuests()
	}
	if id == 0 || int(id) > len(a.questDefs) {
		return
	}

	def := &a.questDefs[id-1]
	switch def.EndType {
	case 1:
		a.questMessage(fmt.Sprintf("%s: objective %d = %d", def.Name, slot+1, active.Counters[slot]))
	case 2:
		if int(slot) >= len(def.MobIDs) || def.MobCounts[slot] == 0 {
			return
		}
		// EP5.4 reports all configured mob objectives after a counter update.
		for i, mobID := range def.MobIDs {
			required := def.MobCounts[i]
			if mobID == 0 || required == 0 {
				continue
			}
			if a.assets != nil {
				if mob, err := a.assets.MonsterDefinition(mobID); err == nil && mob.Name != "" {
					a.questKillMessage(def, 351, mob.Name, active.Counters[i], required)
					continue
				}
			}
			if i == int(slot) {
				a.questMessage(fmt.Sprintf("%s: objective %d = %d", def.Name, slot+1, active.Counters[slot]))
			}
		}
	case 3:
		if slot == 0 && def.PVPKills != 0 {
			a.questKillMessage(def, 349, "", active.Counters[0], def.PVPKills)
		}
	}
}

func (a *App) questKillMessage(def *sdata.Quest, message int, target string, current, required uint8) {
	values := map[string]string{"p": def.Name, "t": target}
	// The native 0x0905 callback at 0x5745F0 selects the capped variant for either flag.
	if def.ResultType != 0 || def.UserSelect != 0 {
		message++
		current = min(current, required)
		values["v2"] = strconv.Itoa(int(required))
	}
	values["v"] = strconv.Itoa(int(current))
	a.systemMessage(6, message, values)
}

// Presentation follows confirmed server messages, including automatic progress.
func (a *App) confirmedQuest(p []byte) {
	switch proto.Op(p) {
	case proto.SelectCharacter:
		if a.state.Phase == game.PhaseLoading {
			a.questStartedAt = nil
			a.questSnapshotAt = nil
		}
	case proto.QuestList:
		if quests, err := proto.DecodeQuestList(p); err == nil {
			if a.questSnapshotAt == nil {
				a.questSnapshotAt = make(map[uint16]time.Time)
			}
			at := time.Now()
			for _, q := range quests {
				a.questSnapshotAt[q.ID] = at
			}
		}
	case proto.QuestStart:
		_, id, err := proto.DecodeQuestStart(p)
		if err == nil {
			if a.questStartedAt == nil {
				a.questStartedAt = make(map[uint16]time.Time)
			}
			a.questStartedAt[id] = time.Now()
			delete(a.questSnapshotAt, id)
			if a.msg.Get(219) != "" {
				a.systemMessage(3, 219, nil)
			} else {
				a.questMessage("Quest started: " + a.questName(id))
			}
		}
	case proto.AddItem:
		if item, err := proto.DecodeItemAdd(p); err == nil && isQuestInventoryItem(item.Type) {
			a.offerItemQuest([2]uint8{item.Bag, item.Slot}, sdata.ItemKey{Type: item.Type, TypeID: item.TypeID}, 3)
		}
	case proto.QuestCounter:
		id, slot, _, err := proto.DecodeQuestCounter(p)
		if err == nil {
			a.confirmedQuestProgress(id, slot)
		}
	case proto.QuestChoose:
		id, err := proto.DecodeQuestChoice(p)
		if err == nil {
			a.questMessage(a.msg.Format(12315, map[string]string{"v": fmt.Sprint(id)}))
		}
	case proto.QuestEnd:
		result, err := proto.DecodeQuestEnd(p)
		if err == nil {
			if result.Success != 0 && a.wv != nil && a.assets != nil {
				a.startSkillImpact(a.state.SelfID, "Effect/up.EFT", 1)
				a.playSound("bgm_quest02.wav", a.wv.pos[0], a.wv.pos[2])
			}
			delete(a.questTracker.expanded, result.ID)
			delete(a.questStartedAt, result.ID)
			delete(a.questSnapshotAt, result.ID)
			// A timed or automatic quest can end before the quest UI has ever
			// loaded definitions. Its repeatability still determines whether
			// Rapture records it as finished.
			if !a.questDataLoaded && a.arc != nil {
				_ = a.loadQuests()
			}
			if result.ID > 0 && int(result.ID) <= len(a.questDefs) {
				if a.state.FinishedQuests == nil {
					a.state.FinishedQuests = make(map[uint16]uint8)
				}
				if a.questDefs[result.ID-1].TickReceive == 0 {
					a.state.FinishedQuests[result.ID] = result.Success
				} else {
					delete(a.state.FinishedQuests, result.ID)
				}
			}
			dialogue := result.Success != 0 && a.shop.npcID != 0 && a.npcQuest.selected == result.ID && a.npcQuest.ending
			if dialogue {
				copy := result
				a.npcQuest.result = &copy
				if a.npcQuestUI != nil {
					a.npcQuestUI.ResetDescriptionScroll()
				}
			}
			status := "ended"
			if result.Success != 0 {
				status = "completed"
			}
			if a.msg.Get(243) != "" {
				id := 244
				if result.Success != 0 {
					id = 243
				}
				a.systemMessage(3, id, nil)
			} else {
				a.questMessage(a.msg.Format(12316, map[string]string{"v": status, "v2": a.questName(result.ID)}))
			}
			if result.Success != 0 && !dialogue {
				var def sdata.Quest
				if result.ID > 0 && int(result.ID) <= len(a.questDefs) {
					def = a.questDefs[result.ID-1]
				}
				a.questMessage(a.questResultText(def, &result))
			} else if result.Success == 0 && result.ID > 0 && int(result.ID) <= len(a.questDefs) {
				if text := questText(a.questDefs[result.ID-1].Descriptions[3]); text != "" {
					a.questMessage(text)
				}
			}
		}
	}
}

// Commands provide access while the native quest dialogue is being built.
// Requests never change quest state until the server acknowledges them.
func (a *App) questCommand(text string) bool {
	fields := strings.Fields(text)
	if len(fields) == 0 || !strings.EqualFold(fields[0], "/quest") {
		return false
	}
	if err := a.loadQuests(); err != nil {
		a.questMessageID(12150)
		return true
	}
	if len(fields) == 1 || fields[1] == "list" {
		ids := make([]int, 0, len(a.state.Quests))
		for id := range a.state.Quests {
			ids = append(ids, int(id))
		}
		sort.Ints(ids)
		for _, number := range ids {
			id := uint16(number)
			q := a.state.Quests[id]
			a.questMessage(fmt.Sprintf("%d: %s (%d/%d/%d)", id, a.questName(id), q.Counters[0], q.Counters[1], q.Counters[2]))
		}
		if npc := a.state.Entities[a.shop.npcID]; npc != nil {
			for _, entry := range a.npcQuestEntries(npc) {
				if entry.ending {
					a.questMessage(a.msg.Format(12317, map[string]string{"v": fmt.Sprint(entry.id), "v2": a.questName(entry.id)}))
				} else {
					a.questMessage(a.msg.Format(12318, map[string]string{"v": fmt.Sprint(entry.id), "v2": a.questName(entry.id)}))
				}
			}
		}
		a.questMessageID(12175)
		return true
	}
	if len(fields) < 3 {
		a.questMessageID(12176)
		return true
	}
	parsed, err := strconv.ParseUint(fields[2], 10, 16)
	if err != nil || parsed == 0 {
		a.questMessageID(12177)
		return true
	}
	id := uint16(parsed)
	var packet []byte
	switch strings.ToLower(fields[1]) {
	case "start", "end":
		if strings.EqualFold(fields[1], "start") {
			if !a.questDataLoaded && a.arc != nil {
				_ = a.loadQuests()
			}
			if int(id) <= len(a.questDefs) && a.questDefs[id-1].StartType == 3 {
				packet = proto.QuestStartPacket(uint32(a.questDefs[id-1].StartItemID), id)
				break
			}
		} else if int(id) <= len(a.questDefs) {
			q := a.questDefs[id-1]
			if q.EndType >= 1 && q.EndType <= 3 && q.EndNPCType == 0 {
				packet = proto.QuestEndPacket(0, id)
				break
			}
		}
		npc := a.state.Entities[a.shop.npcID]
		if npc == nil || npc.Kind != game.KindNpc {
			a.questMessageID(12178)
			return true
		}
		if strings.EqualFold(fields[1], "start") {
			packet = proto.QuestStartPacket(npc.ID, id)
		} else {
			packet = proto.QuestEndPacket(npc.ID, id)
		}
	case "abandon":
		packet = proto.QuestAbandonPacket(id)
	case "choose":
		if len(fields) != 4 {
			a.questMessageID(12179)
			return true
		}
		choice, e := strconv.ParseUint(fields[3], 10, 8)
		if e != nil || choice > 2 || !a.state.QuestChoosing || a.state.QuestChoice != id {
			a.questMessageID(12180)
			return true
		}
		packet = proto.QuestChoosePacket(id, uint8(choice))
	default:
		a.questMessageID(12176)
		return true
	}
	if err := a.world.Send(packet); err != nil {
		a.backToLoginID(12149)
	}
	return true
}
