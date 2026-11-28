package scene

// SkillActionSound follows the common cast/release branches in 0x5780a0,
// 0x57bdc0 and 0x57e320. IDs 0x21..0x26 are registered at 0x4264b6..0x42653b.
// Physical attacks use separate weapon/voice cues, not a generic spell sound.
func SkillActionSound(action int) string {
	switch action {
	case 85:
		return "ch_maready01.ogg"
	case 87:
		return "ef_in_firing01.ogg"
	case 88:
		return "ch_maready02.ogg"
	case 90:
		return "ch_maattack02.ogg"
	case 91:
		return "ch_buready01.ogg"
	case 93:
		return "ch_buattack01.ogg"
	}
	return ""
}

// SkillReleaseSounds follows the physical release branches at 0x57cf21
// and 0x57f10e. The animation selects the cue before checking the equipped
// weapon for the additional projectile release sound.
func SkillReleaseSounds(animation uint16, attack, effect, weapon uint8) []string {
	if (effect != 1 && effect != 3 && effect != 5) || (attack != 1 && attack != 2) {
		return nil
	}
	var out []string
	switch uint8(animation) {
	case 100:
		out = append(out, "ch_skill001.ogg")
	case 102:
		out = append(out, "ch_skill003.ogg")
	case 103:
		out = append(out, "ch_skill004.ogg")
	case 104:
		out = append(out, "ch_skill005.ogg")
	case 107:
		out = append(out, "ch_skill008.ogg")
	}
	switch weaponKind(weapon) {
	case 11, 13, 14:
		out = append(out, "ef_cst_crobow01.ogg")
	}
	return out
}

// HealImpactSound is the cue Game.exe 0x57d896 plays on the healed target
// when a heal skill (effect 2) lands as a normal or critical result.
const HealImpactSound = "ef_dmg_heal01.ogg"
