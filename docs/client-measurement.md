# Reproducible client CPU and rendering measurements

Use the explicit integration measurement to establish a baseline before changing
rendering or animation. It runs the real application with locally owned loose
assets in a hidden OpenGL 4.5 window. It does not connect to a game server, change
production code paths, edit a database, or replace personal client settings.

## Run on Linux

The measurements use a private runner with a working Linux desktop GL context and local client data.

Close other graphics-heavy applications first. `quiet` is an operator-provided
classification, not a claim of exclusive hardware ownership. The runner captures
NVIDIA temperature/clocks/utilization and process telemetry when available;
telemetry failure does not select another renderer. On shared hardware use
`--load-label shared` and do not treat its GPU timings as a standalone baseline.
No other applications are stopped automatically.

The runner builds `internal/app` with the `integration` tag and invokes only
`TestClientPerformanceBaseline` with explicit opt-in. Without both
`GX_MEASURE_CLIENT=1` and `SHAIYA_TEST_RENDER=1` that test skips. Preferences use a
fresh temporary directory (XDG on Linux, AppData on Windows); ordinary gameplay
settings and external font/data resources are preserved. The Python orchestration
and hardware collection command above target Linux; this is not native Windows
performance evidence.

Each repetition shuffles a fixed case list using a reproducible seed. Each case
uses a fresh process/context, explicit default graphics settings, frozen camera
and advancing 1/60-second simulation steps. Warmup is complete scene readiness at
zero animation step followed by 120 advancing frames. Missing stress actors,
missing idle animations, failed glow or an unverified visible-water reflection
fail the fixture rather than silently producing an easier scene.

## Cases

| Case | Workload |
|---|---|
| `world-stream` | Original map 1 field and local character |
| `world-mobs` | Existing monster preview |
| `world-players` | Existing remote-player preview |
| `world-effect` | Existing animated EFT preview |
| `world-hud` | Existing HUD/target/skill preview |
| `water-visible` | Real map 1 deep-water camera; asserts a reflection target and water visibility |
| `dungeon` | Original map 15 at the existing movement/visible-marker integration position |
| `crowd-200` | 100 users, 60 monsters and 40 NPCs, including distant/off-screen actors |
| `crowd-400` | 200 users, 120 monsters and 80 NPCs |
| `crowd-effects` | 200 actors plus repeated real EFT impacts, using the normal 32-impact cap |
| `crowd-network` | 200 actors plus predecoded movement batches through the real packet pump |

Crowd users vary race, sex, valid class/body variants and sword equipment.
Monsters use four original definitions; NPCs use a known original model. The
fixture validates that every expected actor loads. Positions are fixed, some
intentionally beyond the near camera view. These synthetic scenes are not a
specific live town, player-count capacity test or server load test.

