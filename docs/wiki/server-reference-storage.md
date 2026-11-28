# SQLite package API

> Generated from Go source by `tools/generate_server_reference.py`; do not edit by hand.

[Package responsibilities](server-packages.md) · [Server architecture](server-architecture.md)

Every production Go file is listed, including OS-specific alternatives. Signatures include private helpers to make ownership and call boundaries inspectable. Declaration comments are source documentation, not an independent guarantee of behavior; the architecture guide explains transaction, concurrency and failure contracts.

## server/internal/db

All authoritative SQLite access, definitions, migrations, guarded economy transactions and crash/session ownership.

Tests: [account_session_test.go](../../server/internal/db/account_session_test.go), [auction_test.go](../../server/internal/db/auction_test.go), [cash_test.go](../../server/internal/db/cash_test.go), [character_creation_test.go](../../server/internal/db/character_creation_test.go), [chars_integration_test.go](../../server/internal/db/chars_integration_test.go), [consumable_links_test.go](../../server/internal/db/consumable_links_test.go), [defs_integration_test.go](../../server/internal/db/defs_integration_test.go), [economy_atomic_test.go](../../server/internal/db/economy_atomic_test.go), [gm_test.go](../../server/internal/db/gm_test.go), [group_write_test.go](../../server/internal/db/group_write_test.go), [isolation_test.go](../../server/internal/db/isolation_test.go), [keeps_test.go](../../server/internal/db/keeps_test.go), [legacy_backup_test.go](../../server/internal/db/legacy_backup_test.go), [mall_transaction_test.go](../../server/internal/db/mall_transaction_test.go), [migrations_test.go](../../server/internal/db/migrations_test.go), [mob_quest_test.go](../../server/internal/db/mob_quest_test.go), [production_integration_test.go](../../server/internal/db/production_integration_test.go), [quest_start_atomic_test.go](../../server/internal/db/quest_start_atomic_test.go), [skill_cleanup_test.go](../../server/internal/db/skill_cleanup_test.go), [snapshot_atomic_test.go](../../server/internal/db/snapshot_atomic_test.go), [social_test.go](../../server/internal/db/social_test.go), [sqlite_test.go](../../server/internal/db/sqlite_test.go).

### account.go

