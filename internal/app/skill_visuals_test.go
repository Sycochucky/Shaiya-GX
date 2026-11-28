package app

import (
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
)

func TestSkillEffectsRouteDirectAndProjectileResults(t *testing.T) {
	for _, tc := range []struct {
		id                     uint16
		attack, effect, target uint8
		result                 uint8
		flight, impact         string
	}{
		{37, 3, 3, 3, 0, "CA_ice.EFT", "DA_target_generic.EFT"},
		{37, 3, 3, 3, 3, "CA_ice.EFT", "DA_target_generic.EFT"},
		{30, 0, 4, 2, 0, "", "DA_target_bash.EFT"},
		{31, 0, 4, 2, 0, "", "DA_target_concentration.EFT"},
		{57, 1, 1, 3, 1, "", "DA_target_generic.EFT"},
		{35, 3, 3, 3, 0, "", "DA_target_root.EFT"},
		{42, 3, 3, 3, 0, "", "DA_target_slumber.EFT"},
		{57, 1, 1, 3, 3, "", "DA_target_generic.EFT"}, // A missed spell still plays its effect.
		{30, 0, 4, 2, 5, "", ""},
		{30, 0, 4, 2, 4, "", ""}, // Area announcements do not duplicate impacts.
		{38, 1, 1, 7, 0, "", "DA_target_generic.EFT"},
		{85, 3, 3, 6, 4, "", "DA_area_frozen.EFT"},
		{30, 0, 0, 0, 0, "", ""},
		{65535, 1, 1, 3, 0, "", ""},
	} {
		r := proto.SkillResult{ID: tc.id, Level: 1, Target: 42, Result: tc.result}
		d := sdata.Skill{Attack: tc.attack, Effect: tc.effect, Target: tc.target}
		flight, impact := skillEffectLibraries(r, d)
		wantFlight, wantImpact := tc.flight, tc.impact
		if wantFlight != "" {
			wantFlight = "Effect/" + wantFlight
		}
		if wantImpact != "" {
			wantImpact = "Effect/" + wantImpact
		}
		if flight != wantFlight || impact != wantImpact {
			t.Fatalf("skill %d result %d: %q/%q, want %q/%q", tc.id, tc.result, flight, impact, wantFlight, wantImpact)
		}
		for _, invalid := range []proto.SkillResult{{ID: tc.id, Target: 42}, {ID: tc.id, Level: 100, Target: 42}, {ID: tc.id, Level: 1, Result: 5}} {
			if f, i := skillEffectLibraries(invalid, d); f != "" || i != "" {
				t.Fatal("invalid result created a visual recipe", invalid)
			}
		}
	}
}

func TestSkillEffectPlayerAnchorKeepsActorWorldUnchanged(t *testing.T) {
	a := inputApp()
	a.state.SelfID = 1
	a.wv.pos = render.V3{10, 20, 30}
	world := render.Basis(a.wv.pos, render.V3{0, 0, 1}, render.V3{0, 1, 0})
	a.wv.self = &scene.Character{World: world}
	if m, ok := a.effectTargetMatrix(1); !ok || m[13] != world[13] || a.wv.self.World != world {
		t.Fatal("player effect offset mutated the actor", m, ok)
	}
	a.wv.self = nil
	if m, ok := a.effectTargetMatrix(1); !ok || m[13] != world[13] {
		t.Fatal("unloaded player did not keep the same effect origin", m, ok)
	}
	a.state.Entities[11] = &game.Entity{ID: 11, Kind: game.KindMob, X: 10, Y: 20, Z: 30}
	if m, ok := a.effectTargetMatrix(11); !ok || m[13] != world[13] {
		t.Fatal("mob effect incorrectly used the player offset", m, ok)
	}
}

func TestCasterAreaIgnoresSelectionAndPreservesRecipientEffects(t *testing.T) {
	a := inputApp()
	a.state.SelfID, a.state.Target = 1, 2
	a.skillDefs = map[sdata.SkillKey]sdata.Skill{{ID: 85, Level: 1}: {Attack: 3, Effect: 3, Target: 6}}
	p := a.learnedSkillPacket(proto.LearnedSkill{ID: 85, Level: 1, Number: 4})
	r := proto.NewReader(p)
	if proto.Op(p) != proto.SkillUse || r.U8() != 4 || r.U32() != 0 {
		t.Fatal("caster area depends on selected victim", p)
	}
	if flight, body := skillEffectLibraries(proto.SkillResult{ID: 85, Level: 1, Result: 4, Source: 1}, a.skillDefs[sdata.SkillKey{ID: 85, Level: 1}]); flight != "" || body != "Effect/DA_area_frozen.EFT" {
		t.Fatal("empty area lost its body effect", flight, body)
	}
	result := proto.SkillResult{ID: 85, Level: 1, Source: 1, Target: 2}
	if file, _ := skillSecondaryLibrary(result, sdata.Skill{Effect: 3}); file != "Effect/DA_area_frozen.EFT" {
		t.Fatal("debuff recipient lost its original frozen effect", file)
	}
}

