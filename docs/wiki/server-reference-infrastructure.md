# Infrastructure and command package API

> Generated from Go source by `tools/generate_server_reference.py`; do not edit by hand.

[Package responsibilities](server-packages.md) · [Server architecture](server-architecture.md)

Every production Go file is listed, including OS-specific alternatives. Signatures include private helpers to make ownership and call boundaries inspectable. Declaration comments are source documentation, not an independent guarantee of behavior; the architecture guide explains transaction, concurrency and failure contracts.

## server/ai

Embedded/default Lua catalog, actor state machines, host API contract, bounded script execution and optional editable catalog loading.

Tests: [runtime_test.go](../../server/ai/runtime_test.go).

### methods.go

[server/ai/methods.go](../../server/ai/methods.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`methods`](../../server/ai/methods.go#L3)<br>`var methods` | See the source definition and the subsystem ownership contract. |

### runtime.go

[server/ai/runtime.go](../../server/ai/runtime.go#L1)

Package ai executes the authoritative mob policies in modern, pure-Go Lua.

| Declaration | Responsibility / source comment |
|---|---|
| [`sources`](../../server/ai/runtime.go#L21)<br>`var sources embed.FS` | See the source definition and the subsystem ownership contract. |
| [`Host`](../../server/ai/runtime.go#L25)<br>`type Host` | Host supplies world primitives. Lua owns decisions; unavailable primitives must return an error rather than silently bypass an encounter rule. |
| [`Program`](../../server/ai/runtime.go#L29)<br>`type Program` | See the source definition and the subsystem ownership contract. |
| [`Load`](../../server/ai/runtime.go#L36)<br>`func Load(source fs.FS) (*Program, error)` | See the source definition and the subsystem ownership contract. |
| [`defaultOnce`](../../server/ai/runtime.go#L88)<br>`var defaultOnce sync.Once` | See the source definition and the subsystem ownership contract. |
| [`defaultProgram`](../../server/ai/runtime.go#L89)<br>`var defaultProgram *Program` | See the source definition and the subsystem ownership contract. |
| [`defaultError`](../../server/ai/runtime.go#L90)<br>`var defaultError error` | See the source definition and the subsystem ownership contract. |
| [`Default`](../../server/ai/runtime.go#L92)<br>`func Default() (*Program, error)` | See the source definition and the subsystem ownership contract. |
| [`VM`](../../server/ai/runtime.go#L97)<br>`type VM` | See the source definition and the subsystem ownership contract. |
| [`NewVM`](../../server/ai/runtime.go#L107)<br>`func (p *Program) NewVM() (*VM, error)` | See the source definition and the subsystem ownership contract. |
| [`Close`](../../server/ai/runtime.go#L170)<br>`func (v *VM) Close()` | See the source definition and the subsystem ownership contract. |
| [`module`](../../server/ai/runtime.go#L171)<br>`func (v *VM) module(name string) (rt.Value, error)` | See the source definition and the subsystem ownership contract. |
| [`call`](../../server/ai/runtime.go#L199)<br>`func (v *VM) call(fn rt.Value, args ...rt.Value) (result rt.Value, err error)` | See the source definition and the subsystem ownership contract. |
| [`luaValue`](../../server/ai/runtime.go#L212)<br>`func luaValue(value any) (rt.Value, error)` | See the source definition and the subsystem ownership contract. |
| [`Actor`](../../server/ai/runtime.go#L235)<br>`type Actor` | See the source definition and the subsystem ownership contract. |
| [`NewActor`](../../server/ai/runtime.go#L240)<br>`func (v *VM) NewActor(id uint8, host Host) (*Actor, error)` | See the source definition and the subsystem ownership contract. |
| [`Event`](../../server/ai/runtime.go#L256)<br>`func (a *Actor) Event(host Host, event string, args ...any) error` | See the source definition and the subsystem ownership contract. |
| [`AcquiresProximity`](../../server/ai/runtime.go#L274)<br>`func (p *Program) AcquiresProximity(id uint8) bool` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type Host interface {
	Call(string, []any) (any, error)
}
```

```go
type Program struct {
	acquire [256]bool
	mu      sync.Mutex
	source  map[string][]byte
	units   map[string]*code.Unit
}
```

```go
type VM struct {
	mu      sync.Mutex
	program *Program
	r       *rt.Runtime
	modules map[string]rt.Value
	catalog rt.Value
	host    Host
	context rt.Value
}
```

```go
type Actor struct {
	vm    *VM
	value rt.Value
}
```

</details>

## server/cmd/gx_console

Terminal start/shutdown controls plus local account and TLS administration commands.

### main.go

[server/cmd/gx_console/main.go](../../server/cmd/gx_console/main.go#L1)

Command gx_console controls the detached GX services from a terminal.

Direct GX dependencies: `server/internal/admin`, `server/internal/control`.

| Declaration | Responsibility / source comment |
|---|---|
| [`main`](../../server/cmd/gx_console/main.go#L15)<br>`func main()` | See the source definition and the subsystem ownership contract. |
| [`run`](../../server/cmd/gx_console/main.go#L22)<br>`func run() error` | See the source definition and the subsystem ownership contract. |

## server/cmd/gx_export_catalogs



### main.go

[server/cmd/gx_export_catalogs/main.go](../../server/cmd/gx_export_catalogs/main.go#L1)

gx_export_catalogs exports the retired SQLite definition tables without
modifying the source database. GX World reads the resulting files at boot.

Direct GX dependencies: `internal/blessingrules`, `internal/catalogfiles`, `internal/data/sdata`, `internal/enchantrules`, `internal/itemboxes`, `internal/maprules`, `internal/mobaimessages`, `internal/npcquest`, `internal/teleportrunes`, `internal/worlddrops`.

| Declaration | Responsibility / source comment |
|---|---|
| [`main`](../../server/cmd/gx_export_catalogs/main.go#L32)<br>`func main()` | See the source definition and the subsystem ownership contract. |
| [`encode`](../../server/cmd/gx_export_catalogs/main.go#L58)<br>`func encode[T any](value T, readErr error, format func(T) []byte) ([]byte, error)` | See the source definition and the subsystem ownership contract. |
| [`export`](../../server/cmd/gx_export_catalogs/main.go#L65)<br>`func export(ctx context.Context, dbPath, npcFile, mapFile, mobFile string) (map[string][]byte, error)` | See the source definition and the subsystem ownership contract. |

## server/cmd/gx_login

Login handshake, bounded credential verification, handoff creation, server selection and diagnostics.

Tests: [login_integration_test.go](../../server/cmd/gx_login/login_integration_test.go).

### main.go

[server/cmd/gx_login/main.go](../../server/cmd/gx_login/main.go#L1)

Command gx_login is the Shaiya 5.4 login server: RSA/AES key
exchange, credential check through the original
PS_UserData.dbo.usp_Try_GameLogin_Taiwan, server list and the session
handoff to gx_world.

Direct GX dependencies: `internal/authwire`, `internal/itemstack`, `internal/transport`, `server/internal/config`, `server/internal/controlplane`, `server/internal/crypto`, `server/internal/db`, `server/internal/logx`, `server/internal/netx`, `server/internal/packet`, `server/internal/session`.

| Declaration | Responsibility / source comment |
|---|---|
| [`nextConnection`](../../server/cmd/gx_login/main.go#L31)<br>`var nextConnection atomic.Int64` | See the source definition and the subsystem ownership contract. |
| [`loginByIP`](../../server/cmd/gx_login/main.go#L32)<br>`var loginByIP` | See the source definition and the subsystem ownership contract. |
| [`loginByAccount`](../../server/cmd/gx_login/main.go#L33)<br>`var loginByAccount` | See the source definition and the subsystem ownership contract. |
| [`main`](../../server/cmd/gx_login/main.go#L35)<br>`func main()` | See the source definition and the subsystem ownership contract. |
| [`run`](../../server/cmd/gx_login/main.go#L41)<br>`func run() error` | See the source definition and the subsystem ownership contract. |
| [`handshake`](../../server/cmd/gx_login/main.go#L103)<br>`func handshake() []byte` | See the source definition and the subsystem ownership contract. |
| [`loginReply`](../../server/cmd/gx_login/main.go#L114)<br>`func loginReply(result uint8, userUID uint32, permission uint8, identity [16]byte) []byte` | See the source definition and the subsystem ownership contract. |
| [`serverList`](../../server/cmd/gx_login/main.go#L125)<br>`func serverList() []byte` | See the source definition and the subsystem ownership contract. |
| [`selectServerReply`](../../server/cmd/gx_login/main.go#L136)<br>`func selectServerReply(status uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`client`](../../server/cmd/gx_login/main.go#L147)<br>`type client` | See the source definition and the subsystem ownership contract. |
| [`write`](../../server/cmd/gx_login/main.go#L157)<br>`func (c *client) write(frame []byte, label string) error` | See the source definition and the subsystem ownership contract. |
| [`handle`](../../server/cmd/gx_login/main.go#L170)<br>`func handle(conn net.Conn, log *logx.Tag) error` | See the source definition and the subsystem ownership contract. |
| [`frame`](../../server/cmd/gx_login/main.go#L200)<br>`func (c *client) frame(frame []byte, remote string) error` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type client struct {
	conn      net.Conn
	log       *logx.Tag
	send      *crypto.AESCTR
	recv      *crypto.AESCTR
	material  [32]byte
	userUID   uint32
	sessionID int64
}
```

</details>

## server/cmd/gx_mapbuild

Offline collision/navigation generation from locally owned loose assets; not a runtime network service.

### main.go

[server/cmd/gx_mapbuild/main.go](../../server/cmd/gx_mapbuild/main.go#L1)

gx_mapbuild generates server-only collision and navigation data from loose assets.

Direct GX dependencies: `internal/data/archive`, `internal/data/world`, `internal/mapgeometry`, `internal/mapnav`.

| Declaration | Responsibility / source comment |
|---|---|
| [`main`](../../server/cmd/gx_mapbuild/main.go#L23)<br>`func main()` | See the source definition and the subsystem ownership contract. |

## server/cmd/gx_server_control

Standalone desktop service controls, live logs and reward-rate settings.

### logs.go

[server/cmd/gx_server_control/logs.go](../../server/cmd/gx_server_control/logs.go#L1)

Direct GX dependencies: `server/internal/control`.

| Declaration | Responsibility / source comment |
|---|---|
| [`logHistory`](../../server/cmd/gx_server_control/logs.go#L18)<br>`const logHistory = 1000` | See the source definition and the subsystem ownership contract. |
| [`logView`](../../server/cmd/gx_server_control/logs.go#L20)<br>`type logView` | See the source definition and the subsystem ownership contract. |
| [`newLogView`](../../server/cmd/gx_server_control/logs.go#L29)<br>`func newLogView(window fyne.Window, ctx context.Context, root string) *logView` | See the source definition and the subsystem ownership contract. |
| [`poll`](../../server/cmd/gx_server_control/logs.go#L79)<br>`func (v *logView) poll(ctx context.Context, reader *control.LogReader)` | See the source definition and the subsystem ownership contract. |
| [`refresh`](../../server/cmd/gx_server_control/logs.go#L115)<br>`func (v *logView) refresh()` | See the source definition and the subsystem ownership contract. |
| [`follow`](../../server/cmd/gx_server_control/logs.go#L151)<br>`func (v *logView) follow()` | See the source definition and the subsystem ownership contract. |
| [`newLogRow`](../../server/cmd/gx_server_control/logs.go#L157)<br>`func newLogRow() fyne.CanvasObject` | See the source definition and the subsystem ownership contract. |
| [`renderRow`](../../server/cmd/gx_server_control/logs.go#L168)<br>`func (v *logView) renderRow(id widget.ListItemID, object fyne.CanvasObject)` | See the source definition and the subsystem ownership contract. |
| [`logColumns`](../../server/cmd/gx_server_control/logs.go#L190)<br>`type logColumns` | See the source definition and the subsystem ownership contract. |
| [`logWidths`](../../server/cmd/gx_server_control/logs.go#L192)<br>`var logWidths` | See the source definition and the subsystem ownership contract. |
| [`Layout`](../../server/cmd/gx_server_control/logs.go#L194)<br>`func (*logColumns) Layout(objects []fyne.CanvasObject, size fyne.Size)` | See the source definition and the subsystem ownership contract. |
| [`MinSize`](../../server/cmd/gx_server_control/logs.go#L205)<br>`func (*logColumns) MinSize(objects []fyne.CanvasObject) fyne.Size` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type logView struct {
	content                   fyne.CanvasObject
	list                      *widget.List
	footer, empty             *widget.Label
	service, level, readError string
	all, visible              []control.LogEntry
	paused                    bool
}
```

```go
type logColumns struct{}
```

</details>

### main.go

[server/cmd/gx_server_control/main.go](../../server/cmd/gx_server_control/main.go#L1)

Command gx_server_control is the standalone local GX Server Control desktop app.

Direct GX dependencies: `server/internal/control`.

| Declaration | Responsibility / source comment |
|---|---|
| [`serviceRow`](../../server/cmd/gx_server_control/main.go#L22)<br>`type serviceRow` | See the source definition and the subsystem ownership contract. |
| [`dashboard`](../../server/cmd/gx_server_control/main.go#L28)<br>`type dashboard` | See the source definition and the subsystem ownership contract. |
| [`main`](../../server/cmd/gx_server_control/main.go#L42)<br>`func main()` | See the source definition and the subsystem ownership contract. |
| [`build`](../../server/cmd/gx_server_control/main.go#L68)<br>`func (d *dashboard) build()` | See the source definition and the subsystem ownership contract. |
| [`strong`](../../server/cmd/gx_server_control/main.go#L116)<br>`func strong(text string) *widget.Label` | See the source definition and the subsystem ownership contract. |
| [`statusCell`](../../server/cmd/gx_server_control/main.go#L120)<br>`func statusCell(dot *canvas.Circle, label *widget.Label) fyne.CanvasObject` | See the source definition and the subsystem ownership contract. |
| [`poll`](../../server/cmd/gx_server_control/main.go#L124)<br>`func (d *dashboard) poll()` | See the source definition and the subsystem ownership contract. |
| [`refresh`](../../server/cmd/gx_server_control/main.go#L148)<br>`func (d *dashboard) refresh()` | See the source definition and the subsystem ownership contract. |
| [`setEnabled`](../../server/cmd/gx_server_control/main.go#L222)<br>`func setEnabled(button *widget.Button, enabled bool)` | See the source definition and the subsystem ownership contract. |
| [`run`](../../server/cmd/gx_server_control/main.go#L230)<br>`func (d *dashboard) run(action control.Action, targets []control.Service)` | See the source definition and the subsystem ownership contract. |
| [`apply`](../../server/cmd/gx_server_control/main.go#L260)<br>`func (d *dashboard) apply(action control.Action, targets []control.Service)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type serviceRow struct {
	status, endpoint     *widget.Label
	dot                  *canvas.Circle
	start, stop, restart *widget.Button
}
```

```go
type dashboard struct {
	window                         fyne.Window
	manager                        *control.Manager
	ctx                            context.Context
	rows                           map[control.Service]serviceRow
	status                         map[control.Service]control.Status
	overall, players, notice       *widget.Label
	overallDot                     *canvas.Circle
	start, stop, restart, settings *widget.Button
	busy                           bool
	checked                        bool
	revision                       atomic.Uint64
}
```

</details>

### settings.go

[server/cmd/gx_server_control/settings.go](../../server/cmd/gx_server_control/settings.go#L1)

Direct GX dependencies: `server/internal/config`.

| Declaration | Responsibility / source comment |
|---|---|
| [`openSettings`](../../server/cmd/gx_server_control/settings.go#L14)<br>`func (d *dashboard) openSettings()` | See the source definition and the subsystem ownership contract. |

### theme.go

[server/cmd/gx_server_control/theme.go](../../server/cmd/gx_server_control/theme.go#L1)

Direct GX dependencies: `server/internal/control`.

| Declaration | Responsibility / source comment |
|---|---|
| [`controlTheme`](../../server/cmd/gx_server_control/theme.go#L14)<br>`type controlTheme` | See the source definition and the subsystem ownership contract. |
| [`base`](../../server/cmd/gx_server_control/theme.go#L17)<br>`var base` | See the source definition and the subsystem ownership contract. |
| [`surface`](../../server/cmd/gx_server_control/theme.go#L18)<br>`var surface` | See the source definition and the subsystem ownership contract. |
| [`border`](../../server/cmd/gx_server_control/theme.go#L19)<br>`var border` | See the source definition and the subsystem ownership contract. |
| [`ink`](../../server/cmd/gx_server_control/theme.go#L20)<br>`var ink` | See the source definition and the subsystem ownership contract. |
| [`quiet`](../../server/cmd/gx_server_control/theme.go#L21)<br>`var quiet` | See the source definition and the subsystem ownership contract. |
| [`teal`](../../server/cmd/gx_server_control/theme.go#L22)<br>`var teal` | See the source definition and the subsystem ownership contract. |
| [`healthy`](../../server/cmd/gx_server_control/theme.go#L23)<br>`var healthy` | See the source definition and the subsystem ownership contract. |
| [`warning`](../../server/cmd/gx_server_control/theme.go#L24)<br>`var warning` | See the source definition and the subsystem ownership contract. |
| [`fault`](../../server/cmd/gx_server_control/theme.go#L25)<br>`var fault` | See the source definition and the subsystem ownership contract. |
| [`faultSurface`](../../server/cmd/gx_server_control/theme.go#L26)<br>`var faultSurface` | See the source definition and the subsystem ownership contract. |
| [`rgb`](../../server/cmd/gx_server_control/theme.go#L29)<br>`func rgb(r, g, b uint8) color.NRGBA` | See the source definition and the subsystem ownership contract. |
| [`colors`](../../server/cmd/gx_server_control/theme.go#L31)<br>`var colors` | See the source definition and the subsystem ownership contract. |
| [`Color`](../../server/cmd/gx_server_control/theme.go#L48)<br>`func (t controlTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color` | See the source definition and the subsystem ownership contract. |
| [`statusColor`](../../server/cmd/gx_server_control/theme.go#L55)<br>`func statusColor(state control.State) color.Color` | See the source definition and the subsystem ownership contract. |
| [`panel`](../../server/cmd/gx_server_control/theme.go#L68)<br>`func panel(body fyne.CanvasObject) fyne.CanvasObject` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type controlTheme struct{ fyne.Theme }
```

</details>

## server/cmd/gx_world

World bootstrap/lease, definition/map loading, owned scheduler, bounded sessions and graceful storage shutdown.

Tests: [combat_integration_test.go](../../server/cmd/gx_world/combat_integration_test.go), [fixtures_integration_test.go](../../server/cmd/gx_world/fixtures_integration_test.go), [itemmall_integration_test.go](../../server/cmd/gx_world/itemmall_integration_test.go), [journey_integration_test.go](../../server/cmd/gx_world/journey_integration_test.go), [load_integration_test.go](../../server/cmd/gx_world/load_integration_test.go), [party_integration_test.go](../../server/cmd/gx_world/party_integration_test.go), [quest_atomic_integration_test.go](../../server/cmd/gx_world/quest_atomic_integration_test.go), [quest_integration_test.go](../../server/cmd/gx_world/quest_integration_test.go), [spell_fluidity_integration_test.go](../../server/cmd/gx_world/spell_fluidity_integration_test.go), [storage_load_integration_test.go](../../server/cmd/gx_world/storage_load_integration_test.go), [trade_integration_test.go](../../server/cmd/gx_world/trade_integration_test.go), [traffic_probe_test.go](../../server/cmd/gx_world/traffic_probe_test.go), [view_integration_test.go](../../server/cmd/gx_world/view_integration_test.go), [world_integration_test.go](../../server/cmd/gx_world/world_integration_test.go).

### main.go

[server/cmd/gx_world/main.go](../../server/cmd/gx_world/main.go#L1)

Command gx_world is the Shaiya 5.4 game server on the original
EP4.5 tables translated to SQLite: character select, world simulation (maps,
NPCs, mobs, combat, skills), items, trade and the item mall.

Direct GX dependencies: `internal/transport`, `server/internal/config`, `server/internal/controlplane`, `server/internal/db`, `server/internal/logx`, `server/internal/netx`, `server/internal/world`.

| Declaration | Responsibility / source comment |
|---|---|
| [`nextConnection`](../../server/cmd/gx_world/main.go#L25)<br>`var nextConnection atomic.Int64` | See the source definition and the subsystem ownership contract. |
| [`main`](../../server/cmd/gx_world/main.go#L27)<br>`func main()` | See the source definition and the subsystem ownership contract. |
| [`run`](../../server/cmd/gx_world/main.go#L33)<br>`func run() error` | See the source definition and the subsystem ownership contract. |

## server/internal/admin

Local account creation/password reset and private TLS provisioning; never a public admin endpoint.

Tests: [certificates_test.go](../../server/internal/admin/certificates_test.go).

### admin.go

[server/internal/admin/admin.go](../../server/internal/admin/admin.go#L1)

Package admin supplies local operator commands without a web admin service.

Direct GX dependencies: `server/internal/config`, `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`secret`](../../server/internal/admin/admin.go#L17)<br>`func secret() (string, error)` | See the source definition and the subsystem ownership contract. |
| [`Run`](../../server/internal/admin/admin.go#L39)<br>`func Run(args []string) error` | See the source definition and the subsystem ownership contract. |

### certificates.go

[server/internal/admin/certificates.go](../../server/internal/admin/certificates.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`certificates`](../../server/internal/admin/certificates.go#L21)<br>`func certificates(args []string) error` | See the source definition and the subsystem ownership contract. |

## server/internal/config

Validated environment policy, optional atomic runtime.env parsing, paths and SQLite DSN.

Tests: [config_test.go](../../server/internal/config/config_test.go), [rates_test.go](../../server/internal/config/rates_test.go).

### config.go

[server/internal/config/config.go](../../server/internal/config/config.go#L1)

Package config resolves runtime settings from the environment, falling
back to .local/runtime.env for local overrides. SQLite needs no credentials.

| Declaration | Responsibility / source comment |
|---|---|
| [`ServerID`](../../server/internal/config/config.go#L15)<br>`const ServerID = 1` | See the source definition and the subsystem ownership contract. |
| [`LoadRuntimeEnv`](../../server/internal/config/config.go#L19)<br>`func LoadRuntimeEnv() error` | LoadRuntimeEnv reads KEY=VALUE lines from .local/runtime.env into the process environment without overriding variables that are already set. |
| [`ReadRuntimeEnv`](../../server/internal/config/config.go#L45)<br>`func ReadRuntimeEnv(root string) (map[string]string, error)` | ReadRuntimeEnv validates a runtime.env without changing the process environment. |
| [`RuntimeEnvPath`](../../server/internal/config/config.go#L57)<br>`func RuntimeEnvPath(root string) string` | See the source definition and the subsystem ownership contract. |
| [`parseRuntimeEnv`](../../server/internal/config/config.go#L59)<br>`func parseRuntimeEnv(scanner *bufio.Scanner) (map[string]string, error)` | See the source definition and the subsystem ownership contract. |
| [`LoadRuntimeEnvFrom`](../../server/internal/config/config.go#L89)<br>`func LoadRuntimeEnvFrom(root string) error` | LoadRuntimeEnvFrom is LoadRuntimeEnv for the checkout at root. |
| [`env`](../../server/internal/config/config.go#L105)<br>`func env(key, fallback string) string` | See the source definition and the subsystem ownership contract. |
| [`DatabasePath`](../../server/internal/config/config.go#L113)<br>`func DatabasePath() string` | DatabasePath is the mutable runtime copy, never the versioned seed. |
| [`DatabaseSeed`](../../server/internal/config/config.go#L114)<br>`func DatabaseSeed() string` | See the source definition and the subsystem ownership contract. |
| [`DSN`](../../server/internal/config/config.go#L117)<br>`func DSN(_ string) (string, error)` | DSN applies connection settings to every connection, including both servers. |
| [`ReadDSN`](../../server/internal/config/config.go#L122)<br>`func ReadDSN() (string, error)` | ReadDSN opens a WAL reader that cannot mutate gameplay data. |
| [`sqliteDSN`](../../server/internal/config/config.go#L124)<br>`func sqliteDSN(readOnly bool) (string, error)` | See the source definition and the subsystem ownership contract. |
| [`DataDir`](../../server/internal/config/config.go#L148)<br>`func DataDir() string` | DataDir holds map data (data/maps/*.svmap). |
| [`PublicIP`](../../server/internal/config/config.go#L151)<br>`func PublicIP() string` | PublicIP is the world address sent to clients after server selection. |
| [`LoginListen`](../../server/internal/config/config.go#L153)<br>`func LoginListen() string` | See the source definition and the subsystem ownership contract. |
| [`WorldListen`](../../server/internal/config/config.go#L154)<br>`func WorldListen() string` | See the source definition and the subsystem ownership contract. |
| [`DefaultCombatDEXFirst`](../../server/internal/config/config.go#L157)<br>`const DefaultCombatDEXFirst = 500` | See the source definition and the subsystem ownership contract. |
| [`DefaultCombatDEXSecond`](../../server/internal/config/config.go#L158)<br>`const DefaultCombatDEXSecond = 800` | See the source definition and the subsystem ownership contract. |
| [`CombatDEXThresholds`](../../server/internal/config/config.go#L162)<br>`func CombatDEXThresholds() (uint32, uint32, error)` | CombatDEXThresholds reads the two points where DEX gains change rate. |
| [`resolvePaths`](../../server/internal/config/config.go#L177)<br>`func resolvePaths(root string)` | Launchers and direct binaries resolve storage/data against their checkout. |
| [`sqliteURIPath`](../../server/internal/config/config.go#L189)<br>`func sqliteURIPath(path string) string` | A Windows drive is part of the URI path, never its authority. Without the leading slash net/url emits file://C:/..., which SQLite rejects. |
| [`FallDamageEnabled`](../../server/internal/config/config.go#L197)<br>`func FallDamageEnabled() bool` | FallDamageEnabled defaults to the GX landing rule; an explicit false opts out. |

### gameini.go

[server/internal/config/gameini.go](../../server/internal/config/gameini.go#L1)

Direct GX dependencies: `internal/gameini`.

| Declaration | Responsibility / source comment |
|---|---|
| [`GameINI`](../../server/internal/config/gameini.go#L5)<br>`type GameINI` | See the source definition and the subsystem ownership contract. |
| [`ReadGameINI`](../../server/internal/config/gameini.go#L7)<br>`func ReadGameINI(path string) (GameINI, error)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type GameINI = gameini.File
```

</details>

### policy.go

[server/internal/config/policy.go](../../server/internal/config/policy.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`NetworkPolicy`](../../server/internal/config/policy.go#L11)<br>`type NetworkPolicy` | NetworkPolicy is immutable for a connection. Startup and acceptance validate the same bounded overrides; a malformed value never silently disables a cap. |
| [`Network`](../../server/internal/config/policy.go#L16)<br>`func Network(service string) (NetworkPolicy, error)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type NetworkPolicy struct {
	Connections, PerIP, IPAttempts, AccountAttempts, InputFrames, InputBytes, OutputPackets, OutputBytes int
	Establishments, Economy                                                                              int
}
```

</details>

### rates.go

[server/internal/config/rates.go](../../server/internal/config/rates.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`MobEXPKey`](../../server/internal/config/rates.go#L11)<br>`const MobEXPKey = "RAPTURE_MOB_EXP_PERCENT"` | See the source definition and the subsystem ownership contract. |
| [`QuestEXPKey`](../../server/internal/config/rates.go#L12)<br>`const QuestEXPKey = "RAPTURE_QUEST_EXP_PERCENT"` | See the source definition and the subsystem ownership contract. |
| [`PvPKillKey`](../../server/internal/config/rates.go#L13)<br>`const PvPKillKey = "RAPTURE_PVP_KILL_MULTIPLIER"` | See the source definition and the subsystem ownership contract. |
| [`RewardRates`](../../server/internal/config/rates.go#L16)<br>`type RewardRates` | See the source definition and the subsystem ownership contract. |
| [`DefaultRewardRates`](../../server/internal/config/rates.go#L22)<br>`func DefaultRewardRates() RewardRates` | See the source definition and the subsystem ownership contract. |
| [`RuntimeValue`](../../server/internal/config/rates.go#L27)<br>`func RuntimeValue(settings map[string]string, key string) string` | RuntimeValue gives explicitly inherited values precedence over the local file. |
| [`ParseRewardRates`](../../server/internal/config/rates.go#L35)<br>`func ParseRewardRates(mob, quest, kills string) (RewardRates, error)` | ParseRewardRates accepts whole percentages (100 = 1×) and whole PvP multipliers. |
| [`ReadRewardRates`](../../server/internal/config/rates.go#L58)<br>`func ReadRewardRates(settings map[string]string) (RewardRates, error)` | See the source definition and the subsystem ownership contract. |
| [`ScaleEXP`](../../server/internal/config/rates.go#L63)<br>`func ScaleEXP(amount, percent uint32) uint32` | ScaleEXP applies a percentage after native reward rounding, with saturation. |

<details><summary>Type definitions and owned fields</summary>

```go
type RewardRates struct {
	MobEXPPercent   uint32
	QuestEXPPercent uint32
	PvPKills        uint32
}
```

</details>

## server/internal/control

Shared desktop and terminal lifecycle, executable-scoped status and shutdown, detached launch, reward settings and log tailing.

Tests: [logs_test.go](../../server/internal/control/logs_test.go), [settings_test.go](../../server/internal/control/settings_test.go).

### control.go

[server/internal/control/control.go](../../server/internal/control/control.go#L1)

Package control manages the detached Login and World processes for a single GX installation.

Direct GX dependencies: `server/internal/config`, `server/internal/controlplane`, `server/internal/filelock`.

| Declaration | Responsibility / source comment |
|---|---|
| [`Service`](../../server/internal/control/control.go#L21)<br>`type Service` | See the source definition and the subsystem ownership contract. |
| [`Login`](../../server/internal/control/control.go#L24)<br>`const Login = "gx_login"` | See the source definition and the subsystem ownership contract. |
| [`World`](../../server/internal/control/control.go#L25)<br>`const World = "gx_world"` | See the source definition and the subsystem ownership contract. |
| [`Services`](../../server/internal/control/control.go#L28)<br>`func Services() []Service` | See the source definition and the subsystem ownership contract. |
| [`Label`](../../server/internal/control/control.go#L30)<br>`func (s Service) Label() string` | See the source definition and the subsystem ownership contract. |
| [`State`](../../server/internal/control/control.go#L37)<br>`type State` | See the source definition and the subsystem ownership contract. |
| [`Stopped`](../../server/internal/control/control.go#L40)<br>`const Stopped = "Stopped"` | See the source definition and the subsystem ownership contract. |
| [`Starting`](../../server/internal/control/control.go#L41)<br>`const Starting = "Starting"` | See the source definition and the subsystem ownership contract. |
| [`Ready`](../../server/internal/control/control.go#L42)<br>`const Ready = "Ready"` | See the source definition and the subsystem ownership contract. |
| [`Stopping`](../../server/internal/control/control.go#L43)<br>`const Stopping = "Stopping"` | See the source definition and the subsystem ownership contract. |
| [`Failed`](../../server/internal/control/control.go#L44)<br>`const Failed = "Failed"` | See the source definition and the subsystem ownership contract. |
| [`Unknown`](../../server/internal/control/control.go#L45)<br>`const Unknown = "Unknown"` | See the source definition and the subsystem ownership contract. |
| [`Status`](../../server/internal/control/control.go#L48)<br>`type Status` | See the source definition and the subsystem ownership contract. |
| [`Action`](../../server/internal/control/control.go#L56)<br>`type Action` | See the source definition and the subsystem ownership contract. |
| [`Start`](../../server/internal/control/control.go#L59)<br>`const Start = "Start"` | See the source definition and the subsystem ownership contract. |
| [`Stop`](../../server/internal/control/control.go#L60)<br>`const Stop = "Stop"` | See the source definition and the subsystem ownership contract. |
| [`Restart`](../../server/internal/control/control.go#L61)<br>`const Restart = "Restart"` | See the source definition and the subsystem ownership contract. |
| [`Manager`](../../server/internal/control/control.go#L64)<br>`type Manager` | See the source definition and the subsystem ownership contract. |
| [`New`](../../server/internal/control/control.go#L70)<br>`func New(root string) *Manager` | See the source definition and the subsystem ownership contract. |
| [`Installation`](../../server/internal/control/control.go#L75)<br>`func Installation() (*Manager, error)` | Installation opens the portable root containing the controller's build directory. |
| [`Root`](../../server/internal/control/control.go#L86)<br>`func (m *Manager) Root() string` | See the source definition and the subsystem ownership contract. |
| [`exeName`](../../server/internal/control/control.go#L88)<br>`func exeName(service Service) string` | See the source definition and the subsystem ownership contract. |
| [`binary`](../../server/internal/control/control.go#L96)<br>`func (m *Manager) binary(service Service) string` | See the source definition and the subsystem ownership contract. |
| [`lock`](../../server/internal/control/control.go#L100)<br>`func (m *Manager) lock(ctx context.Context) (func(), error)` | See the source definition and the subsystem ownership contract. |
| [`current`](../../server/internal/control/control.go#L110)<br>`func (m *Manager) current(ctx context.Context, service Service, found []int) Status` | See the source definition and the subsystem ownership contract. |
| [`Snapshot`](../../server/internal/control/control.go#L132)<br>`func (m *Manager) Snapshot(ctx context.Context) map[Service]Status` | See the source definition and the subsystem ownership contract. |
| [`Apply`](../../server/internal/control/control.go#L160)<br>`func (m *Manager) Apply(ctx context.Context, action Action, targets []Service, progress func(Service, Status)) error` | Apply serializes GUI and CLI changes to this installation. Restart starts only services whose old processes were verified stopped. |
| [`start`](../../server/internal/control/control.go#L218)<br>`func (m *Manager) start(ctx context.Context, service Service) error` | See the source definition and the subsystem ownership contract. |
| [`stop`](../../server/internal/control/control.go#L276)<br>`func (m *Manager) stop(ctx context.Context, service Service) error` | See the source definition and the subsystem ownership contract. |
| [`lastStderr`](../../server/internal/control/control.go#L312)<br>`func lastStderr(path string) string` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type Service string
```

```go
type State string
```

```go
type Status struct {
	State    State
	PID      int
	Endpoint string
	Players  int
	Detail   string
}
```

```go
type Action string
```

```go
type Manager struct {
	root    string
	mu      sync.Mutex
	failure map[Service]string
}
```

</details>

### detach_other.go

[server/internal/control/detach_other.go](../../server/internal/control/detach_other.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`detach`](../../server/internal/control/detach_other.go#L12)<br>`func detach(cmd *exec.Cmd)` | detach starts the server in its own session so closing the console or app (or its terminal) does not take the server down. |

### detach_windows.go

[server/internal/control/detach_windows.go](../../server/internal/control/detach_windows.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`createNoWindow`](../../server/internal/control/detach_windows.go#L10)<br>`const createNoWindow = 0x08000000` | See the source definition and the subsystem ownership contract. |
| [`detach`](../../server/internal/control/detach_windows.go#L13)<br>`func detach(cmd *exec.Cmd)` | detach starts the server without opening a console window. |

### logs.go

[server/internal/control/logs.go](../../server/internal/control/logs.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`logReadLimit`](../../server/internal/control/logs.go#L15)<br>`const logReadLimit = 256 << 10` | See the source definition and the subsystem ownership contract. |
| [`LogEntry`](../../server/internal/control/logs.go#L17)<br>`type LogEntry` | See the source definition and the subsystem ownership contract. |
| [`logPosition`](../../server/internal/control/logs.go#L24)<br>`type logPosition` | See the source definition and the subsystem ownership contract. |
| [`LogReader`](../../server/internal/control/logs.go#L31)<br>`type LogReader` | LogReader tracks complete JSON records without holding files open between polls, so the Windows services can rotate their logs. |
| [`NewLogReader`](../../server/internal/control/logs.go#L36)<br>`func NewLogReader(root string) *LogReader` | See the source definition and the subsystem ownership contract. |
| [`Poll`](../../server/internal/control/logs.go#L40)<br>`func (r *LogReader) Poll() ([]LogEntry, error)` | See the source definition and the subsystem ownership contract. |
| [`readLog`](../../server/internal/control/logs.go#L57)<br>`func readLog(path string, service Service, pos logPosition) (logPosition, []LogEntry, error)` | See the source definition and the subsystem ownership contract. |
| [`readLogRecords`](../../server/internal/control/logs.go#L91)<br>`func readLogRecords(f *os.File, service Service, size, offset int64) (int64, []LogEntry, error)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type LogEntry struct {
	Time    time.Time
	Level   string
	Service Service
	Message string
}
```

```go
type logPosition struct {
	file   os.FileInfo
	offset int64
}
```

```go
type LogReader struct {
	root      string
	positions map[Service]logPosition
}
```

</details>

### process_linux.go

[server/internal/control/process_linux.go](../../server/internal/control/process_linux.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`processes`](../../server/internal/control/process_linux.go#L9)<br>`func (m *Manager) processes() (map[Service][]int, error)` | See the source definition and the subsystem ownership contract. |

### process_windows.go

[server/internal/control/process_windows.go](../../server/internal/control/process_windows.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`processes`](../../server/internal/control/process_windows.go#L12)<br>`func (m *Manager) processes() (map[Service][]int, error)` | See the source definition and the subsystem ownership contract. |

### settings.go

[server/internal/control/settings.go](../../server/internal/control/settings.go#L1)

Direct GX dependencies: `server/internal/config`.

| Declaration | Responsibility / source comment |
|---|---|
| [`RewardSettings`](../../server/internal/control/settings.go#L16)<br>`func (m *Manager) RewardSettings() (map[string]string, map[string]bool, error)` | RewardSettings returns the editable local values and tracks inherited overrides. |
| [`SaveRewardSettings`](../../server/internal/control/settings.go#L36)<br>`func (m *Manager) SaveRewardSettings(ctx context.Context, mob, quest, kills string) error` | See the source definition and the subsystem ownership contract. |
| [`replaceSetting`](../../server/internal/control/settings.go#L94)<br>`func replaceSetting(content, key, value, eol string) string` | Update the last assignment, which is the one the runtime parser uses. |

## server/internal/controlplane

Loopback health/metrics, authenticated readiness and shutdown, optional protected profiling and private PID/executable descriptors.

Tests: [controlplane_test.go](../../server/internal/controlplane/controlplane_test.go).

### controlplane.go

[server/internal/controlplane/controlplane.go](../../server/internal/controlplane/controlplane.go#L1)

Package controlplane owns loopback diagnostics and authenticated local shutdown.

Direct GX dependencies: `server/internal/telemetry`.

| Declaration | Responsibility / source comment |
|---|---|
| [`descriptor`](../../server/internal/controlplane/controlplane.go#L25)<br>`type descriptor` | See the source definition and the subsystem ownership contract. |
| [`Server`](../../server/internal/controlplane/controlplane.go#L31)<br>`type Server` | See the source definition and the subsystem ownership contract. |
| [`ServiceStatus`](../../server/internal/controlplane/controlplane.go#L41)<br>`type ServiceStatus` | ServiceStatus contains the operational data exposed to the local controller. |
| [`loopback`](../../server/internal/controlplane/controlplane.go#L46)<br>`func loopback(addr string) bool` | See the source definition and the subsystem ownership contract. |
| [`Start`](../../server/internal/controlplane/controlplane.go#L51)<br>`func Start(ctx context.Context, service, addr string, cancel context.CancelFunc, status func() ServiceStatus, observe ...func()) (*Server, error)` | See the source definition and the subsystem ownership contract. |
| [`Address`](../../server/internal/controlplane/controlplane.go#L162)<br>`func (s *Server) Address() string` | See the source definition and the subsystem ownership contract. |
| [`Close`](../../server/internal/controlplane/controlplane.go#L163)<br>`func (s *Server) Close()` | See the source definition and the subsystem ownership contract. |
| [`localRequest`](../../server/internal/controlplane/controlplane.go#L180)<br>`func localRequest(ctx context.Context, root, service string, pid int, executable, method, path string) (*http.Response, error)` | See the source definition and the subsystem ownership contract. |
| [`localClient`](../../server/internal/controlplane/controlplane.go#L208)<br>`var localClient` | See the source definition and the subsystem ownership contract. |
| [`ReadStatus`](../../server/internal/controlplane/controlplane.go#L215)<br>`func ReadStatus(ctx context.Context, root, service string, pid int, executable string) (ServiceStatus, error)` | ReadStatus validates the descriptor against the running installation. |
| [`RequestStop`](../../server/internal/controlplane/controlplane.go#L230)<br>`func RequestStop(ctx context.Context, root, service string, pid int, executable string) error` | RequestStop validates the PID and executable before sending the private token. |

<details><summary>Type definitions and owned fields</summary>

```go
type descriptor struct {
	PID        int    `json:"pid"`
	Executable string `json:"executable"`
	Address    string `json:"address"`
	Token      string `json:"token"`
}
```

```go
type Server struct {
	http       *http.Server
	listener   net.Listener
	done       chan struct{}
	path       string
	descriptor descriptor
	once       sync.Once
}
```

```go
type ServiceStatus struct {
	Endpoint string `json:"endpoint"`
	Players  int    `json:"players"`
}
```

</details>

## server/internal/crypto

Legacy login RSA/HMAC, AES-CTR and expanded world XOR compatibility; authenticated remote protection belongs to TLS.

Tests: [crypto_test.go](../../server/internal/crypto/crypto_test.go).

### crypto.go

[server/internal/crypto/crypto.go](../../server/internal/crypto/crypto.go#L1)

Package crypto implements the Shaiya 5.4 login/world transport security:
an RSA key exchange at login, HMAC-SHA256 key derivation, a little-endian
counter AES-128-CTR stream and the world handshake's expanded XOR.

| Declaration | Responsibility / source comment |
|---|---|
| [`exponent`](../../server/internal/crypto/crypto.go#L15)<br>`var exponent` | See the source definition and the subsystem ownership contract. |
| [`privateExponentBE`](../../server/internal/crypto/crypto.go#L23)<br>`var privateExponentBE` | See the source definition and the subsystem ownership contract. |
| [`modulusBE`](../../server/internal/crypto/crypto.go#L33)<br>`var modulusBE` | See the source definition and the subsystem ownership contract. |
| [`reversed`](../../server/internal/crypto/crypto.go#L43)<br>`func reversed(b []byte) []byte` | See the source definition and the subsystem ownership contract. |
| [`LoginPublicKey`](../../server/internal/crypto/crypto.go#L51)<br>`func LoginPublicKey() (exp [64]byte, modulusLE [128]byte)` | LoginPublicKey returns the exponent and the little-endian modulus sent in the 0xA101 handshake. |
| [`leInt`](../../server/internal/crypto/crypto.go#L56)<br>`func leInt(b []byte) *big.Int` | See the source definition and the subsystem ownership contract. |
| [`leBytes`](../../server/internal/crypto/crypto.go#L58)<br>`func leBytes(v *big.Int) []byte` | See the source definition and the subsystem ownership contract. |
| [`DecryptLoginRSA`](../../server/internal/crypto/crypto.go#L70)<br>`func DecryptLoginRSA(encrypted []byte) []byte` | DecryptLoginRSA decrypts the client's little-endian RSA block. The client's secret is always 128 bytes (0x401160 only clears the top bit of the last byte), so the result keeps the modulus width: dropping high zero bytes would change the HMAC key and fail about 1 login in 128. |
| [`DeriveLoginAESMaterial`](../../server/internal/crypto/crypto.go#L81)<br>`func DeriveLoginAESMaterial(decrypted []byte) [32]byte` | DeriveLoginAESMaterial is HMAC-SHA256(key=decrypted, msg=modulus LE): the first 16 bytes are the AES key, the last 16 the IV. |
| [`DeriveWorldAESIV`](../../server/internal/crypto/crypto.go#L90)<br>`func DeriveWorldAESIV(loginIV [16]byte) [16]byte` | DeriveWorldAESIV is SHA-256(login IV)[:16]. |
| [`ExpandedXOR`](../../server/internal/crypto/crypto.go#L98)<br>`func ExpandedXOR(input []byte) []byte` | ExpandedXOR applies the world handshake's SHA-256 chained XOR stream. |
| [`AESCTR`](../../server/internal/crypto/crypto.go#L116)<br>`type AESCTR` | AESCTR is AES-128 in counter mode with a little-endian counter that starts at byte 0 (Go's cipher.NewCTR increments big-endian). |
| [`NewAESCTR`](../../server/internal/crypto/crypto.go#L123)<br>`func NewAESCTR(key, iv [16]byte) *AESCTR` | See the source definition and the subsystem ownership contract. |
| [`Transform`](../../server/internal/crypto/crypto.go#L132)<br>`func (c *AESCTR) Transform(input []byte) []byte` | Transform encrypts or decrypts input (the operation is symmetric). |

<details><summary>Type definitions and owned fields</summary>

```go
type AESCTR struct {
	block   cipher.Block
	counter [16]byte
	mask    [16]byte
	pos     int
}
```

</details>

## server/internal/filelock

OS-released cross-process startup/ownership locks for Linux and Windows.

Tests: [lock_test.go](../../server/internal/filelock/lock_test.go).

### lock.go

[server/internal/filelock/lock.go](../../server/internal/filelock/lock.go#L1)

Package filelock provides crash-released process locks for runtime setup.

| Declaration | Responsibility / source comment |
|---|---|
| [`Acquire`](../../server/internal/filelock/lock.go#L12)<br>`func Acquire(ctx context.Context, path string) (func(), error)` | Acquire returns a release function. The lock file is intentionally retained: removing its inode would let different processes lock different files. |

### lock_unix.go

[server/internal/filelock/lock_unix.go](../../server/internal/filelock/lock_unix.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`tryLock`](../../server/internal/filelock/lock_unix.go#L11)<br>`func tryLock(f *os.File) (bool, error)` | See the source definition and the subsystem ownership contract. |
| [`unlock`](../../server/internal/filelock/lock_unix.go#L18)<br>`func unlock(f *os.File)` | See the source definition and the subsystem ownership contract. |

### lock_windows.go

[server/internal/filelock/lock_windows.go](../../server/internal/filelock/lock_windows.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`tryLock`](../../server/internal/filelock/lock_windows.go#L9)<br>`func tryLock(f *os.File) (bool, error)` | See the source definition and the subsystem ownership contract. |
| [`unlock`](../../server/internal/filelock/lock_windows.go#L16)<br>`func unlock(f *os.File)` | See the source definition and the subsystem ownership contract. |

## server/internal/logx

Concurrent structured JSON logs, session tags, bounded file rotation and opt-in debug output.

Tests: [logx_test.go](../../server/internal/logx/logx_test.go).

### logx.go

[server/internal/logx/logx.go](../../server/internal/logx/logx.go#L1)

Package logx writes structured JSON logs with bounded disk retention.

| Declaration | Responsibility / source comment |
|---|---|
| [`Logger`](../../server/internal/logx/logx.go#L15)<br>`type Logger` | See the source definition and the subsystem ownership contract. |
| [`New`](../../server/internal/logx/logx.go#L24)<br>`func New(name string) (*Logger, error)` | See the source definition and the subsystem ownership contract. |
| [`rotate`](../../server/internal/logx/logx.go#L51)<br>`func (l *Logger) rotate() error` | See the source definition and the subsystem ownership contract. |
| [`record`](../../server/internal/logx/logx.go#L79)<br>`func (l *Logger) record(level, message, session string, attrs ...any)` | See the source definition and the subsystem ownership contract. |
| [`Printf`](../../server/internal/logx/logx.go#L114)<br>`func (l *Logger) Printf(format string, args ...any)` | See the source definition and the subsystem ownership contract. |
| [`Event`](../../server/internal/logx/logx.go#L117)<br>`func (l *Logger) Event(message string, attrs ...any)` | See the source definition and the subsystem ownership contract. |
| [`Close`](../../server/internal/logx/logx.go#L118)<br>`func (l *Logger) Close() error` | See the source definition and the subsystem ownership contract. |
| [`Tagged`](../../server/internal/logx/logx.go#L128)<br>`func (l *Logger) Tagged(tag string) *Tag` | See the source definition and the subsystem ownership contract. |
| [`Tag`](../../server/internal/logx/logx.go#L130)<br>`type Tag` | See the source definition and the subsystem ownership contract. |
| [`SetTag`](../../server/internal/logx/logx.go#L135)<br>`func (t *Tag) SetTag(tag string)` | See the source definition and the subsystem ownership contract. |
| [`Tag`](../../server/internal/logx/logx.go#L136)<br>`func (t *Tag) Tag() string` | See the source definition and the subsystem ownership contract. |
| [`Printf`](../../server/internal/logx/logx.go#L142)<br>`func (t *Tag) Printf(format string, args ...any)` | See the source definition and the subsystem ownership contract. |
| [`Debugf`](../../server/internal/logx/logx.go#L145)<br>`func (t *Tag) Debugf(format string, args ...any)` | See the source definition and the subsystem ownership contract. |
| [`Errorf`](../../server/internal/logx/logx.go#L151)<br>`func (l *Logger) Errorf(format string, args ...any)` | See the source definition and the subsystem ownership contract. |
| [`Errorf`](../../server/internal/logx/logx.go#L154)<br>`func (t *Tag) Errorf(format string, args ...any)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type Logger struct {
	mu              sync.Mutex
	file            *os.File
	name, path      string
	debug           bool
	bytes, maxBytes int64
	backups         int
}
```

```go
type Tag struct {
	logger *Logger
	tag    atomic.Pointer[string]
}
```

</details>

## server/internal/netx

Server TCP/TLS admission, keepalive, bounded frame parsing, rate limiting and ordered bounded output.

Tests: [server_test.go](../../server/internal/netx/server_test.go), [writer_test.go](../../server/internal/netx/writer_test.go).

### netx.go

[server/internal/netx/netx.go](../../server/internal/netx/netx.go#L1)

Package netx holds TCP helpers shared by the login and world servers.

| Declaration | Responsibility / source comment |
|---|---|
| [`Tune`](../../server/internal/netx/netx.go#L14)<br>`func Tune(conn net.Conn)` | Tune enables TCP keepalive (idle clients are never reaped; vanished peers are) and disables Nagle so small packets such as skill acks go out immediately. |
| [`Limiter`](../../server/internal/netx/netx.go#L27)<br>`type Limiter` | Limiter caps simultaneous connections. |
| [`NewLimiter`](../../server/internal/netx/netx.go#L32)<br>`func NewLimiter(max int64) *Limiter` | See the source definition and the subsystem ownership contract. |
| [`Acquire`](../../server/internal/netx/netx.go#L34)<br>`func (l *Limiter) Acquire() bool` | See the source definition and the subsystem ownership contract. |
| [`Release`](../../server/internal/netx/netx.go#L46)<br>`func (l *Limiter) Release()` | See the source definition and the subsystem ownership contract. |
| [`Active`](../../server/internal/netx/netx.go#L47)<br>`func (l *Limiter) Active() int64` | See the source definition and the subsystem ownership contract. |
| [`Frames`](../../server/internal/netx/netx.go#L51)<br>`type Frames` | Frames splits a byte stream into length-prefixed Shaiya frames (length:u16 counts the whole frame). |
| [`Push`](../../server/internal/netx/netx.go#L55)<br>`func (f *Frames) Push(data []byte) ([][]byte, bool)` | Push appends received bytes and returns every complete frame. A frame length below 4 is a protocol error. |
| [`Abort`](../../server/internal/netx/netx.go#L79)<br>`func Abort(conn net.Conn)` | Abort interrupts pending TLS/TCP IO before closing, so error handling cannot block simulation while attempting a graceful TLS close notification. |

<details><summary>Type definitions and owned fields</summary>

```go
type Limiter struct {
	max    int64
	active atomic.Int64
}
```

```go
type Frames struct{ buf []byte }
```

</details>

### server.go

[server/internal/netx/server.go](../../server/internal/netx/server.go#L1)

Direct GX dependencies: `server/internal/telemetry`.

| Declaration | Responsibility / source comment |
|---|---|
| [`admitted`](../../server/internal/netx/server.go#L13)<br>`var admitted` | See the source definition and the subsystem ownership contract. |
| [`rejected`](../../server/internal/netx/server.go#L14)<br>`var rejected` | See the source definition and the subsystem ownership contract. |
| [`connections`](../../server/internal/netx/server.go#L15)<br>`var connections` | See the source definition and the subsystem ownership contract. |
| [`Serve`](../../server/internal/netx/server.go#L19)<br>`func Serve(ctx context.Context, listener net.Listener, capacity, perIP int, handle func(net.Conn), admission ...*RateLimiter) error` | Serve bounds connections globally and by source. Cancellation closes sockets and waits for handlers to finish persisting before the caller closes storage. |
| [`RateLimiter`](../../server/internal/netx/server.go#L110)<br>`type RateLimiter` | RateLimiter has bounded memory and fails closed when its source budget fills. Idle entries are reclaimed opportunistically, without background goroutines. |
| [`bucket`](../../server/internal/netx/server.go#L118)<br>`type bucket` | See the source definition and the subsystem ownership contract. |
| [`NewRateLimiter`](../../server/internal/netx/server.go#L123)<br>`func NewRateLimiter(capacity, burst int, interval time.Duration) *RateLimiter` | See the source definition and the subsystem ownership contract. |
| [`Allow`](../../server/internal/netx/server.go#L126)<br>`func (l *RateLimiter) Allow(key string) bool` | See the source definition and the subsystem ownership contract. |
| [`AllowN`](../../server/internal/netx/server.go#L127)<br>`func (l *RateLimiter) AllowN(key string, units int) bool` | See the source definition and the subsystem ownership contract. |
| [`allow`](../../server/internal/netx/server.go#L128)<br>`func (l *RateLimiter) allow(key string, now time.Time) bool` | See the source definition and the subsystem ownership contract. |
| [`allowN`](../../server/internal/netx/server.go#L129)<br>`func (l *RateLimiter) allowN(key string, units int, now time.Time) bool` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type RateLimiter struct {
	mu       sync.Mutex
	entries  map[string]bucket
	capacity int
	burst    float64
	interval time.Duration
	sweep    time.Time
}
```

```go
type bucket struct {
	tokens float64
	at     time.Time
}
```

</details>

### writer.go

[server/internal/netx/writer.go](../../server/internal/netx/writer.go#L1)

Direct GX dependencies: `server/internal/telemetry`.

| Declaration | Responsibility / source comment |
|---|---|
| [`writerBytes`](../../server/internal/netx/writer.go#L12)<br>`var writerBytes` | See the source definition and the subsystem ownership contract. |
| [`writerOverflows`](../../server/internal/netx/writer.go#L13)<br>`var writerOverflows` | See the source definition and the subsystem ownership contract. |
| [`ErrBackpressure`](../../server/internal/netx/writer.go#L14)<br>`var ErrBackpressure` | See the source definition and the subsystem ownership contract. |
| [`Writer`](../../server/internal/netx/writer.go#L19)<br>`type Writer` | Writer preserves packet order without blocking simulation on a slow client. Its memory budget includes the packet currently being written. Overflow closes the connection rather than dropping authoritative messages silently. |
| [`NewWriter`](../../server/internal/netx/writer.go#L30)<br>`func NewWriter(conn net.Conn, packets, maxBytes int) *Writer` | See the source definition and the subsystem ownership contract. |
| [`Write`](../../server/internal/netx/writer.go#L35)<br>`func (w *Writer) Write(packet []byte) error` | See the source definition and the subsystem ownership contract. |
| [`failLocked`](../../server/internal/netx/writer.go#L58)<br>`func (w *Writer) failLocked(err error)` | See the source definition and the subsystem ownership contract. |
| [`run`](../../server/internal/netx/writer.go#L71)<br>`func (w *Writer) run()` | See the source definition and the subsystem ownership contract. |
| [`Close`](../../server/internal/netx/writer.go#L96)<br>`func (w *Writer) Close() error` | Close drains queued output with per-write deadlines. Call after producers stop. |

<details><summary>Type definitions and owned fields</summary>

```go
type Writer struct {
	mu       sync.Mutex
	conn     net.Conn
	queue    chan []byte
	done     chan struct{}
	bytes    int
	maxBytes int
	closed   bool
	err      error
}
```

</details>

## server/internal/packet

Server opcode constants and explicit little-endian frame writer; no Go struct layout on the wire.

### packet.go

[server/internal/packet/packet.go](../../server/internal/packet/packet.go#L1)

Package packet holds the Shaiya 5.4 opcodes and the little-endian frame
writer/readers shared by the login and world servers. Frame layout:
length:u16 (whole frame), opcode:u16, payload.

Direct GX dependencies: `internal/itemstack`, `internal/proto`, `internal/textwire`.

| Declaration | Responsibility / source comment |
|---|---|
| [`Opcode`](../../server/internal/packet/packet.go#L16)<br>`type Opcode` | See the source definition and the subsystem ownership contract. |
| [`CharacterList`](../../server/internal/packet/packet.go#L21)<br>`const CharacterList = 0x0101` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CreateCharacter`](../../server/internal/packet/packet.go#L22)<br>`const CreateCharacter = 0x0102` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`DeleteCharacter`](../../server/internal/packet/packet.go#L25)<br>`const DeleteCharacter = 0x0103` | Request charId:u32; response result:u8 (0=ok), charId:u32. Soft delete, restorable for 2 hours. |
| [`SelectCharacter`](../../server/internal/packet/packet.go#L26)<br>`const SelectCharacter = 0x0104` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RestoreCharacter`](../../server/internal/packet/packet.go#L28)<br>`const RestoreCharacter = 0x010F` | Same shape as DeleteCharacter. |
| [`RenameCharacter`](../../server/internal/packet/packet.go#L30)<br>`const RenameCharacter = 0x010E` | Request charId:u32, newName; response result:u8 with 1 = ok, 0 = fail. |
| [`CharacterDetails`](../../server/internal/packet/packet.go#L31)<br>`const CharacterDetails = 0x0105` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterItems`](../../server/internal/packet/packet.go#L32)<br>`const CharacterItems = 0x0106` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`Logout`](../../server/internal/packet/packet.go#L33)<br>`const Logout = 0x0107` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterSkills`](../../server/internal/packet/packet.go#L34)<br>`const CharacterSkills = 0x0108` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`AccountFaction`](../../server/internal/packet/packet.go#L35)<br>`const AccountFaction = 0x0109` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterActiveBuffs`](../../server/internal/packet/packet.go#L36)<br>`const CharacterActiveBuffs = 0x010A` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterSkillBar`](../../server/internal/packet/packet.go#L37)<br>`const CharacterSkillBar = 0x010B` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterEnteredMap`](../../server/internal/packet/packet.go#L38)<br>`const CharacterEnteredMap = 0x0201` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`AutoAttackStop`](../../server/internal/packet/packet.go#L39)<br>`const AutoAttackStop = 0x0212` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`SetMoney`](../../server/internal/packet/packet.go#L40)<br>`const SetMoney = 0x0213` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ExpGain`](../../server/internal/packet/packet.go#L41)<br>`const ExpGain = 0x0207` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PartyRoster`](../../server/internal/packet/packet.go#L42)<br>`const PartyRoster = 0x0B01` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PartyInvite`](../../server/internal/packet/packet.go#L43)<br>`const PartyInvite = 0x0B02` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PartyAnswer`](../../server/internal/packet/packet.go#L44)<br>`const PartyAnswer = 0x0B03` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PartyJoined`](../../server/internal/packet/packet.go#L45)<br>`const PartyJoined = 0x0B04` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PartyLeft`](../../server/internal/packet/packet.go#L46)<br>`const PartyLeft = 0x0B05` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PartyExpelled`](../../server/internal/packet/packet.go#L47)<br>`const PartyExpelled = 0x0B06` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PartyLeader`](../../server/internal/packet/packet.go#L48)<br>`const PartyLeader = 0x0B07` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PartyInviteName`](../../server/internal/packet/packet.go#L49)<br>`const PartyInviteName = 0x0B09` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PartyLeaveAlt`](../../server/internal/packet/packet.go#L50)<br>`const PartyLeaveAlt = 0x0B0D` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PartyBadge`](../../server/internal/packet/packet.go#L51)<br>`const PartyBadge = 0x0520` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidRoster`](../../server/internal/packet/packet.go#L52)<br>`const RaidRoster = 0x0B0B` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidJoined`](../../server/internal/packet/packet.go#L53)<br>`const RaidJoined = 0x0B0C` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidConvert`](../../server/internal/packet/packet.go#L54)<br>`const RaidConvert = 0x0B0E` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidDivision`](../../server/internal/packet/packet.go#L55)<br>`const RaidDivision = 0x0B0F` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidAutoJoin`](../../server/internal/packet/packet.go#L56)<br>`const RaidAutoJoin = 0x0B10` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidAutoJoinName`](../../server/internal/packet/packet.go#L57)<br>`const RaidAutoJoinName = 0x0B11` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidSwap`](../../server/internal/packet/packet.go#L58)<br>`const RaidSwap = 0x0B12` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidSubLeader`](../../server/internal/packet/packet.go#L59)<br>`const RaidSubLeader = 0x0B13` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidInvite`](../../server/internal/packet/packet.go#L60)<br>`const RaidInvite = 0x0B14` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidAnswer`](../../server/internal/packet/packet.go#L61)<br>`const RaidAnswer = 0x0B15` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidDisband`](../../server/internal/packet/packet.go#L62)<br>`const RaidDisband = 0x0B16` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidLeader`](../../server/internal/packet/packet.go#L63)<br>`const RaidLeader = 0x0B17` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidInviteName`](../../server/internal/packet/packet.go#L64)<br>`const RaidInviteName = 0x0B18` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidError`](../../server/internal/packet/packet.go#L65)<br>`const RaidError = 0x0B19` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidExpel`](../../server/internal/packet/packet.go#L66)<br>`const RaidExpel = 0x0B1A` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidPools`](../../server/internal/packet/packet.go#L67)<br>`const RaidPools = 0x0C0C` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PartyPools`](../../server/internal/packet/packet.go#L68)<br>`const PartyPools = 0x0C03` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PartyLevel`](../../server/internal/packet/packet.go#L69)<br>`const PartyLevel = 0x0C09` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PartyMove`](../../server/internal/packet/packet.go#L70)<br>`const PartyMove = 0x0C07` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. member position: u32 id, u16 map, 3 x f32 |
| [`RaidMove`](../../server/internal/packet/packet.go#L71)<br>`const RaidMove = 0x0C12` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PartyEffectAdd`](../../server/internal/packet/packet.go#L72)<br>`const PartyEffectAdd = 0x0C04` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. member buff: u32 id, u16 skill, u8 level |
| [`PartyEffectRemove`](../../server/internal/packet/packet.go#L73)<br>`const PartyEffectRemove = 0x0C05` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidEffectAdd`](../../server/internal/packet/packet.go#L74)<br>`const RaidEffectAdd = 0x0C0E` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RaidEffectRemove`](../../server/internal/packet/packet.go#L75)<br>`const RaidEffectRemove = 0x0C10` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ChatParty`](../../server/internal/packet/packet.go#L76)<br>`const ChatParty = 0x1105` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`Gatekeeper`](../../server/internal/packet/packet.go#L77)<br>`const Gatekeeper = 0x020C` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ShopBuy`](../../server/internal/packet/packet.go#L78)<br>`const ShopBuy = 0x0702` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`LevelUpSelf`](../../server/internal/packet/packet.go#L79)<br>`const LevelUpSelf = 0x0508` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`LevelUpOther`](../../server/internal/packet/packet.go#L80)<br>`const LevelUpOther = 0x051E` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`TargetMobGetState`](../../server/internal/packet/packet.go#L81)<br>`const TargetMobGetState = 0x0304` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`TargetMobHpUpdate`](../../server/internal/packet/packet.go#L82)<br>`const TargetMobHpUpdate = 0x0305` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`TargetClear`](../../server/internal/packet/packet.go#L83)<br>`const TargetClear = 0x030A` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`UseVehicle`](../../server/internal/packet/packet.go#L84)<br>`const UseVehicle = 0x0216` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. empty request; response ok:u8, isOnVehicle:u8 |
| [`ItemBagToBag`](../../server/internal/packet/packet.go#L85)<br>`const ItemBagToBag = 0x0204` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`AddItem`](../../server/internal/packet/packet.go#L86)<br>`const AddItem = 0x0205` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`RemoveItem`](../../server/internal/packet/packet.go#L87)<br>`const RemoveItem = 0x0206` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`SetStatPoint`](../../server/internal/packet/packet.go#L88)<br>`const SetStatPoint = 0x0208` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`StatsReset`](../../server/internal/packet/packet.go#L89)<br>`const StatsReset = 0x0214` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. success:u8, statPoint:u16, str,rec,int,wis,dex,luc:u16 |
| [`LearnNewSkill`](../../server/internal/packet/packet.go#L90)<br>`const LearnNewSkill = 0x0209` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterEnteredPortal`](../../server/internal/packet/packet.go#L91)<br>`const CharacterEnteredPortal = 0x020A` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterMapTeleport`](../../server/internal/packet/packet.go#L92)<br>`const CharacterMapTeleport = 0x020B` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`UserKillCountUpdate`](../../server/internal/packet/packet.go#L93)<br>`const UserKillCountUpdate = 0x020E` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterShape`](../../server/internal/packet/packet.go#L94)<br>`const CharacterShape = 0x0303` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ChangeAppearance`](../../server/internal/packet/packet.go#L96)<br>`const ChangeAppearance = 0x0226` | Request bag,slot,hair,face,height,sex:u8; reply charId:u32,hair,face,height,sex:u8. |
| [`ChangeSexDenied`](../../server/internal/packet/packet.go#L97)<br>`const ChangeSexDenied = 0x0228` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. empty: client shows SysMsg 862 |
| [`WorldDay`](../../server/internal/packet/packet.go#L98)<br>`const WorldDay = 0x0404` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterMove`](../../server/internal/packet/packet.go#L99)<br>`const CharacterMove = 0x0501` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterReviveHitpoints`](../../server/internal/packet/packet.go#L100)<br>`const CharacterReviveHitpoints = 0x0505` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. charId:u32, hp,mp,sp:u32 |
| [`CharacterUserAutoAttack`](../../server/internal/packet/packet.go#L101)<br>`const CharacterUserAutoAttack = 0x0502` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterMobAutoAttack`](../../server/internal/packet/packet.go#L102)<br>`const CharacterMobAutoAttack = 0x0503` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterDeath`](../../server/internal/packet/packet.go#L103)<br>`const CharacterDeath = 0x0504` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. charId:u32, deathType:u8, killerId:u32 |
| [`CharacterMotion`](../../server/internal/packet/packet.go#L104)<br>`const CharacterMotion = 0x0506` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterEquipment`](../../server/internal/packet/packet.go#L105)<br>`const CharacterEquipment = 0x0507` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`MapAddItem`](../../server/internal/packet/packet.go#L106)<br>`const MapAddItem = 0x0401` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`MapRemoveItem`](../../server/internal/packet/packet.go#L107)<br>`const MapRemoveItem = 0x0402` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterReviveInit`](../../server/internal/packet/packet.go#L108)<br>`const CharacterReviveInit = 0x0406` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. charId:u32 |
| [`MobEnter`](../../server/internal/packet/packet.go#L109)<br>`const MobEnter = 0x0601` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`MobLeave`](../../server/internal/packet/packet.go#L110)<br>`const MobLeave = 0x0602` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`MobMove`](../../server/internal/packet/packet.go#L111)<br>`const MobMove = 0x0603` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`MobAttackCharacter`](../../server/internal/packet/packet.go#L112)<br>`const MobAttackCharacter = 0x0605` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. result:u8, mobId:u32, charId:u32, dmgHp,dmgSp,dmgMp:u16 |
| [`MobDeath`](../../server/internal/packet/packet.go#L113)<br>`const MobDeath = 0x0606` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`MobPeriodicDamage`](../../server/internal/packet/packet.go#L114)<br>`const MobPeriodicDamage = 0x0607` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. mobId:u32, skillId:u16, skillLevel:u8, dmgHp,dmgSp,dmgMp:u16 |
| [`MobCurrentHitpoints`](../../server/internal/packet/packet.go#L115)<br>`const MobCurrentHitpoints = 0x0610` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. mobId:u32, currentHp:u32 |
| [`MapNpcEnter`](../../server/internal/packet/packet.go#L116)<br>`const MapNpcEnter = 0x0E01` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`MapNpcLeave`](../../server/internal/packet/packet.go#L117)<br>`const MapNpcLeave = 0x0E02` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`MapNpcMove`](../../server/internal/packet/packet.go#L118)<br>`const MapNpcMove = 0x0E03` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`MapNpcAttackPlayer`](../../server/internal/packet/packet.go#L119)<br>`const MapNpcAttackPlayer = 0x0E05` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`MapNpcAttackMob`](../../server/internal/packet/packet.go#L120)<br>`const MapNpcAttackMob = 0x0E06` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`UseItem`](../../server/internal/packet/packet.go#L121)<br>`const UseItem = 0x050A` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterMaxEnergy`](../../server/internal/packet/packet.go#L122)<br>`const CharacterMaxEnergy = 0x050B` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`SkillUsedSpMp`](../../server/internal/packet/packet.go#L123)<br>`const SkillUsedSpMp = 0x050C` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`SkillCastStarted`](../../server/internal/packet/packet.go#L124)<br>`const SkillCastStarted = 0x0510` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`BuffAdd`](../../server/internal/packet/packet.go#L125)<br>`const BuffAdd = 0x050D` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. id:u32, skillId:u16, skillLevel:u8, countdownSeconds:u32 |
| [`BuffRemove`](../../server/internal/packet/packet.go#L126)<br>`const BuffRemove = 0x050E` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. objectId:u32 |
| [`SkillKeep`](../../server/internal/packet/packet.go#L127)<br>`const SkillKeep = 0x050F` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. objectId:u32, skillId:u16, skillLevel:u8, hp,mp,sp:u16 |
| [`CharacterMaxHpMpSp`](../../server/internal/packet/packet.go#L128)<br>`const CharacterMaxHpMpSp = 0x051F` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterCurrentHitpoints`](../../server/internal/packet/packet.go#L129)<br>`const CharacterCurrentHitpoints = 0x0521` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterAdditionalStats`](../../server/internal/packet/packet.go#L130)<br>`const CharacterAdditionalStats = 0x0526` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`TeleportSavePositionList`](../../server/internal/packet/packet.go#L131)<br>`const TeleportSavePositionList = 0x0225` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`BlessInit`](../../server/internal/packet/packet.go#L132)<br>`const BlessInit = 0x0211` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`QuestList`](../../server/internal/packet/packet.go#L133)<br>`const QuestList = 0x0901` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`QuestStart`](../../server/internal/packet/packet.go#L134)<br>`const QuestStart = 0x0902` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`QuestEnd`](../../server/internal/packet/packet.go#L135)<br>`const QuestEnd = 0x0903` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`QuestStartAlt`](../../server/internal/packet/packet.go#L136)<br>`const QuestStartAlt = 0x0904` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`QuestCounter`](../../server/internal/packet/packet.go#L137)<br>`const QuestCounter = 0x0905` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`QuestChoose`](../../server/internal/packet/packet.go#L138)<br>`const QuestChoose = 0x0907` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`QuestAbandon`](../../server/internal/packet/packet.go#L139)<br>`const QuestAbandon = 0x0908` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`NpcTalk`](../../server/internal/packet/packet.go#L140)<br>`const NpcTalk = 0x0E04` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`QuestFinishedList`](../../server/internal/packet/packet.go#L141)<br>`const QuestFinishedList = 0x0906` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ChatNormal`](../../server/internal/packet/packet.go#L142)<br>`const ChatNormal = 0x1101` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ChatWorld`](../../server/internal/packet/packet.go#L143)<br>`const ChatWorld = 0x1103` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ChatError`](../../server/internal/packet/packet.go#L144)<br>`const ChatError = 0x1106` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ChatMap`](../../server/internal/packet/packet.go#L145)<br>`const ChatMap = 0x1111` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ChatRaid`](../../server/internal/packet/packet.go#L146)<br>`const ChatRaid = 0x1112` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. raid leader/sub-leader notice: u32 sender, u8 length, text |
| [`WarehouseItemList`](../../server/internal/packet/packet.go#L147)<br>`const WarehouseItemList = 0x0711` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ShopSell`](../../server/internal/packet/packet.go#L148)<br>`const ShopSell = 0x0703` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`SkillUseNoTarget`](../../server/internal/packet/packet.go#L149)<br>`const SkillUseNoTarget = 0x0511` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. a skill on a user (CUser::ChkTargetUser) |
| [`SkillPartyMember`](../../server/internal/packet/packet.go#L150)<br>`const SkillPartyMember = 0x0513` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. a party heal/buff reaching one member |
| [`ResurrectRequest`](../../server/internal/packet/packet.go#L151)<br>`const ResurrectRequest = 0x0553` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. 1-byte body |
| [`ItemRepair`](../../server/internal/packet/packet.go#L152)<br>`const ItemRepair = 0x0554` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. bag, slot -> bag, slot, money |
| [`ReturnToTown`](../../server/internal/packet/packet.go#L153)<br>`const ReturnToTown = 0x0556` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. no body; bind point after 10 s |
| [`GuildCreateFailed`](../../server/internal/packet/packet.go#L156)<br>`const GuildCreateFailed = 0x0D01` | Guilds (CUser::PacketGuild and the CGuild senders). |
| [`GuildDisband`](../../server/internal/packet/packet.go#L157)<br>`const GuildDisband = 0x0D03` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildDisbandAgree`](../../server/internal/packet/packet.go#L158)<br>`const GuildDisbandAgree = 0x0D04` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildInvite`](../../server/internal/packet/packet.go#L159)<br>`const GuildInvite = 0x0D05` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildInviteAnswer`](../../server/internal/packet/packet.go#L160)<br>`const GuildInviteAnswer = 0x0D06` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildApply`](../../server/internal/packet/packet.go#L161)<br>`const GuildApply = 0x0D07` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildJoinResult`](../../server/internal/packet/packet.go#L162)<br>`const GuildJoinResult = 0x0D08` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildLeave`](../../server/internal/packet/packet.go#L163)<br>`const GuildLeave = 0x0D09` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildExpel`](../../server/internal/packet/packet.go#L164)<br>`const GuildExpel = 0x0D0A` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildStatus`](../../server/internal/packet/packet.go#L165)<br>`const GuildStatus = 0x0D0C` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildListBegin`](../../server/internal/packet/packet.go#L166)<br>`const GuildListBegin = 0x0D0D` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildListEnd`](../../server/internal/packet/packet.go#L167)<br>`const GuildListEnd = 0x0D0E` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildDisbanded`](../../server/internal/packet/packet.go#L168)<br>`const GuildDisbanded = 0x0D11` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildOnlineList`](../../server/internal/packet/packet.go#L169)<br>`const GuildOnlineList = 0x0D12` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildOfflineList`](../../server/internal/packet/packet.go#L170)<br>`const GuildOfflineList = 0x0D13` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildMemberAdded`](../../server/internal/packet/packet.go#L171)<br>`const GuildMemberAdded = 0x0D14` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildApplicantList`](../../server/internal/packet/packet.go#L172)<br>`const GuildApplicantList = 0x0D16` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildApplicant`](../../server/internal/packet/packet.go#L173)<br>`const GuildApplicant = 0x0D17` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildApplicantRemoved`](../../server/internal/packet/packet.go#L174)<br>`const GuildApplicantRemoved = 0x0D18` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildRemark`](../../server/internal/packet/packet.go#L175)<br>`const GuildRemark = 0x0D20` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildRankingNotice`](../../server/internal/packet/packet.go#L176)<br>`const GuildRankingNotice = 0x0D26` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. map u16, type u8 |
| [`GuildRankingPoints`](../../server/internal/packet/packet.go#L177)<br>`const GuildRankingPoints = 0x0D27` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. points, country max, its guild id |
| [`GuildRankingList`](../../server/internal/packet/packet.go#L178)<br>`const GuildRankingList = 0x0D37` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. count, (guild id, points, rank) |
| [`GuildBankResult`](../../server/internal/packet/packet.go#L179)<br>`const GuildBankResult = 0x0D28` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildBankList`](../../server/internal/packet/packet.go#L180)<br>`const GuildBankList = 0x0D29` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildBankPlaced`](../../server/internal/packet/packet.go#L181)<br>`const GuildBankPlaced = 0x0D31` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildBankTaken`](../../server/internal/packet/packet.go#L182)<br>`const GuildBankTaken = 0x0D32` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildCreate`](../../server/internal/packet/packet.go#L183)<br>`const GuildCreate = 0x0D21` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildCreateAgree`](../../server/internal/packet/packet.go#L184)<br>`const GuildCreateAgree = 0x0D22` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildEtin`](../../server/internal/packet/packet.go#L185)<br>`const GuildEtin = 0x0D23` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildNpcList`](../../server/internal/packet/packet.go#L186)<br>`const GuildNpcList = 0x0D24` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. count, (type u16, level u8, slot u8) |
| [`GuildNpcUpgrade`](../../server/internal/packet/packet.go#L187)<br>`const GuildNpcUpgrade = 0x0D25` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. C->S type u16, level u8; S->C result, type, level, slot |
| [`GuildHouseKeepEtin`](../../server/internal/packet/packet.go#L188)<br>`const GuildHouseKeepEtin = 0x0D33` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. keep etin u32 |
| [`GuildHouseBuy`](../../server/internal/packet/packet.go#L189)<br>`const GuildHouseBuy = 0x0D34` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. result u8, money u32 |
| [`GuildDonate`](../../server/internal/packet/packet.go#L190)<br>`const GuildDonate = 0x0D35` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. count, (bag, slot) of the etin items consumed |
| [`GuildList`](../../server/internal/packet/packet.go#L191)<br>`const GuildList = 0x0D2F` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GuildNew`](../../server/internal/packet/packet.go#L192)<br>`const GuildNew = 0x0D30` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ChatGuild`](../../server/internal/packet/packet.go#L193)<br>`const ChatGuild = 0x1104` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ChatWhisper`](../../server/internal/packet/packet.go#L194)<br>`const ChatWhisper = 0x1102` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ChatShout`](../../server/internal/packet/packet.go#L195)<br>`const ChatShout = 0x1107` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ChatWorldShout`](../../server/internal/packet/packet.go#L196)<br>`const ChatWorldShout = 0x1108` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`SkillDetectInfo`](../../server/internal/packet/packet.go#L197)<br>`const SkillDetectInfo = 0x0558` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. attr:u32, level:u16, mode:u8 |
| [`CharacterResurrect`](../../server/internal/packet/packet.go#L199)<br>`const CharacterResurrect = 0x0551` | charId:u32, 5-byte blob 04 46 AF 37 01 (verbatim), x,y,z:f32. |
| [`SkillUse`](../../server/internal/packet/packet.go#L200)<br>`const SkillUse = 0x0517` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`SkillExtraDamage`](../../server/internal/packet/packet.go#L201)<br>`const SkillExtraDamage = 0x051A` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CharacterMountShape`](../../server/internal/packet/packet.go#L202)<br>`const CharacterMountShape = 0x051D` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. characterId:u32, shape:u8 |
| [`CharacterDeathState`](../../server/internal/packet/packet.go#L203)<br>`const CharacterDeathState = 0x051C` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. charId:u32, 0x07, 0x02 |
| [`AccountPoints`](../../server/internal/packet/packet.go#L204)<br>`const AccountPoints = 0x2601` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PointBuy`](../../server/internal/packet/packet.go#L205)<br>`const PointBuy = 0x2602` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PointGiftSend`](../../server/internal/packet/packet.go#L206)<br>`const PointGiftSend = 0x2603` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PointStoredItemList`](../../server/internal/packet/packet.go#L207)<br>`const PointStoredItemList = 0x2606` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`PointStoredItemGet`](../../server/internal/packet/packet.go#L208)<br>`const PointStoredItemGet = 0x2607` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CashPoint`](../../server/internal/packet/packet.go#L209)<br>`const CashPoint = 0x2605` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`BankItemList`](../../server/internal/packet/packet.go#L210)<br>`const BankItemList = 0xB101` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`LoginHandshake`](../../server/internal/packet/packet.go#L211)<br>`const LoginHandshake = 0xA101` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`LoginRequest`](../../server/internal/packet/packet.go#L212)<br>`const LoginRequest = 0xA102` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ServerList`](../../server/internal/packet/packet.go#L213)<br>`const ServerList = 0xA201` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`SelectServer`](../../server/internal/packet/packet.go#L214)<br>`const SelectServer = 0xA202` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GameHandshake`](../../server/internal/packet/packet.go#L215)<br>`const GameHandshake = 0xA301` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`Ping`](../../server/internal/packet/packet.go#L216)<br>`const Ping = 0xA303` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ChatNormalAdmin`](../../server/internal/packet/packet.go#L217)<br>`const ChatNormalAdmin = 0xF101` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ChatWorldAdmin`](../../server/internal/packet/packet.go#L218)<br>`const ChatWorldAdmin = 0xF103` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GmCommandError`](../../server/internal/packet/packet.go#L219)<br>`const GmCommandError = 0xF501` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GmCommandGetItem`](../../server/internal/packet/packet.go#L220)<br>`const GmCommandGetItem = 0xF702` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`EnchantAdd`](../../server/internal/packet/packet.go#L221)<br>`const EnchantAdd = 0x0805` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ItemCompose`](../../server/internal/packet/packet.go#L222)<br>`const ItemCompose = 0x0806` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GemAdd`](../../server/internal/packet/packet.go#L223)<br>`const GemAdd = 0x0801` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GemRemove`](../../server/internal/packet/packet.go#L224)<br>`const GemRemove = 0x0802` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CloakAdd`](../../server/internal/packet/packet.go#L225)<br>`const CloakAdd = 0x0803` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`CloakRemove`](../../server/internal/packet/packet.go#L226)<br>`const CloakRemove = 0x0804` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GemAddPossibility`](../../server/internal/packet/packet.go#L227)<br>`const GemAddPossibility = 0x0809` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`GemRemovePossibility`](../../server/internal/packet/packet.go#L228)<br>`const GemRemovePossibility = 0x080A` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ExchangeRequest`](../../server/internal/packet/packet.go#L229)<br>`const ExchangeRequest = 0x0A01` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ExchangeAnswer`](../../server/internal/packet/packet.go#L230)<br>`const ExchangeAnswer = 0x0A02` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ExchangeStart`](../../server/internal/packet/packet.go#L231)<br>`const ExchangeStart = 0x0A03` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ExchangeEnd`](../../server/internal/packet/packet.go#L232)<br>`const ExchangeEnd = 0x0A04` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ExchangeReady`](../../server/internal/packet/packet.go#L233)<br>`const ExchangeReady = 0x0A05` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ExchangeAdd`](../../server/internal/packet/packet.go#L234)<br>`const ExchangeAdd = 0x0A06` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ExchangeRemove`](../../server/internal/packet/packet.go#L235)<br>`const ExchangeRemove = 0x0A07` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ExchangeAddMoney`](../../server/internal/packet/packet.go#L236)<br>`const ExchangeAddMoney = 0x0A08` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ExchangeAddDest`](../../server/internal/packet/packet.go#L237)<br>`const ExchangeAddDest = 0x0A09` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ExchangeConfirm`](../../server/internal/packet/packet.go#L238)<br>`const ExchangeConfirm = 0x0A0A` | Opcodes. Comments carry the confirmed wire shapes recorded while the server was written against captures of the real 5.4 client. |
| [`ErrTooLarge`](../../server/internal/packet/packet.go#L241)<br>`var ErrTooLarge` | See the source definition and the subsystem ownership contract. |
| [`ErrStringTooLarge`](../../server/internal/packet/packet.go#L242)<br>`var ErrStringTooLarge` | See the source definition and the subsystem ownership contract. |
| [`Writer`](../../server/internal/packet/packet.go#L245)<br>`type Writer` | Writer builds one frame; the length prefix is filled in by Bytes. |
| [`New`](../../server/internal/packet/packet.go#L250)<br>`func New(opcode Opcode) *Writer` | See the source definition and the subsystem ownership contract. |
| [`Count`](../../server/internal/packet/packet.go#L257)<br>`func (w *Writer) Count(v uint16)` | See the source definition and the subsystem ownership contract. |
| [`U8`](../../server/internal/packet/packet.go#L264)<br>`func (w *Writer) U8(v uint8)` | See the source definition and the subsystem ownership contract. |
| [`U16`](../../server/internal/packet/packet.go#L265)<br>`func (w *Writer) U16(v uint16)` | See the source definition and the subsystem ownership contract. |
| [`U32`](../../server/internal/packet/packet.go#L266)<br>`func (w *Writer) U32(v uint32)` | See the source definition and the subsystem ownership contract. |
| [`I32`](../../server/internal/packet/packet.go#L267)<br>`func (w *Writer) I32(v int32)` | See the source definition and the subsystem ownership contract. |
| [`U64`](../../server/internal/packet/packet.go#L268)<br>`func (w *Writer) U64(v uint64)` | See the source definition and the subsystem ownership contract. |
| [`F32`](../../server/internal/packet/packet.go#L269)<br>`func (w *Writer) F32(v float32)` | See the source definition and the subsystem ownership contract. |
| [`Bytes`](../../server/internal/packet/packet.go#L270)<br>`func (w *Writer) Bytes(v []byte)` | See the source definition and the subsystem ownership contract. |
| [`Bool`](../../server/internal/packet/packet.go#L271)<br>`func (w *Writer) Bool(v bool)` | See the source definition and the subsystem ownership contract. |
| [`FixedString`](../../server/internal/packet/packet.go#L281)<br>`func (w *Writer) FixedString(value string, size int)` | FixedString writes value null-padded to size bytes; an oversized value makes Finish fail instead of truncating silently. |
| [`Len`](../../server/internal/packet/packet.go#L292)<br>`func (w *Writer) Len() int` | See the source definition and the subsystem ownership contract. |
| [`Finish`](../../server/internal/packet/packet.go#L295)<br>`func (w *Writer) Finish() ([]byte, error)` | Finish returns the frame with its length prefix set. |
| [`MustFinish`](../../server/internal/packet/packet.go#L307)<br>`func (w *Writer) MustFinish() []byte` | MustFinish is for frames whose shape is fixed and cannot overflow. |
| [`ErrShort`](../../server/internal/packet/packet.go#L315)<br>`var ErrShort` | See the source definition and the subsystem ownership contract. |
| [`U16`](../../server/internal/packet/packet.go#L317)<br>`func U16(b []byte) (uint16, error)` | See the source definition and the subsystem ownership contract. |
| [`U32`](../../server/internal/packet/packet.go#L324)<br>`func U32(b []byte) (uint32, error)` | See the source definition and the subsystem ownership contract. |
| [`F32`](../../server/internal/packet/packet.go#L331)<br>`func F32(b []byte) (float32, error)` | See the source definition and the subsystem ownership contract. |
| [`FixedText`](../../server/internal/packet/packet.go#L337)<br>`func FixedText(b []byte) string` | FixedText decodes NUL-terminated UTF-8 with a legacy ANSI fallback. |
| [`NewReader`](../../server/internal/packet/packet.go#L339)<br>`func NewReader(p []byte) *proto.Reader` | See the source definition and the subsystem ownership contract. |
| [`BaseOpcode`](../../server/internal/packet/packet.go#L340)<br>`func BaseOpcode(v uint16) Opcode` | See the source definition and the subsystem ownership contract. |
| [`GXForgeQuote`](../../server/internal/packet/packet.go#L342)<br>`const GXForgeQuote = 0x7F03` | See the source definition and the subsystem ownership contract. |
| [`GXPartyInviteName`](../../server/internal/packet/packet.go#L345)<br>`const GXPartyInviteName = 0x7F15` | GXPartyInviteName is the global /party <name> invitation request. |
| [`RaidLevel`](../../server/internal/packet/packet.go#L348)<br>`const RaidLevel = 0x0C13` | RaidLevel is the original member level notice, parallel to PartyLevel. |

<details><summary>Type definitions and owned fields</summary>

```go
type Opcode uint16
```

```go
type Writer struct {
	data []byte
	err  error
}
```

</details>

## server/internal/password

Versioned Argon2id hashing/verification, bounded workers, legacy upgrade and normalized failure work.

Tests: [password_test.go](../../server/internal/password/password_test.go).

### password.go

[server/internal/password/password.go](../../server/internal/password/password.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`prefix`](../../server/internal/password/password.go#L13)<br>`const prefix = "$argon2id$v=19$m=19456,t=2,p=1$"` | See the source definition and the subsystem ownership contract. |
| [`work`](../../server/internal/password/password.go#L15)<br>`var work` | See the source definition and the subsystem ownership contract. |
| [`derive`](../../server/internal/password/password.go#L17)<br>`func derive(ctx context.Context, value string, salt []byte) ([]byte, error)` | See the source definition and the subsystem ownership contract. |
| [`Hash`](../../server/internal/password/password.go#L30)<br>`func Hash(value string) (string, error)` | See the source definition and the subsystem ownership contract. |
| [`HashContext`](../../server/internal/password/password.go#L31)<br>`func HashContext(ctx context.Context, value string) (string, error)` | See the source definition and the subsystem ownership contract. |
| [`Verify`](../../server/internal/password/password.go#L45)<br>`func Verify(stored, value string) (bool, bool)` | See the source definition and the subsystem ownership contract. |
| [`Dummy`](../../server/internal/password/password.go#L52)<br>`func Dummy(ctx context.Context, value string) error` | Dummy performs the same bounded KDF work for an unknown account. This reduces user-enumeration timing differences; network/database timing is not constant. |
| [`VerifyContext`](../../server/internal/password/password.go#L56)<br>`func VerifyContext(ctx context.Context, stored, value string) (bool, bool, error)` | See the source definition and the subsystem ownership contract. |

## server/internal/scheduler

Single cancellable serial simulation timeline, deterministic stepping and duration/delay histograms.

Tests: [scheduler_test.go](../../server/internal/scheduler/scheduler_test.go).

### scheduler.go

[server/internal/scheduler/scheduler.go](../../server/internal/scheduler/scheduler.go#L1)

Package scheduler owns the cancellable world background timeline. All jobs
execute serially; an overdue job runs once rather than replaying a tick storm.

Direct GX dependencies: `server/internal/telemetry`.

| Declaration | Responsibility / source comment |
|---|---|
| [`Job`](../../server/internal/scheduler/scheduler.go#L13)<br>`type Job` | See the source definition and the subsystem ownership contract. |
| [`task`](../../server/internal/scheduler/scheduler.go#L18)<br>`type task` | See the source definition and the subsystem ownership contract. |
| [`Scheduler`](../../server/internal/scheduler/scheduler.go#L24)<br>`type Scheduler` | See the source definition and the subsystem ownership contract. |
| [`New`](../../server/internal/scheduler/scheduler.go#L30)<br>`func New(start time.Time, jobs []Job) *Scheduler` | See the source definition and the subsystem ownership contract. |
| [`Step`](../../server/internal/scheduler/scheduler.go#L44)<br>`func (s *Scheduler) Step(ctx context.Context, now time.Time)` | Step also permits deterministic tests without sleeping or starting goroutines. |
| [`Run`](../../server/internal/scheduler/scheduler.go#L62)<br>`func (s *Scheduler) Run(ctx context.Context)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type Job struct {
	Name     string
	Interval time.Duration
	Run      func(time.Time)
}
```

```go
type task struct {
	Job
	next     time.Time
	duration *telemetry.Histogram
	late     *telemetry.Histogram
}
```

```go
type Scheduler struct {
	mu    sync.Mutex
	tasks []task
	now   func() time.Time
}
```

</details>

## server/internal/session

Atomic single-use login-to-world SQLite handoffs, random identities and credential-version validation.

Tests: [session_test.go](../../server/internal/session/session_test.go).

### session.go

[server/internal/session/session.go](../../server/internal/session/session.go#L1)

Package session hands an authenticated login over to the world server.
Login writes an atomic SQLite ticket; the client then connects to
world and presents the same identity token in its 0xA301 handshake.

Direct GX dependencies: `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`MaxAge`](../../server/internal/session/session.go#L20)<br>`const MaxAge = 10 * time.Minute` | MaxAge bounds how long a handoff stays valid after login. |
| [`Handoff`](../../server/internal/session/session.go#L22)<br>`type Handoff` | See the source definition and the subsystem ownership contract. |
| [`NewIdentity`](../../server/internal/session/session.go#L31)<br>`func NewIdentity() [16]byte` | NewIdentity returns a fresh random token for one login. |
| [`NewSessionID`](../../server/internal/session/session.go#L40)<br>`func NewSessionID() int64` | NewSessionID returns the bigint session id passed to the login procedures. |
| [`ErrInvalid`](../../server/internal/session/session.go#L48)<br>`var ErrInvalid` | See the source definition and the subsystem ownership contract. |
| [`Save`](../../server/internal/session/session.go#L51)<br>`func Save(h Handoff) error` | Save replaces the previous login ticket in the configured runtime database. |
| [`Consume`](../../server/internal/session/session.go#L78)<br>`func Consume(userUID uint32, identity [16]byte) (Handoff, error)` | Consume first checks the token on a read-only connection. A valid candidate is then consumed by one guarded DELETE that rechecks expiry and revocation. |
| [`PurgeExpired`](../../server/internal/session/session.go#L124)<br>`func PurgeExpired() error` | PurgeExpired drops old key material in bounded batches. |

<details><summary>Type definitions and owned fields</summary>

```go
type Handoff struct {
	Identity          [16]byte
	AESMaterial       [32]byte // login AES key (16) + IV (16)
	UserUID           uint32
	SessionID         int64
	CredentialVersion int64
}
```

</details>

## server/internal/telemetry

Bounded registry of atomic counters/gauges and duration histograms in Prometheus text format.

### metrics.go

[server/internal/telemetry/metrics.go](../../server/internal/telemetry/metrics.go#L1)

Package telemetry exposes bounded, process-wide metrics without player labels.

| Declaration | Responsibility / source comment |
|---|---|
| [`Number`](../../server/internal/telemetry/metrics.go#L14)<br>`type Number` | See the source definition and the subsystem ownership contract. |
| [`Add`](../../server/internal/telemetry/metrics.go#L16)<br>`func (n *Number) Add(v int64)` | See the source definition and the subsystem ownership contract. |
| [`Set`](../../server/internal/telemetry/metrics.go#L17)<br>`func (n *Number) Set(v int64)` | See the source definition and the subsystem ownership contract. |
| [`Value`](../../server/internal/telemetry/metrics.go#L18)<br>`func (n *Number) Value() int64` | See the source definition and the subsystem ownership contract. |
| [`Histogram`](../../server/internal/telemetry/metrics.go#L20)<br>`type Histogram` | See the source definition and the subsystem ownership contract. |
| [`limits`](../../server/internal/telemetry/metrics.go#L26)<br>`var limits` | See the source definition and the subsystem ownership contract. |
| [`Observe`](../../server/internal/telemetry/metrics.go#L28)<br>`func (h *Histogram) Observe(d time.Duration)` | See the source definition and the subsystem ownership contract. |
| [`registry`](../../server/internal/telemetry/metrics.go#L41)<br>`var registry` | See the source definition and the subsystem ownership contract. |
| [`valid`](../../server/internal/telemetry/metrics.go#L47)<br>`func valid(name string)` | See the source definition and the subsystem ownership contract. |
| [`NumberMetric`](../../server/internal/telemetry/metrics.go#L57)<br>`func NumberMetric(name string) *Number` | See the source definition and the subsystem ownership contract. |
| [`DurationMetric`](../../server/internal/telemetry/metrics.go#L71)<br>`func DurationMetric(name string) *Histogram` | See the source definition and the subsystem ownership contract. |
| [`WritePrometheus`](../../server/internal/telemetry/metrics.go#L88)<br>`func WritePrometheus(w io.Writer)` | WritePrometheus writes the standard text exposition format. Names come only from fixed code registrations; user IDs and source addresses are never labels. |

<details><summary>Type definitions and owned fields</summary>

```go
type Number struct{ value atomic.Int64 }
```

```go
type Histogram struct {
	count   atomic.Uint64
	nanos   atomic.Uint64
	buckets [8]atomic.Uint64
}
```

</details>
