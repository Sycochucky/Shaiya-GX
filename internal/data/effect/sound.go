package effect

// TargetSound follows the detailed target selector at 0x577390 and mode 1
// of Game.exe 0x5761b0. Some target recipes use an area EFT but have no cue
// in this mode. The converted water asset has its normal .ogg extension.
func TargetSound(skill uint16) string {
	switch skill {
	case 35:
		return "ef_dmg_hold01.ogg"
	case 36:
		return "ef_dmg_light01.ogg"
	case 51, 134, 145, 151, 153, 157, 177, 302, 310, 314, 315:
		return "ef_dmg_earth01.ogg"
	case 63, 77, 97, 107, 127, 146, 172, 274, 275:
		return "ef_dmg_transf01.ogg"
	case 85:
		return "ef_dmg_water01.ogg"
	case 87, 95, 174, 180, 282, 309, 312:
		return "ef_dmg_boom01.ogg"
	}
	return ""
}

// AreaSound follows the area selector at 0x577020 and mode 0 of 0x5761b0.
func AreaSound(skill uint16) string {
	switch skill {
	case 50, 69, 83:
		return "ef_dmg_earth01.ogg"
	case 52, 84:
		return "ef_dmg_wind01.ogg"
	case 53:
		return "ef_dmg_hail01.ogg"
	case 68, 163, 184:
		return "ef_dmg_meteo01.ogg"
	case 70:
		return "ef_dmg_light01.ogg"
	case 71, 162, 353, 355, 356:
		return "ef_dmg_blizzard01.ogg"
	case 81, 174, 301:
		return "ef_dmg_boom01.ogg"
	}
	return ""
}
