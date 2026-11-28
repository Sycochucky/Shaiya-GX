package world

import (
	"github.com/Spelunkern/Shaiya-GX/server/internal/netx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/telemetry"
	"net"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/crypto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
)

// connWriter serializes every write to one client socket: the owning
// session and broadcasts from other sessions share it.
type connWriter struct {
	mu     sync.Mutex
	conn   net.Conn
	queued *netx.Writer
}

func (w *connWriter) write(b []byte) error {
	if w.queued != nil {
		return w.queued.Write(b)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := w.conn.Write(b)
	return err
}

type pendingTeleport struct {
	MapID   uint16
	X, Y, Z float32
}

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

// ConnectedPlayer is a player in the world, reachable from any session for
// broadcasts. Cross-session work is queued in mailboxes and applied by the
// owning session's loop. Paired trades hold both owners' state locks.
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

	mailMu                     sync.Mutex
	teleport                   *pendingTeleport
	summonOffer                *partySummonOffer
	forced                     *forcedTeleport // a GM's /asummon waiting for this user's session
	tradeRefresh               *tradeRefresh
	damage                     []pendingDamage
	lifeSkill                  *pendingLifeSkill
	soulKill                   *pendingSoulKill
	friendly                   []pendingFriendly
	userHits                   []pendingUserHit
	combatMailClosed           bool
	pvpKillCount, pvpRankCount uint32
	duelEnds                   []pendingDuelEnd
	duelWins                   int
	itemGrants                 []db.Item

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

// QueueQuestKill credits a kill to the player's quests (party members get
// it too, as mob_death does).
func (p *ConnectedPlayer) QueueQuestKill(mobID uint16) {
	// Only active quests can consume a kill; each has two uint8 counters.
	relevant := false
	for _, id := range mobQuestLinks[mobID] {
		if p.hasQuest(id) {
			relevant = true
			break
		}
	}
	if !relevant {
		return
	}
	p.rewardMu.Lock()
	defer p.rewardMu.Unlock()
	if p.questKills == nil {
		p.questKills = map[uint16]uint8{}
	}
	if p.questKills[mobID] < 255 {
		p.questKills[mobID]++
	}
}
func (p *ConnectedPlayer) pendingQuestKills() map[uint16]uint8 {
	p.rewardMu.Lock()
	defer p.rewardMu.Unlock()
	result := make(map[uint16]uint8, len(p.questKills))
	for id, count := range p.questKills {
		result[id] = count
	}
	return result
}
func (p *ConnectedPlayer) ackQuestKills(id uint16, count uint8) {
	p.rewardMu.Lock()
	defer p.rewardMu.Unlock()
	if p.questKills[id] <= count {
		delete(p.questKills, id)
	} else {
		p.questKills[id] -= count
	}
}

func (p *ConnectedPlayer) setQuests(ids []uint16) {
	m := make(map[uint16]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	p.questMu.Lock()
	p.questIDs = m
	p.questMu.Unlock()
}

func (p *ConnectedPlayer) hasQuest(id uint16) bool {
	p.questMu.Lock()
	defer p.questMu.Unlock()
	return p.questIDs[id]
}

type perfectExperience struct {
	before experienceBatch
	amount uint32
}

type rewardSnapshot struct {
	experience experienceBatch
	perfect    []perfectExperience
	gold       uint32
}

// An award's integer hundreds and remainder can be aggregated independently.
// This keeps ordinary rewards bounded while retaining rounding per kill for any bonus.
type experienceBatch struct {
	hundreds   uint64
	remainders [100]uint64
}

func experienceScaled(amount uint32, bonus uint32) uint32 {
	return uint32(min(uint64(^uint32(0)), uint64(amount)+uint64(amount)*uint64(bonus)/100))
}

func (b *experienceBatch) add(amount uint32) {
	b.hundreds += uint64(amount / 100)
	if remainder := amount % 100; remainder != 0 {
		b.remainders[remainder]++
	}
}

func (b *experienceBatch) scaled(bonus uint32) uint32 {
	const limit = uint64(^uint32(0))
	factor := 100 + uint64(bonus)
	if b.hundreds > limit/factor {
		return uint32(limit)
	}
	total := b.hundreds * factor
	for remainder, count := range b.remainders {
		value := uint64(remainder) * factor / 100
		if value != 0 && count > (limit-total)/value {
			return uint32(limit)
		}
		total += count * value
	}
	return uint32(total)
}

func (b *experienceBatch) subtract(credited experienceBatch) {
	b.hundreds -= credited.hundreds
	for remainder, count := range credited.remainders {
		b.remainders[remainder] -= count
	}
}

// QueueExp carries a kill reward to its recipient's owning session.
func (p *ConnectedPlayer) QueueExp(amount uint32, perfect bool) {
	p.rewardMu.Lock()
	defer p.rewardMu.Unlock()
	if perfect {
		// Preserve award order for the Perfect Party level-up boundary.
		p.perfectExp = append(p.perfectExp, perfectExperience{p.experience, amount})
		p.experience = experienceBatch{}
		return
	}
	p.experience.add(amount)
}
func (p *ConnectedPlayer) QueueGold(amount uint32) {
	p.rewardMu.Lock()
	defer p.rewardMu.Unlock()
	p.gold = satAdd(p.gold, amount)
}
func (p *ConnectedPlayer) takeRewardTotals(bonus uint32) (uint32, uint32) {
	p.rewardMu.Lock()
	defer p.rewardMu.Unlock()
	exp, gold := p.experience.scaled(bonus), p.gold
	for _, event := range p.perfectExp {
		exp = satAdd(exp, satAdd(event.before.scaled(bonus), experienceScaled(event.amount, bonus)))
	}
	p.perfectExp = p.perfectExp[:0]
	p.experience = experienceBatch{}
	p.gold = 0
	return exp, gold
}
func (p *ConnectedPlayer) peekRewardTotals() rewardSnapshot {
	p.rewardMu.Lock()
	defer p.rewardMu.Unlock()
	return rewardSnapshot{p.experience, slices.Clone(p.perfectExp), p.gold}
}
func (p *ConnectedPlayer) ackRewardTotals(snapshot rewardSnapshot) {
	p.rewardMu.Lock()
	defer p.rewardMu.Unlock()
	remaining := copy(p.perfectExp, p.perfectExp[len(snapshot.perfect):])
	p.perfectExp = p.perfectExp[:remaining]
	tail := &p.experience
	if remaining != 0 {
		// A new perfect award moved the old trailing EXP into its prefix.
		tail = &p.perfectExp[0].before
	}
	tail.subtract(snapshot.experience)
	p.gold = satSub(p.gold, snapshot.gold)
}
func (p *ConnectedPlayer) TakeRewards() (exp, gold []uint32) {
	e, g := p.takeRewardTotals(0)
	if e > 0 {
		exp = []uint32{e}
	}
	if g > 0 {
		gold = []uint32{g}
	}
	return
}

const maxPendingCombat = 4096
const maxPendingItemGrants = 256

var mailboxOverflows = telemetry.NumberMetric("gx_player_mailbox_overflows_total")

func (p *ConnectedPlayer) mailboxOverflow() {
	mailboxOverflows.Add(1)
	// Bound combat backlog by terminating an unresponsive connection. Rewards
	// use fixed accumulators and are flushed separately on session termination.
	if p.out != nil && p.out.conn != nil {
		netx.Abort(p.out.conn)
	}
}

func bucket(mapID uint16) map[uint32]*ConnectedPlayer {
	b, ok := realm.registry[mapID]
	if !ok {
		b = map[uint32]*ConnectedPlayer{}
		realm.registry[mapID] = b
	}
	return b
}

// RegisterPlayer adds a player to its map's realm.registry.
func RegisterPlayer(ch db.Character, items []db.Item, isAdmin bool, faction uint8, out *connWriter, log *logx.Logger, owner ...*Session) *ConnectedPlayer {
	p := &ConnectedPlayer{CharacterID: ch.ID, CharacterName: ch.Name, IsAdmin: isAdmin, out: out, log: log,
		gmVisible: true, mapID: ch.MapID, character: ch, items: append([]db.Item(nil), items...), faction: faction}
	if len(owner) > 0 {
		p.owner = owner[0]
		if p.owner != nil {
			p.owner.syncQuestIDs(p)
		}
	}
	p.lifeState.Store(1)
	realm.registryMu.Lock()
	bucket(ch.MapID)[ch.ID] = p
	realm.registryMu.Unlock()
	return p
}

// Unregister removes the player and cancels any trade it was part of.
func (p *ConnectedPlayer) Unregister() {
	clearVehicle(p)

	clearSocialSession(p.CharacterID)
	p.dropFromOthers()
	realm.registryMu.Lock()
	delete(bucket(p.MapID()), p.CharacterID)
	realm.registryMu.Unlock()
	p.closeUserHits()
	tradeCancelForDisconnect(p.CharacterID, p.log)
	duelLeave(p, 0, false) // PvPChkLeaveByZoneLeave
	guildDuelLeave(p, nil, 1)
	partyLeaveOnDisconnect(p)
}

func (p *ConnectedPlayer) SetMap(mapID uint16) {
	p.mu.Lock()
	old := p.mapID
	p.mapID = mapID
	if old != mapID {
		p.assistTarget = 0
	}
	p.mu.Unlock()
	if old == mapID {
		return
	}
	realm.registryMu.Lock()
	if cur, ok := bucket(old)[p.CharacterID]; ok {
		delete(bucket(old), p.CharacterID)
		bucket(mapID)[p.CharacterID] = cur
	}
	realm.registryMu.Unlock()
}

func (p *ConnectedPlayer) MapID() uint16 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.mapID
}

func (p *ConnectedPlayer) Faction() uint8 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.faction
}