[server/internal/db/account.go](../../server/internal/db/account.go#L1)

Direct GX dependencies: `server/internal/config`, `server/internal/password`.

| Declaration | Responsibility / source comment |
|---|---|
| [`LoginResult`](../../server/internal/db/account.go#L11)<br>`type LoginResult` | LoginResult is what PS_UserData.dbo.usp_Try_GameLogin_Taiwan selects. |
| [`TryGameLogin`](../../server/internal/db/account.go#L19)<br>`func TryGameLogin(userID, passwordValue string, sessionID int64, userIP string) (LoginResult, error)` | TryGameLogin translates the original procedure into one SQLite transaction, including account status, the login log and UserLoginStatus. |
| [`TryGameLogout`](../../server/internal/db/account.go#L122)<br>`func TryGameLogout(userUID int32, sessionID int64) error` | Logout and the original two-minute online-point credit commit together. |
| [`LoginDecision`](../../server/internal/db/account.go#L164)<br>`func LoginDecision(login LoginResult) (result uint8, permission uint8)` | LoginDecision maps the procedure's Status to the client's login result byte and permission (auth status), as recovered from the original ps_login (CUser login handling; see Eternity src/auth/login_protocol.cpp). Permission 11 is a regular account; 1..5 are the admin tiers. |
| [`CanPlay`](../../server/internal/db/account.go#L222)<br>`func CanPlay(status int32) bool` | Account is the account state the world server needs. CanPlay rejects unrecognized authorization states rather than granting permissions from arbitrary legacy status bits. |
| [`Account`](../../server/internal/db/account.go#L226)<br>`type Account` | See the source definition and the subsystem ownership contract. |
| [`ErrNotFound`](../../server/internal/db/account.go#L237)<br>`var ErrNotFound` | See the source definition and the subsystem ownership contract. |
| [`AccountByUID`](../../server/internal/db/account.go#L241)<br>`func AccountByUID(userUID int32) (Account, error)` | AccountByUID reads Users_Master plus the realm's UserMaxGrow row (PS_GameData.dbo.usp_Read_User_MaxGrow_R). |

<details><summary>Type definitions and owned fields</summary>

```go
type LoginResult struct {
	Status            int32
	UserUID           int32
	CredentialVersion int64
}
```

```go
type Account struct {
	UserUID           int32
	CredentialVersion int64
	UserID            string
	Status            int16
	Admin             bool
	AdminLevel        uint8
	Point             int32
	Country           uint8 // 0 light, 1 dark, 2 none chosen yet
}
```

</details>

### accounts_admin.go

[server/internal/db/accounts_admin.go](../../server/internal/db/accounts_admin.go#L1)

Direct GX dependencies: `internal/authwire`, `server/internal/password`.

| Declaration | Responsibility / source comment |
|---|---|
| [`CreateAccount`](../../server/internal/db/accounts_admin.go#L13)<br>`func CreateAccount(name, secret string, admin bool) (int32, error)` | CreateAccount never stores plaintext and never replaces an existing account. |
| [`SetAccountPassword`](../../server/internal/db/accounts_admin.go#L52)<br>`func SetAccountPassword(name, secret string) error` | SetAccountPassword revokes pending tickets and active sessions through a monotonically increasing credential version. World checks it every 30 seconds. |
| [`revokeCredentialTx`](../../server/internal/db/accounts_admin.go#L78)<br>`func revokeCredentialTx(c context.Context, tx *sql.Tx, uid uint32) error` | See the source definition and the subsystem ownership contract. |
| [`UpgradeLegacyPasswords`](../../server/internal/db/accounts_admin.go#L89)<br>`func UpgradeLegacyPasswords() (upgraded, needsReset int, err error)` | UpgradeLegacyPasswords hashes passwords still stored in the legacy format. Each update compares the original row so a concurrent password reset wins. Invalid legacy credentials are counted and can be reset with the account CLI. |
| [`legacyPassword`](../../server/internal/db/accounts_admin.go#L129)<br>`type legacyPassword` | See the source definition and the subsystem ownership contract. |
| [`legacyPasswords`](../../server/internal/db/accounts_admin.go#L134)<br>`func legacyPasswords(reader *sql.DB, after int64) ([]legacyPassword, error)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type legacyPassword struct {
	uid    int64
	secret string
}
```

</details>

### auction.go

[server/internal/db/auction.go](../../server/internal/db/auction.go#L1)

Direct GX dependencies: `internal/proto`.

| Declaration | Responsibility / source comment |
|---|---|
| [`ErrAuctionChanged`](../../server/internal/db/auction.go#L15)<br>`var ErrAuctionChanged` | See the source definition and the subsystem ownership contract. |
| [`ErrAuctionInvalid`](../../server/internal/db/auction.go#L16)<br>`var ErrAuctionInvalid` | See the source definition and the subsystem ownership contract. |
| [`ErrAuctionFunds`](../../server/internal/db/auction.go#L17)<br>`var ErrAuctionFunds` | See the source definition and the subsystem ownership contract. |
| [`ErrAuctionItem`](../../server/internal/db/auction.go#L18)<br>`var ErrAuctionItem` | See the source definition and the subsystem ownership contract. |
| [`ErrAuctionWatchFull`](../../server/internal/db/auction.go#L19)<br>`var ErrAuctionWatchFull` | See the source definition and the subsystem ownership contract. |
| [`AuctionListing`](../../server/internal/db/auction.go#L22)<br>`type AuctionListing` | See the source definition and the subsystem ownership contract. |
| [`AuctionClaim`](../../server/internal/db/auction.go#L32)<br>`type AuctionClaim` | See the source definition and the subsystem ownership contract. |
| [`auctionTx`](../../server/internal/db/auction.go#L38)<br>`func auctionTx(run func(context.Context, *sql.Tx) error) error` | See the source definition and the subsystem ownership contract. |
| [`auctionClaim`](../../server/internal/db/auction.go#L48)<br>`func auctionClaim(c context.Context, tx *sql.Tx, id int64, owner uint32, gold int64, reason string, now int64) error` | See the source definition and the subsystem ownership contract. |
| [`settleAuction`](../../server/internal/db/auction.go#L53)<br>`func settleAuction(c context.Context, tx *sql.Tx, a AuctionListing, buyer uint32, price uint32, state string, now int64) error` | See the source definition and the subsystem ownership contract. |
| [`auctionRefund`](../../server/internal/db/auction.go#L74)<br>`func auctionRefund(c context.Context, tx *sql.Tx, a AuctionListing, now int64) error` | See the source definition and the subsystem ownership contract. |
| [`auctionFields`](../../server/internal/db/auction.go#L81)<br>`const auctionFields = "ID,Seller,Bidder,Price,Buyout,Deposit,Ends,State,Version"` | See the source definition and the subsystem ownership contract. |
| [`scanAuction`](../../server/internal/db/auction.go#L83)<br>`func scanAuction(scan func(...any) error, a *AuctionListing) error` | See the source definition and the subsystem ownership contract. |
| [`ExpireAuctions`](../../server/internal/db/auction.go#L87)<br>`func ExpireAuctions() error` | See the source definition and the subsystem ownership contract. |
| [`PostAuction`](../../server/internal/db/auction.go#L123)<br>`func PostAuction(a AuctionListing, change CharacterChange) error` | See the source definition and the subsystem ownership contract. |
| [`ChangeAuction`](../../server/internal/db/auction.go#L138)<br>`func ChangeAuction(owner uint32, req proto.AuctionRequest, change *CharacterChange) error` | See the source definition and the subsystem ownership contract. |
| [`AuctionClaims`](../../server/internal/db/auction.go#L212)<br>`func AuctionClaims(owner uint32, view int, id int64) ([]AuctionClaim, error)` | See the source definition and the subsystem ownership contract. |
| [`CollectAuction`](../../server/internal/db/auction.go#L244)<br>`func CollectAuction(claim AuctionClaim, change CharacterChange) error` | See the source definition and the subsystem ownership contract. |
| [`ListAuctions`](../../server/internal/db/auction.go#L266)<br>`func ListAuctions(owner uint32, req proto.AuctionRequest) (proto.AuctionReply, error)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type AuctionListing struct {
	Seller, Bidder                   uint32
	SellerName, Name, State, Request string
	Category                         int
	Level                            uint16
	Item                             Item
	ID, Version, Ends                int64
	Price, Buyout, Deposit           uint32
}
```

```go
type AuctionClaim struct {
	ID   int64
	Item Item
	Gold uint32
}
```

</details>

### character_exit.go

[server/internal/db/character_exit.go](../../server/internal/db/character_exit.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`SaveCharacterExit`](../../server/internal/db/character_exit.go#L10)<br>`func SaveCharacterExit(ch Character, buffs []AppliedBuff, skills []CharacterSkill, quests []CharQuest) error` | SaveCharacterExit commits the final character state and offline timers together. Other sessions can share the durable commit. |

### chars.go

[server/internal/db/chars.go](../../server/internal/db/chars.go#L1)

Direct GX dependencies: `server/internal/config`.

| Declaration | Responsibility / source comment |
|---|---|
| [`Character`](../../server/internal/db/chars.go#L16)<br>`type Character` | Character mirrors PS_GameData.dbo.Chars. CurrentHP/MP/SP are base values (see world.CharacterStats); Chars.HP/MP/SP store them. |
| [`NewCharacter`](../../server/internal/db/chars.go#L41)<br>`type NewCharacter` | NewCharacter is a character to create: the client's request plus the starting stats, pools and position the world server derives for it. |
| [`DeleteWindow`](../../server/internal/db/chars.go#L51)<br>`const DeleteWindow = 2 * time.Hour` | DeleteWindow is how long a deleted character stays restorable. |
| [`MoneyCap`](../../server/internal/db/chars.go#L54)<br>`const MoneyCap = 4_000_000_000` | MoneyCap keeps gold within the signed Money column's bit pattern. |
| [`BaseStatCap`](../../server/internal/db/chars.go#L58)<br>`const BaseStatCap = 32000` | BaseStatCap is the configured character base-stat limit, within the signed smallint character columns and SaveCharacter parameters. |
| [`u16`](../../server/internal/db/chars.go#L60)<br>`func u16(v int64) uint16` | See the source definition and the subsystem ownership contract. |
| [`CharactersForAccount`](../../server/internal/db/chars.go#L64)<br>`func CharactersForAccount(userUID uint32) ([]Character, error)` | CharactersForAccount lists live characters (usp_Read_Chars_R) plus those deleted within DeleteWindow, flagged PendingDelete. |
| [`CharacterDetail`](../../server/internal/db/chars.go#L144)<br>`func CharacterDetail(userUID, charID uint32) (Character, error)` | CharacterDetail loads a playable character with usp_Read_Chars_Detail2_R (which also stamps JoinDate). It returns ErrNotFound unless the character belongs to userUID and is not deleted. |
| [`readCharacterDetail`](../../server/internal/db/chars.go#L165)<br>`func readCharacterDetail(userUID, charID uint32) (Character, error)` | See the source definition and the subsystem ownership contract. |
| [`CharacterIDInSlot`](../../server/internal/db/chars.go#L213)<br>`func CharacterIDInSlot(userUID uint32, slot uint8) (uint32, bool, error)` | CharacterIDInSlot runs usp_Read_CharSlots_R. |
| [`NameTaken`](../../server/internal/db/chars.go#L231)<br>`func NameTaken(name string) (bool, error)` | NameTaken reports whether a live character, or one deleted within the restore window, already uses name. |
| [`CreateCharacter`](../../server/internal/db/chars.go#L248)<br>`func CreateCharacter(account Account, nc NewCharacter) (uint32, error)` | CreateCharacter runs usp_Create_Char_R. The procedure itself rejects duplicate names (return -2) and fixes SkillPoint at 5. The parameters mirror the original CUser::ProcessCharacterCreate: level 1, no stat or skill points, Vg 200. |
| [`CreateCharacterWithItems`](../../server/internal/db/chars.go#L254)<br>`func CreateCharacterWithItems(account Account, nc NewCharacter, items []Item) (uint32, error)` | CreateCharacterWithItems commits the character and initial inventory together. Slot/name validation is repeated under the writer lock, not trusted to callers. |
| [`DeleteCharacter`](../../server/internal/db/chars.go#L311)<br>`func DeleteCharacter(userUID, charID uint32) (bool, error)` | DeleteCharacter soft-deletes a character and removes its guild membership in one transaction. A guild master must resolve the guild's leadership first. |
| [`RestoreCharacter`](../../server/internal/db/chars.go#L365)<br>`func RestoreCharacter(userUID, charID uint32) (bool, error)` | RestoreCharacter undoes a deletion within DeleteWindow. The EP4.5 procedures have no restore; this is the inverse of usp_Delete_Char_E's Chars update. |
| [`EnableRename`](../../server/internal/db/chars.go#L387)<br>`func EnableRename(charID uint32) error` | EnableRename runs Usp_Save_Char_EnableRename_E (name-change stone). |
| [`RenameCharacter`](../../server/internal/db/chars.go#L393)<br>`func RenameCharacter(userUID, charID uint32, oldName, newName string) (bool, error)` | RenameCharacter commits a pending rename with usp_Update_Char_Rename and consumes one RenameCnt. |
| [`MapForSave`](../../server/internal/db/chars.go#L437)<br>`var MapForSave` | MapForSave maps a world zone to the map stored in Chars.Map (the world replaces it so instanced zones save their real map). |
| [`SaveCharacter`](../../server/internal/db/chars.go#L441)<br>`func SaveCharacter(ch *Character) error` | SaveCharacter persists level, points, stats, HP/MP/SP, position, exp, money and kill counters with usp_Save_Char_Info_E. |
| [`writeCharacter`](../../server/internal/db/chars.go#L454)<br>`func writeCharacter(c context.Context, executor sqlExecutor, ch *Character) (sql.Result, error)` | See the source definition and the subsystem ownership contract. |
| [`SaveAppearance`](../../server/internal/db/chars.go#L478)<br>`func SaveAppearance(charID uint32, hair, face, size, sex uint8) error` | SaveAppearance runs usp_Save_Char_BodyChg_E. |
| [`CharacterByName`](../../server/internal/db/chars.go#L486)<br>`func CharacterByName(name string) (charID, userUID uint32, found bool, err error)` | CharacterByName finds a live character's account and id by name. |
| [`CharacterByID`](../../server/internal/db/chars.go#L503)<br>`func CharacterByID(charID uint32) (Character, error)` | CharacterByID loads any live character regardless of account (used to credit an offline caster). |
| [`execSQL`](../../server/internal/db/chars.go#L521)<br>`func execSQL(database, name string, args ...any) error` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type Character struct {
	ID, AccountID                     uint32
	Slot                              uint8
	Name                              string
	Level                             uint16
	MapID                             uint16
	PosX, PosY, PosZ                  float32
	Dir                               int16
	Race, Mode                        uint8 // Family, Grow
	Hair, Face, Height                uint8
	Class, Gender                     uint8 // Job, Sex
	Str, Rec, Int, Wis, Dex, Luc      uint16
	StatPoint, SkillPoint             uint16
	MaxHP, MaxMP, MaxSP               uint32
	CurrentHP, CurrentMP, CurrentSP   uint32
	Exp, Money                        uint32
	Kills, Deaths, Victories, Defeats uint32
	KillLevel, DeadLevel              uint8
	Hg, Vg                            int16
	Cg, Og, Ig                        uint8
	PendingDelete, RenamePending      bool
}
```

```go
type NewCharacter struct {
	Slot, Race, Mode, Hair, Face, Height, Class, Gender uint8
	Name                                                string
	MapID                                               uint16
	PosX, PosY, PosZ                                    float32
	Str, Dex, Rec, Int, Wis, Luc                        uint16
	HP, MP, SP                                          uint16
}
```

</details>

### db.go

[server/internal/db/db.go](../../server/internal/db/db.go#L1)

Package db owns the portable SQLite store. Original EP table/column names
are retained; database names become table prefixes in the single file.

Direct GX dependencies: `server/internal/config`, `server/internal/filelock`, `server/internal/telemetry`.

| Declaration | Responsibility / source comment |
|---|---|
| [`UserData`](../../server/internal/db/db.go#L24)<br>`const UserData = "PS_UserData"` | See the source definition and the subsystem ownership contract. |
| [`GameData`](../../server/internal/db/db.go#L25)<br>`const GameData = "PS_GameData"` | See the source definition and the subsystem ownership contract. |
| [`GameDefs`](../../server/internal/db/db.go#L26)<br>`const GameDefs = "PS_GameDefs"` | See the source definition and the subsystem ownership contract. |
| [`Timeout`](../../server/internal/db/db.go#L28)<br>`const Timeout = 10 * time.Second` | See the source definition and the subsystem ownership contract. |
| [`mu`](../../server/internal/db/db.go#L31)<br>`var mu sync.RWMutex` | See the source definition and the subsystem ownership contract. |
| [`pools`](../../server/internal/db/db.go#L32)<br>`var pools` | See the source definition and the subsystem ownership contract. |
| [`Initialize`](../../server/internal/db/db.go#L37)<br>`func Initialize() error` | Initialize copies the versioned seed once. A process lock and atomic rename prevent login/world from observing a partially copied database at startup. |
| [`runtimeExists`](../../server/internal/db/db.go#L83)<br>`func runtimeExists(path string) (bool, error)` | See the source definition and the subsystem ownership contract. |
| [`Pool`](../../server/internal/db/db.go#L99)<br>`func Pool(database string) (*sql.DB, error)` | Pool shares one durable writer per process. SQLite coordinates the login and world writers through WAL; each process serializes its own write transactions. |
| [`openPool`](../../server/internal/db/db.go#L138)<br>`func openPool(name string) *sql.DB` | See the source definition and the subsystem ownership contract. |
| [`ReadPool`](../../server/internal/db/db.go#L146)<br>`func ReadPool() (*sql.DB, error)` | ReadPool serves queries whose decisions do not require the writer lock. Always revalidate state inside the write transaction before committing. |
| [`Open`](../../server/internal/db/db.go#L172)<br>`func Open(databases ...string) error` | See the source definition and the subsystem ownership contract. |
| [`ctx`](../../server/internal/db/db.go#L180)<br>`func ctx() (context.Context, context.CancelFunc)` | See the source definition and the subsystem ownership contract. |
| [`longCtx`](../../server/internal/db/db.go#L183)<br>`func longCtx() (context.Context, context.CancelFunc)` | See the source definition and the subsystem ownership contract. |
| [`Close`](../../server/internal/db/db.go#L189)<br>`func Close() error` | Close runs bounded planner maintenance and closes process-owned connections. Call only after network handlers have drained. |
| [`Observe`](../../server/internal/db/db.go#L209)<br>`func Observe()` | Observe reports bounded database contention and WAL size without acquiring the writer connection for checkpointing. Maintenance never truncates a live WAL. |
| [`dbWaits`](../../server/internal/db/db.go#L230)<br>`var dbWaits` | See the source definition and the subsystem ownership contract. |
| [`dbWaitNanos`](../../server/internal/db/db.go#L231)<br>`var dbWaitNanos` | See the source definition and the subsystem ownership contract. |
| [`dbOpen`](../../server/internal/db/db.go#L232)<br>`var dbOpen` | See the source definition and the subsystem ownership contract. |
| [`dbWAL`](../../server/internal/db/db.go#L233)<br>`var dbWAL` | See the source definition and the subsystem ownership contract. |
| [`dbReaderWaits`](../../server/internal/db/db.go#L235)<br>`var dbReaderWaits` | See the source definition and the subsystem ownership contract. |
| [`dbReaderWaitNanos`](../../server/internal/db/db.go#L236)<br>`var dbReaderWaitNanos` | See the source definition and the subsystem ownership contract. |

### defs.go

[server/internal/db/defs.go](../../server/internal/db/defs.go#L1)

Direct GX dependencies: `internal/element`, `internal/itemstack`, `internal/potential`.

| Declaration | Responsibility / source comment |
|---|---|
| [`ItemKey`](../../server/internal/db/defs.go#L17)<br>`type ItemKey` | See the source definition and the subsystem ownership contract. |
| [`ItemDef`](../../server/internal/db/defs.go#L20)<br>`type ItemDef` | ItemDef is one PS_GameDefs.dbo.Items row. |
| [`MaxStack`](../../server/internal/db/defs.go#L53)<br>`func (d *ItemDef) MaxStack() uint16` | MaxStack is the stack size (Items.Count, at least 1). |
| [`SkillKey`](../../server/internal/db/defs.go#L55)<br>`type SkillKey` | See the source definition and the subsystem ownership contract. |
| [`SkillDef`](../../server/internal/db/defs.go#L63)<br>`type SkillDef` | SkillDef is one PS_GameDefs.dbo.Skills row. ClassEligible is indexed by Chars.Job: Attackfighter, Defensefighter, Patrolrogue, Shootrogue, Attackmage, Defensemage. EP4.5 has three AbilityType/Value pairs. |
| [`MobAttack`](../../server/internal/db/defs.go#L118)<br>`type MobAttack` | MobAttack is one of a mob's three attack slots (AttackTypeN..AttackokN). |
| [`MobDef`](../../server/internal/db/defs.go#L129)<br>`type MobDef` | MobDef is the subset of PS_GameDefs.dbo.Mobs the world uses. |
| [`MobDrop`](../../server/internal/db/defs.go#L162)<br>`type MobDrop` | See the source definition and the subsystem ownership contract. |
| [`ProductItem`](../../server/internal/db/defs.go#L169)<br>`type ProductItem` | See the source definition and the subsystem ownership contract. |
| [`Product`](../../server/internal/db/defs.go#L174)<br>`type Product` | See the source definition and the subsystem ownership contract. |
| [`StatusKey`](../../server/internal/db/defs.go#L182)<br>`type StatusKey` | StatusKey indexes StatusDefs by level and job (Chars.Job). |
| [`Status`](../../server/internal/db/defs.go#L188)<br>`type Status` | Status is a StatusDefs row: base max HP/SP/MP for a level and job. |
| [`BaseItem`](../../server/internal/db/defs.go#L192)<br>`type BaseItem` | BaseItem is a BaseItemsDefs row: an item a new character of that family and job starts with, in inventory bag 1 at Slot. |
| [`BaseItemKey`](../../server/internal/db/defs.go#L199)<br>`type BaseItemKey` | BaseItemKey indexes BaseItemsDefs by family (race) and job. |
| [`Defs`](../../server/internal/db/defs.go#L202)<br>`type Defs` | Defs holds every PS_GameDefs table the world needs. |
| [`defsOnce`](../../server/internal/db/defs.go#L217)<br>`var defsOnce sync.Once` | See the source definition and the subsystem ownership contract. |
| [`defs`](../../server/internal/db/defs.go#L218)<br>`var defs *Defs` | See the source definition and the subsystem ownership contract. |
| [`defsErr`](../../server/internal/db/defs.go#L219)<br>`var defsErr error` | See the source definition and the subsystem ownership contract. |
| [`SetDefsForTest`](../../server/internal/db/defs.go#L223)<br>`func SetDefsForTest(d *Defs)` | SetDefsForTest installs definitions without a database (unit tests). |
| [`ReloadDefsForTest`](../../server/internal/db/defs.go#L229)<br>`func ReloadDefsForTest() (*Defs, error)` | ReloadDefsForTest replaces test definitions with the database's again. |
| [`Definitions`](../../server/internal/db/defs.go#L236)<br>`func Definitions() (*Defs, error)` | Definitions returns the cached definitions, loading them on first use. |
| [`MustDefs`](../../server/internal/db/defs.go#L242)<br>`func MustDefs() *Defs` | MustDefs is for code paths that run only after a successful startup load. |
| [`clampU8`](../../server/internal/db/defs.go#L250)<br>`func clampU8(v int64) uint8` | See the source definition and the subsystem ownership contract. |
| [`clampU16`](../../server/internal/db/defs.go#L254)<br>`func clampU16(v int64) uint16` | See the source definition and the subsystem ownership contract. |
| [`normalizeSkill`](../../server/internal/db/defs.go#L259)<br>`func normalizeSkill(s *SkillDef, defenceType, fixRange uint8, keep uint32, aggro uint16, need [15]uint8, shield uint8)` | normalizeSkill applies CGameData::LoadConstSkill's load-time rules. |
| [`MobRespawnSeconds`](../../server/internal/db/defs.go#L305)<br>`func MobRespawnSeconds(special uint8) uint64` | MobRespawnSeconds maps Mobs.Attackspecial3 to the respawn delay. |
| [`loadDefs`](../../server/internal/db/defs.go#L316)<br>`func loadDefs() (*Defs, error)` | See the source definition and the subsystem ownership contract. |
| [`loadSkillCatalog`](../../server/internal/db/defs.go#L588)<br>`func loadSkillCatalog(ctx context.Context, pool *sql.DB, table, levelColumn string) (map[SkillKey]*SkillDef, error)` | See the source definition and the subsystem ownership contract. |
| [`Item`](../../server/internal/db/defs.go#L646)<br>`func (d *Defs) Item(itemType, typeID uint8) *ItemDef` | Item returns the definition for (type, typeID), or nil. |
| [`Skill`](../../server/internal/db/defs.go#L649)<br>`func (d *Defs) Skill(skillID uint16, level uint8) *SkillDef` | Skill returns the definition for (skillID, level), or nil. |
| [`NPCSkillWireLevel`](../../server/internal/db/defs.go#L654)<br>`const NPCSkillWireLevel = 100` | NPCSkillWireLevel distinguishes NPC effects from player effects on the wire. |
| [`NPCSkill`](../../server/internal/db/defs.go#L657)<br>`func (d *Defs) NPCSkill(skillID uint16) *SkillDef` | NPCSkill selects rank one from the independent NPC catalog. |
| [`EffectSkill`](../../server/internal/db/defs.go#L662)<br>`func (d *Defs) EffectSkill(skillID uint16, wireLevel uint8) *SkillDef` | EffectSkill resolves a stored or transmitted effect's level discriminator. |
| [`MobQuestItem`](../../server/internal/db/defs.go#L671)<br>`func MobQuestItem(index int32) ItemKey` | MobQuestItem decodes LoadConstMob's quest-only index. This is not the ordinary Type*1000+TypeID format. Native item IDs are truncated to a byte. |

<details><summary>Type definitions and owned fields</summary>

```go
type ItemKey struct{ Type, TypeID uint8 }
```

```go
type ItemDef struct {
	ItemID                                 int32
	Name                                   string
	Type, TypeID                           uint8
	ReqLevel                               uint16
	Country                                uint8
	Classes                                [6]bool // Attackfighter..Defensemage, indexed by Job
	Grow                                   uint8
	ReqStr, ReqDex, ReqRec, ReqInt, ReqWis int16
	ReqLuc                                 int32
	ReqIg                                  uint8
	Range                                  int16
	AttackTime                             uint8
	Speed                                  uint8 // native equipment / Sonic movement tier bonus
	Special                                uint8
	Attrib                                 uint8
	Slot                                   uint8
	Quality                                uint16
	Effect1, Effect2, Effect3, Effect4     int16
	ConstHP, ConstSP, ConstMP              int16
	ConstStr, ConstDex, ConstRec           int16
	ConstInt, ConstWis, ConstLuc           int16
	Buy, Sell                              int32
	Grade                                  uint16
	Drop                                   uint16 // drop-limit interval selector (0 = unlimited)
	Server                                 uint8
	Count                                  uint16
	ReqVg                                  int16 // linking hammer bonus points / extraction hammer success percent
	ReqOg                                  uint8 // stItemInfo byOG: not storable in the guild bank
	ExpAbsorb                              uint8 // Items.Exp: damage absorbed while equipped
}
```

```go
type SkillKey struct {
	SkillID uint16
	Level   uint8
}
```

```go
type SkillDef struct {
	SkillID        uint16
	SkillLevel     uint8
	Name           string
	Country        uint8
	ReqLevel       uint16
	SkillPointCost uint8
	PrevSkillID    uint16
	// ChangeType 101 and ChangeLevel: a transformation into the monster of that id (Degeneration).
	ChangeType, ChangeLevel                  uint16
	ClassEligible                            [6]bool
	MPCost, SPCost                           uint16
	ResetTimeSeconds                         uint16
	DamageHP, AddDamageHP                    int32
	HealHP, HealSP, HealMP                   int32
	TypeDetail                               uint16
	KeepTime                                 uint16
	TimeHealHP, TimeHealSP, TimeHealMP       int32
	ReadyTime                                uint8
	AbilityType                              [3]uint8
	AbilityValue                             [3]int16
	TimeDamageHP, TimeDamageSP, TimeDamageMP int32
	AttackRange                              uint8
	TypeAttack                               uint8
	TargetType, ApplyRange, MultiAttack      uint8

	// CGameData::LoadConstSkill (0x004bffe0) normalizes these on load.
	Kind           uint8 // clamp(TypeAttack, 1, 3) - 1: 0 melee, 1 ranged, 2 magic
	TypeEffect     uint8 // 0 passive, 1 attack, 2 heal, 3 debuff, 4 buff, ...
	DamageType     uint8 // AttackDamage mode 0..5
	TimeDamageType uint8 // TimeDamage mode 0, 3, 4, 5
	TypeShow       uint8 // 0: not learnable
	DamageSP       int32
	DamageMP       int32
	AddDamageSP    int32
	AddDamageMP    int32
	AttrType       uint8 // element override, 0 when > 8
	StateType      uint8 // 0 when > 15
	SuccessType    uint8
	SuccessValue   uint8
	Weapon1        uint8 // passive weapon types, at most 19
	Weapon2        uint8
	WeaponValue    uint8
	DefenceType    uint8 // clamp(DefenceType, 1, 3) - 1
	DefenceValue   uint8
	Aggro          uint16 // Arrow
	LimitHP        uint8
	KeepSeconds    uint32 // KeepTime, in hours when FixRange is 1
	KeepOnDeath    bool   // FixRange != 0: survives CUser::ClearApplySkillByDeath
	NeedWeapon     [15]bool
	NeedShield     bool
	MobSkill       bool // SkillLevel 100 rows are mob skills
}
```

```go
type MobAttack struct {
	Type         uint16 // 0 melee, 1 ranged, 2 magic
	TimeMS       uint32
	Range        uint16
	Attack, Plus uint16 // base damage and random variance
	Attrib       uint8
	Special      uint8
	OK           uint8 // 0 unused, 1/2 always, 3 HP 20-40%, 4 HP <= 5%, 5 HP < 20%, 6 target the weakest attacker
}
```

```go
type MobDef struct {
	MobID              uint16
	Name               string
	Level              uint16
	MaxHP              uint32
	MaxMP, MaxSP       uint32
	MoneyMin, MoneyMax uint32 // Money1 (base) and Money2 (random range)
	RespawnSeconds     uint64 // 0: does not respawn
	Exp                uint32
	Attack             uint32
	AI                 uint8
	ChaseRange         float32
	Dex, Wis, Luc      uint16
	Defense, Magic     uint16
	ShieldDefense      uint32 // Attacktime3, as the original reads it
	Attrib             uint8
	Size               uint8
	QuestItemID        int32 // Native quest index: 1..1000 -> type 27, then 28/29.
	NormalTimeMS       uint32
	NormalStep         uint16
	ChaseTimeMS        uint32
	ChaseStep          uint16
	Attacks            [3]MobAttack
	// CGameData::LoadConstMob reuses attack slot 3: Attack3 is the 5 s
	// recovery (nRecoverValue), Attackplus3 != 0 an obelisk, Attackattrib3
	// the country (1 light -> 0, 2 fury -> 1, otherwise 2: none).
	RecoverValue int32
	Obelisk      bool
	Country      uint8
	ResistState  [15]uint8 // immune to skills of StateType i+1 when non-zero
	ResistSkill  [6]uint16 // immune to these skill IDs
}
```

```go
type MobDrop struct {
	MobID     uint16
	ItemOrder uint8
	Grade     uint16
	DropRate  uint32
}
```

```go
type ProductItem struct {
	Type, TypeID uint8
	Count        uint16
}
```

```go
type Product struct {
	Code    string
	Name    string
	BuyCost uint32
	Items   []ProductItem
}
```

```go
type StatusKey struct {
	Level uint16
	Job   uint8
}
```

```go
type Status struct{ HP, SP, MP uint32 }
```

```go
type BaseItem struct {
	Type, TypeID, Slot uint8
	Count              uint16
	Quality            uint16
}
```

```go
type BaseItemKey struct{ Family, Job uint8 }
```

```go
type Defs struct {
	Items        map[ItemKey]*ItemDef
	ItemsByGrade map[uint16][]*ItemDef
	Skills       map[SkillKey]*SkillDef
	NPCSkills    map[SkillKey]*SkillDef
	Mobs         map[uint16]*MobDef
	MobDrops     map[uint16][]MobDrop
	Exp          map[uint16]uint32
	LevelCap     uint16
	Products     map[string]*Product
	Status       map[StatusKey]Status
	BaseItems    map[BaseItemKey][]BaseItem
}
```

</details>

### enchant_legacy.go

[server/internal/db/enchant_legacy.go](../../server/internal/db/enchant_legacy.go#L1)

Direct GX dependencies: `internal/catalogfiles`, `internal/enchantrules`, `server/internal/config`.

| Declaration | Responsibility / source comment |
|---|---|
| [`ReadLegacyEnchant`](../../server/internal/db/enchant_legacy.go#L17)<br>`func ReadLegacyEnchant(path string) (enchantrules.Catalog, error)` | ReadLegacyEnchant is used only to initialize a pre-migration installation. |
| [`initializeEnchant11`](../../server/internal/db/enchant_legacy.go#L23)<br>`func initializeEnchant11(ctx context.Context, tx *sql.Tx) error` | initializeEnchant11 runs inside migration 11 and never imports an INI again after that migration is recorded, including when both caps are later zero. |

### errors.go

[server/internal/db/errors.go](../../server/internal/db/errors.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`StorageUnavailable`](../../server/internal/db/errors.go#L12)<br>`func StorageUnavailable(err error) bool` | StorageUnavailable distinguishes an unsafe storage failure from a refused gameplay plan or a transient busy database. The world must stop admitting mutations if durable storage is no longer writable or trustworthy. |
| [`failureCallback`](../../server/internal/db/errors.go#L24)<br>`type failureCallback` | See the source definition and the subsystem ownership contract. |
| [`failureHandler`](../../server/internal/db/errors.go#L26)<br>`var failureHandler atomic.Pointer[failureCallback]` | See the source definition and the subsystem ownership contract. |
| [`OnStorageFailure`](../../server/internal/db/errors.go#L28)<br>`func OnStorageFailure(handler func(error)) func()` | See the source definition and the subsystem ownership contract. |
| [`reportStorageFailure`](../../server/internal/db/errors.go#L33)<br>`func reportStorageFailure(err error)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type failureCallback struct{ run func(error) }
```

</details>

### faction.go

[server/internal/db/faction.go](../../server/internal/db/faction.go#L1)

Direct GX dependencies: `server/internal/config`.

| Declaration | Responsibility / source comment |
|---|---|
| [`FactionNone`](../../server/internal/db/faction.go#L10)<br>`const FactionNone = 2` | FactionNone is reported while the account has not chosen a faction yet. |
| [`SaveCountry`](../../server/internal/db/faction.go#L14)<br>`func SaveCountry(userUID int32, country uint8) error` | SaveCountry runs PS_GameData.dbo.usp_Save_Char_Country_E, inserting the UserMaxGrow row the first time an account picks a faction. |

### gameplay.go

[server/internal/db/gameplay.go](../../server/internal/db/gameplay.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`sqlExecutor`](../../server/internal/db/gameplay.go#L9)<br>`type sqlExecutor` | See the source definition and the subsystem ownership contract. |
| [`CharacterChange`](../../server/internal/db/gameplay.go#L15)<br>`type CharacterChange` | CharacterChange is a plan made under the owning sessions' state locks. Nothing is published in memory or to clients until CommitGameplay succeeds. |
| [`QuestCompletion`](../../server/internal/db/gameplay.go#L21)<br>`type QuestCompletion` | See the source definition and the subsystem ownership contract. |
| [`CommitGameplay`](../../server/internal/db/gameplay.go#L29)<br>`func CommitGameplay(changes []CharacterChange, quest *QuestCompletion) error` | CommitGameplay makes both trade participants, or one quest's entire reward, durable together. Expected inventory and money protect against stale plans. |
| [`commitGameplayTx`](../../server/internal/db/gameplay.go#L39)<br>`func commitGameplayTx(c context.Context, tx *sql.Tx, changes []CharacterChange, quest *QuestCompletion) error` | See the source definition and the subsystem ownership contract. |
| [`requireOne`](../../server/internal/db/gameplay.go#L94)<br>`func requireOne(result sql.Result) error` | See the source definition and the subsystem ownership contract. |
| [`sameItemState`](../../server/internal/db/gameplay.go#L105)<br>`func sameItemState(a, b Item) bool` | See the source definition and the subsystem ownership contract. |
| [`sameInventory`](../../server/internal/db/gameplay.go#L109)<br>`func sameInventory(a, b []Item) bool` | See the source definition and the subsystem ownership contract. |
| [`writeInventoryChange`](../../server/internal/db/gameplay.go#L130)<br>`func writeInventoryChange(c context.Context, tx *sql.Tx, change CharacterChange) error` | See the source definition and the subsystem ownership contract. |
| [`CommitItems`](../../server/internal/db/gameplay.go#L192)<br>`func CommitItems(charID uint32, before, after []Item, consumedSlots ...[2]uint8) error` | CommitItems applies an inventory-only plan without overwriting character money/stats. Used by item grants whose character state is owned elsewhere. |
| [`CommitStatusItem`](../../server/internal/db/gameplay.go#L212)<br>`func CommitStatusItem(change CharacterChange, buffs []AppliedBuff) error` | CommitStatusItem saves the consumed item and remaining effects together. |

<details><summary>Type definitions and owned fields</summary>

```go
type sqlExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}
```

```go
type CharacterChange struct {
	Before, After           Character
	ItemsBefore, ItemsAfter []Item
	ConsumedSlots           [][2]uint8 // bag/slot pairs whose last unit was consumed
}
```

```go
type QuestCompletion struct {
	CharacterID uint32
	QuestID     uint16
	Repeatable  bool
}
```

</details>

### gm.go

[server/internal/db/gm.go](../../server/internal/db/gm.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`ApplyGMItemGrant`](../../server/internal/db/gm.go#L8)<br>`func ApplyGMItemGrant(charID uint32, changed, added []Item) error` | ApplyGMItemGrant atomically persists a planned grant. Existing items may only change count; new items receive unique IDs before becoming visible. The caller installs its inventory copy only after this transaction succeeds. |

### guild_creation.go

[server/internal/db/guild_creation.go](../../server/internal/db/guild_creation.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`CreateGuildGroup`](../../server/internal/db/guild_creation.go#L11)<br>`func CreateGuildGroup(id uint32, name, masterUserID, remark string, country uint8, master Character, fee uint32, members []GuildCharRow) error` | CreateGuildGroup commits the group, membership, mirror and master's fee in one transaction. The world holds the master's state lock until publication. |

### guildbank.go

[server/internal/db/guildbank.go](../../server/internal/db/guildbank.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`CommitGuildBank`](../../server/internal/db/guildbank.go#L9)<br>`func CommitGuildBank(guildID uint32, change CharacterChange, before, after []Item) error` | See the source definition and the subsystem ownership contract. |
| [`writeGuildBankChange`](../../server/internal/db/guildbank.go#L33)<br>`func writeGuildBankChange(c context.Context, tx *sql.Tx, guildID uint32, before, after []Item) error` | See the source definition and the subsystem ownership contract. |

### guildhouse.go

[server/internal/db/guildhouse.go](../../server/internal/db/guildhouse.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`GuildNpcRow`](../../server/internal/db/guildhouse.go#L10)<br>`type GuildNpcRow` | GuildNpcRow is one usp_Read_GuildNpcLv_R row: a guild house NPC and its level in the guild's slot (Number). |
| [`LoadGuildNpcLevels`](../../server/internal/db/guildhouse.go#L18)<br>`func LoadGuildNpcLevels() ([]GuildNpcRow, error)` | LoadGuildNpcLevels reads every live guild NPC level (usp_Read_GuildNpcLv_R). |
| [`CommitGuildHouse`](../../server/internal/db/guildhouse.go#L47)<br>`func CommitGuildHouse(g GuildRow, npc *GuildNpcRow, change *CharacterChange) error` | CommitGuildHouse stores a guild house change atomically: GuildDetails' house fields (the DB agent's 0x0B0F/0x0B10/0x0B12 writes), optionally the NPC level row (usp_Save_GuildNpcLv_Add_E / _Mod_E: the row of that type takes the level and slot) and the character's own change (money or the consumed etin items). |
| [`saveGuildHouseTx`](../../server/internal/db/guildhouse.go#L74)<br>`func saveGuildHouseTx(c context.Context, tx *sql.Tx, g GuildRow) error` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type GuildNpcRow struct {
	GuildID uint32
	Type    uint16
	Level   uint8
	Slot    uint8
}
```

</details>

### guilds.go

[server/internal/db/guilds.go](../../server/internal/db/guilds.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`mirrorGuildTx`](../../server/internal/db/guilds.go#L24)<br>`func mirrorGuildTx(c context.Context, tx *sql.Tx, guildID uint32) error` | mirrorGuild is what Trig_Guilds_Insert/Update/Delete did through the linked server PS_USERDB01 (disabled, see tools/sqlserver.py prepare): PS_UserData.dbo.CreatedGuilds holds a copy of the Guilds row for server 1. The EP4.5 CreatedGuilds also has a NOT NULL UpdateDate the triggers never set; it gets the time of the change. |
| [`guildOfMember`](../../server/internal/db/guilds.go#L35)<br>`func guildOfMember(charID uint32) uint32` | guildOfMember is the live guild of a character, for the mirror after procedures keyed by CharID. |
| [`GuildRow`](../../server/internal/db/guilds.go#L48)<br>`type GuildRow` | GuildRow is one usp_Read_Guild2_R row. |
| [`GuildCharRow`](../../server/internal/db/guilds.go#L67)<br>`type GuildCharRow` | GuildCharRow is one usp_Read_GuildChar_R row. |
| [`LoadGuilds`](../../server/internal/db/guilds.go#L77)<br>`func LoadGuilds() ([]GuildRow, []GuildCharRow, error)` | LoadGuilds reads every live guild and its members. |
| [`NextGuildID`](../../server/internal/db/guilds.go#L127)<br>`func NextGuildID() (uint32, error)` | MaxGuildID is usp_Read_MaxGuildID_R: the next free guild id. |
| [`execStatus`](../../server/internal/db/guilds.go#L139)<br>`func execStatus(query string, args ...any) (int, error)` | See the source definition and the subsystem ownership contract. |
| [`CreateGuild`](../../server/internal/db/guilds.go#L155)<br>`func CreateGuild(id uint32, name, masterUserID string, masterCharID uint32, masterName string, country uint8, remark string) error` | CreateGuild is usp_Create_Guild2_E (0 on success). |
| [`AddGuildChar`](../../server/internal/db/guilds.go#L183)<br>`func AddGuildChar(guildID, charID uint32, rank uint8) error` | AddGuildChar is usp_Create_GuildChar_E (rank 0 for an applicant). |
| [`SaveGuildCharRank`](../../server/internal/db/guilds.go#L211)<br>`func SaveGuildCharRank(charID uint32, rank uint8) error` | SaveGuildCharRank is usp_Save_GuildChar_E (rank 1 also moves the master). |
| [`DeleteGuildChar`](../../server/internal/db/guilds.go#L238)<br>`func DeleteGuildChar(guildID, charID uint32) error` | DeleteGuildChar is usp_Delete_GuildChar_E. |
| [`DeleteLastGuildChar`](../../server/internal/db/guilds.go#L239)<br>`func DeleteLastGuildChar(guildID, charID uint32) error` | See the source definition and the subsystem ownership contract. |
| [`deleteGuildChar`](../../server/internal/db/guilds.go#L240)<br>`func deleteGuildChar(guildID, charID uint32, disband bool) error` | See the source definition and the subsystem ownership contract. |
| [`DeleteGuild`](../../server/internal/db/guilds.go#L276)<br>`func DeleteGuild(guildID uint32) error` | DeleteGuild is usp_Delete_Guild_E. |
| [`SaveGuildRemark`](../../server/internal/db/guilds.go#L296)<br>`func SaveGuildRemark(guildID uint32, remark string) error` | SaveGuildRemark stores the notice in GuildDetails (usp_Save_GuildRemark_E's update or insert). |
| [`SaveGuildRanking`](../../server/internal/db/guilds.go#L302)<br>`func SaveGuildRanking(g GuildRow) error` | SaveGuildRanking is usp_Save_GuildInfo_E2 (the DB agent's 0x0B16 batch): guild points, rank and the house fields, then the CreatedGuilds copy. |
| [`ReadGodBless`](../../server/internal/db/guilds.go#L325)<br>`func ReadGodBless() (light, dark uint32, err error)` | ReadGodBless is usp_Read_GodBless_R (ps_dbAgent CWorld::DBKillCountLoad): the light and dark country kill counts, 0 without a row. |
| [`ErrNoWorldInfo`](../../server/internal/db/guilds.go#L345)<br>`var ErrNoWorldInfo` | ErrNoWorldInfo is usp_Save_GodBless_E's -1: PS_GameData.dbo.WorldInfo has no row yet (the world calendar creates it). |
| [`SaveGodBless`](../../server/internal/db/guilds.go#L349)<br>`func SaveGodBless(light, dark uint32) error` | SaveGodBless is usp_Save_GodBless_E (CWorld::DBKillCountSave): 1 saved, -1 without a WorldInfo row. |
| [`SaveWorldInfo`](../../server/internal/db/guilds.go#L359)<br>`func SaveWorldInfo(lastWorldTime int32) error` | SaveWorldInfo is usp_Save_WorldInfo_E (the DB agent's world calendar save), which creates the WorldInfo row when missing. |
| [`ReadWorldInfo`](../../server/internal/db/guilds.go#L385)<br>`func ReadWorldInfo() (int32, error)` | ReadWorldInfo is usp_Read_WorldInfo_E: the packed world calendar base (0 without a row). |

<details><summary>Type definitions and owned fields</summary>

```go
type GuildRow struct {
	ID         uint32
	Name       string
	MasterName string
	Country    uint8
	Points     uint32
	Rank       uint8
	Etin       uint32
	Remark     string
	BuyHouse   uint8
	UseHouse   uint8
	EtinReturn uint32
	KeepEtin   uint32
	// MasterAdmin: the guild master's account is a GM (test guilds keep an
	// always-active house).
	MasterAdmin bool
}
```

```go
type GuildCharRow struct {
	CharID  uint32
	Name    string
	Level   uint16
	Job     uint8
	GuildID uint32
	Rank    uint8 // GuildLevel
}
```

</details>

### items.go

[server/internal/db/items.go](../../server/internal/db/items.go#L1)

Direct GX dependencies: `server/internal/config`.

| Declaration | Responsibility / source comment |
|---|---|
| [`Item`](../../server/internal/db/items.go#L18)<br>`type Item` | Item is one CharItems or UserStoredItems row; UID (ItemUID) is the key the original procedures use. |
| [`ItemID`](../../server/internal/db/items.go#L31)<br>`func (it *Item) ItemID() int32` | ItemID is Type*1000+TypeID, the original encoding. |
| [`NewItemUID`](../../server/internal/db/items.go#L34)<br>`func NewItemUID() int64` | NewItemUID returns a random positive 63-bit ItemUID. |
| [`scanItems`](../../server/internal/db/items.go#L42)<br>`func scanItems(rows *sql.Rows, stored bool) ([]Item, error)` | See the source definition and the subsystem ownership contract. |
| [`CharacterItems`](../../server/internal/db/items.go#L75)<br>`func CharacterItems(charID uint32) ([]Item, error)` | CharacterItems runs usp_Read_Char_Items_Detail_R. |
| [`itemQuerier`](../../server/internal/db/items.go#L85)<br>`type itemQuerier` | See the source definition and the subsystem ownership contract. |
| [`characterItems`](../../server/internal/db/items.go#L89)<br>`func characterItems(c context.Context, query itemQuerier, charID uint32) ([]Item, error)` | See the source definition and the subsystem ownership contract. |
| [`makeTime`](../../server/internal/db/items.go#L108)<br>`func makeTime(it *Item) string` | See the source definition and the subsystem ownership contract. |
| [`makeType`](../../server/internal/db/items.go#L116)<br>`func makeType(it *Item) string` | See the source definition and the subsystem ownership contract. |
| [`AddCharacterItem`](../../server/internal/db/items.go#L124)<br>`func AddCharacterItem(charID uint32, it *Item) error` | AddCharacterItem runs usp_Save_Char_Item_Add_E; a zero UID gets a new one. |
| [`insertCharacterItem`](../../server/internal/db/items.go#L134)<br>`func insertCharacterItem(c context.Context, executor sqlExecutor, charID uint32, it *Item) error` | See the source definition and the subsystem ownership contract. |
| [`SaveCharacterItem`](../../server/internal/db/items.go#L157)<br>`func SaveCharacterItem(charID uint32, it *Item) error` | SaveCharacterItem writes bag, slot, quality, gems, count and craft name with usp_Save_Char_Item_Mod_E. A zero count deletes the item. |
| [`idList`](../../server/internal/db/items.go#L172)<br>`func idList(uids []int64) string` | See the source definition and the subsystem ownership contract. |
| [`DeleteCharacterItems`](../../server/internal/db/items.go#L181)<br>`func DeleteCharacterItems(charID uint32, uids ...int64) error` | DeleteCharacterItems runs usp_Save_Char_Item_Del_E for the given UIDs. |
| [`StoredItems`](../../server/internal/db/items.go#L190)<br>`func StoredItems(userUID uint32) ([]Item, error)` | StoredItems runs usp_Read_User_StoredItems_Detail_R (account warehouse). |
| [`storedItems`](../../server/internal/db/items.go#L199)<br>`func storedItems(c context.Context, query itemQuerier, userUID uint32) ([]Item, error)` | See the source definition and the subsystem ownership contract. |
| [`AddStoredItem`](../../server/internal/db/items.go#L222)<br>`func AddStoredItem(userUID uint32, it *Item) error` | AddStoredItem runs usp_Save_User_StoredItem_Add_E. |
| [`SaveStoredItem`](../../server/internal/db/items.go#L240)<br>`func SaveStoredItem(userUID uint32, it *Item) error` | SaveStoredItem runs usp_Save_User_StoredItem_Mod_E; zero count deletes. |
| [`DeleteStoredItems`](../../server/internal/db/items.go#L256)<br>`func DeleteStoredItems(userUID uint32, uids ...int64) error` | DeleteStoredItems runs usp_Save_User_StoredItem_Del_E. |
| [`GuildStoredItems`](../../server/internal/db/items.go#L265)<br>`func GuildStoredItems(guildID uint32) ([]Item, error)` | GuildStoredItems runs usp_Read_Guild_StoredItems_Detail_R. |
| [`guildStoredItems`](../../server/internal/db/items.go#L274)<br>`func guildStoredItems(c context.Context, query itemQuerier, guildID uint32) ([]Item, error)` | See the source definition and the subsystem ownership contract. |
| [`AddGuildStoredItem`](../../server/internal/db/items.go#L296)<br>`func AddGuildStoredItem(guildID uint32, it *Item) error` | AddGuildStoredItem runs usp_Save_Guild_StoredItem_Add_E. |
| [`insertGuildStoredItem`](../../server/internal/db/items.go#L305)<br>`func insertGuildStoredItem(c context.Context, executor sqlExecutor, guildID uint32, it *Item) error` | See the source definition and the subsystem ownership contract. |
| [`SaveGuildStoredItem`](../../server/internal/db/items.go#L324)<br>`func SaveGuildStoredItem(guildID uint32, it *Item) error` | SaveGuildStoredItem runs usp_Save_Guild_StoredItem_Mod_E. |
| [`DeleteGuildStoredItems`](../../server/internal/db/items.go#L336)<br>`func DeleteGuildStoredItems(guildID uint32, uids ...int64) error` | DeleteGuildStoredItems runs usp_Save_Guild_StoredItem_Del_E. |

<details><summary>Type definitions and owned fields</summary>

```go
type Item struct {
	UID          int64
	Bag, Slot    uint8
	Type, TypeID uint8
	Quality      uint16
	Gems         [6]uint8
	Count        uint16
	CraftName    string
	MakeTime     time.Time
	MakeType     string
}
```

```go
type itemQuerier interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}
```

</details>

### keeps.go

[server/internal/db/keeps.go](../../server/internal/db/keeps.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`Keep`](../../server/internal/db/keeps.go#L8)<br>`type Keep` | Keep mirrors the original Keeps row / DB-agent 0206 and 0207 record. Country packs runtime owner (low nibble) and last conqueror+1 (high nibble). ResetTime is the original packed calendar value, not a Unix timestamp. |
| [`ReadKeeps`](../../server/internal/db/keeps.go#L14)<br>`func ReadKeeps() (map[uint32]Keep, error)` | See the source definition and the subsystem ownership contract. |
| [`SaveKeeps`](../../server/internal/db/keeps.go#L42)<br>`func SaveKeeps(keeps []Keep) error` | SaveKeeps updates/creates only the supplied identities, atomically. Unrelated keeps and all account/character data survive a world restart or capture. |

<details><summary>Type definitions and owned fields</summary>

```go
type Keep struct {
	ID        uint32
	Country   uint8
	ResetTime uint32
}
```

</details>

### legacy_backup.go

[server/internal/db/legacy_backup.go](../../server/internal/db/legacy_backup.go#L1)

Direct GX dependencies: `server/internal/config`, `server/internal/filelock`.

| Declaration | Responsibility / source comment |
|---|---|
| [`pruneLegacyVersion`](../../server/internal/db/legacy_backup.go#L18)<br>`const pruneLegacyVersion = 26` | See the source definition and the subsystem ownership contract. |
| [`migrateRuntime`](../../server/internal/db/legacy_backup.go#L22)<br>`func migrateRuntime(ctx context.Context, pool *sql.DB) error` | Keep the pre-prune schema available to the legacy exporters. A separate process lock serializes backup and migration across Login and World. |
| [`backupLegacyRuntime`](../../server/internal/db/legacy_backup.go#L47)<br>`func backupLegacyRuntime(ctx context.Context, pool *sql.DB, path string) error` | See the source definition and the subsystem ownership contract. |
| [`verifyLegacyBackup`](../../server/internal/db/legacy_backup.go#L92)<br>`func verifyLegacyBackup(ctx context.Context, archive string) error` | See the source definition and the subsystem ownership contract. |

### mall.go

[server/internal/db/mall.go](../../server/internal/db/mall.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`GiftBoxSlots`](../../server/internal/db/mall.go#L14)<br>`const GiftBoxSlots = 240` | See the source definition and the subsystem ownership contract. |
| [`Points`](../../server/internal/db/mall.go#L17)<br>`func Points(userUID uint32) (uint32, error)` | Points reads the account's item-mall balance. |
| [`ErrNotEnoughPoints`](../../server/internal/db/mall.go#L33)<br>`var ErrNotEnoughPoints` | ErrNotEnoughPoints means the balance does not cover the cost. |
| [`SpendPoints`](../../server/internal/db/mall.go#L37)<br>`func SpendPoints(userUID, charID, cost uint32, productCode, targetName string, useType uint8) (uint32, error)` | SpendPoints guards the balance and records the original NCash PointLog in the same transaction. useType: 1 buy, 2 gift. |
| [`spendPointsTx`](../../server/internal/db/mall.go#L54)<br>`func spendPointsTx(c context.Context, tx *sql.Tx, userUID, charID, cost uint32, productCode, targetName string, useType uint8) (uint32, error)` | spendPointsTx keeps charging and delivery in the same transaction. |
| [`PurchaseMallItems`](../../server/internal/db/mall.go#L69)<br>`func PurchaseMallItems(userUID, charID uint32, product *Product, items []Item) (uint32, error)` | PurchaseMallItems atomically charges, logs and inserts a validated delivery. The caller publishes packets and updates session memory only after commit. |
| [`GiftEntry`](../../server/internal/db/mall.go#L104)<br>`type GiftEntry` | GiftEntry is one gift-box slot. |
| [`GiftBox`](../../server/internal/db/mall.go#L111)<br>`func GiftBox(userUID uint32) ([]GiftEntry, error)` | GiftBox runs usp_Read_User_StoredPointItems. |
| [`GiftOK`](../../server/internal/db/mall.go#L142)<br>`const GiftOK = 0` | GiftResult codes match the original usp_Save_User_GiftPointItems_NCash and the client's gift result byte. |
| [`GiftNoSuchChar`](../../server/internal/db/mall.go#L143)<br>`const GiftNoSuchChar = 1` | GiftResult codes match the original usp_Save_User_GiftPointItems_NCash and the client's gift result byte. |
| [`GiftSameAccount`](../../server/internal/db/mall.go#L144)<br>`const GiftSameAccount = 2` | GiftResult codes match the original usp_Save_User_GiftPointItems_NCash and the client's gift result byte. |
| [`GiftBoxFull`](../../server/internal/db/mall.go#L145)<br>`const GiftBoxFull = 3` | GiftResult codes match the original usp_Save_User_GiftPointItems_NCash and the client's gift result byte. |
| [`GiftNotEnoughSlots`](../../server/internal/db/mall.go#L146)<br>`const GiftNotEnoughSlots = 4` | GiftResult codes match the original usp_Save_User_GiftPointItems_NCash and the client's gift result byte. |
| [`GiftInsertError`](../../server/internal/db/mall.go#L147)<br>`const GiftInsertError = 5` | GiftResult codes match the original usp_Save_User_GiftPointItems_NCash and the client's gift result byte. |
| [`StoreGift`](../../server/internal/db/mall.go#L152)<br>`func StoreGift(targetName string, senderUID uint32, product *Product) (uint8, uint32, error)` | StoreGift puts every item of product into the gift box of the account owning targetName, in free slots, atomically. |
| [`PurchaseGift`](../../server/internal/db/mall.go#L158)<br>`func PurchaseGift(targetName string, senderUID, charID uint32, product *Product) (uint8, uint32, uint32, error)` | PurchaseGift charges the sender and stores every gift in one transaction. |
| [`storeGift`](../../server/internal/db/mall.go#L162)<br>`func storeGift(targetName string, senderUID, charID uint32, product *Product, charge bool) (result uint8, targetUID, points uint32, err error)` | See the source definition and the subsystem ownership contract. |
| [`ClaimGift`](../../server/internal/db/mall.go#L252)<br>`func ClaimGift(userUID, charID uint32, giftSlot uint8, it *Item) (bool, error)` | ClaimGift moves gift-box slot into the character's inventory as it, in one transaction (the effect of usp_Save_User_StoredPointItems). It returns false if the slot was already claimed. |
| [`errRollback`](../../server/internal/db/mall.go#L307)<br>`var errRollback` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type GiftEntry struct {
	Slot         uint8
	Type, TypeID uint8
	Count        uint16
}
```

</details>

### migrations.go

[server/internal/db/migrations.go](../../server/internal/db/migrations.go#L1)

Direct GX dependencies: `internal/blessingrules`, `internal/enchantrules`, `internal/itemboxes`, `internal/maprules`, `internal/mobaimessages`, `internal/npcquest`, `internal/teleportrunes`, `internal/worlddrops`.

| Declaration | Responsibility / source comment |
|---|---|
| [`migration`](../../server/internal/db/migrations.go#L22)<br>`type migration` | See the source definition and the subsystem ownership contract. |
| [`experienceMigration`](../../server/internal/db/migrations.go#L28)<br>`var experienceMigration string` | See the source definition and the subsystem ownership contract. |
| [`npcSkillsMigration`](../../server/internal/db/migrations.go#L31)<br>`var npcSkillsMigration string` | See the source definition and the subsystem ownership contract. |
| [`auctionMigration`](../../server/internal/db/migrations.go#L34)<br>`var auctionMigration string` | See the source definition and the subsystem ownership contract. |
| [`pruneLegacyMigration`](../../server/internal/db/migrations.go#L37)<br>`var pruneLegacyMigration string` | See the source definition and the subsystem ownership contract. |
| [`percentageSkillsMigration`](../../server/internal/db/migrations.go#L40)<br>`var percentageSkillsMigration string` | See the source definition and the subsystem ownership contract. |
| [`restoreConsumablesMigration`](../../server/internal/db/migrations.go#L43)<br>`var restoreConsumablesMigration string` | See the source definition and the subsystem ownership contract. |
| [`removeGrowthModesMigration`](../../server/internal/db/migrations.go#L46)<br>`var removeGrowthModesMigration string` | See the source definition and the subsystem ownership contract. |
| [`consumableStacksMigration`](../../server/internal/db/migrations.go#L49)<br>`var consumableStacksMigration string` | See the source definition and the subsystem ownership contract. |
| [`forgeHammersMigration`](../../server/internal/db/migrations.go#L52)<br>`var forgeHammersMigration string` | See the source definition and the subsystem ownership contract. |
| [`migrations`](../../server/internal/db/migrations.go#L54)<br>`var migrations` | See the source definition and the subsystem ownership contract. |
| [`migrate`](../../server/internal/db/migrations.go#L237)<br>`func migrate(ctx context.Context, pool *sql.DB, steps []migration) error` | Base user_version=1 remains compatible with the distributed development seed. GX migrations have their own transactional, checksummed ledger. |

<details><summary>Type definitions and owned fields</summary>

```go
type migration struct {
	Version   int
	Name, SQL string
}
```

</details>

### npcquest.go

[server/internal/db/npcquest.go](../../server/internal/db/npcquest.go#L1)

Direct GX dependencies: `internal/data/sdata`, `internal/npcquest`, `server/internal/config`.

| Declaration | Responsibility / source comment |
|---|---|
| [`initializeNPCQuest`](../../server/internal/db/npcquest.go#L16)<br>`func initializeNPCQuest(ctx context.Context, tx *sql.Tx) error` | See the source definition and the subsystem ownership contract. |

### operator.go

[server/internal/db/operator.go](../../server/internal/db/operator.go#L1)

Direct GX dependencies: `internal/pvprank`, `server/internal/config`, `server/internal/operatorwire`.

| Declaration | Responsibility / source comment |
|---|---|
| [`operatorAccountLimit`](../../server/internal/db/operator.go#L15)<br>`const operatorAccountLimit = 100` | See the source definition and the subsystem ownership contract. |
| [`rowQueryer`](../../server/internal/db/operator.go#L17)<br>`type rowQueryer` | See the source definition and the subsystem ownership contract. |
| [`OperatorAccounts`](../../server/internal/db/operator.go#L21)<br>`func OperatorAccounts(query string) ([]operatorwire.Account, bool, error)` | See the source definition and the subsystem ownership contract. |
| [`operatorAccount`](../../server/internal/db/operator.go#L49)<br>`func operatorAccount(c context.Context, q rowQueryer, uid uint32) (operatorwire.Account, error)` | See the source definition and the subsystem ownership contract. |
| [`OperatorAccount`](../../server/internal/db/operator.go#L59)<br>`func OperatorAccount(uid uint32) (operatorwire.Account, error)` | See the source definition and the subsystem ownership contract. |
| [`SetOperatorAccount`](../../server/internal/db/operator.go#L69)<br>`func SetOperatorAccount(expected operatorwire.Account, gm bool, points uint32) error` | See the source definition and the subsystem ownership contract. |
| [`OperatorCharacterSnapshot`](../../server/internal/db/operator.go#L98)<br>`func OperatorCharacterSnapshot(ch Character) (operatorwire.Response, error)` | OperatorCharacterSnapshot exposes the editable fields and experience bounds. |
| [`operatorEXPBounds`](../../server/internal/db/operator.go#L113)<br>`func operatorEXPBounds(level uint16) (uint32, uint32, error)` | See the source definition and the subsystem ownership contract. |
| [`OperatorStats`](../../server/internal/db/operator.go#L128)<br>`func OperatorStats(uid, charID uint32) (operatorwire.Response, error)` | See the source definition and the subsystem ownership contract. |
| [`ValidateOperatorStats`](../../server/internal/db/operator.go#L153)<br>`func ValidateOperatorStats(snapshot operatorwire.Response, expected, next operatorwire.Stats) error` | ValidateOperatorStats rejects stale forms and checks the native field bounds. |
| [`SetOperatorStats`](../../server/internal/db/operator.go#L167)<br>`func SetOperatorStats(uid, charID uint32, expected, next operatorwire.Stats) error` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type rowQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}
```

</details>

### ownership.go

[server/internal/db/ownership.go](../../server/internal/db/ownership.go#L1)

Direct GX dependencies: `server/internal/config`, `server/internal/filelock`.

| Declaration | Responsibility / source comment |
|---|---|
| [`AcquireWorld`](../../server/internal/db/ownership.go#L17)<br>`func AcquireWorld(ctx context.Context) (func(), error)` | AcquireWorld enforces one authoritative simulation process per runtime. The OS releases ownership on crash; a new owner removes stale online leases. |
| [`AcquireLogin`](../../server/internal/db/ownership.go#L36)<br>`func AcquireLogin(ctx context.Context) (func(), error)` | See the source definition and the subsystem ownership contract. |
| [`AcquireWorldAdministration`](../../server/internal/db/ownership.go#L41)<br>`func AcquireWorldAdministration(ctx context.Context) (func(), error)` | AcquireWorldAdministration prevents offline edits from racing World startup. |
| [`lockRuntime`](../../server/internal/db/ownership.go#L45)<br>`func lockRuntime(ctx context.Context, service string) (func(), error)` | See the source definition and the subsystem ownership contract. |
| [`AcquireAccount`](../../server/internal/db/ownership.go#L64)<br>`func AcquireAccount(uid uint32, id int64) (bool, error)` | See the source definition and the subsystem ownership contract. |
| [`ReleaseAccount`](../../server/internal/db/ownership.go#L78)<br>`func ReleaseAccount(uid uint32, id int64) error` | See the source definition and the subsystem ownership contract. |
| [`releases`](../../server/internal/db/ownership.go#L89)<br>`var releases` | See the source definition and the subsystem ownership contract. |
| [`RetryAccountRelease`](../../server/internal/db/ownership.go#L96)<br>`func RetryAccountRelease(uid uint32, id int64)` | Failed disconnect cleanup is retried with its exact session identity, never deleting a newer account lease. Cardinality is bounded by admission capacity. |
| [`RetryAccountReleases`](../../server/internal/db/ownership.go#L103)<br>`func RetryAccountReleases()` | See the source definition and the subsystem ownership contract. |

### quests.go

[server/internal/db/quests.go](../../server/internal/db/quests.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`CharQuest`](../../server/internal/db/quests.go#L13)<br>`type CharQuest` | CharQuest is an active quest row: Delay is minutes left (timed quests) or minutes elapsed (untimed), as the original saves it. |
| [`ErrQuestStartTerm`](../../server/internal/db/quests.go#L20)<br>`var ErrQuestStartTerm` | See the source definition and the subsystem ownership contract. |
| [`FinishedQuest`](../../server/internal/db/quests.go#L23)<br>`type FinishedQuest` | FinishedQuest is a finished quest row (Success 1 completed, 0 failed). |
| [`CharacterQuests`](../../server/internal/db/quests.go#L29)<br>`func CharacterQuests(charID uint32) ([]CharQuest, error)` | CharacterQuests runs usp_Read_Char_Quest_R. |
| [`FinishedQuests`](../../server/internal/db/quests.go#L55)<br>`func FinishedQuests(charID uint32) ([]FinishedQuest, error)` | FinishedQuests runs usp_Read_Char_FinishedQuest_R. |
| [`AddCharacterQuest`](../../server/internal/db/quests.go#L82)<br>`func AddCharacterQuest(charID uint32, q CharQuest) error` | AddCharacterQuest runs usp_Save_Char_Quest_Add_E. |
| [`CommitQuestStart`](../../server/internal/db/quests.go#L90)<br>`func CommitQuestStart(charID uint32, q CharQuest, before, after []Item, interval time.Duration) error` | CommitQuestStart commits the inventory cost, starting grants and quest row together. A duplicate start or changed inventory rolls back the whole plan. |
| [`SaveCharacterQuest`](../../server/internal/db/quests.go#L134)<br>`func SaveCharacterQuest(charID uint32, q CharQuest) error` | SaveCharacterQuest runs usp_Save_Char_Quest_Mod_E. |
| [`FinishCharacterQuest`](../../server/internal/db/quests.go#L141)<br>`func FinishCharacterQuest(charID uint32, questID uint16, success uint8) error` | FinishCharacterQuest runs usp_Save_Char_FinishedQuest_E on the active row. |
| [`DeleteCharacterQuest`](../../server/internal/db/quests.go#L146)<br>`func DeleteCharacterQuest(charID uint32, questID uint16) error` | DeleteCharacterQuest runs usp_Save_Char_Quest_Del_E (repeatable quests). |
| [`changeActiveQuest`](../../server/internal/db/quests.go#L150)<br>`func changeActiveQuest(query string, args ...any) error` | See the source definition and the subsystem ownership contract. |
| [`SaveCharacterQuests`](../../server/internal/db/quests.go#L167)<br>`func SaveCharacterQuests(charID uint32, quests []CharQuest) error` | SaveCharacterQuests atomically snapshots active counters. A stale/finished quest cannot be updated after another operation completed it. |
| [`saveQuestsTx`](../../server/internal/db/quests.go#L180)<br>`func saveQuestsTx(c context.Context, tx *sql.Tx, charID uint32, quests []CharQuest) error` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type CharQuest struct {
	QuestID   uint16
	Delay     int16
	Counts    [3]uint8
	StartedAt int64 // Unix milliseconds for GX timing, 0 for legacy rows.
}
```

```go
type FinishedQuest struct {
	QuestID uint16
	Success uint8
}
```

</details>

### skill_reset.go

[server/internal/db/skill_reset.go](../../server/internal/db/skill_reset.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`ResetSkillCooldowns`](../../server/internal/db/skill_reset.go#L8)<br>`func ResetSkillCooldowns(charID uint32) (map[uint16]time.Time, error)` | See the source definition and the subsystem ownership contract. |
| [`CommitSkillReset`](../../server/internal/db/skill_reset.go#L32)<br>`func CommitSkillReset(change CharacterChange, deadlines map[uint16]time.Time) error` | See the source definition and the subsystem ownership contract. |

### skills.go

[server/internal/db/skills.go](../../server/internal/db/skills.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`CharacterSkill`](../../server/internal/db/skills.go#L12)<br>`type CharacterSkill` | CharacterSkill is one learned skill. Delay is the remaining cooldown in seconds as last saved (CharSkills.Delay). |
| [`CharacterSkills`](../../server/internal/db/skills.go#L20)<br>`func CharacterSkills(charID uint32) ([]CharacterSkill, error)` | CharacterSkills runs usp_Read_Char_Skills_R. |
| [`AddCharacterSkill`](../../server/internal/db/skills.go#L46)<br>`func AddCharacterSkill(charID uint32, s CharacterSkill) error` | AddCharacterSkill runs usp_Save_Char_Skill_Add_E. |
| [`SaveCharacterSkill`](../../server/internal/db/skills.go#L54)<br>`func SaveCharacterSkill(charID uint32, s CharacterSkill) error` | SaveCharacterSkill runs usp_Save_Char_Skill_Mod_E (level and remaining cooldown). |
| [`QuickSlot`](../../server/internal/db/skills.go#L60)<br>`type QuickSlot` | QuickSlot is one skill-bar entry. |
| [`QuickSlots`](../../server/internal/db/skills.go#L66)<br>`func QuickSlots(charID uint32) ([]QuickSlot, error)` | QuickSlots runs usp_Read_Char_QuickSlots_R. |
| [`SaveQuickSlots`](../../server/internal/db/skills.go#L93)<br>`func SaveQuickSlots(charID uint32, slots []QuickSlot) error` | SaveQuickSlots atomically applies the complete bar using two set operations. Sending the same bar does not rewrite its unchanged slots. |
| [`AppliedBuff`](../../server/internal/db/skills.go#L131)<br>`type AppliedBuff` | AppliedBuff is a persisted active buff with its remaining seconds. |
| [`AppliedBuffs`](../../server/internal/db/skills.go#L138)<br>`func AppliedBuffs(charID uint32) ([]AppliedBuff, error)` | AppliedBuffs runs usp_Read_Char_ApplySkills_R. |
| [`SaveAppliedBuffs`](../../server/internal/db/skills.go#L167)<br>`func SaveAppliedBuffs(charID uint32, buffs []AppliedBuff) error` | SaveAppliedBuffs replaces the persisted buffs (usp_Save_Char_ApplySkill_Del_E with DeleteAll, then usp_Save_Char_ApplySkill_Add_E2 per buff). |
| [`saveBuffsTx`](../../server/internal/db/skills.go#L177)<br>`func saveBuffsTx(c context.Context, tx *sql.Tx, charID uint32, buffs []AppliedBuff) error` | See the source definition and the subsystem ownership contract. |
| [`CommitLearnSkill`](../../server/internal/db/skills.go#L194)<br>`func CommitLearnSkill(before, after Character, previous *CharacterSkill, learned CharacterSkill) error` | CommitLearnSkill guards both the previous skill level and available points; learning never grants a skill without committing its cost. |
| [`SaveSkillDelays`](../../server/internal/db/skills.go#L231)<br>`func SaveSkillDelays(charID uint32, skills []CharacterSkill) error` | See the source definition and the subsystem ownership contract. |
| [`saveSkillDelaysTx`](../../server/internal/db/skills.go#L244)<br>`func saveSkillDelaysTx(c context.Context, tx *sql.Tx, charID uint32, skills []CharacterSkill) error` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type CharacterSkill struct {
	SkillID uint16
	Level   uint8
	Number  uint8
	Delay   int16
}
```

```go
type QuickSlot struct {
	QuickBar, QuickSlot, Bag uint8
	Number                   uint16
}
```

```go
type AppliedBuff struct {
	SkillID     uint16
	Level       uint8
	LeftSeconds uint32
}
```

</details>

### social.go

[server/internal/db/social.go](../../server/internal/db/social.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`SocialContact`](../../server/internal/db/social.go#L5)<br>`type SocialContact` | See the source definition and the subsystem ownership contract. |
| [`AddFriends`](../../server/internal/db/social.go#L11)<br>`func AddFriends(a, b Character) error` | See the source definition and the subsystem ownership contract. |
| [`SocialContacts`](../../server/internal/db/social.go#L39)<br>`func SocialContacts(charID uint32, blocked bool) ([]SocialContact, error)` | See the source definition and the subsystem ownership contract. |
| [`SocialCharacter`](../../server/internal/db/social.go#L65)<br>`func SocialCharacter(name string) (Character, error)` | See the source definition and the subsystem ownership contract. |
| [`AddSocial`](../../server/internal/db/social.go#L76)<br>`func AddSocial(charID uint32, target Character, blocked bool) error` | See the source definition and the subsystem ownership contract. |
| [`RemoveSocial`](../../server/internal/db/social.go#L109)<br>`func RemoveSocial(charID, target uint32, blocked bool) error` | See the source definition and the subsystem ownership contract. |
| [`SocialMemo`](../../server/internal/db/social.go#L123)<br>`func SocialMemo(charID, target uint32, memo string, blocked bool) error` | See the source definition and the subsystem ownership contract. |
| [`SocialBlocked`](../../server/internal/db/social.go#L137)<br>`func SocialBlocked(charID, sender uint32) bool` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type SocialContact struct {
	ID         uint32
	Name, Memo string
	Class      uint8
}
```

</details>

### teleport_points.go

[server/internal/db/teleport_points.go](../../server/internal/db/teleport_points.go#L1)

Direct GX dependencies: `internal/proto`.

| Declaration | Responsibility / source comment |
|---|---|
| [`TeleportPoint`](../../server/internal/db/teleport_points.go#L5)<br>`type TeleportPoint` | See the source definition and the subsystem ownership contract. |
| [`TeleportPoints`](../../server/internal/db/teleport_points.go#L11)<br>`func TeleportPoints(charID uint32) (points [proto.TeleportPointCount]TeleportPoint, err error)` | See the source definition and the subsystem ownership contract. |
| [`SaveTeleportPoint`](../../server/internal/db/teleport_points.go#L37)<br>`func SaveTeleportPoint(charID uint32, slot uint8, p TeleportPoint) error` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type TeleportPoint struct {
	Saved   bool
	Map     uint16
	X, Y, Z float32
}
```

</details>

### teleport_runes.go

[server/internal/db/teleport_runes.go](../../server/internal/db/teleport_runes.go#L1)

Direct GX dependencies: `internal/teleportrunes`.

| Declaration | Responsibility / source comment |
|---|---|
| [`TeleportRuneDestination`](../../server/internal/db/teleport_runes.go#L5)<br>`type TeleportRuneDestination` | See the source definition and the subsystem ownership contract. |
| [`TeleportRunes`](../../server/internal/db/teleport_runes.go#L10)<br>`func TeleportRunes() (map[int16][2]TeleportRuneDestination, error)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type TeleportRuneDestination struct {
	MapID   uint16
	X, Y, Z float32
}
```

</details>

### transactions.go

[server/internal/db/transactions.go](../../server/internal/db/transactions.go#L1)

Direct GX dependencies: `server/internal/telemetry`.

| Declaration | Responsibility / source comment |
|---|---|
| [`writeGroupSize`](../../server/internal/db/transactions.go#L14)<br>`const writeGroupSize = 32` | See the source definition and the subsystem ownership contract. |
| [`writeQueueLimit`](../../server/internal/db/transactions.go#L15)<br>`const writeQueueLimit = 2048` | See the source definition and the subsystem ownership contract. |
| [`writeGatherTime`](../../server/internal/db/transactions.go#L16)<br>`const writeGatherTime = time.Millisecond` | See the source definition and the subsystem ownership contract. |
| [`groupedWrite`](../../server/internal/db/transactions.go#L19)<br>`type groupedWrite` | See the source definition and the subsystem ownership contract. |
| [`writeQueue`](../../server/internal/db/transactions.go#L25)<br>`type writeQueue` | See the source definition and the subsystem ownership contract. |
| [`writeQueues`](../../server/internal/db/transactions.go#L31)<br>`var writeQueues sync.Map` | keyed by the owning writer pool |
| [`groupedCommits`](../../server/internal/db/transactions.go#L32)<br>`var groupedCommits` | See the source definition and the subsystem ownership contract. |
| [`groupedOperations`](../../server/internal/db/transactions.go#L33)<br>`var groupedOperations` | See the source definition and the subsystem ownership contract. |
| [`groupWrite`](../../server/internal/db/transactions.go#L38)<br>`func groupWrite(c context.Context, pool *sql.DB, fn func(context.Context, *sql.Tx) error) error` | groupWrite shares a durable commit across independent operations. Each operation has its own savepoint and learns its result only after COMMIT. A caller whose request is already executing waits for the commit outcome. |
| [`run`](../../server/internal/db/transactions.go#L59)<br>`func (q *writeQueue) run(pool *sql.DB)` | See the source definition and the subsystem ownership contract. |
| [`commitWriteGroup`](../../server/internal/db/transactions.go#L80)<br>`func commitWriteGroup(pool *sql.DB, tasks []groupedWrite)` | See the source definition and the subsystem ownership contract. |
| [`transactionTime`](../../server/internal/db/transactions.go#L124)<br>`var transactionTime` | See the source definition and the subsystem ownership contract. |
| [`transactionFailures`](../../server/internal/db/transactions.go#L125)<br>`var transactionFailures` | See the source definition and the subsystem ownership contract. |
| [`withTx`](../../server/internal/db/transactions.go#L127)<br>`func withTx(c context.Context, pool *sql.DB, fn func(*sql.Tx) error) (err error)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type groupedWrite struct {
	context.Context
	run  func(context.Context, *sql.Tx) error
	done chan error
}
```

```go
type writeQueue struct {
	sync.Mutex
	pending []groupedWrite
	running bool
}
```

</details>

### warehouse.go

[server/internal/db/warehouse.go](../../server/internal/db/warehouse.go#L1)

Direct GX dependencies: `server/internal/config`.

| Declaration | Responsibility / source comment |
|---|---|
| [`CommitWarehouse`](../../server/internal/db/warehouse.go#L12)<br>`func CommitWarehouse(userUID uint32, change CharacterChange, before, after []Item) error` | CommitWarehouse guards the whole account inventory under the same write lock as character items and withdrawal fees, before any success is published. |
| [`writeWarehouseChange`](../../server/internal/db/warehouse.go#L36)<br>`func writeWarehouseChange(c context.Context, tx *sql.Tx, userUID uint32, before, after []Item) error` | See the source definition and the subsystem ownership contract. |
