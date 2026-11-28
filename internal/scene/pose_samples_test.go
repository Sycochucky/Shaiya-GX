package scene

import (
	"slices"
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/render"
)

func TestPoseSamplesShareOnlyEqualLoopPhases(t *testing.T) {
	clip := poseTestClip(2)
	first := &Character{clip: clip, Time: .25}
	second := &Character{clip: clip, Time: .25}
	different := &Character{clip: clip, Time: .75}
	var poses PoseSamples
	poses.Sample(first)
	poses.Sample(second)
	if len(poses.samples) != 1 || !slices.Equal(first.finals(), second.finals()) {
		t.Fatal("equal looping poses did not share a sample")
	}
	if &first.lastFinals[0] == &second.lastFinals[0] {
		t.Fatal("actors share writable pose storage")
	}
	poses.Sample(different)
	if len(poses.samples) != 2 || !slices.Equal(different.finals(), clip.Finals(.75)) {
		t.Fatal("a different phase reused the wrong pose")
	}
	poses.Sample(&Character{clip: clip, oneShot: true, Time: .25})
	poses.Sample(&Character{clip: clip, blendLeft: .05, Time: .25})
	if len(poses.samples) != 2 {
		t.Fatal("transitional pose entered the loop cache")
	}
	poses.Reset()
	if len(poses.samples) != 0 {
		t.Fatal("pose cache survived the update")
	}
}

func TestCrowdPaletteSharingKeepsBoneMasksAndPhases(t *testing.T) {
	clip := poseTestClip(3)
	binds := make([]render.M4, 3)
	batch := CharacterBatch{
		partOffsets: make(map[*skinPart][2]int32),
		sharedPoses: make(map[crowdPaletteKey]int32),
	}
	part := func() *skinPart {
		return &skinPart{bones: binds, pose: make([]render.M4, 3)}
	}
	first := batch.paletteOffset(&Character{clip: clip, Time: .25}, part(), 3, []int{0, 2})
	same := batch.paletteOffset(&Character{clip: clip, Time: .25}, part(), 3, []int{0, 2})
	mask := batch.paletteOffset(&Character{clip: clip, Time: .25}, part(), 3, []int{1, 2})
	phase := batch.paletteOffset(&Character{clip: clip, Time: .50}, part(), 3, []int{0, 2})
	if first != same || first == mask || first == phase || len(batch.palettes.Matrices) != 9 {
		t.Fatalf("incorrect palette ranges: %v %v %v %v (%d matrices)", first, same, mask, phase, len(batch.palettes.Matrices))
	}
}
