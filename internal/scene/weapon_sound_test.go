package scene

import (
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/data/model"
)

func TestWeaponSoundGroupsAndAliases(t *testing.T) {
	for _, weapon := range []uint8{1, 45} {
		swing, hit := WeaponSounds(weapon, true, 0)
		if swing != "ch_att_swordone003.ogg" || hit != "ch_hit_swordone003.ogg" {
			t.Fatal("critical sword did not use native group 12/15", weapon, swing, hit)
		}
	}
	for weapon := uint8(45); weapon <= 65; weapon++ {
		kind := weaponKind(weapon)
		for variant := uint8(0); variant < 3; variant++ {
			a, b := WeaponSounds(weapon, variant == 2, variant)
			c, d := WeaponSounds(kind, variant == 2, variant)
			if a != c || b != d || WeaponSwingTime(weapon) != WeaponSwingTime(kind) {
				t.Fatal("EP5 weapon alias changed its cue", weapon)
			}
		}
	}
	for _, weapon := range []uint8{0, 16, 26, 66, 95, 255} {
		if a, b := WeaponSounds(weapon, false, 0); a != "" || b != "" {
			t.Fatal("unknown/unarmed context acquired a weapon cue", weapon)
		}
	}
	if _, hit := WeaponSounds(14, true, 0); hit != "ch_hit_crobow003.ogg" {
		t.Fatal("crossbow lost its shipped hit sound", hit)
	}
	if _, hit := WeaponSounds(14, false, 1); hit != "ch_hit_crobow002.ogg" {
		t.Fatal("crossbow normal hit variant", hit)
	}
	for _, tc := range []struct {
		weapon uint8
		frame  float32
	}{{1, 9}, {6, 11}, {11, 12}, {13, 1}, {14, 12}} {
		if WeaponSwingTime(tc.weapon) != tc.frame/30 {
			t.Fatal("native swing frame mismatch", tc)
		}
	}
}

func TestContactUsesFirstPositiveMarkerAndActionToken(t *testing.T) {
	c := &Character{actions: []model.Action{{Metadata: [4]float32{1, 2, 12, 6}}}, clip: &Clip{ani: &model.ANI{End: 30}}}
	if c.ContactTime() != .2 {
		t.Fatal("contact did not select the first native phase", c.ContactTime())
	}
	c.actions[0].Metadata[2], c.actions[0].Metadata[3] = 0, -1
	if c.ContactTime() != 0 {
		t.Fatal("missing marker received an invented contact")
	}
	old := c.ActionID()
	c.ReleaseAction()
	if c.ActionID() == old {
		t.Fatal("release did not invalidate queued cues")
	}
}

// Game.exe selects the release cue by animation on the physical path and
// adds ef_cst_crobow01 for ranged weapons.
func TestSkillReleaseSounds(t *testing.T) {
	for _, tc := range []struct {
		animation              uint16
		attack, effect, weapon uint8
		want                   []string
	}{
		{100, 1, 1, 1, []string{"ch_skill001.ogg"}},
		{102, 1, 1, 1, []string{"ch_skill003.ogg"}},
		{103, 2, 1, 1, []string{"ch_skill004.ogg"}},
		{104, 1, 1, 1, []string{"ch_skill005.ogg"}},
		{107, 1, 1, 1, []string{"ch_skill008.ogg"}},
		{101, 1, 1, 1, nil},
		{105, 1, 1, 1, nil},
		{110, 1, 1, 1, nil},
		{103, 1, 3, 1, []string{"ch_skill004.ogg"}},
		{101, 1, 3, 1, nil},
		{101, 1, 3, 13, []string{"ef_cst_crobow01.ogg"}},
		{103, 3, 3, 13, nil},
		{0, 1, 1, 13, []string{"ef_cst_crobow01.ogg"}},
		{0, 2, 1, 14, []string{"ef_cst_crobow01.ogg"}},
		{102, 1, 1, 64, []string{"ch_skill003.ogg", "ef_cst_crobow01.ogg"}},
		{0, 0, 1, 13, nil},
		{0, 1, 2, 13, nil},
		{0, 1, 4, 13, nil},
	} {
		got := SkillReleaseSounds(tc.animation, tc.attack, tc.effect, tc.weapon)
		if len(got) != len(tc.want) {
			t.Fatalf("%+v: got %v want %v", tc, got, tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("%+v: got %v want %v", tc, got, tc.want)
			}
		}
	}
}
