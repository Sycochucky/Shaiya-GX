package app

import (
	"slices"
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/audio"
	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
)

func TestAttackAudioFollowsActionAndCancelsStaleActor(t *testing.T) {
	actor := &scene.Character{}
	cue := attackSound{actionClock: actionClock{actor, actor.ActionID(), .4}}
	actor.Time = .39
	if due, alive := cue.ready(actor); due || !alive {
		t.Fatal("weapon cue fired before its contact")
	}
	actor.Time = .4
	if due, alive := cue.ready(actor); !due || !alive {
		t.Fatal("weapon cue missed its contact")
	}
	for _, other := range []*scene.Character{nil, {Time: 1}} {
		if due, alive := cue.ready(other); due || alive {
			t.Fatal("removed/replaced actor kept its old cue")
		}
	}
	actor.ReleaseAction()
	if due, alive := cue.ready(actor); due || alive {
		t.Fatal("interrupted action kept its old cue")
	}
}

func TestSkillReleaseSpecialCasesAndSecondaryResults(t *testing.T) {
	def := sdata.Skill{Animation: 102, Attack: 1, Effect: 3, Target: 3}
	for _, tc := range []struct {
		id   uint16
		want []string
	}{
		{134, []string{"ch_skill003.ogg"}},
		{38, []string{"ef_in_spin01.ogg"}},
		{156, []string{"ef_in_spin01.ogg"}},
		{341, []string{"ef_in_spin01.ogg"}},
		{56, nil},
	} {
		if got := skillReleaseSounds(tc.id, def, 1); !slices.Equal(got, tc.want) {
			t.Errorf("skill %d release = %v, want %v", tc.id, got, tc.want)
		}
	}

	def.Animation = 101
	a := inputApp()
	a.sound = &audio.Engine{}
	a.state.SelfID = 1
	a.state.LastSkill = proto.SkillResult{Source: 1, Target: 2, ID: 134, Level: 1}
	a.skillDefs = map[sdata.SkillKey]sdata.Skill{{ID: 134, Level: 1}: def}
	a.clock = 1
	a.confirmedSound(proto.NewWriter(proto.SkillUse).Bytes())
	key := skillSoundKey{source: 1, id: 134, level: 1}
	if a.skillSoundEvents[key] != 1 {
		t.Fatal("primary result missed its release")
	}
	a.clock = 2
	for _, op := range []proto.Opcode{proto.SkillPartyMember, proto.SkillMobOther} {
		a.confirmedSound(proto.NewWriter(op).Bytes())
	}
	if a.skillSoundEvents[key] != 1 {
		t.Fatal("late recipient replayed the caster release")
	}
}
