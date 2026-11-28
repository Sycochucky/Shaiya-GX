package app

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"

	"github.com/Spelunkern/Shaiya-GX/internal/clientlog"
	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/data/world"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
)

func (a *App) audioError(file string, err error) {
	if err != nil && !a.soundErrors[file] {
		clientlog.Warnf("audio %s: %v", file, err)
		a.soundErrors[file] = true
	}
}
func (a *App) playSound(name string, x, z float32) {
	a.playSoundWithin(name, x, z, 40)
}

func (a *App) playSoundWithin(name string, x, z float32, radius float64) {
	if a.sound == nil || name == "" || strings.EqualFold(name, "LOAD") {
		return
	}
	gain := 1.0
	if a.wv != nil {
		dx, dz := float64(x-a.wv.pos[0]), float64(z-a.wv.pos[2])
		gain = math.Max(0, 1-math.Hypot(dx, dz)/radius)
	}
	if gain == 0 {
		return
	}
	file := a.audioPath(name, false)
	a.audioError(file, a.sound.Play(file, gain))
}
func mapMusic(w *world.WLD, x, z float32) string {
	if w == nil {
		return ""
	}
	for _, zone := range w.MusicZones {
		if x < zone.Box.Min[0] || x > zone.Box.Max[0] || z < zone.Box.Min[2] || z > zone.Box.Max[2] {
			continue
		}
		if zone.Asset >= 0 && int(zone.Asset) < len(w.MusicAssets) {
			return w.MusicAssets[zone.Asset]
		}
	}
	return ""
}
func (a *App) updateAudio() {
	if a.sound == nil {
		return
	}
	a.updateAttackSounds()
	a.audioError("device", a.sound.Update())
	music := ""
	switch a.screen {
	case screenLogin, screenServers:
		music = "bgm_title.ogg"
	case screenWorld:
		// World music and emitters share the fading map mix.
	case screenCharacters:
		key := "select_a"
		if a.state != nil && a.state.Faction.Faction == 1 {
			key = "select_b"
		}
		if a.creating {
			key = "creation"
		}
		if w := a.worlds[key]; w != nil {
			x, z := selectActorPos[0], selectActorPos[2]
			if a.creating {
				x, z = 68, 66
			}
			music = mapMusic(w.WLD, x, z)
		}
	}
	if a.loading.active {
		music = ""
		if a.loading.presented {
			music = "bgm_load0001.ogg"
		}
	}
	if music != "" {
		music = a.audioPath(music, true)
	}
	a.audioError(music, a.sound.Music(music))
	a.updateMapAudio()
}
func (a *App) confirmedSound(p []byte) {
	if a.sound == nil || a.wv == nil {
		return
	}
	switch proto.Op(p) {
	// Native 0056ee70: every successful visible consumable uses the same
	// potion cue inside ten units; local bag-zero failures use UI error.
	case proto.UseItem:
		r := a.state.LastItemUse
		if r.Bag != 0 {
			if pos, ok := a.effectPosition(r.ID); ok {
				distance := pos.Sub(a.wv.pos).Len()
				if distance < 10 {
					name := a.audioPath("ch_potion01.wav", false)
					a.audioError(name, a.sound.Play(name, float64(1-distance/10)))
				}
			}
		} else if r.ID == a.state.SelfID {
			name := a.audioPath("ui_error001.wav", false)
			a.audioError(name, a.sound.Play(name, 1))
		}
	// Native 0x56e5xx (item result): a refused pickup plays the UI error, a
	// granted one ch_pikup001 on the local character.
	case proto.AddItem:
		name := "ch_pikup001.ogg"
		if len(p) == 4 && p[2] == 0 {
			name = "ui_error001.ogg"
		}
		a.playSound(name, a.wv.pos[0], a.wv.pos[2])
	// Native 0x56b910 (buy), 0x56ba20 (sell): success pays with ui_pay001, a
	// refusal plays the UI error. 0x56bb10: every repair plays ui_repair001.
	case proto.ShopBuy:
		if r := a.state.LastShop; r.Result == 0 {
			a.playUISound("ui_pay001.ogg")
		} else if r.Result <= 2 {
			a.playUISound("ui_error001.ogg")
		}
	case proto.ShopSell:
		if r := a.state.LastShop; r.Result == 0 {
			a.playUISound("ui_pay001.ogg")
		} else if r.Result == 1 {
			a.playUISound("ui_error001.ogg")
		}
	case proto.ItemRepair:
		a.playUISound("ui_repair001.ogg")
	case proto.SkillCastStarted:
		r := a.state.LastCast
		a.playSkillSound(r.Source, r.ID, r.Level, true)
	case proto.SkillUse, proto.SkillUseNoTarget, proto.SkillUserDrain, proto.SkillMobDrain:
		r := a.state.LastSkill
		a.playSkillHitSound(r)
		a.playSkillHealSound(r)
		if r.Result <= 4 && !(r.Result == 3 && r.Target == 0) {
			a.playSkillSound(r.Source, r.ID, r.Level, false)
		}
	case proto.SkillPartyMember, proto.SkillMobOther:
		a.playSkillHitSound(a.state.LastSkill)
	case proto.MobSkillUse, proto.MobSkillDrain, proto.MobSkillOther:
		r := a.state.LastSkill
		if r.HP > 0 {
			a.playMobHitSound(r.Target, r.Result)
		}
	case proto.CharacterDeath:
		a.playCharacterVoice(proto.NewReader(p).U32(), voiceDeath)
	}
}

