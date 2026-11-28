// Package game holds the client's view of the world. It is pure: no
// sockets, no GL. The loop that owns a State feeds it every payload with
// Apply and sends the requests it builds through the network layer.
package game

import (
	"encoding/binary"
	"fmt"
	"math"
	"slices"

	"github.com/Spelunkern/Shaiya-GX/internal/proto"
)

// Phase follows the original screens.
type Phase int

const (
	PhaseConnecting Phase = iota
	PhaseFaction          // account has no country yet
	PhaseCharSelect       // 0x0101 list + 0x0109 received
	PhaseLoading          // 0x0104 accepted, initial state arriving
	PhaseInWorld          // 0x0201 for ourselves received
)

func (p Phase) String() string {
	return [...]string{"connecting", "faction", "char-select", "loading", "in-world"}[p]
}

type Kind uint8

const (
	KindUser Kind = iota + 1
	KindMob
	KindNpc
)

// Entity is anything placed in view.
type Entity struct {
	ClimbMode                    uint8
	ClimbIndex                   uint16
	Summoned                     bool
	IsGM                         bool
	Concealed                    bool
	MobSpeedKnown                bool
	MobHeightKnown               bool
	MobCountry                   uint8
	MobCountryKnown              bool
	MobMoveDurationMS            uint32
	MobAttackSpeed, MobMoveSpeed uint8
	Sitting                      bool
	Gesture                      uint8
	GestureSerial                uint64
	PartyBadge                   uint8
	ShopTitle                    string // personal market message, empty when not selling
	MountShape                   uint8
	Vehicle                      proto.VehicleState
	VehicleKnown                 bool
	Mounted                      bool
	GuildID                      uint32
	HP                           int32
	HealthKnown                  bool
	PvPKills                     uint32
	PvPKillsKnown                bool
	Dead                         bool
	Kind                         Kind
	ID                           uint32
	TypeID                       uint16 // mob id or npc type id
	NpcType                      uint8
	NpcAttackTarget              uint32
	NpcAttackSerial              uint64
	X, Y, Z                      float32
	Motion                       uint8
	MoveSerial                   uint64
	PositionResetSerial          uint64
	Dir                          uint16
	Shape                        *proto.UserShape
	Speed                        *proto.CharacterSpeed
}

// ChatLine is one received message.
type ChatLine struct {
	Items            proto.ChatItems
	ItemType, ItemID uint8
	Raid             bool
	MessageID        int
	Style            uint8
	Op               proto.Opcode
	Direction        uint8
	SenderID         uint32
	Sender           string
	Text             string
	Repetitions      uint16
	ChatProtected    bool
	chatSerial       uint64
}

// Event tells the caller what a payload changed; Text is for logs.
type Event struct {
	MapLoadAck bool // Send 0x0201 after initial state or a self cross-map placement.
	ChatQuiet  bool // The player chat was filtered and must stay out of ancillary presentation.
	Op         proto.Opcode
	Text       string
}

type State struct {
	Economy                proto.EconomySettings
	LastMobSpeech          proto.MobSpeech
	RaidNotice             string // last 0x1112 raid notice text
	RaidNoticeSerial       uint64
	TargetEffects          proto.TargetEffectSnapshot
	TargetEffectsSerial    uint64
	LastSkillFailure       proto.SkillFailure
	LastSkillCooldown      proto.SkillCooldown
	OJLimits               map[[2]uint8]proto.OJLimit
	EnchantRules           *proto.EnchantRules
	ChestCloakFits         map[[2]uint8]uint8
	LastSkillDash          proto.SkillDash
	LastPlayerCorrection   proto.PlayerCorrection
	PhysicsMapID           uint16
	PhysicsTime            float64
	MapQuestNPCs           []proto.QuestNPC
	MapQuestNPCsReady      bool
	MapQuestMobs           []proto.QuestMobSpawn
	MapQuestMobsReady      bool
	questMobStaging        []proto.QuestMobSpawn
	questMobStagingActive  bool
	QuestMapRevision       uint64
	QuestInventoryRevision uint64
	questNPCStaging        []proto.QuestNPC
	questNPCStagingActive  bool
	GMTurbo                bool
	GMConcealed, GMImmune  bool // /char on and /attack on, published by GXGMModes
	Social                 SocialState
	LastDetectInfo         proto.DetectInfo
	LastPeriodicEffect     proto.PeriodicEffect
	LastField              proto.Field
	LastKillNotice         string      // name of the player the user killed
	MyShop                 MyShopState // the player's own personal market
	ShopSigns              map[uint32]string
	ShopView               ShopViewState    // another player's market, open
	LastLootNotice         proto.LootNotice // the party member who just received loot
	LootNoticeSerial       uint64
	PlayerForms            map[uint32]proto.PlayerFormAppearance
	Evolution              proto.EvolutionBar
	Transforms             map[uint32]uint16             // character -> monster it looks like
	UserHealth             map[uint32]proto.TargetHealth // HP of the selected user
	DeathLossFree          bool
	LastMotion             proto.CharacterMotionResult
	Blessing               BlessingState
	Weather                proto.Weather
	Mall                   MallState
	Guild                  GuildState
	GuildDuel              GuildDuelState
	Exchange               ExchangeState
	Duel                   DuelState
	Quests                 map[uint16]proto.ActiveQuest
	FinishedQuests         map[uint16]uint8
	QuestChoice            uint16
	QuestChoosing          bool
	LastQuestEnd           proto.QuestEndResult
	Warehouse              map[uint8]proto.InventoryItem
	WarehouseReady         bool
	WarehouseGold          uint32
	Party                  PartyState
	LastItemUse            proto.ItemUseResult
	LastVehicle            proto.VehicleResult
	VehicleInviteID        uint32
	VehicleInviteSerial    uint64
	VehicleReply           uint8
	VehicleReplySerial     uint64
	VehicleSerial          uint64
	vehicleShapeExpected   bool
	LastGatekeeper         proto.GatekeeperResult
	GatekeeperSerial       uint64
	LastCompose            proto.ComposeResult
	ComposeSerial          uint64
	LastEnchant            proto.EnchantResult
	EnchantSerial          uint64
	LastGemAdd             proto.GemAddResult
	LastGemRemove          proto.GemRemoveResult
	LastEmblem             proto.EmblemResult
	EmblemSerial           uint64
	EmblemRemoveSerial     uint64
	GemRemoveSerial        uint64
	GemAddSerial           uint64
	LastShop               proto.ShopResult
	SkillBarReady          bool
	Skills                 map[uint8]proto.LearnedSkill
	QuickSlots             []proto.QuickSlot
	LastLearn              proto.LearnResult
	LastSkill              proto.SkillResult
	LastSkillExtra         proto.SkillResult
	LastReflectedDamage    proto.ReflectedDamage
	LastCast               proto.SkillCast
	GroundItems            map[uint32]proto.GroundItem
	InventoryReady         bool
	Inventory              map[[2]uint8]proto.InventoryItem
	Target                 uint32
	PreviousTarget         uint32
	Pools                  proto.Pools
	LastAttack             proto.AttackResult
	LastFallDamage         proto.FallDamage
	LastDrowningDamage     proto.EnvironmentalDamage
	LastDrowningDeath      uint32
	Oxygen                 proto.OxygenStatus
	PvPCountsKnown         [4]bool
	PvPRewards             proto.PvPRewards
	LastPvPReward          proto.PvPRewardResult
	Relics                 []proto.Relic // current map's relics (0x2101/0x2102)
	LastRelicNotice        proto.RelicDestroyed
	LastServerNotice       string
	LastPortalNotice       proto.PortalOwnerNotice
	Phase                  Phase
	Faction                proto.Faction
	Slots                  map[uint8]proto.CharacterSlot // non-empty slots only

	SelfID          uint32
	MapID           uint16
	Details         proto.Details
	Stats           proto.AdditionalStats
	StatsKnown      bool
	LastStatPoints  proto.StatPoints
	Buffs           []proto.ActiveBuff
	LastBuff        proto.ActiveBuff
	LastBuffRemoved uint32
	Self            Entity

	Entities  map[uint32]*Entity
	Chat      []ChatLine
	chatGuard chatProtection

	Unknown map[proto.Opcode]int // opcodes received but not modeled
}