func (p *ConnectedPlayer) PositionXZ() (float32, float32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.character.PosX, p.character.PosZ
}

func (p *ConnectedPlayer) UpdateCharacter(ch db.Character) {
	p.mu.Lock()
	p.character = ch
	p.mu.Unlock()
}

func (p *ConnectedPlayer) Snapshot() db.Character {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.character
}

func (p *ConnectedPlayer) UpdateItems(items []db.Item) {
	p.mu.Lock()
	p.items = append([]db.Item(nil), items...)
	p.mu.Unlock()
}

func (p *ConnectedPlayer) UpdateMotion(motion uint8, direction uint16) {
	p.mu.Lock()
	p.motion, p.direction = motion, direction
	p.mu.Unlock()
}

func (p *ConnectedPlayer) SetSitting(v bool) {
	p.mu.Lock()
	p.sitting = v
	p.mu.Unlock()
}

func (p *ConnectedPlayer) IsSitting() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.sitting
}

func (p *ConnectedPlayer) GMVisible() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.gmVisible
}

func (p *ConnectedPlayer) IsAlive() bool { return p.lifeState.Load()&1 != 0 }

// SetAlive is called by the owning session. A death has a distinct state
// even if the same character revives and dies again during a prepared cast.
func (p *ConnectedPlayer) SetAlive(alive bool) {
	if state := p.lifeState.Load(); (state&1 != 0) != alive {
		if alive {
			p.SetSitting(false)
		}
		p.lifeState.Store(state + 1)
	}
}

