package ui

import (
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
)

func TestStatusDisplayUsesConfirmedAttributes(t *testing.T) {
	state := StatusState{
		Base:  proto.StatPoints{20, 21, 22, 23, 24, 25},
		Stats: proto.AdditionalStats{Str: 28, Rec: 32, Int: 17, Wis: 65536, Dex: 0, Luc: 38},
	}
	if got, want := state.displayedAttributes(), ([6]uint32{20, 22, 23, 24, 21, 25}); got != want {
		t.Fatalf("base attributes before the first snapshot = %v, want %v", got, want)
	}
	state.Known = true
	if got, want := state.displayedAttributes(), ([6]uint32{28, 32, 17, 65536, 0, 38}); got != want {
		t.Fatalf("confirmed attribute totals = %v, want %v", got, want)
	}
}

func TestStatusDraftLimitsCancellationAndWireOrder(t *testing.T) {
	v, err := NewStatus(marketTextures{})
	if err != nil {
		t.Fatal(err)
	}
	state := StatusState{
		Points: 3,
		Base:   proto.StatPoints{65535, 10, 10, 10, 10, 10},
		Stats:  proto.AdditionalStats{Str: 65000, Rec: 66000, Dex: 66001},
		Known:  true,
	}
	click := func(x, y int) bool {
		t.Helper()
		v.Update(&platform.Input{MouseX: x, MouseY: y, Pressed: true, MouseDown: true}, 1024, 768, state, 0)
		apply, _ := v.Update(&platform.Input{MouseX: x, MouseY: y, Clicked: true}, 1024, 768, state, 0)
		return apply
	}
	x, y := StatusOrigin(1024, 768, false)
	click(x+1, y+157)
	if !v.Expanded {
		t.Fatal("allocation tab did not expand")
	}
	x, y = StatusOrigin(1024, 768, true)
	click(x+86, y+129) // STR saturated: no point may be wasted.
	if v.Draft.Total() != 0 {
		t.Fatal("allocation overflow")
	}
	click(x+86, y+149) // REC row -> wire index 2.
	click(x+86, y+209) // DEX row -> wire index 1.
	click(x+86, y+209)
	click(x+86, y+209)
	if v.Draft != (proto.StatPoints{0, 2, 1, 0, 0, 0}) || state.Points != 3 || state.Base[1] != 10 {
		t.Fatal("draft/order/budget changed confirmed state", v.Draft, state)
	}
	state.Busy = true
	if click(x+16, y+270) {
		t.Fatal("duplicate apply while pending")
	}
	state.Busy = false
	if !click(x+16, y+270) {
		t.Fatal("valid allocation did not apply")
	}
	click(x+81, y+270)
	if v.Expanded || v.Draft.Total() != 0 {
		t.Fatal("cancel retained unconfirmed points")
	}
}

// 0x5037f0: the side tab only opens the sheet; Apply and Cancel fold it.
func TestStatusTabOnlyOpensAndApplyFoldsSheet(t *testing.T) {
	v, err := NewStatus(marketTextures{})
	if err != nil {
		t.Fatal(err)
	}
	state := StatusState{Points: 3, Base: proto.StatPoints{10, 10, 10, 10, 10, 10}}
	click := func(x, y int) bool {
		v.Update(&platform.Input{MouseX: x, MouseY: y, Pressed: true, MouseDown: true}, 1024, 768, state, 0)
		apply, _ := v.Update(&platform.Input{MouseX: x, MouseY: y, Clicked: true}, 1024, 768, state, 0)
		return apply
	}
	x, y := StatusOrigin(1024, 768, false)
	click(x+1, y+157)
	if !v.Expanded {
		t.Fatal("tab did not open")
	}
	x, y = StatusOrigin(1024, 768, true)
	click(x+132+1, y+157) // the tab is hidden while open
	if !v.Expanded {
		t.Fatal("tab folded the open sheet")
	}
	if !click(x+16, y+270) {
		t.Fatal("empty Apply must still be accepted")
	}
}