func New() *State {
	return &State{Skills: map[uint8]proto.LearnedSkill{}, GroundItems: map[uint32]proto.GroundItem{}, Inventory: map[[2]uint8]proto.InventoryItem{}, Slots: map[uint8]proto.CharacterSlot{}, Entities: map[uint32]*Entity{}, Unknown: map[proto.Opcode]int{}}
}

// Apply updates the state from one payload. Malformed packets are
// reported as an error and leave the state as it was.
func (s *State) Apply(p []byte) (result Event, applyErr error) {
	op := proto.Op(p)
	defer func() {
		if applyErr == nil && questMapStateChanged(op) {
			s.QuestMapRevision++
		}
		if applyErr == nil && questMapInventoryChanged(op) {
			s.QuestInventoryRevision++
		}
	}()
	ev := Event{Op: op}
	fail := ev.packetError
	switch op {
	case proto.GXItemCooldownReset:
		if len(p) != 2 {
			return fail(fmt.Errorf("invalid cooldown reset length %d", len(p)))
		}
	case proto.GXChestCloakFits:
		rows, err := proto.DecodeChestCloakFits(p)
		if err != nil {
			return fail(err)
		}
		if s.ChestCloakFits == nil {
			s.ChestCloakFits = make(map[[2]uint8]uint8)
		}
		for _, row := range rows {
			s.ChestCloakFits[[2]uint8{row.Type, row.TypeID}] = row.Fit
		}
	case proto.GXEconomySettings:
		settings, err := proto.DecodeEconomySettings(p)
		if err != nil {
			return fail(err)
		}
		s.Economy = settings
	case proto.GXEnchantRules:
		rules, err := proto.DecodeEnchantRules(p)
		if err != nil {
			return fail(err)
		}
		s.EnchantRules = &rules
	case proto.GXOJLimits:
		limits, err := proto.DecodeOJLimits(p)
		if err != nil {
			return fail(err)
		}
		if s.OJLimits == nil {
			s.OJLimits = make(map[[2]uint8]proto.OJLimit)
		}
		for _, limit := range limits {
			s.OJLimits[[2]uint8{limit.Type, limit.TypeID}] = limit
		}
	case proto.GXQuestMobs:
		if err := s.applyQuestMobs(p); err != nil {
			return fail(err)
		}
	case proto.GXQuestNPCs:
		if err := s.applyQuestNPCs(p); err != nil {
			return fail(err)
		}
	case proto.FriendList, proto.FriendInvite, proto.FriendAnswer, proto.FriendAdded, proto.FriendDelete, proto.FriendMemo, proto.FriendPresence, proto.BlockList, proto.BlockAdd, proto.BlockDelete, proto.BlockMemo, proto.PartySearchRequest, proto.PartySearchList, proto.SocialError:
		if err := s.applySocial(p); err != nil {
			return fail(err)
		}
	case proto.CharacterMotion:
		m, err := proto.DecodeCharacterMotion(p)
		if err != nil {
			return fail(err)
		}
		s.LastMotion = m
		e := &s.Self
		if m.ID != s.SelfID {
			e = s.Entities[m.ID]
			if e == nil {
				e = &Entity{Kind: KindUser, ID: m.ID}
				s.Entities[m.ID] = e
			}
		}
		if m.Motion == proto.MotionSitDown {
			e.Sitting = true
		} else if m.Motion == proto.MotionStandUp {
			e.Sitting = false
		} else if _, ok := proto.GestureAction(m.Motion); ok || m.Motion == proto.MotionJump || m.Motion == proto.MotionBackstep || m.Motion == proto.MotionEvadeLeft || m.Motion == proto.MotionEvadeRight {
			e.Gesture, e.GestureSerial = m.Motion, e.GestureSerial+1
		}
	case proto.GXWeather:
		value, err := proto.DecodeWeather(p)
		if err != nil {
			return fail(err)
		}
		s.Weather = value
	case proto.GXBlessingSnapshot:
		value, err := proto.DecodeBlessingSnapshot(p)
		if err != nil {
			return fail(err)
		}
		s.Blessing.applySnapshot(value)
	case proto.BlessInit, proto.CountryKillCount:
		value, err := proto.DecodeBlessing(p)
		if err != nil {
			return fail(err)
		}
		s.Blessing.apply(op, value)
	case proto.PvPRewardStatus:
		value, err := proto.DecodePvPRewards(p)
		if err != nil {
			return fail(err)
		}
		s.PvPRewards = value
	case proto.PvPRewardKill, proto.PvPRewardDeath:
		value, err := proto.DecodePvPRewardResult(p, op == proto.PvPRewardDeath)
		if err != nil {
			return fail(err)
		}
		s.LastPvPReward = value
		if value.OK {
			if op == proto.PvPRewardKill {
				s.Details.StatPoint += uint16(value.Value)
				if s.PvPRewards.KillStage < proto.PvPRewardFinishedStage {
					s.PvPRewards.KillStage++
				}
			} else {
				s.Details.Money = uint32(min(uint64(s.Details.Money)+uint64(value.Value), uint64(^uint32(0))))
				if s.PvPRewards.DeathStage < proto.PvPRewardFinishedStage {
					s.PvPRewards.DeathStage++
				}
			}
		}
	case proto.UserKillCountUpdate:
		value, err := proto.DecodePvPCountUpdate(p)
		if err != nil {
			return fail(err)
		}
		if value.Index < uint8(len(s.PvPCountsKnown)) {
			s.PvPCountsKnown[value.Index] = true
		}
		switch value.Index {
		case 0:
			s.Details.Kills = value.Value
			s.Self.PvPKills, s.Self.PvPKillsKnown = value.Value, true
			if s.Self.Shape != nil {
				s.Self.Shape.Kills = value.Value
			}
		case 1:
			s.Details.Deaths = value.Value
		case 2:
			s.Details.Wins = value.Value
		case 3:
			s.Details.Losses = value.Value
		}
	case proto.RelicList:
		list, err := proto.DecodeRelicList(p)
		if err != nil {
			return fail(err)
		}
		s.Relics = list
	case proto.RelicOwner:
		v, err := proto.DecodeRelicOwner(p)
		if err != nil {
			return fail(err)
		}
		// Native 0x564f20 updates only an ID announced in the snapshot.
		for i := range s.Relics {
			if s.Relics[i].ID == v.ID {
				s.Relics[i].Owner = v.Owner
			}
		}
	case proto.PortalNotice:
		v, err := proto.DecodePortalNotice(p)
		if err != nil {
			return fail(err)
		}
		s.LastPortalNotice = v
	case proto.GXServerNotice:
		text, err := proto.DecodeServerNotice(p)
		if err != nil {
			return fail(err)
		}
		s.LastServerNotice = text
		ev.Text = text
	case proto.RelicNotice:
		v, err := proto.DecodeRelicNotice(p)
		if err != nil {
			return fail(err)
		}
		s.LastRelicNotice = v
	case proto.CharacterPvPKills:
		value, err := proto.DecodeVisiblePvPKills(p)
		if err != nil {
			return fail(err)
		}
		entity := &s.Self
		if value.ID != s.SelfID {
			entity = s.Entities[value.ID]
			if entity == nil {
				entity = &Entity{Kind: KindUser, ID: value.ID}
				s.Entities[value.ID] = entity
			}
		}
		entity.PvPKills, entity.PvPKillsKnown = value.Kills, true
		if entity.Shape != nil {
			entity.Shape.Kills = value.Kills
		}
		if value.ID == s.SelfID {
			s.Details.Kills = value.Kills
		}
	case proto.GXVehicleState:
		v, err := proto.DecodeVehicleState(p)
		if err != nil {
			return fail(err)
		}
		e := &s.Self
		if v.ID != s.SelfID {
			e = s.Entities[v.ID]
			if e == nil {
				e = &Entity{Kind: KindUser, ID: v.ID}
				s.Entities[v.ID] = e
			}
		}
		e.Vehicle = v
		e.VehicleKnown = true
		e.Mounted = v.State >= proto.VehicleDriver
		e.MountShape = v.Model
	case proto.VehicleInvite:
		r := proto.NewReader(p)
		s.VehicleInviteID = r.U32()
		if err := r.Err(); err != nil {
			return fail(err)
		}
		s.VehicleInviteSerial++
	case proto.VehicleAnswer:
		r := proto.NewReader(p)
		s.VehicleReply = r.U8()
		if err := r.Err(); err != nil {
			return fail(err)
		}
		s.VehicleReplySerial++
	case proto.VehicleCast:
		// The canonical GX state follows this native announcement.
	case proto.VehicleLeave:
		// Passenger links are authoritative in GXVehicleState.
	case proto.UseVehicle:
		result, err := proto.DecodeVehicleResult(p)
		if err != nil {
			return fail(err)
		}
		s.LastVehicle = result
		s.VehicleSerial++
		if result.Result == 1 && !s.Self.VehicleKnown {
			s.Self.Mounted = !s.Self.Mounted
			s.vehicleShapeExpected = true
		}
	case proto.CharacterMountShape:
		shape, err := proto.DecodeMountShape(p)
		if err != nil {
			return fail(err)
		}
		if shape.ID == s.SelfID {
			if s.Self.VehicleKnown {
				break
			}
			s.Self.MountShape = shape.Shape
			if s.vehicleShapeExpected {
				s.vehicleShapeExpected = false
			} else {
				s.Self.Mounted = shape.Shape != 0
			}
		} else {
			e := s.Entities[shape.ID]
			if e == nil {
				e = &Entity{Kind: KindUser, ID: shape.ID}
				s.Entities[shape.ID] = e
			}
			if !e.VehicleKnown {
				e.MountShape, e.Mounted = shape.Shape, shape.Shape != 0
			}
		}
	case proto.Gatekeeper:
		result, err := proto.DecodeGatekeeperResult(p)
		if err != nil {
			return fail(err)
		}
		s.Details.Money = result.Money
		s.LastGatekeeper = result
		s.GatekeeperSerial++
	case proto.ItemCompose:
		result, err := proto.DecodeComposeResult(p)
		if err != nil {
			return fail(err)
		}
		key := [2]uint8{result.Bag, result.Slot}
		item, ok := s.Inventory[key]
		if result.Result == 0 && !ok {
			return fail(fmt.Errorf("recreation references missing equipment"))
		}
		if result.Result == 0 {
			item.Craft = result.Craft
			s.Inventory[key] = item
		}
		s.LastCompose = result
		s.ComposeSerial++
	case proto.EnchantAdd:
		result, err := proto.DecodeEnchantResult(p)
		if err != nil {
			return fail(err)
		}
		lapKey := [2]uint8{result.LapisiaBag, result.LapisiaSlot}
		matKey := [2]uint8{result.MaterialBag, result.MaterialSlot}
		itemKey := [2]uint8{result.ItemBag, result.ItemSlot}
		lap, lapOK := s.Inventory[lapKey]
		mat, matOK := s.Inventory[matKey]
		item, itemOK := s.Inventory[itemKey]
		if !lapOK || !matOK || !itemOK {
			return fail(fmt.Errorf("enchant response references missing item"))
		}
		lap.Count, mat.Count, item.Craft = result.LapisiaCount, result.MaterialCount, result.Craft
		s.applyItem(lap)
		s.applyItem(mat)
		s.Inventory[itemKey] = item
		s.Details.Money = result.Money
		s.LastEnchant = result
		s.EnchantSerial++
	case proto.ItemRepair:
		result, err := proto.DecodeItemRepair(p)
		if err != nil {
			return fail(err)
		}
		s.Details.Money = result.Money
	case proto.GemRemove:
		result, err := proto.DecodeGemRemoveResult(p)
		if err != nil {
			return fail(err)
		}
		key := [2]uint8{result.ItemBag, result.ItemSlot}
		item, ok := s.Inventory[key]
		if result.Result == 1 && !ok {
			return fail(fmt.Errorf("gem removal references missing item"))
		}
		if ok && result.Result == 1 {
			for i, removed := range result.Removed {
				if removed != 0 {
					item.Gems[i] = 0
				}
			}
			s.Inventory[key] = item
		}
		s.Details.Money = result.Money
		s.LastGemRemove = result
		s.GemRemoveSerial++
	case proto.CloakAdd:
		result, err := proto.DecodeEmblemResult(p)
		if err != nil {
			return fail(err)
		}
		key := [2]uint8{result.Bag, result.Slot}
		item, ok := s.Inventory[key]
		if result.Result == 1 && !ok {
			return fail(fmt.Errorf("emblem references missing item"))
		}
		if ok && result.Result == 1 {
			item.Gems = result.Pattern
			s.Inventory[key] = item
		}
		s.Details.Money = result.Money
		s.LastEmblem = result
		s.EmblemSerial++
	case proto.CloakRemove:
		result, err := proto.DecodeEmblemRemoved(p)
		if err != nil {
			return fail(err)
		}
		key := [2]uint8{result.Bag, result.Slot}
		if item, ok := s.Inventory[key]; ok {
			item.Gems = [6]uint8{}
			s.Inventory[key] = item
		}
		s.EmblemRemoveSerial++
	case proto.GemAdd:
		result, err := proto.DecodeGemAddResult(p)
		if err != nil {
			return fail(err)
		}
		gemKey, itemKey := [2]uint8{result.GemBag, result.GemSlot}, [2]uint8{result.ItemBag, result.ItemSlot}
		gem, gemOK := s.Inventory[gemKey]
		item, itemOK := s.Inventory[itemKey]
		if !gemOK || (result.Result == 1 && !itemOK) {
			return fail(fmt.Errorf("gem result references missing item"))
		}
		if result.Result == 1 {
			item.Gems[result.Position] = result.GemID
			s.Inventory[itemKey] = item
		}
		gem.Count = result.GemCount
		s.applyItem(gem)
		s.Details.Money = result.Money
		s.LastGemAdd = result
		s.GemAddSerial++
	case proto.GuildDuelRequest, proto.GuildDuelAnswer, proto.GuildDuelCountdown, proto.GuildDuelStart, proto.GuildDuelResult, proto.GuildDuelLeave:
		if err := s.applyGuildDuel(p); err != nil {
			return fail(err)
		}
	case proto.GuildListBegin, proto.GuildListEnd, proto.GuildList, proto.GuildNew, proto.GuildDisbanded, proto.GuildOnlineList, proto.GuildOfflineList, proto.GuildMemberAdded, proto.GuildApplicantList, proto.GuildApplicant, proto.GuildApplicantRemoved, proto.GuildApply, proto.GuildStatus, proto.GuildJoinResult, proto.GuildCreate, proto.GuildCreateAgree, proto.GuildCreateFailed, proto.GuildLeave, proto.GuildDisband, proto.GuildDisbandAgree, proto.GuildExpel, proto.GuildBankList, proto.GuildBankPlaced, proto.GuildBankTaken, proto.GuildBankResult, proto.GuildRankingNotice, proto.GuildRankingPoints, proto.GuildRankingList, proto.GuildEtin, proto.GuildRemark, proto.GuildNpcList, proto.GuildNpcUpgrade, proto.GuildHouseBuy, proto.GuildHouseKeepEtin, proto.GuildDonate:
		if err := s.applyGuild(p); err != nil {
			return fail(err)
		}
	case proto.MyShopBegin, proto.MyShopCancel, proto.MyShopOpen, proto.MyShopStop, proto.MyShopAdd, proto.MyShopRemove, proto.MyShopPrice, proto.MyShopSold, proto.MyShopEnter, proto.MyShopList, proto.MyShopStock, proto.MyShopBuy, proto.MyShopClosed, proto.MyShopLeave, proto.ShopBubble:
		if err := s.applyMyShop(p); err != nil {
			return fail(err)
		}
	case proto.DuelRequest, proto.DuelAnswer, proto.DuelCountdown, proto.DuelStart, proto.DuelResult, proto.DuelEnd, proto.DuelWindow, proto.DuelWindowEnd, proto.DuelReady, proto.DuelAdd, proto.DuelRemove, proto.DuelMoney, proto.DuelAddOther:
		if err := s.applyDuel(p); err != nil {
			return fail(err)
		}
	case proto.ExchangeRequest, proto.ExchangeAnswer, proto.ExchangeStart, proto.ExchangeEnd, proto.ExchangeReady, proto.ExchangeAdd, proto.ExchangeAddDest, proto.ExchangeAddMoney, proto.ExchangeConfirm:
		if err := s.applyExchange(p); err != nil {
			return fail(err)
		}
	case proto.QuestList, proto.QuestFinishedList, proto.QuestStart, proto.QuestCounter, proto.QuestChoose, proto.QuestEnd:
		return s.applyQuestsPacket(p, ev)
	case proto.PartyRoster, proto.PartyJoined, proto.PartyInvite, proto.PartyInviteName, proto.PartyAnswer, proto.PartyLeft, proto.PartyExpelled, proto.PartyLeader, proto.PartyBadge, proto.PartyPools, proto.PartyLevel, proto.RaidLevel, proto.PartyMove, proto.RaidMove, proto.PartyEffectAdd, proto.PartyEffectRemove, proto.RaidEffectAdd, proto.RaidEffectRemove, proto.RaidRoster, proto.RaidJoined, proto.PartyLeaveAlt, proto.RaidConvert, proto.RaidDivision, proto.RaidAutoJoin, proto.RaidSwap, proto.RaidSubLeader, proto.RaidInvite, proto.RaidAnswer, proto.RaidDisband, proto.RaidLeader, proto.RaidInviteName, proto.RaidError, proto.RaidExpel, proto.RaidPools:
		if err := s.applyParty(p); err != nil {
			return fail(err)
		}
	case proto.CashPoint, proto.PointBuy, proto.PointGiftSend, proto.PointStoredItemList, proto.PointStoredItemGet:
		if err := s.applyMall(p); err != nil {
			return fail(err)
		}
	case proto.CharacterActiveBuffs:
		buffs, err := proto.DecodeActiveBuffs(p)
		if err != nil {
			return fail(err)
		}
		s.Buffs = nil
		for _, b := range buffs {
			if err := s.addBuff(b); err != nil {
				return fail(err)
			}
		}
	case proto.BuffAdd:
		b, err := proto.DecodeBuffAdd(p)
		if err != nil {
			return fail(err)
		}
		if err := s.addBuff(b); err != nil {
			return fail(err)
		}
		s.LastBuff = b
	case proto.BuffRemove:
		id, err := proto.DecodeU32(p)
		if err != nil {
			return fail(err)
		}
		s.removeBuff(id)
		s.LastBuffRemoved = id
	case proto.SetStatPoint:
		d, err := proto.DecodeStatPoints(p)
		if err != nil {
			return fail(err)
		}
		s.applyStatPoints(d)
	case proto.StatsReset:
		d, err := proto.DecodeStatReset(p)
		if err != nil {
			return fail(err)
		}
		if d.Result == 1 {
			s.Details.StatPoint = d.Points
			for i, c := range s.Slots {
				if c.ID == s.SelfID {
					c.Str, c.Rec, c.Int, c.Wis, c.Dex, c.Luc = d.Str, d.Rec, d.Int, d.Wis, d.Dex, d.Luc
					s.Slots[i] = c
				}
			}
		}
	case proto.CharacterAdditionalStats:
		stats, err := proto.DecodeAdditionalStats(p)
		if err != nil {
			return fail(err)
		}
		s.Stats, s.StatsKnown = stats, true
	case proto.CharacterMaxEnergy:
		m, err := proto.DecodeMaxEnergy(p)
		if err != nil {
			return fail(err)
		}
		if m.ID == s.SelfID {
			switch m.Kind {
			case 0:
				s.Details.MaxHP = m.Value
			case 1:
				s.Details.MaxSP = m.Value
			case 2:
				s.Details.MaxMP = m.Value
			}
		}
	case proto.CharacterDeathState:
		m, err := proto.DecodeCharacterSpeed(p)
		if err != nil {
			return fail(err)
		}
		e := &s.Self
		if m.ID != s.SelfID {
			e = s.Entities[m.ID]
			if e == nil {
				e = &Entity{Kind: KindUser, ID: m.ID}
				s.Entities[m.ID] = e
			}
		}
		e.Speed = &m
	case proto.UseItem:
		result, err := proto.DecodeItemUse(p)
		if err != nil {
			return fail(err)
		}
		if result.ID == s.SelfID {
			s.LastItemUse = result
			if result.Type != 0 {
				it := s.Inventory[[2]uint8{result.Bag, result.Slot}]
				it.Bag, it.Slot, it.Type, it.TypeID, it.Count = result.Bag, result.Slot, result.Type, result.TypeID, result.Count
				s.applyItem(it)
			}
		}
	case proto.ShopBuy, proto.ShopSell:
		result, err := proto.DecodeShopResult(p)
		if err != nil {
			return fail(err)
		}
		s.LastShop = result
		if result.Result == 0 {
			s.Details.Money = result.Money
			it := s.Inventory[[2]uint8{result.Bag, result.Slot}]
			it.Bag, it.Slot, it.Type, it.TypeID, it.Count = result.Bag, result.Slot, result.Type, result.TypeID, result.Count
			s.applyItem(it)
		}
	case proto.CharacterSkills:
		skills, err := proto.DecodeSkills(p)
		if err != nil {
			return fail(err)
		}
		s.Skills = map[uint8]proto.LearnedSkill{}
		for _, k := range skills {
			s.Skills[k.Number] = k
		}
	case proto.CharacterSkillBar:
		slots, err := proto.DecodeQuickSlots(p)
		if err != nil {
			return fail(err)
		}
		s.QuickSlots = slots
		s.SkillBarReady = true
		ev.MapLoadAck = s.Phase == PhaseLoading
	case proto.LearnNewSkill:
		result, err := proto.DecodeLearnResult(p)
		if err != nil {
			return fail(err)
		}
		s.LastLearn = result
		if result.Result == 0 {
			s.Skills[result.Number] = proto.LearnedSkill{ID: result.ID, Level: result.Level, Number: result.Number}
		}
	case proto.SkillCastStarted:
		cast, err := proto.DecodeSkillCast(p)
		if err != nil {
			return fail(err)
		}
		s.LastCast = cast
	case proto.TargetEffects:
		snapshot, err := proto.DecodeTargetEffects(p)
		if err != nil {
			return fail(err)
		}
		s.TargetEffects = snapshot
		s.TargetEffectsSerial++
	case proto.TargetEffectAdd, proto.TargetEffectRemove:
		snapshot, e, err := proto.DecodeTargetEffectChange(p)
		if err != nil {
			return fail(err)
		}
		if snapshot.ID == s.TargetEffects.ID && snapshot.Kind == s.TargetEffects.Kind {
			kept := s.TargetEffects.Effects[:0]
			for _, old := range s.TargetEffects.Effects {
				if old.Skill != e.Skill || old.Level != e.Level {
					kept = append(kept, old)
				}
			}
			if proto.Op(p) == proto.TargetEffectAdd {
				kept = append(kept, e)
			}
			s.TargetEffects.Effects = kept
			s.TargetEffectsSerial++
		}
	case proto.GXSkillCooldown:
		result, err := proto.DecodeSkillCooldown(p)
		if err != nil {
			return fail(err)
		}
		s.LastSkillCooldown = result
	case proto.GXLootNotice:
		n, err := proto.DecodeLootNotice(p)
		if err != nil {
			return fail(err)
		}
		s.LastLootNotice = n
		s.LootNoticeSerial++
	case proto.GXKillNotice:
		name, err := proto.DecodeKillNotice(p)
		if err != nil {
			return fail(err)
		}
		s.LastKillNotice = name
	case proto.GXTargetHealth:
		result, err := proto.DecodeTargetHealth(p)
		if err != nil {
			return fail(err)
		}
		s.UserHealth = map[uint32]proto.TargetHealth{result.User: result} // only the selected user is kept
	case proto.GXFormAppearance:
		form, err := proto.DecodePlayerForm(p)
		if err != nil {
			return fail(err)
		}
		if form.Kind == proto.FormNone {
			delete(s.PlayerForms, form.User)
		} else {
			if s.PlayerForms == nil {
				s.PlayerForms = map[uint32]proto.PlayerFormAppearance{}
			}
			s.PlayerForms[form.User] = form
		}
	case proto.GXEvolutionBar:
		bar, err := proto.DecodeEvolutionBar(p)
		if err != nil {
			return fail(err)
		}
		s.Evolution = bar
	case proto.GXTransform:
		result, err := proto.DecodeTransform(p)
		if err != nil {
			return fail(err)
		}
		if s.Transforms == nil {
			s.Transforms = map[uint32]uint16{}
		}
		if result.Mob == 0 {
			delete(s.Transforms, result.User)
		} else {
			s.Transforms[result.User] = result.Mob
		}
	case proto.GXField:
		result, err := proto.DecodeField(p)
		if err != nil {
			return fail(err)
		}
		s.LastField = result
	case proto.GXSkillFailure:
		result, err := proto.DecodeSkillFailure(p)
		if err != nil {
			return fail(err)
		}
		s.LastSkillFailure = result
	case proto.GXLadder:
		result, err := proto.DecodeLadderState(p)
		if err != nil {
			return fail(err)
		}
		if result.ID == s.SelfID {
			s.Self.ClimbMode, s.Self.ClimbIndex = result.Mode, result.Index
		} else if e := s.Entities[result.ID]; e != nil {
			e.ClimbMode, e.ClimbIndex = result.Mode, result.Index
		}
	case proto.GXOxygen:
		status, err := proto.DecodeOxygen(p)
		if err != nil {
			return fail(err)
		}
		s.Oxygen = status
	case proto.GXDrowningDamage:
		result, err := proto.DecodeDrowningDamage(p)
		if err != nil {
			return fail(err)
		}
		s.LastDrowningDamage = result
	case proto.GXDrowningDeath:
		id, err := proto.DecodeDrowningDeath(p)
		if err != nil {
			return fail(err)
		}
		s.LastDrowningDeath = id
	case proto.GXFallDamage:
		result, err := proto.DecodeFallDamage(p)
		if err != nil {
			return fail(err)
		}
		s.LastFallDamage = result
	case proto.SkillKeep, proto.MobPeriodicDamage:
		result, err := proto.DecodePeriodicEffect(p)
		if err != nil {
			return fail(err)
		}
		s.LastPeriodicEffect = result
	case proto.SkillDetectInfo:
		result, err := proto.DecodeDetectInfo(p)
		if err != nil {
			return fail(err)
		}
		s.LastDetectInfo = result
	case proto.SkillUse, proto.SkillUseNoTarget, proto.SkillUserDrain, proto.SkillMobDrain, proto.SkillPartyMember, proto.SkillMobOther, proto.MobSkillUse, proto.MobSkillDrain, proto.MobSkillOther:
		result, err := proto.DecodeSkillResult(p)
		if err != nil {
			return fail(err)
		}
		s.LastSkill = result
	case proto.SkillUserExtraDamage, proto.SkillExtraDamage:
		result, err := proto.DecodeSkillExtra(p)
		if err != nil {
			return fail(err)
		}
		s.LastSkillExtra = result
	case proto.SkillUserReflectedDamage, proto.SkillMobReflectedDamage:
		result, err := proto.DecodeReflectedDamage(p)
		if err != nil {
			return fail(err)
		}
		s.LastReflectedDamage = result
	case proto.SkillUsedSpMp:
		r := proto.NewReader(p)
		sp, mp := r.U32(), r.U32()
		if err := r.Err(); err != nil {
			return fail(err)
		}
		s.Pools.SP -= min(s.Pools.SP, sp)
		s.Pools.MP -= min(s.Pools.MP, mp)
	case proto.ExpGain:
		gain, err := proto.DecodeExperienceGain(p)
		if err != nil {
			return fail(err)
		}
		s.Details.Exp = uint32(min(uint64(s.Details.Exp)+uint64(gain.Base)+uint64(gain.Perfect), uint64(^uint32(0))))
	case proto.LevelUpSelf, proto.LevelUpOther:
		r := proto.NewReader(p)
		id, level, stats, skills := r.U32(), r.U16(), r.U16(), r.U16()
		prev, next := r.U32(), r.U32()
		if err := r.Err(); err != nil {
			return fail(err)
		}
		for i, c := range s.Slots {
			if c.ID == id {
				// Native/Rapture levelUp grants +1 to the job's primary
				// attribute per level; it is not an allocated stat point.
				if level > c.Level && c.Class < 6 {
					var d proto.StatPoints
					d[[...]int{0, 2, 1, 5, 3, 4}[c.Class]] = level - c.Level
					addAttributes(&c, d)
				}
				c.Level = level
				s.Slots[i] = c
			}
		}
		if id == s.SelfID {
			s.Details.StatPoint, s.Details.SkillPoint = stats, skills
			s.Details.PrevExp, s.Details.NextExp = prev, next
		}
	case proto.CharacterMaxHpMpSp:
		r := proto.NewReader(p)
		id, hp, mp, sp := r.U32(), r.U32(), r.U32(), r.U32()
		if err := r.Err(); err != nil {
			return fail(err)
		}
		if id == s.SelfID {
			s.Details.MaxHP, s.Details.MaxMP, s.Details.MaxSP = hp, mp, sp
		}
	case proto.MapAddItem, proto.MapRemoveItem, proto.CharacterItems, proto.ItemBagToBag, proto.WarehouseItemList, proto.AddItem, proto.RemoveItem, proto.SetMoney:
		return s.applyInventoryPacket(p, ev)
	case proto.MobSpeed:
		r := proto.NewReader(p)
		id, attack, move := r.U32(), r.U8(), r.U8()
		if err := r.Err(); err != nil {
			return fail(err)
		}
		if e := s.Entities[id]; e != nil && e.Kind == KindMob {
			e.MobSpeedKnown = true
			e.MobAttackSpeed, e.MobMoveSpeed = attack, move
		}
	case proto.TargetMobGetState:
		r := proto.NewReader(p)
		id, hp := r.U32(), r.I32()
		attack, move := r.U8(), r.U8()
		if err := r.Err(); err != nil {
			return fail(err)
		}
		if e := s.Entities[id]; e != nil && e.Kind == KindMob {
			e.HP = hp
			e.HealthKnown = true
			e.MobSpeedKnown = true
			e.MobAttackSpeed, e.MobMoveSpeed = attack, move
			e.Dead = hp == 0
		}
	case proto.CharacterResurrect:
		m, err := proto.DecodeResurrection(p)
		if err != nil {
			return fail(err)
		}
		e := s.Entities[m.ID]
		if m.ID == s.SelfID {
			e = &s.Self
			s.Details.Exp = m.Exp
			s.Target = 0
			s.Duel.Outcome = 0
			s.DeathLossFree = false
			s.GuildDuel.deathLeavePending = false
		}
		if e != nil {
			e.Dead = false
			e.Sitting = false
			e.PositionResetSerial++
			e.ClimbMode = 0
			e.X, e.Y, e.Z = m.X, m.Y, m.Z
			e.Motion = 0
		}
	case proto.CharacterReviveHitpoints:
		r := proto.NewReader(p)
		id := r.U32()
		hp, mp, sp := r.U32(), r.U32(), r.U32()
		if err := r.Err(); err != nil {
			return fail(err)
		}
		if id == s.SelfID {
			s.Pools = proto.Pools{HP: hp, MP: mp, SP: sp}
		}
	case proto.CharacterCurrentHitpoints:
		pools, err := proto.DecodePools(p)
		if err != nil {
			return fail(err)
		}
		s.Pools = pools
	case proto.TargetMobHpUpdate, proto.MobCurrentHitpoints:
		hp, err := proto.DecodeMobHealth(p)
		if err != nil {
			return fail(err)
		}
		if e := s.Entities[hp.ID]; e != nil && e.Kind == KindMob {
			e.HP = hp.HP
			e.HealthKnown = true
			e.Dead = hp.HP == 0
		}
	case proto.CharacterMobAutoAttack, proto.CharacterUserAutoAttack, proto.MobAttackCharacter:
		attack, err := proto.DecodeAttackResult(p)
		if err != nil {
			return fail(err)
		}
		s.LastAttack = attack
	case proto.MobDeath, proto.CharacterDeath:
		r := proto.NewReader(p)
		id := r.U32()
		deathType := r.U8()
		killer := r.U32()
		if err := r.Err(); err != nil {
			return fail(err)
		}
		if id == s.SelfID && op == proto.CharacterDeath {
			if !s.Self.Dead {
				s.DeathLossFree = deathType == 2 && killer == 0 || deathType == 1 && (s.Duel.Outcome == 2 || s.GuildDuel.deathLeavePending)
				s.GuildDuel.deathLeavePending = false
			}
			s.Self.Dead = true
			s.Oxygen = proto.OxygenStatus{}
			s.Pools.HP = 0
		}
		if e := s.Entities[id]; e != nil {
			e.Dead = true
			e.HP = 0
			e.HealthKnown = true
		}
		if s.Target == id {
			s.Target = 0
		}
	case proto.AccountFaction:
		f, err := proto.DecodeFaction(p)
		if err != nil {
			return fail(err)
		}
		s.Faction = f
		if f.Faction > 1 {
			s.Phase = PhaseFaction
		} else if s.Phase < PhaseLoading {
			s.Phase = PhaseCharSelect
		}
		ev.Text = fmt.Sprintf("faction=%d", f.Faction)
	case proto.CharacterList:
		c, err := proto.DecodeCharacterSlot(p)
		if err != nil {
			return fail(err)
		}
		if c.ID == 0 {
			delete(s.Slots, c.Slot)
		} else {
			s.Slots[c.Slot] = c
			ev.Text = fmt.Sprintf("slot %d: %s lv%d id=%d", c.Slot, c.Name, c.Level, c.ID)
		}
	case proto.CreateCharacter:
		v, err := proto.DecodeU8(p)
		if err != nil {
			return fail(err)
		}
		ev.Text = fmt.Sprintf("create result=%d", v)
	case proto.DeleteCharacter, proto.RestoreCharacter:
		r, err := proto.DecodeIDResult(p)
		if err != nil {
			return fail(err)
		}
		if r.Result == 0 {
			for slot, ch := range s.Slots {
				if ch.ID == r.ID {
					ch.PendingDelete = op == proto.DeleteCharacter
					s.Slots[slot] = ch
				}
			}
		}
		ev.Text = fmt.Sprintf("result=%d id=%d", r.Result, r.ID)
	case proto.SelectCharacter:
		r, err := proto.DecodeIDResult(p)
		if err != nil {
			return fail(err)
		}
		if r.Result == 0 {
			s.SelfID, s.Phase = r.ID, PhaseLoading
			s.Chat = nil // every chat pane, combat and system line belongs to the previous character
			s.Oxygen = proto.OxygenStatus{}
			s.StatsKnown = false
			s.DeathLossFree = false
			s.Blessing = BlessingState{}
			s.Weather = proto.Weather{}
			s.PvPCountsKnown = [4]bool{}
			s.PvPRewards = proto.PvPRewards{}
			s.LastPvPReward = proto.PvPRewardResult{}
			s.Mall = MallState{}
			s.Exchange = ExchangeState{}
			s.Duel = DuelState{}
			s.Guild = GuildState{}
			s.Relics = nil
			s.LastRelicNotice = proto.RelicDestroyed{}
			s.LastPortalNotice = proto.PortalOwnerNotice{}
			s.GuildDuel = GuildDuelState{}
			s.Quests, s.FinishedQuests = nil, nil
			s.QuestChoosing, s.QuestChoice = false, 0
			s.Buffs = nil
			s.Inventory = map[[2]uint8]proto.InventoryItem{}
			s.LastVehicle, s.VehicleSerial, s.vehicleShapeExpected = proto.VehicleResult{}, 0, false
			s.LastGatekeeper, s.GatekeeperSerial = proto.GatekeeperResult{}, 0
			s.LastCompose, s.ComposeSerial = proto.ComposeResult{}, 0
			s.LastEnchant, s.EnchantSerial = proto.EnchantResult{}, 0
			s.LastGemAdd, s.GemAddSerial = proto.GemAddResult{}, 0
			s.LastGemRemove, s.GemRemoveSerial = proto.GemRemoveResult{}, 0
			s.LastEmblem, s.EmblemSerial, s.EmblemRemoveSerial = proto.EmblemResult{}, 0, 0
			s.InventoryReady = false
			s.Warehouse, s.WarehouseReady = nil, false
			s.Skills = map[uint8]proto.LearnedSkill{}
			s.QuickSlots = nil
			s.SkillBarReady = false
			s.MapQuestNPCs, s.MapQuestNPCsReady = nil, false
			s.questNPCStaging, s.questNPCStagingActive = nil, false
			s.resetQuestMobs()
			s.GroundItems = map[uint32]proto.GroundItem{}
			s.Entities = map[uint32]*Entity{}
			s.MyShop, s.ShopView, s.ShopSigns = MyShopState{}, ShopViewState{}, nil
		}
		ev.Text = fmt.Sprintf("select result=%d id=%d", r.Result, r.ID)
	case proto.CharacterDetails:
		d, err := proto.DecodeDetails(p)
		if err != nil {
			return fail(err)
		}
		s.Details = d
		s.Guild.Name = d.Guild
		s.Self.Kind, s.Self.ID = KindUser, s.SelfID
		s.Self.X, s.Self.Y, s.Self.Z, s.Self.Dir = d.X, d.Y, d.Z, d.Dir
		ev.Text = fmt.Sprintf("details hp=%d pos=(%.1f,%.1f,%.1f) money=%d", d.MaxHP, d.X, d.Y, d.Z, d.Money)
	case proto.CharacterMapTeleport:
		m, err := proto.DecodeMapPosition(p)
		if err != nil {
			return fail(err)
		}
		if m.ID == s.SelfID {
			if s.MapID != m.MapID {
				ev.MapLoadAck = s.Phase == PhaseInWorld
				s.Weather = proto.Weather{}
				s.Relics = nil // a map without relics sends no 0x2101
				s.MapQuestNPCs, s.MapQuestNPCsReady = nil, false
				s.questNPCStaging, s.questNPCStagingActive = nil, false
				s.resetQuestMobs()
				s.Guild.RankingPointsKnown = false
				s.GroundItems = map[uint32]proto.GroundItem{}
				s.Entities = map[uint32]*Entity{}
				for id := range s.PlayerForms {
					if id != s.SelfID {
						delete(s.PlayerForms, id)
					}
				}
				for id := range s.Transforms {
					if id != s.SelfID {
						delete(s.Transforms, id)
					}
				}
				s.MyShop, s.ShopView, s.ShopSigns = MyShopState{}, ShopViewState{}, nil
				s.Self.ShopTitle = ""
				s.Target = 0
			}
			s.Self.Sitting = false
			s.MapID, s.Self.X, s.Self.Y, s.Self.Z = m.MapID, m.X, m.Y, m.Z
		}
		ev.Text = fmt.Sprintf("map %d pos=(%.1f,%.1f,%.1f)", m.MapID, m.X, m.Y, m.Z)
	case proto.GXSkillDash:
		d, err := proto.DecodeSkillDash(p)
		if err != nil {
			return fail(err)
		}
		s.LastSkillDash = d
	case proto.GXPlayerCorrection:
		correction, err := proto.DecodePlayerCorrection(p)
		if err != nil {
			return fail(err)
		}
		if correction.ID == s.SelfID {
			s.Self.X, s.Self.Y, s.Self.Z = correction.X, correction.Y, correction.Z
			s.Self.Motion, s.Self.Dir = correction.Motion, correction.Direction
			s.LastPlayerCorrection = correction
		}
	case proto.GXPhysicsClock:
		if len(p) != 12 {
			return fail(fmt.Errorf("invalid physics clock"))
		}
		clock := math.Float64frombits(binary.LittleEndian.Uint64(p[4:12]))
		if math.IsNaN(clock) || math.IsInf(clock, 0) || clock < 0 {
			return fail(fmt.Errorf("invalid physics time"))
		}
		s.PhysicsMapID = binary.LittleEndian.Uint16(p[2:4])
		s.PhysicsTime = clock
	case proto.GXVisibility:
		if len(p) != 7 || p[6] > 1 {
			return fail(fmt.Errorf("invalid GX visibility"))
		}
		id := binary.LittleEndian.Uint32(p[2:6])
		if id == s.SelfID {
			s.Self.Concealed = p[6] != 0
		} else if e := s.Entities[id]; e != nil {
			e.Concealed = p[6] != 0
		}
	case proto.GXGMModes:
		m, err := proto.DecodeGMModes(p)
		if err != nil {
			return fail(err)
		}
		if m.ID == s.SelfID && s.Self.IsGM {
			s.GMConcealed, s.GMImmune, s.GMTurbo = m.Concealed, m.Immune, m.Turbo
		}
	case proto.GXGMState:
		g, err := proto.DecodeGMState(p)
		if err != nil {
			return fail(err)
		}
		if g.ID == s.SelfID {
			s.Self.IsGM, s.GMTurbo = g.GM, g.GM && g.Turbo
			if !g.GM {
				s.GMConcealed, s.GMImmune = false, false
			}
			if g.Message != "" {
				s.addChat(ChatLine{Op: proto.ChatError, Text: g.Message, Style: 4})
			}
		}
	case proto.CharacterEnteredMap, proto.UserLeave, proto.MobLeave, proto.MapNpcLeave, proto.ChangeAppearance, proto.CharacterEquipment, proto.CharacterShape, proto.CharacterMove, proto.MobEnter, proto.GXMobEnter, proto.MobMove, proto.GXMobMove, proto.MapNpcEnter, proto.MapNpcMove, proto.MapNpcAttackPlayer, proto.MapNpcAttackMob:
		return s.applyEntitiesPacket(p, ev)
	case proto.MobSpeechText, proto.MobSpeechIndex, proto.ChatNormal, proto.ChatParty, proto.ChatShout, proto.ChatNormalAdmin, proto.ChatWorld, proto.ChatGuild, proto.ChatMap, proto.ChatRaid, proto.ChatWorldShout, proto.ChatWorldAdmin, proto.ChatWhisper, proto.ChatError:
		return s.applyChatPacket(p, ev)
	case proto.Logout:
		s.Party = PartyState{}
		s.Guild = GuildState{}
		s.Relics = nil
		s.LastRelicNotice = proto.RelicDestroyed{}
		s.LastPortalNotice = proto.PortalOwnerNotice{}
		s.GuildDuel = GuildDuelState{}
		s.Phase, s.SelfID = PhaseCharSelect, 0
		s.PlayerForms, s.Transforms, s.Evolution = nil, nil, proto.EvolutionBar{}
		s.Target = 0
		s.Entities = map[uint32]*Entity{}
		s.chatGuard.senders = nil
		s.MyShop, s.ShopView, s.ShopSigns = MyShopState{}, ShopViewState{}, nil
		ev.Text = "back to character select"
	default:
		s.Unknown[op]++
	}
	return ev, nil
}