func (p *ConnectedPlayer) QueueTeleport(mapID uint16, x, y, z float32) {
	p.mailMu.Lock()
	p.teleport = &pendingTeleport{mapID, x, y, z}
	p.mailMu.Unlock()
}

func (p *ConnectedPlayer) TakeTeleport() *pendingTeleport {
	p.mailMu.Lock()
	defer p.mailMu.Unlock()
	t := p.teleport
	p.teleport = nil
	return t
}

// tradeRefresh tells the other side of a completed trade to reload its
// items and apply the money change itself (only the owner mutates its
// character).
type tradeRefresh struct {
	Packets    [][]byte
	MoneyDelta int64
}

func (p *ConnectedPlayer) QueueTradeRefresh(packets [][]byte, moneyDelta int64) {
	p.mailMu.Lock()
	p.tradeRefresh = &tradeRefresh{packets, moneyDelta}
	p.mailMu.Unlock()
}

func (p *ConnectedPlayer) TakeTradeRefresh() *tradeRefresh {
	p.mailMu.Lock()
	defer p.mailMu.Unlock()
	r := p.tradeRefresh
	p.tradeRefresh = nil
	return r
}

func (p *ConnectedPlayer) QueueDamage(d pendingDamage) {
	p.mailMu.Lock()
	if len(p.damage) >= maxPendingCombat {
		p.mailMu.Unlock()
		p.mailboxOverflow()
		return
	}
	p.damage = append(p.damage, d)
	p.mailMu.Unlock()
}

func (p *ConnectedPlayer) TakeDamage() []pendingDamage {
	p.mailMu.Lock()
	defer p.mailMu.Unlock()
	d := p.damage
	p.damage = nil
	return d
}

func (p *ConnectedPlayer) SitMotionPacket() []byte {
	return buildCharacterMotion(p.CharacterID, MotionSitDown)
}

// EnteredMapPacket introduces an already-present character to a viewer.
func (p *ConnectedPlayer) EnteredMapPacket() []byte {
	gid := guildIDFor(p.CharacterID)
	p.mu.Lock()
	defer p.mu.Unlock()
	return buildCharacterEnteredMapGuild(&p.character, p.IsAdmin, false, gid)
}

// ShapePacket never embeds sitting; sit is sent as a separate 0x0506.
func (p *ConnectedPlayer) ShapePacket() []byte {
	name := guildNameFor(p.CharacterID)
	p.mu.Lock()
	defer p.mu.Unlock()
	return buildCharacterShapeGuild(&p.character, p.faction, p.items, 0, name)
}

