package app

import (
	"strings"
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/data/text"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

func TestEquipRefusalUsesNativeUpperChat(t *testing.T) {
	a := inputApp()
	a.msg = text.SysMsg{229: "Cannot equip <i>"}
	a.items = map[sdata.ItemKey]*sdata.Item{{Type: 16, TypeID: 2}: {Name: "Test Helmet"}}
	w := proto.NewWriter(proto.GXEquipRefusal)
	w.U8(16)
	w.U8(2)
	a.applyWorldPacket(w.Bytes())
	if len(a.state.Chat) != 1 || len(a.screenNotices) != 0 {
		t.Fatalf("equipment refusal must appear only in chat: %v, notices %v", a.state.Chat, a.screenNotices)
	}
	line := a.state.Chat[0]
	if line.Text != "Cannot equip Test Helmet" || line.MessageID != 229 || line.Style != 17 {
		t.Fatalf("wrong equipment notice: %+v", line)
	}
}

func TestAcknowledgedEventsUseRuntimeMessagesAndNativePalette(t *testing.T) {
	a := inputApp()
	a.state.SelfID = 1
	a.state.Entities[2] = &game.Entity{ID: 2, Kind: game.KindUser, Shape: &proto.UserShape{Name: "Opponent"}}
	a.msg = text.SysMsg{50: "XP <b>", 53: "Gold <b>", 1002: "Critical <t> <v>", 1003: "Received critical <t> <v>", 1004: "Miss <t>", 347: "No such character"}
	xp := proto.NewWriter(proto.ExpGain)
	xp.U32(1234)
	xp.U32(0)
	if _, err := a.state.Apply(xp.Bytes()); err != nil {
		t.Fatal(err)
	}
	a.confirmedSystemChat(xp.Bytes(), 0)
	if got := a.state.Chat[0]; got.Text != "XP 1234" || got.MessageID != 50 || got.Style != 3 {
		t.Fatal(got)
	}
	for _, tc := range []struct {
		source, target uint32
		result         uint8
		want           string
		rgb            uint32
	}{{1, 2, 1, "Critical Opponent 73", 0xff9b32}, {2, 1, 1, "Received critical Opponent 73", 0xff0000}, {1, 2, 3, "Miss Opponent", 0xff9b32}} {
		a.state.LastAttack = proto.AttackResult{Source: tc.source, Target: tc.target, Result: tc.result, HP: 73}
		a.confirmedSystemChat(proto.NewWriter(proto.CharacterUserAutoAttack).Bytes(), 0)
		got := chatEntry(a.msg, a.state.Chat[len(a.state.Chat)-1])
		if strings.NewReplacer("{c3}", "", "{/c}", "").Replace(got.Text) != tc.want || got.Color != ui.RGBColor(tc.rgb) || !got.System {
			t.Fatal(got)
		}
	}
	a.state.Details.Money = 100
	before := len(a.state.Chat)
	a.confirmedSystemChat(proto.NewWriter(proto.SetMoney).Bytes(), 100)
	a.confirmedSystemChat(proto.NewWriter(proto.SetMoney).Bytes(), 110)
	if len(a.state.Chat) != before {
		t.Fatal("balance snapshot or spending invented acquired gold")
	}
	a.confirmedSystemChat(proto.NewWriter(proto.SetMoney).Bytes(), 90)
	if a.state.Chat[len(a.state.Chat)-1].Text != "Gold 10" {
		t.Fatal(a.state.Chat)
	}
	got := chatEntry(a.msg, game.ChatLine{Op: proto.ChatError, Style: 1, MessageID: 347, Text: "hardcoded fallback"})
	if got.Text != "No such character" {
		t.Fatal("chat error ignored runtime table", got)
	}
}

func TestPerfectPartyExperienceUsesBothNativeMessagesAndUpdatesProgress(t *testing.T) {
	a := inputApp()
	a.msg = text.SysMsg{50: "XP <b>", 66: "Perfect party <b>"}
	a.state.Details.Exp = 90
	xp := proto.NewWriter(proto.ExpGain)
	xp.U32(468)
	xp.U32(468)
	a.applyWorldPacket(xp.Bytes())
	if a.state.Details.Exp != 1026 || len(a.state.Chat) != 2 {
		t.Fatalf("experience %d, messages %+v", a.state.Details.Exp, a.state.Chat)
	}
	base, bonus := a.state.Chat[0], a.state.Chat[1]
	if base.MessageID != 50 || base.Text != "XP 468" || bonus.MessageID != 66 || bonus.Text != "Perfect party 468" || bonus.Style != 3 {
		t.Fatalf("unexpected experience messages: %+v, %+v", base, bonus)
	}
}

func TestSkillChatUsesNumericEmphasisWithoutHPDamageSuffix(t *testing.T) {
	a := inputApp()
	a.state.SelfID = 1
	a.state.Entities[2] = &game.Entity{ID: 2, Kind: game.KindUser, Shape: &proto.UserShape{Name: "Opponent"}}
	a.skillDefs = map[sdata.SkillKey]sdata.Skill{{ID: 4, Level: 1}: {Name: "Fire", Effect: 1}, {ID: 5, Level: 1}: {Name: "Heal", Effect: 2}, {ID: 6, Level: 1}: {Name: "Blessing", Effect: 4}}
	a.msg = text.SysMsg{1100: "Used <s>.", 1101: "Hit <t> for <v> <hp>damage using <s>.", 1103: "Restored <v><hp> using <s>.", 416: "HP", 417: "Stamina", 418: "Mana"}
	a.state.LastSkill = proto.SkillResult{Source: 1, Target: 2, ID: 4, Level: 1, HP: 73}
	a.confirmedSystemChat(proto.NewWriter(proto.SkillUseNoTarget).Bytes(), 0)
	if len(a.state.Chat) != 2 || a.state.Chat[0].MessageID != 1100 || a.state.Chat[1].Text != "Hit Opponent for {c3}73{/c} damage using Fire." {
		t.Fatal(a.state.Chat)
	}
	a.state.LastSkill = proto.SkillResult{Source: 1, Target: 1, ID: 5, Level: 1, HP: 30}
	a.confirmedSystemChat(proto.NewWriter(proto.SkillUseNoTarget).Bytes(), 0)
	if got := a.state.Chat[len(a.state.Chat)-1].Text; got != "Restored {c3}30{/c}HP using Heal." {
		t.Fatal(got)
	}
	a.state.LastSkill = proto.SkillResult{Source: 1, ID: 6, Level: 1, Result: 4}
	a.confirmedSystemChat(proto.NewWriter(proto.SkillUseNoTarget).Bytes(), 0)
	if got := a.state.Chat[len(a.state.Chat)-1]; got.MessageID != 1100 || got.Text != "Used Blessing." {
		t.Fatal(got)
	}
	if a.combatPoolLabel(1, true) != "Mana" || a.combatPoolLabel(2, true) != "Stamina" {
		t.Fatal("resource labels ignored localization")
	}
}

func TestPeriodicAndExtraDamageChatPreservesWireResourceMeaning(t *testing.T) {
	a := inputApp()
	a.state.SelfID = 1
	a.skillDefs = map[sdata.SkillKey]sdata.Skill{{ID: 4, Level: 1}: {Name: "Poison", Effect: 3}, {ID: 5, Level: 1}: {Name: "Regen", Effect: 4}}
	a.msg = text.SysMsg{1310: "Tick <v><hp> from <s>", 1103: "Heal <v><hp> from <s>", 1105: "Extra <v><hp> from <s>", 416: "HP", 417: "SP", 418: "MP"}
	a.state.LastPeriodicEffect = proto.PeriodicEffect{Target: 1, Skill: 4, Level: 1, HP: 7, Resource1: 3, Resource2: 5}
	a.confirmedSystemChat(proto.NewWriter(proto.SkillKeep).Bytes(), 0)
	want := []string{"Tick {c3}7{/c} from Poison", "Tick {c3}5{/c}MP from Poison", "Tick {c3}3{/c}SP from Poison"}
	for i, v := range want {
		if a.state.Chat[i].Text != v {
			t.Fatal(a.state.Chat)
		}
	}
	a.state.LastPeriodicEffect = proto.PeriodicEffect{Target: 1, Skill: 5, Level: 1, Resource1: 8, Resource2: 6}
	a.confirmedSystemChat(proto.NewWriter(proto.SkillKeep).Bytes(), 0)
	if a.state.Chat[3].Text != "Heal {c3}8{/c}MP from Regen" || a.state.Chat[4].Text != "Heal {c3}6{/c}SP from Regen" {
		t.Fatal(a.state.Chat)
	}
	a.state.LastSkillExtra = proto.SkillResult{Source: 1, Target: 2, ID: 4, Level: 1, HP: 9}
	a.confirmedSystemChat(proto.NewWriter(proto.SkillExtraDamage).Bytes(), 0)
	if got := a.state.Chat[len(a.state.Chat)-1]; got.MessageID != 1105 || got.Text != "Extra {c3}9{/c} from Poison" {
		t.Fatal(got)
	}
	count := len(a.state.Chat)
	a.confirmedSystemChat(proto.NewWriter(proto.SkillUserReflectedDamage).Bytes(), 0)
	if len(a.state.Chat) != count {
		t.Fatal("reflection invented chat absent in native callbacks")
	}
	a.state.LastPeriodicEffect.Target = 999
	a.confirmedSystemChat(proto.NewWriter(proto.SkillKeep).Bytes(), 0)
	if len(a.state.Chat) != count {
		t.Fatal("unrelated pulse polluted chat")
	}
}

func TestScreenshotNoticeUsesNativeSysMsg(t *testing.T) {
	a := inputApp()
	a.state.Phase = game.PhaseInWorld
	a.msg = text.SysMsg{359: "saved <b>", 360: "capture failed"}
	a.ReportScreenshot("capture.png")
	a.ReportScreenshot("")
	lines := a.state.Chat
	if len(lines) != 2 || lines[0].MessageID != 359 || lines[0].Style != 7 || lines[0].Text != "saved capture.png" ||
		lines[1].MessageID != 360 || lines[1].Style != 17 || lines[1].Text != "capture failed" {
		t.Fatalf("wrong localized screenshot notices: %+v", lines)
	}
}