func (a *App) playSkillSound(source uint32, id uint16, level uint8, preparing bool) {
	pos, visible := a.effectPosition(source)
	if !visible || level == 0 || level == 100 {
		return
	}
	a.loadSkills()
	d, ok := a.skillDefs[sdata.SkillKey{ID: id, Level: level}]
	if !ok || d.Effect == 0 {
		return
	}
	action := scene.SkillCastAction(d.Animation, d.Effect)
	if !preparing {
		action = scene.SkillAction(d.Animation, d.Attack, d.Effect, 0, 0)
	}
	key := skillSoundKey{source, id, level, preparing}
	if at, ok := a.skillSoundEvents[key]; ok && a.clock-at < .15 {
		return
	}
	if a.skillSoundEvents == nil {
		a.skillSoundEvents = map[skillSoundKey]float64{}
	}
	// Bound the cache by retaining only the recent burst window.
	for k, at := range a.skillSoundEvents {
		if a.clock-at >= .15 {
			delete(a.skillSoundEvents, k)
		}
	}
	a.skillSoundEvents[key] = a.clock
	name := scene.SkillActionSound(action)
	a.playSound(name, pos[0], pos[2])
	if !preparing {
		for _, extra := range skillReleaseSounds(id, d, a.playerWeapon(source)) {
			a.playSound(extra, pos[0], pos[2])
		}
	}
}

func skillReleaseSounds(id uint16, d sdata.Skill, weapon uint8) []string {
	// These actions return before the normal physical release cue.
	switch id {
	case 38, 156, 341:
		return []string{"ef_in_spin01.ogg"}
	case 56:
		return nil
	}
	return scene.SkillReleaseSounds(d.Animation, d.Attack, d.Effect, weapon)
}

func (a *App) playEntitySound(target uint32, file string) {
	if a.sound == nil || file == "" || target == 0 {
		return
	}
	if pos, ok := a.effectPosition(target); ok {
		a.playSound(file, pos[0], pos[2])
	}
}

// playSkillHealSound is the native heal impact (0x57d896): a heal that lands
// plays ef_dmg_heal01 on its target, not on the caster.
func (a *App) playSkillHealSound(r proto.SkillResult) {
	if r.Result > 1 || r.Target == 0 || r.Level == 0 || r.Level == 100 {
		return
	}
	if d, ok := a.skillDefinition(r.ID, r.Level); ok && d.Effect == 2 {
		if pos, ok := a.effectPosition(r.Target); ok {
			a.playSound(scene.HealImpactSound, pos[0], pos[2])
		}
	}
}

// playUISound plays a global interface cue (native position 0x22094c8).
func (a *App) playUISound(name string) {
	if a.sound == nil {
		return
	}
	file := a.audioPath(name, false)
	a.audioError(file, a.sound.Play(file, .6))
}

func (a *App) playUI() {
	if a.sound != nil {
		a.audioError("Sound/ui_button001.ogg", a.sound.Play("Sound/ui_button001.ogg", 0.4))
	}
}

type skillSoundKey struct {
	source    uint32
	id        uint16
	level     uint8
	preparing bool
}

// Trap and Demon Hushing have no sound in the original (no voice for clips
// 108/109 and no impact id in 0x5761b0). GX gives their damage the registered
// but otherwise unused hit cue ef_dmg_hit01 (0x426613, id 0x30) on the target.
func (a *App) playSkillHitSound(r proto.SkillResult) {
	if r.HP == 0 || r.Result > 1 || r.Target == 0 {
		return
	}
	switch r.ID {
	case 113, 169, 171:
		if pos, ok := a.effectPosition(r.Target); ok {
			a.playSound("ef_dmg_hit01.ogg", pos[0], pos[2])
		}
	}
}

// playMobHitSound is the impact of a creature's blow on a character (Game.exe
// 0x4320fb, 0x433ad5 and 0x433c30): a normal hit picks mob_hit001 or 002 at
// random, a critical uses mob_hit003, and only hits (result 0 or 1) sound.
func (a *App) playMobHitSound(target uint32, result uint8) {
	if result > 1 || target == 0 {
		return
	}
	pos, ok := a.effectPosition(target)
	if !ok {
		return
	}
	name := "mob_hit003.ogg"
	if result == 0 {
		name = fmt.Sprintf("mob_hit00%d.ogg", 1+rand.IntN(2))
	}
	a.playSound(name, pos[0], pos[2])
}