func (p *ConnectedPlayer) MovePacket() []byte {
	p.mu.Lock()
	defer p.mu.Unlock()
	return buildCharacterMove(&p.character, p.motion, p.direction)
}

func (p *ConnectedPlayer) EquipmentPacket(slot uint8) []byte {
	p.mu.Lock()
	defer p.mu.Unlock()
	return buildCharacterEquipment(&p.character, p.items, slot)
}

// Send writes a frame to this player with the post-handshake expanded XOR
// encoding, which is stateless and safe from any goroutine.
func (p *ConnectedPlayer) Send(frame []byte, label string) error {
	if !p.canReceiveWorldPacket(frame) {
		return nil
	}
	frame = p.redactPartyPositions(frame)
	return p.out.write(appendWorldFrame(nil, frame))
}

func appendWorldFrame(dst, frame []byte) []byte {
	if len(frame) <= 2 {
		return append(dst, frame...)
	}
	dst = append(dst, frame[:2]...)
	return append(dst, crypto.ExpandedXOR(frame[2:])...)
}

func playersWhere(mapID uint16, keep func(*ConnectedPlayer) bool) []*ConnectedPlayer {
	realm.registryMu.Lock()
	candidates := make([]*ConnectedPlayer, 0, len(realm.registry[mapID]))
	for _, p := range realm.registry[mapID] {
		candidates = append(candidates, p)
	}
	realm.registryMu.Unlock()
	out := candidates[:0]
	for _, p := range candidates {
		if keep(p) {
			out = append(out, p)
		}
	}
	return out
}

func PlayersOnMap(mapID uint16) []*ConnectedPlayer {
	return playersWhere(mapID, func(*ConnectedPlayer) bool { return true })
}

func PlayersNear(mapID uint16, x, z, radius float32) []*ConnectedPlayer {
	return playersWhere(mapID, func(p *ConnectedPlayer) bool {
		px, pz := p.PositionXZ()
		return DistanceXZ(px, pz, x, z) <= radius
	})
}

func PlayersNearExcluding(mapID uint16, characterID uint32, x, z, radius float32) []*ConnectedPlayer {
	return playersWhere(mapID, func(p *ConnectedPlayer) bool {
		if p.CharacterID == characterID {
			return false
		}
		px, pz := p.PositionXZ()
		return DistanceXZ(px, pz, x, z) <= radius
	})
}

// PlayerCount reports characters registered in the world.
func PlayerCount() int {
	realm.registryMu.Lock()
	defer realm.registryMu.Unlock()
	count := 0
	for _, players := range realm.registry {
		count += len(players)
	}
	return count
}

func AllPlayers() []*ConnectedPlayer {
	realm.registryMu.Lock()
	defer realm.registryMu.Unlock()
	var out []*ConnectedPlayer
	for _, b := range realm.registry {
		for _, p := range b {
			out = append(out, p)
		}
	}
	return out
}

func FindPlayerByName(name string) *ConnectedPlayer {
	for _, p := range AllPlayers() {
		if strings.EqualFold(p.CharacterName, name) {
			return p
		}
	}
	return nil
}

func FindPlayerByID(id uint32) *ConnectedPlayer {
	for _, p := range AllPlayers() {
		if p.CharacterID == id {
			return p
		}
	}
	return nil
}

// Hidden GMs are visible only to other GMs.
func canSeeCharacter(viewerIsAdmin, subjectIsAdmin, subjectGMVisible bool) bool {
	return !subjectIsAdmin || subjectGMVisible || viewerIsAdmin
}

// broadcast sends frame to each player, logging failures.
func broadcast(players []*ConnectedPlayer, frame []byte, label string, log *logx.Logger) {
	for _, p := range players {
		if err := p.Send(frame, label); err != nil && log != nil {
			log.Printf("GX Server World failed to deliver %s to character_id=%d: %v", label, p.CharacterID, err)
		}
	}
}

// UpdateElements publishes the owning session's native event state for
// detection; no other session reads or mutates its applied skill list.
func (p *ConnectedPlayer) UpdateElements(attack, defense uint8) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.elementKnown = true
	p.attackElement, p.defenseElement = attack, defense
}

func (p *ConnectedPlayer) QueueTradeCommit(trade *tradeSession) {
	p.mailMu.Lock()
	p.tradeCommit = trade
	p.mailMu.Unlock()
}
func (p *ConnectedPlayer) takeTradeCommit() *tradeSession {
	p.mailMu.Lock()
	defer p.mailMu.Unlock()
	trade := p.tradeCommit
	p.tradeCommit = nil
	return trade
}