func TestStatusHoldRepeatsPlusAtTenPointsPerSecond(t *testing.T) {
	v, err := NewStatus(marketTextures{})
	if err != nil {
		t.Fatal(err)
	}
	v.Expanded = true
	x, y := StatusOrigin(1024, 768, true)
	state := StatusState{Points: 100}
	held := platform.Input{MouseX: x + 86, MouseY: y + 129, MouseDown: true}
	press := func() {
		in := held
		in.Pressed = true
		v.Update(&in, 1024, 768, state, 0)
	}
	release := func() {
		in := held
		in.MouseDown, in.Clicked = false, true
		v.Update(&in, 1024, 768, state, 0)
	}

	// A press adds one point at once; releasing adds nothing more.
	press()
	if v.Draft[0] != 1 {
		t.Fatal("the press itself must add one point", v.Draft[0])
	}
	v.Update(&held, 1024, 768, state, .09)
	if v.Draft[0] != 1 {
		t.Fatal("repeat started before 100 ms")
	}
	v.Update(&held, 1024, 768, state, .02)
	if v.Draft[0] != 2 {
		t.Fatal("first repeat at 100 ms", v.Draft[0])
	}
	release()
	if v.Draft[0] != 2 {
		t.Fatal("release added a point")
	}

	start := v.Draft[0]
	press()
	for range 60 {
		v.Update(&held, 1024, 768, state, 1.0/60)
	}
	release()
	if gained := v.Draft[0] - start; gained != 11 {
		t.Fatalf("one second held must give the press plus ten repeats, gained %d", gained)
	}
	start = v.Draft[0]

	press()
	start++
	outside := held
	outside.MouseX -= 50
	v.Update(&outside, 1024, 768, state, .5)
	v.Update(&held, 1024, 768, state, .5)
	release()
	if v.Draft[0] != start {
		t.Fatal("repeat resumed after leaving the original button", v.Draft[0], start)
	}

	state.Base[0] = 65534 - v.Draft[0]
	press()
	v.Update(&held, 1024, 768, state, .8)
	release()
	if uint32(state.Base[0])+uint32(v.Draft[0]) != 65535 {
		t.Fatal("repeat exceeded attribute limit", v.Draft[0])
	}
	state.Base[0] = 0
	state.Points = v.Draft[0] + 2
	press()
	v.Update(&held, 1024, 768, state, .8)
	release()
	if v.Draft[0] != state.Points {
		t.Fatal("repeat exceeded available points", v.Draft[0])
	}
}

func TestStatusMinusButtonRepeatsLikeTheNativeRepeatButtons(t *testing.T) {
	v, err := NewStatus(marketTextures{})
	if err != nil {
		t.Fatal(err)
	}
	v.Expanded = true
	v.Draft[0] = 25
	state := StatusState{Points: 100, Base: [6]uint16{10}}
	x, y := StatusOrigin(1024, 768, true)
	held := platform.Input{MouseX: x + 112, MouseY: y + 129, MouseDown: true} // STR minus
	press := held
	press.Pressed = true
	v.Update(&press, 1024, 768, state, 0)
	for range 30 {
		v.Update(&held, 1024, 768, state, 1.0/60)
	}
	release := held
	release.MouseDown, release.Clicked = false, true
	v.Update(&release, 1024, 768, state, 0)
	if v.Draft[0] != 25-1-5 {
		t.Fatal("half a second on the minus button must take back the press plus five", v.Draft[0])
	}
}

func TestStatusAttributeHover(t *testing.T) {
	const w, h = 1024, 768
	v := Status{}
	for _, expanded := range []bool{false, true} {
		v.Expanded = expanded
		v.Position.Move(82, 41, w, h, StatusWidth, StatusHeight)
		x, y := v.Origin(w, h)
		mainX := x
		if expanded {
			mainX += 132
		}
		for i := range 6 {
			mx, my := mainX+90+(i/3)*115, y+174+(i%3)*20
			if got := v.AttributeMessageAt(mx, my, w, h); got != 5283+i {
				t.Errorf("main attribute %d, expanded=%v: SysMsg %d", i, expanded, got)
			}
			if expanded {
				if got := v.AttributeMessageAt(x+60, y+133+i*20, w, h); got != 5283+i {
					t.Errorf("allocation attribute %d: SysMsg %d", i, got)
				}
			}
		}
		for _, point := range [][2]int{{mainX + 90, y + 188}, {mainX + 145, y + 174}, {mainX + 90, y + 265}, {x - 1, y + 174}} {
			if got := v.AttributeMessageAt(point[0], point[1], w, h); got != 0 {
				t.Errorf("point %v outside an attribute, expanded=%v: SysMsg %d", point, expanded, got)
			}
		}
	}
}
