# World package API

> Generated from Go source by `tools/generate_server_reference.py`; do not edit by hand.

[Package responsibilities](server-packages.md) · [Server architecture](server-architecture.md)

Every production Go file is listed, including OS-specific alternatives. Signatures include private helpers to make ownership and call boundaries inspectable. Declaration comments are source documentation, not an independent guarantee of behavior; the architecture guide explains transaction, concurrency and failure contracts.

## server/internal/world

Authoritative realm, sessions, maps, actors, game rules, protocol adapters and commit-before-publication application actions.

Tests: [absorption_skill_test.go](../../server/internal/world/absorption_skill_test.go), [absorption_test.go](../../server/internal/world/absorption_test.go), [attack_movement_test.go](../../server/internal/world/attack_movement_test.go), [attack_speed_test.go](../../server/internal/world/attack_speed_test.go), [auction_test.go](../../server/internal/world/auction_test.go), [blessing_lifecycle_test.go](../../server/internal/world/blessing_lifecycle_test.go), [bots_integration_test.go](../../server/internal/world/bots_integration_test.go), [bots_test.go](../../server/internal/world/bots_test.go), [change_sex_test.go](../../server/internal/world/change_sex_test.go), [character_enchant_test.go](../../server/internal/world/character_enchant_test.go), [chat_items_test.go](../../server/internal/world/chat_items_test.go), [chat_test.go](../../server/internal/world/chat_test.go), [cloak_fits_test.go](../../server/internal/world/cloak_fits_test.go), [combat_integration_test.go](../../server/internal/world/combat_integration_test.go), [combatformula_test.go](../../server/internal/world/combatformula_test.go), [cooldown_reduction_test.go](../../server/internal/world/cooldown_reduction_test.go), [degeneration_test.go](../../server/internal/world/degeneration_test.go), [detection_test.go](../../server/internal/world/detection_test.go), [drop_oj_limits_integration_test.go](../../server/internal/world/drop_oj_limits_integration_test.go), [drop_oj_limits_test.go](../../server/internal/world/drop_oj_limits_test.go), [drops_test.go](../../server/internal/world/drops_test.go), [durability_test.go](../../server/internal/world/durability_test.go), [elements_test.go](../../server/internal/world/elements_test.go), [emblem_test.go](../../server/internal/world/emblem_test.go), [enchant_ini_test.go](../../server/internal/world/enchant_ini_test.go), [equipment_rejection_test.go](../../server/internal/world/equipment_rejection_test.go), [experience_integration_test.go](../../server/internal/world/experience_integration_test.go), [experience_test.go](../../server/internal/world/experience_test.go), [fall_damage_test.go](../../server/internal/world/fall_damage_test.go), [firecracker_test.go](../../server/internal/world/firecracker_test.go), [firethorn_test.go](../../server/internal/world/firethorn_test.go), [forge_quote_test.go](../../server/internal/world/forge_quote_test.go), [forge_rules_test.go](../../server/internal/world/forge_rules_test.go), [free_mall_test.go](../../server/internal/world/free_mall_test.go), [gem_safety_test.go](../../server/internal/world/gem_safety_test.go), [gm_erase_test.go](../../server/internal/world/gm_erase_test.go), [gm_level_test.go](../../server/internal/world/gm_level_test.go), [gm_map_test.go](../../server/internal/world/gm_map_test.go), [gm_set_stats_test.go](../../server/internal/world/gm_set_stats_test.go), [gm_test.go](../../server/internal/world/gm_test.go), [grb_gm_integration_test.go](../../server/internal/world/grb_gm_integration_test.go), [grb_integration_test.go](../../server/internal/world/grb_integration_test.go), [guild_house_shop_integration_test.go](../../server/internal/world/guild_house_shop_integration_test.go), [guild_integration_test.go](../../server/internal/world/guild_integration_test.go), [guild_shop_session_test.go](../../server/internal/world/guild_shop_session_test.go), [guildhouse_test.go](../../server/internal/world/guildhouse_test.go), [healing_delta_test.go](../../server/internal/world/healing_delta_test.go), [healthassist_test.go](../../server/internal/world/healthassist_test.go), [inventory_operation_test.go](../../server/internal/world/inventory_operation_test.go), [isolation_test.go](../../server/internal/world/isolation_test.go), [item_shortcuts_test.go](../../server/internal/world/item_shortcuts_test.go), [itemboxes_test.go](../../server/internal/world/itemboxes_test.go), [jump_motion_regression_test.go](../../server/internal/world/jump_motion_regression_test.go), [killcount_test.go](../../server/internal/world/killcount_test.go), [ladders_test.go](../../server/internal/world/ladders_test.go), [lapis_debuff_test.go](../../server/internal/world/lapis_debuff_test.go), [lapis_mob_test.go](../../server/internal/world/lapis_mob_test.go), [loot_test.go](../../server/internal/world/loot_test.go), [mailbox_test.go](../../server/internal/world/mailbox_test.go), [map_rules_test.go](../../server/internal/world/map_rules_test.go), [mapitems_test.go](../../server/internal/world/mapitems_test.go), [miss_feedback_test.go](../../server/internal/world/miss_feedback_test.go), [mob_ai_leak_test.go](../../server/internal/world/mob_ai_leak_test.go), [mob_ai_test.go](../../server/internal/world/mob_ai_test.go), [mob_hop_test.go](../../server/internal/world/mob_hop_test.go), [mob_navigation_test.go](../../server/internal/world/mob_navigation_test.go), [mob_quest_drop_test.go](../../server/internal/world/mob_quest_drop_test.go), [mob_respawn_test.go](../../server/internal/world/mob_respawn_test.go), [mob_sacrifice_test.go](../../server/internal/world/mob_sacrifice_test.go), [myshop_integration_test.go](../../server/internal/world/myshop_integration_test.go), [myshop_test.go](../../server/internal/world/myshop_test.go), [notices_test.go](../../server/internal/world/notices_test.go), [npc_guard_test.go](../../server/internal/world/npc_guard_test.go), [npc_probe_test.go](../../server/internal/world/npc_probe_test.go), [npc_skill_repairs_test.go](../../server/internal/world/npc_skill_repairs_test.go), [oj_limits_test.go](../../server/internal/world/oj_limits_test.go), [oxygen_test.go](../../server/internal/world/oxygen_test.go), [party_global_test.go](../../server/internal/world/party_global_test.go), [passives_test.go](../../server/internal/world/passives_test.go), [periodic_healing_test.go](../../server/internal/world/periodic_healing_test.go), [persistence_audit_test.go](../../server/internal/world/persistence_audit_test.go), [phase_test.go](../../server/internal/world/phase_test.go), [player_feedback_test.go](../../server/internal/world/player_feedback_test.go), [player_forms_test.go](../../server/internal/world/player_forms_test.go), [player_movement_test.go](../../server/internal/world/player_movement_test.go), [player_speed_test.go](../../server/internal/world/player_speed_test.go), [player_state_fixes_test.go](../../server/internal/world/player_state_fixes_test.go), [portal_bounds_test.go](../../server/internal/world/portal_bounds_test.go), [potential_test.go](../../server/internal/world/potential_test.go), [previous_skill_test.go](../../server/internal/world/previous_skill_test.go), [protocol_layout_test.go](../../server/internal/world/protocol_layout_test.go), [pvp_test.go](../../server/internal/world/pvp_test.go), [pvpreward_test.go](../../server/internal/world/pvpreward_test.go), [quest_integration_test.go](../../server/internal/world/quest_integration_test.go), [quest_item_integration_test.go](../../server/internal/world/quest_item_integration_test.go), [quest_map_test.go](../../server/internal/world/quest_map_test.go), [quest_requirements_integration_test.go](../../server/internal/world/quest_requirements_integration_test.go), [raid_chat_test.go](../../server/internal/world/raid_chat_test.go), [raid_skill_test.go](../../server/internal/world/raid_skill_test.go), [rebirth_test.go](../../server/internal/world/rebirth_test.go), [recreation_runes_test.go](../../server/internal/world/recreation_runes_test.go), [reflect_test.go](../../server/internal/world/reflect_test.go), [relic_boot_integration_test.go](../../server/internal/world/relic_boot_integration_test.go), [relic_config_test.go](../../server/internal/world/relic_config_test.go), [relics_test.go](../../server/internal/world/relics_test.go), [revive_on_exit_test.go](../../server/internal/world/revive_on_exit_test.go), [rules_test.go](../../server/internal/world/rules_test.go), [runtime_test.go](../../server/internal/world/runtime_test.go), [sacrifice_test.go](../../server/internal/world/sacrifice_test.go), [safezones_test.go](../../server/internal/world/safezones_test.go), [shop_sell_test.go](../../server/internal/world/shop_sell_test.go), [skill_cadence_test.go](../../server/internal/world/skill_cadence_test.go), [skill_dash_test.go](../../server/internal/world/skill_dash_test.go), [skill_failure_test.go](../../server/internal/world/skill_failure_test.go), [skill_range_test.go](../../server/internal/world/skill_range_test.go), [soul_burn_test.go](../../server/internal/world/soul_burn_test.go), [spin_combo_test.go](../../server/internal/world/spin_combo_test.go), [stack_counts_test.go](../../server/internal/world/stack_counts_test.go), [status_items_test.go](../../server/internal/world/status_items_test.go), [summon_rune_test.go](../../server/internal/world/summon_rune_test.go), [target_effects_test.go](../../server/internal/world/target_effects_test.go), [taunt_test.go](../../server/internal/world/taunt_test.go), [teleport_rune_integration_test.go](../../server/internal/world/teleport_rune_integration_test.go), [teleport_stone_test.go](../../server/internal/world/teleport_stone_test.go), [trap_test.go](../../server/internal/world/trap_test.go), [utility_items_test.go](../../server/internal/world/utility_items_test.go), [vehicle_test.go](../../server/internal/world/vehicle_test.go), [visibility_test.go](../../server/internal/world/visibility_test.go), [worldcal_test.go](../../server/internal/world/worldcal_test.go), [worlddrops_test.go](../../server/internal/world/worlddrops_test.go).

### abilities.go

[server/internal/world/abilities.go](../../server/internal/world/abilities.go#L1)

Direct GX dependencies: `internal/skillchance`, `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`addSigned`](../../server/internal/world/abilities.go#L9)<br>`func addSigned(u uint32, v int32) uint32` | addSigned adds a signed change to an unsigned stat, flooring at 0. |
| [`applyAbility`](../../server/internal/world/abilities.go#L19)<br>`func applyAbility(e *EquipmentStats, effect, kind uint8, value int32) bool` | applyAbility is CUser::SetSkillAbility (0x00485e00) for the abilities that change combat stats. Debuffs (effect 3) do not raise maximum pools. It reports whether the ability type is one it models. |
| [`applyPassives`](../../server/internal/world/abilities.go#L86)<br>`func applyPassives(e *EquipmentStats, skills []db.CharacterSkill)` | applyPassives is CUser::InitPassive / AddSkillPassive (0x00489a80) over the learned skills in load order. Weapon masteries keep a per-type table and each one adds the table's running total for the equipped weapon, as the original does. |

### attack_movement.go

[server/internal/world/attack_movement.go](../../server/internal/world/attack_movement.go#L1)

Direct GX dependencies: `internal/attackspeed`, `internal/proto`.

| Declaration | Responsibility / source comment |
|---|---|
| [`attackMovementAllowed`](../../server/internal/world/attack_movement.go#L10)<br>`func (s *Session) attackMovementAllowed() bool` | Attacks and spell casts require footing. Only the attacker is checked. |
| [`commitAttackStop`](../../server/internal/world/attack_movement.go#L16)<br>`func (s *Session) commitAttackStop(now time.Time)` | See the source definition and the subsystem ownership contract. |

### auction.go

[server/internal/world/auction.go](../../server/internal/world/auction.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`auctionNPC`](../../server/internal/world/auction.go#L15)<br>`func (s *Session) auctionNPC(id uint32) bool` | See the source definition and the subsystem ownership contract. |
| [`auctionReply`](../../server/internal/world/auction.go#L28)<br>`func (s *Session) auctionReply(reply proto.AuctionReply)` | See the source definition and the subsystem ownership contract. |
| [`handleAuction`](../../server/internal/world/auction.go#L36)<br>`func (s *Session) handleAuction(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`postAuction`](../../server/internal/world/auction.go#L94)<br>`func (s *Session) postAuction(req proto.AuctionRequest) error` | See the source definition and the subsystem ownership contract. |
| [`payAuction`](../../server/internal/world/auction.go#L149)<br>`func (s *Session) payAuction(req proto.AuctionRequest) error` | See the source definition and the subsystem ownership contract. |
| [`collectAuction`](../../server/internal/world/auction.go#L169)<br>`func (s *Session) collectAuction(req proto.AuctionRequest) (string, error)` | See the source definition and the subsystem ownership contract. |

### bot_visibility.go

[server/internal/world/bot_visibility.go](../../server/internal/world/bot_visibility.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`botEntrancesPerTick`](../../server/internal/world/bot_visibility.go#L5)<br>`const botEntrancesPerTick = 32` | See the source definition and the subsystem ownership contract. |
| [`botObserver`](../../server/internal/world/bot_visibility.go#L7)<br>`type botObserver` | See the source definition and the subsystem ownership contract. |
| [`clearWorldBots`](../../server/internal/world/bot_visibility.go#L12)<br>`func clearWorldBots()` | See the source definition and the subsystem ownership contract. |
| [`tickWorldBots`](../../server/internal/world/bot_visibility.go#L32)<br>`func tickWorldBots(now time.Time)` | See the source definition and the subsystem ownership contract. |
| [`updateObservers`](../../server/internal/world/bot_visibility.go#L54)<br>`func (a *worldBotArena) updateObservers(players []*ConnectedPlayer)` | See the source definition and the subsystem ownership contract. |
| [`updateObserver`](../../server/internal/world/bot_visibility.go#L73)<br>`func (a *worldBotArena) updateObserver(observer *botObserver, viewer *ConnectedPlayer)` | See the source definition and the subsystem ownership contract. |
| [`add`](../../server/internal/world/bot_visibility.go#L96)<br>`func (observer *botObserver) add(frame []byte)` | See the source definition and the subsystem ownership contract. |
| [`emit`](../../server/internal/world/bot_visibility.go#L100)<br>`func (a *worldBotArena) emit(frame []byte, subject, target int)` | See the source definition and the subsystem ownership contract. |
| [`flush`](../../server/internal/world/bot_visibility.go#L108)<br>`func (a *worldBotArena) flush()` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type botObserver struct {
	visible []bool
	pending []byte
}
```

</details>

### bot_wire.go

[server/internal/world/bot_wire.go](../../server/internal/world/bot_wire.go#L1)

Direct GX dependencies: `server/internal/crypto`, `server/internal/netx`.

| Declaration | Responsibility / source comment |
|---|---|
| [`botWire`](../../server/internal/world/bot_wire.go#L16)<br>`type botWire` | Each synthetic controller owns a private post-login TCP stream. It uses normal world framing and AES, with credentials and account state absent. |
| [`openBotWire`](../../server/internal/world/bot_wire.go#L22)<br>`func openBotWire(listener *net.TCPListener) (*botWire, error)` | See the source definition and the subsystem ownership contract. |
| [`close`](../../server/internal/world/bot_wire.go#L38)<br>`func (w *botWire) close()` | See the source definition and the subsystem ownership contract. |
| [`request`](../../server/internal/world/bot_wire.go#L43)<br>`func (w *botWire) request(payload []byte) ([]byte, error)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type botWire struct {
	client, world net.Conn
	send, receive *crypto.AESCTR
	frames        netx.Frames
}
```

</details>

### bots.go

[server/internal/world/bots.go](../../server/internal/world/bots.go#L1)

Direct GX dependencies: `internal/mapgeometry`, `internal/proto`, `server/internal/config`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`botHealth`](../../server/internal/world/bots.go#L20)<br>`const botHealth = 12000` | See the source definition and the subsystem ownership contract. |
| [`botTick`](../../server/internal/world/bots.go#L21)<br>`const botTick = 100 * time.Millisecond` | See the source definition and the subsystem ownership contract. |
| [`botHomeRange`](../../server/internal/world/bots.go#L22)<br>`const botHomeRange = 5` | See the source definition and the subsystem ownership contract. |
| [`botMeleeSpacing`](../../server/internal/world/bots.go#L23)<br>`const botMeleeSpacing = 2.15` | See the source definition and the subsystem ownership contract. |
| [`worldBot`](../../server/internal/world/bots.go#L25)<br>`type worldBot` | See the source definition and the subsystem ownership contract. |
| [`worldBotArena`](../../server/internal/world/bots.go#L39)<br>`type worldBotArena` | See the source definition and the subsystem ownership contract. |
| [`worldBots`](../../server/internal/world/bots.go#L47)<br>`var worldBots struct { 	sync.Mutex 	nextID uint32 	arena  *worldBotArena }` | See the source definition and the subsystem ownership contract. |
| [`gmBots`](../../server/internal/world/bots.go#L53)<br>`func (s *Session) gmBots(args []string) error` | See the source definition and the subsystem ownership contract. |
| [`separatedBotHome`](../../server/internal/world/bots.go#L140)<br>`func separatedBotHome(g *mapgeometry.Geometry, x, y, z float32, homes []mapgeometry.Position) (mapgeometry.Position, bool)` | When terrain blocks a formation point, look for a nearby spot before opening any connections. Sending blocked bots back to the GM stacks them. |
| [`botEquipment`](../../server/internal/world/bots.go#L171)<br>`func botEquipment(defs *db.Defs) (bySlot [8][]*db.ItemDef)` | See the source definition and the subsystem ownership contract. |
| [`equip`](../../server/internal/world/bots.go#L187)<br>`func (b *worldBot) equip(defs *db.Defs, gear [8][]*db.ItemDef)` | See the source definition and the subsystem ownership contract. |
| [`advance`](../../server/internal/world/bots.go#L213)<br>`func (a *worldBotArena) advance(i int, now time.Time)` | See the source definition and the subsystem ownership contract. |
| [`applyRequest`](../../server/internal/world/bots.go#L303)<br>`func (a *worldBotArena) applyRequest(i int, payload []byte, target int, now time.Time)` | See the source definition and the subsystem ownership contract. |
| [`resolveAttack`](../../server/internal/world/bots.go#L356)<br>`func (a *worldBotArena) resolveAttack(i, target int, skill *db.SkillDef, now time.Time)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type worldBot struct {
	castAt             time.Time
	castTarget         int
	castSkill          *db.SkillDef
	wire               *botWire
	char               db.Character
	home               mapgeometry.Position
	faction            uint8
	items              []db.Item
	skills             []*db.SkillDef
	target             int
	nextAction, revive time.Time
}
```

```go
type worldBotArena struct {
	failed   bool
	owner    uint32
	geometry *mapgeometry.Geometry
	bots     []worldBot
	visible  map[*ConnectedPlayer]*botObserver
}
```

</details>

### character.go

[server/internal/world/character.go](../../server/internal/world/character.go#L1)

Direct GX dependencies: `internal/element`, `internal/itemtype`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`defaultCharacterID`](../../server/internal/world/character.go#L11)<br>`const defaultCharacterID = 1` | See the source definition and the subsystem ownership contract. |
| [`basePoolPerLevel`](../../server/internal/world/character.go#L12)<br>`const basePoolPerLevel = 100` | See the source definition and the subsystem ownership contract. |
| [`hpPerReaction`](../../server/internal/world/character.go#L13)<br>`const hpPerReaction = 5` | See the source definition and the subsystem ownership contract. |
| [`spPerDexterity`](../../server/internal/world/character.go#L14)<br>`const spPerDexterity = 5` | See the source definition and the subsystem ownership contract. |
| [`mpPerWisdom`](../../server/internal/world/character.go#L15)<br>`const mpPerWisdom = 5` | See the source definition and the subsystem ownership contract. |
| [`specialNone`](../../server/internal/world/character.go#L17)<br>`const specialNone = 0` | See the source definition and the subsystem ownership contract. |
| [`specialPercentHealingPotion`](../../server/internal/world/character.go#L18)<br>`const specialPercentHealingPotion = 17` | See the source definition and the subsystem ownership contract. |
| [`specialChangeAppearance`](../../server/internal/world/character.go#L19)<br>`const specialChangeAppearance = 44` | See the source definition and the subsystem ownership contract. |
| [`specialChangeSex`](../../server/internal/world/character.go#L20)<br>`const specialChangeSex = 45` | See the source definition and the subsystem ownership contract. |
| [`specialChangeName`](../../server/internal/world/character.go#L21)<br>`const specialChangeName = 51` | See the source definition and the subsystem ownership contract. |
| [`specialResetStats`](../../server/internal/world/character.go#L22)<br>`const specialResetStats = 34` | See the source definition and the subsystem ownership contract. |
| [`specialRecreationRune`](../../server/internal/world/character.go#L23)<br>`const specialRecreationRune = 62` | See the source definition and the subsystem ownership contract. |
| [`InventoryFirstBag`](../../server/internal/world/character.go#L25)<br>`const InventoryFirstBag = 1` | See the source definition and the subsystem ownership contract. |
| [`InventoryLastBag`](../../server/internal/world/character.go#L26)<br>`const InventoryLastBag = 5` | See the source definition and the subsystem ownership contract. |
| [`InventorySlotCount`](../../server/internal/world/character.go#L27)<br>`const InventorySlotCount = 24` | See the source definition and the subsystem ownership contract. |
| [`FactionLight`](../../server/internal/world/character.go#L29)<br>`const FactionLight = 0` | See the source definition and the subsystem ownership contract. |
| [`FactionDark`](../../server/internal/world/character.go#L30)<br>`const FactionDark = 1` | See the source definition and the subsystem ownership contract. |
| [`FactionNone`](../../server/internal/world/character.go#L31)<br>`const FactionNone = 2` | See the source definition and the subsystem ownership contract. |
| [`MaxCharacterSlots`](../../server/internal/world/character.go#L32)<br>`const MaxCharacterSlots = 5` | See the source definition and the subsystem ownership contract. |
| [`MotionSitDown`](../../server/internal/world/character.go#L34)<br>`const MotionSitDown = 1` | See the source definition and the subsystem ownership contract. |
| [`MotionStandUp`](../../server/internal/world/character.go#L35)<br>`const MotionStandUp = 0` | See the source definition and the subsystem ownership contract. |
| [`consumableItemTypes`](../../server/internal/world/character.go#L38)<br>`var consumableItemTypes` | See the source definition and the subsystem ownership contract. |
| [`isConsumableType`](../../server/internal/world/character.go#L40)<br>`func isConsumableType(t uint8) bool` | See the source definition and the subsystem ownership contract. |
| [`satAdd`](../../server/internal/world/character.go#L49)<br>`func satAdd(a, b uint32) uint32` | See the source definition and the subsystem ownership contract. |
| [`satSub`](../../server/internal/world/character.go#L56)<br>`func satSub(a, b uint32) uint32` | See the source definition and the subsystem ownership contract. |
| [`satMul`](../../server/internal/world/character.go#L63)<br>`func satMul(a, b uint32) uint32` | See the source definition and the subsystem ownership contract. |
| [`EquipmentStats`](../../server/internal/world/character.go#L72)<br>`type EquipmentStats` | EquipmentStats aggregates equipped items, gems and craft bonuses. |
| [`CategoryMods`](../../server/internal/world/character.go#L109)<br>`type CategoryMods` | CategoryMods are SetSkillAbility's per-category additions. |
| [`itemElement`](../../server/internal/world/character.go#L113)<br>`func itemElement(it *db.Item, d *db.ItemDef) uint8` | itemElement applies GX precedence: intrinsic element, otherwise the first elemental lapis in socket order. Later elemental lapis do not add an affinity. |
| [`activeLapis`](../../server/internal/world/character.go#L132)<br>`func activeLapis(it *db.Item, definition *db.ItemDef) [6]*db.ItemDef` | activeLapis keeps all non-elemental lapis but excludes every inactive elemental lapis from stats, recovery, speed and triggered-item effects. |
| [`gemItemType`](../../server/internal/world/character.go#L155)<br>`const gemItemType = 30` | See the source definition and the subsystem ownership contract. |
| [`specialRecoverGem`](../../server/internal/world/character.go#L158)<br>`const specialRecoverGem = 16` | specialRecoverGem is the gem Special that turns its pools into recovery. |
| [`IsWeaponItemType`](../../server/internal/world/character.go#L161)<br>`func IsWeaponItemType(t uint8) bool` | IsWeaponItemType: Type 1-15 and 45-65. |
| [`IsArmorItemType`](../../server/internal/world/character.go#L164)<br>`func IsArmorItemType(t uint8) bool` | IsArmorItemType: Type 16-21, 31-36 and 66-92. |
| [`CraftStats`](../../server/internal/world/character.go#L170)<br>`type CraftStats` | CraftStats is the decoded 20-digit Craftname: str dex rec int wis luc hp/100 mp/100 sp/100 enchant, two digits each. |
| [`ParseCraftName`](../../server/internal/world/character.go#L176)<br>`func ParseCraftName(name string) (CraftStats, bool)` | See the source definition and the subsystem ownership contract. |
| [`pos16`](../../server/internal/world/character.go#L193)<br>`func pos16(v int16) uint32` | See the source definition and the subsystem ownership contract. |
| [`isAccessoryType`](../../server/internal/world/character.go#L195)<br>`func isAccessoryType(t uint8) bool` | See the source definition and the subsystem ownership contract. |
| [`addItemDef`](../../server/internal/world/character.go#L198)<br>`func (e *EquipmentStats) addItemDef(itemType uint8, d *db.ItemDef)` | addItemDef adds full item effects; legacy stored quality never reduces stats. |
| [`realType`](../../server/internal/world/character.go#L220)<br>`func realType(t uint8) uint8` | realType is GetItemRType (0x004bdad0): item types grouped by family. |
| [`GetEquipmentStats`](../../server/internal/world/character.go#L223)<br>`func GetEquipmentStats(items []db.Item) EquipmentStats` | GetEquipmentStats sums every bag-0 item, its gems and craft bonuses. |
| [`CharacterStats`](../../server/internal/world/character.go#L287)<br>`type CharacterStats` | CharacterStats derives maximum pools from stats/equipment. Current pools are absolute durable values; bonuses must never be re-added on every read. |
| [`StatsFor`](../../server/internal/world/character.go#L323)<br>`func StatsFor(ch *db.Character, e EquipmentStats) CharacterStats` | See the source definition and the subsystem ownership contract. |
| [`recoverBaseSitting`](../../server/internal/world/character.go#L363)<br>`const recoverBaseSitting = 28` | InitEquipment's base recovery: 28 per pool sitting, 4 standing, 0 in combat. |
| [`recoverBaseStanding`](../../server/internal/world/character.go#L364)<br>`const recoverBaseStanding = 4` | InitEquipment's base recovery: 28 per pool sitting, 4 standing, 0 in combat. |
| [`recoveryAmount`](../../server/internal/world/character.go#L367)<br>`func recoveryAmount(value, maxValue uint32, special uint8) uint32` | See the source definition and the subsystem ownership contract. |
| [`charID`](../../server/internal/world/character.go#L374)<br>`func charID(ch *db.Character) uint32` | See the source definition and the subsystem ownership contract. |
| [`charName`](../../server/internal/world/character.go#L381)<br>`func charName(ch *db.Character) string` | See the source definition and the subsystem ownership contract. |
| [`charPos`](../../server/internal/world/character.go#L388)<br>`func charPos(ch *db.Character) (float32, float32, float32)` | See the source definition and the subsystem ownership contract. |
| [`buildCharacterDetails`](../../server/internal/world/character.go#L400)<br>`func buildCharacterDetails(ch *db.Character, s *CharacterStats) []byte` | buildCharacterDetails (0x0105), per the client's reader (0x0058dbd0): stat point, skill point (u16), max HP, MP, SP (u32), direction (u16), then eleven u32 and the 25-byte guild name. The three exp fields are the cumulative exp at the level's start, the total for the next level and the current total; their order is not yet confirmed in game. |
| [`buildAdditionalStats`](../../server/internal/world/character.go#L440)<br>`func buildAdditionalStats(s *CharacterStats) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildStatsReset`](../../server/internal/world/character.go#L449)<br>`func buildStatsReset(statPoint, str, rec, intel, wis, dex, luc uint16) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildCurrentHitpoints`](../../server/internal/world/character.go#L458)<br>`func buildCurrentHitpoints(s *CharacterStats) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildMaxHPMPSP`](../../server/internal/world/character.go#L466)<br>`func buildMaxHPMPSP(ch *db.Character, s *CharacterStats) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildMaxEnergy`](../../server/internal/world/character.go#L476)<br>`func buildMaxEnergy(ch *db.Character, hitpointType uint8, value uint32) []byte` | buildMaxEnergy (0x050B): hitpointType 0 hp, 1 sp, 2 mp. |
| [`buildCharacterEnteredMap`](../../server/internal/world/character.go#L485)<br>`func buildCharacterEnteredMap(ch *db.Character, isAdmin, firstLoad bool) []byte` | buildCharacterEnteredMap: marker 0 already present, 1 own first load, 2 admin. |
| [`buildCharacterEnteredMapGuild`](../../server/internal/world/character.go#L489)<br>`func buildCharacterEnteredMapGuild(ch *db.Character, isAdmin, firstLoad bool, guildID uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildCharacterMapPosition`](../../server/internal/world/character.go#L509)<br>`func buildCharacterMapPosition(ch *db.Character) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildCharacterMove`](../../server/internal/world/character.go#L524)<br>`func buildCharacterMove(ch *db.Character, motion uint8, direction uint16) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildCharacterMotion`](../../server/internal/world/character.go#L536)<br>`func buildCharacterMotion(characterID uint32, motion uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildLeaveWorldAck`](../../server/internal/world/character.go#L543)<br>`func buildLeaveWorldAck(characterID uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`equippedItem`](../../server/internal/world/character.go#L549)<br>`func equippedItem(items []db.Item, slot uint8) *db.Item` | See the source definition and the subsystem ownership contract. |
| [`enchantLevel`](../../server/internal/world/character.go#L558)<br>`func enchantLevel(it *db.Item) uint8` | See the source definition and the subsystem ownership contract. |
| [`buildCharacterEquipment`](../../server/internal/world/character.go#L565)<br>`func buildCharacterEquipment(ch *db.Character, items []db.Item, slot uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildChangeAppearance`](../../server/internal/world/character.go#L587)<br>`func buildChangeAppearance(characterID uint32, hair, face, height, sex uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildCharacterShape`](../../server/internal/world/character.go#L595)<br>`func buildCharacterShape(ch *db.Character, faction uint8, items []db.Item, motion uint8) []byte` | buildCharacterShape (0x0303). The persistent motion is 1 only for sitting. |
| [`buildCharacterShapeGuild`](../../server/internal/world/character.go#L599)<br>`func buildCharacterShapeGuild(ch *db.Character, faction uint8, items []db.Item, motion uint8, guildName string) []byte` | See the source definition and the subsystem ownership contract. |
| [`statPackets`](../../server/internal/world/character.go#L636)<br>`func statPackets(ch *db.Character, s *CharacterStats) []labeled` | See the source definition and the subsystem ownership contract. |
| [`buildEmptyCharacterListSlot`](../../server/internal/world/character.go#L647)<br>`func buildEmptyCharacterListSlot(slot uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildCharacterListSlot`](../../server/internal/world/character.go#L654)<br>`func buildCharacterListSlot(ch *db.Character, items []db.Item) []byte` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type EquipmentStats struct {
	Str, Rec, Int, Wis, Dex, Luc uint32
	MaxHP, MaxMP, MaxSP          uint32
	// CUser::GemEquipmentAdd/ItemEquipmentAdd: Effect1 is the item attack
	// and Effect2 its random variance (weapons and accessories), Effect3
	// physical and Effect4 magic defense (everything but weapons).
	ItemAttack, Variance uint32
	Defense, Resistance  uint32
	Absorption           uint32
	AttackAttr           uint8 // the weapon's (or its gem's) element
	DefenseAttr          uint8 // the armour's (slot 1) element, CUser::SetAttrDef
	RangedWeapon         bool
	// CUser::SetSkillAbility / AddSkillPassive / gems: per-category
	// (melee, ranged, magic) additions and other modifiers.
	Mods          [3]CategoryMods
	Critical      int32
	Range         int32 // added to the weapon's reach
	AttackSpeed   int32 // added to the swing-speed index
	MoveSpeed     int32
	SitRecoverPct [3]int32 // abilities 10-12: extra sitting recovery % (HP, SP, MP)
	// Flat recovery from Special 16 gems (CUser +0x1250 sitting, +0x1268
	// standing; HP, SP, MP).
	GemRecover [3]uint32
	// Country kill count tier (CUser::UpdateKCStatus) percents.
	KCPoolPct    [3]int32 // types 1-3: max HP, MP, SP
	KCRecoverPct [3]int32 // types 10-12: sitting/standing recovery, of max HP, SP, MP
	KCCombatPct  [3]int32 // types 13-15: in-combat recovery, of max HP, SP, MP
	KCGuardPct   [2]int32 // types 18-19: melee/ranged defense
	SPCostCut    int32    // ability 13 / 14: percent off skill costs
	MPCostCut    int32
	CooldownCut  int32    // ability 19: seconds off every spell's reuse while active
	Disabled     [3]int32 // abilities 32-34: no attacks of that category
	WeaponType   uint8    // real type of the equipped weapon (0 none)
	ShieldType   uint8    // real type of the equipped shield (0 none)
}
```

```go
type CategoryMods struct{ Accuracy, Attack, Defense, Evasion int32 }
```

```go
type CraftStats struct {
	Str, Dex, Rec, Int, Wis, Luc uint32
	MaxHP, MaxMP, MaxSP          uint32
	EnchantLevel                 uint8
}
```

```go
type CharacterStats struct {
	Str, Rec, Int, Wis, Dex, Luc    uint32
	MaxHP, MaxMP, MaxSP             uint32
	CurrentHP, CurrentMP, CurrentSP uint32
	// Inclusive display ranges for 0x0526: attack to attack + variance.
	MinAttack, MaxAttack            uint32
	MinMagic, MaxMagic              uint32
	Defense, Resistance, Absorption uint32
	// Combat values per category (melee, ranged, magic). Chances use
	// hundredths of one percentage point.
	Level        uint16
	Attack       [3]int32
	Accuracy     [3]int64
	Evasion      [3]int64
	Guard        [3]int32 // defense per category
	Critical     int64
	Variance     uint32
	ItemAttack   uint32 // the weapon's own attack (CUser +0x12d8)
	AttackAttr   uint8
	DefenseAttr  uint8
	RangedWeapon bool
	Range        int32
	AttackSpeed  int32
	MoveSpeed    int32
	Disabled     [3]int32
	SitRecover   [3]int32
	// Recover is CUser::UpdateRecover's flat amounts per mode (0 sitting,
	// 1 standing, 2 in combat) and pool (HP, SP, MP): InitEquipment's 28
	// sitting and 4 standing, plus gems and the country tier.
	Recover   [3][3]uint32
	SPCostCut int32
	MPCostCut int32
	// CooldownCut is ability 19: seconds taken off every spell's reuse.
	CooldownCut int32
}
```

</details>

### chat.go

[server/internal/world/chat.go](../../server/internal/world/chat.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`chatNormalRadius`](../../server/internal/world/chat.go#L17)<br>`const chatNormalRadius = 30` | Chat follows CUser::PacketChatE 0x00470490. Every channel takes a length byte (the text including its terminator): 0 or 1 is ignored, above 128 the connection is closed (0x004d43a0 with reason 9). PSendViewChat |
| [`chatShoutRadius`](../../server/internal/world/chat.go#L18)<br>`const chatShoutRadius = 100` | Chat follows CUser::PacketChatE 0x00470490. Every channel takes a length byte (the text including its terminator): 0 or 1 is ignored, above 128 the connection is closed (0x004d43a0 with reason 9). PSendViewRange |
| [`chatShoutDelay`](../../server/internal/world/chat.go#L19)<br>`const chatShoutDelay = 30 * time.Second` | Chat follows CUser::PacketChatE 0x00470490. Every channel takes a length byte (the text including its terminator): 0 or 1 is ignored, above 128 the connection is closed (0x004d43a0 with reason 9). |
| [`chatMaxLength`](../../server/internal/world/chat.go#L20)<br>`const chatMaxLength = 0x80` | Chat follows CUser::PacketChatE 0x00470490. Every channel takes a length byte (the text including its terminator): 0 or 1 is ignored, above 128 the connection is closed (0x004d43a0 with reason 9). |
| [`chatLength`](../../server/internal/world/chat.go#L25)<br>`func (s *Session) chatLength(p []byte, at int) bool` | chatLength validates the length byte at p[at]: ok false drops the message, and closes the connection when it is too long. |
| [`handleChat`](../../server/internal/world/chat.go#L40)<br>`func (s *Session) handleChat(op packet.Opcode, p []byte)` | See the source definition and the subsystem ownership contract. |
| [`handlePublicChat`](../../server/internal/world/chat.go#L68)<br>`func (s *Session) handlePublicChat(op packet.Opcode, message string, items proto.ChatItems)` | handlePublicChat delivers the normal, trade, shout and area channels. |
| [`handleWhisper`](../../server/internal/world/chat.go#L111)<br>`func (s *Session) handleWhisper(p []byte)` | handleWhisper (0x1102): target name[21] (20 characters), length, text. Another character (by name, same country) gets direction 0 with the sender's name; the sender gets direction 1 with the target's name; otherwise 0x1106 result 1. |
| [`buildWhisper`](../../server/internal/world/chat.go#L133)<br>`func buildWhisper(direction uint8, name, message string) []byte` | See the source definition and the subsystem ownership contract. |
| [`chatItemSnapshots`](../../server/internal/world/chat.go#L142)<br>`func (s *Session) chatItemSnapshots(p []byte, textOffset int) (items proto.ChatItems, ok bool)` | Build display data from server-owned inventory, never from chat text. |
| [`chatWithItems`](../../server/internal/world/chat.go#L161)<br>`func chatWithItems(frame []byte, items proto.ChatItems) []byte` | See the source definition and the subsystem ownership contract. |

### cloak_fits.go

[server/internal/world/cloak_fits.go](../../server/internal/world/cloak_fits.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`chestCloakFits`](../../server/internal/world/cloak_fits.go#L10)<br>`func chestCloakFits(defs *db.Defs) []proto.ChestCloakFit` | See the source definition and the subsystem ownership contract. |
| [`sendChestCloakFits`](../../server/internal/world/cloak_fits.go#L28)<br>`func (s *Session) sendChestCloakFits()` | See the source definition and the subsystem ownership contract. |

### combat.go

[server/internal/world/combat.go](../../server/internal/world/combat.go#L1)

Direct GX dependencies: `internal/attackspeed`, `internal/proto`, `internal/skilltiming`, `server/internal/config`, `server/internal/db`, `server/internal/netx`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`weaponSlot`](../../server/internal/world/combat.go#L16)<br>`const weaponSlot = 5` | See the source definition and the subsystem ownership contract. |
| [`autoAttackRangeGrace`](../../server/internal/world/combat.go#L17)<br>`const autoAttackRangeGrace = 2.0` | See the source definition and the subsystem ownership contract. |
| [`autoAttackCancelRange`](../../server/internal/world/combat.go#L18)<br>`const autoAttackCancelRange = 80.0` | See the source definition and the subsystem ownership contract. |
| [`attackSuccessNormal`](../../server/internal/world/combat.go#L19)<br>`const attackSuccessNormal = 0` | See the source definition and the subsystem ownership contract. |
| [`attackResultCritical`](../../server/internal/world/combat.go#L20)<br>`const attackResultCritical = 1` | See the source definition and the subsystem ownership contract. |
| [`attackResultMiss`](../../server/internal/world/combat.go#L21)<br>`const attackResultMiss = 3` | See the source definition and the subsystem ownership contract. |
| [`deathTypeMob`](../../server/internal/world/combat.go#L22)<br>`const deathTypeMob = 0x02` | See the source definition and the subsystem ownership contract. |
| [`weaponCombat`](../../server/internal/world/combat.go#L25)<br>`type weaponCombat` | See the source definition and the subsystem ownership contract. |
| [`weaponAttackDelay`](../../server/internal/world/combat.go#L33)<br>`func weaponAttackDelay(tier, weaponBase uint8) (time.Duration, bool)` | weaponAttackDelay takes the native zero-based attack index, not index+1. |
| [`equippedWeaponCombat`](../../server/internal/world/combat.go#L37)<br>`func equippedWeaponCombat(items []db.Item) (weaponCombat, bool)` | See the source definition and the subsystem ownership contract. |
| [`sendNear`](../../server/internal/world/combat.go#L61)<br>`func sendNear(mapID uint16, x, z, radius float32, frame []byte, label string)` | See the source definition and the subsystem ownership contract. |
| [`autoAttackSwing`](../../server/internal/world/combat.go#L68)<br>`func (s *Session) autoAttackSwing(target uint32) bool` | autoAttackSwing resolves one swing; false cancels the auto-attack. |
| [`resolveMobDamage`](../../server/internal/world/combat.go#L175)<br>`func (s *Session) resolveMobDamage(r *MobDamageResult) bool` | resolveMobDamage handles the aftermath of any damage to a mob: death, drops and exp, or an HP sync. Returns whether the mob is still alive. |
| [`sendMobDeath`](../../server/internal/world/combat.go#L191)<br>`func sendMobDeath(mapID uint16, m *MapMob, killerID uint32, label string)` | See the source definition and the subsystem ownership contract. |
| [`applyIncomingMobDamage`](../../server/internal/world/combat.go#L201)<br>`func (s *Session) applyIncomingMobDamage(d pendingDamage) bool` | applyIncomingMobDamage applies one queued mob attack; true when it killed. Slot attacks with a skill (Attacktype = a level-100 Skills row) go through CMob::AttackUser(skill); others are CMob::Attack. |
| [`applyMobSkill`](../../server/internal/world/combat.go#L279)<br>`func (s *Session) applyMobSkill(d pendingDamage, md *db.MobDef, mc *mobCombat, sk *db.SkillDef, stats *CharacterStats) bool` | applyMobSkill sends the cast with its animation slot (0x060B or 0x060C) and any subsequent results without a slot (0x060D). |
| [`buildMobSkillPacket`](../../server/internal/world/combat.go#L406)<br>`func buildMobSkillPacket(op packet.Opcode, result uint8, mobID, target uint32, sk *db.SkillDef, damage [3]uint16, healed *[3]uint16, attackSlot ...uint8) []byte` | buildMobSkillPacket: result, mob id, target id, a byte the client reads before the skill id (0x060B/0x060C only), skill id, level, HP/SP/MP damage and, for 0x060C, the healed HP/SP/MP. |
| [`takeMobDamage`](../../server/internal/world/combat.go#L432)<br>`func (s *Session) takeMobDamage(d pendingDamage, damage uint32, stats *CharacterStats) bool` | takeMobDamage lowers the character's HP; true when it died. |
| [`afterMobDamage`](../../server/internal/world/combat.go#L460)<br>`func (s *Session) afterMobDamage(d pendingDamage, killed bool)` | See the source definition and the subsystem ownership contract. |
| [`persist`](../../server/internal/world/combat.go#L484)<br>`func (s *Session) persist()` | persist saves the active character (usp_Save_Char_Info_E) and refreshes the shared snapshot. |
| [`applyExpGain`](../../server/internal/world/combat.go#L506)<br>`func applyExpGain(ch *db.Character, gained uint32) (levels uint16, credited uint32)` | applyExpGain returns the levels gained and experience actually credited. |
| [`experienceCredit`](../../server/internal/world/combat.go#L525)<br>`type experienceCredit` | See the source definition and the subsystem ownership contract. |
| [`applyExperienceAward`](../../server/internal/world/combat.go#L530)<br>`func applyExperienceAward(ch *db.Character, amount uint32, perfect bool) experienceCredit` | CUser::AddExpFromUser (0x458240) grants an equal extra amount to a normal seven-member party when the base award alone is below the next threshold. Process any level crossed by the bonus immediately using the GX curve. |
| [`expBounds`](../../server/internal/world/combat.go#L544)<br>`func expBounds(ch *db.Character) (prev, next uint32)` | expBounds returns the cumulative thresholds used by the character's bar. |
| [`rewardKill`](../../server/internal/world/combat.go#L557)<br>`func rewardKill(killer *ConnectedPlayer, killerID uint32, mapID uint16, m *MapMob, log logPrinter)` | rewardKill hands out a kill's loot and exp: drops on the ground, gold split across the party (or dropped for a solo killer) and each share of exp queued to its recipient's session. An offline DoT caster gets the solo exp straight in the database. |
| [`experienceBonus`](../../server/internal/world/combat.go#L616)<br>`func (s *Session) experienceBonus() uint32` | Active skill IDs are unique within each skill catalog. Each EXP ability contributes percentage points to the total applied to monster rewards. |
| [`notifyExpGain`](../../server/internal/world/combat.go#L628)<br>`func (s *Session) notifyExpGain(credited experienceCredit, levels uint16)` | See the source definition and the subsystem ownership contract. |
| [`buildExpGain`](../../server/internal/world/combat.go#L649)<br>`func buildExpGain(credited experienceCredit) []byte` | buildExpGain (0x0207): base experience and the extra Perfect Party EXP. |
| [`buildLevelUp`](../../server/internal/world/combat.go#L658)<br>`func buildLevelUp(op packet.Opcode, ch *db.Character) []byte` | buildLevelUp (0x0508 self / 0x051E others): character id, level, stat points, skill points, exp at the level's start and for the next level. |
| [`grantOfflineKillExp`](../../server/internal/world/combat.go#L671)<br>`func grantOfflineKillExp(characterID uint32, mobID, mobLevel uint16, mobExp uint32, log logPrinter)` | grantOfflineKillExp credits a DoT kill to a caster who went offline. |
| [`logPrinter`](../../server/internal/world/combat.go#L688)<br>`type logPrinter` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type weaponCombat struct {
	ItemType, TypeID uint8
	AttackTime       uint8
	Range            float32
	Delay            time.Duration
}
```

```go
type experienceCredit struct{ base, perfect uint32 }
```

```go
type logPrinter interface {
	Printf(format string, args ...any)
}
```

</details>

### combat_cadence.go

[server/internal/world/combat_cadence.go](../../server/internal/world/combat_cadence.go#L1)

Direct GX dependencies: `internal/skilltiming`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`actionInterval`](../../server/internal/world/combat_cadence.go#L11)<br>`func (s *Session) actionInterval() (time.Duration, bool)` | actionInterval is the shared interval at the live attack index and weapon. |
| [`ownClip`](../../server/internal/world/combat_cadence.go#L18)<br>`func (s *Session) ownClip() time.Duration` | ownClip is the animation floor of the last skill; any later action stamp (basic attack, cast, item) makes it stale on its own. |
| [`actionRecovery`](../../server/internal/world/combat_cadence.go#L27)<br>`func (s *Session) actionRecovery() (time.Duration, bool)` | actionRecovery is how long after the last action the next one may start: the shared interval, or the last skill's own clip when that is longer. |
| [`combatActionReady`](../../server/internal/world/combat_cadence.go#L34)<br>`func (s *Session) combatActionReady(now time.Time) bool` | Both basic attack routes and skills recover from lastActionAt. Use the live equipment/buff tier so haste, slow and gear changes affect recovery immediately. |
| [`skillActionReady`](../../server/internal/world/combat_cadence.go#L42)<br>`func (s *Session) skillActionReady(def *db.SkillDef, now time.Time) bool` | See the source definition and the subsystem ownership contract. |
| [`pendingSpinCombo`](../../server/internal/world/combat_cadence.go#L53)<br>`type pendingSpinCombo` | See the source definition and the subsystem ownership contract. |
| [`queueSpinCombo`](../../server/internal/world/combat_cadence.go#L59)<br>`func (s *Session) queueSpinCombo(def *db.SkillDef, op packet.Opcode, payload []byte) bool` | See the source definition and the subsystem ownership contract. |
| [`processSpinCombo`](../../server/internal/world/combat_cadence.go#L67)<br>`func (s *Session) processSpinCombo()` | See the source definition and the subsystem ownership contract. |
| [`pendingSkillUse`](../../server/internal/world/combat_cadence.go#L82)<br>`type pendingSkillUse` | pendingSkillUse is the classic one-slot selection: a request that reaches the server before the shared execution deadline is kept (latest wins) and runs the moment the interval allows, instead of being refused by a few milliseconds of network jitter. Own reuse, costs and validation are checked again at release. |
| [`holdSkillUse`](../../server/internal/world/combat_cadence.go#L88)<br>`func (s *Session) holdSkillUse(def *db.SkillDef, op packet.Opcode, payload []byte, now time.Time) bool` | See the source definition and the subsystem ownership contract. |
| [`processPendingSkill`](../../server/internal/world/combat_cadence.go#L100)<br>`func (s *Session) processPendingSkill()` | See the source definition and the subsystem ownership contract. |
| [`minWake`](../../server/internal/world/combat_cadence.go#L123)<br>`const minWake = 5 * time.Millisecond` | minWake keeps a blocked pending action from spinning the loop. |
| [`nextWake`](../../server/internal/world/combat_cadence.go#L128)<br>`func (s *Session) nextWake() time.Duration` | nextWake is how long the session loop may sleep: the usual tick, shortened to the moment a held skill, a spin combo or the next auto-attack swing becomes due, so those run at their exact deadline instead of the next 100 ms tick. |

<details><summary>Type definitions and owned fields</summary>

```go
type pendingSpinCombo struct {
	payload       []byte
	op            packet.Opcode
	due, previous time.Time
}
```

```go
type pendingSkillUse struct {
	payload []byte
	op      packet.Opcode
	expires time.Time
}
```

</details>

### combatformula.go

[server/internal/world/combatformula.go](../../server/internal/world/combatformula.go#L1)

Direct GX dependencies: `internal/element`, `server/internal/config`, `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`attackMelee`](../../server/internal/world/combatformula.go#L13)<br>`const attackMelee = iota` | See the source definition and the subsystem ownership contract. |
| [`attackRanged`](../../server/internal/world/combatformula.go#L14)<br>`const attackRanged` | See the source definition and the subsystem ownership contract. |
| [`attackMagic`](../../server/internal/world/combatformula.go#L15)<br>`const attackMagic` | See the source definition and the subsystem ownership contract. |
| [`maxDamage`](../../server/internal/world/combatformula.go#L19)<br>`const maxDamage = 30000` | See the source definition and the subsystem ownership contract. |
| [`fullChance`](../../server/internal/world/combatformula.go#L20)<br>`const fullChance = 10000` | Hundredths of one percentage point. |
| [`combatDEXFirst`](../../server/internal/world/combatformula.go#L24)<br>`var combatDEXFirst uint32` | See the source definition and the subsystem ownership contract. |
| [`combatDEXSecond`](../../server/internal/world/combatformula.go#L25)<br>`var combatDEXSecond uint32` | See the source definition and the subsystem ownership contract. |
| [`dexChance`](../../server/internal/world/combatformula.go#L28)<br>`func dexChance(dex, first, second uint32) int64` | See the source definition and the subsystem ownership contract. |
| [`setCombatChances`](../../server/internal/world/combatformula.go#L34)<br>`func setCombatChances(accuracy, evasion *[3]int64, dex, wis uint32, mods [3]CategoryMods)` | See the source definition and the subsystem ownership contract. |
| [`boundedCombatStat`](../../server/internal/world/combatformula.go#L42)<br>`func boundedCombatStat(value int64) int32` | See the source definition and the subsystem ownership contract. |
| [`setAttack`](../../server/internal/world/combatformula.go#L48)<br>`func setAttack(s *CharacterStats, e EquipmentStats)` | setAttack uses the same attribute scaling for melee and ranged. Category abilities still apply to their own attack, while melee defense is shared. |
| [`mobCombat`](../../server/internal/world/combatformula.go#L87)<br>`type mobCombat` | mobCombat contains the combat values derived from a monster's catalog and active effects. Its attack templates supply the weapon part separately. |
| [`mobCombatFor`](../../server/internal/world/combatformula.go#L99)<br>`func mobCombatFor(d *db.MobDef) mobCombat` | See the source definition and the subsystem ownership contract. |
| [`rollCombatChance`](../../server/internal/world/combatformula.go#L109)<br>`func rollCombatChance(chance int64) bool` | See the source definition and the subsystem ownership contract. |
| [`effectiveEvasion`](../../server/internal/world/combatformula.go#L113)<br>`func effectiveEvasion(defenderEvasion, attackerAccuracy int64) int64` | See the source definition and the subsystem ownership contract. |
| [`mobLevelPenalty`](../../server/internal/world/combatformula.go#L118)<br>`func mobLevelPenalty(playerLevel uint16, mobLevel int32) int64` | mobLevelPenalty is the percentage adjustment for a monster above a player. |
| [`playerHitChance`](../../server/internal/world/combatformula.go#L122)<br>`func playerHitChance(s *CharacterStats, m *mobCombat, kind int) int64` | See the source definition and the subsystem ownership contract. |
| [`playerDodgeChance`](../../server/internal/world/combatformula.go#L127)<br>`func playerDodgeChance(s *CharacterStats, m *mobCombat, kind int) int64` | See the source definition and the subsystem ownership contract. |
| [`playerHits`](../../server/internal/world/combatformula.go#L132)<br>`func playerHits(s *CharacterStats, m *mobCombat, kind int) bool` | See the source definition and the subsystem ownership contract. |
| [`mobHits`](../../server/internal/world/combatformula.go#L136)<br>`func mobHits(m *mobCombat, s *CharacterStats, kind int) bool` | See the source definition and the subsystem ownership contract. |
| [`userHits`](../../server/internal/world/combatformula.go#L140)<br>`func userHits(att, def *CharacterStats, kind int) bool` | See the source definition and the subsystem ownership contract. |
| [`skillSuccess`](../../server/internal/world/combatformula.go#L146)<br>`func skillSuccess(def *db.SkillDef) bool` | skillSuccess retains a definition's independent success roll. Accuracy and evasion still apply, even to skills with a guaranteed catalog roll. |
| [`playerSkillHits`](../../server/internal/world/combatformula.go#L150)<br>`func playerSkillHits(s *CharacterStats, m *mobCombat, md *db.MobDef, def *db.SkillDef) bool` | See the source definition and the subsystem ownership contract. |
| [`mobSkillHits`](../../server/internal/world/combatformula.go#L162)<br>`func mobSkillHits(m *mobCombat, def *db.SkillDef, s *CharacterStats) bool` | See the source definition and the subsystem ownership contract. |
| [`userSkillHits`](../../server/internal/world/combatformula.go#L166)<br>`func userSkillHits(att, def *CharacterStats, sk *db.SkillDef) bool` | See the source definition and the subsystem ownership contract. |
| [`hitDamage`](../../server/internal/world/combatformula.go#L170)<br>`type hitDamage` | See the source definition and the subsystem ownership contract. |
| [`criticalChance`](../../server/internal/world/combatformula.go#L183)<br>`func criticalChance(critical, targetLuc int64) int64` | criticalChance is the attacker's critical chance (0.2 points per LUC plus bonuses) less 0.01 percentage points per point of the target's LUC. |
| [`weaponHit`](../../server/internal/world/combatformula.go#L189)<br>`func (h hitDamage) weaponHit(bonus, multiplier int64) (int32, bool)` | weaponHit computes the same damage for all three categories. Using the element table's exact tenths keeps truncation at the end of the calculation. |
| [`directAmount`](../../server/internal/world/combatformula.go#L209)<br>`func directAmount(amount, levelPercent, limit int64) int32` | directAmount scales nonzero immediate skill damage and preserves zero components. It is also used for the level adjustment on extra impacts. |
| [`extraSkillDamage`](../../server/internal/world/combatformula.go#L216)<br>`func extraSkillDamage(def *db.SkillDef, levelPercent int64) [3]uint32` | See the source definition and the subsystem ownership contract. |
| [`damagePools`](../../server/internal/world/combatformula.go#L225)<br>`type damagePools` | See the source definition and the subsystem ownership contract. |
| [`currentDamagePools`](../../server/internal/world/combatformula.go#L227)<br>`func currentDamagePools(s *CharacterStats) damagePools` | See the source definition and the subsystem ownership contract. |
| [`maximumDamagePools`](../../server/internal/world/combatformula.go#L231)<br>`func maximumDamagePools(s *CharacterStats) damagePools` | See the source definition and the subsystem ownership contract. |
| [`skillHit`](../../server/internal/world/combatformula.go#L235)<br>`func (h hitDamage) skillHit(def *db.SkillDef, caster, current, maximum damagePools) ([3]int32, bool)` | See the source definition and the subsystem ownership contract. |
| [`characterHitDamage`](../../server/internal/world/combatformula.go#L281)<br>`func characterHitDamage(s *CharacterStats, kind int, guard int32, defenseAttr uint8, levelPercent, targetLuc int64) hitDamage` | See the source definition and the subsystem ownership contract. |
| [`playerDamage`](../../server/internal/world/combatformula.go#L288)<br>`func playerDamage(s *CharacterStats, m *mobCombat, kind int) (int32, bool)` | See the source definition and the subsystem ownership contract. |
| [`playerSkillDamage`](../../server/internal/world/combatformula.go#L293)<br>`func playerSkillDamage(s *CharacterStats, m *mobCombat, def *db.SkillDef, current, maximum damagePools) ([3]int32, bool)` | See the source definition and the subsystem ownership contract. |
| [`userDamage`](../../server/internal/world/combatformula.go#L299)<br>`func userDamage(att, def *CharacterStats, kind int) (int32, bool)` | See the source definition and the subsystem ownership contract. |
| [`userSkillDamage`](../../server/internal/world/combatformula.go#L304)<br>`func userSkillDamage(att, def *CharacterStats, sk *db.SkillDef, caster damagePools) ([3]int32, bool)` | See the source definition and the subsystem ownership contract. |
| [`mobHitDamage`](../../server/internal/world/combatformula.go#L310)<br>`func mobHitDamage(m *mobCombat, a *db.MobAttack, s *CharacterStats, kind int, attr uint8) hitDamage` | See the source definition and the subsystem ownership contract. |
| [`mobDamage`](../../server/internal/world/combatformula.go#L321)<br>`func mobDamage(m *mobCombat, a *db.MobAttack, s *CharacterStats) (int32, bool)` | See the source definition and the subsystem ownership contract. |
| [`mobSkillDamage`](../../server/internal/world/combatformula.go#L325)<br>`func mobSkillDamage(m *mobCombat, md *db.MobDef, def *db.SkillDef, s *CharacterStats, mobHP uint32) ([3]int32, bool)` | See the source definition and the subsystem ownership contract. |
| [`f32`](../../server/internal/world/combatformula.go#L332)<br>`func f32(v int32) float64` | f32 preserves the original periodic-damage calculation's input precision. |

<details><summary>Type definitions and owned fields</summary>

```go
type mobCombat struct {
	DamageFactor float32
	Level        int32
	Accuracy     [3]int64
	Evasion      [3]int64
	Guard        [3]int32
	Attack       [3]int32
	Critical     int64
	Luc          int64 // its LUC lowers the critical chance of whoever hits it
	Attrib       uint8
}
```

```go
type hitDamage struct {
	attack       int64
	variance     uint32
	guard        int32
	critical     int64
	targetLuc    int64 // each point of the target's LUC removes 0.01 percentage points
	attackAttr   uint8
	defenseAttr  uint8
	levelPercent int64
}
```

```go
type damagePools [3]uint32
```

</details>

### cooldown_reduction.go

[server/internal/world/cooldown_reduction.go](../../server/internal/world/cooldown_reduction.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`abilityCooldownCut`](../../server/internal/world/cooldown_reduction.go#L12)<br>`const abilityCooldownCut = 19` | See the source definition and the subsystem ownership contract. |
| [`cooldownCut`](../../server/internal/world/cooldown_reduction.go#L26)<br>`func (s *Session) cooldownCut() uint16` | cooldownCut is the seconds currently taken off spell reuse. |
| [`sendSkillCooldown`](../../server/internal/world/cooldown_reduction.go#L33)<br>`func (s *Session) sendSkillCooldown(skillID, remaining, total uint16)` | sendSkillCooldown tells the client a skill's timer: remaining seconds and the full-bar total (0 keeps the one it has). |
| [`effectCooldownCut`](../../server/internal/world/cooldown_reduction.go#L42)<br>`func effectCooldownCut(b *ActiveBuff) int32` | effectCooldownCut is the reduction a newly applied effect carries. |
| [`reduceCooldowns`](../../server/internal/world/cooldown_reduction.go#L53)<br>`func (s *Session) reduceCooldowns(now time.Time, seconds int32)` | reduceCooldowns takes seconds off every running spell reuse right now. |
| [`learnedDef`](../../server/internal/world/cooldown_reduction.go#L82)<br>`func (s *Session) learnedDef(id uint16) *db.SkillDef` | learnedDef is the definition of a learned skill at its learned level. |

### crafting.go

[server/internal/world/crafting.go](../../server/internal/world/crafting.go#L1)

Direct GX dependencies: `internal/forgecost`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`lapisiaItemType`](../../server/internal/world/crafting.go#L17)<br>`const lapisiaItemType = 95` | See the source definition and the subsystem ownership contract. |
| [`maxGemSlots`](../../server/internal/world/crafting.go#L18)<br>`const maxGemSlots = 6` | See the source definition and the subsystem ownership contract. |
| [`composeResultSuccess`](../../server/internal/world/crafting.go#L19)<br>`const composeResultSuccess = 0` | See the source definition and the subsystem ownership contract. |
| [`composeResultFailure`](../../server/internal/world/crafting.go#L20)<br>`const composeResultFailure = 1` | See the source definition and the subsystem ownership contract. |
| [`craftString`](../../server/internal/world/crafting.go#L23)<br>`func craftString(pairs [10]uint32) string` | See the source definition and the subsystem ownership contract. |
| [`craftPairs`](../../server/internal/world/crafting.go#L31)<br>`func craftPairs(c CraftStats) [10]uint32` | See the source definition and the subsystem ownership contract. |
| [`buildEnchantResult`](../../server/internal/world/crafting.go#L37)<br>`func buildEnchantResult(result, lBag, lSlot uint8, lCount uint16, mBag, mSlot uint8, mCount uint16, iBag, iSlot uint8, money uint32, craft string) []byte` | See the source definition and the subsystem ownership contract. |
| [`lapisiaMoney`](../../server/internal/world/crafting.go#L50)<br>`func lapisiaMoney(reqIg uint8) uint32` | lapisiaMoney is GetGemMoneyByGrade (0x0045f6e0) by the lapisia's ReqIg. |
| [`lapisiaFits`](../../server/internal/world/crafting.go#L55)<br>`func lapisiaFits(lap *db.ItemDef, rt uint8) bool` | lapisiaFits checks the target families selected by the lapisia fields (ReqLevel weapons, Country type 16, the six job flags types 17-22, Grow 23, ReqStr 25). |
| [`handleEnchantAdd`](../../server/internal/world/crafting.go#L77)<br>`func (s *Session) handleEnchantAdd(p []byte)` | handleEnchantAdd (0x0805) is CUser::ItemLapisianAdd (0x0045fb20): lapisia bag/slot, material bag/slot, material count, item bag/slot. The item (count 1, Server set) needs its ReqLuc as material; below the configured cap, ReqIg selects the gold cost and the configured rate determines success independently of material quantity. Either way the lapisia and the material count are used up; equipment is preserved. |
| [`planEnchantAdd`](../../server/internal/world/crafting.go#L78)<br>`func (s *Session) planEnchantAdd(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`adjustCount`](../../server/internal/world/crafting.go#L172)<br>`func (s *Session) adjustCount(uid int64, count uint16)` | adjustCount sets an item's count by UID (deleting at zero) and persists. |
| [`buildGemAddResult`](../../server/internal/world/crafting.go#L188)<br>`func buildGemAddResult(result, gBag, gSlot uint8, gCount uint16, iBag, iSlot, pos, gemID uint8, money uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`maxPools`](../../server/internal/world/crafting.go#L199)<br>`func (s *Session) maxPools() [3]uint32` | maxPools returns the effective max HP/SP/MP (0x050B order 0, 1, 2). |
| [`sendChangedPools`](../../server/internal/world/crafting.go#L205)<br>`func (s *Session) sendChangedPools(before [3]uint32, why string)` | sendChangedPools refreshes derived stats and sends 0x050B only for pools that changed. |
| [`gemFits`](../../server/internal/world/crafting.go#L220)<br>`func gemFits(item *db.Item, target, gemDef *db.ItemDef) bool` | gemFits is CItem::EnableGem (0x004ba050): the item's Slot is a capacity each linked gem uses its own Slot of; the same gem twice is refused. |
| [`findSpecialItem`](../../server/internal/world/crafting.go#L239)<br>`func (s *Session) findSpecialItem(special uint8) int` | findSpecialItem finds a usable item of real type 26 with the given Special in the bags (CUser::FindItemGemAddItemProtect looks for '9', the Lucky Charm). |
| [`specialGemSuccessUp`](../../server/internal/world/crafting.go#L253)<br>`const specialGemSuccessUp = 36` | additive percentage-point bonus |
| [`specialGemExtract`](../../server/internal/world/crafting.go#L254)<br>`const specialGemExtract = 58` | direct extraction success percentage |
| [`specialGemProtect`](../../server/internal/world/crafting.go#L255)<br>`const specialGemProtect = 57` | legacy Lucky Charm ID; no longer needed/consumed |
| [`handleGemAdd`](../../server/internal/world/crafting.go#L265)<br>`func (s *Session) handleGemAdd(p []byte)` | handleGemAdd (0x0801) is CUser::ItemGemAdd (0x00461310): gem bag/slot, item bag/slot, optional hammer bag/slot. The item (count 1) must take the gem (the lapisia-style target fields and EnableGem's capacity); the gem's GetGemMoneyByGrade is paid; base chance is 289 - 8 x ReqIg plus blessing (at least 1), plus hammer ReqVg percentage points, up to 100%. The gem is always used up. GX preserves the equipment on every failure; legacy ReqVg destruction flags and Lucky Charms do not affect linking. |
| [`planGemAdd`](../../server/internal/world/crafting.go#L266)<br>`func (s *Session) planGemAdd(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`handleGemRemove`](../../server/internal/world/crafting.go#L365)<br>`func (s *Session) handleGemRemove(p []byte)` | handleGemRemove extracts one selected socket with an extraction hammer. Obsolete requests without a hammer are rejected without consuming anything. |
| [`planGemRemove`](../../server/internal/world/crafting.go#L366)<br>`func (s *Session) planGemRemove(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`jewelryItemTypes`](../../server/internal/world/crafting.go#L475)<br>`var jewelryItemTypes` | See the source definition and the subsystem ownership contract. |
| [`rollsStatsOnly`](../../server/internal/world/crafting.go#L479)<br>`func rollsStatsOnly(t uint8) bool` | rollsStatsOnly is CItem::ReGenerationCraftExpansion's split: weapons (real types 1-15) and real types 22, 23 and 25 roll the six stats only. |
| [`rollCraftName`](../../server/internal/world/crafting.go#L486)<br>`func rollCraftName(count, maxValue uint8, statsOnly bool, previous string) string` | rollCraftName picks `count` distinct stats (6 for weapons/jewelry, 9 otherwise) at 1..maxValue each, keeping the enchant level. |
| [`buildComposeResult`](../../server/internal/world/crafting.go#L501)<br>`func buildComposeResult(result, bag, slot uint8, craft string) []byte` | See the source definition and the subsystem ownership contract. |
| [`handleItemCompose`](../../server/internal/world/crafting.go#L510)<br>`func (s *Session) handleItemCompose(p []byte)` | handleItemCompose: rune bag/slot, item bag/slot. Items.ReqWis is the max roll and Items.Server the number of rolled stats. |
| [`planItemCompose`](../../server/internal/world/crafting.go#L511)<br>`func (s *Session) planItemCompose(p []byte)` | See the source definition and the subsystem ownership contract. |

### degeneration.go

[server/internal/world/degeneration.go](../../server/internal/world/degeneration.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`detailDegeneration`](../../server/internal/world/degeneration.go#L9)<br>`const detailDegeneration = 90` | debuff: a creature that cannot attack or cast and crawls |
| [`degenerationSlow`](../../server/internal/world/degeneration.go#L11)<br>`const degenerationSlow = 4` | degenerationSlow takes the move tier to its lowest whatever buffs there are. |
| [`degenerated`](../../server/internal/world/degeneration.go#L16)<br>`func (s *Session) degenerated() bool` | degenerated is Degeneration (TypeEffect 3, TypeDetail 90) on this user: the ChangeLevel of the skill names the monster it looks like (visual only). |
| [`cannotAct`](../../server/internal/world/degeneration.go#L27)<br>`func (s *Session) cannotAct() bool` | cannotAct: stunned, asleep or transformed users cannot attack or use spells. They can still use consumables and change equipment, which do not ask this. |
| [`transformedMob`](../../server/internal/world/degeneration.go#L29)<br>`func (s *Session) transformedMob() uint16` | See the source definition and the subsystem ownership contract. |
| [`buildTransform`](../../server/internal/world/degeneration.go#L31)<br>`func buildTransform(user uint32, mob uint16) []byte` | See the source definition and the subsystem ownership contract. |
| [`tickTransform`](../../server/internal/world/degeneration.go#L39)<br>`func (s *Session) tickTransform()` | tickTransform publishes a change and keeps the view snapshot for late observers. |

### detection.go

[server/internal/world/detection.go](../../server/internal/world/detection.go#L1)

Direct GX dependencies: `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`detailDetection`](../../server/internal/world/detection.go#L10)<br>`const detailDetection = 93` | buff: reveals the enemies hiding around the caster |
| [`detectionTick`](../../server/internal/world/detection.go#L11)<br>`const detectionTick = time.Second` | how often the area is swept |
| [`RevealStealth`](../../server/internal/world/detection.go#L15)<br>`func (p *ConnectedPlayer) RevealStealth()` | RevealStealth asks a user's session to end its stealth. |
| [`takeReveal`](../../server/internal/world/detection.go#L21)<br>`func (p *ConnectedPlayer) takeReveal() bool` | See the source definition and the subsystem ownership contract. |
| [`tickReveal`](../../server/internal/world/detection.go#L31)<br>`func (s *Session) tickReveal()` | tickReveal ends the stealth of a user an enemy's Detection has found (RemApplySkillHide): it is not hidden again until it casts Stealth again. |
| [`tickDetection`](../../server/internal/world/detection.go#L43)<br>`func (s *Session) tickDetection(now time.Time)` | tickDetection is CUser::ChkFindHide, run by ChkKeepSkill for a Detection buff (TypeDetail 93): on every sweep each living enemy user within ApplyRange of the caster loses its stealth. The area is the caster's current position, so it follows the caster; allies, Game Masters who are concealed and users that are not hidden are untouched. The first sweep runs when the buff is applied. |

### dispatch.go

[server/internal/world/dispatch.go](../../server/internal/world/dispatch.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/netx`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`dispatch`](../../server/internal/world/dispatch.go#L12)<br>`func (s *Session) dispatch(op packet.Opcode, p []byte)` | dispatch is the protocol adapter. Application methods own validation and state changes; guarded storage operations own persistence and durability. |
| [`phaseAllows`](../../server/internal/world/dispatch.go#L307)<br>`func (s *Session) phaseAllows(op packet.Opcode) bool` | Phase admission is enforced before reaching application handlers. Character management cannot replace an actor that is already registered in the world. |
| [`economicOpcode`](../../server/internal/world/dispatch.go#L324)<br>`func economicOpcode(op packet.Opcode) bool` | See the source definition and the subsystem ownership contract. |

### drops.go

[server/internal/world/drops.go](../../server/internal/world/drops.go#L1)

Direct GX dependencies: `internal/catalogfiles`, `internal/mapgeometry`, `internal/worlddrops`, `server/internal/db`, `server/internal/logx`.

| Declaration | Responsibility / source comment |
|---|---|
| [`chanceSucceeds`](../../server/internal/world/drops.go#L24)<br>`func chanceSucceeds(percent uint32) bool` | See the source definition and the subsystem ownership contract. |
| [`dropLimitIntervals`](../../server/internal/world/drops.go#L35)<br>`var dropLimitIntervals` | dropLimitIntervals are the original Drop selectors (ms per window). |
| [`gradeEntry`](../../server/internal/world/drops.go#L37)<br>`type gradeEntry` | See the source definition and the subsystem ownership contract. |
| [`dropCatalog`](../../server/internal/world/drops.go#L46)<br>`var dropCatalog struct { 	mu         sync.Mutex 	grades     map[uint16][]*gradeEntry 	worldRules []worlddrops.Rule }` | See the source definition and the subsystem ownership contract. |
| [`isQuestItemType`](../../server/internal/world/drops.go#L52)<br>`func isQuestItemType(t uint8) bool` | See the source definition and the subsystem ownership contract. |
| [`buildDropCatalog`](../../server/internal/world/drops.go#L56)<br>`func buildDropCatalog()` | buildDropCatalog groups items by grade, unrestricted items first, then those with a Server or Drop restriction, as CGameData's init does. |
| [`loadWorldDrops`](../../server/internal/world/drops.go#L91)<br>`func loadWorldDrops(log *logx.Logger) error` | See the source definition and the subsystem ownership contract. |
| [`pickGradeItem`](../../server/internal/world/drops.go#L105)<br>`func pickGradeItem(grade uint16) *db.ItemDef` | pickGradeItem is MobDropCatalog::select: a random item of the grade, dropped only if its time limit allows it. |
| [`rollWorldDrops`](../../server/internal/world/drops.go#L127)<br>`func rollWorldDrops(level uint16, roll func(uint64) uint64, drop func(worlddrops.Rule))` | See the source definition and the subsystem ownership contract. |
| [`dropPosition`](../../server/internal/world/drops.go#L140)<br>`func dropPosition(x, z float32) (float32, float32)` | dropPosition scatters a drop within half a unit along each horizontal axis. |
| [`mobDropPosition`](../../server/internal/world/drops.go#L146)<br>`func mobDropPosition(mob *MapMob) (x, y, z float32)` | mobDropPosition keeps CMob::GetDropPosRand's scatter around the monster and uses its authoritative height to select the floor. |
| [`floorMobDrop`](../../server/internal/world/drops.go#L162)<br>`func floorMobDrop(g *mapgeometry.Geometry, origin, candidate mapgeometry.Position) mapgeometry.Position` | See the source definition and the subsystem ownership contract. |
| [`rollMobDrops`](../../server/internal/world/drops.go#L183)<br>`func rollMobDrops(mapID uint16, mob *MapMob, ownerID, killerID uint32, party []*ConnectedPlayer, log logPrinter) []MapItem` | rollMobDrops rolls one kill's loot, registers it on the map and returns the new map items.  With two or more party members in range, the gold is split evenly among them instead of dropped (at least 1 each). |
| [`newDroppedItem`](../../server/internal/world/drops.go#L251)<br>`func newDroppedItem(def *db.ItemDef) db.Item` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type gradeEntry struct {
	def *db.ItemDef
	// Drop limit: at most one drop per window, the window start
	// randomized within its last tenth.
	limited          bool
	delay, randDelay time.Duration
	nextAt           time.Time
}
```

</details>

### duel.go

[server/internal/world/duel.go](../../server/internal/world/duel.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`duelNone`](../../server/internal/world/duel.go#L48)<br>`const duelNone = 0` | See the source definition and the subsystem ownership contract. |
| [`duelAsking`](../../server/internal/world/duel.go#L49)<br>`const duelAsking = 1` | See the source definition and the subsystem ownership contract. |
| [`duelAsked`](../../server/internal/world/duel.go#L50)<br>`const duelAsked = 2` | See the source definition and the subsystem ownership contract. |
| [`duelCountdown`](../../server/internal/world/duel.go#L51)<br>`const duelCountdown = 3` | See the source definition and the subsystem ownership contract. |
| [`duelFighting`](../../server/internal/world/duel.go#L52)<br>`const duelFighting = 4` | See the source definition and the subsystem ownership contract. |
| [`duelWindow`](../../server/internal/world/duel.go#L53)<br>`const duelWindow = 5` | See the source definition and the subsystem ownership contract. |
| [`duelRequestTime`](../../server/internal/world/duel.go#L55)<br>`const duelRequestTime = 10 * time.Second` | See the source definition and the subsystem ownership contract. |
| [`duelCountdownTime`](../../server/internal/world/duel.go#L56)<br>`const duelCountdownTime = 5 * time.Second` | See the source definition and the subsystem ownership contract. |
| [`duelRange`](../../server/internal/world/duel.go#L57)<br>`const duelRange = 40` | See the source definition and the subsystem ownership contract. |
| [`duelWagerSlots`](../../server/internal/world/duel.go#L58)<br>`const duelWagerSlots = 8` | See the source definition and the subsystem ownership contract. |
| [`duelCountCap`](../../server/internal/world/duel.go#L59)<br>`const duelCountCap = 0xFFFF` | See the source definition and the subsystem ownership contract. |
| [`duelPick`](../../server/internal/world/duel.go#L62)<br>`type duelPick` | See the source definition and the subsystem ownership contract. |
| [`duelState`](../../server/internal/world/duel.go#L69)<br>`type duelState` | duelState is a user's CExchangePvP and PvP fields (guarded by duelMu). |
| [`duelMu`](../../server/internal/world/duel.go#L87)<br>`var duelMu sync.Mutex` | See the source definition and the subsystem ownership contract. |
| [`duelPacket`](../../server/internal/world/duel.go#L89)<br>`func duelPacket(op packet.Opcode, write func(*packet.Writer)) []byte` | See the source definition and the subsystem ownership contract. |
| [`duelResult`](../../server/internal/world/duel.go#L97)<br>`func duelResult(op packet.Opcode, result uint8, id uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`duelByte`](../../server/internal/world/duel.go#L101)<br>`func duelByte(op packet.Opcode, v uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`pendingDuelEnd`](../../server/internal/world/duel.go#L106)<br>`type pendingDuelEnd` | pendingDuelEnd asks a session to finish PvPModeEnd's own part. |
| [`queueDuelEnd`](../../server/internal/world/duel.go#L108)<br>`func (p *ConnectedPlayer) queueDuelEnd(e pendingDuelEnd)` | See the source definition and the subsystem ownership contract. |
| [`takeDuelEnds`](../../server/internal/world/duel.go#L119)<br>`func (p *ConnectedPlayer) takeDuelEnds() []pendingDuelEnd` | See the source definition and the subsystem ownership contract. |
| [`duelStateOf`](../../server/internal/world/duel.go#L128)<br>`func duelStateOf(p *ConnectedPlayer) (int, uint32)` | duelStateOf reads a user's duel state and opponent. |
| [`duelFoes`](../../server/internal/world/duel.go#L135)<br>`func duelFoes(a, b *ConnectedPlayer) bool` | duelFoes reports two users fighting each other (IsTeam's duel case). |
| [`inActiveDuel`](../../server/internal/world/duel.go#L142)<br>`func inActiveDuel(p *ConnectedPlayer) bool` | inActiveDuel is IsTeamUsable's state-4 check. |
| [`modeEndLocked`](../../server/internal/world/duel.go#L150)<br>`func (p *ConnectedPlayer) modeEndLocked()` | modeEnd is PvPModeEnd under duelMu (the debuffs and the ledger are cleared in the user's own session). |
| [`dropWagerLocked`](../../server/internal/world/duel.go#L158)<br>`func dropWagerLocked(src, owner *ConnectedPlayer)` | dropWagerLocked is ItemDropPvP: src's held wager lies at owner's feet, owned by owner alone. |
| [`endDuelLocked`](../../server/internal/world/duel.go#L182)<br>`func endDuelLocked(a, b *ConnectedPlayer, toA, toB []byte, label string)` | endDuelLocked ends a duel for both with each wager back to its owner; msg builds what each side receives (nil sends nothing). |
| [`duelLeave`](../../server/internal/world/duel.go#L200)<br>`func duelLeave(p *ConnectedPlayer, fightCode uint8, toSelf bool)` | duelLeave is the shared shape of PvPChkLeaveByAttack / ByDeath / ByZoneLeave for p; fightCode is the 0x2406 result, toSelf whether p also gets it. |
| [`duelHitBy`](../../server/internal/world/duel.go#L238)<br>`func duelHitBy(p *ConnectedPlayer, attacker uint32)` | duelHitBy is PvPChkLeaveByAttack when attacker (0 for a mob) hits p. |
| [`handleDuel`](../../server/internal/world/duel.go#L250)<br>`func (s *Session) handleDuel(op packet.Opcode, p []byte)` | handleDuel dispatches 0x2401-0x240C. |
| [`duelRequest`](../../server/internal/world/duel.go#L291)<br>`func (s *Session) duelRequest(target uint32)` | See the source definition and the subsystem ownership contract. |
| [`duelAnswer`](../../server/internal/world/duel.go#L320)<br>`func (s *Session) duelAnswer(accept bool)` | See the source definition and the subsystem ownership contract. |
| [`duelGiveUp`](../../server/internal/world/duel.go#L351)<br>`func (s *Session) duelGiveUp()` | See the source definition and the subsystem ownership contract. |
| [`windowPeerLocked`](../../server/internal/world/duel.go#L363)<br>`func (s *Session) windowPeerLocked() *ConnectedPlayer` | windowPeerLocked returns the opponent while both are in the window. |
| [`cancelReadyLocked`](../../server/internal/world/duel.go#L376)<br>`func cancelReadyLocked(a, b *ConnectedPlayer)` | cancelReadyLocked is ExchangeCancleReadyPvP. |
| [`duelReady`](../../server/internal/world/duel.go#L386)<br>`func (s *Session) duelReady(action uint8)` | See the source definition and the subsystem ownership contract. |
| [`duelAddItem`](../../server/internal/world/duel.go#L423)<br>`func (s *Session) duelAddItem(bag, slot uint8, count uint16, wslot uint8)` | See the source definition and the subsystem ownership contract. |
| [`duelRemoveItem`](../../server/internal/world/duel.go#L454)<br>`func (s *Session) duelRemoveItem(wslot uint8)` | See the source definition and the subsystem ownership contract. |
| [`duelMoney`](../../server/internal/world/duel.go#L468)<br>`func (s *Session) duelMoney(m uint32)` | See the source definition and the subsystem ownership contract. |
| [`takeWager`](../../server/internal/world/duel.go#L486)<br>`func takeWager(charID uint32, d *duelState, items *[]db.Item, money uint32) ([][]byte, uint32)` | takeWager is ExchangePvP over one inventory: the gold (up to what the owner has) and each picked item move into the held wager; a partial stack becomes a new item without quality or gems. It returns the 0x0206 packets. |
| [`exchangeWagerSelf`](../../server/internal/world/duel.go#L524)<br>`func (s *Session) exchangeWagerSelf()` | exchangeWagerSelf runs ExchangePvP on this session's own inventory. |
| [`exchangeWagerOther`](../../server/internal/world/duel.go#L541)<br>`func exchangeWagerOther(o *ConnectedPlayer, log logPrinter)` | exchangeWagerOther runs ExchangePvP for the opponent through the database; its session reloads with the trade-refresh mailbox. |
| [`tickDuel`](../../server/internal/world/duel.go#L553)<br>`func (s *Session) tickDuel()` | tickDuel is PvPChkLeaveByTime and the session's part of PvPModeEnd. |
| [`duelCheckMove`](../../server/internal/world/duel.go#L607)<br>`func (s *Session) duelCheckMove()` | duelCheckMove is PvPChkLeaveByMove after this user moved. |
| [`duelDeath`](../../server/internal/world/duel.go#L635)<br>`func (s *Session) duelDeath(winner *ConnectedPlayer) bool` | duelDeath is PvPGetDeathType 4 + PvPChkLoseDeath for this (dead) user killed by winner; false when the death is not the duel's. |
| [`queueDuelWin`](../../server/internal/world/duel.go#L664)<br>`func (p *ConnectedPlayer) queueDuelWin()` | See the source definition and the subsystem ownership contract. |
| [`takeDuelWins`](../../server/internal/world/duel.go#L670)<br>`func (p *ConnectedPlayer) takeDuelWins() int` | See the source definition and the subsystem ownership contract. |
| [`applyDuelWins`](../../server/internal/world/duel.go#L679)<br>`func (s *Session) applyDuelWins()` | applyDuelWins books victories in the winner's session. |
| [`guildDuelState`](../../server/internal/world/duel.go#L694)<br>`func (p *ConnectedPlayer) guildDuelState() uint32` | guildDuelState is CUser +0x54e4 (non-zero in a guild duel), under duelMu. |

<details><summary>Type definitions and owned fields</summary>

```go
type duelPick struct {
	Bag, Slot uint8
	Count     uint16
	Item      db.Item
}
```

```go
type duelState struct {
	state     int
	opponent  uint32
	deadline  time.Time
	center    [2]float32 // kept after the duel, as the original
	picks     [duelWagerSlots]*duelPick
	money     uint32
	ready     bool
	held      []db.Item
	heldMoney uint32
	// Guild duel (CUser +0x54e4 opposing guild, +0x54e8 center, +0x54f8
	// enrolled, +0x54fc active).
	gOpp    uint32
	gCenter [2]float32
	gJoined bool
	gActive bool
}
```

```go
type pendingDuelEnd struct{ ledger, debuffs, stopAttack bool }
```

</details>

### elementstate.go

[server/internal/world/elementstate.go](../../server/internal/world/elementstate.go#L1)

Direct GX dependencies: `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`elementSource`](../../server/internal/world/elementstate.go#L8)<br>`type elementSource` | Native elemental fields are event state, not an unordered sum of buffs. RemApplySkillBuff 0x485130 and RemApplySkillDebuff 0x485720 restore equipment even if a different elemental effect remains in the applied-skill list. |
| [`elementRuntime`](../../server/internal/world/elementstate.go#L14)<br>`type elementRuntime` | See the source definition and the subsystem ownership contract. |
| [`elementSource`](../../server/internal/world/elementstate.go#L22)<br>`func (s *Session) elementSource(slot uint8) elementSource` | See the source definition and the subsystem ownership contract. |
| [`applyElement`](../../server/internal/world/elementstate.go#L28)<br>`func (s *Session) applyElement(b ActiveBuff)` | See the source definition and the subsystem ownership contract. |
| [`removeElement`](../../server/internal/world/elementstate.go#L43)<br>`func (s *Session) removeElement(b ActiveBuff)` | See the source definition and the subsystem ownership contract. |
| [`syncElements`](../../server/internal/world/elementstate.go#L71)<br>`func (s *Session) syncElements(base EquipmentStats)` | See the source definition and the subsystem ownership contract. |
| [`replaceAttackElement`](../../server/internal/world/elementstate.go#L110)<br>`func (s *Session) replaceAttackElement(def *db.SkillDef)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type elementSource struct {
	UID          int64
	Type, TypeID uint8
	Gems         [6]uint8
	Present      bool
}
```

```go
type elementRuntime struct {
	ready           bool
	weapon, armor   elementSource
	attack, defense uint8
	attackFlags     [9]bool
	defenseSkill    bool
}
```

</details>

### emblem.go

[server/internal/world/emblem.go](../../server/internal/world/emblem.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`emblemFee`](../../server/internal/world/emblem.go#L13)<br>`const emblemFee = 450000` | Cloak emblems (the 0x08xx group handled by CUser::PacketGem, 0x0046bd40). 0x0803 is CUser::ItemCloakAdd (0x00464ad0) and 0x0804 CUser::ItemCloakRemove (0x00464d40). The pattern is stored in the cloak's six gem bytes: front, back and emblem shape (index + 1, 0 = none) followed by the palette index of each. The NPC is not validated by the original handlers. |
| [`buildEmblemAddResult`](../../server/internal/world/emblem.go#L15)<br>`func buildEmblemAddResult(result, bag, slot uint8, data [6]uint8, gold uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`emblemTarget`](../../server/internal/world/emblem.go#L25)<br>`func (s *Session) emblemTarget(bag, slot uint8) int` | emblemTarget returns the index of a carried cloak. The original accepts only a populated slot whose item definition type is 24. |
| [`handleCloakAdd`](../../server/internal/world/emblem.go#L36)<br>`func (s *Session) handleCloakAdd(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`planCloakAdd`](../../server/internal/world/emblem.go#L37)<br>`func (s *Session) planCloakAdd(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`handleCloakRemove`](../../server/internal/world/emblem.go#L63)<br>`func (s *Session) handleCloakRemove(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`planCloakRemove`](../../server/internal/world/emblem.go#L64)<br>`func (s *Session) planCloakRemove(p []byte)` | See the source definition and the subsystem ownership contract. |

### enchant_rules.go

[server/internal/world/enchant_rules.go](../../server/internal/world/enchant_rules.go#L1)

Direct GX dependencies: `internal/catalogfiles`, `internal/enchantrules`, `internal/proto`, `server/internal/logx`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`defaultEnchantMaxLevel`](../../server/internal/world/enchant_rules.go#L15)<br>`const defaultEnchantMaxLevel = 20` | See the source definition and the subsystem ownership contract. |
| [`enchantMaxLevel`](../../server/internal/world/enchant_rules.go#L16)<br>`const enchantMaxLevel = enchantrules.MaxLevel` | armor encodes 50+level in the two-digit Craftname pair |
| [`enchantRulesConfig`](../../server/internal/world/enchant_rules.go#L22)<br>`type enchantRulesConfig` | Runtime enchant configuration is loaded atomically from ItemEnchant.ini. These initial values retain the historical defaults before startup loading. Step N is the probability of N -> N+1, in successes per million. |
| [`enchantConfig`](../../server/internal/world/enchant_rules.go#L27)<br>`var enchantConfig` | See the source definition and the subsystem ownership contract. |
| [`weaponEnchantAttackBonus`](../../server/internal/world/enchant_rules.go#L34)<br>`var weaponEnchantAttackBonus` | Bonuses are indexed by the resulting enchantment level, starting at zero. |
| [`armorEnchantAbsorption`](../../server/internal/world/enchant_rules.go#L35)<br>`var armorEnchantAbsorption` | See the source definition and the subsystem ownership contract. |
| [`validateEnchantRules`](../../server/internal/world/enchant_rules.go#L37)<br>`func validateEnchantRules() error` | See the source definition and the subsystem ownership contract. |
| [`applyEnchantRules`](../../server/internal/world/enchant_rules.go#L49)<br>`func applyEnchantRules(rows enchantrules.Catalog) error` | See the source definition and the subsystem ownership contract. |
| [`loadEnchantRules`](../../server/internal/world/enchant_rules.go#L66)<br>`func loadEnchantRules(log *logx.Logger) error` | See the source definition and the subsystem ownership contract. |
| [`sendEnchantRules`](../../server/internal/world/enchant_rules.go#L80)<br>`func (s *Session) sendEnchantRules()` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type enchantRulesConfig struct {
	WeaponCap, ArmorCap     uint8
	WeaponRate, DefenseRate [enchantMaxLevel + 1]uint32
}
```

</details>

### entities.go

[server/internal/world/entities.go](../../server/internal/world/entities.go#L1)

Direct GX dependencies: `internal/mapnav`, `internal/npcquest`, `internal/proto`, `server/ai`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`VisibilityCellSize`](../../server/internal/world/entities.go#L20)<br>`const VisibilityCellSize = 40` | Visibility and combat radii, matching the real server's cell view model. |
| [`ViewportRadius`](../../server/internal/world/entities.go#L21)<br>`const ViewportRadius = 120.0` | Visibility and combat radii, matching the real server's cell view model. |
| [`CombatRadius`](../../server/internal/world/entities.go#L22)<br>`const CombatRadius = BroadcastRadius` | Visibility and combat radii, matching the real server's cell view model. CZone::PSendView |
| [`MobAttackRange`](../../server/internal/world/entities.go#L23)<br>`const MobAttackRange = 3.0` | Visibility and combat radii, matching the real server's cell view model. |
| [`MobAttackInterval`](../../server/internal/world/entities.go#L24)<br>`const MobAttackInterval = 2500 * time.Millisecond` | Visibility and combat radii, matching the real server's cell view model. |
| [`mobMotionRun`](../../server/internal/world/entities.go#L26)<br>`const mobMotionRun = 1` | Visibility and combat radii, matching the real server's cell view model. |
| [`mobMotionWalk`](../../server/internal/world/entities.go#L27)<br>`const mobMotionWalk = 0` | Visibility and combat radii, matching the real server's cell view model. |
| [`NpcPatrolPoint`](../../server/internal/world/entities.go#L30)<br>`type NpcPatrolPoint` | See the source definition and the subsystem ownership contract. |
| [`MapNpc`](../../server/internal/world/entities.go#L35)<br>`type MapNpc` | See the source definition and the subsystem ownership contract. |
| [`MobArea`](../../server/internal/world/entities.go#L53)<br>`type MobArea` | See the source definition and the subsystem ownership contract. |
| [`normalized`](../../server/internal/world/entities.go#L55)<br>`func (a MobArea) normalized() MobArea` | See the source definition and the subsystem ownership contract. |
| [`clampX`](../../server/internal/world/entities.go#L58)<br>`func (a MobArea) clampX(x float32) float32` | See the source definition and the subsystem ownership contract. |
| [`clampZ`](../../server/internal/world/entities.go#L59)<br>`func (a MobArea) clampZ(z float32) float32` | See the source definition and the subsystem ownership contract. |
| [`contains`](../../server/internal/world/entities.go#L60)<br>`func (a MobArea) contains(x, z float32) bool` | See the source definition and the subsystem ownership contract. |
| [`MapMob`](../../server/internal/world/entities.go#L64)<br>`type MapMob` | See the source definition and the subsystem ownership contract. |
| [`MapPortal`](../../server/internal/world/entities.go#L122)<br>`type MapPortal` | See the source definition and the subsystem ownership contract. |
| [`Contains`](../../server/internal/world/entities.go#L133)<br>`func (p *MapPortal) Contains(x, y, z float32) bool` | See the source definition and the subsystem ownership contract. |
| [`IsOpen`](../../server/internal/world/entities.go#L140)<br>`func (p *MapPortal) IsOpen() bool` | See the source definition and the subsystem ownership contract. |
| [`AllowsFaction`](../../server/internal/world/entities.go#L141)<br>`func (p *MapPortal) AllowsFaction(f uint8) bool` | See the source definition and the subsystem ownership contract. |
| [`AllowsLevel`](../../server/internal/world/entities.go#L150)<br>`func (p *MapPortal) AllowsLevel(level uint16) bool` | See the source definition and the subsystem ownership contract. |
| [`MapSpawnArea`](../../server/internal/world/entities.go#L154)<br>`type MapSpawnArea` | MapSpawnArea is one stMapBind of the svmap: a rebirth group of boxes. Faction is its type: 1 light, 2 fury, 0 and 3 both. |
| [`SpawnBox`](../../server/internal/world/entities.go#L159)<br>`type SpawnBox` | See the source definition and the subsystem ownership contract. |
| [`distanceSq`](../../server/internal/world/entities.go#L166)<br>`func (a *MapSpawnArea) distanceSq(x, z float32) float32` | distanceSq is measured, as CMap::GetBindPos does, to the lower corner of the group's first box. |
| [`allowsFaction`](../../server/internal/world/entities.go#L170)<br>`func (a *MapSpawnArea) allowsFaction(f uint8) bool` | See the source definition and the subsystem ownership contract. |
| [`RandomPoint`](../../server/internal/world/entities.go#L184)<br>`func (a *MapSpawnArea) RandomPoint() (x, y, z float32)` | RandomPoint is GetBindPos's pick: a random box of the group and a random whole-unit offset inside it, 0.9 above its floor. |
| [`MapEntities`](../../server/internal/world/entities.go#L196)<br>`type MapEntities` | MapEntities is one map's NPCs, mobs, portals and spawn areas. |
| [`mapShard`](../../server/internal/world/entities.go#L203)<br>`type mapShard` | See the source definition and the subsystem ownership contract. |
| [`mapGrid`](../../server/internal/world/entities.go#L216)<br>`type mapGrid` | See the source definition and the subsystem ownership contract. |
| [`shard`](../../server/internal/world/entities.go#L221)<br>`func shard(mapID uint16) *mapShard` | See the source definition and the subsystem ownership contract. |
| [`shardIDs`](../../server/internal/world/entities.go#L232)<br>`func shardIDs() []uint16` | See the source definition and the subsystem ownership contract. |
| [`cloneMob`](../../server/internal/world/entities.go#L242)<br>`func cloneMob(m *MapMob) MapMob` | See the source definition and the subsystem ownership contract. |
| [`snapshotAt`](../../server/internal/world/entities.go#L256)<br>`func (m *MapMob) snapshotAt(now time.Time) MapMob` | See the source definition and the subsystem ownership contract. |
| [`EntitiesOnMap`](../../server/internal/world/entities.go#L264)<br>`func EntitiesOnMap(mapID uint16) MapEntities` | EntitiesOnMap returns a copy of a map's entities. |
| [`RegisterMapGrid`](../../server/internal/world/entities.go#L277)<br>`func RegisterMapGrid(mapID uint16, size uint32, mask []byte)` | See the source definition and the subsystem ownership contract. |
| [`mapSize`](../../server/internal/world/entities.go#L284)<br>`func mapSize(mapID uint16) (float32, bool)` | mapSize is the svmap's side length. |
| [`hasGrid`](../../server/internal/world/entities.go#L291)<br>`func hasGrid(mapID uint16) bool` | See the source definition and the subsystem ownership contract. |
| [`abs32`](../../server/internal/world/entities.go#L298)<br>`func abs32(v float32) float32` | See the source definition and the subsystem ownership contract. |
| [`DistanceXZ`](../../server/internal/world/entities.go#L300)<br>`func DistanceXZ(ax, az, bx, bz float32) float32` | See the source definition and the subsystem ownership contract. |
| [`cell`](../../server/internal/world/entities.go#L305)<br>`type cell` | See the source definition and the subsystem ownership contract. |
| [`coordCell`](../../server/internal/world/entities.go#L307)<br>`func coordCell(v float32, size float32) int32` | See the source definition and the subsystem ownership contract. |
| [`cellFor`](../../server/internal/world/entities.go#L314)<br>`func cellFor(x, z float32) cell` | See the source definition and the subsystem ownership contract. |
| [`isNeighborCell`](../../server/internal/world/entities.go#L319)<br>`func isNeighborCell(a, b cell) bool` | isNeighborCell is the 5x5 view window. |
| [`EntitiesNear`](../../server/internal/world/entities.go#L331)<br>`func EntitiesNear(mapID uint16, x, z float32) MapEntities` | EntitiesNear returns the entities in the 5x5 cell window around (x, z). |
| [`EntitiesInView`](../../server/internal/world/entities.go#L337)<br>`func EntitiesInView(mapID uint16, viewer cell) MapEntities` | EntitiesInView returns the NPCs and the visible mobs whose cell is in the 5x5 window around the viewer's cell. |
| [`IsMapWalkable`](../../server/internal/world/entities.go#L365)<br>`func IsMapWalkable(mapID uint16, x, z float32) bool` | See the source definition and the subsystem ownership contract. |
| [`walkableInMask`](../../server/internal/world/entities.go#L375)<br>`func walkableInMask(size uint32, mask []byte, x, z float32) bool` | See the source definition and the subsystem ownership contract. |
| [`MobByID`](../../server/internal/world/entities.go#L394)<br>`func MobByID(mapID uint16, id uint32) (MapMob, bool)` | See the source definition and the subsystem ownership contract. |
| [`MobDamageResult`](../../server/internal/world/entities.go#L406)<br>`type MobDamageResult` | See the source definition and the subsystem ownership contract. |
| [`DamageMob`](../../server/internal/world/entities.go#L416)<br>`func DamageMob(mapID uint16, id uint32, damage uint32) (MobDamageResult, bool)` | DamageMob applies damage unless the mob is dead or resetting home. |
| [`DamageMobFrom`](../../server/internal/world/entities.go#L419)<br>`func DamageMobFrom(mapID uint16, id uint32, damage, attacker uint32) (MobDamageResult, bool)` | See the source definition and the subsystem ownership contract. |
| [`PlayerCombatSnapshot`](../../server/internal/world/entities.go#L463)<br>`type PlayerCombatSnapshot` | See the source definition and the subsystem ownership contract. |
| [`MobAttackAction`](../../server/internal/world/entities.go#L478)<br>`type MobAttackAction` | See the source definition and the subsystem ownership contract. |
| [`mapPacket`](../../server/internal/world/entities.go#L489)<br>`type mapPacket` | mapPacket is a packet to broadcast near (x, z) on a map. |
| [`MobCombatTick`](../../server/internal/world/entities.go#L495)<br>`type MobCombatTick` | See the source definition and the subsystem ownership contract. |
| [`viewerBuckets`](../../server/internal/world/entities.go#L501)<br>`type viewerBuckets` | See the source definition and the subsystem ownership contract. |
| [`bucketViewers`](../../server/internal/world/entities.go#L503)<br>`func bucketViewers(viewers [][2]float32) viewerBuckets` | See the source definition and the subsystem ownership contract. |
| [`nearAny`](../../server/internal/world/entities.go#L512)<br>`func (b viewerBuckets) nearAny(x, z float32, radius float32) bool` | See the source definition and the subsystem ownership contract. |
| [`TickMobRespawns`](../../server/internal/world/entities.go#L530)<br>`func TickMobRespawns(viewersByMap map[uint16][][2]float32, maxPerMap int) []entityMove` | TickMobRespawns revives due mobs at a fresh position in their area. |
| [`stepPatrol`](../../server/internal/world/entities.go#L607)<br>`func (n *MapNpc) stepPatrol(now time.Time, mapID uint16) bool` | CNpc::UpdateMove pauses at each waypoint and reverses at route endpoints. |
| [`mobGroup`](../../server/internal/world/entities.go#L654)<br>`type mobGroup` | mobGroup is stMobRespawn: an svmap area entry of Count mobs. The first corpse removal from a full group starts its timer (RespawnTime, +-60 min for Attackspecial3 >= 9, CZone::UpdateMobDeath); CZone::UpdateMobRespawn then brings back one mob each time the timer runs out, at a random spot of the area, restarting the timer. |
| [`died`](../../server/internal/world/entities.go#L660)<br>`func (m *MapMob) died(now time.Time)` | died books a death (shard lock held). |
| [`removeCorpse`](../../server/internal/world/entities.go#L678)<br>`func (m *MapMob) removeCorpse(now time.Time)` | removeCorpse books the native UpdateMobDeath removal (shard lock held). Group counts include corpses until this point; later deaths cannot start another timer while the group is still being replenished. |
| [`respawnSpot`](../../server/internal/world/entities.go#L699)<br>`func (m *MapMob) respawnSpot(mapID uint16) (float32, float32, bool)` | respawnSpot is UpdateMobRespawn's position: lower + rand(width) on each axis; ok false when it is not walkable (the next update tries again). |
| [`relocate`](../../server/internal/world/entities.go#L715)<br>`func (n *MapNpc) relocate(mapID uint16, motion uint8) entityMove` | See the source definition and the subsystem ownership contract. |
| [`mobCorpseDelay`](../../server/internal/world/entities.go#L723)<br>`const mobCorpseDelay = 5 * time.Second` | mobCorpseDelay is CMob::Damage's death tick: the corpse stays 5 s. |
| [`visible`](../../server/internal/world/entities.go#L726)<br>`func (m *MapMob) visible() bool` | visibleMob: alive, or dead with its corpse still shown. |
| [`TickMobCorpses`](../../server/internal/world/entities.go#L730)<br>`func TickMobCorpses() []entityMove` | TickMobCorpses is CZone::UpdateMobDeath: corpses older than 5 s leave the view (0x0602). |
| [`SetMobAggro`](../../server/internal/world/entities.go#L753)<br>`func SetMobAggro(mapID uint16, id, characterID uint32)` | SetMobAggro is CMob::SetTarget for a player's hit: an idle mob takes the attacker as its target. The ledger grows through AddMobAggro, with the original amount (hitAggro): the attacker's REC times ten plus the skill's Arrow, whatever the damage. CMob::AttackAI later turns to the top of the ledger (ReTargetByAggro). |
| [`AddMobAggro`](../../server/internal/world/entities.go#L755)<br>`func AddMobAggro(mapID uint16, id, characterID uint32, amount int64)` | See the source definition and the subsystem ownership contract. |
| [`ForceMobTarget`](../../server/internal/world/entities.go#L774)<br>`func ForceMobTarget(mapID uint16, id, characterID uint32)` | ForceMobTarget is CMob::SetTarget(user, true): the mob turns to the user at once; the next ReTargetByAggro may still pick whoever has more aggro. |
| [`ClearMobAggro`](../../server/internal/world/entities.go#L788)<br>`func ClearMobAggro(mapID uint16, id uint32)` | ClearMobAggro is CMob::Attack after killing its target: back to idle (status 0) with the dead user's aggro gone; UpdateIdle walks it home when it is more than 10 units away. |
| [`mobLeash`](../../server/internal/world/entities.go#L802)<br>`const mobLeash = 45.0` | Mob combat as CMob::UpdateIdle / UpdateAttack / Chase / ReturnHome run it. target farther than this from home or from the mob |
| [`mobHomeDone`](../../server/internal/world/entities.go#L803)<br>`const mobHomeDone = 10.0` | Mob combat as CMob::UpdateIdle / UpdateAttack / Chase / ReturnHome run it. ReturnHome completes within 10 units of home |
| [`TickMobCombat`](../../server/internal/world/entities.go#L807)<br>`func TickMobCombat(players []PlayerCombatSnapshot, maxPerMap int) MobCombatTick` | TickMobCombat advances targeting, attacks, chasing and returning home. |
| [`liveTarget`](../../server/internal/world/entities.go#L864)<br>`func (m *MapMob) liveTarget(players []PlayerCombatSnapshot) *PlayerCombatSnapshot` | See the source definition and the subsystem ownership contract. |
| [`dropTarget`](../../server/internal/world/entities.go#L877)<br>`func (m *MapMob) dropTarget(home bool)` | dropTarget: status 0 (idle) or 3 (return home). |
| [`findTarget`](../../server/internal/world/entities.go#L885)<br>`func (m *MapMob) findTarget(now time.Time, players []PlayerCombatSnapshot, extended bool) *PlayerCombatSnapshot` | findTarget is CMob::FindTarget for AI 1 and 3: the first living user of another country within ChaseRange, cell by cell over the 3x3 cells around the mob. |
| [`targetWeakest`](../../server/internal/world/entities.go#L913)<br>`func (m *MapMob) targetWeakest(players []PlayerCombatSnapshot) *PlayerCombatSnapshot` | targetWeakest is CMob::FindAttackUserMinHP over the damage ledger. |
| [`retargetByAggro`](../../server/internal/world/entities.go#L929)<br>`func (m *MapMob) retargetByAggro(players []PlayerCombatSnapshot)` | retargetByAggro is CMob::ReTargetByAggro: the living attacker on the map with the most aggro. |
| [`stepChase`](../../server/internal/world/entities.go#L944)<br>`func (m *MapMob) stepChase(now time.Time, mapID uint16, t *PlayerCombatSnapshot, d float32, mods *mobMods, tick *MobCombatTick) bool` | stepChase advances toward the target, stopping just inside attack reach. |
| [`stepHome`](../../server/internal/world/entities.go#L980)<br>`func (m *MapMob) stepHome(now time.Time, mapID uint16, mods *mobMods, tick *MobCombatTick) bool` | stepHome runs toward the spawn and clears effects and aggro on arrival. CMob::ReturnHome sends movement mode 1 with the chase step and time. |
| [`AddEntities`](../../server/internal/world/entities.go#L1023)<br>`func AddEntities(mapID uint16, e MapEntities)` | See the source definition and the subsystem ownership contract. |
| [`FindNearestSpawnArea`](../../server/internal/world/entities.go#L1039)<br>`func FindNearestSpawnArea(mapID uint16, faction uint8, x, z float32) (MapSpawnArea, bool)` | FindNearestSpawnArea is CMap::GetBindPos's group choice: the nearest non-empty group open to the faction; when none is, the first group (the original's index stays 0), and ok is false only if that group is empty. |
| [`FindPortalByIndex`](../../server/internal/world/entities.go#L1056)<br>`func FindPortalByIndex(mapID uint16, index uint8) (MapPortal, bool)` | See the source definition and the subsystem ownership contract. |
| [`FindPortalAt`](../../server/internal/world/entities.go#L1065)<br>`func FindPortalAt(mapID uint16, x, y, z float32) (MapPortal, bool)` | See the source definition and the subsystem ownership contract. |
| [`NewPatrolNpc`](../../server/internal/world/entities.go#L1082)<br>`func NewPatrolNpc(npcType uint8, typeID uint16, points []NpcPatrolPoint) (MapNpc, bool)` | See the source definition and the subsystem ownership contract. |
| [`NewStaticMob`](../../server/internal/world/entities.go#L1093)<br>`func NewStaticMob(def *db.MobDef, mobID uint16, x, z float32, area *MobArea, respawnSeconds uint64) MapMob` | NewStaticMob builds a mob from its definition; respawn 0 never respawns. |
| [`pickRespawnPosition`](../../server/internal/world/entities.go#L1116)<br>`func (m *MapMob) pickRespawnPosition(mapID uint16) (float32, float32)` | See the source definition and the subsystem ownership contract. |
| [`mobSpeedOffsets`](../../server/internal/world/entities.go#L1139)<br>`var mobSpeedOffsets` | mobSpeedOffsets are the original move-speed step offsets (0x00542538); mobs start at speed 2 (no offset). |
| [`mobHomeRadius`](../../server/internal/world/entities.go#L1141)<br>`const mobHomeRadius = 10.0` | See the source definition and the subsystem ownership contract. |
| [`advanceRoam`](../../server/internal/world/entities.go#L1146)<br>`func (m *MapMob) advanceRoam(now time.Time, mapID uint16, tick *MobCombatTick)` | advanceRoam is the idle state's wander (update_mob_ai_idle): every 10 s + NormalTime + up to 30 s, a step of NormalStep + 0..4 units in a random direction, staying within 10 units of home. |
| [`chaseStep`](../../server/internal/world/entities.go#L1180)<br>`func (m *MapMob) chaseStep(mods *mobMods) float32` | chaseStep is the maximum path segment for a chase or return home. |
| [`attackReach`](../../server/internal/world/entities.go#L1189)<br>`func (m *MapMob) attackReach() float32` | attackReach is vertical_mobs_attack_range: Size + Attackrange + 0.3. |
| [`mobAttackAnimation`](../../server/internal/world/entities.go#L1197)<br>`const mobAttackAnimation = time.Second` | mobAttackAnimation caps how long a mob holds still while an attack plays. |
| [`later`](../../server/internal/world/entities.go#L1199)<br>`func later(a, b time.Time) time.Time` | See the source definition and the subsystem ownership contract. |
| [`mobAttackIdle`](../../server/internal/world/entities.go#L1208)<br>`const mobAttackIdle = 2500 * time.Millisecond` | mobAttackIdle is AttackAI's wait when no slot is ready (2500 ms plus g_ConstMobAttackSpeed, 0 at normal speed). |
| [`pickAttackSlot`](../../server/internal/world/entities.go#L1213)<br>`func (m *MapMob) pickAttackSlot(now time.Time, mods *mobMods, inReach func(int, float32) bool, rollSpell func() bool) (slotIndex uint8, found, declinedSpell bool)` | pickAttackSlot keeps CMob::AttackAI's slot order and eligibility checks. GX draws once per decision for routine spells. A declined spell remains ready and the mob can still choose a basic or conditional attack. |
| [`useSlot`](../../server/internal/world/entities.go#L1270)<br>`func (m *MapMob) useSlot(slot uint8, now time.Time) []byte` | useSlot starts a slot's cooldown and the next attack time. A mob heal skill (TypeEffect 2) applies at once and returns its HP packet. |
| [`attackSpeedFactor`](../../server/internal/world/entities.go#L1290)<br>`func (m *MapMob) attackSpeedFactor() float32` | See the source definition and the subsystem ownership contract. |
| [`movable`](../../server/internal/world/entities.go#L1298)<br>`func (m *MapMob) movable() bool` | movable: not stunned, asleep or stopped. |
| [`attackInterval`](../../server/internal/world/entities.go#L1303)<br>`func (m *MapMob) attackInterval() time.Duration` | See the source definition and the subsystem ownership contract. |
| [`roamWalkable`](../../server/internal/world/entities.go#L1310)<br>`func roamWalkable(mapID uint16, x, z float32, a MobArea) bool` | See the source definition and the subsystem ownership contract. |
| [`buildNpcEnter`](../../server/internal/world/entities.go#L1316)<br>`func buildNpcEnter(n *MapNpc) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildNpcLeave`](../../server/internal/world/entities.go#L1331)<br>`func buildNpcLeave(id uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildNpcMove`](../../server/internal/world/entities.go#L1337)<br>`func buildNpcMove(n *MapNpc, motion uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`wireHeightKnown`](../../server/internal/world/entities.go#L1347)<br>`func (m *MapMob) wireHeightKnown() bool` | See the source definition and the subsystem ownership contract. |
| [`buildMobEnter`](../../server/internal/world/entities.go#L1351)<br>`func buildMobEnter(m *MapMob, isNew bool) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildMobMove`](../../server/internal/world/entities.go#L1373)<br>`func buildMobMove(m *MapMob, motion uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildMobLeave`](../../server/internal/world/entities.go#L1393)<br>`func buildMobLeave(id uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildMobTargetHP`](../../server/internal/world/entities.go#L1400)<br>`func buildMobTargetHP(m *MapMob) []byte` | buildMobTargetHP (0x0305): mob id, HP (client reader 0x0058f2d0). |
| [`buildMobState`](../../server/internal/world/entities.go#L1408)<br>`func buildMobState(m *MapMob) []byte` | buildMobState (0x0304): mob id, HP, attack speed, move speed. |
| [`buildCharacterMobAttack`](../../server/internal/world/entities.go#L1418)<br>`func buildCharacterMobAttack(characterID, mobID uint32, result uint8, damage uint16) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildMobAttackCharacter`](../../server/internal/world/entities.go#L1429)<br>`func buildMobAttackCharacter(result uint8, mobID, characterID uint32, damage uint16) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildCharacterDeath`](../../server/internal/world/entities.go#L1440)<br>`func buildCharacterDeath(characterID uint32, deathType uint8, killerID uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildCharacterDeathState`](../../server/internal/world/entities.go#L1448)<br>`func buildCharacterDeathState(characterID uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildCharacterReviveInit`](../../server/internal/world/entities.go#L1456)<br>`func buildCharacterReviveInit(characterID uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildCharacterResurrect`](../../server/internal/world/entities.go#L1464)<br>`func buildCharacterResurrect(characterID uint32, kind uint8, exp uint32, x, y, z float32) []byte` | buildCharacterResurrect (0x0551) is CUser::Rebirth's packet: id, kind (1 in place, 4 bind point, 5 party leader), current exp, position. |
| [`buildCharacterReviveHitpoints`](../../server/internal/world/entities.go#L1475)<br>`func buildCharacterReviveHitpoints(characterID, hp, mp, sp uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildMobDeath`](../../server/internal/world/entities.go#L1484)<br>`func buildMobDeath(mobID, killerID uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`deathPositionPacket`](../../server/internal/world/entities.go#L1488)<br>`func (m *MapMob) deathPositionPacket() []byte` | See the source definition and the subsystem ownership contract. |
| [`buildMobDeathBy`](../../server/internal/world/entities.go#L1495)<br>`func buildMobDeathBy(mobID, killerID uint32, deathType uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildMobPeriodicDamage`](../../server/internal/world/entities.go#L1503)<br>`func buildMobPeriodicDamage(mobID uint32, skillID uint16, level uint8, hp, sp, mp uint16) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildMobCurrentHP`](../../server/internal/world/entities.go#L1514)<br>`func buildMobCurrentHP(mobID, hp uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`HealMob`](../../server/internal/world/entities.go#L1522)<br>`func HealMob(mapID uint16, id, amount uint32) (MapMob, uint32, bool)` | HealMob restores HP to a living mob (a drain's heal); returns the mob. |
| [`recordLootDamage`](../../server/internal/world/entities.go#L1538)<br>`func (m *MapMob) recordLootDamage(id, damage uint32)` | recordLootDamage books damage against the attacker for the loot owner. |

<details><summary>Type definitions and owned fields</summary>

```go
type NpcPatrolPoint struct {
	X, Y, Z float32
	Angle   uint16
}
```

```go
type MapNpc struct {
	Summoned         bool
	ID               uint32
	Type             uint8
	TypeID           uint16
	PosX, PosY, PosZ float32
	Angle            uint16
	patrol           []NpcPatrolPoint
	patrolIndex      int
	cell             cell
	hasCell          bool
	backward         bool      // walking the patrol points back
	nextMoveAt       time.Time // CNpc +0x4c
	guardReady       time.Time
	guardTarget      guardTarget
	guardReturning   bool
}
```

```go
type MobArea struct{ LowerX, LowerZ, UpperX, UpperZ float32 }
```

```go
type MapMob struct {
	aiMapID          uint16
	aiSummoned       bool
	aiAttackType     uint8
	lua              mobAIState
	event            *eventActor
	eventGeneration  uint64
	eventFresh       bool
	eventDamage      []eventDamageEntry
	lootDamage       []eventDamageEntry // damage per attacker: who earns the loot
	elementRestored  bool               // RemApplySkillDebuff68 restores base even with another68 active.
	ID               uint32
	MobID            uint16
	Level            uint16
	PosX, PosY, PosZ float32
	nav              *mobNav
	navPath          []mapnav.Point
	navGoal          mapnav.Point
	navRefresh       time.Time
	// Pos is the destination while a movement segment is in progress.
	moveFrom             mapnav.Point
	moveStart            time.Time
	moveDuration         time.Duration
	moveMotion           uint8
	moveSpeedTier        uint8
	stoppedOnDeath       bool
	spawnMinY, spawnMaxY float32
	spawnHeightKnown     bool
	homeY                float32
	homeX, homeZ         float32
	CurrentHP, MaxHP     uint32
	MoneyMin, MoneyMax   uint32
	Exp                  uint32
	Area                 *MobArea
	respawnDelay         time.Duration // 0: no respawn (ungrouped) / immediate (grouped)
	group                *mobGroup     // stMobRespawn: the area entry it belongs to
	respawnAt            time.Time
	roamStep             uint32
	Attack               uint32
	AI                   uint8
	ChaseRange           float32
	AggroTarget          uint32 // 0: none
	lastAttackAt         time.Time
	nextAttackAt         time.Time        // CMob m_dwNextAttackTime
	noTargetUntil        time.Time        // CMob::Init: no target search for 5 s
	aggro                map[uint32]int64 // CDamage ledger: damage and skill aggro
	slotReadyAt          [3]time.Time     // per attack slot (AttackAI)
	returningHome        bool
	def                  *db.MobDef
	nextMoveAt           time.Time   // idle wander
	nextChaseAt          time.Time   // chase / return step
	Effects              []MobEffect // applied skills (TypeEffect 3)
	cell                 cell        // view cell, with the move hysteresis
	hasCell              bool
	diedAt               time.Time
	corpseGone           bool // the corpse left the view (5 s after death)
}
```

```go
type MapPortal struct {
	Index                     uint8
	PortalID                  int32
	FactionOrPortalID         int32
	MinLevel, MaxLevel        uint16
	TargetMapID               uint16
	PosX, PosY, PosZ          float32
	TargetX, TargetY, TargetZ float32
	volume                    *portalVolume
}
```

```go
type MapSpawnArea struct {
	Faction int32
	Boxes   []SpawnBox
}
```

```go
type SpawnBox struct {
	LowerX, LowerY, LowerZ float32
	UpperX, UpperY, UpperZ float32
}
```

```go
type MapEntities struct {
	NPCs       []MapNpc
	Mobs       []MapMob
	Portals    []MapPortal
	SpawnAreas []MapSpawnArea
}
```

```go
type mapShard struct {
	guardClaims      map[guardTarget]time.Time
	items            map[uint32]*mapItemEntry
	lua              *ai.VM
	aiCommands       []mobAICommand
	aiCommandCost    int
	aiPendingSummons int
	npcRevision      atomic.Uint64
	questNPCs        questNPCCatalogCache
	mu               sync.Mutex
	entities         MapEntities
}
```

```go
type mapGrid struct {
	size uint32
	mask []byte
}
```

```go
type cell struct{ x, z int32 }
```

```go
type MobDamageResult struct {
	Mob            MapMob
	Speed          []byte // 0x0609 when the hit woke the mob
	DamageHP       uint16
	AppliedHP      uint32
	Killed         bool
	RespawnSeconds uint64
}
```

```go
type PlayerCombatSnapshot struct {
	CharacterID      uint32
	MapID            uint16
	PosX, PosY, PosZ float32
	Alive            bool
	Faction          uint8
	Cell             cell
	HP               uint32
	Invisible        bool
	// GMUndetected is a GM in /char on or /attack on mode. Unlike skill stealth,
	// which guards see through, the original admin flags hide the user from
	// every NPC (bVisible and bAttackable).
	GMUndetected bool
}
```

```go
type MobAttackAction struct {
	MobGlobalID       uint32
	MobID             uint16
	Attack            uint32
	TargetCharacterID uint32
	Slot              uint8
	MobX, MobZ        float32
	SkillID           uint16
}
```

```go
type mapPacket struct {
	MapID  uint16
	X, Z   float32
	Packet []byte
}
```

```go
type MobCombatTick struct {
	Attacks  []MobAttackAction
	Moves    []entityMove
	Restores []mapPacket
}
```

```go
type viewerBuckets map[cell][][2]float32
```

```go
type mobGroup struct {
	count, alive int
	respawnAt    time.Time
}
```

</details>

### evolution.go

[server/internal/world/evolution.go](../../server/internal/world/evolution.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`evolutionAction`](../../server/internal/world/evolution.go#L11)<br>`type evolutionAction` | See the source definition and the subsystem ownership contract. |
| [`evolutionActions`](../../server/internal/world/evolution.go#L16)<br>`func evolutionActions(mob *db.MobDef) (out [2]evolutionAction)` | See the source definition and the subsystem ownership contract. |
| [`updateEvolutionGrant`](../../server/internal/world/evolution.go#L39)<br>`func (s *Session) updateEvolutionGrant(look formLook)` | See the source definition and the subsystem ownership contract. |
| [`sendEvolutionBar`](../../server/internal/world/evolution.go#L57)<br>`func (s *Session) sendEvolutionBar()` | See the source definition and the subsystem ownership contract. |
| [`evolutionSkill`](../../server/internal/world/evolution.go#L77)<br>`func (s *Session) evolutionSkill(generation uint32, cell uint8) *db.SkillDef` | See the source definition and the subsystem ownership contract. |
| [`handleEvolutionCast`](../../server/internal/world/evolution.go#L85)<br>`func (s *Session) handleEvolutionCast(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`definitionReady`](../../server/internal/world/evolution.go#L111)<br>`func (s *Session) definitionReady(def *db.SkillDef) bool` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type evolutionAction struct {
	skill *db.SkillDef
	slot  uint8
}
```

</details>

### fall_damage.go

[server/internal/world/fall_damage.go](../../server/internal/world/fall_damage.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/config`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`fallDamage`](../../server/internal/world/fall_damage.go#L11)<br>`func fallDamage(drop float32, maxHP uint32) uint32` | GX rule: six safe units; half the previous quadratic damage, capped at max HP. |
| [`applyFallDamage`](../../server/internal/world/fall_damage.go#L18)<br>`func (s *Session) applyFallDamage(drop float32)` | See the source definition and the subsystem ownership contract. |
| [`applyEnvironmentalDamage`](../../server/internal/world/fall_damage.go#L27)<br>`func (s *Session) applyEnvironmentalDamage(amount uint32, opcode proto.Opcode, cause string)` | See the source definition and the subsystem ownership contract. |

### firecracker.go

[server/internal/world/firecracker.go](../../server/internal/world/firecracker.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`specialFirecracker`](../../server/internal/world/firecracker.go#L12)<br>`const specialFirecracker = 46` | See the source definition and the subsystem ownership contract. |
| [`firecrackerRadius`](../../server/internal/world/firecracker.go#L13)<br>`const firecrackerRadius = 40` | See the source definition and the subsystem ownership contract. |
| [`useFirecracker`](../../server/internal/world/firecracker.go#L16)<br>`func (s *Session) useFirecracker(index int, def *db.ItemDef) bool` | See the source definition and the subsystem ownership contract. |

### firethorn.go

[server/internal/world/firethorn.go](../../server/internal/world/firethorn.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`detailFireThorn`](../../server/internal/world/firethorn.go#L12)<br>`const detailFireThorn = 84` | buff: thorns around the caster |
| [`fireThornTick`](../../server/internal/world/firethorn.go#L13)<br>`const fireThornTick = 2 * time.Second` | the pulse interval |
| [`pulseHeightReach`](../../server/internal/world/firethorn.go#L17)<br>`const pulseHeightReach = 4` | pulseHeightReach is how far above or below the pulse's centre a target may stand and still burn: the height counts, but leniently (slopes and steps must not narrow the ground reach). |
| [`fireThornRange`](../../server/internal/world/firethorn.go#L19)<br>`const fireThornRange = 6` | fireThornRange is how far the thorns reach, whatever the skill's ApplyRange. |
| [`isFireThorn`](../../server/internal/world/firethorn.go#L22)<br>`func isFireThorn(b *ActiveBuff) bool` | See the source definition and the subsystem ownership contract. |
| [`tickFireThorn`](../../server/internal/world/firethorn.go#L30)<br>`func (s *Session) tickFireThorn(now time.Time)` | tickFireThorn runs Fire Thorn (TypeEffect 4, TypeDetail 84). While the buff lasts, every two seconds each hostile mob and user within ApplyRange of the caster takes the skill's DamageHP as fixed damage: no hit roll, no stat formula, no defense, no absorption. Mobs also gain aggro on the caster. |
| [`fireThornPulse`](../../server/internal/world/firethorn.go#L53)<br>`func (s *Session) fireThornPulse(def *db.SkillDef)` | See the source definition and the subsystem ownership contract. |
| [`fixedDamagePulse`](../../server/internal/world/firethorn.go#L62)<br>`func (s *Session) fixedDamagePulse(def *db.SkillDef, x, y, z, reach float32)` | fixedDamagePulse hurts every hostile mob and user within ApplyRange of a point with the skill's DamageHP as fixed damage (Fire Thorn around the caster, Fire Bush around its curtain of fire). The client plays DA_self_fire on each target. |
| [`applyFixedHit`](../../server/internal/world/firethorn.go#L102)<br>`func (s *Session) applyFixedHit(h pendingUserHit) bool` | applyFixedHit is the target side of a fixed-damage skill hit (Fire Thorn): the damage is final, so no accuracy, defense or absorption applies. Nearby users get the periodic damage packet 0x050F; true when this character died. |

### forge_quote.go

[server/internal/world/forge_quote.go](../../server/internal/world/forge_quote.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`forgeCost`](../../server/internal/world/forge_quote.go#L10)<br>`func forgeCost(original uint32) uint32` | Costs are the original ReqIg table, with no character-dependent discount. |
| [`gemChance`](../../server/internal/world/forge_quote.go#L11)<br>`func gemChance(_ *db.Character, faction uint8, gem, hammer *db.ItemDef) int` | See the source definition and the subsystem ownership contract. |
| [`validForgeHammer`](../../server/internal/world/forge_quote.go#L18)<br>`func validForgeHammer(d *db.ItemDef, special uint8) bool` | See the source definition and the subsystem ownership contract. |
| [`extractionChance`](../../server/internal/world/forge_quote.go#L21)<br>`func extractionChance(d *db.ItemDef) int` | See the source definition and the subsystem ownership contract. |
| [`gemCost`](../../server/internal/world/forge_quote.go#L22)<br>`func gemCost(_ *db.Character, gem *db.ItemDef) uint32` | See the source definition and the subsystem ownership contract. |
| [`enchantRate`](../../server/internal/world/forge_quote.go#L23)<br>`func enchantRate(item db.Item, _ uint16) (uint32, bool)` | See the source definition and the subsystem ownership contract. |
| [`forgeQuote`](../../server/internal/world/forge_quote.go#L40)<br>`func (s *Session) forgeQuote(q proto.ForgeQuoteRequest) proto.ForgeQuote` | See the source definition and the subsystem ownership contract. |
| [`handleForgeQuote`](../../server/internal/world/forge_quote.go#L105)<br>`func (s *Session) handleForgeQuote(p []byte)` | See the source definition and the subsystem ownership contract. |

### friendly.go

[server/internal/world/friendly.go](../../server/internal/world/friendly.go#L1)

Direct GX dependencies: `internal/proto`, `internal/skillrange`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`targetTypeOther`](../../server/internal/world/friendly.go#L21)<br>`const targetTypeOther = 3` | See the source definition and the subsystem ownership contract. |
| [`targetTypeParty`](../../server/internal/world/friendly.go#L22)<br>`const targetTypeParty = 4` | See the source definition and the subsystem ownership contract. |
| [`targetTypeNParty`](../../server/internal/world/friendly.go#L23)<br>`const targetTypeNParty = 5` | See the source definition and the subsystem ownership contract. |
| [`targetTypeOtherAlt`](../../server/internal/world/friendly.go#L24)<br>`const targetTypeOtherAlt = 7` | See the source definition and the subsystem ownership contract. |
| [`detailHeal`](../../server/internal/world/friendly.go#L26)<br>`const detailHeal = 41` | See the source definition and the subsystem ownership contract. |
| [`detailCure`](../../server/internal/world/friendly.go#L27)<br>`const detailCure = 42` | See the source definition and the subsystem ownership contract. |
| [`detailSacrificeHeal`](../../server/internal/world/friendly.go#L28)<br>`const detailSacrificeHeal = 43` | Sacrifice: transfers the caster's maximum pool to the target |
| [`skillResultPty`](../../server/internal/world/friendly.go#L29)<br>`const skillResultPty = 4` | See the source definition and the subsystem ownership contract. |
| [`pendingFriendly`](../../server/internal/world/friendly.go#L33)<br>`type pendingFriendly` | pendingFriendly is a heal, cure or buff to apply in a target's session. |
| [`QueueFriendly`](../../server/internal/world/friendly.go#L42)<br>`func (p *ConnectedPlayer) QueueFriendly(f pendingFriendly)` | See the source definition and the subsystem ownership contract. |
| [`TakeFriendly`](../../server/internal/world/friendly.go#L53)<br>`func (p *ConnectedPlayer) TakeFriendly() []pendingFriendly` | See the source definition and the subsystem ownership contract. |
| [`teamUsable`](../../server/internal/world/friendly.go#L62)<br>`func (s *Session) teamUsable(p *ConnectedPlayer) bool` | teamUsable is IsTeamUsable (guild duels not yet). |
| [`friendlyTargets`](../../server/internal/world/friendly.go#L86)<br>`func (s *Session) friendlyTargets(def *db.SkillDef, target uint32, completed ...bool) ([]*ConnectedPlayer, bool, bool)` | friendlyTargets resolves the users a heal or buff reaches; ok false is ChkTargetUser's refusal (0x0511 result 11). |
| [`friendlySkill`](../../server/internal/world/friendly.go#L151)<br>`func (s *Session) friendlySkill(k db.CharacterSkill, def *db.SkillDef, target uint32, stats CharacterStats, completed ...bool)` | friendlySkill casts a heal (TypeEffect 2) or buff (4) on its targets. |
| [`applyFriendly`](../../server/internal/world/friendly.go#L214)<br>`func (s *Session) applyFriendly(f pendingFriendly)` | applyFriendly is HealUser / BuffUser on this user. |
| [`tickFriendly`](../../server/internal/world/friendly.go#L240)<br>`func (s *Session) tickFriendly()` | See the source definition and the subsystem ownership contract. |
| [`cancelBuff`](../../server/internal/world/friendly.go#L250)<br>`func (s *Session) cancelBuff(skillID uint32)` | cancelBuff is 0x0511 with 0xFF: CUser::RemApplySkill of an own buff. |
| [`friendlySkillRecipient`](../../server/internal/world/friendly.go#L258)<br>`func (s *Session) friendlySkillRecipient(def *db.SkillDef, target uint32) uint32` | A selected enemy is a combat selection, not a recipient for single-target help. Stale/missing allies and allies out of range remain errors. |
| [`sendFriendlySkillFailure`](../../server/internal/world/friendly.go#L271)<br>`func (s *Session) sendFriendlySkillFailure(k db.CharacterSkill, def *db.SkillDef, target uint32)` | See the source definition and the subsystem ownership contract. |
| [`isRemoteWarehouse`](../../server/internal/world/friendly.go#L281)<br>`func isRemoteWarehouse(b *ActiveBuff) bool` | isRemoteWarehouse marks Warehouse Call (ability 39): it cannot be cancelled by hand, so a stray right click never drops the remote warehouse. |

<details><summary>Type definitions and owned fields</summary>

```go
type pendingFriendly struct {
	Source       uint32
	Op           packet.Opcode
	SkillID      uint16
	Level        uint8
	HP, MP, SP   uint32
	Cure, IsBuff bool
}
```

</details>

### gameini.go

[server/internal/world/gameini.go](../../server/internal/world/gameini.go#L1)

Direct GX dependencies: `internal/catalogfiles`, `internal/maprules`, `server/internal/config`, `server/internal/db`, `server/internal/logx`.

| Declaration | Responsibility / source comment |
|---|---|
| [`gameDataDir`](../../server/internal/world/gameini.go#L17)<br>`func gameDataDir() string` | See the source definition and the subsystem ownership contract. |
| [`iniFile`](../../server/internal/world/gameini.go#L19)<br>`type iniFile` | See the source definition and the subsystem ownership contract. |
| [`readIni`](../../server/internal/world/gameini.go#L21)<br>`func readIni(path string) (iniFile, error)` | See the source definition and the subsystem ownership contract. |
| [`numberRE`](../../server/internal/world/gameini.go#L24)<br>`var numberRE` | numbers extracts every number in s ("(1, 577.98, 78.56, 1762.12)"). |
| [`numbers`](../../server/internal/world/gameini.go#L26)<br>`func numbers(s string) []float64` | See the source definition and the subsystem ownership contract. |
| [`MapPos`](../../server/internal/world/gameini.go#L37)<br>`type MapPos` | MapPos is a map and position from the original configs. |
| [`ZoneConfig`](../../server/internal/world/gameini.go#L43)<br>`type ZoneConfig` | ZoneConfig holds the runtime rules for one map. |
| [`MapConfig`](../../server/internal/world/gameini.go#L64)<br>`type MapConfig` | MapConfig holds the map definitions and race starts loaded at boot. |
| [`mapConfig`](../../server/internal/world/gameini.go#L71)<br>`var mapConfig` | See the source definition and the subsystem ownership contract. |
| [`loadMapProperties`](../../server/internal/world/gameini.go#L77)<br>`func loadMapProperties(log *logx.Logger) error` | See the source definition and the subsystem ownership contract. |
| [`buildMapProperties`](../../server/internal/world/gameini.go#L92)<br>`func buildMapProperties(c maprules.Catalog) (MapConfig, error)` | See the source definition and the subsystem ownership contract. |
| [`rewardRates`](../../server/internal/world/gameini.go#L126)<br>`var rewardRates` | See the source definition and the subsystem ownership contract. |
| [`LoadGameConfig`](../../server/internal/world/gameini.go#L129)<br>`func LoadGameConfig(log *logx.Logger) error` | LoadGameConfig loads the original PS_Game ini files from data/game/. |
| [`rebirthPosition`](../../server/internal/world/gameini.go#L183)<br>`func rebirthPosition(mapID uint16, faction uint8) (MapPos, bool)` | rebirthPosition is where a character of faction revives on mapID, per CMap::LoadConfig: Light and Fury have independent destinations. |
| [`bindPosition`](../../server/internal/world/gameini.go#L201)<br>`func bindPosition(mapID uint16, faction uint8, x, y, z float32) (uint16, float32, float32, float32)` | bindPosition is CMap::GetBindPos 0x004295a0: the configured faction rebirth point when its map exists and the point lies inside it (0.9 above), else a random point of the nearest svmap rebirth group of this map, else the current position (zeroed if outside the map). |
| [`bindPositionFound`](../../server/internal/world/gameini.go#L208)<br>`func bindPositionFound(mapID uint16, faction uint8, x, y, z float32) (uint16, float32, float32, float32, bool)` | bindPositionFound is bindPosition with GetBindPos's result: false when neither a configured point nor a rebirth group applies. |

<details><summary>Type definitions and owned fields</summary>

```go
type iniFile = config.GameINI
```

```go
type MapPos struct {
	MapID   uint16
	X, Y, Z float32
}
```

```go
type ZoneConfig struct {
	Dungeon bool // MapType D (F is a field)
	// Rebirth destinations are configured independently for each faction.
	RebirthLight, RebirthFury *MapPos
	MinLevel, MaxLevel        uint16
	WeatherType               uint8
	WeatherFrequency          time.Duration
	MapCurses                 [3]uint16
	WarType                   string
	CreateType                string // instanced maps (R, H, P)
	// CreateTime (day of week, start hour, hours) entries, up to 7 as
	// CMap::LoadConfig stores them (the ranking battle reads two).
	Schedule []grbSlot
	// CreateMinUser (party members to enter a P map), EnterMaxUser (an
	// instance's capacity, 30 by default) and ExpireTime (minutes; an empty
	// instance goes ExpireTime x 60 s + 30 s after it first empties).
	CreateMinUser, EnterMaxUser int
	Expire                      time.Duration
}
```

```go
type MapConfig struct {
	// Start is the new-character position per race (Chars.Family): human,
	// elf, vail, deatheater.
	Start [4]MapPos
	Zones map[uint16]ZoneConfig
}
```

</details>

### gm.go

[server/internal/world/gm.go](../../server/internal/world/gm.go#L1)

Direct GX dependencies: `internal/gm`, `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`gxGMState`](../../server/internal/world/gm.go#L16)<br>`const gxGMState = 0x7F01` | See the source definition and the subsystem ownership contract. |
| [`gmMaxSpawn`](../../server/internal/world/gm.go#L17)<br>`const gmMaxSpawn = 100` | See the source definition and the subsystem ownership contract. |
| [`gxGMModes`](../../server/internal/world/gm.go#L18)<br>`const gxGMModes = 0x7F29` | See the source definition and the subsystem ownership contract. |
| [`turboActive`](../../server/internal/world/gm.go#L20)<br>`func (s *Session) turboActive() bool` | See the source definition and the subsystem ownership contract. |
| [`gmReply`](../../server/internal/world/gm.go#L22)<br>`func (s *Session) gmReply(message string)` | See the source definition and the subsystem ownership contract. |
| [`sendGMModes`](../../server/internal/world/gm.go#L43)<br>`func (s *Session) sendGMModes()` | sendGMModes publishes the three GM switches to the GM's own client. |
| [`gmModeGuard`](../../server/internal/world/gm.go#L68)<br>`func (s *Session) gmModeGuard() bool` | setGMChar and setGMAttack are the original admin commands (CUser:: PacketAdminCmdE 0xFA01/0xFA02 bVisible and 0xFA05/0xFA06 bAttackable, chat "/char on&#124;off" and "/attack on&#124;off" at Game.exe 0x471c20). "/char on" makes the character transparent; "/attack on" makes it unattackable. gmModeGuard rejects a /char or /attack change that follows another within 750 ms. Each change re-evaluates who sees the GM, so a client that loops (or an input burst) must not be able to flood the world with them. |
| [`setGMChar`](../../server/internal/world/gm.go#L78)<br>`func (s *Session) setGMChar(on bool)` | See the source definition and the subsystem ownership contract. |
| [`setGMAttack`](../../server/internal/world/gm.go#L89)<br>`func (s *Session) setGMAttack(on bool)` | See the source definition and the subsystem ownership contract. |
| [`gmNumbers`](../../server/internal/world/gm.go#L103)<br>`func gmNumbers(args []string, bounds ...uint64) ([]uint64, error)` | See the source definition and the subsystem ownership contract. |
| [`handleGMCommand`](../../server/internal/world/gm.go#L119)<br>`func (s *Session) handleGMCommand(text string) bool` | Known commands are consumed even when refused; they never become public chat. |
| [`gmClearInventory`](../../server/internal/world/gm.go#L249)<br>`func (s *Session) gmClearInventory() error` | Clear carried inventory only. Equipped items and warehouse contents survive. |
| [`planGMGrant`](../../server/internal/world/gm.go#L275)<br>`func planGMGrant(items []db.Item, def *db.ItemDef, quantity int) ([]db.Item, []db.Item, []db.Item, error)` | planGMGrant keeps the entire request atomic, including quantities above 255. |
| [`gmGiveItem`](../../server/internal/world/gm.go#L307)<br>`func (s *Session) gmGiveItem(itemType, id uint8, quantity int) error` | See the source definition and the subsystem ownership contract. |
| [`gmSpawnPositions`](../../server/internal/world/gm.go#L332)<br>`func gmSpawnPositions(mapID uint16, x, z float32, count int, entities *MapEntities) ([][2]float32, error)` | Spawn positions form expanding rings; occupied points and blocked cells are rejected, including entities created by earlier commands. |
| [`gmSpawn`](../../server/internal/world/gm.go#L381)<br>`func (s *Session) gmSpawn(id uint16, count int, npcType uint8) error` | See the source definition and the subsystem ownership contract. |
| [`gmMove`](../../server/internal/world/gm.go#L458)<br>`func (s *Session) gmMove(mapID uint16) error` | See the source definition and the subsystem ownership contract. |
| [`gmSetUsage`](../../server/internal/world/gm.go#L542)<br>`const gmSetUsage = "usage: /set money&#124;kill&#124;str&#124;dex&#124;rec&#124;int&#124;wis&#124;luc&#124;statpoint&#124;skillpoint&#124;level amount"` | See the source definition and the subsystem ownership contract. |
| [`gmSet`](../../server/internal/world/gm.go#L545)<br>`func (s *Session) gmSet(args []string) error` | gmSet changes only the issuing character. Claimed rewards remain claimed. |
| [`gmCure`](../../server/internal/world/gm.go#L634)<br>`func (s *Session) gmCure() error` | gmCure clears all debuffs and cooldowns and restores HP, MP and SP. |

### gm_erase.go

[server/internal/world/gm_erase.go](../../server/internal/world/gm_erase.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`gmMaxEraseRange`](../../server/internal/world/gm_erase.go#L11)<br>`const gmMaxEraseRange = 1000` | gmMaxEraseRange bounds /mera <range>, in world units. |
| [`gmTargetID`](../../server/internal/world/gm_erase.go#L14)<br>`func gmTargetID(args []string) (uint32, bool, error)` | gmTargetID reads the "#<id>" the client writes in place of "t". |
| [`MobIDsWithin`](../../server/internal/world/gm_erase.go#L26)<br>`func MobIDsWithin(mapID uint16, x, z, radius float32) []uint32` | MobIDsWithin lists the living mobs of a map within radius units of (x, z). |
| [`eraseMob`](../../server/internal/world/gm_erase.go#L48)<br>`func (s *Session) eraseMob(id uint32) bool` | eraseMob kills a mob with no credit, drop or experience; its common respawn still follows. False when the mob cannot be killed (already dead, resetting home or scripted invulnerable). |
| [`gmEraseMobs`](../../server/internal/world/gm_erase.go#L63)<br>`func (s *Session) gmEraseMobs(args []string) error` | gmEraseMobs is /mera on the selected target (the client sends #<id>) or /mera <range>. |
| [`gmEraseNPC`](../../server/internal/world/gm_erase.go#L92)<br>`func (s *Session) gmEraseNPC(args []string) error` | gmEraseNPC is /nera t: the NPC leaves its map until the server restarts. |
| [`gmKill`](../../server/internal/world/gm_erase.go#L124)<br>`func (s *Session) gmKill(args []string) error` | gmKill is /kill t: a legitimate, immediate kill of a monster or a player of any faction. A monster drops and rewards its killer as usual. |
| [`applyExecute`](../../server/internal/world/gm_erase.go#L157)<br>`func (s *Session) applyExecute(h pendingUserHit) bool` | applyExecute is the target side of /kill on a player: all its health is lost whatever its immunity, and the death is credited to the GM like any player kill. |

### gm_level.go

[server/internal/world/gm_level.go](../../server/internal/world/gm_level.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`gmSetLevel`](../../server/internal/world/gm_level.go#L10)<br>`func (s *Session) gmSetLevel(args []string) error` | See the source definition and the subsystem ownership contract. |

### gm_map.go

[server/internal/world/gm_map.go](../../server/internal/world/gm_map.go#L1)

Direct GX dependencies: `internal/mapgeometry`, `internal/mapnav`, `internal/proto`.

| Declaration | Responsibility / source comment |
|---|---|
| [`handleGMMapTeleport`](../../server/internal/world/gm_map.go#L12)<br>`func (s *Session) handleGMMapTeleport(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`gmMapTeleport`](../../server/internal/world/gm_map.go#L33)<br>`func (s *Session) gmMapTeleport(mapID uint16, u, v float32) error` | See the source definition and the subsystem ownership contract. |
| [`exactMapDestination`](../../server/internal/world/gm_map.go#L73)<br>`func exactMapDestination(g *mapgeometry.Geometry, x, hint, z float32) (mapgeometry.Position, bool)` | See the source definition and the subsystem ownership contract. |
| [`nearbyMapDestination`](../../server/internal/world/gm_map.go#L90)<br>`func nearbyMapDestination(nav *mapnav.Map, g *mapgeometry.Geometry, mapID uint16, size, x, hint, z float32) (mapgeometry.Position, bool)` | Prefer the exact click, then the nearest navmesh projection. If its player body is obstructed, search bounded expanding rings; never accept a floorless point, a blocked mask cell, or coordinates outside the destination map. |

### gm_travel.go

[server/internal/world/gm_travel.go](../../server/internal/world/gm_travel.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`forcedTeleport`](../../server/internal/world/gm_travel.go#L8)<br>`type forcedTeleport` | forcedTeleport is a GM's /asummon: where the user is taken, at once. |
| [`queueForcedTeleport`](../../server/internal/world/gm_travel.go#L13)<br>`func (p *ConnectedPlayer) queueForcedTeleport(f forcedTeleport)` | See the source definition and the subsystem ownership contract. |
| [`takeForcedTeleport`](../../server/internal/world/gm_travel.go#L19)<br>`func (p *ConnectedPlayer) takeForcedTeleport() *forcedTeleport` | See the source definition and the subsystem ownership contract. |
| [`tickForcedTeleport`](../../server/internal/world/gm_travel.go#L29)<br>`func (s *Session) tickForcedTeleport()` | tickForcedTeleport carries out a GM's /asummon on this user: no consumable, no cast and no confirmation. |
| [`gmTarget`](../../server/internal/world/gm_travel.go#L51)<br>`func (s *Session) gmTarget(name string, mustBeAlive bool) (*ConnectedPlayer, error)` | gmTarget finds the character a GM command names. |
| [`gmGoTo`](../../server/internal/world/gm_travel.go#L64)<br>`func (s *Session) gmGoTo(name string) error` | gmGoTo is /amove: the movement rune without the item or the cast, to where another character is (dead or alive), on any map. |
| [`gmBring`](../../server/internal/world/gm_travel.go#L88)<br>`func (s *Session) gmBring(name string) error` | gmBring is /asummon: the summon rune without the item, the cast or the confirmation; the character is taken to the GM at once. |

<details><summary>Type definitions and owned fields</summary>

```go
type forcedTeleport struct {
	mapID uint16
	pos   [3]float32
}
```

</details>

### grb.go

[server/internal/world/grb.go](../../server/internal/world/grb.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/logx`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`grbCreateType`](../../server/internal/world/grb.go#L26)<br>`const grbCreateType = "R"` | See the source definition and the subsystem ownership contract. |
| [`grbRankDelay`](../../server/internal/world/grb.go#L27)<br>`const grbRankDelay = 30 * time.Minute` | CWorldSendThread::Run +0x1b7740 |
| [`grbNoticeOpen`](../../server/internal/world/grb.go#L28)<br>`const grbNoticeOpen = 100` | See the source definition and the subsystem ownership contract. |
| [`grbNoticeEnd10`](../../server/internal/world/grb.go#L29)<br>`const grbNoticeEnd10 = 101` | See the source definition and the subsystem ownership contract. |
| [`grbNoticeEnd1`](../../server/internal/world/grb.go#L30)<br>`const grbNoticeEnd1 = 102` | See the source definition and the subsystem ownership contract. |
| [`grbUnranked`](../../server/internal/world/grb.go#L31)<br>`const grbUnranked = 31` | See the source definition and the subsystem ownership contract. |
| [`grbRankedGuilds`](../../server/internal/world/grb.go#L32)<br>`const grbRankedGuilds = 30` | See the source definition and the subsystem ownership contract. |
| [`grbNow`](../../server/internal/world/grb.go#L36)<br>`var grbNow` | grbNow is the clock the battle entry checks (tests replace it). |
| [`grbSlots`](../../server/internal/world/grb.go#L39)<br>`func grbSlots(z ZoneConfig) []grbSlot` | grbSlots is the two CreateTime entries IsGuildRankMapOpenTime reads. |
| [`grbSlot`](../../server/internal/world/grb.go#L42)<br>`type grbSlot` | grbSlot is one Map.ini CreateTime entry. |
| [`end`](../../server/internal/world/grb.go#L46)<br>`func (s grbSlot) end() int` | slotHours is the slot's start hour and end hour (past 24 runs into the next day, as the original's wrap). |
| [`nextDay`](../../server/internal/world/grb.go#L48)<br>`func nextDay(d int) int` | See the source definition and the subsystem ownership contract. |
| [`grbOpen`](../../server/internal/world/grb.go#L56)<br>`func grbOpen(slots []grbSlot, now time.Time) bool` | grbOpen is IsGuildRankMapOpenTime. |
| [`grbNotice`](../../server/internal/world/grb.go#L73)<br>`func grbNotice(slots []grbSlot, now time.Time) (uint8, bool)` | grbNotice is GetGuildRankNotices: the start hour at 30 and 10 minutes before opening, 100 at the opening, 101 ten and 102 one minute before the close. |
| [`grbState`](../../server/internal/world/grb.go#L104)<br>`var grbState` | See the source definition and the subsystem ownership contract. |
| [`grbMaps`](../../server/internal/world/grb.go#L113)<br>`func grbMaps() []uint16` | See the source definition and the subsystem ownership contract. |
| [`enterGuildRanking`](../../server/internal/world/grb.go#L125)<br>`func (s *Session) enterGuildRanking(destination uint16) (zoneOut uint16, ok bool)` | enterGuildRanking resolves a teleport to a battle map: the guild's own instance while the battle is open, otherwise no zone (the move fails). |
| [`endGMGuildRanking`](../../server/internal/world/grb.go#L180)<br>`func (s *Session) endGMGuildRanking(final bool)` | endGMGuildRanking closes a GM's ranking participation when the GM leaves the battle map or the game, open or closed: the points earned are credited to the guild (kept for the regular settlement while a real battle still runs) and stored. There is no global notice. |
| [`grbKill`](../../server/internal/world/grb.go#L228)<br>`func grbKill(zone uint16, mob *MapMob, killerID uint32) bool` | grbKill is CMob::ItemDrop's map type 5 branch; true when the mob gave ranking points instead of gold. |
| [`TickGuildRanking`](../../server/internal/world/grb.go#L270)<br>`func TickGuildRanking(now time.Time, log *logx.Logger)` | TickGuildRanking runs the battle schedule: notices, the close and the ranking 30 minutes after it. |
| [`grbCloseInstances`](../../server/internal/world/grb.go#L311)<br>`func grbCloseInstances(mapID uint16, log *logx.Logger)` | grbCloseInstances is MoveAllBindPos at the close, then the empty instances are removed. |
| [`updateGuildRanking`](../../server/internal/world/grb.go#L348)<br>`func updateGuildRanking(log *logx.Logger)` | updateGuildRanking is CGuildMgr::UpdateGuildPointRank: the value is the sum of points from monster kills; guilds with a value are ranked per country by it (equal values: the first guild keeps the place, the others their rank, as the original's map insert), the top 30 get 1-30 and the rest 31. Every guild is stored and each country gets its 0x0D37 list. |

<details><summary>Type definitions and owned fields</summary>

```go
type grbSlot struct{ Day, Hour, Hours int }
```

</details>

### guild.go

[server/internal/world/guild.go](../../server/internal/world/guild.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/logx`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`guildRankMaster`](../../server/internal/world/guild.go#L22)<br>`const guildRankMaster = 1` | See the source definition and the subsystem ownership contract. |
| [`guildRankAdmin`](../../server/internal/world/guild.go#L23)<br>`const guildRankAdmin = 2` | See the source definition and the subsystem ownership contract. |
| [`guildRankMember`](../../server/internal/world/guild.go#L24)<br>`const guildRankMember = 9` | See the source definition and the subsystem ownership contract. |
| [`guildMaxAdmins`](../../server/internal/world/guild.go#L25)<br>`const guildMaxAdmins = 7` | See the source definition and the subsystem ownership contract. |
| [`guildInviteTime`](../../server/internal/world/guild.go#L26)<br>`const guildInviteTime = 30 * time.Second` | See the source definition and the subsystem ownership contract. |
| [`guildListBatch`](../../server/internal/world/guild.go#L27)<br>`const guildListBatch = 14` | 0x0D2F entries per packet |
| [`guildMemberBatch`](../../server/internal/world/guild.go#L28)<br>`const guildMemberBatch = 40` | 0x0D12/0x0D13/0x0D16 entries per packet |
| [`guildStatusDisband`](../../server/internal/world/guild.go#L32)<br>`const guildStatusDisband = 0x65` | 0x0D0C status modes; a level byte (1-80) is a character level change and 200 + rank a rank change. |
| [`guildStatusLeft`](../../server/internal/world/guild.go#L33)<br>`const guildStatusLeft = 0x66` | See the source definition and the subsystem ownership contract. |
| [`guildStatusExpel`](../../server/internal/world/guild.go#L34)<br>`const guildStatusExpel = 0x67` | See the source definition and the subsystem ownership contract. |
| [`guildStatusOnline`](../../server/internal/world/guild.go#L35)<br>`const guildStatusOnline = 0x68` | See the source definition and the subsystem ownership contract. |
| [`guildStatusOffline`](../../server/internal/world/guild.go#L36)<br>`const guildStatusOffline = 0x69` | See the source definition and the subsystem ownership contract. |
| [`guildStatusRank`](../../server/internal/world/guild.go#L37)<br>`const guildStatusRank = 200` | See the source definition and the subsystem ownership contract. |
| [`guildMember`](../../server/internal/world/guild.go#L40)<br>`type guildMember` | See the source definition and the subsystem ownership contract. |
| [`guild`](../../server/internal/world/guild.go#L49)<br>`type guild` | See the source definition and the subsystem ownership contract. |
| [`guildCreate`](../../server/internal/world/guild.go#L78)<br>`type guildCreate` | guildCreate is CGuildCreate: a creation (disband false) or disband vote. |
| [`playerGuildState`](../../server/internal/world/guild.go#L91)<br>`type playerGuildState` | Per-player guild state, guarded by guildMu. |
| [`guildMu`](../../server/internal/world/guild.go#L101)<br>`var guildMu sync.Mutex` | See the source definition and the subsystem ownership contract. |
| [`guildsByID`](../../server/internal/world/guild.go#L102)<br>`var guildsByID` | See the source definition and the subsystem ownership contract. |
| [`guildOfChar`](../../server/internal/world/guild.go#L103)<br>`var guildOfChar` | members and applicants |
| [`nextGuildID`](../../server/internal/world/guild.go#L104)<br>`var nextGuildID uint32` | See the source definition and the subsystem ownership contract. |
| [`guildsLoaded`](../../server/internal/world/guild.go#L105)<br>`var guildsLoaded bool` | See the source definition and the subsystem ownership contract. |
| [`LoadGuilds`](../../server/internal/world/guild.go#L109)<br>`func LoadGuilds(log *logx.Logger) error` | LoadGuilds is CGuildMgr::InitAddGuild / InitAddUser from the database. |
| [`admins`](../../server/internal/world/guild.go#L166)<br>`func (g *guild) admins() int` | See the source definition and the subsystem ownership contract. |
| [`send`](../../server/internal/world/guild.go#L177)<br>`func (g *guild) send(frame []byte, label string, except *ConnectedPlayer)` | send is CGuild::Send: every online member. |
| [`sendAdmin`](../../server/internal/world/guild.go#L186)<br>`func (g *guild) sendAdmin(frame []byte, label string)` | sendAdmin is CGuild::SendAdmin: the online master and sub-masters. |
| [`guildNameFor`](../../server/internal/world/guild.go#L194)<br>`func guildNameFor(charID uint32) string` | See the source definition and the subsystem ownership contract. |
| [`guildIDFor`](../../server/internal/world/guild.go#L203)<br>`func guildIDFor(charID uint32) uint32` | See the source definition and the subsystem ownership contract. |
| [`guildConnect`](../../server/internal/world/guild.go#L214)<br>`func (s *Session) guildConnect()` | guildConnect is CUser::InitGuild + CGuildMgr::ConnectUser + SendCharacterGuild at world entry, then CGuildMgr::SendGuildList. |
| [`guildDisconnect`](../../server/internal/world/guild.go#L251)<br>`func (s *Session) guildDisconnect()` | guildListPackets is CGuildMgr::SendGuildList: 0x0D0D, the country's guilds in 0x0D2F batches of 14, 0x0D0E. |
| [`guildLevelChanged`](../../server/internal/world/guild.go#L271)<br>`func (s *Session) guildLevelChanged()` | guildLevelChanged is CGuild::ChangeUserLevel on a level-up. |
| [`guildChat`](../../server/internal/world/guild.go#L283)<br>`func (s *Session) guildChat(message string, items proto.ChatItems)` | guildChat (0x1104) is CGuild::SendAllGuildChat: sender name, text. |
| [`handleGuild`](../../server/internal/world/guild.go#L297)<br>`func (s *Session) handleGuild(op packet.Opcode, p []byte)` | handleGuild dispatches the 0x0Dxx requests of CUser::PacketGuild. |
| [`fixedText`](../../server/internal/world/guild.go#L387)<br>`func fixedText(p []byte, at, n int) string` | See the source definition and the subsystem ownership contract. |
| [`guildDisbandRequest`](../../server/internal/world/guild.go#L402)<br>`func (s *Session) guildDisbandRequest()` | guildDisbandRequest (0x0D03) starts the disband vote. |
| [`guildDisbandAnswer`](../../server/internal/world/guild.go#L448)<br>`func (s *Session) guildDisbandAnswer(accept bool)` | guildDisbandAnswer (0x0D04): more than 2 acceptances disband it. |
| [`fire`](../../server/internal/world/guild.go#L470)<br>`func (g *guild) fire(log *logx.Logger, alreadyDeleted ...bool)` | fire is CGuildMgr::FireGuild / CGuild::Fire. |
| [`guildRemark`](../../server/internal/world/guild.go#L498)<br>`func (s *Session) guildRemark(p []byte)` | guildRemark (0x0D20) updates the notice for the entire guild. |

<details><summary>Type definitions and owned fields</summary>

```go
type guildMember struct {
	CharID uint32
	Name   string
	Job    uint8
	Level  uint16
	Rank   uint8
	player *ConnectedPlayer // nil when offline
}
```

```go
type guild struct {
	ID         uint32
	Name       string
	Master     string
	Remark     string
	Country    uint8
	Points     uint32
	Rank       uint8
	Etin       uint32
	BuyHouse   uint8
	UseHouse   uint8
	EtinReturn uint32
	KeepEtin   uint32
	// gmHouse: founded by a GM account. A test guild whose house is always
	// active, ranked and paid, with no GRB or etin upkeep.
	gmHouse bool
	// gmGRB: a GM owner is inside the guild's ranking instance outside the
	// normal schedule; points are credited when that GM leaves.
	gmGRB bool
	// Monster points earned in the current guild ranking battle.
	grbPoints  uint32
	gd         guildDuel // guarded by duelMu
	members    map[uint32]*guildMember
	applicants map[uint32]*guildMember
	bank       [guildBankSlots]*db.Item
	npcs       [guildNpcSlots]guildNpcSlot // CGuild +0x47c, guarded by guildMu
}
```

```go
type guildCreate struct {
	disband  bool
	name     string
	remark   string
	masterID uint32
	master   *ConnectedPlayer
	pending  map[uint32]bool // ids that have not accepted yet
	players  []*ConnectedPlayer
	accepted int
	guild    *guild // the guild being disbanded
}
```

```go
type playerGuildState struct {
	guild       *guild
	create      *guildCreate
	inviterID   uint32
	inviteUntil time.Time
	applied     *guild
	userID      string
}
```

</details>

### guild_creation.go

[server/internal/world/guild_creation.go](../../server/internal/world/guild_creation.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/logx`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`guildCreateRequest`](../../server/internal/world/guild_creation.go#L11)<br>`func (s *Session) guildCreateRequest(p []byte)` | guildCreateRequest (0x0D21) starts a solo foundation or a group vote. |
| [`sendAll`](../../server/internal/world/guild_creation.go#L81)<br>`func (c *guildCreate) sendAll(frame []byte, label string)` | See the source definition and the subsystem ownership contract. |
| [`fail`](../../server/internal/world/guild_creation.go#L88)<br>`func (c *guildCreate) fail(result uint8)` | fail is SendGuildCreateFailed / SendGuildFireFailed then ResetCreate. |
| [`reset`](../../server/internal/world/guild_creation.go#L97)<br>`func (c *guildCreate) reset()` | See the source definition and the subsystem ownership contract. |
| [`guildCreateAnswer`](../../server/internal/world/guild_creation.go#L106)<br>`func (s *Session) guildCreateAnswer(accept bool)` | guildCreateAnswer (0x0D22): the group must accept unanimously. |
| [`complete`](../../server/internal/world/guild_creation.go#L130)<br>`func (c *guildCreate) complete(log *logx.Logger, caller *Session) bool` | complete is CGuildMgr::CreateGuild 0x004321e0. |
| [`completePendingGuild`](../../server/internal/world/guild_creation.go#L204)<br>`func (s *Session) completePendingGuild()` | completePendingGuild retries a fully accepted group on the master's own session loop when its state lock was busy during another member's reply. |

### guild_membership.go

[server/internal/world/guild_membership.go](../../server/internal/world/guild_membership.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/logx`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`guildInvite`](../../server/internal/world/guild_membership.go#L10)<br>`func (s *Session) guildInvite(targetID uint32)` | See the source definition and the subsystem ownership contract. |
| [`guildInviteAnswer`](../../server/internal/world/guild_membership.go#L36)<br>`func (s *Session) guildInviteAnswer(accept bool)` | guildInviteAnswer (0x0D06): accept byte, answered to the inviter. |
| [`joinUser`](../../server/internal/world/guild_membership.go#L65)<br>`func (g *guild) joinUser(p *ConnectedPlayer, log *logx.Logger) bool` | joinUser is CGuild::JoinUser: an online user joins at rank 9. |
| [`guildApply`](../../server/internal/world/guild_membership.go#L93)<br>`func (s *Session) guildApply(guildID uint32)` | buildGuildMemberAdded (0x0D14): online byte and the 29-byte member. |
| [`cancelApplication`](../../server/internal/world/guild_membership.go#L122)<br>`func (g *guild) cancelApplication(p *ConnectedPlayer)` | cancelApplication is CGuild::JoinRequestCancle. |
| [`guildAnswerApplicant`](../../server/internal/world/guild_membership.go#L137)<br>`func (s *Session) guildAnswerApplicant(accept bool, charID uint32)` | guildAnswerApplicant (0x0D08): accept byte, char id; JoinOK / JoinNO. |
| [`guildLeave`](../../server/internal/world/guild_membership.go#L193)<br>`func (s *Session) guildLeave()` | guildLeave (0x0D09): CGuild::JoinLeave; the master leaves only alone, and an emptied guild is disbanded. |
| [`guildExpel`](../../server/internal/world/guild_membership.go#L231)<br>`func (s *Session) guildExpel(charID uint32)` | guildExpel (0x0D0A): CGuild::JoinExpel, only lower ranks. |
| [`guildChangeRank`](../../server/internal/world/guild_membership.go#L271)<br>`func (s *Session) guildChangeRank(direction uint8, charID uint32)` | guildChangeRank (0x0D0C): direction (0 promote, 1 demote), char id; CGuild::ChangeGuildUserLevel. |
| [`guildRankChange`](../../server/internal/world/guild_membership.go#L299)<br>`func guildRankChange(actor, target, direction uint8, admins int) (uint8, bool)` | See the source definition and the subsystem ownership contract. |

### guild_packets.go

[server/internal/world/guild_packets.go](../../server/internal/world/guild_packets.go#L1)

Direct GX dependencies: `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`guildStatus`](../../server/internal/world/guild_packets.go#L7)<br>`func guildStatus(mode uint8, charID uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`writeGuildMember`](../../server/internal/world/guild_packets.go#L14)<br>`func writeGuildMember(w *packet.Writer, m *guildMember)` | See the source definition and the subsystem ownership contract. |
| [`memberListPackets`](../../server/internal/world/guild_packets.go#L24)<br>`func (g *guild) memberListPackets() [][]byte` | buildGuildMemberLists is SendGuildUserList: 0x0D12 online and 0x0D13 offline members, 29 bytes each, 40 per packet, none sent when empty. |
| [`writeGuildApplicant`](../../server/internal/world/guild_packets.go#L57)<br>`func writeGuildApplicant(w *packet.Writer, m *guildMember)` | See the source definition and the subsystem ownership contract. |
| [`applicantListPackets`](../../server/internal/world/guild_packets.go#L65)<br>`func (g *guild) applicantListPackets() [][]byte` | applicantListPackets is SendGuildJoinList (0x0D16, 28 bytes each). |
| [`writeGuildInfo`](../../server/internal/world/guild_packets.go#L91)<br>`func writeGuildInfo(w *packet.Writer, g *guild)` | writeGuildInfo is the 120-byte guild record of 0x0D2F and 0x0D30. |
| [`guildSendCountry`](../../server/internal/world/guild_packets.go#L101)<br>`func guildSendCountry(country uint8, frame []byte, label string)` | sendWorld is CWorld::SendAll restricted to a country. |
| [`buildGuildCreateFailed`](../../server/internal/world/guild_packets.go#L109)<br>`func buildGuildCreateFailed(result uint8, name string) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildGuildCreateResult`](../../server/internal/world/guild_packets.go#L119)<br>`func buildGuildCreateResult(result uint8, id uint32, rank uint8, name, remark string) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildGuildJoined`](../../server/internal/world/guild_packets.go#L129)<br>`func buildGuildJoined(result uint8, g *guild, rank uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`guildResult`](../../server/internal/world/guild_packets.go#L138)<br>`func guildResult(op packet.Opcode, result uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`guildResultID`](../../server/internal/world/guild_packets.go#L144)<br>`func guildResultID(op packet.Opcode, result uint8, id uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`guildID32`](../../server/internal/world/guild_packets.go#L151)<br>`func guildID32(op packet.Opcode, id uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildGuildMemberAdded`](../../server/internal/world/guild_packets.go#L159)<br>`func buildGuildMemberAdded(m *guildMember) []byte` | guildNameFor / guildIDFor feed the character packets (0x0105, 0x0201, 0x0303). They must not be called with a player's mu held. |
| [`guildListPackets`](../../server/internal/world/guild_packets.go#L171)<br>`func guildListPackets(country uint8) [][]byte` | guildApply (0x0D07): guild id; CGuild::JoinRequest. |

### guildbank.go

[server/internal/world/guildbank.go](../../server/internal/world/guildbank.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`guildBankBag`](../../server/internal/world/guildbank.go#L16)<br>`const guildBankBag = 0xff` | See the source definition and the subsystem ownership contract. |
| [`guildBankSlots`](../../server/internal/world/guildbank.go#L17)<br>`const guildBankSlots = 240` | See the source definition and the subsystem ownership contract. |
| [`guildBankRanked`](../../server/internal/world/guildbank.go#L18)<br>`const guildBankRanked = 30` | CGuild +0x48 must be at most this |
| [`guildBankBatch`](../../server/internal/world/guildbank.go#L19)<br>`const guildBankBatch = 50` | 0x0D29 entries per packet |
| [`guildBankRefused`](../../server/internal/world/guildbank.go#L20)<br>`const guildBankRefused = 1` | 0x0D28 result: guild not ranked |
| [`guildBankNoRight`](../../server/internal/world/guildbank.go#L21)<br>`const guildBankNoRight = 2` | not an admin / slot beyond the permitted pages |
| [`guildBankMember9`](../../server/internal/world/guildbank.go#L22)<br>`const guildBankMember9 = 4` | rank 9 may not deposit |
| [`loadGuildBank`](../../server/internal/world/guildbank.go#L25)<br>`func loadGuildBank(g *guild) error` | See the source definition and the subsystem ownership contract. |
| [`bankListPackets`](../../server/internal/world/guildbank.go#L41)<br>`func (g *guild) bankListPackets() [][]byte` | bankListPackets is SendBankItemList: 0x0D29, 33-byte entries, 50 each. |
| [`buildGuildBankSlot`](../../server/internal/world/guildbank.go#L75)<br>`func buildGuildBankSlot(op packet.Opcode, slot uint8, it *db.Item, charID uint32) []byte` | buildGuildBankSlot (0x0D31 placed / 0x0D32 taken): slot, item, char id. |
| [`buildGuildBankResult`](../../server/internal/world/guildbank.go#L94)<br>`func buildGuildBankResult(result, arg uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`guildBankCheck`](../../server/internal/world/guildbank.go#L103)<br>`func (s *Session) guildBankCheck(g *guild, rank uint8, slot uint8, adminOnly bool) bool` | guildBankCheck is the shared gate: ranked guild, the rank rule and the permitted slot; it sends the 0x0D28 refusal itself. |
| [`handleGuildBankMove`](../../server/internal/world/guildbank.go#L118)<br>`func (s *Session) handleGuildBankMove(srcBag, srcSlot, destBag, destSlot uint8)` | handleGuildBankMove is the 0x0204 with bag 0xFF on either side. |
| [`guildBankDeposit`](../../server/internal/world/guildbank.go#L170)<br>`func (s *Session) guildBankDeposit(g *guild, rank, bag, slot, dest uint8)` | guildBankDeposit is ItemBagToGuildBank: not from the equipment, no quest items (real type 27) nor ReqOg items; a taken slot only for admins (swap). |
| [`guildBankWithdraw`](../../server/internal/world/guildbank.go#L211)<br>`func (s *Session) guildBankWithdraw(g *guild, rank, src, bag, slot uint8)` | guildBankWithdraw is ItemGuildBankToBag: admins only, 5% of Sell x count (truncated; type 94 free), a taken bag slot swaps into the bank. |
| [`guildBankRearrange`](../../server/internal/world/guildbank.go#L251)<br>`func (s *Session) guildBankRearrange(g *guild, rank, src, dest uint8)` | guildBankRearrange is ItemGuildBankToGuildBank: admins only; equal items merge up to the stack, otherwise the slots swap; every member sees the 0x0204. |

### guildduel.go

[server/internal/world/guildduel.go](../../server/internal/world/guildduel.go#L1)

Direct GX dependencies: `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`gduelNone`](../../server/internal/world/guildduel.go#L38)<br>`const gduelNone = 0` | See the source definition and the subsystem ownership contract. |
| [`gduelAsking`](../../server/internal/world/guildduel.go#L39)<br>`const gduelAsking = 1` | See the source definition and the subsystem ownership contract. |
| [`gduelAsked`](../../server/internal/world/guildduel.go#L40)<br>`const gduelAsked = 2` | See the source definition and the subsystem ownership contract. |
| [`gduelCountdown`](../../server/internal/world/guildduel.go#L41)<br>`const gduelCountdown = 3` | See the source definition and the subsystem ownership contract. |
| [`gduelFighting`](../../server/internal/world/guildduel.go#L42)<br>`const gduelFighting = 4` | See the source definition and the subsystem ownership contract. |
| [`guildDuel`](../../server/internal/world/guildduel.go#L46)<br>`type guildDuel` | guildDuel is CGuild +0x4b4..+0x4d4 (guarded by duelMu). |
| [`guildPlayers`](../../server/internal/world/guildduel.go#L58)<br>`func guildPlayers(g *guild) []*ConnectedPlayer` | guildPlayers snapshots a guild's online members. |
| [`guildAndRank`](../../server/internal/world/guildduel.go#L71)<br>`func guildAndRank(p *ConnectedPlayer) (*guild, uint8)` | guildAndRank is the user's guild and rank (nil, 0 without one). |
| [`guildFromID`](../../server/internal/world/guildduel.go#L85)<br>`func guildFromID(id uint32) *guild` | See the source definition and the subsystem ownership contract. |
| [`sendPvPLocked`](../../server/internal/world/guildduel.go#L92)<br>`func sendPvPLocked(players []*ConnectedPlayer, frame []byte, label string, joined bool)` | sendPvPLocked is CGuild::SendPvP / SendPvPJoinUser: the participants. |
| [`guildDuelFoes`](../../server/internal/world/guildduel.go#L101)<br>`func guildDuelFoes(a, b *ConnectedPlayer) bool` | guildDuelFoes is IsTeam's guild-duel case. |
| [`guildDuelSide`](../../server/internal/world/guildduel.go#L114)<br>`func guildDuelSide(a, b *ConnectedPlayer) bool` | guildDuelSide is IsTeamUsable's guild-duel check: the same side, or neither in one. |
| [`handleGuildDuel`](../../server/internal/world/guildduel.go#L120)<br>`func (s *Session) handleGuildDuel(op packet.Opcode, p []byte)` | See the source definition and the subsystem ownership contract. |
| [`guildDuelRequest`](../../server/internal/world/guildduel.go#L137)<br>`func (s *Session) guildDuelRequest(target uint32)` | See the source definition and the subsystem ownership contract. |
| [`guildDuelAnswer`](../../server/internal/world/guildduel.go#L167)<br>`func (s *Session) guildDuelAnswer(accept bool)` | See the source definition and the subsystem ownership contract. |
| [`tickGuildDuel`](../../server/internal/world/guildduel.go#L230)<br>`func (s *Session) tickGuildDuel()` | tickGuildDuel is PvPChkLeaveByTime's guild part, run by the leader. |
| [`guildDuelCancel`](../../server/internal/world/guildduel.go#L267)<br>`func guildDuelCancel(g, og *guild)` | guildDuelCancel ends a guild duel before the countdown (0x2502 3 to both leaders). |
| [`guildDuelLeave`](../../server/internal/world/guildduel.go#L289)<br>`func guildDuelLeave(p *ConnectedPlayer, g *guild, reason uint8)` | guildDuelLeave is the guild part of PvPChkLeaveByMove / ByDeath / ByZoneLeave / ByGuildLeave for p (g is p's guild, given when p has just left it). |
| [`pvpUserLeave`](../../server/internal/world/guildduel.go#L320)<br>`func pvpUserLeave(g, og *guild, p *ConnectedPlayer, reason uint8)` | pvpUserLeave is CGuild::PvPUserLeave. |
| [`guildDuelDeath`](../../server/internal/world/guildduel.go#L357)<br>`func (s *Session) guildDuelDeath(killer *ConnectedPlayer) bool` | guildDuelDeath is PvPGetDeathType 5: killed by a participant of the opposing guild; the leave runs and the death keeps the exp. |
| [`guildDuelCheckMove`](../../server/internal/world/guildduel.go#L369)<br>`func (s *Session) guildDuelCheckMove()` | guildDuelCheckMove is PvPChkLeaveByMove's guild part. |

<details><summary>Type definitions and owned fields</summary>

```go
type guildDuel struct {
	state     int
	leader    uint32 // +0x4c0, this guild's duel leader
	oppLeader uint32 // +0x4b8
	oppGuild  uint32 // +0x4bc
	center    [2]float32
	mapID     uint16
	count     int
	deadline  time.Time
}
```

</details>

### guildhouse.go

[server/internal/world/guildhouse.go](../../server/internal/world/guildhouse.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`guildNpcGate`](../../server/internal/world/guildhouse.go#L20)<br>`const guildNpcGate = 0x002` | Guild NPC types as stored in the eight CGuild NPC slots. |
| [`guildNpcWeapon`](../../server/internal/world/guildhouse.go#L21)<br>`const guildNpcWeapon = 0x101` | Guild NPC types as stored in the eight CGuild NPC slots. |
| [`guildNpcDefense`](../../server/internal/world/guildhouse.go#L22)<br>`const guildNpcDefense = 0x201` | Guild NPC types as stored in the eight CGuild NPC slots. |
| [`guildNpcGeneral`](../../server/internal/world/guildhouse.go#L23)<br>`const guildNpcGeneral = 0x301` | Guild NPC types as stored in the eight CGuild NPC slots. |
| [`guildNpcSmith`](../../server/internal/world/guildhouse.go#L24)<br>`const guildNpcSmith = 0x003` | Guild NPC types as stored in the eight CGuild NPC slots. |
| [`guildNpcPotion`](../../server/internal/world/guildhouse.go#L25)<br>`const guildNpcPotion = 0x401` | Guild NPC types as stored in the eight CGuild NPC slots. |
| [`guildNpcBank`](../../server/internal/world/guildhouse.go#L26)<br>`const guildNpcBank = 0x006` | Guild NPC types as stored in the eight CGuild NPC slots. |
| [`guildNpcSlots`](../../server/internal/world/guildhouse.go#L28)<br>`const guildNpcSlots = 8` | Guild NPC types as stored in the eight CGuild NPC slots. |
| [`guildNpcServices`](../../server/internal/world/guildhouse.go#L29)<br>`const guildNpcServices = 7` | Guild NPC types as stored in the eight CGuild NPC slots. |
| [`guildEtinSpecial`](../../server/internal/world/guildhouse.go#L30)<br>`const guildEtinSpecial = 52` | Guild NPC types as stored in the eight CGuild NPC slots. Special of the Type 43 etin items: 52..55 are x1, x10, x100, x1000 |
| [`guildEtinType`](../../server/internal/world/guildhouse.go#L31)<br>`const guildEtinType = 43` | Guild NPC types as stored in the eight CGuild NPC slots. '+', the item type CUser::PacketGuild 0x0D35 accepts |
| [`guildNpcOK`](../../server/internal/world/guildhouse.go#L36)<br>`const guildNpcOK = 0` | Results of 0x0D25 (CGuild::NpcUpgrade). |
| [`guildNpcNoEtin`](../../server/internal/world/guildhouse.go#L37)<br>`const guildNpcNoEtin = 1` | Results of 0x0D25 (CGuild::NpcUpgrade). |
| [`guildNpcBadLevel`](../../server/internal/world/guildhouse.go#L38)<br>`const guildNpcBadLevel = 2` | Results of 0x0D25 (CGuild::NpcUpgrade). |
| [`guildNpcRefused`](../../server/internal/world/guildhouse.go#L39)<br>`const guildNpcRefused = 3` | Results of 0x0D25 (CGuild::NpcUpgrade). |
| [`guildHouseBought`](../../server/internal/world/guildhouse.go#L44)<br>`const guildHouseBought = 0` | Results of 0x0D34. |
| [`guildHouseOwned`](../../server/internal/world/guildhouse.go#L45)<br>`const guildHouseOwned = 1` | Results of 0x0D34. |
| [`guildHouseNoMoney`](../../server/internal/world/guildhouse.go#L46)<br>`const guildHouseNoMoney = 2` | Results of 0x0D34. |
| [`guildHouseNoRight`](../../server/internal/world/guildhouse.go#L47)<br>`const guildHouseNoRight = 3` | Results of 0x0D34. |
| [`guildNpcSlot`](../../server/internal/world/guildhouse.go#L51)<br>`type guildNpcSlot` | guildNpcSlot is a CGuild NPC slot (+0x47c, type and level words). |
| [`guildServiceTable`](../../server/internal/world/guildhouse.go#L55)<br>`type guildServiceTable` | guildServiceTable is one g_GuildNpcUpgradeEtin / g_GuildNpcKeepEtin row: index is the NPC level (1-5). |
| [`guildHouseSettings`](../../server/internal/world/guildhouse.go#L63)<br>`type guildHouseSettings` | guildHouseSettings is everything InitGuildHouseIniFile loads, starting from the executable's own defaults. |
| [`defaultGuildHouse`](../../server/internal/world/guildhouse.go#L76)<br>`func defaultGuildHouse() guildHouseSettings` | See the source definition and the subsystem ownership contract. |
| [`guildHouse`](../../server/internal/world/guildhouse.go#L97)<br>`var guildHouse` | guildHouse is the active configuration. |
| [`iniTokens`](../../server/internal/world/guildhouse.go#L101)<br>`func iniTokens(value string) []string` | iniTokens splits a value like ParseString1/ParseString2: the first 254 characters on " ,=()[]{}" and white space. |
| [`iniInt`](../../server/internal/world/guildhouse.go#L109)<br>`func iniInt(token string) int64` | iniUint is the atoi the original parser applies: a decimal prefix. |
| [`parseGuildHouse`](../../server/internal/world/guildhouse.go#L124)<br>`func parseGuildHouse(ini iniFile) guildHouseSettings` | parseGuildHouse applies GuildHouse.ini over the built-in defaults. Missing sections, levels or fields keep their default. Each service section lists its levels as `LevelN = Field value, Field value`; see the file itself. |
| [`loadGuildHouseIni`](../../server/internal/world/guildhouse.go#L211)<br>`func loadGuildHouseIni()` | See the source definition and the subsystem ownership contract. |
| [`service`](../../server/internal/world/guildhouse.go#L221)<br>`func (h *guildHouseSettings) service(npcType uint16) *guildServiceTable` | service is the upgrade table of an NPC type (nil for an unknown type). |
| [`npcLevel`](../../server/internal/world/guildhouse.go#L231)<br>`func (g *guild) npcLevel(npcType uint16) uint16` | npcLevel is the level of the guild's NPC of that type, 0 without it. |
| [`houseEnabled`](../../server/internal/world/guildhouse.go#L242)<br>`func (g *guild) houseEnabled() bool` | houseEnabled is CGuild::EnableGuildHouse without the ranking condition: the house is active and bought. |
| [`forceGMHouse`](../../server/internal/world/guildhouse.go#L248)<br>`func (g *guild) forceGMHouse()` | forceGMHouse keeps a GM-founded guild's house open: bought, active and free of upkeep. |
| [`calculateKeepEtin`](../../server/internal/world/guildhouse.go#L256)<br>`func (g *guild) calculateKeepEtin()` | calculateKeepEtin is CGuild::CalculateKeepEtin: the default plus the upkeep of every NPC at its level. |
| [`bankSlots`](../../server/internal/world/guildhouse.go#L271)<br>`func (g *guild) bankSlots() int` | bankSlots is GetGuildHouseBankAllowSlotCount: PermitSlot at the bank NPC level (the first entry without it). |
| [`row`](../../server/internal/world/guildhouse.go#L279)<br>`func (g *guild) row() db.GuildRow` | See the source definition and the subsystem ownership contract. |
| [`npcListPacket`](../../server/internal/world/guildhouse.go#L285)<br>`func (g *guild) npcListPacket() []byte` | npcListPacket is CGuild::SendNpcLvList (0x0D24); nil without NPCs. |
| [`buildGuildNpcResult`](../../server/internal/world/guildhouse.go#L306)<br>`func buildGuildNpcResult(result uint8, npcType uint16, level, slot uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`guildNpcUpgrade`](../../server/internal/world/guildhouse.go#L317)<br>`func (s *Session) guildNpcUpgrade(g *guild, npcType uint16, level uint8)` | guildNpcUpgrade is 0x0D25 (CGuild::NpcUpgrade), reached for guild ranks 1 and 2 only. Called with guildMu held. |
| [`buildGuildHouseBuy`](../../server/internal/world/guildhouse.go#L400)<br>`func buildGuildHouseBuy(result uint8, money uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`guildHouseBuy`](../../server/internal/world/guildhouse.go#L409)<br>`func (s *Session) guildHouseBuy(g *guild, rank uint8)` | guildHouseBuy is the 0x0D34 branch of CUser::PacketGuild. Called with guildMu held. |
| [`guildEtinValue`](../../server/internal/world/guildhouse.go#L440)<br>`func guildEtinValue(it *db.Item) (value uint32, ok bool)` | guildEtinValue is the etin an item donates: Type 43 with Special 52..55. |
| [`guildDonate`](../../server/internal/world/guildhouse.go#L464)<br>`func (s *Session) guildDonate(g *guild)` | guildDonate is the 0x0D35 branch: every etin item in the bags is consumed, its value added to the guild's etin and the return count incremented. Called with guildMu held. |
| [`settleGuildHouse`](../../server/internal/world/guildhouse.go#L507)<br>`func (g *guild) settleGuildHouse()` | settleGuildHouse is the house part of CGuildMgr::UpdateGuildPointRank for a ranked guild: below the upkeep the house is suspended (UseHouse 2), otherwise the upkeep is paid from the etin. Called with guildMu held. |
| [`inGuildHouseZone`](../../server/internal/world/guildhouse.go#L525)<br>`func (s *Session) inGuildHouseZone() bool` | inGuildHouseZone is CMap::IsGuildHouseMap for the character's zone. |
| [`houseNpcRefusal`](../../server/internal/world/guildhouse.go#L531)<br>`func (s *Session) houseNpcRefusal(result, detail uint8)` | houseNpcRefusal answers a guild house service the guild cannot use with 0x0D28 {result, detail}, as PacketShop and PacketMaininterface do. |
| [`shopBuyRate`](../../server/internal/world/guildhouse.go#L539)<br>`func (g *guild) shopBuyRate(shopType uint8) (rate float64, result, detail uint8)` | shopBuyRate is the price factor of a shop type (1 weapon, 2 defense, 3 general, 4 potion) at the guild's level of that service, or the refusal result and detail. Ranking no longer restricts any service, and NPC ids no longer matter: the discount belongs to the kind of shop. |
| [`gateRate`](../../server/internal/world/guildhouse.go#L550)<br>`func (g *guild) gateRate() float64` | gateRate is the teleport fee factor of the guild's gatekeeper level. The upgrades only change this fee; every gatekeeper stays usable. |
| [`houseRound`](../../server/internal/world/guildhouse.go#L555)<br>`func houseRound(v float64) uint32` | roundHalfAway is the x87 ROUND of the original price products. |
| [`smithIndex`](../../server/internal/world/guildhouse.go#L558)<br>`func (g *guild) smithIndex() int` | smithIndex is the blacksmith NPC level clamped to the six-entry tables. |
| [`houseMix`](../../server/internal/world/guildhouse.go#L563)<br>`func (s *Session) houseMix() (rate float64, bonus int)` | houseMix returns the gem-linking cost factor (RapiceMixDecreRate) and the success bonus (RapiceMixPercentRate, active house only) for the guild smith inside a house map; (1, 0) elsewhere. |
| [`houseShopRate`](../../server/internal/world/guildhouse.go#L584)<br>`func (s *Session) houseShopRate(shopType uint8) (g *guild, rate float64, result, detail uint8)` | houseShopRate resolves the guild's shop factor under guildMu, released on every path. A nil guild means the buyer has none. |
| [`houseStatusPacket`](../../server/internal/world/guildhouse.go#L603)<br>`func (g *guild) houseStatusPacket() []byte` | houseStatusPacket is the GX-only 0x0D33 notice (the EP5.4 client has no reader for it), sent to a guild's members only while the house is bought: the keep etin (bit 31 marks a GM guild) and, as whole percentages, the discounts the guild's current service levels give. The client shows them in tooltips and greys its Buy House option. |

<details><summary>Type definitions and owned fields</summary>

```go
type guildNpcSlot struct{ Type, Level uint16 }
```

```go
type guildServiceTable struct {
	Type uint16
	Cost [7]uint16
	Keep [6]uint16
}
```

```go
type guildHouseSettings struct {
	CreateMoney, BuyMoney, DefaultKeep uint32
	Services                           [guildNpcServices]guildServiceTable // gate, weapon, defense, general, smith, potion, bank
	// ShopFactor is the price factor of a shop type (1 weapon, 2 defense, 3
	// general, 4 potion) by the guild's level of that service (0-3).
	ShopFactor [5][4]float64
	GateRate   [6]float64 // teleport fee factor by gatekeeper level
	RepairRate [6]float64 // smith repair factor by level
	MixRate    [6]float64 // gem link cost factor by smith level
	MixPercent [6]int32   // gem link success bonus by smith level
	BankSlots  [6]uint16  // usable warehouse slots by bank level
}
```

</details>

### healthassist.go

[server/internal/world/healthassist.go](../../server/internal/world/healthassist.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`detailHealthAssist`](../../server/internal/world/healthassist.go#L3)<br>`const detailHealthAssist = 87` | buff: SP and/or MP take the damage before HP |
| [`healthAssistLevel`](../../server/internal/world/healthassist.go#L7)<br>`func (s *Session) healthAssistLevel() uint8` | healthAssistLevel is CUser +0x1338: the level of the active Health Assistant (AddApplySkillBuff case 0x57), 0 without it. |
| [`absorbHealthAssist`](../../server/internal/world/healthassist.go#L21)<br>`func (s *Session) absorbHealthAssist(hp uint32) uint32` | absorbHealthAssist is the Health Assistant part of CUser::Damage (0x0044c259): level 1 spends SP against the hit, level 2 MP, level 3 SP and then MP, point for point; what is left is the HP damage. Other levels do nothing. The amounts really taken are left in s.assisted (HP, SP, MP) so the result packets can show them. |

### instance.go

[server/internal/world/instance.go](../../server/internal/world/instance.go#L1)

Direct GX dependencies: `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`instanceZoneBase`](../../server/internal/world/instance.go#L17)<br>`const instanceZoneBase = 1000` | See the source definition and the subsystem ownership contract. |
| [`realMap`](../../server/internal/world/instance.go#L20)<br>`func realMap(zone uint16) uint16` | realMap is the map a zone shows (itself for real maps). |
| [`isInstanceZone`](../../server/internal/world/instance.go#L32)<br>`func isInstanceZone(zone uint16) bool` | See the source definition and the subsystem ownership contract. |
| [`init`](../../server/internal/world/instance.go#L34)<br>`func init()` | See the source definition and the subsystem ownership contract. |
| [`mapCreateType`](../../server/internal/world/instance.go#L37)<br>`func mapCreateType(zone uint16) string` | mapCreateType is the Map.ini CreateType of a zone's map ("" if none). |
| [`createInstance`](../../server/internal/world/instance.go#L45)<br>`func createInstance(mapID uint16, owner uint32) (uint16, error)` | createInstance builds a fresh copy of mapID from its svmap. |
| [`destroyInstance`](../../server/internal/world/instance.go#L91)<br>`func destroyInstance(zone uint16)` | destroyInstance drops an (empty) instance's entities and grid. |
| [`instanceZones`](../../server/internal/world/instance.go#L112)<br>`func instanceZones(mapID uint16) []uint16` | instanceZones lists the live instances of mapID. |
| [`instanceOwnerOf`](../../server/internal/world/instance.go#L125)<br>`func instanceOwnerOf(zone uint16) uint32` | instanceOwnerOf is the owner id (guild, party) an instance was created for. |

### inventory.go

[server/internal/world/inventory.go](../../server/internal/world/inventory.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`inventoryChunk`](../../server/internal/world/inventory.go#L13)<br>`const inventoryChunk = 50` | See the source definition and the subsystem ownership contract. |
| [`buildInventoryPackets`](../../server/internal/world/inventory.go#L15)<br>`func buildInventoryPackets(items []db.Item) [][]byte` | See the source definition and the subsystem ownership contract. |
| [`isValidItemPosition`](../../server/internal/world/inventory.go#L40)<br>`func isValidItemPosition(bag, slot uint8) bool` | See the source definition and the subsystem ownership contract. |
| [`findItem`](../../server/internal/world/inventory.go#L42)<br>`func findItem(items []db.Item, bag, slot uint8) int` | See the source definition and the subsystem ownership contract. |
| [`findItemByUID`](../../server/internal/world/inventory.go#L51)<br>`func findItemByUID(items []db.Item, uid int64) int` | See the source definition and the subsystem ownership contract. |
| [`writeSlot`](../../server/internal/world/inventory.go#L60)<br>`func writeSlot(w *packet.Writer, bag, slot uint8, it *db.Item)` | See the source definition and the subsystem ownership contract. |
| [`buildBagToBag`](../../server/internal/world/inventory.go#L79)<br>`func buildBagToBag(srcBag, srcSlot uint8, src *db.Item, destBag, destSlot uint8, dest *db.Item) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildAddItem`](../../server/internal/world/inventory.go#L86)<br>`func buildAddItem(it *db.Item) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildSetMoney`](../../server/internal/world/inventory.go#L92)<br>`func buildSetMoney(money uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildAddItemError`](../../server/internal/world/inventory.go#L98)<br>`func buildAddItemError(code uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildUseItemResult`](../../server/internal/world/inventory.go#L105)<br>`func buildUseItemResult(characterID uint32, it *db.Item) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildRemoveItem`](../../server/internal/world/inventory.go#L118)<br>`func buildRemoveItem(it *db.Item, full bool) []byte` | See the source definition and the subsystem ownership contract. |
| [`freeInventorySlots`](../../server/internal/world/inventory.go#L132)<br>`func freeInventorySlots(items []db.Item) [][2]uint8` | See the source definition and the subsystem ownership contract. |
| [`itemsMatch`](../../server/internal/world/inventory.go#L144)<br>`func itemsMatch(a, b *db.Item) bool` | See the source definition and the subsystem ownership contract. |
| [`itemsCanStack`](../../server/internal/world/inventory.go#L149)<br>`func itemsCanStack(src, dest *db.Item) bool` | itemsCanStack: same item, and the destination is not equipped. |
| [`maxStackOf`](../../server/internal/world/inventory.go#L151)<br>`func maxStackOf(t, id uint8) uint16` | See the source definition and the subsystem ownership contract. |
| [`findStackable`](../../server/internal/world/inventory.go#L158)<br>`func findStackable(items []db.Item, tpl *db.Item, maxStack uint16) int` | See the source definition and the subsystem ownership contract. |
| [`wouldFit`](../../server/internal/world/inventory.go#L168)<br>`func wouldFit(items []db.Item, tpl *db.Item, count, maxStack uint16) bool` | wouldFit simulates giveItem on a copy of the inventory. |
| [`giveItem`](../../server/internal/world/inventory.go#L192)<br>`func (s *Session) giveItem(tpl *db.Item, count, maxStack uint16) ([][]byte, uint16, error)` | giveItem stacks onto existing items, then fills free slots, persisting each change. Returns the 0x0205 packets and what did not fit. |
| [`giveItemTo`](../../server/internal/world/inventory.go#L200)<br>`func giveItemTo(charID uint32, items *[]db.Item, tpl *db.Item, count, maxStack uint16) ([][]byte, uint16, error)` | See the source definition and the subsystem ownership contract. |
| [`consumeOne`](../../server/internal/world/inventory.go#L215)<br>`func (s *Session) consumeOne(i int) (db.Item, error)` | consumeOne decrements the item at index i by one, persisting it. |
| [`cooldownForReqIg`](../../server/internal/world/inventory.go#L246)<br>`func cooldownForReqIg(reqIg uint8) time.Duration` | See the source definition and the subsystem ownership contract. |
| [`handleUseItem`](../../server/internal/world/inventory.go#L252)<br>`func (s *Session) handleUseItem(p []byte)` | handleUseItem (0x050A): bag, slot. |
| [`healDelta`](../../server/internal/world/inventory.go#L441)<br>`func (s *Session) healDelta(hp, mp, sp uint32) (uint32, uint32, uint32)` | healDelta raises effective HP/MP/SP by up to the given amounts, applied to the durable absolute fields. Returns the actual amounts gained. |
| [`spendDelta`](../../server/internal/world/inventory.go#L452)<br>`func (s *Session) spendDelta(mp, sp uint32)` | spendDelta lowers effective MP/SP by the given amounts (as deltas). |
| [`handleChangeAppearance`](../../server/internal/world/inventory.go#L460)<br>`func (s *Session) handleChangeAppearance(p []byte)` | handleChangeAppearance (0x0226): bag, slot, hair, face, height, sex. |
| [`genderRestrictedEquipment`](../../server/internal/world/inventory.go#L508)<br>`func (s *Session) genderRestrictedEquipment() bool` | genderRestrictedEquipment follows CUser::PacketMaininterface case 0x226: the sex stone is refused while any of the 13 equipment slots holds a sex-only item (Special 64 or 65), whatever sex was requested. |
| [`broadcastNear`](../../server/internal/world/inventory.go#L523)<br>`func (s *Session) broadcastNear(frame []byte, label string, includeSelf bool)` | broadcastNear sends to viewers within the viewport that may see this character; includeSelf also echoes it to this session. |
| [`handleDropItem`](../../server/internal/world/inventory.go#L533)<br>`func (s *Session) handleDropItem(p []byte)` | handleDropItem (0x0206): bag, slot, count. |
| [`equipmentChanged`](../../server/internal/world/inventory.go#L591)<br>`func (s *Session) equipmentChanged(slots []uint8)` | equipmentChanged pushes 0x0507 for the slots to self and viewers. |
| [`mobGoldBonus`](../../server/internal/world/inventory.go#L606)<br>`func (s *Session) mobGoldBonus(amount uint32) uint32` | mobGoldBonus applies the dedicated mob-gold ability. |
| [`handlePickupItem`](../../server/internal/world/inventory.go#L619)<br>`func (s *Session) handlePickupItem(p []byte)` | handlePickupItem (0x0205): map item id. |
| [`moveItem`](../../server/internal/world/inventory.go#L693)<br>`func (s *Session) moveItem(srcBag, srcSlot, destBag, destSlot uint8) (frame []byte, ok bool)` | moveItem performs an inventory move/swap/merge and returns the 0x0204. |
| [`planMoveItem`](../../server/internal/world/inventory.go#L699)<br>`func (s *Session) planMoveItem(srcBag, srcSlot, destBag, destSlot uint8) ([]byte, bool)` | See the source definition and the subsystem ownership contract. |
| [`handleItemBagToBag`](../../server/internal/world/inventory.go#L753)<br>`func (s *Session) handleItemBagToBag(p []byte)` | handleItemBagToBag (0x0204): src bag/slot, dest bag/slot. |
| [`equipAllowed`](../../server/internal/world/inventory.go#L803)<br>`func (s *Session) equipAllowed(srcBag, srcSlot, destBag, destSlot uint8) bool` | equipAllowed checks both directions of a move that touches the equipment bag. |
| [`equipmentMoveFailure`](../../server/internal/world/inventory.go#L807)<br>`func (s *Session) equipmentMoveFailure(srcBag, srcSlot, destBag, destSlot uint8) string` | See the source definition and the subsystem ownership contract. |
| [`planGiveItem`](../../server/internal/world/inventory.go#L833)<br>`func planGiveItem(items *[]db.Item, tpl *db.Item, count, maxStack uint16) (packets [][]byte, left uint16)` | planGiveItem uses the native placement/stacking rules on a private copy. It allocates identifiers but performs no writes or network operations. |

### inventory_operation.go

[server/internal/world/inventory_operation.go](../../server/internal/world/inventory_operation.go#L1)

Direct GX dependencies: `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`inventoryOperation`](../../server/internal/world/inventory_operation.go#L13)<br>`type inventoryOperation` | inventoryOperation is an application unit of work. It stages character, equipment and packets while the session lock excludes other mutations. No database write or player publication occurs until its guarded commit. |
| [`operationPacket`](../../server/internal/world/inventory_operation.go#L20)<br>`type operationPacket` | See the source definition and the subsystem ownership contract. |
| [`packet`](../../server/internal/world/inventory_operation.go#L26)<br>`func (op *inventoryOperation) packet(frame []byte, label string, view bool)` | See the source definition and the subsystem ownership contract. |
| [`inventoryOperation`](../../server/internal/world/inventory_operation.go#L33)<br>`func (s *Session) inventoryOperation(run func()) bool` | See the source definition and the subsystem ownership contract. |
| [`inventoryOperationWithCommit`](../../server/internal/world/inventory_operation.go#L36)<br>`func (s *Session) inventoryOperationWithCommit(run func(), commit func(db.CharacterChange) error) (ok bool)` | See the source definition and the subsystem ownership contract. |
| [`saveInventoryItem`](../../server/internal/world/inventory_operation.go#L119)<br>`func (s *Session) saveInventoryItem(it *db.Item) error` | See the source definition and the subsystem ownership contract. |
| [`addInventoryItem`](../../server/internal/world/inventory_operation.go#L125)<br>`func (s *Session) addInventoryItem(it *db.Item) error` | See the source definition and the subsystem ownership contract. |
| [`operationPlayer`](../../server/internal/world/inventory_operation.go#L138)<br>`func (s *Session) operationPlayer() *ConnectedPlayer` | See the source definition and the subsystem ownership contract. |
| [`guildPublication`](../../server/internal/world/inventory_operation.go#L144)<br>`func (s *Session) guildPublication(g *guild, frame []byte, label string, except *ConnectedPlayer)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type inventoryOperation struct {
	packets       []operationPacket
	effects       []func()
	consumedSlots [][2]uint8
	player        *ConnectedPlayer
	err           error
}
```

```go
type operationPacket struct {
	frame []byte
	label string
	view  bool
}
```

</details>

### item_skills.go

[server/internal/world/item_skills.go](../../server/internal/world/item_skills.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`itemBuffSkill`](../../server/internal/world/item_skills.go#L11)<br>`func itemBuffSkill(d *db.ItemDef) *db.SkillDef` | All consumable Range/AttackTime buff references use the same native UseItemSkill -> BuffUser presentation, including commercial item skills. |
| [`announceItemSkill`](../../server/internal/world/item_skills.go#L21)<br>`func (s *Session) announceItemSkill(sk *db.SkillDef)` | See the source definition and the subsystem ownership contract. |
| [`canUseExpReset`](../../server/internal/world/item_skills.go#L33)<br>`func (s *Session) canUseExpReset() bool` | Native ItemUse special 50: only the two arena maps, after half a level. |

### itemboxes.go

[server/internal/world/itemboxes.go](../../server/internal/world/itemboxes.go#L1)

Direct GX dependencies: `internal/catalogfiles`, `internal/itemboxes`, `internal/worlddrops`, `server/internal/db`, `server/internal/logx`.

| Declaration | Responsibility / source comment |
|---|---|
| [`specialItemBox`](../../server/internal/world/itemboxes.go#L17)<br>`const specialItemBox = 42` | See the source definition and the subsystem ownership contract. |
| [`boxReward`](../../server/internal/world/itemboxes.go#L19)<br>`type boxReward` | See the source definition and the subsystem ownership contract. |
| [`itemBox`](../../server/internal/world/itemboxes.go#L24)<br>`type itemBox` | See the source definition and the subsystem ownership contract. |
| [`itemBoxCatalog`](../../server/internal/world/itemboxes.go#L30)<br>`var itemBoxCatalog atomic.Pointer[map[int16]itemBox]` | See the source definition and the subsystem ownership contract. |
| [`loadItemBoxes`](../../server/internal/world/itemboxes.go#L32)<br>`func loadItemBoxes(log *logx.Logger) error` | See the source definition and the subsystem ownership contract. |
| [`compileItemBoxes`](../../server/internal/world/itemboxes.go#L43)<br>`func compileItemBoxes(boxes []itemboxes.Box, defs *db.Defs, log logPrinter) map[int16]itemBox` | See the source definition and the subsystem ownership contract. |
| [`roll`](../../server/internal/world/itemboxes.go#L85)<br>`func (b itemBox) roll(draw func(uint64) uint64, award func(*db.ItemDef))` | See the source definition and the subsystem ownership contract. |
| [`boxItem`](../../server/internal/world/itemboxes.go#L104)<br>`func boxItem(def *db.ItemDef) db.Item` | See the source definition and the subsystem ownership contract. |
| [`canOpen`](../../server/internal/world/itemboxes.go#L108)<br>`func (b itemBox) canOpen(items []db.Item) bool` | See the source definition and the subsystem ownership contract. |
| [`openItemBox`](../../server/internal/world/itemboxes.go#L150)<br>`func (s *Session) openItemBox(slot int, index int16) bool` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type boxReward struct {
	items  []*db.ItemDef
	chance uint64
}
```

```go
type itemBox struct {
	boxType uint8
	total   uint64
	rewards []boxReward
}
```

</details>

### itemmall.go

[server/internal/world/itemmall.go](../../server/internal/world/itemmall.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`pointBuyOK`](../../server/internal/world/itemmall.go#L12)<br>`const pointBuyOK = 0` | See the source definition and the subsystem ownership contract. |
| [`pointBuyNoProduct`](../../server/internal/world/itemmall.go#L13)<br>`const pointBuyNoProduct = 1` | See the source definition and the subsystem ownership contract. |
| [`pointBuyNotEnoughPoints`](../../server/internal/world/itemmall.go#L14)<br>`const pointBuyNotEnoughPoints = 2` | See the source definition and the subsystem ownership contract. |
| [`pointBuyNotEnoughSlots`](../../server/internal/world/itemmall.go#L15)<br>`const pointBuyNotEnoughSlots = 3` | See the source definition and the subsystem ownership contract. |
| [`giftProductNotFound`](../../server/internal/world/itemmall.go#L17)<br>`const giftProductNotFound = 6` | See the source definition and the subsystem ownership contract. |
| [`giftNotEnoughPoints`](../../server/internal/world/itemmall.go#L18)<br>`const giftNotEnoughPoints = 10` | See the source definition and the subsystem ownership contract. |
| [`time32`](../../server/internal/world/itemmall.go#L21)<br>`func time32() uint32` | See the source definition and the subsystem ownership contract. |
| [`grantedItem`](../../server/internal/world/itemmall.go#L23)<br>`type grantedItem` | See the source definition and the subsystem ownership contract. |
| [`buildPointBuyResult`](../../server/internal/world/itemmall.go#L28)<br>`func buildPointBuyResult(result uint8, points uint32, code string, cost uint32, items []grantedItem) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildGiftSendResult`](../../server/internal/world/itemmall.go#L43)<br>`func buildGiftSendResult(result uint8, target string, points uint32, code string, cost uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildGiftBoxList`](../../server/internal/world/itemmall.go#L55)<br>`func buildGiftBoxList(userUID uint32) ([]byte, error)` | buildGiftBoxList (0x2606) lists UserStoredPointItems (slots 0-239). |
| [`sendGiftBox`](../../server/internal/world/itemmall.go#L69)<br>`func (s *Session) sendGiftBox()` | See the source definition and the subsystem ownership contract. |
| [`handlePointBuy`](../../server/internal/world/itemmall.go#L78)<br>`func (s *Session) handlePointBuy(p []byte)` | handlePointBuy (0x2602): product code. |
| [`handleGiftSend`](../../server/internal/world/itemmall.go#L146)<br>`func (s *Session) handleGiftSend(p []byte)` | handleGiftSend (0x2603): product code[21], target name[21]. |
| [`handleGiftClaim`](../../server/internal/world/itemmall.go#L190)<br>`func (s *Session) handleGiftClaim(p []byte)` | handleGiftClaim (0x2607): gift-box slot. |

<details><summary>Type definitions and owned fields</summary>

```go
type grantedItem struct {
	Bag, Slot, Type, TypeID uint8
	Count                   uint16
}
```

</details>

### itemoption.go

[server/internal/world/itemoption.go](../../server/internal/world/itemoption.go#L1)

Direct GX dependencies: `internal/itemtype`, `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`gemSkillLevel`](../../server/internal/world/itemoption.go#L26)<br>`const gemSkillLevel = db.NPCSkillWireLevel` | See the source definition and the subsystem ownership contract. |
| [`gemArmourOrder`](../../server/internal/world/itemoption.go#L29)<br>`var gemArmourOrder` | See the source definition and the subsystem ownership contract. |
| [`gemSkillsOf`](../../server/internal/world/itemoption.go#L31)<br>`func gemSkillsOf(it *db.Item) []*db.SkillDef` | See the source definition and the subsystem ownership contract. |
| [`immuneToLapisSkill`](../../server/internal/world/itemoption.go#L51)<br>`func (s *Session) immuneToLapisSkill(skill *db.SkillDef) bool` | immuneToLapisSkill uses the NPC skill namespace and exact ID. Equipped gems remain effective across gaps in armor slots. |
| [`weaponGemSkills`](../../server/internal/world/itemoption.go#L72)<br>`func weaponGemSkills(items []db.Item) []*db.SkillDef` | weaponGemSkills is AttackItemOption's weapon scan. |
| [`armourGemSkills`](../../server/internal/world/itemoption.go#L80)<br>`func armourGemSkills(items []db.Item) []*db.SkillDef` | armourGemSkills is BeAttackedItemOption's armour scan. |
| [`gemCharSkill`](../../server/internal/world/itemoption.go#L92)<br>`func gemCharSkill(sk *db.SkillDef) db.CharacterSkill` | See the source definition and the subsystem ownership contract. |
| [`gemCastsOnUser`](../../server/internal/world/itemoption.go#L97)<br>`func (s *Session) gemCastsOnUser(p *ConnectedPlayer, skills []*db.SkillDef)` | gemCastsOnUser queues this user's gem skills on p. |
| [`gemCastsOnMob`](../../server/internal/world/itemoption.go#L113)<br>`func (s *Session) gemCastsOnMob(mobID uint32, skills []*db.SkillDef)` | gemCastsOnMob casts this user's gem skills on a mob. |
| [`itemOptionsAfterUserHit`](../../server/internal/world/itemoption.go#L129)<br>`func (s *Session) itemOptionsAfterUserHit(h pendingUserHit)` | itemOptionsAfterUserHit runs both sides for a landed user-on-user hit, in the target's (this) session. |
| [`applyGemAttack`](../../server/internal/world/itemoption.go#L142)<br>`func (s *Session) applyGemAttack(target uint32)` | applyGemAttack is AttackItemOption for this (attacking) user on target. |

### killcount.go

[server/internal/world/killcount.go](../../server/internal/world/killcount.go#L1)

Direct GX dependencies: `internal/catalogfiles`, `internal/proto`, `server/internal/db`, `server/internal/logx`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`kcTypes`](../../server/internal/world/killcount.go#L51)<br>`const kcTypes = 28` | See the source definition and the subsystem ownership contract. |
| [`kcHoldTime`](../../server/internal/world/killcount.go#L52)<br>`const kcHoldTime = 10 * time.Minute` | See the source definition and the subsystem ownership contract. |
| [`kcBroadcastEvery`](../../server/internal/world/killcount.go#L53)<br>`const kcBroadcastEvery = 10 * time.Minute` | See the source definition and the subsystem ownership contract. |
| [`kcTier`](../../server/internal/world/killcount.go#L56)<br>`type kcTier` | See the source definition and the subsystem ownership contract. |
| [`kcCountry`](../../server/internal/world/killcount.go#L61)<br>`type kcCountry` | See the source definition and the subsystem ownership contract. |
| [`kcMu`](../../server/internal/world/killcount.go#L70)<br>`var kcMu sync.Mutex` | See the source definition and the subsystem ownership contract. |
| [`kcSendMu`](../../server/internal/world/killcount.go#L71)<br>`var kcSendMu sync.Mutex` | See the source definition and the subsystem ownership contract. |
| [`kcTiers`](../../server/internal/world/killcount.go#L72)<br>`var kcTiers []kcTier` | See the source definition and the subsystem ownership contract. |
| [`kcWorld`](../../server/internal/world/killcount.go#L73)<br>`var kcWorld` | See the source definition and the subsystem ownership contract. |
| [`loadKillStatus`](../../server/internal/world/killcount.go#L77)<br>`func loadKillStatus(log *logx.Logger) error` | loadKillStatus is CGameData::LoadKCStatus. |
| [`kcTierFor`](../../server/internal/world/killcount.go#L97)<br>`func kcTierFor(count uint32) int` | kcTierFor is CGameData::GetKCStatusByCount (-1 below the first tier). |
| [`kcMaxKill`](../../server/internal/world/killcount.go#L107)<br>`func kcMaxKill() uint32` | See the source definition and the subsystem ownership contract. |
| [`setKillCountLocked`](../../server/internal/world/killcount.go#L115)<br>`func setKillCountLocked(c int, count uint32, now time.Time)` | setKillCountLocked is CWorld::SetKillCount / InitKillCount for c. |
| [`InitKillCount`](../../server/internal/world/killcount.go#L130)<br>`func InitKillCount(log *logx.Logger)` | InitKillCount loads the saved counts at world start. |
| [`AddKillCount`](../../server/internal/world/killcount.go#L144)<br>`func AddKillCount(c uint8, n uint32, log logPrinter)` | AddKillCount is CWorld::AddKillCount for country c. |
| [`TickKillCount`](../../server/internal/world/killcount.go#L188)<br>`func TickKillCount(now time.Time, log logPrinter)` | TickKillCount is CWorld::UpdateKillCount. |
| [`kcSave`](../../server/internal/world/killcount.go#L223)<br>`func kcSave(light, dark uint32, log logPrinter)` | kcSave stores both counts; without a WorldInfo row (the world calendar creates it at start) it creates the row first. |
| [`kcStore`](../../server/internal/world/killcount.go#L230)<br>`var kcStore` | kcStore is the storage step (tests replace it). |
| [`kcBroadcast`](../../server/internal/world/killcount.go#L241)<br>`func kcBroadcast(c int, snapshot bool)` | kcBroadcast sends the native country count and, on state changes, the GX snapshot. |
| [`kcApplyCountry`](../../server/internal/world/killcount.go#L266)<br>`func kcApplyCountry(c int)` | kcApplyCountry is ApplyKillCount: every user of c refreshes its stats. |
| [`buildKCStatus`](../../server/internal/world/killcount.go#L275)<br>`func buildKCStatus(c uint8) []byte` | buildKCStatus is CUser::SendKCStatus (0x0211). |
| [`buildBlessingSnapshot`](../../server/internal/world/killcount.go#L293)<br>`func buildBlessingSnapshot(c uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`blessingSnapshotLocked`](../../server/internal/world/killcount.go#L299)<br>`func blessingSnapshotLocked(c uint8, entry bool) []byte` | See the source definition and the subsystem ownership contract. |
| [`kcStatus`](../../server/internal/world/killcount.go#L333)<br>`func kcStatus(c uint8) [kcTypes]int32` | kcStatus is the country's current tier types (zero below the first). |
| [`kcAbility`](../../server/internal/world/killcount.go#L343)<br>`var kcAbility` | kcAbility maps the flat KC types onto SetSkillAbility kinds (same field). |
| [`applyKCStatus`](../../server/internal/world/killcount.go#L346)<br>`func applyKCStatus(e *EquipmentStats, c uint8)` | applyKCStatus adds the country tier's user types to the stats. |
| [`kcPercent`](../../server/internal/world/killcount.go#L360)<br>`func kcPercent(base uint32, v int32) uint32` | kcPercent is UpdateKCStatus' (int)((float)(base * value) * 0.01). |
| [`kcWorldType`](../../server/internal/world/killcount.go#L369)<br>`func kcWorldType(c uint8, typ int) int32` | kcWorldType is a country-wide type (24 gem chance, 25 exp loss cut). |
| [`SubKillCount`](../../server/internal/world/killcount.go#L373)<br>`func SubKillCount(c uint8, n uint32, log logPrinter)` | SubKillCount is the obelisk capture debit. Active blessing freezes both countries, including subtraction; counts cannot wrap below zero. |
| [`addKillCountBatch`](../../server/internal/world/killcount.go#L402)<br>`func addKillCountBatch(country uint8, count uint32, log logPrinter)` | Keep threshold/hold semantics identical to individual kill credits while reducing a burst to at most one update per country tier. |

<details><summary>Type definitions and owned fields</summary>

```go
type kcTier struct {
	Kill   uint32
	Status [kcTypes]int32
}
```

```go
type kcCountry struct {
	count       uint32
	index       int // -1 below the first tier
	next        uint32
	resetAt     time.Time // the top tier's hold
	broadcastAt time.Time
}
```

</details>

### ladders.go

[server/internal/world/ladders.go](../../server/internal/world/ladders.go#L1)

Direct GX dependencies: `internal/mapgeometry`, `internal/proto`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`buildLadderState`](../../server/internal/world/ladders.go#L11)<br>`func buildLadderState(id uint32, index uint16, mode uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`notifyLadder`](../../server/internal/world/ladders.go#L18)<br>`func (s *Session) notifyLadder()` | See the source definition and the subsystem ownership contract. |
| [`handleLadder`](../../server/internal/world/ladders.go#L36)<br>`func (s *Session) handleLadder(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`advanceLadder`](../../server/internal/world/ladders.go#L105)<br>`func (s *Session) advanceLadder(now time.Time, dt float32)` | See the source definition and the subsystem ownership contract. |

### linked_bosses.go

[server/internal/world/linked_bosses.go](../../server/internal/world/linked_bosses.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`initializeBossLocked`](../../server/internal/world/linked_bosses.go#L8)<br>`func initializeBossLocked(mapID uint16, b *linkedBoss, now time.Time)` | See the source definition and the subsystem ownership contract. |
| [`bossRandomDelay`](../../server/internal/world/linked_bosses.go#L32)<br>`func bossRandomDelay(b *linkedBoss) time.Duration` | See the source definition and the subsystem ownership contract. |
| [`bossPortalMap`](../../server/internal/world/linked_bosses.go#L39)<br>`func bossPortalMap(b *linkedBoss) uint16` | See the source definition and the subsystem ownership contract. |
| [`bossDestroyedLocked`](../../server/internal/world/linked_bosses.go#L45)<br>`func bossDestroyedLocked(s *mapShard, b *linkedBoss, now time.Time, out *[]eventPublication)` | See the source definition and the subsystem ownership contract. |
| [`tickBossLocked`](../../server/internal/world/linked_bosses.go#L70)<br>`func tickBossLocked(s *mapShard, b *linkedBoss, now time.Time, out *[]eventPublication)` | See the source definition and the subsystem ownership contract. |
| [`initializeInstanceBosses`](../../server/internal/world/linked_bosses.go#L101)<br>`func initializeInstanceBosses(zone, base uint16)` | Each instance gets its own linked boss phases. Primary-map obelisks are deliberately not copied into party/guild instances. |
| [`syncBossPortal`](../../server/internal/world/linked_bosses.go#L126)<br>`func syncBossPortal(b *linkedBoss)` | See the source definition and the subsystem ownership contract. |

### loot.go

[server/internal/world/loot.go](../../server/internal/world/loot.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`lootGrace`](../../server/internal/world/loot.go#L22)<br>`const lootGrace = 10 * time.Minute` | Universal loot (a GX system that replaces the original item divisions): an item a party or raid earns goes straight into one member's bag. Members in reach of the dead monster (same map, partyRewardRange) take turns at random: the first item goes to anyone, then only to members who have not received one this round, and when everyone in reach has one a new round starts.  The group remembers each member while it exists. Whoever joins a round that already started waits for the next one; whoever leaves or disconnects keeps their place for lootGrace, so leaving and returning cannot be used to get served again and an honest disconnection costs nothing. |
| [`lootEntry`](../../server/internal/world/loot.go#L24)<br>`type lootEntry` | See the source definition and the subsystem ownership contract. |
| [`lootJoin`](../../server/internal/world/loot.go#L30)<br>`func (pt *party) lootJoin(id uint32, now time.Time)` | lootJoin records a member entering (partyMu held). |
| [`lootLeave`](../../server/internal/world/loot.go#L48)<br>`func (pt *party) lootLeave(id uint32, now time.Time)` | lootLeave records a member leaving or disconnecting (partyMu held). |
| [`lootEntryOf`](../../server/internal/world/loot.go#L54)<br>`func (pt *party) lootEntryOf(id uint32) *lootEntry` | See the source definition and the subsystem ownership contract. |
| [`lootPick`](../../server/internal/world/loot.go#L68)<br>`func (pt *party) lootPick(reach, fits []*ConnectedPlayer, now time.Time) *ConnectedPlayer` | lootPick chooses who receives one item among members able to carry it (partyMu held). reach is everyone in range, fits the subset with room. |
| [`lootFits`](../../server/internal/world/loot.go#L101)<br>`func lootFits(p *ConnectedPlayer, tpl *db.Item) bool` | lootFits reports whether the item fits in the member's bags. |
| [`distributeLoot`](../../server/internal/world/loot.go#L111)<br>`func distributeLoot(members []*ConnectedPlayer, tpl *db.Item, now time.Time) *ConnectedPlayer` | distributeLoot picks the recipient for an item earned by the group of members (already limited to the monster's reach); nil sends it to the ground. |
| [`lootGrant`](../../server/internal/world/loot.go#L131)<br>`type lootGrant` | lootGrant is an item queued for a member's session. |
| [`QueueLoot`](../../server/internal/world/loot.go#L137)<br>`func (p *ConnectedPlayer) QueueLoot(g lootGrant)` | See the source definition and the subsystem ownership contract. |
| [`takeLoot`](../../server/internal/world/loot.go#L143)<br>`func (p *ConnectedPlayer) takeLoot() []lootGrant` | See the source definition and the subsystem ownership contract. |
| [`applyLoot`](../../server/internal/world/loot.go#L153)<br>`func (s *Session) applyLoot(p *ConnectedPlayer)` | applyLoot delivers queued group loot to this session's bags; what no longer fits (the bag filled meanwhile) falls to the ground where the monster died. |
| [`announceLoot`](../../server/internal/world/loot.go#L177)<br>`func (s *Session) announceLoot(it db.Item)` | announceLoot tells the rest of the group who received the item (SysMsg 61). |
| [`lootOwner`](../../server/internal/world/loot.go#L205)<br>`func lootOwner(m *MapMob) *ConnectedPlayer` | lootOwner is who earns a monster's loot: the group (or lone player) that dealt the most damage to it, not the last hit nor the aggro holder. Damage by members of one party counts together. It returns the biggest damage dealer of the winning side; nil when nobody online dealt any. |

<details><summary>Type definitions and owned fields</summary>

```go
type lootEntry struct {
	served bool
	left   time.Time // zero while the member is in the group
}
```

```go
type lootGrant struct {
	item    db.Item
	mapID   uint16
	x, y, z float32
}
```

</details>

### map_access.go

[server/internal/world/map_access.go](../../server/internal/world/map_access.go#L1)

Direct GX dependencies: `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`canEnterMap`](../../server/internal/world/map_access.go#L5)<br>`func (s *Session) canEnterMap(mapID, level uint16) bool` | See the source definition and the subsystem ownership contract. |
| [`mapLevelAllowed`](../../server/internal/world/map_access.go#L9)<br>`func mapLevelAllowed(mapID, level uint16, gm bool) bool` | See the source definition and the subsystem ownership contract. |
| [`birthPosition`](../../server/internal/world/map_access.go#L15)<br>`func birthPosition(ch *db.Character) MapPos` | See the source definition and the subsystem ownership contract. |
| [`respawnDestination`](../../server/internal/world/map_access.go#L19)<br>`func (s *Session) respawnDestination(ch *db.Character) (uint16, float32, float32, float32)` | See the source definition and the subsystem ownership contract. |
| [`characterRespawn`](../../server/internal/world/map_access.go#L23)<br>`func characterRespawn(ch *db.Character, faction uint8, gm bool) (uint16, float32, float32, float32)` | See the source definition and the subsystem ownership contract. |
| [`enforceMapLevel`](../../server/internal/world/map_access.go#L31)<br>`func (s *Session) enforceMapLevel()` | See the source definition and the subsystem ownership contract. |

### map_curses.go

[server/internal/world/map_curses.go](../../server/internal/world/map_curses.go#L1)

Direct GX dependencies: `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`mapCurseDef`](../../server/internal/world/map_curses.go#L10)<br>`func mapCurseDef(defs *db.Defs, id uint16) *db.SkillDef` | See the source definition and the subsystem ownership contract. |
| [`refreshMapCurses`](../../server/internal/world/map_curses.go#L18)<br>`func (s *Session) refreshMapCurses(publish bool)` | See the source definition and the subsystem ownership contract. |

### mapitems.go

[server/internal/world/mapitems.go](../../server/internal/world/mapitems.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`mapItemOwnerWindow`](../../server/internal/world/mapitems.go#L15)<br>`const mapItemOwnerWindow = 30 * time.Second` | A dropped item is reserved for its owner for 30 s, open to anyone until 60 s, then despawns. |
| [`mapItemDespawnAge`](../../server/internal/world/mapitems.go#L16)<br>`const mapItemDespawnAge = 60 * time.Second` | See the source definition and the subsystem ownership contract. |
| [`moneyItemType`](../../server/internal/world/mapitems.go#L17)<br>`const moneyItemType = 26` | See the source definition and the subsystem ownership contract. |
| [`MapItem`](../../server/internal/world/mapitems.go#L21)<br>`type MapItem` | MapItem is an item or gold lying on a map. |
| [`mapItemEntry`](../../server/internal/world/mapitems.go#L32)<br>`type mapItemEntry` | See the source definition and the subsystem ownership contract. |
| [`itemClaim`](../../server/internal/world/mapitems.go#L36)<br>`type itemClaim` | See the source definition and the subsystem ownership contract. |
| [`nextMapItemID`](../../server/internal/world/mapitems.go#L42)<br>`func nextMapItemID() uint32` | See the source definition and the subsystem ownership contract. |
| [`addMapItem`](../../server/internal/world/mapitems.go#L43)<br>`func addMapItem(mapID uint16, it MapItem)` | See the source definition and the subsystem ownership contract. |
| [`findMapItem`](../../server/internal/world/mapitems.go#L52)<br>`func findMapItem(mapID uint16, id uint32) (MapItem, bool)` | See the source definition and the subsystem ownership contract. |
| [`claimMapItem`](../../server/internal/world/mapitems.go#L64)<br>`func claimMapItem(mapID uint16, id uint32) (MapItem, *itemClaim, bool)` | Claim reserves one ground object across concurrent players without holding a map mutex during SQLite IO. Failed persistence releases the unchanged object. |
| [`Abort`](../../server/internal/world/mapitems.go#L75)<br>`func (c *itemClaim) Abort()` | See the source definition and the subsystem ownership contract. |
| [`Finish`](../../server/internal/world/mapitems.go#L78)<br>`func (c *itemClaim) Finish(left uint16) (updated MapItem)` | See the source definition and the subsystem ownership contract. |
| [`itemsOnMap`](../../server/internal/world/mapitems.go#L96)<br>`func itemsOnMap(mapID uint16) []MapItem` | See the source definition and the subsystem ownership contract. |
| [`expiredItem`](../../server/internal/world/mapitems.go#L108)<br>`type expiredItem` | See the source definition and the subsystem ownership contract. |
| [`sweepExpiredMapItems`](../../server/internal/world/mapitems.go#L113)<br>`func sweepExpiredMapItems() []expiredItem` | See the source definition and the subsystem ownership contract. |
| [`buildMapAddItem`](../../server/internal/world/mapitems.go#L129)<br>`func buildMapAddItem(m *MapItem) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildMapRemoveItem`](../../server/internal/world/mapitems.go#L146)<br>`func buildMapRemoveItem(id uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`clearMapItems`](../../server/internal/world/mapitems.go#L153)<br>`func clearMapItems(zone uint16)` | clearMapItems drops every item of a zone (instance teardown). |

<details><summary>Type definitions and owned fields</summary>

```go
type MapItem struct {
	ID               uint32
	Item             db.Item
	Money            uint32
	OwnerID          uint32
	NoParty          bool // a player's drop: its party does not share it
	FromMob          bool // mob loot (drop type 1): gold can receive the mob-gold bonus
	DroppedAt        time.Time
	PosX, PosY, PosZ float32
}
```

```go
type mapItemEntry struct {
	item    MapItem
	claimed bool
}
```

```go
type itemClaim struct {
	shard *mapShard
	entry *mapItemEntry
	once  sync.Once
}
```

```go
type expiredItem struct {
	MapID uint16
	Item  MapItem
}
```

</details>

### mob_ai.go

[server/internal/world/mob_ai.go](../../server/internal/world/mob_ai.go#L1)

Direct GX dependencies: `internal/mapnav`, `server/ai`, `server/internal/config`, `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`mobAIProgram`](../../server/internal/world/mob_ai.go#L21)<br>`var mobAIProgram atomic.Pointer[ai.Program]` | All AI decisions enter Lua. This adapter supplies authoritative world primitives under the shard lock; cross-controller operations are deferred. |
| [`mobAIClock`](../../server/internal/world/mob_ai.go#L22)<br>`var mobAIClock` | See the source definition and the subsystem ownership contract. |
| [`loadMobAI`](../../server/internal/world/mob_ai.go#L24)<br>`func loadMobAI() error` | See the source definition and the subsystem ownership contract. |
| [`aiProgram`](../../server/internal/world/mob_ai.go#L45)<br>`func aiProgram() (*ai.Program, error)` | See the source definition and the subsystem ownership contract. |
| [`mobAIState`](../../server/internal/world/mob_ai.go#L52)<br>`type mobAIState` | See the source definition and the subsystem ownership contract. |
| [`resetAI`](../../server/internal/world/mob_ai.go#L76)<br>`func (m *MapMob) resetAI()` | See the source definition and the subsystem ownership contract. |
| [`ensureAI`](../../server/internal/world/mob_ai.go#L85)<br>`func (m *MapMob) ensureAI(h *mobAIHost) bool` | See the source definition and the subsystem ownership contract. |
| [`failAI`](../../server/internal/world/mob_ai.go#L120)<br>`func (m *MapMob) failAI(err error)` | See the source definition and the subsystem ownership contract. |
| [`event`](../../server/internal/world/mob_ai.go#L125)<br>`func (h *mobAIHost) event(name string, args ...any) bool` | See the source definition and the subsystem ownership contract. |
| [`notifyTarget`](../../server/internal/world/mob_ai.go#L136)<br>`func (h *mobAIHost) notifyTarget()` | See the source definition and the subsystem ownership contract. |
| [`mobAIHost`](../../server/internal/world/mob_ai.go#L146)<br>`type mobAIHost` | See the source definition and the subsystem ownership contract. |
| [`queue`](../../server/internal/world/mob_ai.go#L155)<br>`func (h *mobAIHost) queue(kind string, id uint16, count int, x, z float32, mode int) error` | See the source definition and the subsystem ownership contract. |
| [`aiNumber`](../../server/internal/world/mob_ai.go#L180)<br>`func aiNumber(args []any, i int) float64` | See the source definition and the subsystem ownership contract. |
| [`aiBool`](../../server/internal/world/mob_ai.go#L196)<br>`func aiBool(args []any, i int) bool` | See the source definition and the subsystem ownership contract. |
| [`Call`](../../server/internal/world/mob_ai.go#L205)<br>`func (h *mobAIHost) Call(name string, args []any) (any, error)` | See the source definition and the subsystem ownership contract. |
| [`heal`](../../server/internal/world/mob_ai.go#L530)<br>`func (h *mobAIHost) heal(amount uint32)` | See the source definition and the subsystem ownership contract. |
| [`refreshTarget`](../../server/internal/world/mob_ai.go#L541)<br>`func (h *mobAIHost) refreshTarget()` | See the source definition and the subsystem ownership contract. |
| [`recruit`](../../server/internal/world/mob_ai.go#L569)<br>`func (h *mobAIHost) recruit(user uint32, x, z, radius float32, kind string)` | See the source definition and the subsystem ownership contract. |
| [`selectDirection`](../../server/internal/world/mob_ai.go#L593)<br>`func (h *mobAIHost) selectDirection(extra float32)` | See the source definition and the subsystem ownership contract. |
| [`moveSelected`](../../server/internal/world/mob_ai.go#L604)<br>`func (h *mobAIHost) moveSelected(run bool)` | See the source definition and the subsystem ownership contract. |
| [`attack`](../../server/internal/world/mob_ai.go#L639)<br>`func (h *mobAIHost) attack(stationary bool)` | See the source definition and the subsystem ownership contract. |
| [`skill`](../../server/internal/world/mob_ai.go#L698)<br>`func (h *mobAIHost) skill(id uint16)` | See the source definition and the subsystem ownership contract. |
| [`mobAICommand`](../../server/internal/world/mob_ai.go#L741)<br>`type mobAICommand` | See the source definition and the subsystem ownership contract. |
| [`runMobAICommands`](../../server/internal/world/mob_ai.go#L754)<br>`func runMobAICommands(tick *MobCombatTick)` | See the source definition and the subsystem ownership contract. |
| [`isSacraGuard`](../../server/internal/world/mob_ai.go#L875)<br>`func isSacraGuard(m *MapMob) bool` | Sacra's first-floor bosses each send the portal update on death. The summoned encounter pair must both be defeated in this instance. |
| [`hasLivingSacraGuard`](../../server/internal/world/mob_ai.go#L879)<br>`func hasLivingSacraGuard(mobs []MapMob) bool` | See the source definition and the subsystem ownership contract. |
| [`recordMobAIMiss`](../../server/internal/world/mob_ai.go#L888)<br>`func recordMobAIMiss(mapID uint16, id, user uint32)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type mobAIState struct {
	actor                                        *ai.Actor
	rng                                          *rand.Rand
	initialized                                  bool
	fault                                        error
	lastTarget                                   uint32
	suppressHelp                                 bool
	needReset                                    bool
	damage, speed, defense, recovery             float32
	mask                                         uint8
	invulnerable, hold, disabled, recoverEnabled bool
	recoverAt                                    time.Time
	recoverPeriod                                time.Duration
	attacked, recentHit, counting                bool
	hits, attackedCount                          int
	battleType                                   int
	element                                      uint8
	elementSet                                   bool
	selected                                     mapnav.Point
	selectedValid, fixed, arrived                bool
	fixedRun                                     bool
	escapeX, escapeY, escapeZ                    float32
}
```

```go
type mobAIHost struct {
	mob     *MapMob
	shard   *mapShard
	mapID   uint16
	now     time.Time
	players []PlayerCombatSnapshot
	tick    *MobCombatTick
}
```

```go
type mobAICommand struct {
	kind    string
	mapID   uint16
	owner   uint32
	actor   *ai.Actor
	id      uint16
	count   int
	x, y, z float32
	mode    int
	spread  float32
	legacy  bool
}
```

</details>

### mob_ai_speech.go

[server/internal/world/mob_ai_speech.go](../../server/internal/world/mob_ai_speech.go#L1)

Direct GX dependencies: `internal/catalogfiles`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`mobAIStrings`](../../server/internal/world/mob_ai_speech.go#L11)<br>`var mobAIStrings` | See the source definition and the subsystem ownership contract. |
| [`loadMobAIStrings`](../../server/internal/world/mob_ai_speech.go#L13)<br>`func loadMobAIStrings() error` | See the source definition and the subsystem ownership contract. |
| [`buildMobAIMessage`](../../server/internal/world/mob_ai_speech.go#L26)<br>`func buildMobAIMessage(owner uint32, index uint16, text string, indexed bool) []byte` | See the source definition and the subsystem ownership contract. |
| [`publishMobAISpeech`](../../server/internal/world/mob_ai_speech.go#L45)<br>`func publishMobAISpeech(c mobAICommand)` | See the source definition and the subsystem ownership contract. |

### mob_navigation.go

[server/internal/world/mob_navigation.go](../../server/internal/world/mob_navigation.go#L1)

Direct GX dependencies: `internal/mapgeometry`, `internal/mapnav`, `internal/movespeed`, `server/internal/config`.

| Declaration | Responsibility / source comment |
|---|---|
| [`mobNavigation`](../../server/internal/world/mob_navigation.go#L17)<br>`var mobNavigation` | See the source definition and the subsystem ownership contract. |
| [`loadMobNavigation`](../../server/internal/world/mob_navigation.go#L24)<br>`func loadMobNavigation(mapID uint16) (*mapnav.Map, error)` | See the source definition and the subsystem ownership contract. |
| [`point`](../../server/internal/world/mob_navigation.go#L72)<br>`func (m *MapMob) point() mapnav.Point` | See the source definition and the subsystem ownership contract. |
| [`setPoint`](../../server/internal/world/mob_navigation.go#L73)<br>`func (m *MapMob) setPoint(p mapnav.Point)` | See the source definition and the subsystem ownership contract. |
| [`placeOnNavigation`](../../server/internal/world/mob_navigation.go#L75)<br>`func (m *MapMob) placeOnNavigation(nav *mapnav.Map, mapID uint16) bool` | See the source definition and the subsystem ownership contract. |
| [`mobSpawnHasExit`](../../server/internal/world/mob_navigation.go#L133)<br>`func mobSpawnHasExit(nav *mapnav.Map, mapID uint16, spawn mapnav.Point) bool` | See the source definition and the subsystem ownership contract. |
| [`navigate`](../../server/internal/world/mob_navigation.go#L149)<br>`func (m *MapMob) navigate(now time.Time, mapID uint16, goal mapnav.Point, step float32) (mapnav.Point, bool)` | See the source definition and the subsystem ownership contract. |
| [`targetDistance`](../../server/internal/world/mob_navigation.go#L172)<br>`func (m *MapMob) targetDistance(t *PlayerCombatSnapshot) float32` | See the source definition and the subsystem ownership contract. |
| [`canReachTarget`](../../server/internal/world/mob_navigation.go#L178)<br>`func (m *MapMob) canReachTarget(t *PlayerCombatSnapshot) bool` | See the source definition and the subsystem ownership contract. |
| [`mobNav`](../../server/internal/world/mob_navigation.go#L187)<br>`type mobNav` | Mobs retain a lightweight map reference. The decoded geometry cache is bounded instead of keeping every world map resident merely because it has spawns. |
| [`Path`](../../server/internal/world/mob_navigation.go#L189)<br>`func (n *mobNav) Path(a, b mapnav.Point) ([]mapnav.Point, bool)` | See the source definition and the subsystem ownership contract. |
| [`Step`](../../server/internal/world/mob_navigation.go#L196)<br>`func (n *mobNav) Step(a mapnav.Point, path []mapnav.Point, d float32, walkable func(float32, float32) bool) (mapnav.Point, float32, int, bool)` | See the source definition and the subsystem ownership contract. |
| [`Visible`](../../server/internal/world/mob_navigation.go#L203)<br>`func (n *mobNav) Visible(a, b mapnav.Point) bool` | See the source definition and the subsystem ownership contract. |
| [`placeRespawn`](../../server/internal/world/mob_navigation.go#L207)<br>`func (m *MapMob) placeRespawn(mapID uint16) bool` | See the source definition and the subsystem ownership contract. |
| [`mobNavigationTick`](../../server/internal/world/mob_navigation.go#L213)<br>`const mobNavigationTick = 50 * time.Millisecond` | Corner arrivals need a finer scheduler than the legacy half-second sweep. |
| [`movingAt`](../../server/internal/world/mob_navigation.go#L215)<br>`func (m *MapMob) movingAt(now time.Time) bool` | See the source definition and the subsystem ownership contract. |
| [`positionAt`](../../server/internal/world/mob_navigation.go#L219)<br>`func (m *MapMob) positionAt(now time.Time) mapnav.Point` | See the source definition and the subsystem ownership contract. |
| [`moveTo`](../../server/internal/world/mob_navigation.go#L240)<br>`func (m *MapMob) moveTo(mapID uint16, now time.Time, to mapnav.Point, motion uint8) entityMove` | See the source definition and the subsystem ownership contract. |
| [`refreshMoveSpeed`](../../server/internal/world/mob_navigation.go#L273)<br>`func (m *MapMob) refreshMoveSpeed(mapID uint16, now time.Time, tick *MobCombatTick)` | The original client reads the current speed index every frame. Update the active segment from its current position so World and the client keep the same destination and arrival time when a speed effect changes. |
| [`appearance`](../../server/internal/world/mob_navigation.go#L279)<br>`func (m *MapMob) appearance(now time.Time, fresh bool) (enter, move []byte)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type mobNav struct{ id uint16 }
```

</details>

### mobeffects.go

[server/internal/world/mobeffects.go](../../server/internal/world/mobeffects.go#L1)

Direct GX dependencies: `internal/skillchance`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`detailTimeDamage`](../../server/internal/world/mobeffects.go#L28)<br>`const detailTimeDamage = 61` | See the source definition and the subsystem ownership contract. |
| [`detailAbilityDown`](../../server/internal/world/mobeffects.go#L29)<br>`const detailAbilityDown = 62` | See the source definition and the subsystem ownership contract. |
| [`detailDeathSentence`](../../server/internal/world/mobeffects.go#L30)<br>`const detailDeathSentence = 63` | See the source definition and the subsystem ownership contract. |
| [`detailStun`](../../server/internal/world/mobeffects.go#L31)<br>`const detailStun = 64` | See the source definition and the subsystem ownership contract. |
| [`detailStop`](../../server/internal/world/mobeffects.go#L32)<br>`const detailStop = 65` | See the source definition and the subsystem ownership contract. |
| [`detailSleep`](../../server/internal/world/mobeffects.go#L33)<br>`const detailSleep = 66` | See the source definition and the subsystem ownership contract. |
| [`detailAbilityDown2`](../../server/internal/world/mobeffects.go#L34)<br>`const detailAbilityDown2 = 67` | See the source definition and the subsystem ownership contract. |
| [`detailNoElement2`](../../server/internal/world/mobeffects.go#L35)<br>`const detailNoElement2 = 68` | See the source definition and the subsystem ownership contract. |
| [`detailVisionFix`](../../server/internal/world/mobeffects.go#L36)<br>`const detailVisionFix = 72` | See the source definition and the subsystem ownership contract. |
| [`effectTick`](../../server/internal/world/mobeffects.go#L37)<br>`const effectTick = 3 * time.Second` | See the source definition and the subsystem ownership contract. |
| [`MobEffect`](../../server/internal/world/mobeffects.go#L41)<br>`type MobEffect` | MobEffect is one skill applied to a mob. |
| [`mobMods`](../../server/internal/world/mobeffects.go#L49)<br>`type mobMods` | mobMods are CMob::SetSkillAbility's runtime changes. |
| [`mods`](../../server/internal/world/mobeffects.go#L60)<br>`func (m *MapMob) mods() mobMods` | See the source definition and the subsystem ownership contract. |
| [`addAbility`](../../server/internal/world/mobeffects.go#L88)<br>`func (o *mobMods) addAbility(t uint8, v int32)` | See the source definition and the subsystem ownership contract. |
| [`canAct`](../../server/internal/world/mobeffects.go#L120)<br>`func (o mobMods) canAct() bool` | canAct: stunned or sleeping mobs neither attack nor move. |
| [`canMove`](../../server/internal/world/mobeffects.go#L123)<br>`func (o mobMods) canMove() bool` | canMove: a stopped mob still attacks in reach. |
| [`attackSpeedIndex`](../../server/internal/world/mobeffects.go#L127)<br>`func (o mobMods) attackSpeedIndex() uint8` | attackSpeedIndex and moveSpeedIndex are CMob::SendSpeed's values: 255 while blocked, otherwise the speed index (3 and 2 at normal speed). |
| [`moveSpeedIndex`](../../server/internal/world/mobeffects.go#L134)<br>`func (o mobMods) moveSpeedIndex() uint8` | See the source definition and the subsystem ownership contract. |
| [`mobAttackSpeedOffsets`](../../server/internal/world/mobeffects.go#L142)<br>`var mobAttackSpeedOffsets` | mobAttackSpeedOffsets is g_ConstMobAttackSpeed by speed index. |
| [`attackDelayOffset`](../../server/internal/world/mobeffects.go#L144)<br>`func (o mobMods) attackDelayOffset() time.Duration` | See the source definition and the subsystem ownership contract. |
| [`buildMobSpeed`](../../server/internal/world/mobeffects.go#L149)<br>`func buildMobSpeed(m *MapMob) []byte` | buildMobSpeed (0x0609): mob id, attack speed, move speed. |
| [`mobCombatOf`](../../server/internal/world/mobeffects.go#L159)<br>`func mobCombatOf(m *MapMob) mobCombat` | mobCombatOf is mobCombatFor with the mob's applied skills. |
| [`MobCombatByID`](../../server/internal/world/mobeffects.go#L194)<br>`func MobCombatByID(mapID uint16, id uint32) (mobCombat, bool)` | MobCombatByID returns a live mob's combat values with its applied skills. |
| [`ApplyMobEffect`](../../server/internal/world/mobeffects.go#L211)<br>`func ApplyMobEffect(mapID uint16, id, casterID uint32, def *db.SkillDef) (speed []byte, ok bool)` | ApplyMobEffect is CMob::AddApplySkill: the same skill at the same level refreshes; a higher level (levels 1-3) replaces a lower one; a lower one is ignored. Returns the 0x0609 packet when the mob's speed changed. |
| [`removeSleep`](../../server/internal/world/mobeffects.go#L265)<br>`func (m *MapMob) removeSleep() []byte` | removeSleep is CMob::RemApplySkillSleep: damage wakes the mob. Returns the 0x0609 packet when it was asleep. The shard lock is held. |
| [`wakeMob`](../../server/internal/world/mobeffects.go#L282)<br>`func wakeMob(mapID uint16, id uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`mobTimeDamage`](../../server/internal/world/mobeffects.go#L298)<br>`func mobTimeDamage(def *db.SkillDef, cur, maxv [3]uint32, elapsed time.Duration) [3]int32` | mobTimeDamage is the TimeDamage formula for a mob target: mode 0 the raw values, 3 the mob's current pools times the value, 4 and 5 a percentage of its maximum or current pools; then scaled by elapsed/3000 and capped at 30000. |
| [`MobEffectOutcome`](../../server/internal/world/mobeffects.go#L327)<br>`type MobEffectOutcome` | MobEffectOutcome is one damage or expiry event from TickMobEffects. |
| [`TickMobEffects`](../../server/internal/world/mobeffects.go#L343)<br>`func TickMobEffects() []MobEffectOutcome` | TickMobEffects is ChkKeepSkill (detail 61, every 3 s) and ChkEndTimeSkill (expiry, with the last partial tick of 61 and the whole-HP hit of 63). |
| [`tickMobEffects`](../../server/internal/world/mobeffects.go#L345)<br>`func tickMobEffects(now time.Time) []MobEffectOutcome` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type MobEffect struct {
	Def      *db.SkillDef
	CasterID uint32
	endsAt   time.Time
	prevTick time.Time
}
```

```go
type mobMods struct {
	Wis, Dex, Luc          int32
	Critical               int32
	AttackSpeed, MoveSpeed int32
	Cat                    [3]CategoryMods
	Disabled               [3]int32
	Stun, Stop, Sleep      int32
	NoElement              bool
	VisionFix              uint32
}
```

```go
type MobEffectOutcome struct {
	MapID          uint16
	Mob            MapMob
	CasterID       uint32
	SkillID        uint16
	SkillLevel     uint8
	Damage         [3]int32
	AppliedHP      uint32
	Damaged        bool
	Killed         bool
	RespawnSeconds uint64
	Speed          []byte // 0x0609 when an expiry changed the speed
}
```

</details>

### movement.go

[server/internal/world/movement.go](../../server/internal/world/movement.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`positionSaveInterval`](../../server/internal/world/movement.go#L14)<br>`const positionSaveInterval = 5 * time.Second` | See the source definition and the subsystem ownership contract. |
| [`usablePortal`](../../server/internal/world/movement.go#L16)<br>`func (s *Session) usablePortal(x, y, z float32, index *uint8) (MapPortal, bool)` | See the source definition and the subsystem ownership contract. |
| [`savePosition`](../../server/internal/world/movement.go#L42)<br>`func (s *Session) savePosition()` | savePosition persists the character (usp_Save_Char_Info_E). |
| [`syncPlayer`](../../server/internal/world/movement.go#L54)<br>`func (s *Session) syncPlayer()` | See the source definition and the subsystem ownership contract. |
| [`applyTeleport`](../../server/internal/world/movement.go#L63)<br>`func (s *Session) applyTeleport(mapID uint16, x, y, z float32, label string, gmDirect ...bool) bool` | applyTeleport is the single path for every teleport (portals, town, resurrection): moves the character, persists it, acks the client, introduces it to viewers at the destination and refreshes visibility. |
| [`handleCharacterMove`](../../server/internal/world/movement.go#L150)<br>`func (s *Session) handleCharacterMove(p []byte)` | handleCharacterMove (0x0501): motion:u8, direction:u16, x, y, z:f32. |
| [`handleEnteredPortal`](../../server/internal/world/movement.go#L228)<br>`func (s *Session) handleEnteredPortal(index uint8)` | See the source definition and the subsystem ownership contract. |
| [`isTownCommand`](../../server/internal/world/movement.go#L247)<br>`func isTownCommand(message string) bool` | See the source definition and the subsystem ownership contract. |
| [`executeTown`](../../server/internal/world/movement.go#L254)<br>`func (s *Session) executeTown()` | executeTown moves the character to the nearest faction spawn area on its map (also the escape hatch while dead). |
| [`deathExpPercent`](../../server/internal/world/movement.go#L278)<br>`const deathExpPercent = 5` | deathExpPercent is CUser::Death's drop without protection (deathExpLoss). |
| [`expDrop`](../../server/internal/world/movement.go#L282)<br>`func expDrop(ch *db.Character, percent int, faction uint8) uint32` | expDrop is CUser::ExpDrop (0x0045ad50): above level 10, percent of the level's exp span, never below the level's start. |
| [`returnToTownDelay`](../../server/internal/world/movement.go#L302)<br>`const returnToTownDelay = 10 * time.Second` | returnToTownDelay is 0x0556's tick + 10000 (CUser::PacketPC). |
| [`handleReturnToTown`](../../server/internal/world/movement.go#L306)<br>`func (s *Session) handleReturnToTown()` | handleReturnToTown (0x0556): alive and with no move pending, the user is moved to its bind point 10 s later (UpdateResetPosition, type 0). |
| [`cancelMove`](../../server/internal/world/movement.go#L315)<br>`func (s *Session) cancelMove()` | cancelMove is CUser::CancleMove for the type-0 move: moving, attacking, casting or being hit drops it without a reply. |
| [`tickReturnToTown`](../../server/internal/world/movement.go#L317)<br>`func (s *Session) tickReturnToTown()` | See the source definition and the subsystem ownership contract. |

### movement_rune.go

[server/internal/world/movement_rune.go](../../server/internal/world/movement_rune.go#L1)

Direct GX dependencies: `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`specialMovementRune`](../../server/internal/world/movement_rune.go#L15)<br>`const specialMovementRune = 32` | The Movement Rune (Special 32, CUser::ItemUse case 0x20 at 0x465e55): used on a selected member of the user's party, it takes five seconds (move type 5, 0x1388 ms later) and carries the user to where that member stands. The member's state does not matter: a dead member can be moved to. |
| [`movementRuneTime`](../../server/internal/world/movement_rune.go#L16)<br>`const movementRuneTime = 5 * time.Second` | The Movement Rune (Special 32, CUser::ItemUse case 0x20 at 0x465e55): used on a selected member of the user's party, it takes five seconds (move type 5, 0x1388 ms later) and carries the user to where that member stands. The member's state does not matter: a dead member can be moved to. |
| [`movementRuneCast`](../../server/internal/world/movement_rune.go#L19)<br>`type movementRuneCast` | See the source definition and the subsystem ownership contract. |
| [`movementRuneTarget`](../../server/internal/world/movement_rune.go#L25)<br>`func (s *Session) movementRuneTarget(id uint32) (*ConnectedPlayer, string)` | movementRuneTarget is the party member with that id, if the rune can go to it. |
| [`beginMovementRune`](../../server/internal/world/movement_rune.go#L40)<br>`func (s *Session) beginMovementRune(i int, def *db.ItemDef)` | See the source definition and the subsystem ownership contract. |
| [`cancelMovementRune`](../../server/internal/world/movement_rune.go#L86)<br>`func (s *Session) cancelMovementRune()` | See the source definition and the subsystem ownership contract. |
| [`tickMovementRune`](../../server/internal/world/movement_rune.go#L97)<br>`func (s *Session) tickMovementRune(now time.Time)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type movementRuneCast struct {
	target uint32
	ends   time.Time
}
```

</details>

### myshop.go

[server/internal/world/myshop.go](../../server/internal/world/myshop.go#L1)

Direct GX dependencies: `internal/chatguard`, `internal/proto`, `internal/textwire`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`opMyShopBegin`](../../server/internal/world/myshop.go#L21)<br>`const opMyShopBegin = 0x2301` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. enter setup |
| [`opMyShopCancel`](../../server/internal/world/myshop.go#L22)<br>`const opMyShopCancel = 0x2302` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. cancel / market closed |
| [`opMyShopOpen`](../../server/internal/world/myshop.go#L23)<br>`const opMyShopOpen = 0x2303` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. start selling with a message |
| [`opMyShopStop`](../../server/internal/world/myshop.go#L24)<br>`const opMyShopStop = 0x2304` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. stop selling, back to setup |
| [`opMyShopAdd`](../../server/internal/world/myshop.go#L25)<br>`const opMyShopAdd = 0x2305` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. |
| [`opMyShopRemove`](../../server/internal/world/myshop.go#L26)<br>`const opMyShopRemove = 0x2306` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. |
| [`opMyShopPrice`](../../server/internal/world/myshop.go#L27)<br>`const opMyShopPrice = 0x2307` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. |
| [`opMyShopSold`](../../server/internal/world/myshop.go#L28)<br>`const opMyShopSold = 0x2308` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. to the seller: a sale |
| [`opMyShopEnter`](../../server/internal/world/myshop.go#L29)<br>`const opMyShopEnter = 0x2309` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. a buyer opens a market |
| [`opMyShopLeave`](../../server/internal/world/myshop.go#L30)<br>`const opMyShopLeave = 0x230a` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. |
| [`opMyShopList`](../../server/internal/world/myshop.go#L31)<br>`const opMyShopList = 0x230b` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. |
| [`opMyShopClosed`](../../server/internal/world/myshop.go#L32)<br>`const opMyShopClosed = 0x230c` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. to buyers: the market ended |
| [`opMyShopBuy`](../../server/internal/world/myshop.go#L33)<br>`const opMyShopBuy = 0x230d` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. |
| [`opMyShopStock`](../../server/internal/world/myshop.go#L34)<br>`const opMyShopStock = 0x230e` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. to buyers: an item's new count |
| [`opShopBubble`](../../server/internal/world/myshop.go#L35)<br>`const opShopBubble = 0x0523` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. the market's bubble over the seller |
| [`myShopSlots`](../../server/internal/world/myshop.go#L37)<br>`const myShopSlots = proto.MyShopSlots` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. |
| [`myShopBuyersMax`](../../server/internal/world/myshop.go#L38)<br>`const myShopBuyersMax = 16` | Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and stMyShop: a seated player sells items from the bags at a price per unit in a map where markets are allowed. Buyers browse the list and choose how many units to take. Moving, standing, dying, leaving the map or logging out ends the market. |
| [`myShopMaps`](../../server/internal/world/myshop.go#L42)<br>`var myShopMaps` | Maps where a market can be set up (GX Server configuration for now). |
| [`myShopBeginOK`](../../server/internal/world/myshop.go#L46)<br>`const myShopBeginOK = 0` | Results of the begin request (client message in the reply). |
| [`myShopBeginWrongMap`](../../server/internal/world/myshop.go#L47)<br>`const myShopBeginWrongMap = 1` | Results of the begin request (client message in the reply). |
| [`myShopBeginBusy`](../../server/internal/world/myshop.go#L48)<br>`const myShopBeginBusy = 2` | Results of the begin request (client message in the reply). |
| [`myShopBuyOK`](../../server/internal/world/myshop.go#L53)<br>`const myShopBuyOK = 1` | Results of a purchase. |
| [`myShopBuyInvalid`](../../server/internal/world/myshop.go#L54)<br>`const myShopBuyInvalid = 0` | Results of a purchase. |
| [`myShopBuyNoMoney`](../../server/internal/world/myshop.go#L55)<br>`const myShopBuyNoMoney = 2` | Results of a purchase. |
| [`myShopBuySellerCap`](../../server/internal/world/myshop.go#L56)<br>`const myShopBuySellerCap = 3` | Results of a purchase. |
| [`myShopBuyNoRoom`](../../server/internal/world/myshop.go#L57)<br>`const myShopBuyNoRoom = 4` | Results of a purchase. |
| [`myShopEnterOK`](../../server/internal/world/myshop.go#L58)<br>`const myShopEnterOK = 1` | Results of a purchase. |
| [`myShopEnterFailed`](../../server/internal/world/myshop.go#L59)<br>`const myShopEnterFailed = 0` | Results of a purchase. |
| [`myShopEnterTooMany`](../../server/internal/world/myshop.go#L60)<br>`const myShopEnterTooMany = 2` | Results of a purchase. |
| [`myShopStatusNone`](../../server/internal/world/myshop.go#L61)<br>`const myShopStatusNone = 0` | Results of a purchase. |
| [`myShopStatusSetup`](../../server/internal/world/myshop.go#L62)<br>`const myShopStatusSetup = 1` | Results of a purchase. |
| [`myShopStatusSelling`](../../server/internal/world/myshop.go#L63)<br>`const myShopStatusSelling = 2` | Results of a purchase. |
| [`shopSlot`](../../server/internal/world/myshop.go#L66)<br>`type shopSlot` | See the source definition and the subsystem ownership contract. |
| [`myShop`](../../server/internal/world/myshop.go#L72)<br>`type myShop` | See the source definition and the subsystem ownership contract. |
| [`itemIndex`](../../server/internal/world/myshop.go#L79)<br>`func (sl shopSlot) itemIndex(items []db.Item) int` | Match the offered item even when its bag position is occupied by a new item. |
| [`shopAck`](../../server/internal/world/myshop.go#L95)<br>`func shopAck(op packet.Opcode, body func(w *packet.Writer)) []byte` | See the source definition and the subsystem ownership contract. |
| [`shopBubbleFrame`](../../server/internal/world/myshop.go#L104)<br>`func shopBubbleFrame(id uint32, title string) []byte` | shopBubbleFrame is 0x0523: the market message over the seller (off: no text). |
| [`setShopTitle`](../../server/internal/world/myshop.go#L117)<br>`func (p *ConnectedPlayer) setShopTitle(title string)` | See the source definition and the subsystem ownership contract. |
| [`currentShopTitle`](../../server/internal/world/myshop.go#L123)<br>`func (p *ConnectedPlayer) currentShopTitle() string` | See the source definition and the subsystem ownership contract. |
| [`setViewedShop`](../../server/internal/world/myshop.go#L129)<br>`func (p *ConnectedPlayer) setViewedShop(seller *ConnectedPlayer)` | See the source definition and the subsystem ownership contract. |
| [`currentViewedShop`](../../server/internal/world/myshop.go#L135)<br>`func (p *ConnectedPlayer) currentViewedShop() *ConnectedPlayer` | See the source definition and the subsystem ownership contract. |
| [`sendShopUpdate`](../../server/internal/world/myshop.go#L142)<br>`func (p *ConnectedPlayer) sendShopUpdate(seller *ConnectedPlayer, frame []byte, closing bool)` | Order a seller's notifications against the buyer switching or leaving a shop. |
| [`closeMyShopView`](../../server/internal/world/myshop.go#L154)<br>`func (s *Session) closeMyShopView()` | See the source definition and the subsystem ownership contract. |
| [`tickMyShop`](../../server/internal/world/myshop.go#L167)<br>`func (s *Session) tickMyShop()` | See the source definition and the subsystem ownership contract. |
| [`canOpenMyShop`](../../server/internal/world/myshop.go#L184)<br>`func (s *Session) canOpenMyShop() int` | canOpenMyShop is the setup precondition. |
| [`handleMyShop`](../../server/internal/world/myshop.go#L198)<br>`func (s *Session) handleMyShop(op packet.Opcode, p []byte)` | handleMyShop covers the seller's requests 0x2301-0x2307 and the buyer's 0x2309, 0x230a and 0x230d. |
| [`openMyShop`](../../server/internal/world/myshop.go#L281)<br>`func (s *Session) openMyShop(p []byte)` | openMyShop is 0x2303: the market starts with its message and the seller sits. |
| [`cleanShopTitle`](../../server/internal/world/myshop.go#L304)<br>`func cleanShopTitle(title string) string` | See the source definition and the subsystem ownership contract. |
| [`stopMyShop`](../../server/internal/world/myshop.go#L314)<br>`func (s *Session) stopMyShop()` | stopMyShop goes back to setup: buyers are told, the bubble goes away. |
| [`closeMyShopBuyers`](../../server/internal/world/myshop.go#L324)<br>`func (s *Session) closeMyShopBuyers()` | See the source definition and the subsystem ownership contract. |
| [`canBrowseShop`](../../server/internal/world/myshop.go#L332)<br>`func (s *Session) canBrowseShop(seller *ConnectedPlayer) bool` | See the source definition and the subsystem ownership contract. |
| [`refreshMyShop`](../../server/internal/world/myshop.go#L344)<br>`func (s *Session) refreshMyShop() bool` | Reconcile offers when a stack changes or the original item leaves its slot. An empty, published market closes and disconnects its buyers. |
| [`endMyShop`](../../server/internal/world/myshop.go#L372)<br>`func (s *Session) endMyShop()` | endMyShop closes the market for good (stand up, move, die, leave...). |
| [`enterMyShop`](../../server/internal/world/myshop.go#L387)<br>`func (s *Session) enterMyShop(id uint32)` | enterMyShop is 0x2309: a buyer opens the seller's market. |
| [`myShopList`](../../server/internal/world/myshop.go#L420)<br>`func (s *Session) myShopList() []byte` | myShopList is 0x230b: count and, per listed item, its index, price and record. |
| [`buyFromMyShop`](../../server/internal/world/myshop.go#L447)<br>`func (s *Session) buyFromMyShop(idx, count uint8)` | buyFromMyShop is 0x230d: take count units of the listed item idx. |

<details><summary>Type definitions and owned fields</summary>

```go
type shopSlot struct {
	used  bool
	item  db.Item
	price uint32
}
```

```go
type myShop struct {
	status int
	slots  [myShopSlots]shopSlot
	buyers []*ConnectedPlayer
}
```

</details>

### notices.go

[server/internal/world/notices.go](../../server/internal/world/notices.go#L1)

Direct GX dependencies: `internal/catalogfiles`, `internal/proto`, `server/internal/logx`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`noticeCatalog`](../../server/internal/world/notices.go#L16)<br>`var noticeCatalog []catalogfiles.Notice` | See the source definition and the subsystem ownership contract. |
| [`loadServerNotices`](../../server/internal/world/notices.go#L18)<br>`func loadServerNotices(log *logx.Logger) error` | See the source definition and the subsystem ownership contract. |
| [`noticeHour`](../../server/internal/world/notices.go#L37)<br>`type noticeHour` | See the source definition and the subsystem ownership contract. |
| [`parseNoticeHour`](../../server/internal/world/notices.go#L39)<br>`func parseNoticeHour(text string) (noticeHour, bool)` | See the source definition and the subsystem ownership contract. |
| [`scheduledNotice`](../../server/internal/world/notices.go#L51)<br>`type scheduledNotice` | See the source definition and the subsystem ownership contract. |
| [`scheduleNotices`](../../server/internal/world/notices.go#L58)<br>`func scheduleNotices(rows []catalogfiles.Notice, start time.Time) []scheduledNotice` | See the source definition and the subsystem ownership contract. |
| [`due`](../../server/internal/world/notices.go#L81)<br>`func (n *scheduledNotice) due(now time.Time) bool` | See the source definition and the subsystem ownership contract. |
| [`validNoticeMessage`](../../server/internal/world/notices.go#L102)<br>`func validNoticeMessage(message string) bool` | See the source definition and the subsystem ownership contract. |
| [`broadcastNotice`](../../server/internal/world/notices.go#L107)<br>`func broadcastNotice(message string)` | See the source definition and the subsystem ownership contract. |
| [`gmNotice`](../../server/internal/world/notices.go#L116)<br>`func (s *Session) gmNotice(message string) error` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type noticeHour struct{ day, hour, minute int }
```

```go
type scheduledNotice struct {
	message  string
	interval time.Duration
	next     time.Time
	fixed    map[noticeHour]int // Last calendar date sent per weekly slot.
}
```

</details>

### npc_guard.go

[server/internal/world/npc_guard.go](../../server/internal/world/npc_guard.go#L1)

Direct GX dependencies: `internal/mapnav`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`npcTypeGuard`](../../server/internal/world/npc_guard.go#L12)<br>`const npcTypeGuard = 8` | See the source definition and the subsystem ownership contract. |
| [`guardSight`](../../server/internal/world/npc_guard.go#L13)<br>`const guardSight = 20` | See the source definition and the subsystem ownership contract. |
| [`guardLeash`](../../server/internal/world/npc_guard.go#L14)<br>`const guardLeash = 30` | See the source definition and the subsystem ownership contract. |
| [`guardContact`](../../server/internal/world/npc_guard.go#L15)<br>`const guardContact = 3` | See the source definition and the subsystem ownership contract. |
| [`deathTypeGuard`](../../server/internal/world/npc_guard.go#L16)<br>`const deathTypeGuard = 3` | See the source definition and the subsystem ownership contract. |
| [`guardTarget`](../../server/internal/world/npc_guard.go#L19)<br>`type guardTarget` | See the source definition and the subsystem ownership contract. |
| [`npcGuardStrike`](../../server/internal/world/npc_guard.go#L24)<br>`type npcGuardStrike` | See the source definition and the subsystem ownership contract. |
| [`npcTick`](../../server/internal/world/npc_guard.go#L33)<br>`type npcTick` | See the source definition and the subsystem ownership contract. |
| [`tickNpcs`](../../server/internal/world/npc_guard.go#L38)<br>`func tickNpcs(now time.Time, players []PlayerCombatSnapshot, limit int) npcTick` | See the source definition and the subsystem ownership contract. |
| [`point`](../../server/internal/world/npc_guard.go#L81)<br>`func (n *MapNpc) point() mapnav.Point` | See the source definition and the subsystem ownership contract. |
| [`findGuardTarget`](../../server/internal/world/npc_guard.go#L85)<br>`func (n *MapNpc) findGuardTarget(now time.Time, s *mapShard, players []PlayerCombatSnapshot, country uint8, home NpcPatrolPoint, locked guardTarget) (guardTarget, mapnav.Point, *MapMob)` | See the source definition and the subsystem ownership contract. |
| [`guardInterval`](../../server/internal/world/npc_guard.go#L126)<br>`func guardInterval(def *npcDef) time.Duration` | See the source definition and the subsystem ownership contract. |
| [`updateGuard`](../../server/internal/world/npc_guard.go#L133)<br>`func (n *MapNpc) updateGuard(now time.Time, mapID uint16, s *mapShard, players []PlayerCombatSnapshot, out *npcTick) bool` | See the source definition and the subsystem ownership contract. |
| [`guardChase`](../../server/internal/world/npc_guard.go#L174)<br>`func (n *MapNpc) guardChase(now time.Time, mapID uint16, s *mapShard, def *npcDef, home NpcPatrolPoint, target guardTarget, goal mapnav.Point, mob *MapMob, out *npcTick) bool` | See the source definition and the subsystem ownership contract. |
| [`guardMove`](../../server/internal/world/npc_guard.go#L200)<br>`func (n *MapNpc) guardMove(now time.Time, mapID uint16, nav *mapnav.Map, def *npcDef, goal mapnav.Point, reach float32, out *npcTick) bool` | See the source definition and the subsystem ownership contract. |
| [`npcDirection`](../../server/internal/world/npc_guard.go#L243)<br>`func npcDirection(dx, dz float32) uint16` | npcDirection is the 0x0E01 angle in whole degrees: the client turns it into the heading (-sin, cos) of angle*pi/180 (Game.exe 0x5659a0). |
| [`buildNpcAttack`](../../server/internal/world/npc_guard.go#L251)<br>`func buildNpcAttack(npcID, target uint32, mob bool, hp uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`dispatchNpcStrike`](../../server/internal/world/npc_guard.go#L264)<br>`func dispatchNpcStrike(hit npcGuardStrike, log logPrinter)` | See the source definition and the subsystem ownership contract. |
| [`applyGuardStrike`](../../server/internal/world/npc_guard.go#L279)<br>`func (s *Session) applyGuardStrike(hit npcGuardStrike) bool` | The session owns player death. Revalidate map and contact after crossing its mailbox so a teleport cannot apply a stale strike to a new position. |

<details><summary>Type definitions and owned fields</summary>

```go
type guardTarget struct {
	id  uint32
	mob bool
}
```

```go
type npcGuardStrike struct {
	mapID   uint16
	guardID uint32
	target  guardTarget
	mob     MapMob
	x, y, z float32
	hp      uint32
}
```

```go
type npcTick struct {
	moves   []entityMove
	strikes []npcGuardStrike
}
```

</details>

### npcquest.go

[server/internal/world/npcquest.go](../../server/internal/world/npcquest.go#L1)

Direct GX dependencies: `internal/catalogfiles`, `internal/npcquest`, `internal/proto`, `server/internal/db`, `server/internal/logx`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`npcTypeShop`](../../server/internal/world/npcquest.go#L21)<br>`const npcTypeShop = 1` | See the source definition and the subsystem ownership contract. |
| [`npcTypeGatekeeper`](../../server/internal/world/npcquest.go#L22)<br>`const npcTypeGatekeeper = 2` | See the source definition and the subsystem ownership contract. |
| [`npcTypeCount`](../../server/internal/world/npcquest.go#L23)<br>`const npcTypeCount = 13` | See the source definition and the subsystem ownership contract. |
| [`gatekeeperReach`](../../server/internal/world/npcquest.go#L24)<br>`const gatekeeperReach = 30.0` | See the source definition and the subsystem ownership contract. |
| [`npcGate`](../../server/internal/world/npcquest.go#L27)<br>`type npcGate` | See the source definition and the subsystem ownership contract. |
| [`npcDef`](../../server/internal/world/npcquest.go#L33)<br>`type npcDef` | See the source definition and the subsystem ownership contract. |
| [`npcDefs`](../../server/internal/world/npcquest.go#L46)<br>`var npcDefs [npcTypeCount][]npcDef` | See the source definition and the subsystem ownership contract. |
| [`loadNpcQuest`](../../server/internal/world/npcquest.go#L48)<br>`func loadNpcQuest(log *logx.Logger) error` | See the source definition and the subsystem ownership contract. |
| [`buildNPCQuest`](../../server/internal/world/npcquest.go#L61)<br>`func buildNPCQuest(c npcquest.Catalog) ([npcTypeCount][]npcDef, []questDef, map[[2]uint8]questItemLink)` | See the source definition and the subsystem ownership contract. |
| [`findNpcDef`](../../server/internal/world/npcquest.go#L114)<br>`func findNpcDef(npcType uint8, typeID uint16) *npcDef` | findNpcDef is GetNpc (0x0041d640): by type and row. |
| [`npcOnMap`](../../server/internal/world/npcquest.go#L121)<br>`func npcOnMap(mapID uint16, id uint32) (MapNpc, bool)` | See the source definition and the subsystem ownership contract. |
| [`buildShopBuyResult`](../../server/internal/world/npcquest.go#L130)<br>`func buildShopBuyResult(result, bag, slot, itemType, typeID uint8, count uint16, money uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`handleShopBuy`](../../server/internal/world/npcquest.go#L141)<br>`func (s *Session) handleShopBuy(p []byte)` | handleShopBuy (0x0702): npc id, item index, count; CUser::PacketShop (0x0046af10). Result 0 bought, 1 not enough gold, 2 no room. A stack must take the whole purchase; otherwise the first free slot is used. |
| [`planShopBuy`](../../server/internal/world/npcquest.go#L142)<br>`func (s *Session) planShopBuy(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`buildGatekeeperResult`](../../server/internal/world/npcquest.go#L224)<br>`func buildGatekeeperResult(ok bool, money uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`handleGatekeeper`](../../server/internal/world/npcquest.go#L234)<br>`func (s *Session) handleGatekeeper(p []byte)` | handleGatekeeper (0x020C): npc id, destination 0-2. A gatekeeper of the character's faction (or neutral) within 30 units moves it for the destination's fee. |
| [`npcUseMu`](../../server/internal/world/npcquest.go#L302)<br>`var npcUseMu sync.Mutex` | NPC usage (CUser::PacketNpc 0x00467510): 0x0E04 records the NPC a player is talking to; an NPC in use stops patrolling. |
| [`npcUsers`](../../server/internal/world/npcquest.go#L303)<br>`var npcUsers` | NPC usage (CUser::PacketNpc 0x00467510): 0x0E04 records the NPC a player is talking to; an NPC in use stops patrolling. |
| [`useNpc`](../../server/internal/world/npcquest.go#L306)<br>`func (s *Session) useNpc(id uint32)` | See the source definition and the subsystem ownership contract. |
| [`npcInUse`](../../server/internal/world/npcquest.go#L330)<br>`func npcInUse(id uint32) bool` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type npcGate struct {
	MapID   uint16
	X, Y, Z float32
	Money   uint32
}
```

```go
type npcDef struct {
	Type                   uint8
	TypeID                 uint16
	Country                uint8 // 0 light, 1 fury, 2 both
	Name                   string
	ShopType               uint8
	ShopItems              [][2]uint8
	StartQuests, EndQuests []uint16
	// stNpcInfo +0x08 / +0x0C: patrol step and step interval (ms).
	MoveDistance, MoveSpeed int32
	Gates                   [3]npcGate
}
```

</details>

### oj_limits.go

[server/internal/world/oj_limits.go](../../server/internal/world/oj_limits.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`equipmentOJLimits`](../../server/internal/world/oj_limits.go#L10)<br>`func equipmentOJLimits(defs *db.Defs) []proto.OJLimit` | See the source definition and the subsystem ownership contract. |
| [`sendOJLimits`](../../server/internal/world/oj_limits.go#L32)<br>`func (s *Session) sendOJLimits()` | See the source definition and the subsystem ownership contract. |

### oxygen.go

[server/internal/world/oxygen.go](../../server/internal/world/oxygen.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`drowningInterval`](../../server/internal/world/oxygen.go#L10)<br>`const drowningInterval = 2 * time.Second` | See the source definition and the subsystem ownership contract. |
| [`oxygenState`](../../server/internal/world/oxygen.go#L12)<br>`type oxygenState` | See the source definition and the subsystem ownership contract. |
| [`sendOxygen`](../../server/internal/world/oxygen.go#L19)<br>`func (s *Session) sendOxygen()` | See the source definition and the subsystem ownership contract. |
| [`clearOxygen`](../../server/internal/world/oxygen.go#L27)<br>`func (s *Session) clearOxygen()` | See the source definition and the subsystem ownership contract. |
| [`tickOxygen`](../../server/internal/world/oxygen.go#L34)<br>`func (s *Session) tickOxygen(now time.Time)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type oxygenState struct {
	phase      proto.OxygenPhase
	remaining  time.Duration
	at         time.Time
	nextDamage time.Time
}
```

</details>

### party.go

[server/internal/world/party.go](../../server/internal/world/party.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`partyMax`](../../server/internal/world/party.go#L24)<br>`const partyMax = 7` | See the source definition and the subsystem ownership contract. |
| [`partyInviteWindow`](../../server/internal/world/party.go#L25)<br>`const partyInviteWindow = 30 * time.Second` | See the source definition and the subsystem ownership contract. |
| [`partyRewardRange`](../../server/internal/world/party.go#L26)<br>`const partyRewardRange = 100` | See the source definition and the subsystem ownership contract. |
| [`partyErrDeclined`](../../server/internal/world/party.go#L28)<br>`const partyErrDeclined = 0` | See the source definition and the subsystem ownership contract. |
| [`partyErrBusy`](../../server/internal/world/party.go#L29)<br>`const partyErrBusy = 1` | See the source definition and the subsystem ownership contract. |
| [`partyErrNotLead`](../../server/internal/world/party.go#L30)<br>`const partyErrNotLead = 2` | See the source definition and the subsystem ownership contract. |
| [`partyErrInParty`](../../server/internal/world/party.go#L31)<br>`const partyErrInParty = 3` | See the source definition and the subsystem ownership contract. |
| [`partyErrNoTarget`](../../server/internal/world/party.go#L32)<br>`const partyErrNoTarget = 4` | See the source definition and the subsystem ownership contract. |
| [`partyErrSame`](../../server/internal/world/party.go#L33)<br>`const partyErrSame = 5` | See the source definition and the subsystem ownership contract. |
| [`partyExpFactor`](../../server/internal/world/party.go#L37)<br>`var partyExpFactor` | partyExpFactor is MobDeathExperience's party table by member count. |
| [`party`](../../server/internal/world/party.go#L39)<br>`type party` | See the source definition and the subsystem ownership contract. |
| [`partyMu`](../../server/internal/world/party.go#L57)<br>`var partyMu sync.Mutex` | See the source definition and the subsystem ownership contract. |
| [`nextPartyID`](../../server/internal/world/party.go#L58)<br>`var nextPartyID uint32` | See the source definition and the subsystem ownership contract. |
| [`partyOf`](../../server/internal/world/party.go#L62)<br>`func partyOf(p *ConnectedPlayer) *party` | partyOf returns the player's party (under partyMu). |
| [`leaderPlayer`](../../server/internal/world/party.go#L64)<br>`func (pt *party) leaderPlayer() *ConnectedPlayer` | See the source definition and the subsystem ownership contract. |
| [`refreshLevel`](../../server/internal/world/party.go#L72)<br>`func (pt *party) refreshLevel()` | refreshLevel is PartyRuntime::check_experience_level. |
| [`send`](../../server/internal/world/party.go#L82)<br>`func (pt *party) send(frame []byte, label string)` | See the source definition and the subsystem ownership contract. |
| [`writePartyMember`](../../server/internal/world/party.go#L91)<br>`func writePartyMember(w *packet.Writer, p *ConnectedPlayer)` | writePartyMember is the roster/join record: id, name, level, job, HP/max, MP/max, SP/max (u32 in the 5.4 client), instance, position and active effects (none sent). |
| [`partyMove`](../../server/internal/world/party.go#L136)<br>`func (s *Session) partyMove(everyone ...bool)` | partyMove is CParty::SendMove (0x445880): after every move the position goes to the other party/raid members on the same map as 0x0C07 / 0x0C12, so their radar shows allies that are out of view. After a map change everyone hears it once, so members left behind drop the stale marker. |
| [`partyEffect`](../../server/internal/world/party.go#L170)<br>`func (s *Session) partyEffect(add bool, skillID uint16, level uint8)` | partyEffect is CUser::SendSkillApply / SendSkillRemoveParty: an applied or ended skill goes to the whole party (0x0C04/0x0C05) or raid (0x0C0E/0x0C10) as character id, skill and level. |
| [`partyIDPacket`](../../server/internal/world/party.go#L201)<br>`func partyIDPacket(op packet.Opcode, id uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`partyError`](../../server/internal/world/party.go#L207)<br>`func partyError(to *ConnectedPlayer, code uint8, id uint32)` | See the source definition and the subsystem ownership contract. |
| [`partyBadge`](../../server/internal/world/party.go#L216)<br>`func partyBadge(p *ConnectedPlayer)` | partyBadge (0x0520) tells viewers whether the character is in a party (1) or leads one (2). |
| [`enterParty`](../../server/internal/world/party.go#L234)<br>`func enterParty(pt *party, p *ConnectedPlayer) bool` | enterParty is enter_normal_party (partyMu held): announce the newcomer, send it the roster of the members already there. |
| [`leaveParty`](../../server/internal/world/party.go#L263)<br>`func leaveParty(pt *party, p *ConnectedPlayer, expelled bool)` | leaveParty is leave_normal_party (partyMu held). expelled sends 0x0B06 instead of 0x0B05. A party left with one member dissolves. |
| [`partyLeaveOnDisconnect`](../../server/internal/world/party.go#L305)<br>`func partyLeaveOnDisconnect(p *ConnectedPlayer)` | partyLeaveOnDisconnect drops a leaving player from its party and from any pending invitation. |
| [`handlePartyInvite`](../../server/internal/world/party.go#L315)<br>`func (s *Session) handlePartyInvite(op packet.Opcode, p []byte)` | handlePartyInvite (0x0B02 id / 0x0B09 name). |
| [`handlePartyInviteGlobal`](../../server/internal/world/party.go#L350)<br>`func (s *Session) handlePartyInviteGlobal(p []byte)` | handlePartyInviteGlobal is the GX /party <name> request (0x7F15): the same invitation as 0x0B09 without the same-map requirement. Offline players, other factions and, unless the inviter is a GM too, GMs all answer with the generic "unavailable" reply, so a hidden GM cannot be told apart from an offline character. |
| [`partyInviteTarget`](../../server/internal/world/party.go#L365)<br>`func (s *Session) partyInviteTarget(me, target *ConnectedPlayer, targetID uint32, op packet.Opcode, anyMap bool)` | partyInviteTarget validates and delivers an invitation. anyMap lets the target answer from another map (the invitation was global). |
| [`handlePartyAnswer`](../../server/internal/world/party.go#L405)<br>`func (s *Session) handlePartyAnswer(p []byte)` | handlePartyAnswer (0x0B03): 0 accepts. The inviter's party is created on the first acceptance. |
| [`handlePartyLeave`](../../server/internal/world/party.go#L439)<br>`func (s *Session) handlePartyLeave()` | See the source definition and the subsystem ownership contract. |
| [`handlePartyExpel`](../../server/internal/world/party.go#L451)<br>`func (s *Session) handlePartyExpel(p []byte)` | handlePartyExpel (0x0B06) and handlePartyLeader (0x0B07): leader only. |
| [`handlePartyLeader`](../../server/internal/world/party.go#L470)<br>`func (s *Session) handlePartyLeader(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`partyPools`](../../server/internal/world/party.go#L494)<br>`func (s *Session) partyPools()` | partyPools (0x0C03) shares the character's HP/SP/MP with its party. |
| [`partyLevel`](../../server/internal/world/party.go#L522)<br>`func (s *Session) partyLevel()` | partyLevel handles the original 0x0C09/0x0C13 level notices for party and raid. |
| [`partyChat`](../../server/internal/world/party.go#L542)<br>`func (s *Session) partyChat(message string, items proto.ChatItems)` | partyChat (0x1105) relays a message to the party. |
| [`raidChat`](../../server/internal/world/party.go#L559)<br>`func (s *Session) raidChat(message string)` | raidChat is the 0x1112 notice (client writer 0x59ca70, shown as "[Union] text"): only the leader or the sub-leader of a raid may send it, and every member of the raid receives it, whatever the map. |
| [`partyExpLevelShare`](../../server/internal/world/party.go#L579)<br>`func partyExpLevelShare(maxLevel, level uint16) float64` | partyExpLevelShare is the level gap factor against the party's highest level. |
| [`killReward`](../../server/internal/world/party.go#L593)<br>`type killReward` | See the source definition and the subsystem ownership contract. |
| [`perfect`](../../server/internal/world/party.go#L598)<br>`func (pt *party) perfect() bool` | See the source definition and the subsystem ownership contract. |
| [`perfectParty`](../../server/internal/world/party.go#L600)<br>`func (s *Session) perfectParty() bool` | See the source definition and the subsystem ownership contract. |
| [`killRewards`](../../server/internal/world/party.go#L611)<br>`func killRewards(killer *ConnectedPlayer, mapID, mobLevel, mobExp uint16) ([]killReward, bool)` | killRewards uses the credited player's position and the full party size. Match CMob::Death's two truncations before handing recipients their EXP. |
| [`partyMembersNear`](../../server/internal/world/party.go#L651)<br>`func partyMembersNear(killer *ConnectedPlayer, mapID uint16, x, z float32) []*ConnectedPlayer` | partyMembersNear lists the killer's party members on the map within the reward range (nil without a party). |
| [`samePartyIDs`](../../server/internal/world/party.go#L670)<br>`func samePartyIDs(a, b uint32) bool` | samePartyIDs reports whether two characters share a party. |

<details><summary>Type definitions and owned fields</summary>

```go
type party struct {
	id       uint32
	members  []*ConnectedPlayer
	leader   int
	maxLevel uint16

	// Raid state (kind 1), see raid.go.
	kind         uint8
	subLeader    int
	slots        []uint16 // wire slot per member
	autoJoin     bool
	itemDivision int32
	itemDivSeq   int // m_nItemDivSeq, the rotation for divisions 1 and 3

	loot map[uint32]*lootEntry // universal loot rounds, see loot.go
}
```

```go
type killReward struct {
	member *ConnectedPlayer
	exp    uint32
}
```

</details>

### party_assist.go

[server/internal/world/party_assist.go](../../server/internal/world/party_assist.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`handlePartyAssist`](../../server/internal/world/party_assist.go#L8)<br>`func (s *Session) handlePartyAssist(frame []byte)` | See the source definition and the subsystem ownership contract. |
| [`knowsAssistTarget`](../../server/internal/world/party_assist.go#L41)<br>`func (p *ConnectedPlayer) knowsAssistTarget(id uint32) bool` | See the source definition and the subsystem ownership contract. |

### partyzone.go

[server/internal/world/partyzone.go](../../server/internal/world/partyzone.go#L1)

Direct GX dependencies: `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`portalClosed`](../../server/internal/world/partyzone.go#L25)<br>`const portalClosed = 9` | See the source definition and the subsystem ownership contract. |
| [`portalNoParty`](../../server/internal/world/partyzone.go#L26)<br>`const portalNoParty = 4` | See the source definition and the subsystem ownership contract. |
| [`portalPartySmall`](../../server/internal/world/partyzone.go#L27)<br>`const portalPartySmall = 10` | See the source definition and the subsystem ownership contract. |
| [`portalPartyFull`](../../server/internal/world/partyzone.go#L28)<br>`const portalPartyFull = 11` | See the source definition and the subsystem ownership contract. |
| [`partyCreateType`](../../server/internal/world/partyzone.go#L29)<br>`const partyCreateType = "P"` | See the source definition and the subsystem ownership contract. |
| [`houseCreateType`](../../server/internal/world/partyzone.go#L30)<br>`const houseCreateType = "H"` | See the source definition and the subsystem ownership contract. |
| [`instanceSweepTick`](../../server/internal/world/partyzone.go#L31)<br>`const instanceSweepTick = time.Second` | See the source definition and the subsystem ownership contract. |
| [`partyOpen`](../../server/internal/world/partyzone.go#L35)<br>`func partyOpen(z ZoneConfig, now time.Time) bool` | partyOpen is IsPartyInsZoneOpenTime. |
| [`findInstance`](../../server/internal/world/partyzone.go#L40)<br>`func findInstance(mapID uint16, owner uint32) (uint16, bool)` | findInstance is the zone of mapID owned by owner. |
| [`partyOfPlayer`](../../server/internal/world/partyzone.go#L52)<br>`func partyOfPlayer(p *ConnectedPlayer) (uint32, int)` | partyID returns the user's party id and member count (0 without one). |
| [`portalGate`](../../server/internal/world/partyzone.go#L63)<br>`func (s *Session) portalGate(target uint16) uint8` | portalGate is the 0x020A checks for an instanced target map; 0 lets the user through. |
| [`buildPortalResult`](../../server/internal/world/partyzone.go#L84)<br>`func buildPortalResult(index, result uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`enterPartyInstance`](../../server/internal/world/partyzone.go#L92)<br>`func (s *Session) enterPartyInstance(mapID uint16) (uint16, bool)` | enterPartyInstance is GetZoneByMapID's party case. |
| [`instanceExpiryMu`](../../server/internal/world/partyzone.go#L110)<br>`var instanceExpiryMu sync.Mutex` | See the source definition and the subsystem ownership contract. |
| [`instanceExpiry`](../../server/internal/world/partyzone.go#L111)<br>`var instanceExpiry` | See the source definition and the subsystem ownership contract. |
| [`TickInstances`](../../server/internal/world/partyzone.go#L116)<br>`func TickInstances(now time.Time, log logPrinter)` | TickInstances is CWorldThread::Update's instance expiry. The GRB owns its instances until the battle closes, even when the entire guild leaves. |

### periodic_healing.go

[server/internal/world/periodic_healing.go](../../server/internal/world/periodic_healing.go#L1)

Direct GX dependencies: `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`detailPeriodicHeal`](../../server/internal/world/periodic_healing.go#L10)<br>`const detailPeriodicHeal = 81` | See the source definition and the subsystem ownership contract. |
| [`periodicHealTick`](../../server/internal/world/periodic_healing.go#L11)<br>`const periodicHealTick = 3 * time.Second` | See the source definition and the subsystem ownership contract. |
| [`tickPeriodicHeal`](../../server/internal/world/periodic_healing.go#L16)<br>`func (s *Session) tickPeriodicHeal(b *ActiveBuff, now time.Time)` | CUser::ChkKeepSkill applies detail 81 to the buff holder. The catalog's TimeHeal values are per three seconds and do not scale with caster stats. |
| [`periodicHealAmount`](../../server/internal/world/periodic_healing.go#L55)<br>`func periodicHealAmount(perTick uint32, elapsed time.Duration) uint32` | See the source definition and the subsystem ownership contract. |

### player_feedback.go

[server/internal/world/player_feedback.go](../../server/internal/world/player_feedback.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`sendPlayerFeedback`](../../server/internal/world/player_feedback.go#L8)<br>`func (s *Session) sendPlayerFeedback(kind uint8)` | See the source definition and the subsystem ownership contract. |

### player_forms.go

[server/internal/world/player_forms.go](../../server/internal/world/player_forms.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`formLook`](../../server/internal/world/player_forms.go#L12)<br>`type formLook` | See the source definition and the subsystem ownership contract. |
| [`disguiseLook`](../../server/internal/world/player_forms.go#L19)<br>`type disguiseLook` | See the source definition and the subsystem ownership contract. |
| [`formOf`](../../server/internal/world/player_forms.go#L21)<br>`func formOf(def *db.SkillDef) proto.PlayerFormKind` | See the source definition and the subsystem ownership contract. |
| [`copyFormTarget`](../../server/internal/world/player_forms.go#L25)<br>`func (s *Session) copyFormTarget(def *db.SkillDef, target uint32, completed bool) (formLook, bool)` | See the source definition and the subsystem ownership contract. |
| [`castCopiedForm`](../../server/internal/world/player_forms.go#L64)<br>`func (s *Session) castCopiedForm(op packet.Opcode, k db.CharacterSkill, def *db.SkillDef, target uint32, stats CharacterStats, completed bool)` | See the source definition and the subsystem ownership contract. |
| [`rejectCopyForm`](../../server/internal/world/player_forms.go#L82)<br>`func (s *Session) rejectCopyForm(op packet.Opcode, k db.CharacterSkill)` | See the source definition and the subsystem ownership contract. |
| [`activeForm`](../../server/internal/world/player_forms.go#L86)<br>`func (s *Session) activeForm() formLook` | See the source definition and the subsystem ownership contract. |
| [`breakCombatForm`](../../server/internal/world/player_forms.go#L96)<br>`func (s *Session) breakCombatForm()` | See the source definition and the subsystem ownership contract. |
| [`breakableForm`](../../server/internal/world/player_forms.go#L106)<br>`func breakableForm(b *ActiveBuff) bool` | See the source definition and the subsystem ownership contract. |
| [`buildFormAppearance`](../../server/internal/world/player_forms.go#L110)<br>`func buildFormAppearance(id uint32, look formLook) []byte` | See the source definition and the subsystem ownership contract. |
| [`publishForm`](../../server/internal/world/player_forms.go#L136)<br>`func (s *Session) publishForm(old, next formLook)` | See the source definition and the subsystem ownership contract. |
| [`sendFormFrame`](../../server/internal/world/player_forms.go#L150)<br>`func (s *Session) sendFormFrame(frame []byte)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type formLook struct {
	kind       proto.PlayerFormKind
	mob, skill uint16
	rank       uint8
	disguise   *disguiseLook
}
```

```go
type disguiseLook struct{ shape []byte }
```

</details>

### player_movement.go

[server/internal/world/player_movement.go](../../server/internal/world/player_movement.go#L1)

Direct GX dependencies: `internal/mapgeometry`, `internal/movespeed`, `internal/proto`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`playerMovement`](../../server/internal/world/player_movement.go#L15)<br>`type playerMovement` | See the source definition and the subsystem ownership contract. |
| [`playerGeometry`](../../server/internal/world/player_movement.go#L34)<br>`func (s *Session) playerGeometry(mapID uint16) (*mapgeometry.Geometry, error)` | See the source definition and the subsystem ownership contract. |
| [`initPlayerMovement`](../../server/internal/world/player_movement.go#L51)<br>`func (s *Session) initPlayerMovement(now time.Time) error` | See the source definition and the subsystem ownership contract. |
| [`movementSpeed`](../../server/internal/world/player_movement.go#L77)<br>`func (s *Session) movementSpeed() float32` | See the source definition and the subsystem ownership contract. |
| [`startMovementJump`](../../server/internal/world/player_movement.go#L91)<br>`func (s *Session) startMovementJump(now time.Time) bool` | See the source definition and the subsystem ownership contract. |
| [`advancePlayerVertical`](../../server/internal/world/player_movement.go#L114)<br>`func (s *Session) advancePlayerVertical(now time.Time) bool` | advancePlayerVertical uses server time and geometry. Client Y is never an integration input; a fake jump flag cannot grant another impulse in mid-air. |
| [`recoverPlayerHeight`](../../server/internal/world/player_movement.go#L229)<br>`func (s *Session) recoverPlayerHeight() bool` | Recover only from server-simulated positions. A client height sample cannot request a free teleport. |
| [`tickPlayerMovement`](../../server/internal/world/player_movement.go#L246)<br>`func (s *Session) tickPlayerMovement(now time.Time)` | See the source definition and the subsystem ownership contract. |
| [`commitMovement`](../../server/internal/world/player_movement.go#L282)<br>`func (s *Session) commitMovement()` | See the source definition and the subsystem ownership contract. |
| [`correctMovement`](../../server/internal/world/player_movement.go#L287)<br>`func (s *Session) correctMovement(reason string)` | See the source definition and the subsystem ownership contract. |
| [`heightRecovery`](../../server/internal/world/player_movement.go#L306)<br>`const heightRecovery = "height recovery"` | See the source definition and the subsystem ownership contract. |
| [`validatePlayerMovement`](../../server/internal/world/player_movement.go#L308)<br>`func (s *Session) validatePlayerMovement(x, y, z float32, motion uint8, direction uint16, now time.Time) (bool, string)` | See the source definition and the subsystem ownership contract. |
| [`sendPhysicsClock`](../../server/internal/world/player_movement.go#L516)<br>`func (s *Session) sendPhysicsClock()` | See the source definition and the subsystem ownership contract. |
| [`groundTrustedDestination`](../../server/internal/world/player_movement.go#L528)<br>`func (s *Session) groundTrustedDestination(x, y, z float32) (mapgeometry.Position, bool)` | groundTrustedDestination normalizes a server-authored same-map relocation. A live selected session always has geometry; unselected sessions have none. |

<details><summary>Type definitions and owned fields</summary>

```go
type playerMovement struct {
	ladder                         mapgeometry.LadderMotion
	ladderAt                       time.Time
	ladderRetryAt                  time.Time
	slope                          mapgeometry.SlopeMotion
	geometry                       *mapgeometry.Geometry
	mapID                          uint16
	pos                            mapgeometry.Position
	at, packetAt, budgetAt, jumpAt time.Time
	credit                         float64
	speedCorrectionPending         bool
	jumping                        bool
	jumpFall, fall, waterFall      float32
	motion                         uint8
	direction                      uint16
	fallPeak                       float32
	fallTracking, fallReady        bool
}
```

</details>

### playereffects.go

[server/internal/world/playereffects.go](../../server/internal/world/playereffects.go#L1)

Direct GX dependencies: `internal/attackspeed`, `internal/movespeed`, `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`addEffect`](../../server/internal/world/playereffects.go#L26)<br>`func (s *Session) addEffect(def *db.SkillDef, casterMob uint32, sourceUser ...uint32) (ActiveBuff, bool)` | addEffect starts (or replaces) an applied skill and sends 0x050D. The same skill at a lower level does not replace a higher one. |
| [`canAddEffect`](../../server/internal/world/playereffects.go#L95)<br>`func (s *Session) canAddEffect(def *db.SkillDef, sourceUser ...uint32) bool` | canAddEffect also guards item consumption, so a refused lower-level buff cannot consume a candy or start its cooldown. |
| [`buildBuffAdd`](../../server/internal/world/playereffects.go#L117)<br>`func buildBuffAdd(id uint32, skillID uint16, level uint8) []byte` | buildBuffAdd (0x050D): instance id, skill id, level (the client reads nothing more). |
| [`playerSpeeds`](../../server/internal/world/playereffects.go#L125)<br>`type playerSpeeds` | See the source definition and the subsystem ownership contract. |
| [`blocks`](../../server/internal/world/playereffects.go#L128)<br>`func (s *Session) blocks() (actionA, move, actionB int)` | blocks counts the action/movement blocks (details 64, 65, 66). |
| [`actionBlocked`](../../server/internal/world/playereffects.go#L146)<br>`func (s *Session) actionBlocked() bool` | actionBlocked: no attacks or skills while stunned or asleep. |
| [`speeds`](../../server/internal/world/playereffects.go#L152)<br>`func (s *Session) speeds() playerSpeeds` | speeds is CUser::GetAttackSpeed / GetMoveSpeed: 255 while blocked. |
| [`buildCharacterSpeed`](../../server/internal/world/playereffects.go#L175)<br>`func buildCharacterSpeed(id uint32, sp playerSpeeds) []byte` | buildCharacterSpeed encodes 0x051C (character id, attack speed, move speed). |
| [`syncPlayerSpeed`](../../server/internal/world/playereffects.go#L185)<br>`func (s *Session) syncPlayerSpeed(force bool)` | syncPlayerSpeed publishes only server-derived state. Enter-view snapshots carry the same value, including equipment and restored consumable buffs. |
| [`broadcastSpeed`](../../server/internal/world/playereffects.go#L195)<br>`func (s *Session) broadcastSpeed(sp playerSpeeds)` | See the source definition and the subsystem ownership contract. |
| [`removeEffects`](../../server/internal/world/playereffects.go#L219)<br>`func (s *Session) removeEffects(match func(*ActiveBuff) bool, why string)` | removeEffects drops the matching applied skills with 0x050E and refreshes the stats and speeds. |
| [`removeEffectsInternal`](../../server/internal/world/playereffects.go#L223)<br>`func (s *Session) removeEffectsInternal(match func(*ActiveBuff) bool, why string, includeMapCurses bool)` | See the source definition and the subsystem ownership contract. |
| [`wake`](../../server/internal/world/playereffects.go#L255)<br>`func (s *Session) wake()` | wake is CUser::RemApplySkillSleep: damage ends every sleep (detail 66) and every transformation (Degeneration, detail 90). |
| [`applyMobDebuff`](../../server/internal/world/playereffects.go#L262)<br>`func (s *Session) applyMobDebuff(sk *db.SkillDef, mobGlobalID uint32)` | applyMobDebuff leaves a mob skill (TypeEffect 3) on the player. |
| [`playerTimeDamage`](../../server/internal/world/playereffects.go#L274)<br>`func playerTimeDamage(def *db.SkillDef, st *CharacterStats, elapsed time.Duration) [3]int32` | playerTimeDamage is CUser::TimeDamage (0x00486290) on the player. |
| [`tickDebuffs`](../../server/internal/world/playereffects.go#L282)<br>`func (s *Session) tickDebuffs(now time.Time) bool` | tickDebuffs runs the damage of applied debuffs: detail 61 every 3 s and on expiry, detail 63 on expiry. Returns whether the player died. |
| [`blocksAttack`](../../server/internal/world/playereffects.go#L381)<br>`func (s *Session) blocksAttack(kind int) bool` | blocksAttack is the defender's block check before CMob::AttackSuccess: detail 92 blocks everything; 85 blocks DefenceValue attacks of its category and then ends; 86 blocks with a DefenceValue% chance. |
| [`deathExpLoss`](../../server/internal/world/playereffects.go#L412)<br>`func (s *Session) deathExpLoss() int` | deathExpLoss honors the dedicated EXP loss protection ability. |
| [`reflects`](../../server/internal/world/playereffects.go#L428)<br>`func (s *Session) reflects(kind int, md *db.MobDef) bool` | reflects is CUser::ChkMirror (0x0044e080) with a detail-83 buff covering the category (its AbilityValue for that category is non-zero): SuccessType 1 rolls SuccessValue%, 2 rolls LUC x 10 / SuccessValue and 3 WIS x 10 / SuccessValue (CUser +0x1234 / +0x122c). A mob resisting the melee mirror skill cannot be reflected; md is nil for a user attacker. |
| [`mirrorDamage`](../../server/internal/world/playereffects.go#L477)<br>`func (s *Session) mirrorDamage(userID, mobID uint32, d [3]uint32)` | mirrorDamage returns a mirrored hit to its source: a user takes it as the 0x0515 mirrored hit, a mob through 0x051B. |
| [`reflectToMob`](../../server/internal/world/playereffects.go#L490)<br>`func (s *Session) reflectToMob(mobGlobalID uint32, damage [3]uint16)` | reflectToMob sends a reflected hit back to the mob (0x051B: character, mob, HP/SP/MP). |

<details><summary>Type definitions and owned fields</summary>

```go
type playerSpeeds struct{ attack, move uint8 }
```

</details>

### players.go

[server/internal/world/players.go](../../server/internal/world/players.go#L1)

Direct GX dependencies: `server/internal/crypto`, `server/internal/db`, `server/internal/logx`, `server/internal/netx`, `server/internal/telemetry`.

| Declaration | Responsibility / source comment |
|---|---|
| [`connWriter`](../../server/internal/world/players.go#L20)<br>`type connWriter` | connWriter serializes every write to one client socket: the owning session and broadcasts from other sessions share it. |
| [`write`](../../server/internal/world/players.go#L26)<br>`func (w *connWriter) write(b []byte) error` | See the source definition and the subsystem ownership contract. |
| [`pendingTeleport`](../../server/internal/world/players.go#L37)<br>`type pendingTeleport` | See the source definition and the subsystem ownership contract. |
| [`pendingDamage`](../../server/internal/world/players.go#L42)<br>`type pendingDamage` | See the source definition and the subsystem ownership contract. |
| [`ConnectedPlayer`](../../server/internal/world/players.go#L56)<br>`type ConnectedPlayer` | ConnectedPlayer is a player in the world, reachable from any session for broadcasts. Cross-session work is queued in mailboxes and applied by the owning session's loop. Paired trades hold both owners' state locks. |
| [`QueueQuestKill`](../../server/internal/world/players.go#L141)<br>`func (p *ConnectedPlayer) QueueQuestKill(mobID uint16)` | QueueQuestKill credits a kill to the player's quests (party members get it too, as mob_death does). |
| [`pendingQuestKills`](../../server/internal/world/players.go#L162)<br>`func (p *ConnectedPlayer) pendingQuestKills() map[uint16]uint8` | See the source definition and the subsystem ownership contract. |
| [`ackQuestKills`](../../server/internal/world/players.go#L171)<br>`func (p *ConnectedPlayer) ackQuestKills(id uint16, count uint8)` | See the source definition and the subsystem ownership contract. |
| [`setQuests`](../../server/internal/world/players.go#L181)<br>`func (p *ConnectedPlayer) setQuests(ids []uint16)` | See the source definition and the subsystem ownership contract. |
| [`hasQuest`](../../server/internal/world/players.go#L191)<br>`func (p *ConnectedPlayer) hasQuest(id uint16) bool` | See the source definition and the subsystem ownership contract. |
| [`perfectExperience`](../../server/internal/world/players.go#L197)<br>`type perfectExperience` | See the source definition and the subsystem ownership contract. |
| [`rewardSnapshot`](../../server/internal/world/players.go#L202)<br>`type rewardSnapshot` | See the source definition and the subsystem ownership contract. |
| [`experienceBatch`](../../server/internal/world/players.go#L210)<br>`type experienceBatch` | An award's integer hundreds and remainder can be aggregated independently. This keeps ordinary rewards bounded while retaining rounding per kill for any bonus. |
| [`experienceScaled`](../../server/internal/world/players.go#L215)<br>`func experienceScaled(amount uint32, bonus uint32) uint32` | See the source definition and the subsystem ownership contract. |
| [`add`](../../server/internal/world/players.go#L219)<br>`func (b *experienceBatch) add(amount uint32)` | See the source definition and the subsystem ownership contract. |
| [`scaled`](../../server/internal/world/players.go#L226)<br>`func (b *experienceBatch) scaled(bonus uint32) uint32` | See the source definition and the subsystem ownership contract. |
| [`subtract`](../../server/internal/world/players.go#L243)<br>`func (b *experienceBatch) subtract(credited experienceBatch)` | See the source definition and the subsystem ownership contract. |
| [`QueueExp`](../../server/internal/world/players.go#L251)<br>`func (p *ConnectedPlayer) QueueExp(amount uint32, perfect bool)` | QueueExp carries a kill reward to its recipient's owning session. |
| [`QueueGold`](../../server/internal/world/players.go#L262)<br>`func (p *ConnectedPlayer) QueueGold(amount uint32)` | See the source definition and the subsystem ownership contract. |
| [`takeRewardTotals`](../../server/internal/world/players.go#L267)<br>`func (p *ConnectedPlayer) takeRewardTotals(bonus uint32) (uint32, uint32)` | See the source definition and the subsystem ownership contract. |
| [`peekRewardTotals`](../../server/internal/world/players.go#L279)<br>`func (p *ConnectedPlayer) peekRewardTotals() rewardSnapshot` | See the source definition and the subsystem ownership contract. |
| [`ackRewardTotals`](../../server/internal/world/players.go#L284)<br>`func (p *ConnectedPlayer) ackRewardTotals(snapshot rewardSnapshot)` | See the source definition and the subsystem ownership contract. |
| [`TakeRewards`](../../server/internal/world/players.go#L297)<br>`func (p *ConnectedPlayer) TakeRewards() (exp, gold []uint32)` | See the source definition and the subsystem ownership contract. |
| [`maxPendingCombat`](../../server/internal/world/players.go#L308)<br>`const maxPendingCombat = 4096` | See the source definition and the subsystem ownership contract. |
| [`maxPendingItemGrants`](../../server/internal/world/players.go#L309)<br>`const maxPendingItemGrants = 256` | See the source definition and the subsystem ownership contract. |
| [`mailboxOverflows`](../../server/internal/world/players.go#L311)<br>`var mailboxOverflows` | See the source definition and the subsystem ownership contract. |
| [`mailboxOverflow`](../../server/internal/world/players.go#L313)<br>`func (p *ConnectedPlayer) mailboxOverflow()` | See the source definition and the subsystem ownership contract. |
| [`bucket`](../../server/internal/world/players.go#L322)<br>`func bucket(mapID uint16) map[uint32]*ConnectedPlayer` | See the source definition and the subsystem ownership contract. |
| [`RegisterPlayer`](../../server/internal/world/players.go#L332)<br>`func RegisterPlayer(ch db.Character, items []db.Item, isAdmin bool, faction uint8, out *connWriter, log *logx.Logger, owner ...*Session) *ConnectedPlayer` | RegisterPlayer adds a player to its map's realm.registry. |
| [`Unregister`](../../server/internal/world/players.go#L346)<br>`func (p *ConnectedPlayer) Unregister()` | Unregister removes the player and cancels any trade it was part of. |
| [`SetMap`](../../server/internal/world/players.go#L361)<br>`func (p *ConnectedPlayer) SetMap(mapID uint16)` | See the source definition and the subsystem ownership contract. |
| [`MapID`](../../server/internal/world/players.go#L380)<br>`func (p *ConnectedPlayer) MapID() uint16` | See the source definition and the subsystem ownership contract. |
| [`Faction`](../../server/internal/world/players.go#L386)<br>`func (p *ConnectedPlayer) Faction() uint8` | See the source definition and the subsystem ownership contract. |
| [`PositionXZ`](../../server/internal/world/players.go#L392)<br>`func (p *ConnectedPlayer) PositionXZ() (float32, float32)` | See the source definition and the subsystem ownership contract. |
| [`UpdateCharacter`](../../server/internal/world/players.go#L398)<br>`func (p *ConnectedPlayer) UpdateCharacter(ch db.Character)` | See the source definition and the subsystem ownership contract. |
| [`Snapshot`](../../server/internal/world/players.go#L404)<br>`func (p *ConnectedPlayer) Snapshot() db.Character` | See the source definition and the subsystem ownership contract. |
| [`UpdateItems`](../../server/internal/world/players.go#L410)<br>`func (p *ConnectedPlayer) UpdateItems(items []db.Item)` | See the source definition and the subsystem ownership contract. |
| [`UpdateMotion`](../../server/internal/world/players.go#L416)<br>`func (p *ConnectedPlayer) UpdateMotion(motion uint8, direction uint16)` | See the source definition and the subsystem ownership contract. |
| [`SetSitting`](../../server/internal/world/players.go#L422)<br>`func (p *ConnectedPlayer) SetSitting(v bool)` | See the source definition and the subsystem ownership contract. |
| [`IsSitting`](../../server/internal/world/players.go#L428)<br>`func (p *ConnectedPlayer) IsSitting() bool` | See the source definition and the subsystem ownership contract. |
| [`GMVisible`](../../server/internal/world/players.go#L434)<br>`func (p *ConnectedPlayer) GMVisible() bool` | See the source definition and the subsystem ownership contract. |
| [`IsAlive`](../../server/internal/world/players.go#L440)<br>`func (p *ConnectedPlayer) IsAlive() bool` | See the source definition and the subsystem ownership contract. |
| [`SetAlive`](../../server/internal/world/players.go#L444)<br>`func (p *ConnectedPlayer) SetAlive(alive bool)` | SetAlive is called by the owning session. A death has a distinct state even if the same character revives and dies again during a prepared cast. |
| [`QueueTeleport`](../../server/internal/world/players.go#L453)<br>`func (p *ConnectedPlayer) QueueTeleport(mapID uint16, x, y, z float32)` | See the source definition and the subsystem ownership contract. |
| [`TakeTeleport`](../../server/internal/world/players.go#L459)<br>`func (p *ConnectedPlayer) TakeTeleport() *pendingTeleport` | See the source definition and the subsystem ownership contract. |
| [`tradeRefresh`](../../server/internal/world/players.go#L470)<br>`type tradeRefresh` | tradeRefresh tells the other side of a completed trade to reload its items and apply the money change itself (only the owner mutates its character). |
| [`QueueTradeRefresh`](../../server/internal/world/players.go#L475)<br>`func (p *ConnectedPlayer) QueueTradeRefresh(packets [][]byte, moneyDelta int64)` | See the source definition and the subsystem ownership contract. |
| [`TakeTradeRefresh`](../../server/internal/world/players.go#L481)<br>`func (p *ConnectedPlayer) TakeTradeRefresh() *tradeRefresh` | See the source definition and the subsystem ownership contract. |
| [`QueueDamage`](../../server/internal/world/players.go#L489)<br>`func (p *ConnectedPlayer) QueueDamage(d pendingDamage)` | See the source definition and the subsystem ownership contract. |
| [`TakeDamage`](../../server/internal/world/players.go#L500)<br>`func (p *ConnectedPlayer) TakeDamage() []pendingDamage` | See the source definition and the subsystem ownership contract. |
| [`SitMotionPacket`](../../server/internal/world/players.go#L508)<br>`func (p *ConnectedPlayer) SitMotionPacket() []byte` | See the source definition and the subsystem ownership contract. |
| [`EnteredMapPacket`](../../server/internal/world/players.go#L513)<br>`func (p *ConnectedPlayer) EnteredMapPacket() []byte` | EnteredMapPacket introduces an already-present character to a viewer. |
| [`ShapePacket`](../../server/internal/world/players.go#L521)<br>`func (p *ConnectedPlayer) ShapePacket() []byte` | ShapePacket never embeds sitting; sit is sent as a separate 0x0506. |
| [`MovePacket`](../../server/internal/world/players.go#L528)<br>`func (p *ConnectedPlayer) MovePacket() []byte` | See the source definition and the subsystem ownership contract. |
| [`EquipmentPacket`](../../server/internal/world/players.go#L534)<br>`func (p *ConnectedPlayer) EquipmentPacket(slot uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`Send`](../../server/internal/world/players.go#L542)<br>`func (p *ConnectedPlayer) Send(frame []byte, label string) error` | Send writes a frame to this player with the post-handshake expanded XOR encoding, which is stateless and safe from any goroutine. |
| [`appendWorldFrame`](../../server/internal/world/players.go#L550)<br>`func appendWorldFrame(dst, frame []byte) []byte` | See the source definition and the subsystem ownership contract. |
| [`playersWhere`](../../server/internal/world/players.go#L558)<br>`func playersWhere(mapID uint16, keep func(*ConnectedPlayer) bool) []*ConnectedPlayer` | See the source definition and the subsystem ownership contract. |
| [`PlayersOnMap`](../../server/internal/world/players.go#L574)<br>`func PlayersOnMap(mapID uint16) []*ConnectedPlayer` | See the source definition and the subsystem ownership contract. |
| [`PlayersNear`](../../server/internal/world/players.go#L578)<br>`func PlayersNear(mapID uint16, x, z, radius float32) []*ConnectedPlayer` | See the source definition and the subsystem ownership contract. |
| [`PlayersNearExcluding`](../../server/internal/world/players.go#L585)<br>`func PlayersNearExcluding(mapID uint16, characterID uint32, x, z, radius float32) []*ConnectedPlayer` | See the source definition and the subsystem ownership contract. |
| [`PlayerCount`](../../server/internal/world/players.go#L596)<br>`func PlayerCount() int` | PlayerCount reports characters registered in the world. |
| [`AllPlayers`](../../server/internal/world/players.go#L606)<br>`func AllPlayers() []*ConnectedPlayer` | See the source definition and the subsystem ownership contract. |
| [`FindPlayerByName`](../../server/internal/world/players.go#L618)<br>`func FindPlayerByName(name string) *ConnectedPlayer` | See the source definition and the subsystem ownership contract. |
| [`FindPlayerByID`](../../server/internal/world/players.go#L627)<br>`func FindPlayerByID(id uint32) *ConnectedPlayer` | See the source definition and the subsystem ownership contract. |
| [`canSeeCharacter`](../../server/internal/world/players.go#L637)<br>`func canSeeCharacter(viewerIsAdmin, subjectIsAdmin, subjectGMVisible bool) bool` | Hidden GMs are visible only to other GMs. |
| [`broadcast`](../../server/internal/world/players.go#L642)<br>`func broadcast(players []*ConnectedPlayer, frame []byte, label string, log *logx.Logger)` | broadcast sends frame to each player, logging failures. |
| [`UpdateElements`](../../server/internal/world/players.go#L652)<br>`func (p *ConnectedPlayer) UpdateElements(attack, defense uint8)` | UpdateElements publishes the owning session's native event state for detection; no other session reads or mutates its applied skill list. |
| [`QueueTradeCommit`](../../server/internal/world/players.go#L659)<br>`func (p *ConnectedPlayer) QueueTradeCommit(trade *tradeSession)` | See the source definition and the subsystem ownership contract. |
| [`takeTradeCommit`](../../server/internal/world/players.go#L664)<br>`func (p *ConnectedPlayer) takeTradeCommit() *tradeSession` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type connWriter struct {
	mu     sync.Mutex
	conn   net.Conn
	queued *netx.Writer
}
```

```go
type pendingTeleport struct {
	MapID   uint16
	X, Y, Z float32
}
```

```go
type pendingDamage struct {
	Guard       *npcGuardStrike
	SkillID     uint16
	MobGlobalID uint32
	MobID       uint16
	Attack      uint32
	Slot        uint8
	First       bool // the skill's selected target
	StartsCast  bool // one area recipient carries the cast animation
}
```

```go
type ConnectedPlayer struct {
	owner       *Session        // immutable; stateMu guards paired trade access
	tradeCommit *tradeSession   // guarded by mailMu
	ride        vehiclePresence // guarded by vehicleMu

	targetEffects []ActiveBuff
	assistTarget  uint32 // guarded by mu
	CharacterID   uint32
	CharacterName string
	IsAdmin       bool
	out           *connWriter
	log           *logx.Logger

	mu                            sync.Mutex
	mapID                         uint16
	character                     db.Character
	items                         []db.Item
	speedKnown                    bool
	speeds                        playerSpeeds
	elementKnown                  bool
	attackElement, defenseElement uint8
	faction                       uint8
	ladderIndex                   uint16
	ladderMode                    uint8
	motion                        uint8
	direction                     uint16
	sitting                       bool
	gmVisible                     bool
	mount                         uint8 // 0x051D vehicle shape, 0 on foot

	mailMu           sync.Mutex
	teleport         *pendingTeleport
	summonOffer      *partySummonOffer
	forced           *forcedTeleport // a GM's /asummon waiting for this user's session
	tradeRefresh     *tradeRefresh
	damage           []pendingDamage
	lifeSkill        *pendingLifeSkill
	soulKill         *pendingSoulKill
	friendly         []pendingFriendly
	userHits         []pendingUserHit
	combatMailClosed bool
	pvpKillCount     uint32
	duelEnds         []pendingDuelEnd
	duelWins         int
	itemGrants       []db.Item

	guild  playerGuildState // guarded by guildMu
	reveal bool             // an enemy's Detection found this user: stealth ends

	hidden atomic.Bool

	formLook    atomic.Pointer[formLook] // immutable form appearance for users entering the view
	immune      atomic.Bool
	noMobTarget atomic.Bool
	lifeState   atomic.Uint64 // incremented on alive/dead transitions; odd while alive
	debuffed    atomic.Bool   // an active debuff (PvPHaveDebuff), refreshed by its session
	kcRefresh   atomic.Bool   // the country tier changed (ApplyKillCount)

	duel duelState // guarded by duelMu

	// Party state, guarded by partyMu.
	party          *party
	inviterID      uint32
	inviteDeadline time.Time
	inviteAnyMap   bool // the pending party invitation was global (/party name)

	shopMu     sync.Mutex
	shopTitle  string // the market bubble over this character
	viewedShop *ConnectedPlayer

	rewardMu   sync.Mutex
	experience experienceBatch
	perfectExp []perfectExperience
	gold       uint32
	questKills map[uint16]uint8
	lootQueue  []lootGrant // group loot waiting for this session

	questMu  sync.Mutex
	questIDs map[uint16]bool // active quests, for loot eligibility

	view playerView // cell and the entities this client knows
}
```

```go
type perfectExperience struct {
	before experienceBatch
	amount uint32
}
```

```go
type rewardSnapshot struct {
	experience experienceBatch
	perfect    []perfectExperience
	gold       uint32
}
```

```go
type experienceBatch struct {
	hundreds   uint64
	remainders [100]uint64
}
```

```go
type tradeRefresh struct {
	Packets    [][]byte
	MoneyDelta int64
}
```

</details>

### portal_bounds.go

[server/internal/world/portal_bounds.go](../../server/internal/world/portal_bounds.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`portalVolume`](../../server/internal/world/portal_bounds.go#L10)<br>`type portalVolume` | See the source definition and the subsystem ownership contract. |
| [`portalVolumeCatalog`](../../server/internal/world/portal_bounds.go#L17)<br>`type portalVolumeCatalog` | See the source definition and the subsystem ownership contract. |
| [`portalVolumes`](../../server/internal/world/portal_bounds.go#L19)<br>`var portalVolumes portalVolumeCatalog` | See the source definition and the subsystem ownership contract. |
| [`readPortalVolumes`](../../server/internal/world/portal_bounds.go#L21)<br>`func readPortalVolumes(path string) (portalVolumeCatalog, error)` | See the source definition and the subsystem ownership contract. |
| [`loadPortalVolumes`](../../server/internal/world/portal_bounds.go#L45)<br>`func loadPortalVolumes() error` | See the source definition and the subsystem ownership contract. |
| [`apply`](../../server/internal/world/portal_bounds.go#L54)<br>`func (catalog portalVolumeCatalog) apply(mapID uint16, portals []MapPortal)` | See the source definition and the subsystem ownership contract. |
| [`contains`](../../server/internal/world/portal_bounds.go#L68)<br>`func (v *portalVolume) contains(x, y, z, tolerance float32) bool` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type portalVolume struct {
	Target   uint16     `json:"target"`
	Min      [3]float32 `json:"min"`
	Max      [3]float32 `json:"max"`
	ArrivalY *float32   `json:"arrival_y,omitempty"`
}
```

```go
type portalVolumeCatalog map[uint16]map[uint8]portalVolume
```

</details>

### portal_bounds_native.go

[server/internal/world/portal_bounds_native.go](../../server/internal/world/portal_bounds_native.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`cloronC3ArrivalY`](../../server/internal/world/portal_bounds_native.go#L5)<br>`var cloronC3ArrivalY float32` | EP5 WLD portal boxes matched to the distributed SVMAP routes. The C2 gate to C3 lands on the upper entrance corridor. |
| [`nativePortalVolumes`](../../server/internal/world/portal_bounds_native.go#L7)<br>`var nativePortalVolumes` | See the source definition and the subsystem ownership contract. |

### potential.go

[server/internal/world/potential.go](../../server/internal/world/potential.go#L1)

Direct GX dependencies: `internal/potential`, `internal/skilltiming`, `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`isPotential`](../../server/internal/world/potential.go#L11)<br>`func isPotential(def *db.SkillDef) bool` | See the source definition and the subsystem ownership contract. |
| [`tickPotentials`](../../server/internal/world/potential.go#L22)<br>`func (s *Session) tickPotentials(now time.Time)` | tickPotentials runs the Potential skills (TypeEffect 0, TypeDetail 6): a learned Potential whose LimitHP percent is above the character's current HP percent applies itself as an ordinary buff for its KeepTime and starts its reuse timer. It cannot be cast, so the reuse only ever starts here. Its abilities are the ones in the definition (Potential Motive's ability 19 is a cooldown reduction, see cooldown_reduction.go). The effect is applied before the Potential's own reuse starts, and a reduction never shortens it. |
| [`hasBuff`](../../server/internal/world/potential.go#L55)<br>`func (s *Session) hasBuff(skillID uint16) bool` | hasBuff reports whether this skill is already applied. |

### protocol.go

[server/internal/world/protocol.go](../../server/internal/world/protocol.go#L1)

Direct GX dependencies: `internal/textwire`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`labeled`](../../server/internal/world/protocol.go#L11)<br>`type labeled` | labeled pairs a frame with its log label. |
| [`buildWorldHandshake`](../../server/internal/world/protocol.go#L16)<br>`func buildWorldHandshake() []byte` | See the source definition and the subsystem ownership contract. |
| [`buildAccountFaction`](../../server/internal/world/protocol.go#L24)<br>`func buildAccountFaction(faction uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`isValidFaction`](../../server/internal/world/protocol.go#L31)<br>`func isValidFaction(f uint8) bool` | See the source definition and the subsystem ownership contract. |
| [`resultByte`](../../server/internal/world/protocol.go#L33)<br>`func resultByte(ok bool) uint8` | See the source definition and the subsystem ownership contract. |
| [`buildCreateCharacterResult`](../../server/internal/world/protocol.go#L41)<br>`func buildCreateCharacterResult(ok bool) []byte` | buildCreateCharacterResult: one byte, 0 ok, 1 failure ("name in use"). |
| [`buildCharResult`](../../server/internal/world/protocol.go#L47)<br>`func buildCharResult(op packet.Opcode, ok bool, characterID uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildRenameResult`](../../server/internal/world/protocol.go#L55)<br>`func buildRenameResult(ok bool) []byte` | buildRenameResult uses the inverted convention: 1 ok, 0 failure. |
| [`buildWorldDay`](../../server/internal/world/protocol.go#L61)<br>`func buildWorldDay() []byte` | See the source definition and the subsystem ownership contract. |
| [`buildKillCountUpdate`](../../server/internal/world/protocol.go#L67)<br>`func buildKillCountUpdate(index uint8, value uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildZeroCount`](../../server/internal/world/protocol.go#L74)<br>`func buildZeroCount(op packet.Opcode) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildBlessInit`](../../server/internal/world/protocol.go#L80)<br>`func buildBlessInit(faction uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildAccountPoints`](../../server/internal/world/protocol.go#L88)<br>`func buildAccountPoints(points uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildCashPoints`](../../server/internal/world/protocol.go#L95)<br>`func buildCashPoints(points uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildPing`](../../server/internal/world/protocol.go#L101)<br>`func buildPing() []byte` | See the source definition and the subsystem ownership contract. |
| [`decodeText`](../../server/internal/world/protocol.go#L108)<br>`func decodeText(b []byte) string` | GX chat uses UTF-8; legacy original ANSI packets remain readable. |
| [`encodeText`](../../server/internal/world/protocol.go#L109)<br>`func encodeText(s string) []byte` | See the source definition and the subsystem ownership contract. |
| [`maxChatMessageLen`](../../server/internal/world/protocol.go#L111)<br>`const maxChatMessageLen = 127` | See the source definition and the subsystem ownership contract. |
| [`errChat`](../../server/internal/world/protocol.go#L113)<br>`var errChat` | See the source definition and the subsystem ownership contract. |
| [`readChatMessage`](../../server/internal/world/protocol.go#L116)<br>`func readChatMessage(payload []byte) (string, error)` | readChatMessage parses length:u8 (including the terminator) + text. |
| [`writeChatText`](../../server/internal/world/protocol.go#L134)<br>`func writeChatText(w *packet.Writer, message string)` | See the source definition and the subsystem ownership contract. |
| [`buildChatByID`](../../server/internal/world/protocol.go#L144)<br>`func buildChatByID(op packet.Opcode, senderID uint32, message string) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildChatNormal`](../../server/internal/world/protocol.go#L151)<br>`func buildChatNormal(characterID uint32, message string) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildChatByName`](../../server/internal/world/protocol.go#L155)<br>`func buildChatByName(op packet.Opcode, sender, message string) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildChatError`](../../server/internal/world/protocol.go#L166)<br>`func buildChatError(result uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`isChatOpcode`](../../server/internal/world/protocol.go#L172)<br>`func isChatOpcode(op packet.Opcode) bool` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type labeled struct {
	Packet []byte
	Label  string
}
```

</details>

### pvp.go

[server/internal/world/pvp.go](../../server/internal/world/pvp.go#L1)

Direct GX dependencies: `internal/attackspeed`, `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`deathTypeUser`](../../server/internal/world/pvp.go#L28)<br>`const deathTypeUser = 0x01` | See the source definition and the subsystem ownership contract. |
| [`userAttackRange`](../../server/internal/world/pvp.go#L29)<br>`const userAttackRange = 2.3` | See the source definition and the subsystem ownership contract. |
| [`pendingUserHit`](../../server/internal/world/pvp.go#L34)<br>`type pendingUserHit` | pendingUserHit is a user's normal attack (or a mirrored one coming back) to resolve in the target's session. |
| [`QueueUserHit`](../../server/internal/world/pvp.go#L51)<br>`func (p *ConnectedPlayer) QueueUserHit(h pendingUserHit) bool` | See the source definition and the subsystem ownership contract. |
| [`TakeUserHits`](../../server/internal/world/pvp.go#L67)<br>`func (p *ConnectedPlayer) TakeUserHits() []pendingUserHit` | See the source definition and the subsystem ownership contract. |
| [`closeUserHits`](../../server/internal/world/pvp.go#L75)<br>`func (p *ConnectedPlayer) closeUserHits()` | See the source definition and the subsystem ownership contract. |
| [`isTeam`](../../server/internal/world/pvp.go#L92)<br>`func (s *Session) isTeam(p *ConnectedPlayer) bool` | isTeam is CUser::IsTeam without duels (none exist yet): the user itself, anyone on a WarType P map, or the same country. isTeam is CUser::IsTeam: the user itself, else not the duel opponent, then a WarType P map or the same country. |
| [`hostileUser`](../../server/internal/world/pvp.go#L106)<br>`func (s *Session) hostileUser(id uint32) *ConnectedPlayer` | hostileUser finds an attackable user for a normal attack. |
| [`userAttackSwing`](../../server/internal/world/pvp.go#L119)<br>`func (s *Session) userAttackSwing(target uint32) bool` | userAttackSwing is one UpdateAttack step on a user target; false ends the auto-attack. |
| [`tickUserHits`](../../server/internal/world/pvp.go#L169)<br>`func (s *Session) tickUserHits()` | tickUserHits resolves the user attacks queued on this session. |
| [`applyUserHit`](../../server/internal/world/pvp.go#L191)<br>`func (s *Session) applyUserHit(h pendingUserHit) bool` | applyUserHit is the target side of CUser::AttackUser; true when this character died. |
| [`takeUserDamage`](../../server/internal/world/pvp.go#L272)<br>`func (s *Session) takeUserDamage(attackerID, damage uint32) bool` | takeUserDamage is CUser::Damage's HP part; true when it killed. |
| [`afterUserDamage`](../../server/internal/world/pvp.go#L299)<br>`func (s *Session) afterUserDamage(killerID uint32, killed bool)` | See the source definition and the subsystem ownership contract. |
| [`buildCharacterUserAttack`](../../server/internal/world/pvp.go#L320)<br>`func buildCharacterUserAttack(attackerID, targetID uint32, result uint8, damage uint16) []byte` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type pendingUserHit struct {
	AttackerID uint32
	Stats      CharacterStats // the attacker's, at the swing
	Kind       int
	Reflected  bool // a mirrored hit: Damage is final
	Execute    bool // GM /kill: the target dies at once; the kill is credited to AttackerID
	Fixed      bool // fixed skill damage (Fire Thorn): Damage is final, Skill names it
	Trap       bool // a trap sprang on this user: the damage is worked out here (TrabDamage), Skill names it
	Damage     [3]uint32
	Skill      *db.SkillDef  // a hostile skill (pvpskill.go), nil for a normal attack
	First      bool          // the skill's selected target (0x0511/0x0512), else 0x0513
	Caster     damagePools   // the caster's HP/SP/MP (DamageType 3)
	Drain      *pendingDrain // a landed drain to finish in the caster's session
	ItemCast   bool          // cast by a gem (m_bItemAttack): casts no gems itself
	GemAttack  bool          // run AttackItemOption on AttackerID (the target hit)
}
```

</details>

### pvpdeath.go

[server/internal/world/pvpdeath.go](../../server/internal/world/pvpdeath.go#L1)

Direct GX dependencies: `internal/proto`, `internal/pvprank`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`damageByUser`](../../server/internal/world/pvpdeath.go#L37)<br>`const damageByUser = 1` | See the source definition and the subsystem ownership contract. |
| [`damageByMob`](../../server/internal/world/pvpdeath.go#L38)<br>`const damageByMob = 2` | See the source definition and the subsystem ownership contract. |
| [`pvpKillCap`](../../server/internal/world/pvpdeath.go#L39)<br>`const pvpKillCap = pvprank.CounterCap` | See the source definition and the subsystem ownership contract. |
| [`pvpDeathCap`](../../server/internal/world/pvpdeath.go#L40)<br>`const pvpDeathCap = pvprank.CounterCap` | See the source definition and the subsystem ownership contract. |
| [`pvpExpFactor`](../../server/internal/world/pvpdeath.go#L41)<br>`const pvpExpFactor = 1.2` | See the source definition and the subsystem ownership contract. |
| [`damageEntry`](../../server/internal/world/pvpdeath.go#L45)<br>`type damageEntry` | damageEntry is one CDamage node. |
| [`addDamage`](../../server/internal/world/pvpdeath.go#L52)<br>`func (s *Session) addDamage(kind uint8, id, amount uint32)` | addDamage is CDamage::Add. |
| [`topDamage`](../../server/internal/world/pvpdeath.go#L63)<br>`func (s *Session) topDamage() (damageEntry, bool)` | topDamage is CDamage::GetMax: the first entry with the highest damage. |
| [`pvpExp`](../../server/internal/world/pvpdeath.go#L78)<br>`func pvpExp(deadLevel, killerLevel uint16) uint32` | pvpExp is CUser::GetExp: the dead user's level x 50 x ((dead - killer + 1) x 0.2 + 1), at least 1, and 0 when the dead user is 10+ levels above. |
| [`pvpKill`](../../server/internal/world/pvpdeath.go#L88)<br>`type pvpKill` | pvpKill is a kill credited to a user's session (KillPCTo). |
| [`QueuePvPKill`](../../server/internal/world/pvpdeath.go#L90)<br>`func (p *ConnectedPlayer) QueuePvPKill(k pvpKill)` | See the source definition and the subsystem ownership contract. |
| [`takePvPKillCount`](../../server/internal/world/pvpdeath.go#L95)<br>`func (p *ConnectedPlayer) takePvPKillCount() uint32` | See the source definition and the subsystem ownership contract. |
| [`QueueItemGrant`](../../server/internal/world/pvpdeath.go#L104)<br>`func (p *ConnectedPlayer) QueueItemGrant(it db.Item) bool` | QueueItemGrant hands an item straight to a user (ItemGet). |
| [`TakeItemGrants`](../../server/internal/world/pvpdeath.go#L115)<br>`func (p *ConnectedPlayer) TakeItemGrants() []db.Item` | See the source definition and the subsystem ownership contract. |
| [`tickPvP`](../../server/internal/world/pvpdeath.go#L124)<br>`func (s *Session) tickPvP()` | tickPvP applies the kills and items other sessions credited to this one. |
| [`killPCTo`](../../server/internal/world/pvpdeath.go#L149)<br>`func (s *Session) killPCTo()` | killPCTo is CUser::KillPCTo on this (the killing) user. |
| [`killPCToCount`](../../server/internal/world/pvpdeath.go#L150)<br>`func (s *Session) killPCToCount(count uint32)` | See the source definition and the subsystem ownership contract. |
| [`questKillPC`](../../server/internal/world/pvpdeath.go#L177)<br>`func (s *Session) questKillPC()` | questKillPC is CUser::QuestUpdateKillPC: the first active quest that ends on player kills counts one. |
| [`deathByUser`](../../server/internal/world/pvpdeath.go#L194)<br>`func (s *Session) deathByUser()` | deathByUser runs CUser::Death's reward part on this (the dead) user. |
| [`notifyKill`](../../server/internal/world/pvpdeath.go#L241)<br>`func (s *Session) notifyKill(killer *ConnectedPlayer)` | notifyKill tells a user who gets the credit for this user's death which player it killed; the client shows it in red (SysMsg 8). |
| [`killPCFrom`](../../server/internal/world/pvpdeath.go#L249)<br>`func (s *Session) killPCFrom(killer *ConnectedPlayer)` | killPCFrom is CUser::KillPCFrom: the killer's KillPCTo, then this user's death count. |
| [`placeGroundItem`](../../server/internal/world/pvpdeath.go#L261)<br>`func (s *Session) placeGroundItem(it db.Item, owner uint32, noParty bool)` | placeGroundItem drops an item at this user's feet (drop type 3). |

<details><summary>Type definitions and owned fields</summary>

```go
type damageEntry struct {
	Type   uint8
	ID     uint32
	Amount uint32
}
```

```go
type pvpKill struct{ VictimID uint32 }
```

</details>

### pvpreward.go

[server/internal/world/pvpreward.go](../../server/internal/world/pvpreward.go#L1)

Direct GX dependencies: `internal/proto`, `internal/pvprank`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`handlePvPReward`](../../server/internal/world/pvpreward.go#L12)<br>`func (s *Session) handlePvPReward(op packet.Opcode, p []byte)` | CUser::PacketMaininterface 0x0218-0x021A requires non-easy mode and level16+. StatusKillResult/StatusDeathResult redeem exactly one unclaimed stage. |
| [`buildPvPRewardStatus`](../../server/internal/world/pvpreward.go#L46)<br>`func buildPvPRewardStatus(ch *db.Character) []byte` | See the source definition and the subsystem ownership contract. |
| [`claimPvPReward`](../../server/internal/world/pvpreward.go#L53)<br>`func claimPvPReward(ch db.Character, death bool) (db.Character, uint32, bool)` | See the source definition and the subsystem ownership contract. |

### pvpskill.go

[server/internal/world/pvpskill.go](../../server/internal/world/pvpskill.go#L1)

Direct GX dependencies: `internal/skillrange`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`opUserSkillDrain`](../../server/internal/world/pvpskill.go#L37)<br>`const opUserSkillDrain = 0x0512` | See the source definition and the subsystem ownership contract. |
| [`opUserSkillOther`](../../server/internal/world/pvpskill.go#L38)<br>`const opUserSkillOther = 0x0513` | See the source definition and the subsystem ownership contract. |
| [`opUserSkillExtra`](../../server/internal/world/pvpskill.go#L39)<br>`const opUserSkillExtra = 0x0514` | See the source definition and the subsystem ownership contract. |
| [`opUserSkillPvP`](../../server/internal/world/pvpskill.go#L40)<br>`const opUserSkillPvP = packet.SkillUseNoTarget` | 0x0511 |
| [`userSkillCenter`](../../server/internal/world/pvpskill.go#L45)<br>`func (s *Session) userSkillCenter(def *db.SkillDef, target uint32) (*ConnectedPlayer, bool)` | userSkillCenter resolves a 0x0511 hostile skill's target user; ok false is ChkTargetUser's refusal. |
| [`hostileUsersWithin`](../../server/internal/world/pvpskill.go#L54)<br>`func (s *Session) hostileUsersWithin(x, z, r float32) []*ConnectedPlayer` | hostileUsersWithin lists the alive hostile users within r of (x, z). |
| [`userSkillAttack`](../../server/internal/world/pvpskill.go#L69)<br>`func (s *Session) userSkillAttack(k db.CharacterSkill, def *db.SkillDef, target uint32, castPos *[2]float32, stats CharacterStats)` | userSkillAttack runs an attack or debuff skill sent as 0x0511. |
| [`areaUserHits`](../../server/internal/world/pvpskill.go#L134)<br>`func (s *Session) areaUserHits(def *db.SkillDef, k db.CharacterSkill, x, y, z float32)` | areaUserHits queues visible hostile users inside the 3D area radius. |
| [`queueSkillHits`](../../server/internal/world/pvpskill.go#L150)<br>`func (s *Session) queueSkillHits(p *ConnectedPlayer, def *db.SkillDef, first bool, n uint8)` | See the source definition and the subsystem ownership contract. |
| [`applyUserSkillHit`](../../server/internal/world/pvpskill.go#L162)<br>`func (s *Session) applyUserSkillHit(h pendingUserHit, stats CharacterStats) bool` | applyUserSkillHit is the target side of CUser::AttackUser(skill); true when this character died. |
| [`takeUserSkillDamage`](../../server/internal/world/pvpskill.go#L271)<br>`func (s *Session) takeUserSkillDamage(attackerID uint32, d [3]uint32) bool` | takeUserSkillDamage is CUser::Damage with SP and MP damage too. |
| [`pendingDrain`](../../server/internal/world/pvpskill.go#L292)<br>`type pendingDrain` | pendingDrain carries a landed PvP hit to the caster, whose session owns the pools and can put the actual recovery into the native 0x0512 result. |
| [`broadcast`](../../server/internal/world/pvpskill.go#L301)<br>`func (d *pendingDrain) broadcast(healed [3]uint16)` | See the source definition and the subsystem ownership contract. |
| [`applyDrainHeal`](../../server/internal/world/pvpskill.go#L307)<br>`func (s *Session) applyDrainHeal(d *pendingDrain)` | applyDrainHeal is CUser::Heal on the drain's caster. |
| [`buildUserSkillExtra`](../../server/internal/world/pvpskill.go#L325)<br>`func buildUserSkillExtra(caster, target uint32, skillID uint16, level uint8, add [3]uint32) []byte` | buildUserSkillExtra (0x0514): caster, target, skill, level, the AddDamage HP/SP/MP. |

<details><summary>Type definitions and owned fields</summary>

```go
type pendingDrain struct {
	source, target  uint32
	mapID           uint16
	x, z            float32
	skill           *db.SkillDef
	result          uint8
	dealt, recovery [3]uint16
}
```

</details>

### quest_map.go

[server/internal/world/quest_map.go](../../server/internal/world/quest_map.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`questNPCCatalogCache`](../../server/internal/world/quest_map.go#L10)<br>`type questNPCCatalogCache` | See the source definition and the subsystem ownership contract. |
| [`questNPCCatalogState`](../../server/internal/world/quest_map.go#L15)<br>`type questNPCCatalogState` | See the source definition and the subsystem ownership contract. |
| [`questNPCSnapshot`](../../server/internal/world/quest_map.go#L24)<br>`func (sh *mapShard) questNPCSnapshot() (uint64, []proto.QuestNPC)` | Shared immutable map metadata. Patrols use their authored first waypoint; ordinary movement neither invalidates this cache nor streams global positions. |
| [`questNPCFrames`](../../server/internal/world/quest_map.go#L50)<br>`func questNPCFrames(mapID uint16, rows []proto.QuestNPC) [][]byte` | See the source definition and the subsystem ownership contract. |
| [`syncQuestNPCCatalog`](../../server/internal/world/quest_map.go#L81)<br>`func (s *Session) syncQuestNPCCatalog(force bool)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type questNPCCatalogCache struct {
	ready    bool
	revision uint64
	rows     []proto.QuestNPC
}
```

```go
type questNPCCatalogState struct {
	sent     bool
	mapID    uint16
	revision uint64
	rows     []proto.QuestNPC
}
```

</details>

### quest_requirements.go

[server/internal/world/quest_requirements.go](../../server/internal/world/quest_requirements.go#L1)

Direct GX dependencies: `internal/npcquest`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`questGroupReady`](../../server/internal/world/quest_requirements.go#L10)<br>`func (s *Session) questGroupReady(required [6]uint8) bool` | See the source definition and the subsystem ownership contract. |
| [`questStartRestriction`](../../server/internal/world/quest_requirements.go#L32)<br>`func (s *Session) questStartRestriction(q *questDef) string` | Additional configurable requirements are optional. HG and VG at the quest level still resolve alternative prerequisite quest IDs as in EP5. |
| [`questFeedback`](../../server/internal/world/quest_requirements.go#L43)<br>`func (s *Session) questFeedback(message string)` | See the source definition and the subsystem ownership contract. |
| [`questItemNeedsAtomicStart`](../../server/internal/world/quest_requirements.go#L48)<br>`func questItemNeedsAtomicStart(typ, id uint8) bool` | Guarded quest items are charged only when the selected quest is accepted. |
| [`hasRequirements`](../../server/internal/world/quest_requirements.go#L59)<br>`func (r questResult) hasRequirements() bool` | See the source definition and the subsystem ownership contract. |
| [`expired`](../../server/internal/world/quest_requirements.go#L64)<br>`func (q *activeQuest) expired(now time.Time) bool` | See the source definition and the subsystem ownership contract. |
| [`questObjectives`](../../server/internal/world/quest_requirements.go#L68)<br>`func (s *Session) questObjectives(q *activeQuest) (totals [3]int, ready bool)` | See the source definition and the subsystem ownership contract. |
| [`questRewardEligible`](../../server/internal/world/quest_requirements.go#L89)<br>`func (s *Session) questRewardEligible(q *activeQuest, r questResult, totals [3]int, now time.Time) bool` | See the source definition and the subsystem ownership contract. |
| [`retryAutomatically`](../../server/internal/world/quest_requirements.go#L107)<br>`func (q *activeQuest) retryAutomatically() bool` | See the source definition and the subsystem ownership contract. |
| [`usesElapsedSeconds`](../../server/internal/world/quest_requirements.go#L127)<br>`func (d *questDef) usesElapsedSeconds() bool` | See the source definition and the subsystem ownership contract. |

### quests.go

[server/internal/world/quests.go](../../server/internal/world/quests.go#L1)

Direct GX dependencies: `internal/npcquest`, `server/internal/config`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`questItem`](../../server/internal/world/quests.go#L23)<br>`type questItem` | See the source definition and the subsystem ownership contract. |
| [`questResult`](../../server/internal/world/quests.go#L25)<br>`type questResult` | See the source definition and the subsystem ownership contract. |
| [`questDef`](../../server/internal/world/quests.go#L38)<br>`type questDef` | See the source definition and the subsystem ownership contract. |
| [`questItemLink`](../../server/internal/world/quests.go#L73)<br>`type questItemLink` | See the source definition and the subsystem ownership contract. |
| [`questDefs`](../../server/internal/world/quests.go#L76)<br>`var questDefs []questDef` | See the source definition and the subsystem ownership contract. |
| [`questItemLinks`](../../server/internal/world/quests.go#L77)<br>`var questItemLinks` | See the source definition and the subsystem ownership contract. |
| [`mobQuestLinks`](../../server/internal/world/quests.go#L78)<br>`var mobQuestLinks` | See the source definition and the subsystem ownership contract. |
| [`sqliteQuest`](../../server/internal/world/quests.go#L82)<br>`func sqliteQuest(row npcquest.Row) questDef` | sqliteQuest follows the native field order declared by QuestDetails. |
| [`findQuestDef`](../../server/internal/world/quests.go#L134)<br>`func findQuestDef(id uint16) *questDef` | findQuestDef is GetQuest: by row (1-based), not the embedded id. |
| [`buildMobQuestLinks`](../../server/internal/world/quests.go#L143)<br>`func buildMobQuestLinks()` | buildMobQuestLinks maps mobs to the mob-kill quests that count them, ascending by minimum level (MobQuestLinks::build). |
| [`activeQuest`](../../server/internal/world/quests.go#L175)<br>`type activeQuest` | activeQuest is a quest in progress. |
| [`delayMinutes`](../../server/internal/world/quests.go#L186)<br>`func (q *activeQuest) delayMinutes() int16` | See the source definition and the subsystem ownership contract. |
| [`row`](../../server/internal/world/quests.go#L193)<br>`func (q *activeQuest) row() db.CharQuest` | See the source definition and the subsystem ownership contract. |
| [`loadQuests`](../../server/internal/world/quests.go#L202)<br>`func (s *Session) loadQuests() error` | loadQuests restores the character's quests as PacketUserDBSetInfo does. |
| [`questListPackets`](../../server/internal/world/quests.go#L255)<br>`func (s *Session) questListPackets() [][]byte` | questListPackets are 0x0901 (id, minutes left, counters) and 0x0906 (id, success), up to 255 entries each. |
| [`questSnapshot`](../../server/internal/world/quests.go#L286)<br>`func (s *Session) questSnapshot() []db.CharQuest` | questSnapshot captures active timers and counters on exit. |
| [`findQuest`](../../server/internal/world/quests.go#L294)<br>`func (s *Session) findQuest(id uint16) *activeQuest` | See the source definition and the subsystem ownership contract. |
| [`questFinished`](../../server/internal/world/quests.go#L303)<br>`func (s *Session) questFinished(id uint16, onlySuccess bool) bool` | See the source definition and the subsystem ownership contract. |
| [`questPrerequisites`](../../server/internal/world/quests.go#L313)<br>`func (s *Session) questPrerequisites(q *questDef) bool` | questPrerequisites: PrevQuest, HG or VG finished successfully (any). |
| [`questEnableStart`](../../server/internal/world/quests.go#L328)<br>`func (s *Session) questEnableStart(q *questDef) bool` | questEnableStart is CUser::QuestEnableStart: level range, race (Faction), job and prerequisites. |
| [`raceAllowed`](../../server/internal/world/quests.go#L343)<br>`func raceAllowed(country, race uint8) bool` | raceAllowed is the Country/Faction race rule shared with items. |
| [`npcMatches`](../../server/internal/world/quests.go#L359)<br>`func (s *Session) npcMatches(npcID uint32, npcType uint8, typeID uint16) (MapNpc, bool)` | See the source definition and the subsystem ownership contract. |
| [`questNPCReach`](../../server/internal/world/quests.go#L365)<br>`const questNPCReach = 30.0` | Use the same authoritative interaction radius as other NPC services. |
| [`questNPCInReach`](../../server/internal/world/quests.go#L367)<br>`func (s *Session) questNPCInReach(npc MapNpc) bool` | See the source definition and the subsystem ownership contract. |
| [`questItemUse`](../../server/internal/world/quests.go#L371)<br>`type questItemUse` | See the source definition and the subsystem ownership contract. |
| [`useQuestItem`](../../server/internal/world/quests.go#L379)<br>`func (s *Session) useQuestItem(index int) bool` | Native ItemUse (0x4655e0) opens the same window for types 27–29. Type 29 consumes one unit before the client can request the quest. |
| [`handleQuestStart`](../../server/internal/world/quests.go#L404)<br>`func (s *Session) handleQuestStart(p []byte)` | handleQuestStart (0x0902): NPC or item context, followed by the quest row. |
| [`startQuest`](../../server/internal/world/quests.go#L447)<br>`func (s *Session) startQuest(q *questDef, id uint16, npcID uint32)` | See the source definition and the subsystem ownership contract. |
| [`questItemTotals`](../../server/internal/world/quests.go#L530)<br>`func questItemTotals(d *questDef, items []db.Item) (totals [3]int)` | questItemTotals counts actual items in carried bags. |
| [`planQuestItems`](../../server/internal/world/quests.go#L545)<br>`func planQuestItems(d *questDef, r questResult, original []db.Item) (items []db.Item, packets [][]byte, ok bool)` | See the source definition and the subsystem ownership contract. |
| [`handleQuestEnd`](../../server/internal/world/quests.go#L590)<br>`func (s *Session) handleQuestEnd(p []byte)` | handleQuestEnd (0x0903): npc id, quest id (quest_end_request). |
| [`completeQuest`](../../server/internal/world/quests.go#L618)<br>`func (s *Session) completeQuest(q *activeQuest, ctx uint32, success, requested bool)` | completeQuest is quest_requests_complete. |
| [`handleQuestChoose`](../../server/internal/world/quests.go#L680)<br>`func (s *Session) handleQuestChoose(p []byte)` | handleQuestChoose (0x0907): quest id, reward index. |
| [`handleQuestAbandon`](../../server/internal/world/quests.go#L709)<br>`func (s *Session) handleQuestAbandon(p []byte)` | handleQuestAbandon (0x0908): quest id. |
| [`awardQuest`](../../server/internal/world/quests.go#L720)<br>`func (s *Session) awardQuest(q *activeQuest, variant uint8, ctx uint32)` | awardQuest sends the credited EXP, authored gold and item through 0x0903. |
| [`removeQuest`](../../server/internal/world/quests.go#L804)<br>`func (s *Session) removeQuest(q *activeQuest, success uint8) bool` | removeQuest is QuestRuntime::remove: repeatable quests (TickReceive) are deleted, others finish with their success flag. |
| [`removeQuestFromMemory`](../../server/internal/world/quests.go#L819)<br>`func (s *Session) removeQuestFromMemory(q *activeQuest, success uint8)` | See the source definition and the subsystem ownership contract. |
| [`bumpQuestCounter`](../../server/internal/world/quests.go#L831)<br>`func (s *Session) bumpQuestCounter(q *activeQuest, slot int)` | See the source definition and the subsystem ownership contract. |
| [`questKill`](../../server/internal/world/quests.go#L849)<br>`func (s *Session) questKill(mobID uint16)` | questKill is update_mob_kill_quests for one kill of mobID. |
| [`questKillCount`](../../server/internal/world/quests.go#L850)<br>`func (s *Session) questKillCount(mobID uint16, count uint8) bool` | See the source definition and the subsystem ownership contract. |
| [`questNeeds`](../../server/internal/world/quests.go#L910)<br>`func (s *Session) questNeeds(t, id uint8) bool` | questNeeds: an item that starts a quest, or one an active quest ends with. |
| [`questPickup`](../../server/internal/world/quests.go#L927)<br>`func (s *Session) questPickup(t, id uint8)` | questPickup is quest_requests_progress_pick_item. |
| [`questTimers`](../../server/internal/world/quests.go#L951)<br>`func (s *Session) questTimers()` | questTimers fails timed quests whose time ran out. |
| [`questItemDrop`](../../server/internal/world/quests.go#L965)<br>`func questItemDrop(mob *MapMob, recipients []*ConnectedPlayer) *db.ItemDef` | questItemDrop is mob_loot's quest item: Mobs.QuestItemID drops (chance = the item's Quality per cent) when it starts a quest or any recipient has a quest it ends. |
| [`syncQuestIDs`](../../server/internal/world/quests.go#L989)<br>`func (s *Session) syncQuestIDs()` | syncQuestIDs publishes the active quest ids for other sessions' loot. |

<details><summary>Type definitions and owned fields</summary>

```go
type questItem struct{ Type, TypeID, Count uint8 }
```

```go
type questResult struct {
	NeedMobID                 uint16 // counter selector for mob quests
	NeedMobCount              uint8
	NeedItemID, NeedItemCount uint8 // counter selector for item quests
	NeedTime                  uint32
	NeedHG                    uint16
	NeedVG                    int16
	NeedOG                    uint8
	Exp, Money                uint32
	ItemType, ItemTypeID      uint8
	NextQuest                 uint16
}
```

```go
type questDef struct {
	ID                  uint16
	Name                string
	MinLevel, MaxLevel  uint16
	Faction             uint8
	Male, Female        bool
	Jobs                [6]uint8
	HG                  uint16 // also an alternative prerequisite quest
	VG                  int16  // likewise
	CG, OG, IG          uint8
	PrevQuest           uint16
	RequireParty        bool
	PartyJobs           [6]uint8
	MinMinutes          uint32
	MaxMinutes          uint32
	TickStart, TickKeep uint32
	TickReceive         uint32 // nonzero: repeatable
	StartType           uint8
	StartNpcType        uint8
	StartNpcID          uint16
	StartItemType       uint8
	StartItemID         uint8
	StartItems          [3]questItem // given on start
	EndType             uint8        // 0 npc, 1 items, 2 mobs, 3 player kills
	EndNpcType          uint8
	EndNpcID            uint16
	FarmItems           [3]questItem
	PvpKills            uint8
	MobIDs              [2]uint16
	MobCounts           [2]uint8
	ResultType          uint8
	UserSelect          uint8
	Results             [3]questResult
}
```

```go
type questItemLink struct{ Start, End []uint16 }
```

```go
type activeQuest struct {
	ID        uint16
	def       *questDef
	counters  [3]uint8
	startAt   time.Time
	endAt     time.Time // timed quests
	successAt time.Time // MinMinutes elapsed
	success   uint8     // 1: waiting for the reward choice
	rewardNPC uint32
}
```

```go
type questItemUse struct {
	item       [2]uint8
	deadline   time.Time
	pendingUID int64
}
```

</details>

### raid.go

[server/internal/world/raid.go](../../server/internal/world/raid.go#L1)

Direct GX dependencies: `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`raidGroupSize`](../../server/internal/world/raid.go#L21)<br>`const raidGroupSize = 6` | slots per group (party) inside a raid |
| [`raidMax`](../../server/internal/world/raid.go#L22)<br>`const raidMax = 150` | five raids of 30 slots (GX extension of the original 30) |
| [`raidKindNormal`](../../server/internal/world/raid.go#L23)<br>`const raidKindNormal = 0` | See the source definition and the subsystem ownership contract. |
| [`raidKindRaid`](../../server/internal/world/raid.go#L24)<br>`const raidKindRaid = 1` | See the source definition and the subsystem ownership contract. |
| [`raidRosterChunk`](../../server/internal/world/raid.go#L25)<br>`const raidRosterChunk = 6` | See the source definition and the subsystem ownership contract. |
| [`raidErrNoRaid`](../../server/internal/world/raid.go#L26)<br>`const raidErrNoRaid = 12` | See the source definition and the subsystem ownership contract. |
| [`raidErrNoAutoJoin`](../../server/internal/world/raid.go#L27)<br>`const raidErrNoAutoJoin = 13` | See the source definition and the subsystem ownership contract. |
| [`raidErrFull`](../../server/internal/world/raid.go#L28)<br>`const raidErrFull = 14` | See the source definition and the subsystem ownership contract. |
| [`raidExpFactor`](../../server/internal/world/raid.go#L32)<br>`const raidExpFactor = 0.01` | GX uses the lowest native raid EXP factor at every raid size. |
| [`isRaid`](../../server/internal/world/raid.go#L34)<br>`func (pt *party) isRaid() bool` | See the source definition and the subsystem ownership contract. |
| [`subLeaderPlayer`](../../server/internal/world/raid.go#L36)<br>`func (pt *party) subLeaderPlayer() *ConnectedPlayer` | See the source definition and the subsystem ownership contract. |
| [`emptyRaidSlot`](../../server/internal/world/raid.go#L44)<br>`func (pt *party) emptyRaidSlot() int` | emptyRaidSlot is the lowest wire slot not in use. |
| [`raidError`](../../server/internal/world/raid.go#L57)<br>`func raidError(to *ConnectedPlayer, reason, id uint32)` | See the source definition and the subsystem ownership contract. |
| [`enterRaid`](../../server/internal/world/raid.go#L65)<br>`func enterRaid(pt *party, p *ConnectedPlayer) bool` | enterRaid is enter_raid_party (partyMu held). |
| [`sendRaidRoster`](../../server/internal/world/raid.go#L94)<br>`func (pt *party) sendRaidRoster(to *ConnectedPlayer)` | sendRaidRoster sends the roster in 0x0B0B chunks of six; the first byte flags whether more chunks follow. |
| [`removeRaid`](../../server/internal/world/raid.go#L115)<br>`func removeRaid(pt *party, p *ConnectedPlayer, expelled bool)` | removeRaid is remove_raid_party (partyMu held): 0x0B0D or 0x0B1A, the sub-leader takes over a departed leader, a raid of one dissolves. |
| [`handleRaidSettings`](../../server/internal/world/raid.go#L172)<br>`func (s *Session) handleRaidSettings(op packet.Opcode, p []byte)` | handleRaidSettings covers 0x0B0E (convert), 0x0B0F (item division), 0x0B10 (auto-join) and 0x0B16 (disband), all leader-only. |
| [`mayInvite`](../../server/internal/world/raid.go#L239)<br>`func (pt *party) mayInvite(p *ConnectedPlayer) bool` | See the source definition and the subsystem ownership contract. |
| [`handleRaidInvite`](../../server/internal/world/raid.go#L245)<br>`func (s *Session) handleRaidInvite(op packet.Opcode, p []byte)` | handleRaidInvite (0x0B14 id / 0x0B18 name): the leader or sub-leader invites; outside a raid it starts a new party as raid-capable. |
| [`handleRaidAnswer`](../../server/internal/world/raid.go#L303)<br>`func (s *Session) handleRaidAnswer(p []byte)` | handleRaidAnswer (0x0B15): 0 accepts; joins the inviter's raid (or creates its party). |
| [`handleRaidAutoJoin`](../../server/internal/world/raid.go#L341)<br>`func (s *Session) handleRaidAutoJoin(p []byte)` | handleRaidAutoJoin (0x0B11): join the raid of the named member when it allows auto-join. |
| [`handleRaidRoles`](../../server/internal/world/raid.go#L369)<br>`func (s *Session) handleRaidRoles(op packet.Opcode, p []byte)` | handleRaidRoles covers 0x0B12 (swap wire slots; leader or sub-leader), 0x0B13 (sub-leader) and 0x0B17 (leader). |
| [`handleRaidExpel`](../../server/internal/world/raid.go#L442)<br>`func (s *Session) handleRaidExpel(p []byte)` | handleRaidExpel (0x0B1A): leader only. |

### realm.go

[server/internal/world/realm.go](../../server/internal/world/realm.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`realmState`](../../server/internal/world/realm.go#L12)<br>`type realmState` | realmState is the process's authoritative realm. Map shards own entity state and Lua VMs; sessions own characters. These registries own identities/lifetimes, not gameplay mutations. One realm per process keeps deployment lightweight. Separate realms run in separate processes with separate SQLite runtimes. |
| [`newRealm`](../../server/internal/world/realm.go#L31)<br>`func newRealm() *realmState` | See the source definition and the subsystem ownership contract. |
| [`realm`](../../server/internal/world/realm.go#L38)<br>`var realm` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type realmState struct {
	runtimeMu            sync.Mutex
	runtime              *Runtime
	shardsMu             sync.Mutex
	shards               map[uint16]*mapShard
	gridsMu              sync.RWMutex
	grids                map[uint16]mapGrid
	nextNpcID, nextMobID atomic.Uint32
	nextMapItem          atomic.Uint32
	registryMu           sync.Mutex
	registry             map[uint16]map[uint32]*ConnectedPlayer
	activeAccountsMu     sync.Mutex
	activeAccounts       map[uint32]bool
	instanceMu           sync.Mutex
	instanceMaps         map[uint16]uint16
	instanceOwner        map[uint16]uint32
	nextInstance         uint16
}
```

</details>

### rebirth.go

[server/internal/world/rebirth.go](../../server/internal/world/rebirth.go#L1)

Direct GX dependencies: `internal/proto`.

| Declaration | Responsibility / source comment |
|---|---|
| [`reviveAtBind`](../../server/internal/world/rebirth.go#L17)<br>`const reviveAtBind = 2` | 0x0553 request byte. |
| [`reviveWithParty`](../../server/internal/world/rebirth.go#L18)<br>`const reviveWithParty = 3` | 0x0553 request byte. |
| [`reviveRuneAtBind`](../../server/internal/world/rebirth.go#L19)<br>`const reviveRuneAtBind = 4` | 0x0553 request byte. |
| [`reviveRuneHere`](../../server/internal/world/rebirth.go#L20)<br>`const reviveRuneHere = 5` | 0x0553 request byte. |
| [`rebirthKindBind`](../../server/internal/world/rebirth.go#L25)<br>`const rebirthKindBind = 4` | CUser::Rebirth kind byte of 0x0551. |
| [`rebirthKindParty`](../../server/internal/world/rebirth.go#L26)<br>`const rebirthKindParty = 5` | CUser::Rebirth kind byte of 0x0551. |
| [`rebirthDeadline`](../../server/internal/world/rebirth.go#L32)<br>`const rebirthDeadline = 200 * time.Second` | rebirthDeadline is EnterDead's tick + 200000: the dead user is revived at the bind point when it has not chosen. |
| [`partyRebirthWait`](../../server/internal/world/rebirth.go#L33)<br>`const partyRebirthWait = proto.PartyReviveSeconds * time.Second` | See the source definition and the subsystem ownership contract. |
| [`startRebirthTimer`](../../server/internal/world/rebirth.go#L37)<br>`func (s *Session) startRebirthTimer()` | startRebirthTimer is EnterDead's deadline (code 0). |
| [`handleResurrect`](../../server/internal/world/rebirth.go#L42)<br>`func (s *Session) handleResurrect(p []byte)` | handleResurrect (0x0553): mode byte. |
| [`tickRebirth`](../../server/internal/world/rebirth.go#L59)<br>`func (s *Session) tickRebirth(now time.Time)` | tickRebirth is UpdateRebirth's death timer. |
| [`rebirthAtBind`](../../server/internal/world/rebirth.go#L71)<br>`func (s *Session) rebirthAtBind()` | rebirthAtBind applies the normal death penalty and returns the user with full pools. |
| [`killSoulParty`](../../server/internal/world/rebirth.go#L81)<br>`func (s *Session) killSoulParty()` | killSoulParty is KillSoulParty: at the party leader's position with full pools (after the exp loss) when the leader is alive on the same map; otherwise return to the bind point. |
| [`inParty`](../../server/internal/world/rebirth.go#L93)<br>`func (s *Session) inParty() bool` | See the source definition and the subsystem ownership contract. |
| [`partyLeader`](../../server/internal/world/rebirth.go#L102)<br>`func (s *Session) partyLeader() *ConnectedPlayer` | See the source definition and the subsystem ownership contract. |
| [`refillPools`](../../server/internal/world/rebirth.go#L115)<br>`func (s *Session) refillPools()` | refillPools sets current HP, MP and SP to their effective maxima. |
| [`rebirth`](../../server/internal/world/rebirth.go#L122)<br>`func (s *Session) rebirth(mapID uint16, x, y, z float32, kind uint8, restore bool, lost uint32)` | rebirth restores the character and publishes the result to the world. |
| [`reviveOnExit`](../../server/internal/world/rebirth.go#L178)<br>`func (s *Session) reviveOnExit(loss int)` | reviveOnExit settles a dead character that leaves the game, or that an older save kept at 0 HP: it is revived at its bind point with full pools and the usual experience loss, so it re-enters alive and the death is not kept. |
| [`exitDeathLoss`](../../server/internal/world/rebirth.go#L199)<br>`func (s *Session) exitDeathLoss() int` | exitDeathLoss is the experience percent a dead character pays when it leaves. |

### recreation_runes.go

[server/internal/world/recreation_runes.go](../../server/internal/world/recreation_runes.go#L1)

Direct GX dependencies: `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`recreateWithRune`](../../server/internal/world/recreation_runes.go#L14)<br>`func recreateWithRune(rune, target *db.ItemDef, item db.Item) (string, error)` | All recreation runes share Special 62. ReqVg 0 retains the full native roll; 101..109 reroll STR..SP individually; 201..209 set that stat to its maximum. 301..309 remove one stat; 400 removes all nine recreation stats. Existing attributes and enchantment survive a targeted operation. A new stat cannot exceed Server's distinct-attribute budget or the equipment's pool. |

### relic_config.go

[server/internal/world/relic_config.go](../../server/internal/world/relic_config.go#L1)

| Declaration | Responsibility / source comment |
|---|---|
| [`eventSpawn`](../../server/internal/world/relic_config.go#L16)<br>`type eventSpawn` | Obelisk.ini uses repeated [Mob] keys and nested tuples, not an INI key map. Country/mob triplets here stay in wire/config order: neutral, Light, Fury. |
| [`relic`](../../server/internal/world/relic_config.go#L24)<br>`type relic` | See the source definition and the subsystem ownership contract. |
| [`linkedBoss`](../../server/internal/world/relic_config.go#L37)<br>`type linkedBoss` | See the source definition and the subsystem ownership contract. |
| [`relicPortal`](../../server/internal/world/relic_config.go#L54)<br>`type relicPortal` | See the source definition and the subsystem ownership contract. |
| [`relicConfig`](../../server/internal/world/relic_config.go#L63)<br>`type relicConfig` | See the source definition and the subsystem ownership contract. |
| [`eventNumber`](../../server/internal/world/relic_config.go#L69)<br>`var eventNumber` | See the source definition and the subsystem ownership contract. |
| [`eventNumbers`](../../server/internal/world/relic_config.go#L71)<br>`func eventNumbers(s string) ([]float64, error)` | See the source definition and the subsystem ownership contract. |
| [`eventInteger`](../../server/internal/world/relic_config.go#L82)<br>`func eventInteger(v float64, max uint64) bool` | See the source definition and the subsystem ownership contract. |
| [`eventPosition`](../../server/internal/world/relic_config.go#L85)<br>`func eventPosition(n []float64) [3]float32` | See the source definition and the subsystem ownership contract. |
| [`eventMobIDs`](../../server/internal/world/relic_config.go#L88)<br>`func eventMobIDs(n []float64) ([]uint16, error)` | See the source definition and the subsystem ownership contract. |
| [`parseRelicConfig`](../../server/internal/world/relic_config.go#L99)<br>`func parseRelicConfig(raw []byte) (relicConfig, error)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type eventSpawn struct {
	pos    [3]float32
	radius float32
	mobs   []uint16
	count  int
	actors []*eventActor
	next   time.Time
}
```

```go
type relic struct {
	mapID                uint16
	key, id              uint32 // persistent map<<16|number versus sequential client identity
	pos                  [3]float32
	mobs                 [3]uint16
	initial, owner, last uint8
	changeTime           uint32
	principal            *eventActor
	guards               []*eventSpawn
	standing             bool
	next                 time.Time
	pending              *relicCapture
}
```

```go
type linkedBoss struct {
	mapID                  uint16
	key                    uint32
	pos                    [3]float32
	mobs                   []uint16
	guards                 []*eventSpawn
	portalMap              uint16
	portalID               int32
	delay                  time.Duration
	random                 bool
	randomMin, randomRange time.Duration
	randomPositions        [3][3]float32
	stage                  int
	portalOpen             bool
	principal              *eventActor
	next                   time.Time
}
```

```go
type relicPortal struct {
	key, id  uint32
	mapID    uint16
	portalID int32
	refs     [2][]uint32
	open     [2]bool
	country  int32
	pos      [3]float32
}
```

```go
type relicConfig struct {
	relics  []*relic
	bosses  []*linkedBoss
	portals []*relicPortal
}
```

</details>

### relic_portals.go

[server/internal/world/relic_portals.go](../../server/internal/world/relic_portals.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`portalMarker`](../../server/internal/world/relic_portals.go#L10)<br>`func portalMarker(country int32) uint8` | See the source definition and the subsystem ownership contract. |
| [`bindRelicPortals`](../../server/internal/world/relic_portals.go#L21)<br>`func bindRelicPortals(inputs []*relicPortal, saved map[uint32]db.Keep, id *uint32, defaults *[]db.Keep) error` | See the source definition and the subsystem ownership contract. |
| [`setEventPortal`](../../server/internal/world/relic_portals.go#L72)<br>`func setEventPortal(mapID uint16, id, country int32)` | See the source definition and the subsystem ownership contract. |
| [`changeRelicPortal`](../../server/internal/world/relic_portals.go#L87)<br>`func changeRelicPortal(p *relicPortal, open [2]bool, out *[]eventPublication, log logPrinter) bool` | See the source definition and the subsystem ownership contract. |
| [`updateRelicPortalsLocked`](../../server/internal/world/relic_portals.go#L121)<br>`func updateRelicPortalsLocked(out *[]eventPublication, log logPrinter)` | Native all/none hysteresis uses the last nonneutral conqueror, not proximity or the current neutral appearance. Adjacent faction entries exclude each other. |

### relics.go

[server/internal/world/relics.go](../../server/internal/world/relics.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/logx`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`relicListOpcode`](../../server/internal/world/relics.go#L21)<br>`const relicListOpcode = 0x2101` | Original reference: ps_game.exe SHA256 0fb7c9f...12bc547. The controller always locks before a shard. Combat calls it only after releasing the shard; ordinary AI never locks this controller. |
| [`relicOwnerOpcode`](../../server/internal/world/relics.go#L22)<br>`const relicOwnerOpcode = 0x2102` | Original reference: ps_game.exe SHA256 0fb7c9f...12bc547. The controller always locks before a shard. Combat calls it only after releasing the shard; ordinary AI never locks this controller. |
| [`relicNoticeOpcode`](../../server/internal/world/relics.go#L23)<br>`const relicNoticeOpcode = 0x2103` | Original reference: ps_game.exe SHA256 0fb7c9f...12bc547. The controller always locks before a shard. Combat calls it only after releasing the shard; ordinary AI never locks this controller. |
| [`eventActor`](../../server/internal/world/relics.go#L26)<br>`type eventActor` | See the source definition and the subsystem ownership contract. |
| [`relicCapture`](../../server/internal/world/relics.go#L36)<br>`type relicCapture` | See the source definition and the subsystem ownership contract. |
| [`eventPublication`](../../server/internal/world/relics.go#L41)<br>`type eventPublication` | See the source definition and the subsystem ownership contract. |
| [`relics`](../../server/internal/world/relics.go#L47)<br>`var relics` | See the source definition and the subsystem ownership contract. |
| [`runtimeRelicCountry`](../../server/internal/world/relics.go#L58)<br>`func runtimeRelicCountry(owner uint8) uint8` | See the source definition and the subsystem ownership contract. |
| [`wireRelicCountry`](../../server/internal/world/relics.go#L59)<br>`func wireRelicCountry(country uint8) uint8` | See the source definition and the subsystem ownership contract. |
| [`relicKeep`](../../server/internal/world/relics.go#L60)<br>`func relicKeep(r *relic) db.Keep` | See the source definition and the subsystem ownership contract. |
| [`restoreRelic`](../../server/internal/world/relics.go#L63)<br>`func restoreRelic(r *relic, k db.Keep) error` | See the source definition and the subsystem ownership contract. |
| [`buildRelicOwner`](../../server/internal/world/relics.go#L77)<br>`func buildRelicOwner(id uint32, owner uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildRelicNotice`](../../server/internal/world/relics.go#L83)<br>`func buildRelicNotice(mob uint16, old, next uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`publishEvents`](../../server/internal/world/relics.go#L90)<br>`func publishEvents(list []eventPublication)` | See the source definition and the subsystem ownership contract. |
| [`LoadRelics`](../../server/internal/world/relics.go#L102)<br>`func LoadRelics(log *logx.Logger) error` | LoadRelics restores the existing Keeps table, then populates primary maps. No schema/seed rewrite and no synthetic character/quest/kill credit occurs. |
| [`allocateEventActor`](../../server/internal/world/relics.go#L182)<br>`func allocateEventActor(s *mapShard, mapID uint16, r *relic, b *linkedBoss, g *eventSpawn, principal bool) *eventActor` | See the source definition and the subsystem ownership contract. |
| [`spawnEventActor`](../../server/internal/world/relics.go#L191)<br>`func spawnEventActor(s *mapShard, a *eventActor, id uint16, pos [3]float32, now time.Time) bool` | See the source definition and the subsystem ownership contract. |
| [`removeEventActor`](../../server/internal/world/relics.go#L232)<br>`func removeEventActor(s *mapShard, a *eventActor)` | See the source definition and the subsystem ownership contract. |
| [`eventSpawnPosition`](../../server/internal/world/relics.go#L241)<br>`func eventSpawnPosition(mapID uint16, g *eventSpawn) ([3]float32, bool)` | See the source definition and the subsystem ownership contract. |
| [`populateEventSpawn`](../../server/internal/world/relics.go#L259)<br>`func populateEventSpawn(s *mapShard, g *eventSpawn, stage int, now time.Time)` | See the source definition and the subsystem ownership contract. |
| [`spawnRelicLocked`](../../server/internal/world/relics.go#L274)<br>`func spawnRelicLocked(s *mapShard, r *relic, now time.Time)` | See the source definition and the subsystem ownership contract. |
| [`tickEventSpawn`](../../server/internal/world/relics.go#L284)<br>`func tickEventSpawn(s *mapShard, g *eventSpawn, stage int, now time.Time)` | See the source definition and the subsystem ownership contract. |
| [`eventMobDestroyed`](../../server/internal/world/relics.go#L303)<br>`func eventMobDestroyed(killer *ConnectedPlayer, mapID uint16, m *MapMob, log logPrinter)` | eventMobDestroyed receives the already resolved single mob death, including DoTs. It removes dependent populations immediately; controller ticks never depend on nearby players and cannot resurrect a stale guard generation. |
| [`commitRelicCapture`](../../server/internal/world/relics.go#L359)<br>`func commitRelicCapture(r *relic, out *[]eventPublication, log logPrinter) *relicCapture` | See the source definition and the subsystem ownership contract. |
| [`scoreRelicCapture`](../../server/internal/world/relics.go#L384)<br>`func scoreRelicCapture(c relicCapture, log logPrinter)` | See the source definition and the subsystem ownership contract. |
| [`syncEventVisibility`](../../server/internal/world/relics.go#L396)<br>`func syncEventVisibility(mapID uint16)` | See the source definition and the subsystem ownership contract. |
| [`TickRelics`](../../server/internal/world/relics.go#L433)<br>`func TickRelics(now time.Time, log logPrinter)` | TickRelics runs independently of map viewers. The mutex also serializes captures/reset/portal changes; packet transmission follows the lock release. |
| [`ResetRelics`](../../server/internal/world/relics.go#L493)<br>`func ResetRelics(now time.Time, log logPrinter)` | ResetRelics is the blessing-activation reset: configured base ownership, guards removed, no capture scoring/notices, then a ten-minute spawn delay. |
| [`relicListFor`](../../server/internal/world/relics.go#L549)<br>`func relicListFor(mapID uint16) [][]byte` | See the source definition and the subsystem ownership contract. |
| [`sendRelicList`](../../server/internal/world/relics.go#L583)<br>`func (s *Session) sendRelicList()` | See the source definition and the subsystem ownership contract. |
| [`stationary`](../../server/internal/world/relics.go#L588)<br>`func (m *MapMob) stationary() bool` | See the source definition and the subsystem ownership contract. |
| [`mobFriendly`](../../server/internal/world/relics.go#L591)<br>`func mobFriendly(def *db.MobDef, faction uint8) bool` | See the source definition and the subsystem ownership contract. |
| [`eventDamageEntry`](../../server/internal/world/relics.go#L597)<br>`type eventDamageEntry` | Managed principals always use the original fair-combat damage owner, even when another player landed the final hit. Aggro cannot substitute for damage. |
| [`recordEventDamage`](../../server/internal/world/relics.go#L602)<br>`func (m *MapMob) recordEventDamage(id, damage uint32)` | See the source definition and the subsystem ownership contract. |
| [`eventDamageWinner`](../../server/internal/world/relics.go#L614)<br>`func eventDamageWinner(m *MapMob) uint32` | See the source definition and the subsystem ownership contract. |
| [`eventCreditedPlayer`](../../server/internal/world/relics.go#L623)<br>`func eventCreditedPlayer(last *ConnectedPlayer, mapID uint16, m *MapMob) *ConnectedPlayer` | See the source definition and the subsystem ownership contract. |
| [`mobDeathCredit`](../../server/internal/world/relics.go#L638)<br>`func mobDeathCredit(m *MapMob, last uint32) uint32` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type eventActor struct {
	mapID      uint16
	slot       int
	relic      *relic
	boss       *linkedBoss
	spawn      *eventSpawn
	principal  bool
	handled    bool
	generation uint64
}
```

```go
type relicCapture struct {
	owner, old uint8
	mob        uint16
	score      bool
}
```

```go
type eventPublication struct {
	mapID  uint16
	global bool
	frame  []byte
}
```

```go
type eventDamageEntry struct {
	ID     uint32
	Amount uint64
}
```

</details>

### rules.go

[server/internal/world/rules.go](../../server/internal/world/rules.go#L1)

Direct GX dependencies: `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`allowedJob`](../../server/internal/world/rules.go#L18)<br>`func allowedJob(family, job uint8) bool` | allowedJob is the family/job mask of CUser::PacketCharacter (0x0046bf20): humans fighter/defender/priest, elves ranger/archer/mage, vail ranger/mage/oracle, deatheaters warrior/guardian/hunter. |
| [`startingStats`](../../server/internal/world/rules.go#L25)<br>`func startingStats(family, job uint8) (str, dex, rec, intel, wis, luc uint16)` | startingStats are CUser::ProcessCharacterCreate (0x00480e90)'s base stats per job (str, dex, rec, int, wis, luc) plus the family bonus. |
| [`familyPoolBonus`](../../server/internal/world/rules.go#L46)<br>`func familyPoolBonus(family uint8) (hp, mp, sp uint32)` | familyPoolBonus is the per-family flat HP/MP/SP of CUser::InitEquipment. |
| [`statusRow`](../../server/internal/world/rules.go#L62)<br>`func statusRow(level uint16, job uint8) db.Status` | statusRow is StatusDefs for a job and level (falling back to the last defined level). |
| [`basePools`](../../server/internal/world/rules.go#L77)<br>`func basePools(ch *db.Character) (hp, mp, sp uint32)` | basePools are a character's base max HP/MP/SP per CUser::InitEquipment (0x00454560): StatusDefs scaled by 1.5, the family bonus and 5 per point of REC (HP), DEX (SP) and WIS (MP). Equipment adds on top. |
| [`newCharacterPools`](../../server/internal/world/rules.go#L93)<br>`func newCharacterPools(family, job uint8, rec, dex, wis uint16) (hp, mp, sp uint16)` | newCharacterPools are the HP/MP/SP CUser::ProcessCharacterCreate stores: level-1 StatusDefs times 1.5, plus the family bonus and the stat-derived part. |
| [`starterItems`](../../server/internal/world/rules.go#L104)<br>`func starterItems(family, job uint8) []db.Item` | starterItems places BaseItemsDefs as CGameData::LoadBaseItem (0x004bf800) does: the configured Slot and Quality are ignored, the first item of type 1..19 is equipped as the weapon (slot 5) and the rest fill bag 1. |
| [`levelUp`](../../server/internal/world/rules.go#L128)<br>`func levelUp(ch *db.Character)` | levelUp is StatusRules::level_up (CUser level-up path): +1 to the job's primary stat, nine stat points and five skill points, and the pools follow the new level. |
| [`mobExperience`](../../server/internal/world/rules.go#L152)<br>`func mobExperience(mobLevel, base, userLevel uint16) uint32` | mobExperience is CMob::GetExp (0x00491690) without the event multiplier: base * (1 + 0.2 * (mobLevel - userLevel + 1)), nothing when the mob is more than 9 levels above, never negative. |
| [`sittingRecovery`](../../server/internal/world/rules.go#L168)<br>`func sittingRecovery(maxValue uint32, extraPct int32, flat uint32) uint32` | sittingRecovery is CUser::UpdateRecover's sitting amount: 5% of the maximum, plus the sitting-recovery abilities' percent (10-12), plus the flat sitting recovery, truncated (the FISTP runs in chop mode). |
| [`resetStats`](../../server/internal/world/rules.go#L176)<br>`func resetStats(ch *db.Character)` | resetStats undoes stat allocation: the job's starting stats plus the automatic +1 per level, and every stat point the levels granted. |
| [`equipSlotRealType`](../../server/internal/world/rules.go#L195)<br>`var equipSlotRealType` | equipSlotRealType is the real item type each equipment slot takes in CUser::EnableEquipment (0x0045b450); 0 means any weapon, -1 no check. |
| [`isWeaponReal`](../../server/internal/world/rules.go#L198)<br>`func isWeaponReal(rt uint8) bool` | See the source definition and the subsystem ownership contract. |
| [`isOneHandReal`](../../server/internal/world/rules.go#L199)<br>`func isOneHandReal(rt uint8) bool` | See the source definition and the subsystem ownership contract. |
| [`canEquip`](../../server/internal/world/rules.go#L204)<br>`func canEquip(ch *db.Character, def *db.ItemDef, slot uint8, items []db.Item) bool` | canEquip is CUser::EnableEquipment: level, slot type, two-handed weapons against shields, race (Items.Country), class and the sex-only items (Special 64/65). |
| [`equipFailure`](../../server/internal/world/rules.go#L208)<br>`func equipFailure(ch *db.Character, def *db.ItemDef, slot uint8, items []db.Item) string` | See the source definition and the subsystem ownership contract. |

### safezones.go

[server/internal/world/safezones.go](../../server/internal/world/safezones.go#L1)

Direct GX dependencies: `internal/gameini`, `internal/proto`, `server/internal/db`, `server/internal/logx`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`safeFaction`](../../server/internal/world/safezones.go#L17)<br>`type safeFaction` | See the source definition and the subsystem ownership contract. |
| [`safeLight`](../../server/internal/world/safezones.go#L20)<br>`const safeLight = 1 << FactionLight` | See the source definition and the subsystem ownership contract. |
| [`safeFury`](../../server/internal/world/safezones.go#L21)<br>`const safeFury = 1 << FactionDark` | See the source definition and the subsystem ownership contract. |
| [`safeBoth`](../../server/internal/world/safezones.go#L22)<br>`const safeBoth = safeLight &#124; safeFury` | See the source definition and the subsystem ownership contract. |
| [`safePoint`](../../server/internal/world/safezones.go#L25)<br>`type safePoint` | See the source definition and the subsystem ownership contract. |
| [`safeZone`](../../server/internal/world/safezones.go#L27)<br>`type safeZone` | See the source definition and the subsystem ownership contract. |
| [`safeZones`](../../server/internal/world/safezones.go#L32)<br>`var safeZones map[uint16][]safeZone` | See the source definition and the subsystem ownership contract. |
| [`loadSafeZones`](../../server/internal/world/safezones.go#L34)<br>`func loadSafeZones(log *logx.Logger) error` | See the source definition and the subsystem ownership contract. |
| [`parseSafeZones`](../../server/internal/world/safezones.go#L48)<br>`func parseSafeZones(ini iniFile, maps map[uint16]ZoneConfig) (map[uint16][]safeZone, error)` | See the source definition and the subsystem ownership contract. |
| [`sideOfSafeEdge`](../../server/internal/world/safezones.go#L94)<br>`func sideOfSafeEdge(a, b, p safePoint) float64` | See the source definition and the subsystem ownership contract. |
| [`contains`](../../server/internal/world/safezones.go#L100)<br>`func (z safeZone) contains(p safePoint) bool` | EP5.4 0x438650 accepts four strictly same-sided vertical planes. |
| [`safeProtectionAt`](../../server/internal/world/safezones.go#L110)<br>`func safeProtectionAt(zones []safeZone, x, z float32) safeFaction` | See the source definition and the subsystem ownership contract. |
| [`protects`](../../server/internal/world/safezones.go#L120)<br>`func (p safeFaction) protects(faction uint8) bool` | See the source definition and the subsystem ownership contract. |
| [`safeZoneProtection`](../../server/internal/world/safezones.go#L124)<br>`func (s *Session) safeZoneProtection() safeFaction` | See the source definition and the subsystem ownership contract. |
| [`safeZoneBlocksUser`](../../server/internal/world/safezones.go#L130)<br>`func (s *Session) safeZoneBlocksUser(target *ConnectedPlayer) bool` | A faction sanctuary protects residents and lets them attack the other faction. A sanctuary for both factions also blocks outgoing attacks. |
| [`safeZoneBlocksSkill`](../../server/internal/world/safezones.go#L144)<br>`func (s *Session) safeZoneBlocksSkill(def *db.SkillDef, op packet.Opcode, target uint32) bool` | See the source definition and the subsystem ownership contract. |
| [`rejectSafeZoneSkill`](../../server/internal/world/safezones.go#L180)<br>`func (s *Session) rejectSafeZoneSkill(k db.CharacterSkill, prepared bool)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type safeFaction uint8
```

```go
type safePoint struct{ x, z float32 }
```

```go
type safeZone struct {
	faction safeFaction
	corners [4]safePoint
}
```

</details>

### session.go

[server/internal/world/session.go](../../server/internal/world/session.go#L1)

Direct GX dependencies: `internal/proto`, `internal/skilltiming`, `server/internal/config`, `server/internal/crypto`, `server/internal/db`, `server/internal/logx`, `server/internal/netx`, `server/internal/packet`, `server/internal/session`.

| Declaration | Responsibility / source comment |
|---|---|
| [`loopTick`](../../server/internal/world/session.go#L25)<br>`const loopTick = 100 * time.Millisecond` | See the source definition and the subsystem ownership contract. |
| [`recoveryInterval`](../../server/internal/world/session.go#L26)<br>`const recoveryInterval = 5 * time.Second` | See the source definition and the subsystem ownership contract. |
| [`characterListSize`](../../server/internal/world/session.go#L27)<br>`const characterListSize = MaxCharacterSlots` | See the source definition and the subsystem ownership contract. |
| [`Session`](../../server/internal/world/session.go#L32)<br>`type Session` | Session is one client connection. Its loop owns its fields under stateMu. A paired trade may also update them while holding both sessions' state locks. |
| [`acquireAccount`](../../server/internal/world/session.go#L147)<br>`func acquireAccount(uid uint32, id int64) bool` | See the source definition and the subsystem ownership contract. |
| [`releaseAccount`](../../server/internal/world/session.go#L165)<br>`func releaseAccount(uid uint32, id int64)` | See the source definition and the subsystem ownership contract. |
| [`Serve`](../../server/internal/world/session.go#L176)<br>`func Serve(conn net.Conn, logger *logx.Logger, log *logx.Tag) error` | Serve runs one client connection until it closes. |
| [`processInput`](../../server/internal/world/session.go#L213)<br>`func (s *Session) processInput(n int, buf []byte, frames *netx.Frames) error` | processInput never holds the session lock while waiting for socket input. |
| [`close`](../../server/internal/world/session.go#L257)<br>`func (s *Session) close()` | close persists the active character and releases the account. |
| [`leaveWorld`](../../server/internal/world/session.go#L270)<br>`func (s *Session) leaveWorld(disconnect bool) bool` | leaveWorld saves and unregisters the active character. |
| [`send`](../../server/internal/world/session.go#L301)<br>`func (s *Session) send(frame []byte, label string)` | send writes a frame to this client: AES-CTR until the initial character state, the stateless expanded XOR afterwards. |
| [`equipment`](../../server/internal/world/session.go#L323)<br>`func (s *Session) equipment() EquipmentStats` | See the source definition and the subsystem ownership contract. |
| [`stats`](../../server/internal/world/session.go#L348)<br>`func (s *Session) stats() CharacterStats` | See the source definition and the subsystem ownership contract. |
| [`sendStatPackets`](../../server/internal/world/session.go#L356)<br>`func (s *Session) sendStatPackets()` | See the source definition and the subsystem ownership contract. |
| [`drainRewards`](../../server/internal/world/session.go#L369)<br>`func (s *Session) drainRewards(p *ConnectedPlayer)` | drainRewards applies bounded, coalesced rewards before normal simulation or logout. Aggregation keeps individual scaling/rounding and caps unchanged. |
| [`normalize`](../../server/internal/world/session.go#L420)<br>`func normalize(ch *db.Character)` | normalize derives the base pools from StatusDefs (Chars only stores the current HP/MP/SP). Effective clamping waits for equipment and buffs. |
| [`tick`](../../server/internal/world/session.go#L427)<br>`func (s *Session) tick(idle bool)` | tick runs the per-iteration work: mailboxes, buffs, regen, casts and the auto-attack swing. |
| [`applyTradeRefresh`](../../server/internal/world/session.go#L522)<br>`func (s *Session) applyTradeRefresh(r *tradeRefresh)` | See the source definition and the subsystem ownership contract. |
| [`recoveryTick`](../../server/internal/world/session.go#L547)<br>`func (s *Session) recoveryTick()` | recoveryTick is CUser::UpdateRecover every 5 seconds: the mode seen at the previous tick decides the amount (0 sitting: 5% of each maximum plus the sitting abilities' percent plus the flat sitting recovery; 1 standing: the flat standing recovery; 2 in combat: the flat in-combat recovery), then the mode for the next tick is taken (in combat, else sitting or standing). Viewers get 0x0505 (absolute pools) and the character 0x0521. |
| [`inCombat`](../../server/internal/world/session.go#L593)<br>`func (s *Session) inCombat() bool` | inCombat approximates the original attack mode: swinging, or a hit given or taken within the last recovery interval. |
| [`deadGated`](../../server/internal/world/session.go#L598)<br>`func deadGated(op packet.Opcode) bool` | deadGated lists the inputs ignored while dead (revive is 0x0553). |
| [`frame`](../../server/internal/world/session.go#L607)<br>`func (s *Session) frame(f []byte) error` | See the source definition and the subsystem ownership contract. |
| [`handshake`](../../server/internal/world/session.go#L631)<br>`func (s *Session) handshake(f []byte) error` | See the source definition and the subsystem ownership contract. |
| [`sendCharacterSelect`](../../server/internal/world/session.go#L672)<br>`func (s *Session) sendCharacterSelect()` | See the source definition and the subsystem ownership contract. |
| [`logErr`](../../server/internal/world/session.go#L694)<br>`func (s *Session) logErr(what string, err error)` | See the source definition and the subsystem ownership contract. |
| [`handleCreateCharacter`](../../server/internal/world/session.go#L701)<br>`func (s *Session) handleCreateCharacter(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`createCharacter`](../../server/internal/world/session.go#L710)<br>`func (s *Session) createCharacter(nc db.NewCharacter)` | See the source definition and the subsystem ownership contract. |
| [`handleRename`](../../server/internal/world/session.go#L756)<br>`func (s *Session) handleRename(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`selectCharacter`](../../server/internal/world/session.go#L777)<br>`func (s *Session) selectCharacter(id uint32)` | See the source definition and the subsystem ownership contract. |
| [`sendInitialState`](../../server/internal/world/session.go#L917)<br>`func (s *Session) sendInitialState(st *CharacterStats)` | sendInitialState is the captured character-select sequence; afterwards the server->client direction switches to the expanded XOR scheme. |
| [`handleMapLoaded`](../../server/internal/world/session.go#L970)<br>`func (s *Session) handleMapLoaded()` | handleMapLoaded (0x0201 from the client): send the map state, register the player and exchange visibility with nearby players. |
| [`handleMotion`](../../server/internal/world/session.go#L1020)<br>`func (s *Session) handleMotion(p []byte)` | handleMotion (0x0506): sit/stand persist; other motions are one-shot. The sender gets its own echo, which the client needs to play it. |
| [`handleLogout`](../../server/internal/world/session.go#L1063)<br>`func (s *Session) handleLogout()` | handleLogout (0x0107) returns to character select on the same connection; the account stays held. |

<details><summary>Type definitions and owned fields</summary>

```go
type Session struct {
	gmForced          bool      // a GM's /asummon is moving this user: no party or guild instance redirection
	assisted          [3]uint32 // HP, SP, MP of the last hit after Health Assistant
	targetHealthSent  [3]uint32 // user, HP and maximum last told for the selected user
	moveRune          *movementRuneCast
	teleportRune      *teleportRuneCast
	traps             []placedTrap
	nextTrapAt        time.Time
	economy           *netx.RateLimiter
	awaitingMapLoad   bool
	inputReads        *netx.RateLimiter
	lastInputTick     time.Time
	persistenceFailed bool
	operation         *inventoryOperation
	inputFrames       *netx.RateLimiter
	inputBytes        *netx.RateLimiter
	stateMu           sync.Mutex
	shop              myShop // the personal market this character sets up
	effectWatchID     uint32
	effectWatchKind   uint8
	effectWatchAt     time.Time
	effectWatchHash   [32]byte
	speedKnown        bool
	lastSpeeds        playerSpeeds
	movement          playerMovement
	oxygen            oxygenState
	questNPCCatalog   questNPCCatalogState
	gmTurbo           bool
	gmImmune          bool
	gmConcealed       bool
	lastGMModeAt      time.Time
	elements          elementRuntime
	socialChecked     time.Time
	socialReloadAt    time.Time
	socialFriends     []db.SocialContact
	socialRefresh     atomic.Bool
	conn              net.Conn
	out               *connWriter
	log               *logx.Tag
	logger            *logx.Logger

	aesRecv, aesSend *crypto.AESCTR
	expanded         bool

	accountID         uint32
	sessionID         int64
	credentialVersion int64
	authorizationAt   time.Time
	account           db.Account
	faction           uint8
	isAdmin           bool

	characters []db.Character
	char       *db.Character
	items      []db.Item
	stored     []db.Item
	skills     []db.CharacterSkill
	quickSlots []db.QuickSlot
	points     uint32
	player     *ConnectedPlayer

	itemCooldowns                   map[uint8]time.Time
	skillCooldowns                  map[uint16]time.Time
	evolutionReadyAt                map[uint16]time.Time
	evolutionSerial, evolutionGrant uint32
	evolutionActions                [2]evolutionAction
	buffs                           []ActiveBuff
	cast                            *pendingCast
	mounted                         bool
	mountItemUID                    int64
	dead                            bool
	deathLoss                       int // CUser::Death exp-loss percent, kept until revive
	statusItemProtection            uint16
	rebirthAt                       time.Time // UpdateRebirth deadline while dead
	rebirthParty                    bool      // 0x0553 mode 3 waiting for the party leader

	autoAttack       bool
	autoAttackTarget uint32
	autoAttackUser   bool          // the target is a user (0x0502), else a mob
	damageLedger     []damageEntry // CDamage: HP damage taken per attacker since the last death
	duelDeathMark    bool          // the last death was a lost duel (m_nVehicleStatus -1)
	itemAttacking    bool          // m_bItemAttack: casting gem skills
	lastSkillID      uint16
	lastSkillAt      time.Time
	pendingCombo     *pendingSpinCombo
	pendingSkill     *pendingSkillUse
	gmGRBGuild       uint32 // guild whose ranking participation this GM started
	dash             *proto.SkillDash
	lastActionAt     time.Time
	clipFloorAt      time.Time     // the lastActionAt the floor below belongs to
	clipFloor        time.Duration // own animation of the last skill
	attackMoveUntil  time.Time

	positionDirty        bool
	lastPositionSave     time.Time
	lastRegenAt          time.Time
	lastQuestCheck       time.Time
	lastWeather          uint8
	weatherOverride      uint8
	weatherOverrideUntil time.Time
	weatherCheck         time.Time
	usingNpc             uint32
	auctionContact       bool
	quests               []*activeQuest
	finishedQuests       []db.FinishedQuest
	lastQuestItem        questItemUse
	regenMode            int // CUser +0x1274: 0 sitting, 1 standing, 2 in combat (next tick's recovery)
	lastCombatAt         time.Time
	teleportStone        db.Item
	teleportCast         *teleportStoneCast
	summonCastEnd        time.Time
	shoutReadyAt         time.Time // 0x1107 cooldown (CUser +0x5848)
	returnAt             time.Time // 0x0556 pending move (CUser +0x581c, type 0)
}
```

</details>

### skill_dash.go

[server/internal/world/skill_dash.go](../../server/internal/world/skill_dash.go#L1)

Direct GX dependencies: `internal/mapgeometry`, `internal/proto`, `internal/skilltiming`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`beginStinger`](../../server/internal/world/skill_dash.go#L15)<br>`func (s *Session) beginStinger(def *db.SkillDef, destination [3]float32) bool` | Stinger is a bounded server-authored sweep, never a client speed allowance. The selected target is captured once, as in the original result handler. |
| [`tickStinger`](../../server/internal/world/skill_dash.go#L44)<br>`func (s *Session) tickStinger() bool` | See the source definition and the subsystem ownership contract. |
| [`buildSkillDash`](../../server/internal/world/skill_dash.go#L86)<br>`func buildSkillDash(d proto.SkillDash) []byte` | See the source definition and the subsystem ownership contract. |

### skill_range.go

[server/internal/world/skill_range.go](../../server/internal/world/skill_range.go#L1)

Direct GX dependencies: `internal/mapgeometry`, `internal/proto`, `internal/skillrange`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`skillLineClear`](../../server/internal/world/skill_range.go#L11)<br>`func (s *Session) skillLineClear(a, b skillrange.Position) bool` | See the source definition and the subsystem ownership contract. |
| [`skillTargetReach`](../../server/internal/world/skill_range.go#L16)<br>`func (s *Session) skillTargetReach(p skillrange.Position, base, size float32, completed bool) bool` | See the source definition and the subsystem ownership contract. |
| [`mobSkillReach`](../../server/internal/world/skill_range.go#L24)<br>`func (s *Session) mobSkillReach(m *MapMob, def *db.SkillDef, completed bool) bool` | See the source definition and the subsystem ownership contract. |
| [`userSkillReach`](../../server/internal/world/skill_range.go#L31)<br>`func (s *Session) userSkillReach(p *ConnectedPlayer, def *db.SkillDef, completed bool) bool` | See the source definition and the subsystem ownership contract. |
| [`skillAreaContains`](../../server/internal/world/skill_range.go#L38)<br>`func (s *Session) skillAreaContains(center, target skillrange.Position, radius float32) bool` | See the source definition and the subsystem ownership contract. |
| [`sendSkillRangeFailure`](../../server/internal/world/skill_range.go#L42)<br>`func (s *Session) sendSkillRangeFailure(k db.CharacterSkill, target ...skillrange.Position)` | See the source definition and the subsystem ownership contract. |
| [`sendSkillMessage`](../../server/internal/world/skill_range.go#L50)<br>`func (s *Session) sendSkillMessage(skillID, message uint16)` | See the source definition and the subsystem ownership contract. |
| [`sendUserSkillRangeFailure`](../../server/internal/world/skill_range.go#L56)<br>`func (s *Session) sendUserSkillRangeFailure(k db.CharacterSkill, p *ConnectedPlayer)` | See the source definition and the subsystem ownership contract. |
| [`sendMobSkillRangeFailure`](../../server/internal/world/skill_range.go#L64)<br>`func (s *Session) sendMobSkillRangeFailure(k db.CharacterSkill, m *MapMob)` | See the source definition and the subsystem ownership contract. |

### skillattack.go

[server/internal/world/skillattack.go](../../server/internal/world/skillattack.go#L1)

Direct GX dependencies: `internal/skillrange`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`opSkillDrain`](../../server/internal/world/skillattack.go#L19)<br>`const opSkillDrain = 0x0518` | See the source definition and the subsystem ownership contract. |
| [`opSkillOther`](../../server/internal/world/skillattack.go#L20)<br>`const opSkillOther = 0x0519` | See the source definition and the subsystem ownership contract. |
| [`resultAnnounce`](../../server/internal/world/skillattack.go#L21)<br>`const resultAnnounce = 4` | See the source definition and the subsystem ownership contract. |
| [`detailAttack`](../../server/internal/world/skillattack.go#L23)<br>`const detailAttack = 21` | See the source definition and the subsystem ownership contract. |
| [`detailMulti`](../../server/internal/world/skillattack.go#L24)<br>`const detailMulti = 22` | See the source definition and the subsystem ownership contract. |
| [`detailSacrifice`](../../server/internal/world/skillattack.go#L25)<br>`const detailSacrifice = 23` | See the source definition and the subsystem ownership contract. |
| [`detailNoElement`](../../server/internal/world/skillattack.go#L26)<br>`const detailNoElement = 24` | See the source definition and the subsystem ownership contract. |
| [`detailDrain`](../../server/internal/world/skillattack.go#L27)<br>`const detailDrain = 27` | See the source definition and the subsystem ownership contract. |
| [`typeEffectAttack`](../../server/internal/world/skillattack.go#L28)<br>`const typeEffectAttack = 1` | See the source definition and the subsystem ownership contract. |
| [`typeEffectHeal`](../../server/internal/world/skillattack.go#L29)<br>`const typeEffectHeal = 2` | See the source definition and the subsystem ownership contract. |
| [`typeEffectDebuff`](../../server/internal/world/skillattack.go#L30)<br>`const typeEffectDebuff = 3` | See the source definition and the subsystem ownership contract. |
| [`typeEffectBuff`](../../server/internal/world/skillattack.go#L31)<br>`const typeEffectBuff = 4` | See the source definition and the subsystem ownership contract. |
| [`skillCost`](../../server/internal/world/skillattack.go#L36)<br>`func skillCost(cost uint16, reduction int32, pool uint32) uint32` | skillCost is CUser::UseSkill's cost: the reduction is a percentage of the cost; a cost of 1 spends the whole pool. |
| [`skillCosts`](../../server/internal/world/skillattack.go#L44)<br>`func (s *Session) skillCosts(def *db.SkillDef, st *CharacterStats) (sp, mp uint32)` | See the source definition and the subsystem ownership contract. |
| [`skillAttack`](../../server/internal/world/skillattack.go#L49)<br>`func (s *Session) skillAttack(k db.CharacterSkill, def *db.SkillDef, target uint32, castPos *[2]float32, stats CharacterStats)` | skillAttack runs an attack skill (TypeEffect 1) or a debuff (3) on mobs. |
| [`skillStrike`](../../server/internal/world/skillattack.go#L123)<br>`func (s *Session) skillStrike(target uint32, def *db.SkillDef, k db.CharacterSkill, first bool)` | skillStrike is one target of a skill: detail 22 hits MultiAttack times, stopping when the mob or the caster dies. |
| [`afterSkillStrikes`](../../server/internal/world/skillattack.go#L137)<br>`func (s *Session) afterSkillStrikes(def *db.SkillDef)` | Detail 23 pays the caster's entire HP once after an accepted attack resolves, including a miss. EP4.5 does this after the target loop at 0x451484. |
| [`skillMobsWithin`](../../server/internal/world/skillattack.go#L163)<br>`func (s *Session) skillMobsWithin(x, y, z, r float32) []uint32` | See the source definition and the subsystem ownership contract. |
| [`skillHit`](../../server/internal/world/skillattack.go#L176)<br>`func (s *Session) skillHit(target uint32, def *db.SkillDef, k db.CharacterSkill, first bool) bool` | skillHit is CUser::AttackMob with a skill. It returns whether the mob is still alive. |
| [`buildSkillDrainResult`](../../server/internal/world/skillattack.go#L281)<br>`func buildSkillDrainResult(result uint8, sender, target uint32, skillID uint16, level uint8, damage, healed [3]uint16) []byte` | buildSkillDrainResult (0x0518): the 0x0517 layout plus the healed HP/SP/MP. |
| [`buildDrainPacket`](../../server/internal/world/skillattack.go#L287)<br>`func buildDrainPacket(op packet.Opcode, result uint8, sender, target uint32, skillID uint16, level uint8, damage, healed [3]uint16) []byte` | buildDrainPacket is the drain layout shared by 0x0518 (on a mob) and 0x0512 (on a user). |
| [`inSkillRange`](../../server/internal/world/skillattack.go#L305)<br>`func (s *Session) inSkillRange(m *MapMob, def *db.SkillDef, st *CharacterStats) bool` | inSkillRange checks launch reach, including mob size and body allowance, current height and line of sight. Release grace is reserved for prepared casts. |
| [`clearBuffsOnDeath`](../../server/internal/world/skillattack.go#L311)<br>`func (s *Session) clearBuffsOnDeath()` | clearBuffsOnDeath is CUser::ClearApplySkillByDeath: every active skill ends except those with FixRange set. |

### skillcast.go

[server/internal/world/skillcast.go](../../server/internal/world/skillcast.go#L1)

Direct GX dependencies: `internal/proto`, `internal/skillchance`, `internal/skilltiming`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`skillResultNormal`](../../server/internal/world/skillcast.go#L14)<br>`const skillResultNormal = 0` | See the source definition and the subsystem ownership contract. |
| [`targetTypeSelf`](../../server/internal/world/skillcast.go#L16)<br>`const targetTypeSelf = 2` | See the source definition and the subsystem ownership contract. |
| [`targetTypeEnemiesNearCaster`](../../server/internal/world/skillcast.go#L17)<br>`const targetTypeEnemiesNearCaster = 6` | See the source definition and the subsystem ownership contract. |
| [`targetTypeEnemiesNearTarget`](../../server/internal/world/skillcast.go#L18)<br>`const targetTypeEnemiesNearTarget = 7` | See the source definition and the subsystem ownership contract. |
| [`ActiveBuff`](../../server/internal/world/skillcast.go#L23)<br>`type ActiveBuff` | ActiveBuff is a self-buff on the session's character: stat bonuses and/or a heal-over-time. |
| [`pendingCast`](../../server/internal/world/skillcast.go#L42)<br>`type pendingCast` | See the source definition and the subsystem ownership contract. |
| [`buffAbility`](../../server/internal/world/skillcast.go#L52)<br>`type buffAbility` | See the source definition and the subsystem ownership contract. |
| [`addBuffs`](../../server/internal/world/skillcast.go#L59)<br>`func addBuffs(e EquipmentStats, buffs []ActiveBuff) EquipmentStats` | addBuffs accumulates numeric abilities. Elemental application/removal is event-driven in elementstate.go; replaying the remaining list is not native. |
| [`newBuff`](../../server/internal/world/skillcast.go#L73)<br>`func newBuff(id uint32, def *db.SkillDef, now, ends time.Time, log logPrinter) ActiveBuff` | See the source definition and the subsystem ownership contract. |
| [`buildCharacterActiveBuffs`](../../server/internal/world/skillcast.go#L104)<br>`func buildCharacterActiveBuffs(buffs []ActiveBuff) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildSkillUseResult`](../../server/internal/world/skillcast.go#L121)<br>`func buildSkillUseResult(op packet.Opcode, result uint8, sender, target uint32, skillID uint16, level uint8, hp, sp, mp uint16) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildSkillUsedSpMp`](../../server/internal/world/skillcast.go#L134)<br>`func buildSkillUsedSpMp(sp, mp uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`knownSkill`](../../server/internal/world/skillcast.go#L141)<br>`func (s *Session) knownSkill(number uint8) (db.CharacterSkill, *db.SkillDef, bool)` | See the source definition and the subsystem ownership contract. |
| [`skillReady`](../../server/internal/world/skillcast.go#L151)<br>`func (s *Session) skillReady(id uint16) bool` | See the source definition and the subsystem ownership contract. |
| [`previousSkillCooling`](../../server/internal/world/skillcast.go#L159)<br>`func (s *Session) previousSkillCooling(def *db.SkillDef) bool` | previousSkillCooling: PrevSkillID names the skill that must be cooling down for this one to be usable. SysMsg 315 is the client's "A previous skill is required". |
| [`msgPreviousSkillRequired`](../../server/internal/world/skillcast.go#L163)<br>`const msgPreviousSkillRequired = 315` | See the source definition and the subsystem ownership contract. |
| [`msgUnavailableEquipment`](../../server/internal/world/skillcast.go#L167)<br>`const msgUnavailableEquipment = 314` | msgUnavailableEquipment is the native text for skill-use result 7 (client 0x5764d0 -> SysMsg 0x13a). |
| [`handleUseSkill`](../../server/internal/world/skillcast.go#L171)<br>`func (s *Session) handleUseSkill(op packet.Opcode, p []byte)` | handleUseSkill (0x0517 / 0x0511): bar number, target id. The ack echoes the request opcode. |
| [`beginSkill`](../../server/internal/world/skillcast.go#L198)<br>`func (s *Session) beginSkill(op packet.Opcode, k db.CharacterSkill, def *db.SkillDef, target, evolution uint32, p []byte)` | See the source definition and the subsystem ownership contract. |
| [`processPendingCast`](../../server/internal/world/skillcast.go#L344)<br>`func (s *Session) processPendingCast()` | See the source definition and the subsystem ownership contract. |
| [`startCooldown`](../../server/internal/world/skillcast.go#L360)<br>`func (s *Session) startCooldown(k db.CharacterSkill, def *db.SkillDef, prepared ...bool)` | See the source definition and the subsystem ownership contract. |
| [`resolveSkill`](../../server/internal/world/skillcast.go#L398)<br>`func (s *Session) resolveSkill(target uint32, number uint8, op packet.Opcode, cast *pendingCast)` | See the source definition and the subsystem ownership contract. |
| [`resolveSkillDefinition`](../../server/internal/world/skillcast.go#L413)<br>`func (s *Session) resolveSkillDefinition(target uint32, k db.CharacterSkill, def *db.SkillDef, op packet.Opcode, cast *pendingCast)` | See the source definition and the subsystem ownership contract. |
| [`sendBuffStats`](../../server/internal/world/skillcast.go#L470)<br>`func (s *Session) sendBuffStats(why string)` | See the source definition and the subsystem ownership contract. |
| [`processBuffs`](../../server/internal/world/skillcast.go#L480)<br>`func (s *Session) processBuffs()` | processBuffs ticks heal-over-time and debuff damage, then expires applied skills (CUser::ChkKeepSkill / ChkEndTimeSkill). |
| [`processBuffsAt`](../../server/internal/world/skillcast.go#L484)<br>`func (s *Session) processBuffsAt(now time.Time)` | See the source definition and the subsystem ownership contract. |
| [`persistBuffs`](../../server/internal/world/skillcast.go#L507)<br>`func (s *Session) persistBuffs()` | persistBuffs snapshots remaining buff time (frozen while offline). |
| [`buffSnapshot`](../../server/internal/world/skillcast.go#L516)<br>`func (s *Session) buffSnapshot() []db.AppliedBuff` | See the source definition and the subsystem ownership contract. |
| [`buffSnapshotExcept`](../../server/internal/world/skillcast.go#L520)<br>`func (s *Session) buffSnapshotExcept(exclude func(*ActiveBuff) bool) []db.AppliedBuff` | See the source definition and the subsystem ownership contract. |
| [`persistSkillDelays`](../../server/internal/world/skillcast.go#L531)<br>`func (s *Session) persistSkillDelays()` | See the source definition and the subsystem ownership contract. |
| [`skillDelaySnapshot`](../../server/internal/world/skillcast.go#L538)<br>`func (s *Session) skillDelaySnapshot() []db.CharacterSkill` | skillDelaySnapshot captures changes to remaining cooldown. |
| [`restoreBuffs`](../../server/internal/world/skillcast.go#L553)<br>`func restoreBuffs(charID uint32, log logPrinter) ([]ActiveBuff, error)` | See the source definition and the subsystem ownership contract. |
| [`skillWeaponAllowed`](../../server/internal/world/skillcast.go#L574)<br>`func (s *Session) skillWeaponAllowed(def *db.SkillDef) bool` | skillWeaponAllowed follows CUser::EnableUseSkillUser: a skill needs an equipped weapon (error 7 otherwise) and, unless it is a shield skill, the weapon's type must be listed in the skill's NeedWeapon flags. Skills that list no weapon at all (item-granted buffs) are not restricted. |

<details><summary>Type definitions and owned fields</summary>

```go
type ActiveBuff struct {
	ID                     uint32
	MapCurse               bool
	SkillID                uint16
	SkillLevel             uint8
	EndsAt                 time.Time
	Effect, Detail         uint8
	Attr                   uint8 // AttrType, for the element buffs 94/95
	DefenceType            uint8
	BlockLeft              uint32 // detail 85 blocks left / 86 chance
	CasterMob              uint32 // the mob that applied a debuff
	Form                   formLook
	CasterUser             uint32 // the user that applied a debuff (its damage counts as theirs)
	prevTick               time.Time
	Abilities              [3]buffAbility
	HealHP, HealSP, HealMP uint32
	nextTickAt             time.Time // zero: no heal component
}
```

```go
type pendingCast struct {
	targetMobID uint32
	number      uint8
	op          packet.Opcode
	readyAt     time.Time
	targetPos   *[2]float32 // Non-nil marks an accepted prepared cast; distance is checked only at launch.
	corpse      corpseTarget
	evolution   uint32 // granting form generation for a temporary monster skill
}
```

```go
type buffAbility struct {
	Type  uint8
	Value int32
}
```

</details>

### skillonce.go

[server/internal/world/skillonce.go](../../server/internal/world/skillonce.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`typeEffectOnce`](../../server/internal/world/skillonce.go#L14)<br>`const typeEffectOnce = 6` | TypeEffect 6 skills are CUser::SkillOnce 0x00453610 (reached through UseSkillUser / AttackUser), switched on the skill's Type (TypeDetail). |
| [`detailLifeSkill`](../../server/internal/world/skillonce.go#L16)<br>`const detailLifeSkill = 121` | TypeEffect 6 skills are CUser::SkillOnce 0x00453610 (reached through UseSkillUser / AttackUser), switched on the skill's Type (TypeDetail). resurrection: revive a dead user (LifeSkill) |
| [`detailKillSoul`](../../server/internal/world/skillonce.go#L17)<br>`const detailKillSoul = 122` | TypeEffect 6 skills are CUser::SkillOnce 0x00453610 (reached through UseSkillUser / AttackUser), switched on the skill's Type (TypeDetail). send an enemy corpse to its bind point |
| [`detailDetect`](../../server/internal/world/skillonce.go#L18)<br>`const detailDetect = 123` | TypeEffect 6 skills are CUser::SkillOnce 0x00453610 (reached through UseSkillUser / AttackUser), switched on the skill's Type (TypeDetail). reveal an enemy's defence element, level, mode |
| [`detailCheck`](../../server/internal/world/skillonce.go#L19)<br>`const detailCheck = 124` | TypeEffect 6 skills are CUser::SkillOnce 0x00453610 (reached through UseSkillUser / AttackUser), switched on the skill's Type (TypeDetail). success roll only |
| [`detailReturn`](../../server/internal/world/skillonce.go#L20)<br>`const detailReturn = 125` | TypeEffect 6 skills are CUser::SkillOnce 0x00453610 (reached through UseSkillUser / AttackUser), switched on the skill's Type (TypeDetail). move the caster to its bind point |
| [`skillResultFailed`](../../server/internal/world/skillonce.go#L22)<br>`const skillResultFailed = 3` | TypeEffect 6 skills are CUser::SkillOnce 0x00453610 (reached through UseSkillUser / AttackUser), switched on the skill's Type (TypeDetail). |
| [`skillResultNotTarget`](../../server/internal/world/skillonce.go#L23)<br>`const skillResultNotTarget = 11` | TypeEffect 6 skills are CUser::SkillOnce 0x00453610 (reached through UseSkillUser / AttackUser), switched on the skill's Type (TypeDetail). |
| [`pendingLifeSkill`](../../server/internal/world/skillonce.go#L27)<br>`type pendingLifeSkill` | pendingLifeSkill asks a dead user's session to revive at the caster. |
| [`QueueLifeSkill`](../../server/internal/world/skillonce.go#L32)<br>`func (p *ConnectedPlayer) QueueLifeSkill(l pendingLifeSkill)` | See the source definition and the subsystem ownership contract. |
| [`TakeLifeSkill`](../../server/internal/world/skillonce.go#L38)<br>`func (p *ConnectedPlayer) TakeLifeSkill() *pendingLifeSkill` | See the source definition and the subsystem ownership contract. |
| [`corpseTarget`](../../server/internal/world/skillonce.go#L46)<br>`type corpseTarget` | See the source definition and the subsystem ownership contract. |
| [`pendingSoulKill`](../../server/internal/world/skillonce.go#L51)<br>`type pendingSoulKill` | See the source definition and the subsystem ownership contract. |
| [`queueSoulKill`](../../server/internal/world/skillonce.go#L58)<br>`func (p *ConnectedPlayer) queueSoulKill(k pendingSoulKill) bool` | See the source definition and the subsystem ownership contract. |
| [`takeSoulKill`](../../server/internal/world/skillonce.go#L68)<br>`func (p *ConnectedPlayer) takeSoulKill() *pendingSoulKill` | See the source definition and the subsystem ownership contract. |
| [`soulBurnCorpse`](../../server/internal/world/skillonce.go#L77)<br>`func (s *Session) soulBurnCorpse(p *ConnectedPlayer) corpseTarget` | Soul Burn follows the normal PvP team and visibility rules with a dead target. |
| [`soulBurnRejection`](../../server/internal/world/skillonce.go#L88)<br>`func soulBurnRejection(source uint32, k db.CharacterSkill) []byte` | See the source definition and the subsystem ownership contract. |
| [`rejectSoulBurn`](../../server/internal/world/skillonce.go#L92)<br>`func (s *Session) rejectSoulBurn(k db.CharacterSkill)` | See the source definition and the subsystem ownership contract. |
| [`equipmentDefenseAttr`](../../server/internal/world/skillonce.go#L98)<br>`func (p *ConnectedPlayer) equipmentDefenseAttr() uint8` | equipmentDefenseAttr reads CUser +0x1284: live authoritative defense element, including application/removal events; equipment-only before entry. |
| [`buildDetectInfo`](../../server/internal/world/skillonce.go#L110)<br>`func buildDetectInfo(attr uint32, level uint16) []byte` | buildDetectInfo (0x0558): target defence element u32, level u16, mode u8. |
| [`userSkillTarget`](../../server/internal/world/skillonce.go#L120)<br>`func (s *Session) userSkillTarget(def *db.SkillDef, target uint32) (*ConnectedPlayer, bool)` | userSkillTarget resolves a TypeEffect 6 target: the caster itself for self skills (TargetType 2), else a user on the caster's map. |
| [`inUserSkillRange`](../../server/internal/world/skillonce.go#L132)<br>`func (s *Session) inUserSkillRange(p *ConnectedPlayer, def *db.SkillDef) bool` | inUserSkillRange checks launch reach and visibility without release grace. |
| [`skillOnceSucceeds`](../../server/internal/world/skillonce.go#L143)<br>`func (s *Session) skillOnceSucceeds(def *db.SkillDef, self bool) bool` | skillOnceSucceeds is CUser::AttackSuccess 0x00456b10 for these skills: a block (ChkBlock; checked when the caster is the target) fails it, then SuccessType 1 rolls SuccessValue%. Types 2 and 3 add 100 x caster/(caster+target) LUC or WIS in integer division, which is 0 unless the target's stat is 0, so they roll SuccessValue% too. The level-based default and the state resistances are not modelled: every EP4.5 TypeEffect 6 skill is SuccessType 1. |
| [`skillOnce`](../../server/internal/world/skillonce.go#L156)<br>`func (s *Session) skillOnce(k db.CharacterSkill, def *db.SkillDef, target uint32, stats CharacterStats, cast *pendingCast)` | skillOnce runs a TypeEffect 6 skill after its cast: SP/MP and the reset time are spent whatever the outcome (AttackUser 0x00451d1d). |
| [`tickSkillOnce`](../../server/internal/world/skillonce.go#L233)<br>`func (s *Session) tickSkillOnce()` | tickSkillOnce applies queued skill effects on this (dead) user. |
| [`lifeSkill`](../../server/internal/world/skillonce.go#L258)<br>`func (s *Session) lifeSkill(l *pendingLifeSkill)` | lifeSkill is CUser::LifeSkill 0x00459d20: the corpse leaves the view, 0x0551 kind 4 with the user's exp and the caster's position goes to the view, the pools are full, no exp is lost. |

<details><summary>Type definitions and owned fields</summary>

```go
type pendingLifeSkill struct {
	CasterID uint32
	X, Y, Z  float32
}
```

```go
type corpseTarget struct {
	player *ConnectedPlayer
	state  uint64
}
```

```go
type pendingSoulKill struct {
	source *ConnectedPlayer
	state  uint64
	mapID  uint16
	skill  db.CharacterSkill
}
```

</details>

### skills.go

[server/internal/world/skills.go](../../server/internal/world/skills.go#L1)

Direct GX dependencies: `internal/proto`, `internal/skilltiming`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`skillBarCount`](../../server/internal/world/skills.go#L15)<br>`const skillBarCount = 5` | See the source definition and the subsystem ownership contract. |
| [`skillBarSlotCount`](../../server/internal/world/skills.go#L16)<br>`const skillBarSlotCount = 10` | See the source definition and the subsystem ownership contract. |
| [`quickSlotSkillBag`](../../server/internal/world/skills.go#L17)<br>`const quickSlotSkillBag = 100` | See the source definition and the subsystem ownership contract. |
| [`maxQuickSlots`](../../server/internal/world/skills.go#L18)<br>`const maxQuickSlots = 128` | See the source definition and the subsystem ownership contract. |
| [`mountSlot`](../../server/internal/world/skills.go#L20)<br>`const mountSlot = 13` | See the source definition and the subsystem ownership contract. |
| [`mountItemType`](../../server/internal/world/skills.go#L21)<br>`const mountItemType = 42` | See the source definition and the subsystem ownership contract. |
| [`unmountedShape`](../../server/internal/world/skills.go#L22)<br>`const unmountedShape = 0x00` | See the source definition and the subsystem ownership contract. |
| [`buildCharacterSkills`](../../server/internal/world/skills.go#L25)<br>`func buildCharacterSkills(skills []db.CharacterSkill) []byte` | See the source definition and the subsystem ownership contract. |
| [`isValidQuickSlot`](../../server/internal/world/skills.go#L38)<br>`func isValidQuickSlot(q db.QuickSlot) bool` | See the source definition and the subsystem ownership contract. |
| [`hasItemShortcut`](../../server/internal/world/skills.go#L43)<br>`func (s *Session) hasItemShortcut(bag, slot uint8) bool` | See the source definition and the subsystem ownership contract. |
| [`clearConsumedShortcuts`](../../server/internal/world/skills.go#L49)<br>`func (s *Session) clearConsumedShortcuts(slots [][2]uint8)` | See the source definition and the subsystem ownership contract. |
| [`discardEmptyItemShortcuts`](../../server/internal/world/skills.go#L59)<br>`func (s *Session) discardEmptyItemShortcuts(slots []db.QuickSlot) []db.QuickSlot` | See the source definition and the subsystem ownership contract. |
| [`buildCharacterSkillBar`](../../server/internal/world/skills.go#L65)<br>`func buildCharacterSkillBar(slots []db.QuickSlot) []byte` | See the source definition and the subsystem ownership contract. |
| [`parseSkillBar`](../../server/internal/world/skills.go#L84)<br>`func parseSkillBar(p []byte) []db.QuickSlot` | parseSkillBar (0x010B): count, then bar, slot, bag, number:u16. |
| [`handleSkillBar`](../../server/internal/world/skills.go#L104)<br>`func (s *Session) handleSkillBar(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`buildLearnSkillError`](../../server/internal/world/skills.go#L130)<br>`func buildLearnSkillError(reason uint8, skillID uint16, level uint8) []byte` | buildLearnSkillError (0x0209): reason, skill id, level. |
| [`skillCountryAllows`](../../server/internal/world/skills.go#L141)<br>`func skillCountryAllows(country, family uint8) bool` | skillCountryAllows is EnableLearnSkill's Country byte by family (0 human, 1 elf, 2 vail, 3 deatheater). |
| [`buildLearnSkillResult`](../../server/internal/world/skills.go#L159)<br>`func buildLearnSkillResult(ok bool, number uint8, skillID uint16, level uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`handleLearnSkill`](../../server/internal/world/skills.go#L169)<br>`func (s *Session) handleLearnSkill(p []byte)` | handleLearnSkill (0x0209): skill id:u16, level. |
| [`handleSetStatPoint`](../../server/internal/world/skills.go#L265)<br>`func (s *Session) handleSetStatPoint(p []byte)` | handleSetStatPoint (0x0208): six u16 deltas str, dex, rec, int, wis, luc. |

### social.go

[server/internal/world/social.go](../../server/internal/world/social.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`socialState`](../../server/internal/world/social.go#L13)<br>`var socialState` | Native EP5 0x2201..0x220e, writers 0x59bb30..0x59bd80. |
| [`socialInvite`](../../server/internal/world/social.go#L19)<br>`type socialInvite` | See the source definition and the subsystem ownership contract. |
| [`buildSocialList`](../../server/internal/world/social.go#L24)<br>`func buildSocialList(rows []db.SocialContact, blocked bool) []byte` | See the source definition and the subsystem ownership contract. |
| [`truncateSocialMemo`](../../server/internal/world/social.go#L48)<br>`func truncateSocialMemo(v string) string` | See the source definition and the subsystem ownership contract. |
| [`sendSocial`](../../server/internal/world/social.go#L55)<br>`func (s *Session) sendSocial()` | See the source definition and the subsystem ownership contract. |
| [`tickSocial`](../../server/internal/world/social.go#L77)<br>`func (s *Session) tickSocial()` | See the source definition and the subsystem ownership contract. |
| [`handleSocial`](../../server/internal/world/social.go#L90)<br>`func (s *Session) handleSocial(op packet.Opcode, p []byte)` | See the source definition and the subsystem ownership contract. |
| [`sendPartySearch`](../../server/internal/world/social.go#L232)<br>`func (s *Session) sendPartySearch()` | See the source definition and the subsystem ownership contract. |
| [`clearSocialSession`](../../server/internal/world/social.go#L262)<br>`func clearSocialSession(id uint32)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type socialInvite struct {
	from  uint32
	until time.Time
}
```

</details>

### status_items.go

[server/internal/world/status_items.go](../../server/internal/world/status_items.go#L1)

Direct GX dependencies: `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`protectFaint`](../../server/internal/world/status_items.go#L10)<br>`const protectFaint = 1 << 7` | See the source definition and the subsystem ownership contract. |
| [`protectDeath`](../../server/internal/world/status_items.go#L11)<br>`const protectDeath = 1 << 8` | See the source definition and the subsystem ownership contract. |
| [`statusItemRule`](../../server/internal/world/status_items.go#L16)<br>`func statusItemRule(special uint8) (states, protection uint16, allDebuffs bool)` | EP4.5 CUser::ItemUseRemState (0x4653e0) selects the original StateType, with Perfect Cure matching every debuff. |
| [`useStatusItem`](../../server/internal/world/status_items.go#L45)<br>`func (s *Session) useStatusItem(index int, def *db.ItemDef) bool` | See the source definition and the subsystem ownership contract. |
| [`blocksStatusSkill`](../../server/internal/world/status_items.go#L86)<br>`func (s *Session) blocksStatusSkill(def *db.SkillDef) bool` | The original player and mob hit checks consume the matching flag before rolling success (0x456b39 and 0x4a2419). |

### summon_rune.go

[server/internal/world/summon_rune.go](../../server/internal/world/summon_rune.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`specialPartySummon`](../../server/internal/world/summon_rune.go#L14)<br>`const specialPartySummon = 31` | See the source definition and the subsystem ownership contract. |
| [`summonCastTime`](../../server/internal/world/summon_rune.go#L16)<br>`const summonCastTime = 5 * time.Second` | EP4.5 Special 31 at 0x465d1a adds 0x1388 ms at 0x465d9f (move type 4). |
| [`summonOfferTime`](../../server/internal/world/summon_rune.go#L17)<br>`const summonOfferTime = time.Minute` | See the source definition and the subsystem ownership contract. |
| [`partySummonOffer`](../../server/internal/world/summon_rune.go#L20)<br>`type partySummonOffer` | See the source definition and the subsystem ownership contract. |
| [`summonOfferSerial`](../../server/internal/world/summon_rune.go#L27)<br>`var summonOfferSerial atomic.Uint32` | See the source definition and the subsystem ownership contract. |
| [`partySummonTargets`](../../server/internal/world/summon_rune.go#L31)<br>`func partySummonTargets(caster *ConnectedPlayer) (*party, []*ConnectedPlayer)` | partySummonTargets returns the whole raid for its leader and the caller's own raid group for every other raid member. |
| [`summonResult`](../../server/internal/world/summon_rune.go#L58)<br>`func (s *Session) summonResult(message string)` | See the source definition and the subsystem ownership contract. |
| [`publishSummonCast`](../../server/internal/world/summon_rune.go#L65)<br>`func (s *Session) publishSummonCast(active bool)` | See the source definition and the subsystem ownership contract. |
| [`beginPartySummon`](../../server/internal/world/summon_rune.go#L81)<br>`func (s *Session) beginPartySummon(i int, def *db.ItemDef)` | See the source definition and the subsystem ownership contract. |
| [`cancelPartySummon`](../../server/internal/world/summon_rune.go#L123)<br>`func (s *Session) cancelPartySummon()` | See the source definition and the subsystem ownership contract. |
| [`tickPartySummon`](../../server/internal/world/summon_rune.go#L134)<br>`func (s *Session) tickPartySummon(now time.Time)` | See the source definition and the subsystem ownership contract. |
| [`handleSummonRune`](../../server/internal/world/summon_rune.go#L178)<br>`func (s *Session) handleSummonRune(p []byte)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type partySummonOffer struct {
	id      uint32
	caster  *ConnectedPlayer
	group   *party
	expires time.Time
}
```

</details>

### svmap.go

[server/internal/world/svmap.go](../../server/internal/world/svmap.go#L1)

Direct GX dependencies: `internal/mapnav`, `server/internal/config`, `server/internal/db`, `server/internal/logx`.

| Declaration | Responsibility / source comment |
|---|---|
| [`svmapReader`](../../server/internal/world/svmap.go#L22)<br>`type svmapReader` | See the source definition and the subsystem ownership contract. |
| [`errShortSvmap`](../../server/internal/world/svmap.go#L27)<br>`var errShortSvmap` | See the source definition and the subsystem ownership contract. |
| [`take`](../../server/internal/world/svmap.go#L29)<br>`func (r *svmapReader) take(n int) ([]byte, error)` | See the source definition and the subsystem ownership contract. |
| [`i32`](../../server/internal/world/svmap.go#L38)<br>`func (r *svmapReader) i32() (int32, error)` | See the source definition and the subsystem ownership contract. |
| [`u32`](../../server/internal/world/svmap.go#L46)<br>`func (r *svmapReader) u32() (uint32, error)` | See the source definition and the subsystem ownership contract. |
| [`u16`](../../server/internal/world/svmap.go#L51)<br>`func (r *svmapReader) u16() (uint16, error)` | See the source definition and the subsystem ownership contract. |
| [`f32`](../../server/internal/world/svmap.go#L59)<br>`func (r *svmapReader) f32() (float32, error)` | See the source definition and the subsystem ownership contract. |
| [`vec3`](../../server/internal/world/svmap.go#L64)<br>`type vec3` | See the source definition and the subsystem ownership contract. |
| [`vec3`](../../server/internal/world/svmap.go#L66)<br>`func (r *svmapReader) vec3() (vec3, error)` | See the source definition and the subsystem ownership contract. |
| [`count`](../../server/internal/world/svmap.go#L79)<br>`func (r *svmapReader) count(label string) (int, error)` | See the source definition and the subsystem ownership contract. |
| [`svmapLoad`](../../server/internal/world/svmap.go#L90)<br>`type svmapLoad` | See the source definition and the subsystem ownership contract. |
| [`mapDataDir`](../../server/internal/world/svmap.go#L97)<br>`func mapDataDir() string` | See the source definition and the subsystem ownership contract. |
| [`ensureMapDataDir`](../../server/internal/world/svmap.go#L99)<br>`func ensureMapDataDir() error` | See the source definition and the subsystem ownership contract. |
| [`LoadSvmaps`](../../server/internal/world/svmap.go#L112)<br>`func LoadSvmaps(log *logx.Logger) error` | LoadSvmaps loads every data/maps/<id>.svmap into the world. |
| [`loadSvmapFile`](../../server/internal/world/svmap.go#L181)<br>`func loadSvmapFile(mapID uint16, path string, defs *db.Defs) (svmapLoad, error)` | See the source definition and the subsystem ownership contract. |
| [`readMonsterAreas`](../../server/internal/world/svmap.go#L234)<br>`func readMonsterAreas(r *svmapReader, mapSize uint32, mask []byte, defs *db.Defs) ([]MapMob, error)` | See the source definition and the subsystem ownership contract. |
| [`positionInArea`](../../server/internal/world/svmap.go#L290)<br>`func positionInArea(lower, upper vec3, index, mapSize uint32, mask []byte) vec3` | See the source definition and the subsystem ownership contract. |
| [`areaCandidate`](../../server/internal/world/svmap.go#L300)<br>`func areaCandidate(lower, upper vec3, index uint32) vec3` | See the source definition and the subsystem ownership contract. |
| [`readNpcs`](../../server/internal/world/svmap.go#L306)<br>`func readNpcs(r *svmapReader) ([]MapNpc, error)` | See the source definition and the subsystem ownership contract. |
| [`yawToAngle`](../../server/internal/world/svmap.go#L350)<br>`func yawToAngle(yaw float32) uint16` | yawToAngle is the client's conversion: degrees = -yaw*180/pi in [0, 360). |
| [`readPortals`](../../server/internal/world/svmap.go#L361)<br>`func readPortals(r *svmapReader) ([]MapPortal, error)` | See the source definition and the subsystem ownership contract. |
| [`readSpawnAreas`](../../server/internal/world/svmap.go#L404)<br>`func readSpawnAreas(r *svmapReader) ([]MapSpawnArea, error)` | readSpawnAreas reads the rebirth groups as CMap::LoadMap does: per group count, type, an unused word, then count boxes (lower, upper). |

<details><summary>Type definitions and owned fields</summary>

```go
type svmapReader struct {
	data []byte
	pos  int
}
```

```go
type vec3 struct{ x, y, z float32 }
```

```go
type svmapLoad struct {
	mapSize  uint32
	cellSize uint32
	mask     []byte
	entities MapEntities
}
```

</details>

### sweeps.go

[server/internal/world/sweeps.go](../../server/internal/world/sweeps.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/logx`, `server/internal/scheduler`, `server/internal/session`.

| Declaration | Responsibility / source comment |
|---|---|
| [`Runtime`](../../server/internal/world/sweeps.go#L16)<br>`type Runtime` | StartSweeps launches the background world ticks: map item despawn, mob roam/respawn/combat, NPC patrol and mob DoTs. |
| [`Close`](../../server/internal/world/sweeps.go#L25)<br>`func (r *Runtime) Close()` | Close stops producers before storage closes and waits for the current job. |
| [`StartSweeps`](../../server/internal/world/sweeps.go#L35)<br>`func StartSweeps(ctx context.Context, log *logx.Logger, onFatal ...context.CancelFunc) *Runtime` | See the source definition and the subsystem ownership contract. |
| [`snapshotCombatPlayers`](../../server/internal/world/sweeps.go#L150)<br>`func snapshotCombatPlayers(players []*ConnectedPlayer) []PlayerCombatSnapshot` | See the source definition and the subsystem ownership contract. |
| [`queueMobAttack`](../../server/internal/world/sweeps.go#L164)<br>`func queueMobAttack(a MobAttackAction)` | queueMobAttack hands a mob's attack to its target, or for an area mob skill (TargetType 6 around the mob, 7 around the target) to every living player within ApplyRange. |
| [`viewersByMap`](../../server/internal/world/sweeps.go#L204)<br>`func viewersByMap() map[uint16][][2]float32` | See the source definition and the subsystem ownership contract. |
| [`dispatchMoves`](../../server/internal/world/sweeps.go#L213)<br>`func dispatchMoves(list []entityMove)` | See the source definition and the subsystem ownership contract. |
| [`runDotSweep`](../../server/internal/world/sweeps.go#L219)<br>`func runDotSweep(log *logx.Logger)` | See the source definition and the subsystem ownership contract. |
| [`storageFailure`](../../server/internal/world/sweeps.go#L241)<br>`func storageFailure(err error)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type Runtime struct {
	state  *realmState
	fatal  context.CancelFunc
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}
```

</details>

### target_effects.go

[server/internal/world/target_effects.go](../../server/internal/world/target_effects.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`publishPlayerEffects`](../../server/internal/world/target_effects.go#L11)<br>`func (s *Session) publishPlayerEffects()` | See the source definition and the subsystem ownership contract. |
| [`handleTargetEffects`](../../server/internal/world/target_effects.go#L19)<br>`func (s *Session) handleTargetEffects(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`tickTargetEffects`](../../server/internal/world/target_effects.go#L34)<br>`func (s *Session) tickTargetEffects(now time.Time)` | See the source definition and the subsystem ownership contract. |
| [`sendTargetHealth`](../../server/internal/world/target_effects.go#L90)<br>`func (s *Session) sendTargetHealth(p *ConnectedPlayer)` | sendTargetHealth gives the selected user's HP, once and again whenever it changes. |

### taunt.go

[server/internal/world/taunt.go](../../server/internal/world/taunt.go#L1)

Direct GX dependencies: `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`detailTaunt`](../../server/internal/world/taunt.go#L9)<br>`const detailTaunt = 69` | buff: pure aggro on the mobs around the caster |
| [`hitAggro`](../../server/internal/world/taunt.go#L14)<br>`func hitAggro(stats *CharacterStats, arrow uint16) int64` | hitAggro is the aggro one hit adds to a mob's ledger (CMob::Damage): the attacker's REC times ten plus the skill's Arrow, whatever the damage dealt. Normal attacks have no skill, so only the REC term. |
| [`tauntMobs`](../../server/internal/world/taunt.go#L23)<br>`func (s *Session) tauntMobs(def *db.SkillDef, stats *CharacterStats)` | tauntMobs is CUser::TauntMob (AddApplySkillBuff, TypeDetail 69): every living hostile mob within ApplyRange of the caster turns to the caster at once and its ledger gains the Arrow plus ten times the caster's REC, with no damage and no other effect. The mob's next retarget may still prefer whoever holds more aggro. |

### teleport_rune.go

[server/internal/world/teleport_rune.go](../../server/internal/world/teleport_rune.go#L1)

Direct GX dependencies: `internal/catalogfiles`, `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`specialTeleportRune`](../../server/internal/world/teleport_rune.go#L10)<br>`const specialTeleportRune = 47` | See the source definition and the subsystem ownership contract. |
| [`guildHouseCreateType`](../../server/internal/world/teleport_rune.go#L11)<br>`const guildHouseCreateType = "H"` | See the source definition and the subsystem ownership contract. |
| [`teleportRunes`](../../server/internal/world/teleport_rune.go#L13)<br>`var teleportRunes map[int16][2]db.TeleportRuneDestination` | See the source definition and the subsystem ownership contract. |
| [`teleportRuneCast`](../../server/internal/world/teleport_rune.go#L15)<br>`type teleportRuneCast` | See the source definition and the subsystem ownership contract. |
| [`loadTeleportRunes`](../../server/internal/world/teleport_rune.go#L21)<br>`func loadTeleportRunes() error` | See the source definition and the subsystem ownership contract. |
| [`canCastTeleportRune`](../../server/internal/world/teleport_rune.go#L38)<br>`func (s *Session) canCastTeleportRune() bool` | See the source definition and the subsystem ownership contract. |
| [`activeGuildHouse`](../../server/internal/world/teleport_rune.go#L50)<br>`func (s *Session) activeGuildHouse() uint32` | See the source definition and the subsystem ownership contract. |
| [`resolveTeleportRune`](../../server/internal/world/teleport_rune.go#L64)<br>`func (s *Session) resolveTeleportRune(index int16) (uint16, [3]float32, uint32, string)` | A configured house destination automatically requires the user's own operational house, regardless of which item references the destination. |
| [`refuseTeleportRune`](../../server/internal/world/teleport_rune.go#L96)<br>`func (s *Session) refuseTeleportRune(message string)` | See the source definition and the subsystem ownership contract. |
| [`beginTeleportRune`](../../server/internal/world/teleport_rune.go#L103)<br>`func (s *Session) beginTeleportRune(i int, def *db.ItemDef)` | See the source definition and the subsystem ownership contract. |
| [`cancelTeleportRune`](../../server/internal/world/teleport_rune.go#L123)<br>`func (s *Session) cancelTeleportRune()` | See the source definition and the subsystem ownership contract. |
| [`tickTeleportRune`](../../server/internal/world/teleport_rune.go#L134)<br>`func (s *Session) tickTeleportRune(now time.Time)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type teleportRuneCast struct {
	item  db.Item
	index int16
	ends  time.Time
}
```

</details>

### teleport_stone.go

[server/internal/world/teleport_stone.go](../../server/internal/world/teleport_stone.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`specialTeleportStone`](../../server/internal/world/teleport_stone.go#L12)<br>`const specialTeleportStone = 39` | See the source definition and the subsystem ownership contract. |
| [`teleportStoneCastDuration`](../../server/internal/world/teleport_stone.go#L16)<br>`const teleportStoneCastDuration = 5 * time.Second` | Special 39 branches to 0x466097 in EP4.5 ps_game.exe. At 0x4661f3 it adds 0x1388 milliseconds before scheduling a type-2 move. |
| [`teleportStoneCast`](../../server/internal/world/teleport_stone.go#L18)<br>`type teleportStoneCast` | See the source definition and the subsystem ownership contract. |
| [`cancelTravelCast`](../../server/internal/world/teleport_stone.go#L24)<br>`func (s *Session) cancelTravelCast()` | See the source definition and the subsystem ownership contract. |
| [`cancelRecordedPointCast`](../../server/internal/world/teleport_stone.go#L31)<br>`func (s *Session) cancelRecordedPointCast()` | See the source definition and the subsystem ownership contract. |
| [`tickTeleportStone`](../../server/internal/world/teleport_stone.go#L40)<br>`func (s *Session) tickTeleportStone(now time.Time)` | See the source definition and the subsystem ownership contract. |
| [`sendTeleportStone`](../../server/internal/world/teleport_stone.go#L51)<br>`func (s *Session) sendTeleportStone(action uint8, ok bool, message string, points [proto.TeleportPointCount]db.TeleportPoint)` | See the source definition and the subsystem ownership contract. |
| [`openTeleportStone`](../../server/internal/world/teleport_stone.go#L69)<br>`func (s *Session) openTeleportStone(it db.Item)` | See the source definition and the subsystem ownership contract. |
| [`teleportStoneField`](../../server/internal/world/teleport_stone.go#L85)<br>`func teleportStoneField(mapID uint16) bool` | See the source definition and the subsystem ownership contract. |
| [`teleportStoneMapAllowed`](../../server/internal/world/teleport_stone.go#L90)<br>`func (s *Session) teleportStoneMapAllowed(target uint16) bool` | See the source definition and the subsystem ownership contract. |
| [`teleportPortalAllowed`](../../server/internal/world/teleport_stone.go#L97)<br>`func (s *Session) teleportPortalAllowed(target uint16) bool` | See the source definition and the subsystem ownership contract. |
| [`handleTeleportStone`](../../server/internal/world/teleport_stone.go#L122)<br>`func (s *Session) handleTeleportStone(p []byte)` | See the source definition and the subsystem ownership contract. |
| [`resolveTeleportStone`](../../server/internal/world/teleport_stone.go#L137)<br>`func (s *Session) resolveTeleportStone(action, slot uint8, complete bool)` | See the source definition and the subsystem ownership contract. |
| [`commitTeleportItem`](../../server/internal/world/teleport_stone.go#L210)<br>`func (s *Session) commitTeleportItem(i int, mapID uint16, dest [3]float32) (db.Item, error)` | commitTeleportItem saves the destination and consumption in one transaction. The caller has already resolved instance ownership and physical ground. |
| [`teleportStoneDueling`](../../server/internal/world/teleport_stone.go#L233)<br>`func (s *Session) teleportStoneDueling() bool` | A recorded point cannot be used to leave a wager or an active match. |

<details><summary>Type definitions and owned fields</summary>

```go
type teleportStoneCast struct {
	slot   uint8
	ends   time.Time
	points [proto.TeleportPointCount]db.TeleportPoint
}
```

</details>

### trade.go

[server/internal/world/trade.go](../../server/internal/world/trade.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/logx`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`tradeSlots`](../../server/internal/world/trade.go#L15)<br>`const tradeSlots = 8` | See the source definition and the subsystem ownership contract. |
| [`offeredItem`](../../server/internal/world/trade.go#L17)<br>`type offeredItem` | See the source definition and the subsystem ownership contract. |
| [`tradeOffer`](../../server/internal/world/trade.go#L22)<br>`type tradeOffer` | See the source definition and the subsystem ownership contract. |
| [`tradeSession`](../../server/internal/world/trade.go#L29)<br>`type tradeSession` | See the source definition and the subsystem ownership contract. |
| [`offer`](../../server/internal/world/trade.go#L36)<br>`func (t *tradeSession) offer(id uint32) *tradeOffer` | See the source definition and the subsystem ownership contract. |
| [`partner`](../../server/internal/world/trade.go#L46)<br>`func (t *tradeSession) partner(id uint32) uint32` | See the source definition and the subsystem ownership contract. |
| [`unready`](../../server/internal/world/trade.go#L53)<br>`func (t *tradeSession) unready()` | See the source definition and the subsystem ownership contract. |
| [`tradeMu`](../../server/internal/world/trade.go#L58)<br>`var tradeMu sync.Mutex` | See the source definition and the subsystem ownership contract. |
| [`tradeRequests`](../../server/internal/world/trade.go#L59)<br>`var tradeRequests` | requester -> target |
| [`tradeSessions`](../../server/internal/world/trade.go#L60)<br>`var tradeSessions` | See the source definition and the subsystem ownership contract. |
| [`tradeFor`](../../server/internal/world/trade.go#L63)<br>`func tradeFor(id uint32) *tradeSession` | See the source definition and the subsystem ownership contract. |
| [`endTrade`](../../server/internal/world/trade.go#L69)<br>`func endTrade(a, b uint32)` | See the source definition and the subsystem ownership contract. |
| [`tradePacket`](../../server/internal/world/trade.go#L76)<br>`func tradePacket(op packet.Opcode, b ...byte) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildTradeReady`](../../server/internal/world/trade.go#L82)<br>`func buildTradeReady(kind uint8, canceled bool) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildTradeConfirm`](../../server/internal/world/trade.go#L86)<br>`func buildTradeConfirm(kind uint8, confirmed bool) []byte` | See the source definition and the subsystem ownership contract. |
| [`boolByte`](../../server/internal/world/trade.go#L90)<br>`func boolByte(v bool) byte` | See the source definition and the subsystem ownership contract. |
| [`sendTradeCancel`](../../server/internal/world/trade.go#L97)<br>`func sendTradeCancel(p *ConnectedPlayer)` | See the source definition and the subsystem ownership contract. |
| [`broadcastUnready`](../../server/internal/world/trade.go#L104)<br>`func broadcastUnready(players ...*ConnectedPlayer)` | See the source definition and the subsystem ownership contract. |
| [`tradeCancelForDisconnect`](../../server/internal/world/trade.go#L116)<br>`func tradeCancelForDisconnect(id uint32, log *logx.Logger)` | tradeCancelForDisconnect drops pending requests and cancels any trade. |
| [`buildTradeStart`](../../server/internal/world/trade.go#L136)<br>`func buildTradeStart(partner uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`startTrade`](../../server/internal/world/trade.go#L143)<br>`func startTrade(a, b *ConnectedPlayer)` | See the source definition and the subsystem ownership contract. |
| [`handleTradeRequest`](../../server/internal/world/trade.go#L157)<br>`func (s *Session) handleTradeRequest(target uint32)` | handleTradeRequest (0x0A01): target character id. Two crossing requests start the trade directly. |
| [`handleTradeAnswer`](../../server/internal/world/trade.go#L187)<br>`func (s *Session) handleTradeAnswer(accepted bool)` | handleTradeAnswer (0x0A02): the byte is "declined" (0 accepts). |
| [`writeOffered`](../../server/internal/world/trade.go#L215)<br>`func writeOffered(w *packet.Writer, o *offeredItem)` | See the source definition and the subsystem ownership contract. |
| [`buildTradeAddDest`](../../server/internal/world/trade.go#L232)<br>`func buildTradeAddDest(slot uint8, o *offeredItem) []byte` | See the source definition and the subsystem ownership contract. |
| [`mutateOffer`](../../server/internal/world/trade.go#L241)<br>`func (s *Session) mutateOffer(fn func(*tradeOffer)) (*tradeSession, *ConnectedPlayer, bool)` | mutateOffer applies fn to this player's offer, unreadies both sides and returns the partner (nil if offline). |
| [`handleTradeAdd`](../../server/internal/world/trade.go#L261)<br>`func (s *Session) handleTradeAdd(bag, slot uint8, count uint16, dest uint8)` | handleTradeAdd (0x0A06): bag, slot, count, trade slot. |
| [`handleTradeRemove`](../../server/internal/world/trade.go#L301)<br>`func (s *Session) handleTradeRemove(dest uint8)` | See the source definition and the subsystem ownership contract. |
| [`buildTradeMoney`](../../server/internal/world/trade.go#L315)<br>`func buildTradeMoney(kind uint8, money uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`handleTradeMoney`](../../server/internal/world/trade.go#L322)<br>`func (s *Session) handleTradeMoney(requested uint32)` | See the source definition and the subsystem ownership contract. |
| [`handleTradeReady`](../../server/internal/world/trade.go#L338)<br>`func (s *Session) handleTradeReady(v uint8)` | handleTradeReady (0x0A05): 0 ready, 1 unready, 2+ cancel. |
| [`handleTradeConfirm`](../../server/internal/world/trade.go#L371)<br>`func (s *Session) handleTradeConfirm(confirmed bool)` | handleTradeConfirm (0x0A0A): confirmed flag. |
| [`offerValid`](../../server/internal/world/trade.go#L394)<br>`func offerValid(items []db.Item, o *tradeOffer, money uint32) bool` | See the source definition and the subsystem ownership contract. |
| [`offerFits`](../../server/internal/world/trade.go#L413)<br>`func offerFits(receiver []db.Item, o *tradeOffer) bool` | See the source definition and the subsystem ownership contract. |
| [`planTransferOffer`](../../server/internal/world/trade.go#L445)<br>`func planTransferOffer(o *tradeOffer, from, to *[]db.Item) (fromPk, toPk [][]byte, err error)` | planTransferOffer updates private inventories and builds their result packets. |
| [`queueTradeCommit`](../../server/internal/world/trade.go#L471)<br>`func queueTradeCommit(trade *tradeSession)` | One coordinator retries on its next tick if its partner is busy. Neither actor waits for the other's state lock while holding its own. |
| [`finalizeTrade`](../../server/internal/world/trade.go#L482)<br>`func (s *Session) finalizeTrade(trade *tradeSession, partner *ConnectedPlayer)` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type offeredItem struct {
	Item  db.Item
	Count uint16
}
```

```go
type tradeOffer struct {
	items     [tradeSlots]*offeredItem
	money     uint32
	ready     bool
	confirmed bool
}
```

```go
type tradeSession struct {
	mu     sync.Mutex
	a, b   uint32
	offerA tradeOffer
	offerB tradeOffer
}
```

</details>

### trap.go

[server/internal/world/trap.go](../../server/internal/world/trap.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`typeEffectField`](../../server/internal/world/trap.go#L13)<br>`const typeEffectField = 5` | CUser::SkillField: a trap or field on the ground |
| [`detailTrap`](../../server/internal/world/trap.go#L14)<br>`const detailTrap = 101` | springs once on the first enemy, then it is gone |
| [`detailFireBush`](../../server/internal/world/trap.go#L15)<br>`const detailFireBush = 102` | a curtain of fire: fixed damage to every enemy in it, every half second |
| [`fireBushTick`](../../server/internal/world/trap.go#L16)<br>`const fireBushTick = 500 * time.Millisecond` | See the source definition and the subsystem ownership contract. |
| [`trapSweep`](../../server/internal/world/trap.go#L17)<br>`const trapSweep = 250 * time.Millisecond` | See the source definition and the subsystem ownership contract. |
| [`trapDamageMax`](../../server/internal/world/trap.go#L18)<br>`const trapDamageMax = 30000` | See the source definition and the subsystem ownership contract. |
| [`placedTrap`](../../server/internal/world/trap.go#L23)<br>`type placedTrap` | placedTrap is a CTrab: owner, position, skill and the time it ends. It is invisible to everybody; only its cast and its hit are announced. |
| [`skillField`](../../server/internal/world/trap.go#L35)<br>`func (s *Session) skillField(k db.CharacterSkill, def *db.SkillDef, stats CharacterStats, completed ...bool)` | skillField is CUser::SkillField (AttackUser, TypeEffect 5): the trap is left on the ground where the caster stands, lasts the skill's KeepTime and is announced like any cast (0x0511, result 0, no target). There is no limit of traps: the skill's own reuse time is the only one. |
| [`trapDamage`](../../server/internal/world/trap.go#L69)<br>`func trapDamage(def *db.SkillDef, current, maximum [3]uint32) [3]uint32` | trapDamage is TrabDamage 0x00422480. DamageType 4 takes the value as a percentage of the target's maximum pool and 5 of its current pool; any other type is the value itself. Each non-zero part is at least 1 and at most 30000. |
| [`tickTraps`](../../server/internal/world/trap.go#L93)<br>`func (s *Session) tickTraps(now time.Time)` | tickTraps is CZone::UpdateTrab for this user's traps: an expired trap goes away, and the first enemy (user or mob) within ApplyRange of a trap springs it. A trap left in another map is dropped. |
| [`springTrap`](../../server/internal/world/trap.go#L121)<br>`func (s *Session) springTrap(t placedTrap) bool` | springTrap hits the nearest enemy in reach and reports whether it did. |
| [`trapHitsMob`](../../server/internal/world/trap.go#L165)<br>`func (s *Session) trapHitsMob(id uint32, def *db.SkillDef)` | trapHitsMob is the mob branch of UpdateTrab: TrabDamage on the mob's HP, CMob::Damage with the skill's Arrow as aggro, 0x0519 to the view, and the mob turns to the owner when it survives. |
| [`applyTrapHit`](../../server/internal/world/trap.go#L190)<br>`func (s *Session) applyTrapHit(h pendingUserHit) bool` | applyTrapHit is the victim side of a sprung trap: its stealth ends (RemApplySkillHide), TrabDamage on its pools, CUser::Damage (HP, SP and MP), 0x0513 to the view; true when this character died. |
| [`mobsNear`](../../server/internal/world/trap.go#L225)<br>`func (s *Session) mobsNear(x, z, r float32, height ...float32) []uint32` | mobsNear lists the living mobs whose centre is within r of a point, measured on the ground plane. The original area loops (UpdateTrab, TauntMob) use the distance in X and Z only, so a slope, a mob's height or a wall between them never narrows a fire's or a trap's reach.  An optional height (y, reach) also keeps only the mobs within reach of it vertically. |

<details><summary>Type definitions and owned fields</summary>

```go
type placedTrap struct {
	mapID   uint16
	x, y, z float32
	def     *db.SkillDef
	endsAt  time.Time
	nextHit time.Time // Fire Bush: the next pulse; zero for a trap
}
```

</details>

### utility_items.go

[server/internal/world/utility_items.go](../../server/internal/world/utility_items.go#L1)

Direct GX dependencies: `server/internal/db`.

| Declaration | Responsibility / source comment |
|---|---|
| [`specialSkillReset`](../../server/internal/world/utility_items.go#L11)<br>`const specialSkillReset = 33` | See the source definition and the subsystem ownership contract. |
| [`specialTargetResurrection`](../../server/internal/world/utility_items.go#L12)<br>`const specialTargetResurrection = 41` | See the source definition and the subsystem ownership contract. |
| [`targetResurrectionRange`](../../server/internal/world/utility_items.go#L15)<br>`const targetResurrectionRange = CombatRadius` | EP4.5 checks the same zone without an explicit distance in ItemUse/LifeSkill. GX caps use at the combat visibility distance. |
| [`resetSkillsWithItem`](../../server/internal/world/utility_items.go#L18)<br>`func (s *Session) resetSkillsWithItem(index int) bool` | See the source definition and the subsystem ownership contract. |
| [`resurrectWithItem`](../../server/internal/world/utility_items.go#L49)<br>`func (s *Session) resurrectWithItem(index int, id uint32) bool` | See the source definition and the subsystem ownership contract. |
| [`remainingSkillDelay`](../../server/internal/world/utility_items.go#L98)<br>`func (s *Session) remainingSkillDelay(skillID uint16) int16` | See the source definition and the subsystem ownership contract. |

### vehicle.go

[server/internal/world/vehicle.go](../../server/internal/world/vehicle.go#L1)

Direct GX dependencies: `internal/movespeed`, `internal/proto`, `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`vehicleBoardRange`](../../server/internal/world/vehicle.go#L13)<br>`const vehicleBoardRange = 10` | See the source definition and the subsystem ownership contract. |
| [`vehicleCastDuration`](../../server/internal/world/vehicle.go#L14)<br>`const vehicleCastDuration = 2 * time.Second` | See the source definition and the subsystem ownership contract. |
| [`vehicleMu`](../../server/internal/world/vehicle.go#L16)<br>`var vehicleMu sync.Mutex` | See the source definition and the subsystem ownership contract. |
| [`vehicleCooldowns`](../../server/internal/world/vehicle.go#L17)<br>`var vehicleCooldowns` | See the source definition and the subsystem ownership contract. |
| [`vehiclePresence`](../../server/internal/world/vehicle.go#L19)<br>`type vehiclePresence` | See the source definition and the subsystem ownership contract. |
| [`remainingVehicleMS`](../../server/internal/world/vehicle.go#L26)<br>`func remainingVehicleMS(end, now time.Time) uint32` | See the source definition and the subsystem ownership contract. |
| [`vehicleState`](../../server/internal/world/vehicle.go#L32)<br>`func (p *ConnectedPlayer) vehicleState() proto.VehicleState` | See the source definition and the subsystem ownership contract. |
| [`publishVehicle`](../../server/internal/world/vehicle.go#L41)<br>`func publishVehicle(p *ConnectedPlayer)` | See the source definition and the subsystem ownership contract. |
| [`clearVehicle`](../../server/internal/world/vehicle.go#L56)<br>`func clearVehicle(p *ConnectedPlayer)` | Only shared presence changes here; each session applies its own position. |
| [`mountDefinition`](../../server/internal/world/vehicle.go#L98)<br>`func (s *Session) mountDefinition() (*db.Item, *db.ItemDef)` | See the source definition and the subsystem ownership contract. |
| [`validMount`](../../server/internal/world/vehicle.go#L105)<br>`func validMount(d *db.ItemDef) bool` | See the source definition and the subsystem ownership contract. |
| [`vehicleResult`](../../server/internal/world/vehicle.go#L108)<br>`func (s *Session) vehicleResult(ok, on bool)` | See the source definition and the subsystem ownership contract. |
| [`mountForbiddenMap`](../../server/internal/world/vehicle.go#L117)<br>`func mountForbiddenMap(mapID uint16) bool` | mountForbiddenMap reports maps that do not allow mounts: dungeons (Map.ini MapType D), including their party and guild instances. |
| [`keepsMountAcrossTeleport`](../../server/internal/world/vehicle.go#L126)<br>`func (s *Session) keepsMountAcrossTeleport(mapID uint16, label string) bool` | keepsMountAcrossTeleport lets a lone rider keep the mount through every kind of travel (portals, gatekeepers, GM moves, town return, queued moves) when the destination allows mounts. Dungeons, a driver with a passenger and rebirth leave the character on foot. |
| [`handleUseVehicle`](../../server/internal/world/vehicle.go#L134)<br>`func (s *Session) handleUseVehicle()` | See the source definition and the subsystem ownership contract. |
| [`dismount`](../../server/internal/world/vehicle.go#L167)<br>`func (s *Session) dismount()` | See the source definition and the subsystem ownership contract. |
| [`cancelVehicleCast`](../../server/internal/world/vehicle.go#L174)<br>`func (s *Session) cancelVehicleCast()` | See the source definition and the subsystem ownership contract. |
| [`tickVehicle`](../../server/internal/world/vehicle.go#L180)<br>`func (s *Session) tickVehicle(now time.Time)` | See the source definition and the subsystem ownership contract. |
| [`mountSpeed`](../../server/internal/world/vehicle.go#L255)<br>`func (s *Session) mountSpeed() float32` | See the source definition and the subsystem ownership contract. |
| [`vehicleNear`](../../server/internal/world/vehicle.go#L264)<br>`func vehicleNear(a, b *ConnectedPlayer) bool` | See the source definition and the subsystem ownership contract. |
| [`vehicleReply`](../../server/internal/world/vehicle.go#L272)<br>`func vehicleReply(p *ConnectedPlayer, code uint8)` | See the source definition and the subsystem ownership contract. |
| [`handleVehicleInvite`](../../server/internal/world/vehicle.go#L279)<br>`func (s *Session) handleVehicleInvite(packetData []byte)` | See the source definition and the subsystem ownership contract. |
| [`handleVehicleAnswer`](../../server/internal/world/vehicle.go#L317)<br>`func (s *Session) handleVehicleAnswer(data []byte)` | See the source definition and the subsystem ownership contract. |
| [`vehicleBlocked`](../../server/internal/world/vehicle.go#L370)<br>`func (s *Session) vehicleBlocked() bool` | See the source definition and the subsystem ownership contract. |
| [`ridingVehicle`](../../server/internal/world/vehicle.go#L375)<br>`func (s *Session) ridingVehicle() bool` | See the source definition and the subsystem ownership contract. |
| [`buildVehicleState`](../../server/internal/world/vehicle.go#L379)<br>`func buildVehicleState(v proto.VehicleState) []byte` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type vehiclePresence struct {
	state                              proto.VehicleState
	castEnd                            time.Time
	inviteFrom, inviteTo, inviteDriver uint32
	inviteEnd                          time.Time
}
```

</details>

### view.go

[server/internal/world/view.go](../../server/internal/world/view.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`viewMargin`](../../server/internal/world/view.go#L24)<br>`const viewMargin = 10` | See the source definition and the subsystem ownership contract. |
| [`BroadcastRadius`](../../server/internal/world/view.go#L25)<br>`const BroadcastRadius = 60.0` | See the source definition and the subsystem ownership contract. |
| [`holds`](../../server/internal/world/view.go#L29)<br>`func (c cell) holds(x, z float32) bool` | holds reports whether (x, z) is inside the cell's view window. |
| [`cellAfterMove`](../../server/internal/world/view.go#L37)<br>`func cellAfterMove(prev cell, has bool, x, z float32) cell` | cellAfterMove keeps the previous cell while the position stays in its view window, otherwise the cell under the position. |
| [`viewKind`](../../server/internal/world/view.go#L44)<br>`type viewKind` | See the source definition and the subsystem ownership contract. |
| [`viewMob`](../../server/internal/world/view.go#L47)<br>`const viewMob = iota` | See the source definition and the subsystem ownership contract. |
| [`viewNpc`](../../server/internal/world/view.go#L48)<br>`const viewNpc` | See the source definition and the subsystem ownership contract. |
| [`viewItem`](../../server/internal/world/view.go#L49)<br>`const viewItem` | See the source definition and the subsystem ownership contract. |
| [`viewUser`](../../server/internal/world/view.go#L50)<br>`const viewUser` | See the source definition and the subsystem ownership contract. |
| [`playerView`](../../server/internal/world/view.go#L54)<br>`type playerView` | playerView is a user's view cell and the entities its client knows. |
| [`set`](../../server/internal/world/view.go#L61)<br>`func (v *playerView) set(kind viewKind) map[uint32]bool` | See the source definition and the subsystem ownership contract. |
| [`viewCell`](../../server/internal/world/view.go#L68)<br>`func (p *ConnectedPlayer) viewCell() (cell, bool)` | See the source definition and the subsystem ownership contract. |
| [`moveViewCell`](../../server/internal/world/view.go#L75)<br>`func (p *ConnectedPlayer) moveViewCell(x, z float32) bool` | moveViewCell applies the hysteresis; it returns whether the cell changed. |
| [`resetView`](../../server/internal/world/view.go#L86)<br>`func (p *ConnectedPlayer) resetView() [4]map[uint32]bool` | resetView forgets everything (map change, teleport, leaving the world) and returns what was known. |
| [`knows`](../../server/internal/world/view.go#L95)<br>`func (p *ConnectedPlayer) knows(kind viewKind, id uint32) bool` | See the source definition and the subsystem ownership contract. |
| [`know`](../../server/internal/world/view.go#L102)<br>`func (p *ConnectedPlayer) know(kind viewKind, id uint32) bool` | know marks an entity known; true when it was not. |
| [`forget`](../../server/internal/world/view.go#L114)<br>`func (p *ConnectedPlayer) forget(kind viewKind, id uint32) bool` | forget unmarks an entity; true when it was known. |
| [`knownIDs`](../../server/internal/world/view.go#L124)<br>`func (p *ConnectedPlayer) knownIDs(kind viewKind) []uint32` | See the source definition and the subsystem ownership contract. |
| [`viewing`](../../server/internal/world/view.go#L135)<br>`func (p *ConnectedPlayer) viewing(c cell) bool` | viewing reports whether the player's 5x5 view covers the cell. |
| [`PlayersViewing`](../../server/internal/world/view.go#L141)<br>`func PlayersViewing(mapID uint16, c cell) []*ConnectedPlayer` | PlayersViewing are the users whose view covers the cell. |
| [`sendView`](../../server/internal/world/view.go#L147)<br>`func sendView(mapID uint16, x, z float32, frame []byte, label string)` | sendView is CZone::SendView: the frame to the users viewing the cell of (x, z). |
| [`viewEnter`](../../server/internal/world/view.go#L154)<br>`func viewEnter(p *ConnectedPlayer, kind viewKind, id uint32, frame []byte, label string)` | viewEnter sends an entity's enter packet to a user that did not know it. |
| [`viewLeave`](../../server/internal/world/view.go#L161)<br>`func viewLeave(p *ConnectedPlayer, kind viewKind, id uint32, frame []byte, label string)` | viewLeave sends an entity's leave packet to a user that knew it. |
| [`entityMove`](../../server/internal/world/view.go#L169)<br>`type entityMove` | entityMove is one mob or NPC update for the users around it: a cell change (or an appearance with Fresh), then its move packet. |
| [`leaveFrame`](../../server/internal/world/view.go#L180)<br>`func leaveFrame(kind viewKind, id uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`dispatchEntityMove`](../../server/internal/world/view.go#L193)<br>`func dispatchEntityMove(e entityMove)` | dispatchEntityMove is SendMoveMob / SendMoveNpc followed by the move. |
| [`dispatchEntityLeave`](../../server/internal/world/view.go#L218)<br>`func dispatchEntityLeave(mapID uint16, kind viewKind, id uint32, label string)` | dispatchEntityLeave removes an entity from every user that knows it (a mob's corpse disappearing, an item picked up or expired). |
| [`dispatchItemEnter`](../../server/internal/world/view.go#L226)<br>`func dispatchItemEnter(mapID uint16, it *MapItem, label string)` | dispatchItemEnter shows a new map item to the users viewing its cell. |
| [`buildUserLeave`](../../server/internal/world/view.go#L234)<br>`func buildUserLeave(id uint32) []byte` | buildUserLeave (0x0202): the character id. |
| [`refreshView`](../../server/internal/world/view.go#L244)<br>`func (s *Session) refreshView()` | refreshView is CZone::SendMoveUser for the user's own cell change (and its first placement): the leave packets of what left the 5x5 view, the enter packets of what entered it (SendCellInfoLeave / SendCellInfoEnter), and the user's own enter or leave for the other users concerned. |
| [`refreshUserView`](../../server/internal/world/view.go#L249)<br>`func (s *Session) refreshUserView()` | refreshUserView re-evaluates only who can see whom: a visibility change never alters the monsters, NPCs and items in view, and walking all of them on every toggle is wasted work on crowded maps. |
| [`refreshViewFor`](../../server/internal/world/view.go#L251)<br>`func (s *Session) refreshViewFor(usersOnly bool)` | See the source definition and the subsystem ownership contract. |
| [`refreshUsers`](../../server/internal/world/view.go#L309)<br>`func (s *Session) refreshUsers(p *ConnectedPlayer, ch *db.Character, c cell)` | See the source definition and the subsystem ownership contract. |
| [`enterFrames`](../../server/internal/world/view.go#L345)<br>`func (p *ConnectedPlayer) enterFrames() [][]byte` | enterFrames is what CCell::SendCellInfoEnter shows of a user. |
| [`userEnterFrames`](../../server/internal/world/view.go#L375)<br>`func (s *Session) userEnterFrames() [][]byte` | See the source definition and the subsystem ownership contract. |
| [`clearView`](../../server/internal/world/view.go#L384)<br>`func (s *Session) clearView(sendLeaves bool)` | clearView forgets the user's view: every known entity gets its leave packet (sendLeaves) and the other users that saw it get 0x0202. |
| [`dropFromOthers`](../../server/internal/world/view.go#L407)<br>`func (p *ConnectedPlayer) dropFromOthers()` | dropFromOthers sends 0x0202 to every user that knows this one. |
| [`sendToViewers`](../../server/internal/world/view.go#L417)<br>`func (s *Session) sendToViewers(frame []byte, label string)` | sendToViewers is CZone::SendView for the user's own updates: the users that currently show it. |
| [`placeView`](../../server/internal/world/view.go#L437)<br>`func (s *Session) placeView()` | placeView puts the user at its current position and sends its view. |
| [`setMountShape`](../../server/internal/world/view.go#L445)<br>`func (p *ConnectedPlayer) setMountShape(shape uint8)` | See the source definition and the subsystem ownership contract. |
| [`mountShape`](../../server/internal/world/view.go#L451)<br>`func (p *ConnectedPlayer) mountShape() uint8` | See the source definition and the subsystem ownership contract. |

<details><summary>Type definitions and owned fields</summary>

```go
type viewKind uint8
```

```go
type playerView struct {
	mu    sync.Mutex
	has   bool
	cell  cell
	known [4]map[uint32]bool
}
```

```go
type entityMove struct {
	MapID    uint16
	Kind     viewKind
	ID       uint32
	From, To cell
	Fresh    bool // just appeared: no previous cell
	Changed  bool
	Move     []byte // 0x0603 / 0x0E03, nil for an appearance
	Enter    []byte // 0x0601 / 0x0E01 at the new position
}
```

</details>

### visibility.go

[server/internal/world/visibility.go](../../server/internal/world/visibility.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`gxVisibility`](../../server/internal/world/visibility.go#L12)<br>`const gxVisibility = 0x7F02` | GX visibility describes only entities the viewer is permitted to receive. Forbidden viewers receive a native leave and no subsequent movement/shape. |
| [`canSee`](../../server/internal/world/visibility.go#L14)<br>`func (p *ConnectedPlayer) canSee(subject *ConnectedPlayer) bool` | See the source definition and the subsystem ownership contract. |
| [`visibilityPacket`](../../server/internal/world/visibility.go#L27)<br>`func (p *ConnectedPlayer) visibilityPacket() []byte` | See the source definition and the subsystem ownership contract. |
| [`invulnerable`](../../server/internal/world/visibility.go#L38)<br>`func (s *Session) invulnerable() bool` | See the source definition and the subsystem ownership contract. |
| [`syncVisibility`](../../server/internal/world/visibility.go#L40)<br>`func (s *Session) syncVisibility()` | See the source definition and the subsystem ownership contract. |
| [`breakStealth`](../../server/internal/world/visibility.go#L78)<br>`func (s *Session) breakStealth()` | An accepted attack (including a miss), or actual damage, ends stealth. GM visibility and invulnerability are independent command modes. |
| [`canReceiveWorldPacket`](../../server/internal/world/visibility.go#L88)<br>`func (p *ConnectedPlayer) canReceiveWorldPacket(frame []byte) bool` | Central outbound guard also covers combat, speed, cast and effect broadcasts that use proximity delivery rather than the character view list. |
| [`redactPartyPositions`](../../server/internal/world/visibility.go#L124)<br>`func (p *ConnectedPlayer) redactPartyPositions(frame []byte) []byte` | Party membership remains visible, but a hidden member's coordinates are not disclosed to an unauthorized viewer by roster/join packets. |

### warehouse.go

[server/internal/world/warehouse.go](../../server/internal/world/warehouse.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`WarehouseBag`](../../server/internal/world/warehouse.go#L14)<br>`const WarehouseBag = 100` | The account warehouse (PS_GameData.dbo.UserStoredItems) is bag 100 in the client's move packets. |
| [`WarehouseSlots`](../../server/internal/world/warehouse.go#L15)<br>`const WarehouseSlots = 240` | The account warehouse (PS_GameData.dbo.UserStoredItems) is bag 100 in the client's move packets. CUser::IsBank: slots below 0xF0 |
| [`findStored`](../../server/internal/world/warehouse.go#L18)<br>`func findStored(items []db.Item, slot uint8) int` | See the source definition and the subsystem ownership contract. |
| [`buildWarehouseItemList`](../../server/internal/world/warehouse.go#L27)<br>`func buildWarehouseItemList(items []db.Item) []byte` | See the source definition and the subsystem ownership contract. |
| [`withMoney`](../../server/internal/world/warehouse.go#L46)<br>`func withMoney(frame []byte, money uint32) []byte` | withMoney appends the character's money to a withdraw 0x0204 and rewrites the frame length. |
| [`copyPtr`](../../server/internal/world/warehouse.go#L52)<br>`func copyPtr(it *db.Item) *db.Item` | See the source definition and the subsystem ownership contract. |
| [`handleWarehouseMove`](../../server/internal/world/warehouse.go#L60)<br>`func (s *Session) handleWarehouseMove(srcBag, srcSlot, destBag, destSlot uint8)` | See the source definition and the subsystem ownership contract. |
| [`warehousePlan`](../../server/internal/world/warehouse.go#L99)<br>`type warehousePlan` | warehousePlan owns private copies; planning cannot write storage or sockets. |
| [`warehouseRearrange`](../../server/internal/world/warehouse.go#L104)<br>`func (s *warehousePlan) warehouseRearrange(src, dest uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`warehouseDeposit`](../../server/internal/world/warehouse.go#L141)<br>`func (s *warehousePlan) warehouseDeposit(srcBag, srcSlot, dest uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`warehouseWithdraw`](../../server/internal/world/warehouse.go#L186)<br>`func (s *warehousePlan) warehouseWithdraw(src, destBag, destSlot uint8) []byte` | See the source definition and the subsystem ownership contract. |
| [`buildShopSellResult`](../../server/internal/world/warehouse.go#L246)<br>`func buildShopSellResult(result, bag, slot, itemType, typeID uint8, remaining uint16, money uint32) []byte` | See the source definition and the subsystem ownership contract. |
| [`handleShopSell`](../../server/internal/world/warehouse.go#L255)<br>`func (s *Session) handleShopSell(p []byte)` | handleShopSell (0x0703): bag, slot, count. Items.Sell is per unit. |

<details><summary>Type definitions and owned fields</summary>

```go
type warehousePlan struct {
	char          *db.Character
	items, stored []db.Item
}
```

</details>

### weather.go

[server/internal/world/weather.go](../../server/internal/world/weather.go#L1)

Direct GX dependencies: `internal/proto`, `server/internal/packet`.

| Declaration | Responsibility / source comment |
|---|---|
| [`weatherDuration`](../../server/internal/world/weather.go#L11)<br>`const weatherDuration = 10 * time.Minute` | See the source definition and the subsystem ownership contract. |
| [`weatherStarted`](../../server/internal/world/weather.go#L13)<br>`var weatherStarted time.Time` | See the source definition and the subsystem ownership contract. |
| [`weatherAt`](../../server/internal/world/weather.go#L15)<br>`func weatherAt(zone ZoneConfig, start, now time.Time) uint8` | See the source definition and the subsystem ownership contract. |
| [`updateWeather`](../../server/internal/world/weather.go#L30)<br>`func (s *Session) updateWeather(now time.Time, force bool)` | See the source definition and the subsystem ownership contract. |
| [`gmWeather`](../../server/internal/world/weather.go#L49)<br>`func (s *Session) gmWeather(args []string) error` | See the source definition and the subsystem ownership contract. |

### worldcal.go

[server/internal/world/worldcal.go](../../server/internal/world/worldcal.go#L1)

Direct GX dependencies: `server/internal/db`, `server/internal/logx`.

| Declaration | Responsibility / source comment |
|---|---|
| [`worldCalMu`](../../server/internal/world/worldcal.go#L20)<br>`var worldCalMu sync.Mutex` | See the source definition and the subsystem ownership contract. |
| [`worldBase`](../../server/internal/world/worldcal.go#L21)<br>`var worldBase` | See the source definition and the subsystem ownership contract. |
| [`worldElapsedAt`](../../server/internal/world/worldcal.go#L22)<br>`var worldElapsedAt time.Time` | See the source definition and the subsystem ownership contract. |
| [`worldElapsedSec`](../../server/internal/world/worldcal.go#L23)<br>`var worldElapsedSec uint32` | See the source definition and the subsystem ownership contract. |
| [`packWorldTime`](../../server/internal/world/worldcal.go#L26)<br>`func packWorldTime(t time.Time) int32` | See the source definition and the subsystem ownership contract. |
| [`unpackWorldTime`](../../server/internal/world/worldcal.go#L32)<br>`func unpackWorldTime(v int32) time.Time` | See the source definition and the subsystem ownership contract. |
| [`InitWorldCalendar`](../../server/internal/world/worldcal.go#L40)<br>`func InitWorldCalendar(log *logx.Logger)` | InitWorldCalendar loads (or creates) the world base and samples the elapsed seconds, as the DB agent does at connection. |
| [`worldDay`](../../server/internal/world/worldcal.go#L62)<br>`func worldDay() uint32` | worldDay is CWorld::GetWorldDay. |
| [`worldBasePacked`](../../server/internal/world/worldcal.go#L72)<br>`func worldBasePacked() int32` | worldBasePacked is the stored base, for creating the WorldInfo row. |