For live PvP stress tests, run [`/bots`](wiki/gm-commands.md#pvp-stress-test-bots)
on a loopback GX World. Historical offline bot results remain in the
[performance report](performance.md).

`crowd-network` injects 200 payloads every six simulated frames: 2,000 payloads
per simulated second. Payload construction is outside the pump timer; decoding,
state reduction and confirmation dispatch use `pumpWorld`. Update/draw then run
without a mock socket object, avoiding accidental outbound requests. There is
no measured TCP/TLS/server round trip. The pure
`BenchmarkClientMovementReducers` separately isolates 200 user movement reductions
per operation, including explicit wire decoding but excluding UI/GL/socket work.

## Metrics and timing boundaries

| JSON field | Meaning |
|---|---|
| `update` | Wall time spent in `App.Update` |
| `packet_pump` | Wall time in injected packet pumping; zero on non-batch frames |
| `draw_submission` | Wall time spent in `App.Draw`, including driver calls and any waits inside them |
| `cpu_submission` | Per-frame sum of pump + update + draw wall time; not exclusive CPU execution time |
| `gpu_queue_wait` | Explicit wait before reusing a slot in the three-frame fence ring |
| `wall_frame` | Whole measured iteration, including query collection, queue waits, injection, update, draw and flush |
| `gpu_elapsed` | Distinct GPU timestamp pair around each draw; includes elapsed submission gaps/preemption, not only shader-active time |
| `throughput_wall_ms_per_frame` | Total loop elapsed time including final GPU drain, divided by frame count |
| `loading_frame_wall` | Synchronous scene-readiness draw samples, outside steady-state statistics |
| `setup_and_warmup_ms` | Context/app/scene initialization plus readiness and fixed warmup |
| `allocs_per_frame`, `allocated_bytes_per_frame`, `gc_cycles` | Process Go allocation deltas over the measured span |
| `rss_end_bytes`, `gpu_owned_bytes` | Final resident bytes and renderer-owned GPU storage estimate; neither is a full driver allocation/peak-memory measure |
| `visible_remote_actors_end`, `loaded_remote_actors` | End-of-run primary-frustum visibility and successfully loaded fixture actors |

Statistics report mean, nearest-rank p50/p95/p99 and maximum. Raw per-frame samples
are retained. GPU samples are associated with their exact frame and consumed only
once, unlike repeatedly reading a last-available HUD value. Timestamp queries are
read only after availability is reported. Explicit fences bound this context to
three in-flight frames; rotating slots alone is not proof that GPU work finished.
There is no `glFinish` inside steady-state per-frame rendering. Warmup and final
sample drain deliberately synchronize outside the submission measurements.

The window does not swap/present and VSync is disabled for this test. Animation
advances at fixed simulated 60 Hz even when throughput is faster/slower. The HUD
FPS text therefore does not establish presented FPS. Queue waits may move work
out of a particular GL call compared with an unbounded submission loop; compare
the same harness/mode, and inspect submission, wait and whole-loop metrics together.

## Ablation and profiles

The measurement cases select diagnostic settings explicitly.
Set `GX_MEASURE_NO_BINDLESS=1` or `GX_MEASURE_NO_DRAW_PARAMETERS=1` to
force either graphics fallback in the integration test. Both can be set
in the same run.

Variants are `default`, `effects`, `no-effects`, `no-glow`, `no-shadow` and `no-water`.
Use `crowd-400:no-effects` to measure remote player batching. Disabling a feature
is a diagnostic ablation, not an implemented optimization or a permissible visual
change. Effects overlap and savings must not be added together.

`--profile 'world-stream,crowd-200,crowd-network,crowd-effects'` runs additional,
separate profiled passes after unprofiled measurements. Profiles cover only the
steady loop, exclude loading/capture/report generation, and use at least 3,000
frames. They are not pooled into baseline timing statistics. `go tool pprof`
reports actual CPU samples; time attributed to `runtime.cgocall` does not isolate
the Go/C transition cost and can include driver execution or spin waiting.

To run just the pure decoder/state benchmark using a built test binary:

```sh
work/client-measure/baseline/app.test -test.run '^$' \
  -test.bench '^BenchmarkClientMovementReducers$' -test.benchmem -test.benchtime=2s -test.count=3
```

## Artifacts and interpretation

Output contains manifest/production-source/fixture/binary hashes, per-run logs and settings,
`measurement.json` with raw samples, diagnostic `frame.png`, aggregate
`summary.json`/`report.md` and optional CPU profiles/top tables. Keep original
asset screenshots, private paths and machine telemetry outside Git. A screenshot
is an inspection artifact; random/time-dependent effects are not asserted to be
pixel-identical across runs.

Application loading is cold per process, but filesystem/driver caches can be warm.
This does not measure cold boot, actual server traffic, audio playback, active
input latency, present/compositor behavior or minimized-window scheduling. Use a
separate live/presented experiment for those questions. Do not infer maximum
player capacity, guaranteed FPS or an optimization's gain from one preview.

Before/after comparisons need the same production source identity except for the
intended change, assets/settings/camera/workload, repeated interleaved cases and
comparable GPU/CPU contention. Investigate wide between-run medians instead of
reporting false precision. Start with the measured bottleneck and preserve visual
and gameplay parity; rerecord this baseline after relevant production changes.
