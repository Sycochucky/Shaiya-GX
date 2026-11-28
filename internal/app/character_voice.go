package app

import (
	"math/rand/v2"

	"github.com/Spelunkern/Shaiya-GX/internal/game"
)

type characterVoice uint8

const (
	voiceAttack characterVoice = iota
	voiceDamage
	voiceDeath
)

var characterAttackVoices = [4][2]string{
	{"ch_ohatm001.ogg", "ch_ohatw001.ogg"}, // Human
	{"ch_ohatm002.ogg", "ch_ohatw002.ogg"}, // Elf
	{"ch_ohatm003.ogg", "ch_ohatw003.ogg"}, // Vail
	{"ch_ohatm004.ogg", "ch_ohatw004.ogg"}, // Deatheater
}

var characterVoiceStems = [4][2]string{
	{"ch_hum", "ch_huw"},
	{"ch_elm", "ch_elw"},
	{"ch_vim", "ch_viw"},
	{"ch_dem", "ch_dew"},
}

func characterVoiceFile(race, gender uint8, cue characterVoice) string {
	if race >= uint8(len(characterVoiceStems)) || gender > 1 {
		return ""
	}
	switch cue {
	case voiceAttack:
		return characterAttackVoices[race][gender]
	case voiceDamage:
		return characterVoiceStems[race][gender] + "_dam.ogg"
	case voiceDeath:
		return characterVoiceStems[race][gender] + "_die.ogg"
	}
	return ""
}

func (a *App) playCharacterVoice(id uint32, cue characterVoice) {
	if id == 0 || a.sound == nil || a.wv == nil || !a.clientSettings().CharacterSound {
		return
	}
	var race, gender uint8
	if id == a.state.SelfID {
		slot, ok := a.selfSlot()
		if !ok {
			return
		}
		race, gender = slot.Race, slot.Gender
	} else if e := a.state.Entities[id]; e != nil && e.Kind == game.KindUser && e.Shape != nil {
		race, gender = e.Shape.Race, e.Shape.Gender
	} else {
		return
	}
	if pos, ok := a.effectPosition(id); ok {
		a.playSoundWithin(characterVoiceFile(race, gender, cue), pos[0], pos[2], 10)
	}
}

// EP5 always plays the reaction on critical damage and samples normal hits.
func (a *App) playCharacterDamageVoice(id uint32, result uint8, hp uint16) {
	// Game.exe 0x40ff40 plays the weapon's damage action with the pain voice.
	if hp > 0 && result <= 1 {
		if actor := a.playerActor(id); actor != nil {
			actor.PlayDamageReaction(a.playerWeapon(id))
		}
	}
	if hp > 0 && (result == 1 || result == 0 && rand.IntN(3) == 0) {
		a.playCharacterVoice(id, voiceDamage)
	}
}
