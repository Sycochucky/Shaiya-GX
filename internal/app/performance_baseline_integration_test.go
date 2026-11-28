//go:build integration

package app

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/data/archive"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/render/gl"
	"github.com/Spelunkern/Shaiya-GX/internal/session"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

type clientMeasureStats struct {
	Samples int     `json:"samples"`
	Mean    float64 `json:"mean_ms"`
	P50     float64 `json:"p50_ms"`
	P95     float64 `json:"p95_ms"`
	P99     float64 `json:"p99_ms"`
	Max     float64 `json:"max_ms"`
}

func clientMeasureSummary(values []float64) clientMeasureStats {
	values = append([]float64(nil), values...)
	sort.Float64s(values)
	if len(values) == 0 {
		return clientMeasureStats{}
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	at := func(q float64) float64 { return values[max(0, int(math.Ceil(q*float64(len(values))))-1)] }
	return clientMeasureStats{len(values), sum / float64(len(values)), at(.5), at(.95), at(.99), values[len(values)-1]}
}
func clientMeasureInt(t *testing.T, key string, fallback, lo, hi int) int {
	t.Helper()
	s := os.Getenv(key)
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < lo || v > hi {
		t.Fatalf("%s must be %d..%d", key, lo, hi)
	}
	return v
}

// TestClientPerformanceBaseline is an opt-in measurement, not a pass/fail FPS
// target. Each process owns one scene/context and isolated client preferences.
// It never connects to a server or modifies a runtime database.
func TestClientPerformanceBaseline(t *testing.T) {
	if os.Getenv("GX_MEASURE_CLIENT") != "1" || os.Getenv("SHAIYA_TEST_RENDER") != "1" {
		t.Skip("explicit GL measurement with local assets")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	settingsDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", settingsDir)
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", settingsDir)
	}
	sceneName := os.Getenv("GX_MEASURE_SCENE")
	if sceneName == "" {
		sceneName = "world-stream"
	}
	variant := os.Getenv("GX_MEASURE_VARIANT")
	if variant == "" {
		variant = "default"
	}
	frames := clientMeasureInt(t, "GX_MEASURE_FRAMES", 900, 30, 30000)
	width := clientMeasureInt(t, "GX_MEASURE_WIDTH", 1280, 320, 3840)
	height := clientMeasureInt(t, "GX_MEASURE_HEIGHT", 720, 200, 2160)
	startSetup := time.Now()
	win, err := platform.Open(platform.Config{Title: "GX client baseline", Width: width, Height: height, Hidden: true, VSync: false})
	if err != nil {
		t.Fatal(err)
	}
	defer win.Close()
	if err = render.InitWithOptions(render.InitOptions{
		DisableBindless:       os.Getenv("GX_MEASURE_NO_BINDLESS") == "1",
		DisableDrawParameters: os.Getenv("GX_MEASURE_NO_DRAW_PARAMETERS") == "1",
	}); err != nil {
		t.Fatal(err)
	}
	renderer, vendor, version := gl.GoStr(gl.GetString(gl.RENDERER)), gl.GoStr(gl.GetString(gl.VENDOR)), gl.GoStr(gl.GetString(gl.VERSION))
	arc, err := archive.OpenDirectory(os.Getenv("SHAIYA_DATA"))
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()
	a, err := New(Config{}, arc)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	// Make settings explicit so a personal settings file cannot bias comparisons.
	a.gameMenu.Settings = ui.DefaultClientSettings()
	switch variant {
	case "default":
	case "geometry-medium":
		a.gameMenu.Settings.Details = 1
	case "geometry-low":
		a.gameMenu.Settings.Details = 0
	case "effects":
		a.gameMenu.Settings.SkillEffects = true
	case "no-effects":
		a.gameMenu.Settings.SkillEffects = false
	case "no-glow":
		a.gameMenu.Settings.Glow = 0
	case "no-shadow":
		a.gameMenu.Settings.Shadows = false
	case "no-water":
		a.gameMenu.Settings.Water = false
	default:
		t.Fatalf("unknown measurement variant %q", variant)
	}
	base := sceneName
	switch sceneName {
	case "crowd-200", "crowd-400", "crowd-network", "crowd-effects", "water-visible", "dungeon":
		base = "world-stream"
	}
	if err = a.Preview(base); err != nil {
		t.Fatal(err)
	}
	if sceneName == "dungeon" {
		a.releaseWorld()
		a.state = game.New()
		a.state.MapID, a.state.SelfID = 15, 1
		a.state.Phase = game.PhaseInWorld
		a.state.Self.X, a.state.Self.Y, a.state.Self.Z = 175, 7.6, 26
		a.state.Slots[0] = proto.CharacterSlot{ID: 1, Race: 0, Class: 0, Height: 2, Name: "Baseline"}
		a.screen = screenWorld
		a.enterWorld()
		a.wv.facing = render.V3{0, 0, 1}
		a.wv.resetCamera()
	}
	if sceneName == "water-visible" {
		v := a.wv
		found := false
		for x := float32(20); x < float32(v.world.WLD.MapSize)-20 && !found; x += 10 {
			for z := float32(20); z < float32(v.world.WLD.MapSize)-20; z += 10 {
				if floor, deep := v.world.WaterFloor(x, z); deep && floor < -8 && v.world.Height(x, z+10) < -8 {
					v.place(render.V3{x, 0, z})
					v.eye = render.V3{x, 6, z - 10}
					v.facing = render.V3{0, 0, 1}
					found = true
					break
				}
			}
		}
		if !found {
			t.Fatal("no verified deep-water viewpoint")
		}
	}
	var ids []uint32
	if sceneName == "crowd-200" || sceneName == "crowd-400" || sceneName == "crowd-network" || sceneName == "crowd-effects" {
		count := 200
		if sceneName == "crowd-400" {
			count = 400
		}
		a.state.Entities = map[uint32]*game.Entity{}
		for i := 0; i < count; i++ {
			id := uint32(i + 100)
			x := a.wv.pos[0] + float32(i%20-10)*3
			z := a.wv.pos[2] + float32(i/20-3)*4
			if i%5 == 0 {
				z += 120
			} // Also exercise loaded off-screen/distant actors.
			y := a.wv.world.Height(x, z)
			e := &game.Entity{ID: id, X: x, Y: y, Z: z}
			switch i % 10 {
			case 0, 1, 2, 3, 4:
				e.Kind = game.KindUser
				e.Shape = &proto.UserShape{ID: id, Race: uint8(i % 4), Gender: uint8(i / 4 % 2), Class: [4][3]uint8{{0, 1, 5}, {2, 3, 4}, {2, 3, 4}, {0, 1, 2}}[i%4][i/8%3], Height: 2, Name: fmt.Sprintf("Actor%03d", i)}
				if i%2 == 0 {
					e.Shape.Equipment[5] = [3]uint8{1, 1, 0}
				}
			case 5, 6, 7:
				e.Kind = game.KindMob
				e.TypeID = []uint16{1, 2, 3, 2009}[i%4]
				e.MobHeightKnown = true
			default:
				e.Kind = game.KindNpc
				e.NpcType = 1
				e.TypeID = 6
			}
			a.state.Entities[id] = e
			ids = append(ids, id)
		}
		a.wv.eye = a.wv.pos.Add(render.V3{0, 9, -18})
	}
	// Freeze camera across all iterations; animation time still advances at 60 Hz.
	eye, target := a.wv.eye, a.wv.pos
	if sceneName == "dungeon" {
		target = a.wv.target()
	}
	if sceneName == "water-visible" {
		target = eye.Add(render.V3{0, -3, 10})
	}
	if len(ids) > 0 {
		target = a.wv.pos.Add(render.V3{0, 2, 12})
	}
	camera := func() { a.wv.eye = eye; a.wv.cameraReady = true; a.wv.cameraEye = eye; a.wv.cameraTarget = target }
	effectTick := 0
	effects := func() {
		if sceneName != "crowd-effects" {
			return
		}
		if effectTick%120 == 0 {
			for i := 0; i < 12; i++ {
				a.startSkillImpactAt(ids[(i*7+21)%len(ids)], "Effect/CA_ice.EFT", 0, 0)
			}
		}
		effectTick++
	}
	var input platform.Input
	run := func(dt float64) {
		if dt > 0 {
			effects()
		}
		a.Update(&input, width, height, dt)
		camera()
		a.Draw(width, height)
	}
	cold := make([]float64, 0, 500)
	for i := 0; i < 500; i++ {
		started := time.Now()
		run(0)
		gl.Finish()
		cold = append(cold, float64(time.Since(started))/1e6)
		if a.SceneReady() {
			break
		}
		if i == 499 {
			t.Fatal("scene did not finish loading")
		}
	}
	for range 120 {
		run(1.0 / 60)
		gl.Finish()
	}
	setupMS := float64(time.Since(startSetup)) / 1e6
	if !a.SceneReady() {
		t.Fatal("scene lost readiness")
	}
	if a.sceneGlowFailed && a.gameMenu.Settings.Glow > 0 {
		t.Fatal("glow failed")
	}
	if sceneName == "water-visible" {
		if !a.wv.world.WaterVisible(a.scene3d.View, a.scene3d.Proj) {
			t.Fatal("water fixture did not show a water surface")
		}
		wantsReflection := a.gameMenu.Settings.Water && a.gameMenu.Settings.Details > 0
		if (a.wv.draws.Reflected > 0 || a.wv.draws.ReflectionReused) != wantsReflection {
			t.Fatalf("water fixture reflection draws = %d, expected reflection = %t", a.wv.draws.Reflected, wantsReflection)
		}
	}
	loaded := 0
	for _, v := range a.wv.actors {
		if v.character != nil && v.character.PlayingClip("000_normal") {
			loaded++
		}
	}
	for _, v := range a.wv.mobs {
		if v.actor != nil {
			loaded++
		}
	}
	for _, v := range a.wv.npcs {
		if v.actor != nil {
			loaded++
		}
	}
	if len(ids) > 0 && loaded != len(ids) {
		t.Fatalf("loaded %d/%d stress actors", loaded, len(ids))
	}
	// Prebuild two movement batches. Enqueue itself is outside the pump timer;
	// decoder/reducer/effect-dispatch work executes through the real pumpWorld.
	var batches [2][][]byte
	if sceneName == "crowd-network" {
		for phase := range batches {
			for _, id := range ids {
				e := a.state.Entities[id]
				op := proto.CharacterMove
				if e.Kind == game.KindMob {
					op = proto.GXMobMove
				}
				if e.Kind == game.KindNpc {
					op = proto.MapNpcMove
				}
				w := proto.NewWriter(op)
				w.U32(id)
				w.U8(1)
				if e.Kind == game.KindUser {
					w.U16(0)
				}
				w.F32(e.X + float32(phase)*.1)
				w.F32(e.Y)
				w.F32(e.Z)
				if e.Kind == game.KindMob {
					w.U32(100)
				}
				p := w.Bytes()
				probe := game.New()
				if _, err := probe.Apply(p); err != nil {
					t.Fatal("invalid traffic fixture", err)
				}
				batches[phase] = append(batches[phase], p)
			}
		}
	}
	ch := make(chan []byte, len(ids))
	incoming := &session.World{Packets: ch}
	// Timestamp queries are associated with exact frame IDs and read only once.
	// A three-frame fence ring bounds queue depth; it is NOT a synchronization-free
	// triple-buffer assumption. Waits are reported separately from submission.
	var queries [16][2]uint32
	pending := [16]int{}
	for i := range pending {
		pending[i] = -1
		gl.CreateQueries(gl.TIMESTAMP, 2, &queries[i][0])
	}
	defer func() {
		for i := range queries {
			gl.DeleteQueries(2, &queries[i][0])
		}
	}()
	var fences [3]uintptr
	defer func() {
		for _, f := range fences {
			if f != 0 {
				gl.DeleteSync(f)
			}
		}
	}()
	gpu := make([]float64, frames)
	for i := range gpu {
		gpu[i] = -1
	}
	collect := func() {
		for i, frame := range pending {
			if frame < 0 {
				continue
			}
			var available int32
			gl.GetQueryObjectiv(queries[i][1], gl.QUERY_RESULT_AVAILABLE, &available)
			if available == 0 {
				continue
			}
			var begin, end uint64
			gl.GetQueryObjectui64v(queries[i][0], gl.QUERY_RESULT, &begin)
			gl.GetQueryObjectui64v(queries[i][1], gl.QUERY_RESULT, &end)
			if end >= begin {
				gpu[frame] = float64(end-begin) / 1e6
			}
			pending[i] = -1
		}
	}
	update, draw, pump, wait, wall, calls := make([]float64, frames), make([]float64, frames), make([]float64, frames), make([]float64, frames), make([]float64, frames), make([]float64, frames)
	profilePath := os.Getenv("GX_MEASURE_CPU_PROFILE")
	stopProfile := func() {}
	if profilePath != "" {
		f, err := os.Create(profilePath)
		if err != nil {
			t.Fatal(err)
		}
		if err = pprof.StartCPUProfile(f); err != nil {
			f.Close()
			t.Fatal(err)
		}
		done := false
		stopProfile = func() {
			if !done {
				pprof.StopCPUProfile()
				f.Close()
				done = true
			}
		}
		defer stopProfile()
	}
	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	startedRun := time.Now()
	packets := 0
	for frame := range frames {
		frameStart := time.Now()
		slot := frame % len(fences)
		started := time.Now()
		if f := fences[slot]; f != 0 {
			result := gl.ClientWaitSync(f, gl.SYNC_FLUSH_COMMANDS_BIT, 10_000_000_000)
			if result != gl.ALREADY_SIGNALED && result != gl.CONDITION_SATISFIED {
				t.Fatalf("GPU fence wait result %#x", result)
			}
			gl.DeleteSync(f)
			fences[slot] = 0
		}
		wait[frame] = float64(time.Since(started)) / 1e6
		collect()
		// Ten batches per simulated second, equivalent to a 2,000 payload/s stress
		// input. This measures decoded-packet pumping, not sockets/TLS/server latency.
		if sceneName == "crowd-network" && frame%6 == 0 {
			if len(ch) != 0 {
				t.Fatal("traffic fixture did not drain before the next burst")
			}
			for _, p := range batches[(frame/6)%2] {
				ch <- p
			}
		}
		if pendingPackets := len(ch); pendingPackets != 0 {
			a.world = incoming
			started = time.Now()
			a.pumpWorld()
			pump[frame] = float64(time.Since(started)) / 1e6
			a.world = nil
			packets += pendingPackets - len(ch)
		}
		started = time.Now()
		effects()
		a.Update(&input, width, height, 1.0/60)
		update[frame] = float64(time.Since(started)) / 1e6
		camera()
		q := frame % len(queries)
		timed := pending[q] < 0
		if timed {
			gl.QueryCounter(queries[q][0], gl.TIMESTAMP)
		}
		started = time.Now()
		a.Draw(width, height)
		draw[frame] = float64(time.Since(started)) / 1e6
		if timed {
			gl.QueryCounter(queries[q][1], gl.TIMESTAMP)
			pending[q] = frame
		}
		fences[slot] = gl.FenceSync(gl.SYNC_GPU_COMMANDS_COMPLETE, 0)
		gl.Flush()
		calls[frame] = float64(render.DrawCalls())
		wall[frame] = float64(time.Since(frameStart)) / 1e6
	}
	gl.Finish()
	elapsedMS := float64(time.Since(startedRun)) / 1e6
	stopProfile()
	collect()
	runtime.ReadMemStats(&memAfter)
	if code := gl.GetError(); code != gl.NO_ERROR {
		t.Fatalf("measurement GL error %#x", code)
	}
	validGPU := make([]float64, 0, frames)
	for _, v := range gpu {
		if v >= 0 {
			validGPU = append(validGPU, v)
		}
	}
	submit := make([]float64, frames)
	for i := range submit {
		submit[i] = update[i] + draw[i] + pump[i]
	}
	output := os.Getenv("GX_MEASURE_OUTPUT")
	if output == "" {
		output = t.TempDir()
	}
	if err = os.MkdirAll(output, 0700); err != nil {
		t.Fatal(err)
	}
	// Untimed diagnostic screenshot: not a deterministic pixel-parity assertion.
	capture := filepath.Join(output, "frame.png")
	if err = render.SavePNG(capture, width, height); err != nil {
		t.Fatal(err)
	}
	rss, _ := platform.ResidentBytes()
	visible := 0
	frustum := render.ViewFrustum(a.scene3d.View, a.scene3d.VisibilityProjection())
	for _, actor := range a.wv.actors {
		if actor.character != nil && actor.character.Visible(frustum, a.scene3d.ClipPlane) {
			visible++
		}
	}
	for _, actor := range a.wv.mobs {
		if actor.actor != nil && actor.actor.Visible(frustum, a.scene3d.ClipPlane) {
			visible++
		}
	}
	for _, actor := range a.wv.npcs {
		if actor.actor != nil && actor.actor.Visible(frustum, a.scene3d.ClipPlane) {
			visible++
		}
	}
	effectActive := a.wv.effect != nil && !a.wv.effect.Done()
	result := map[string]any{
		"schema": 1, "scene": sceneName, "variant": variant, "width": width, "height": height, "frames": frames,
		"renderer": renderer, "vendor": vendor, "gl_version": version, "go_version": runtime.Version(), "gomaxprocs": runtime.GOMAXPROCS(0),
		"mode":     "hidden window, no swap/vsync, maximum 3 GPU frames in flight, fixed simulation dt=1/60",
		"profiled": profilePath != "", "camera_eye": eye, "camera_target": target, "settings": a.gameMenu.Settings,
		"map_id": a.state.MapID, "entities": len(a.state.Entities), "loaded_remote_actors": loaded, "visible_remote_actors_end": visible, "preview_effect_active_end": effectActive, "resolved_effects_end": len(a.wv.impacts), "packets_applied": packets, "simulated_packet_rate": float64(packets) / float64(frames) * 60,
		"setup_and_warmup_ms": setupMS, "loading_frame_wall": clientMeasureSummary(cold),
		"update": clientMeasureSummary(update), "draw_submission": clientMeasureSummary(draw), "packet_pump": clientMeasureSummary(pump),
		"cpu_submission": clientMeasureSummary(submit), "gpu_queue_wait": clientMeasureSummary(wait), "wall_frame": clientMeasureSummary(wall), "gpu_elapsed": clientMeasureSummary(validGPU),
		"throughput_wall_ms_per_frame": elapsedMS / float64(frames), "draw_calls_mean": clientMeasureSummary(calls).Mean,
		"allocs_per_frame": float64(memAfter.Mallocs-memBefore.Mallocs) / float64(frames), "allocated_bytes_per_frame": float64(memAfter.TotalAlloc-memBefore.TotalAlloc) / float64(frames),
		"gc_cycles": memAfter.NumGC - memBefore.NumGC, "rss_end_bytes": rss, "gpu_owned_bytes": render.GPUBytes(), "missing_asset_keys": len(a.assets.Missing),
		"samples_ms": map[string][]float64{"update": update, "draw": draw, "pump": pump, "gpu_wait": wait, "wall": wall, "gpu": gpu},
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(output, "measurement.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("scene=%s variant=%s CPU submission p50=%.3fms p95=%.3fms p99=%.3fms GPU samples=%d/%d remote actors=%d packets=%d", sceneName, variant, clientMeasureSummary(submit).P50, clientMeasureSummary(submit).P95, clientMeasureSummary(submit).P99, len(validGPU), frames, loaded, packets)
}

// BenchmarkClientMovementReducers isolates decoding/state mutation from OpenGL,
// socket scheduling and the UI dispatcher. One operation is 200 decoded payloads.
func BenchmarkClientMovementReducers(b *testing.B) {
	state := game.New()
	state.Phase = game.PhaseInWorld
	packets := make([][]byte, 200)
	bytes := 0
	for i := range packets {
		id := uint32(i + 100)
		state.Entities[id] = &game.Entity{Kind: game.KindUser, ID: id}
		w := proto.NewWriter(proto.CharacterMove)
		w.U32(id)
		w.U8(1)
		w.U16(0)
		w.F32(float32(i))
		w.F32(10)
		w.F32(50)
		packets[i] = w.Bytes()
		bytes += len(packets[i])
	}
	b.ReportAllocs()
	b.SetBytes(int64(bytes))
	for b.Loop() {
		for _, p := range packets {
			if _, err := state.Apply(p); err != nil {
				b.Fatal(err)
			}
		}
	}
	b.ReportMetric(float64(len(packets)), "payloads/op")
}