func (e Event) packetError(err error) (Event, error) {
	return e, fmt.Errorf("0x%04X: %w", uint16(e.Op), err)
}

func (s *State) addChat(l ChatLine) { s.AddChat(l) }

// AddChat retains the most recent 100 entries independently in each pane.
func (s *State) AddChat(l ChatLine) {
	s.resolveChatSender(&l)
	if !s.blockedChat(l) {
		s.appendChat(l)
	}
}

func (s *State) resolveChatSender(l *ChatLine) {
	if l.Sender == "" && l.SenderID != 0 {
		if e := s.Entities[l.SenderID]; e != nil && e.Shape != nil {
			l.Sender = e.Shape.Name
		}
		if l.Sender == "" {
			for _, slot := range s.Slots {
				if slot.ID == l.SenderID {
					l.Sender = slot.Name
					break
				}
			}
		}
	}
}

func (s *State) appendChat(l ChatLine) {
	if len(s.Chat) >= 100 {
		system := l.SystemPane()
		remaining := 100
		for i := len(s.Chat) - 1; i >= 0; i-- {
			if s.Chat[i].SystemPane() == system {
				remaining--
				if remaining == 0 {
					s.Chat = slices.Delete(s.Chat, i, i+1)
					break
				}
			}
		}
	}
	s.Chat = append(s.Chat, l)
}

func (s *State) applyItem(it proto.InventoryItem) {
	if it.Bag == proto.GuildBankBag {
		if s.Guild.Bank == nil {
			s.Guild.Bank = make(map[uint8]proto.InventoryItem)
		}
		if it.Type == 0 || it.Count == 0 {
			delete(s.Guild.Bank, it.Slot)
		} else {
			s.Guild.Bank[it.Slot] = it
		}
		return
	}
	if it.Bag == proto.WarehouseBag {
		if s.Warehouse == nil {
			s.Warehouse = map[uint8]proto.InventoryItem{}
		}
		if it.Type == 0 || it.Count == 0 {
			delete(s.Warehouse, it.Slot)
		} else {
			s.Warehouse[it.Slot] = it
		}
		return
	}
	key := [2]uint8{it.Bag, it.Slot}
	if it.Type == 0 || it.Count == 0 {
		delete(s.Inventory, key)
	} else {
		s.Inventory[key] = it
	}
}