func TestSecondaryAttackUsesNativeElementSequence(t *testing.T) {
	r := proto.SkillResult{ID: 85, Level: 3, Source: 1, Target: 2}
	for attr, sequence := range []int{4, 2, 0, 3, 1} {
		file, got := skillSecondaryLibrary(r, sdata.Skill{Effect: 1, Attribute: uint8(attr)})
		if file != "Effect/DA_target_Adamage.EFT" || got != sequence {
			t.Fatal("secondary attack ignored original element sequence", attr, file, got)
		}
	}
	r.Result = 3
	if file, _ := skillSecondaryLibrary(r, sdata.Skill{Effect: 3}); file == "" {
		t.Fatal("miss dropped the recipient effect")
	}
	r.Result = 5
	if file, _ := skillSecondaryLibrary(r, sdata.Skill{Effect: 3}); file != "" {
		t.Fatal("rejection emitted recipient effect")
	}
}

func TestBlizzardAndRelatedAreaAnnouncementsUseOriginalAreaSelector(t *testing.T) {
	for _, tc := range []struct {
		id   uint16
		file string
	}{{71, "blizzard"}, {53, "hailstorm"}, {68, "meteor"}} {
		d := sdata.Skill{Attack: 3, Effect: 3, Target: 7}
		r := proto.SkillResult{Source: 1, Target: 2, ID: tc.id, Level: 1, Result: 4}
		flight, impact := skillEffectLibraries(r, d)
		want := "Effect/DA_area_" + tc.file + ".EFT"
		if flight != "" || impact != want {
			t.Fatal("area announcement lost original library", tc.id, flight, impact)
		}
		r.Result = 0
		if file, _ := skillSecondaryLibrary(r, d); file != "" {
			t.Fatal("recipient duplicated central area cloud", tc.id, file)
		}
	}
}

func TestTargetCenteredAreasKeepRecipientEffectsWithoutDuplicatingAnnouncement(t *testing.T) {
	for _, id := range []uint16{48, 50, 52, 69, 70, 72, 83, 84} {
		d := sdata.Skill{Attack: 3, Effect: 1, Target: 7}
		r := proto.SkillResult{ID: id, Level: 1, Source: 1, Target: 2, Result: 4}
		if _, file := skillEffectLibraries(r, d); file != "" {
			t.Fatal("recipient area incorrectly duplicated at center", id, file)
		}
		r.Result = 0
		if file, _ := skillSecondaryLibrary(r, d); file == "" {
			t.Fatal("missing recipient area", id)
		}
	}
}

func TestSkillEffectSoundsFollowResultPhase(t *testing.T) {
	for _, tc := range []struct {
		id             uint16
		target, result uint8
		want           string
	}{
		{134, 3, 0, "ef_dmg_earth01.ogg"},
		{309, 3, 0, "ef_dmg_boom01.ogg"},
		{310, 3, 0, "ef_dmg_earth01.ogg"},
		{307, 3, 0, ""},
		{322, 3, 0, ""},
		{323, 3, 0, ""},
		{85, 6, 4, "ef_dmg_water01.ogg"},
		{157, 6, 4, "ef_dmg_earth01.ogg"},
		{69, 7, 4, "ef_dmg_earth01.ogg"},
		{84, 7, 4, "ef_dmg_wind01.ogg"},
		{53, 7, 4, "ef_dmg_hail01.ogg"},
		{312, 7, 4, ""},
		{315, 7, 4, ""},
		{134, 3, 5, ""},
	} {
		r := proto.SkillResult{ID: tc.id, Level: 1, Result: tc.result}
		if got := skillResultEffectSound(r, sdata.Skill{Effect: 3, Target: tc.target}); got != tc.want {
			t.Errorf("skill %d result %d: got %q, want %q", tc.id, tc.result, got, tc.want)
		}
	}
}

func TestMobSkillSoundUsesTheNPCVisualAndPacketRoute(t *testing.T) {
	for _, tc := range []struct {
		visual     uint16
		effect     uint8
		startsCast bool
		file, cue  string
	}{
		{85, 3, true, "DA_area_frozen.EFT", "ef_dmg_water01.ogg"},
		{85, 3, false, "DA_area_frozen.EFT", "ef_dmg_water01.ogg"},
		{81, 1, true, "", ""},
		{81, 1, false, "DA_target_valribomb.EFT", "ef_dmg_boom01.ogg"},
		{81, 3, false, "", ""},
		{52, 3, false, "", ""},
		{53, 1, false, "", ""},
		{68, 1, false, "", ""},
		{71, 1, false, "", ""},
		{307, 1, false, "DA_area_turbulent.EFT", ""},
	} {
		file, cue := mobSkillImpact(sdata.Skill{VisualCode: tc.visual, Effect: tc.effect}, tc.startsCast)
		wantFile := tc.file
		if wantFile != "" {
			wantFile = "Effect/" + wantFile
		}
		if file != wantFile || cue != tc.cue {
			t.Errorf("mob visual %d effect %d first %t: got %q/%q, want %q/%q", tc.visual, tc.effect, tc.startsCast, file, cue, wantFile, tc.cue)
		}
	}
}
