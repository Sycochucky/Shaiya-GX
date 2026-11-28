package world

import (
	"errors"
	"fmt"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/skilltiming"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/crypto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/netx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
	"github.com/Spelunkern/Shaiya-GX/server/internal/session"
)

const (
	loopTick          = 100 * time.Millisecond
	recoveryInterval  = 5 * time.Second
	characterListSize = MaxCharacterSlots
)

// Session is one client connection. Its loop owns its fields under stateMu.
// A paired trade may also update them while holding both sessions' state locks.
type Session struct {
	economyConfirmedAmount *uint32
	gmForced               bool      // a GM's /asummon is moving this user: no party or guild instance redirection
	assisted               [3]uint32 // HP, SP, MP of the last hit after Health Assistant
	targetHealthSent       [3]uint32 // user, HP and maximum last told for the selected user
	moveRune               *movementRuneCast
	teleportRune           *teleportRuneCast
	traps                  []placedTrap
	nextTrapAt             time.Time
	economy                *netx.RateLimiter
	awaitingMapLoad        bool
	inputReads             *netx.RateLimiter
	lastInputTick          time.Time
	persistenceFailed      bool
	operation              *inventoryOperation
	inputFrames            *netx.RateLimiter
	inputBytes             *netx.RateLimiter
	stateMu                sync.Mutex
	shop                   myShop // the personal market this character sets up
	effectWatchID          uint32
	effectWatchKind        uint8
	effectWatchAt          time.Time
	effectWatchHash        [32]byte
	speedKnown             bool
	lastSpeeds             playerSpeeds
	movement               playerMovement
	oxygen                 oxygenState
	questNPCCatalog        questNPCCatalogState
	questMobCatalog        questMobCatalogState
	gmTurbo                bool
	gmImmune               bool
	gmConcealed            bool
	lastGMModeAt           time.Time
	elements               elementRuntime
	socialChecked          time.Time
	socialReloadAt         time.Time
	socialFriends          []db.SocialContact
	socialRefresh          atomic.Bool
	conn                   net.Conn
	out                    *connWriter
	log                    *logx.Tag
	logger                 *logx.Logger

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

func (s *Session) acquireAccount(uid uint32, id int64) bool {
	realm.activeAccountsMu.Lock()
	if realm.activeAccounts[uid] != nil {
		realm.activeAccountsMu.Unlock()
		return false
	}
	if s.conn != nil {
		banned, err := db.OperatorIPBanned(netx.PeerIP(s.conn))
		if err != nil || banned {
			realm.activeAccountsMu.Unlock()
			storageFailure(err)
			return false
		}
	}
	realm.activeAccounts[uid] = s
	realm.activeAccountsMu.Unlock()
	ok, err := db.AcquireAccount(uid, id)
	if err != nil || !ok {
		realm.activeAccountsMu.Lock()
		delete(realm.activeAccounts, uid)
		realm.activeAccountsMu.Unlock()
		storageFailure(err)
		return false
	}
	return true
}
func releaseAccount(uid uint32, id int64) {
	if err := db.ReleaseAccount(uid, id); err != nil {
		storageFailure(err)
		db.RetryAccountRelease(uid, id)
	}
	realm.activeAccountsMu.Lock()
	delete(realm.activeAccounts, uid)
	realm.activeAccountsMu.Unlock()
}

// Serve runs one client connection until it closes.
func Serve(conn net.Conn, logger *logx.Logger, log *logx.Tag) error {
	policy, err := config.Network("world")
	if err != nil {
		return err
	}
	s := &Session{economy: netx.NewRateLimiter(1, policy.Economy, time.Second), inputReads: netx.NewRateLimiter(1, policy.InputFrames*4, time.Second), inputFrames: netx.NewRateLimiter(1, policy.InputFrames, time.Second), inputBytes: netx.NewRateLimiter(1, policy.InputBytes, time.Second), conn: conn, out: &connWriter{conn: conn, queued: netx.NewWriter(conn, policy.OutputPackets, policy.OutputBytes)}, log: log, logger: logger,
		faction:       FactionNone,
		itemCooldowns: map[uint8]time.Time{}, skillCooldowns: map[uint16]time.Time{},
		lastPositionSave: time.Now(), lastRegenAt: time.Now()}
	defer s.out.queued.Close()
	defer s.close()
	log.Printf("GX Server World client connected")
	handshakeDeadline := time.Now().Add(30 * time.Second)
	var frames netx.Frames
	buf := make([]byte, 4096)
	for {
		if s.aesRecv == nil && time.Now().After(handshakeDeadline) {
			return fmt.Errorf("world handshake timed out")
		}
		s.stateMu.Lock()
		wake := s.nextWake()
		s.stateMu.Unlock()
		conn.SetReadDeadline(time.Now().Add(wake))
		n, err := conn.Read(buf)
		var nerr net.Error
		timeout := err != nil && errors.As(err, &nerr) && nerr.Timeout()
		if err != nil && !timeout && !errors.Is(err, os.ErrDeadlineExceeded) {
			log.Printf("GX Server World client disconnected")
			return nil
		}
		if err := s.processInput(n, buf, &frames); err != nil {
			return err
		}
	}
}

// processInput never holds the session lock while waiting for socket input.
func (s *Session) processInput(n int, buf []byte, frames *netx.Frames) error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.accountID != 0 && time.Now().After(s.authorizationAt) {
		a, err := db.AccountByUID(int32(s.accountID))
		if err != nil || !db.CanPlay(int32(a.Status)) || a.Admin != s.isAdmin || a.CredentialVersion != s.credentialVersion {
			return fmt.Errorf("session authorization revoked")
		}
		s.authorizationAt = time.Now().Add(30 * time.Second)
	}
	if n == 0 {
		s.tick(true)
		s.lastInputTick = time.Now()
		return nil
	}
	if s.inputReads != nil && !s.inputReads.Allow("") {
		return fmt.Errorf("world fragmented input budget exceeded")
	}
	if s.inputBytes != nil && !s.inputBytes.AllowN("", n) {
		return fmt.Errorf("world input byte budget exceeded")
	}
	list, ok := frames.Push(buf[:n])
	if !ok {
		return fmt.Errorf("malformed world frame")
	}
	if len(list) > 0 && s.inputFrames != nil && !s.inputFrames.AllowN("", len(list)) {
		return fmt.Errorf("world input packet budget exceeded")
	}
	// Complete commands remain responsive; partial TCP frames cannot trigger
	// expensive simulation work on every byte or starve the idle timeline.
	if len(list) > 0 || time.Since(s.lastInputTick) >= loopTick {
		s.tick(false)
		s.lastInputTick = time.Now()
	}
	for _, frame := range list {
		if err := s.frame(frame); err != nil {
			return err
		}
	}

	return nil
}

// close persists the active character and releases the account.
func (s *Session) close() {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	s.leaveWorld(true)
	if s.accountID != 0 {
		if err := db.TryGameLogout(int32(s.accountID), s.sessionID); err != nil {
			s.log.Printf("GX Server World logout procedure failed: %v", err)
		}
		releaseAccount(s.accountID, s.sessionID)
	}
}

// leaveWorld saves and unregisters the active character.
func (s *Session) leaveWorld(disconnect bool) bool {
	if s.char == nil {
		return true
	}
	if s.player != nil {
		s.endMyShop()
		s.closeMyShopView()
		s.drainRewards(s.player)
		s.tickPvP()
		s.applyDuelWins()
	}
	s.reviveOnExit(s.exitDeathLoss())
	if err := db.SaveCharacterExit(*s.char, s.buffSnapshot(), s.skillDelaySnapshot(), s.questSnapshot()); err != nil {
		s.logErr("save character exit", err)
		if !disconnect {
			return false
		}
	}
	s.useNpc(0)
	s.endGMGuildRanking(true)
	s.guildDisconnect()
	if s.player != nil {
		s.noteDungeonExit(s.char.MapID)
		s.activity("player.leave", s.char.Name+" left World", "disconnected", disconnect)
		s.player.Unregister()
		s.player = nil
	}
	s.char = nil
	return true
}

// send writes a frame to this client: AES-CTR until the initial character
// state, the stateless expanded XOR afterwards.
func (s *Session) send(frame []byte, label string) {
	if s.persistenceFailed {
		return
	}
	if s.operation != nil {
		s.operation.packet(frame, label, false)
		return
	}
	if s.aesSend == nil || len(frame) <= 2 {
		return
	}
	var body []byte
	if s.expanded {
		body = crypto.ExpandedXOR(frame[2:])
	} else {
		body = s.aesSend.Transform(frame[2:])
	}
	if err := s.out.write(append(append([]byte{}, frame[:2]...), body...)); err != nil {
		s.log.Printf("GX Server World failed to send %s: %v", label, err)
	}
}

func (s *Session) equipment() EquipmentStats {
	e := GetEquipmentStats(s.items)
	s.syncElements(e)
	applyPassives(&e, s.skills)
	e = addBuffs(e, s.buffs)
	applyKCStatus(&e, s.faction)
	e.AttackAttr, e.DefenseAttr = s.elements.attack, s.elements.defense
	for _, b := range s.buffs {
		if !b.MapCurse {
			continue
		}
		switch {
		case b.Effect == typeEffectBuff && b.Detail == 94 && b.Attr < 9:
			e.AttackAttr = b.Attr
		case b.Effect == typeEffectBuff && b.Detail == 95:
			e.DefenseAttr = b.Attr
		case b.Effect == typeEffectDebuff && b.Detail == 68:
			e.DefenseAttr = 0
		}
	}
	if s.player != nil {
		s.player.UpdateElements(e.AttackAttr, e.DefenseAttr)
	}
	return e
}
func (s *Session) stats() CharacterStats {
	st := StatsFor(s.char, s.equipment())
	if s.char != nil {
		s.char.CurrentHP, s.char.CurrentMP, s.char.CurrentSP = st.CurrentHP, st.CurrentMP, st.CurrentSP
	}
	return st
}

func (s *Session) sendStatPackets() {
	if s.char == nil {
		return
	}
	s.syncPlayerSpeed(false)
	st := s.stats()
	for _, p := range statPackets(s.char, &st) {
		s.send(p.Packet, p.Label)
	}
}

// drainRewards applies bounded, coalesced rewards before normal simulation or
// logout. Aggregation keeps individual scaling/rounding and caps unchanged.
func (s *Session) drainRewards(p *ConnectedPlayer) {
	s.applyLoot(p)
	for mob, count := range p.pendingQuestKills() {
		if s.questKillCount(mob, count) {
			p.ackQuestKills(mob, count)
		}
	}
	snapshot := p.peekRewardTotals()
	bonus, gold := s.experienceBonus(), snapshot.gold
	trailingExperience := snapshot.experience.scaled(bonus)
	if len(snapshot.perfect) == 0 && trailingExperience == 0 && gold == 0 {
		return
	}
	before := *s.char
	after := before
	var credited experienceCredit
	for _, event := range snapshot.perfect {
		_, ordinary := applyExpGain(&after, event.before.scaled(bonus))
		award := applyExperienceAward(&after, experienceScaled(event.amount, bonus), true)
		credited.base += ordinary + award.base
		credited.perfect += award.perfect
	}
	_, ordinary := applyExpGain(&after, trailingExperience)
	credited.base += ordinary
	levels := after.Level - before.Level
	if credited.base == 0 && credited.perfect == 0 && gold == 0 {
		p.ackRewardTotals(snapshot)
		return
	}
	if levels > 0 {
		full := StatsFor(&after, s.equipment())
		after.CurrentHP, after.CurrentMP, after.CurrentSP = full.MaxHP, full.MaxMP, full.MaxSP
	}
	after.Money = uint32(min(uint64(after.Money)+uint64(gold), db.MoneyCap))
	change := db.CharacterChange{Before: before, After: after, ItemsBefore: s.items, ItemsAfter: s.items}
	if err := db.CommitGameplay([]db.CharacterChange{change}, nil); err != nil {
		s.logErr("reward transaction", err)
		return
	}
	*s.char = after
	p.ackRewardTotals(snapshot)
	s.syncPlayer()
	s.partyPools()
	s.notifyExpGain(credited, levels)
	s.goldActivity(before.Money, "party loot")
	if gold > 0 {
		s.send(buildSetMoney(s.char.Money), "0x0213 party gold share")
	}
}

// normalize derives the base pools from StatusDefs (Chars only stores the
// current HP/MP/SP). Effective clamping waits for equipment and buffs.
func normalize(ch *db.Character) {
	ch.MaxHP, ch.MaxMP, ch.MaxSP = basePools(ch)
	// Equipped/buff maxima are not available yet; clamp after they are loaded.
}

// tick runs the per-iteration work: mailboxes, buffs, regen, casts and
// the auto-attack swing.
func (s *Session) tick(idle bool) {
	if s.char == nil {
		return
	}
	s.finishExpiredDungeon()
	if s.auctionContact && !s.auctionNPC(s.usingNpc) {
		s.useNpc(0)
	}
	if p := s.player; p != nil {
		if t := p.TakeTeleport(); t != nil {
			s.applyTeleport(t.MapID, t.X, t.Y, t.Z, "queued teleport")
		}
		if r := p.TakeTradeRefresh(); r != nil {
			s.applyTradeRefresh(r)
		}
		if trade := p.takeTradeCommit(); trade != nil {
			if other := FindPlayerByID(trade.partner(p.CharacterID)); other != nil {
				s.finalizeTrade(trade, other)
			}
		}

		s.drainRewards(p)
		for _, d := range p.TakeDamage() {
			if s.dead {
				break
			}
			if s.applyIncomingMobDamage(d) {
				s.dead, s.autoAttack, s.cast = true, false, nil
				s.lastActionAt = time.Time{}
				s.dash = nil
				s.pendingCombo = nil
				s.pendingSkill = nil
				s.lastSkillID = 0
				s.clearBuffsOnDeath()
			}
		}
	}
	s.processBuffs()
	s.tickPotentials(time.Now())
	s.tickFireThorn(time.Now())
	s.tickDetection(time.Now())
	s.tickTransform()
	s.tickTraps(time.Now())
	s.tickReveal()
	s.tickSkillOnce()
	s.tickFriendly()
	s.tickSocial()
	s.tickUserHits()
	s.tickPvP()
	s.tickDuel()
	s.tickMyShop()
	if s.player != nil && s.player.kcRefresh.Swap(false) {
		s.sendStatPackets() // ApplyKillCountAllUser -> UpdateKCStatus
	}
	s.tickGuildDuel()
	s.applyDuelWins()
	s.tickReturnToTown()
	s.tickTeleportStone(time.Now())
	s.tickTeleportRune(time.Now())
	s.tickPartySummon(time.Now())
	s.tickMovementRune(time.Now())
	s.tickForcedTeleport()
	s.completePendingGuild()
	s.tickRebirth(time.Now())
	if s.char == nil {
		return
	}
	s.updateWeather(time.Now(), false)
	s.syncPlayerSpeed(false)
	s.publishPlayerEffects()
	s.tickTargetEffects(time.Now())
	s.tickVehicle(time.Now())
	now := time.Now()
	s.tickPlayerMovement(now)
	s.tickOxygen(now)
	s.recoveryTick()
	s.syncQuestNPCCatalog(false)
	s.syncQuestMobCatalog(false)
	if time.Since(s.lastQuestCheck) >= 5*time.Second {
		s.lastQuestCheck = time.Now()
		s.questTimers()
	}
	s.processSpinCombo()
	s.processPendingSkill()
	s.processPendingCast()
	if idle && s.autoAttack {
		swing := s.autoAttackSwing
		if s.autoAttackUser {
			swing = s.userAttackSwing
		}
		if !swing(s.autoAttackTarget) {
			s.autoAttack = false
		}
	}
}

func (s *Session) applyTradeRefresh(r *tradeRefresh) {
	items, err := db.CharacterItems(s.char.ID)
	if err != nil {
		s.log.Printf("GX Server World trade refresh failed: %v", err)
		return
	}
	s.items = items
	s.char.Money = uint32(min(max(int64(s.char.Money)+r.MoneyDelta, 0), db.MoneyCap))
	s.persist()
	if s.player != nil {
		s.player.UpdateItems(s.items)
	}
	for _, p := range r.Packets {
		s.send(p, "0x0206/0x0205 trade refresh item")
	}
	s.send(buildSetMoney(s.char.Money), "0x0213 trade refresh money")
}

// recoveryTick is CUser::UpdateRecover every 5 seconds: the mode seen at
// the previous tick decides the amount (0 sitting: 5% of each maximum plus
// the sitting abilities' percent plus the flat sitting recovery; 1
// standing: the flat standing recovery; 2 in combat: the flat in-combat
// recovery), then the mode for the next tick is taken (in combat, else
// sitting or standing). Viewers get 0x0505 (absolute pools) and the character
// 0x0521.
func (s *Session) recoveryTick() {
	if time.Since(s.lastRegenAt) < recoveryInterval {
		return
	}
	s.lastRegenAt = time.Now()
	mode := s.regenMode
	switch {
	case s.inCombat():
		s.regenMode = 2
	case s.player != nil && s.player.IsSitting():
		s.regenMode = 0
	default:
		s.regenMode = 1
	}
	if s.dead {
		return
	}
	st := s.stats()
	if st.CurrentHP == 0 {
		return
	}
	var hp, mp, sp uint32
	r := st.Recover[mode]
	if mode == 0 {
		hp = sittingRecovery(st.MaxHP, st.SitRecover[0], r[0])
		sp = sittingRecovery(st.MaxSP, st.SitRecover[1], r[1])
		mp = sittingRecovery(st.MaxMP, st.SitRecover[2], r[2])
	} else {
		hp, sp, mp = r[0], r[1], r[2]
	}
	if s.oxygen.phase == proto.OxygenDraining && s.oxygen.remaining == 0 {
		hp = 0
	}
	hp, mp, sp = s.healDelta(hp, mp, sp)
	if hp+mp+sp == 0 {
		return
	}
	s.persist()
	after := s.stats()
	sendNear(s.char.MapID, s.char.PosX, s.char.PosZ, BroadcastRadius,
		buildCharacterReviveHitpoints(s.char.ID, after.CurrentHP, after.CurrentMP, after.CurrentSP), "0x0505 recover pools")
	s.send(buildCurrentHitpoints(&after), "0x0521 recover set")
}

// inCombat approximates the original attack mode: swinging, or a hit
// given or taken within the last recovery interval.
func (s *Session) inCombat() bool {
	return s.autoAttack || time.Since(s.lastCombatAt) < recoveryInterval
}

// deadGated lists the inputs ignored while dead (revive is 0x0553).
func deadGated(op packet.Opcode) bool {
	switch op {
	case packet.CharacterMove, packet.CharacterMobAutoAttack, packet.CharacterUserAutoAttack, packet.SkillUse, packet.SkillUseNoTarget,
		packet.UseItem, packet.ChangeAppearance, packet.UseVehicle:
		return true
	}
	return false
}

func (s *Session) frame(f []byte) error {
	if s.aesRecv == nil {
		op, _ := packet.U16(f[2:4])
		if packet.Opcode(op) == packet.GameHandshake && len(f) >= 24 {
			return s.handshake(f)
		}
		return fmt.Errorf("world handshake required")
	}
	p := s.aesRecv.Transform(f[2:])
	if len(p) < 2 {
		return nil
	}
	o, _ := packet.U16(p)
	op := packet.BaseOpcode(o)
	if op != packet.CharacterMove {
		s.log.Debugf("GX Server World opcode=0x%04X", o)
	}
	if s.dead && deadGated(op) {
		return nil
	}
	s.dispatch(op, p)
	return nil
}

func (s *Session) handshake(f []byte) error {
	uid, _ := packet.U32(f[4:8])
	var identity [16]byte
	copy(identity[:], f[8:24])
	h, err := session.Consume(uid, identity)
	if err != nil {
		s.log.Printf("GX Server World rejected handshake for user_uid=%d: %v", uid, err)
		return fmt.Errorf("invalid world login handoff")
	}
	if !s.acquireAccount(uid, h.SessionID) {
		s.log.Printf("GX Server World rejected user_uid=%d: already connected", uid)
		return fmt.Errorf("account already connected")
	}
	s.accountID, s.sessionID = uid, h.SessionID
	s.credentialVersion = h.CredentialVersion
	s.authorizationAt = time.Now().Add(30 * time.Second)
	if a, err := db.AccountByUID(int32(uid)); err == nil {
		if !db.CanPlay(int32(a.Status)) || a.CredentialVersion != h.CredentialVersion {
			return fmt.Errorf("account authorization changed")
		}
		s.account = a
		s.gmTurbo = false
		s.faction, s.isAdmin = a.Country, a.Admin
		if !isValidFaction(s.faction) {
			s.faction = FactionNone
		}
		s.log.SetTag(s.log.Tag() + "/" + a.UserID)
	} else {
		return fmt.Errorf("world account unavailable: %w", err)
	}
	var key, loginIV [16]byte
	copy(key[:], h.AESMaterial[:16])
	copy(loginIV[:], h.AESMaterial[16:])
	iv := crypto.DeriveWorldAESIV(loginIV)
	s.aesRecv, s.aesSend = crypto.NewAESCTR(key, iv), crypto.NewAESCTR(key, iv)
	s.log.Printf("GX Server World session ready user_uid=%d faction=%d admin=%t", uid, s.faction, s.isAdmin)
	s.send(buildWorldHandshake(), "0xA301 handshake ok")
	s.sendCharacterSelect()
	return nil
}

func (s *Session) sendCharacterSelect() {
	chars, err := db.CharactersForAccount(s.accountID)
	if err != nil {
		s.log.Printf("GX Server World failed to list characters: %v", err)
	}
	s.characters = chars
	if s.faction != FactionNone {
		for slot := uint8(0); slot < characterListSize; slot++ {
			frame := buildEmptyCharacterListSlot(slot)
			for i := range chars {
				if chars[i].Slot == slot {
					normalize(&chars[i])
					items, _ := db.CharacterItems(chars[i].ID)
					// The roster lists live characters before deleted ones.
					// A deleted replacement must never hide a restored slot.
					frame = buildCharacterListSlot(&chars[i], items)
					break
				}
			}
			s.send(frame, "0x0101 character slot")
		}
	}
	s.send(buildAccountFaction(s.faction), "0x0109 account faction")
}

func (s *Session) logErr(what string, err error) {
	if err != nil {
		s.log.Errorf("GX Server World %s failed: %v", what, err)
		storageFailure(err)
	}
}

func (s *Session) handleCreateCharacter(p []byte) {
	if len(p) < 10 {
		return
	}
	nc := db.NewCharacter{Slot: p[2], Race: p[3], Hair: p[5], Face: p[6], Height: p[7], Class: p[8],
		Gender: p[9], Name: packet.FixedText(p[10:])}
	s.createCharacter(nc)
}

func (s *Session) createCharacter(nc db.NewCharacter) {
	reject := func(why string) {
		s.log.Printf("GX Server World rejected create character name=%q slot=%d: %s", nc.Name, nc.Slot, why)
		s.send(buildCreateCharacterResult(false), "0x0102 create character rejected")
	}
	if nc.Slot >= MaxCharacterSlots {
		reject("slot out of range")
		return
	}
	// CUser::PacketCharacter: spaces are dropped and at least two
	// characters must remain; the family/job pair and mode must be valid.
	nc.Name = strings.ReplaceAll(nc.Name, " ", "")
	if len(nc.Name) < 2 || len(nc.Name) > 18 {
		reject("name too short")
		return
	}
	if !allowedJob(nc.Race, nc.Class) {
		reject("invalid race, class or mode")
		return
	}
	if _, used, err := db.CharacterIDInSlot(s.accountID, nc.Slot); err != nil || used {
		reject("slot occupied")
		return
	}
	if taken, err := db.NameTaken(nc.Name); err != nil || taken {
		reject("name in use")
		return
	}
	// Map.ini start position per race; CUser::ProcessCharacterCreate stats
	// and pools.
	start := mapConfig.Start[nc.Race]
	nc.MapID, nc.PosX, nc.PosY, nc.PosZ = start.MapID, start.X, start.Y, start.Z
	nc.Str, nc.Dex, nc.Rec, nc.Int, nc.Wis, nc.Luc = startingStats(nc.Race, nc.Class)
	nc.HP, nc.MP, nc.SP = newCharacterPools(nc.Race, nc.Class, nc.Rec, nc.Dex, nc.Wis)
	id, err := db.CreateCharacterWithItems(s.account, nc, starterItems(nc.Race, nc.Class))
	if err != nil {
		s.log.Printf("GX Server World create character failed: %v", err)
		reject("database")
		return
	}

	s.log.Printf("GX Server World created character id=%d name=%q map=%d", id, nc.Name, nc.MapID)
	s.send(buildCreateCharacterResult(true), "0x0102 create character ok")
	s.sendCharacterSelect()
}

func (s *Session) handleRename(p []byte) {
	if len(p) < 6 {
		return
	}
	id, _ := packet.U32(p[2:])
	name := packet.FixedText(p[6:])
	ok := false
	if strings.TrimSpace(name) != "" {
		if taken, err := db.NameTaken(name); err == nil && !taken {
			for _, c := range s.characters {
				if c.ID == id && c.RenamePending {
					ok, err = db.RenameCharacter(s.accountID, id, c.Name, name)
					s.logErr("rename character", err)
				}
			}
		}
	}
	s.log.Printf("GX Server World rename character id=%d name=%q ok=%t", id, name, ok)
	s.send(buildRenameResult(ok), "0x010E rename character")
}

func (s *Session) selectCharacter(id uint32) {
	if s.char != nil {
		return
	}
	failed := func(what string, err error) {
		s.logErr(what, err)
		s.char = nil
		s.send(buildCharResult(packet.SelectCharacter, false, id), "character data unavailable")
	}
	s.teleportStone = db.Item{}
	s.teleportCast = nil
	s.teleportRune = nil
	s.summonCastEnd = time.Time{}
	ch, err := db.CharacterDetail(s.accountID, id)
	if err != nil {
		s.log.Printf("GX Server World select character id=%d failed: %v", id, err)
		s.send(buildCharResult(packet.SelectCharacter, false, id), "0x0104 select character failed")
		return
	}
	normalize(&ch)
	originalMap := ch.MapID
	g, geometryErr := s.playerGeometry(ch.MapID)
	relocatedHeight := geometryErr == nil && !g.ValidPlayerHeight(ch.PosY)
	// Instanced maps are not kept across sessions. Out-of-range saves
	// take precedence over the instance's usual bind point.
	if relocatedHeight || !s.canEnterMap(ch.MapID, ch.Level) {
		p := birthPosition(&ch)
		ch.MapID, ch.PosX, ch.PosY, ch.PosZ = p.MapID, p.X, p.Y, p.Z
	} else if mapCreateType(ch.MapID) != "" {
		ch.MapID, ch.PosX, ch.PosY, ch.PosZ = s.respawnDestination(&ch)
	}
	s.char = &ch
	s.log.SetTag(strings.SplitN(s.log.Tag(), "/", 3)[0] + "/" + s.account.UserID + "/" + ch.Name)
	s.items, err = db.CharacterItems(id)
	if err != nil {
		failed("load items", err)
		return
	}
	s.skills, err = db.CharacterSkills(id)
	if err != nil {
		failed("load skills", err)
		return
	}
	s.effectWatchID = 0
	s.lastActionAt = time.Time{}
	s.dash = nil
	s.pendingCombo = nil
	s.pendingSkill = nil
	s.lastSkillID = 0
	s.skillCooldowns, s.itemCooldowns = map[uint16]time.Time{}, map[uint8]time.Time{}
	s.evolutionReadyAt, s.evolutionGrant, s.evolutionActions = nil, 0, [2]evolutionAction{}
	for i := range s.skills {
		k := &s.skills[i]
		if d := db.MustDefs().Skill(k.SkillID, k.Level); d != nil && skilltiming.ReuseSeconds(d.ResetTimeSeconds) == 0 {
			k.Delay = 0
		}
		if k.Delay > 0 {
			s.skillCooldowns[k.SkillID] = time.Now().Add(time.Duration(k.Delay) * time.Second)
		}
	}
	retainedCooldowns, err := db.ResetSkillCooldowns(id)
	if err != nil {
		failed("load retained skill cooldowns", err)
		return
	}
	for skillID, deadline := range retainedCooldowns {
		if deadline.After(s.skillCooldowns[skillID]) {
			s.skillCooldowns[skillID] = deadline
		}
	}
	for i := range s.skills {
		s.skills[i].Delay = s.remainingSkillDelay(s.skills[i].SkillID)
	}
	s.speedKnown = false
	s.movement = playerMovement{}
	s.oxygen = oxygenState{}
	s.elements = elementRuntime{}
	// GMs start transparent and unattackable (the original admin switches).
	s.gmConcealed, s.gmImmune = s.isAdmin, s.isAdmin
	s.buffs, err = restoreBuffs(id, s.log)
	if err != nil {
		failed("load buffs", err)
		return
	}

	s.quickSlots, err = db.QuickSlots(id)
	if err != nil {
		failed("load quick slots", err)
		return
	}
	count := len(s.quickSlots)
	s.quickSlots = s.discardEmptyItemShortcuts(s.quickSlots)
	if len(s.quickSlots) != count {
		if err := db.SaveQuickSlots(id, s.quickSlots); err != nil {
			failed("remove empty item shortcuts", err)
			return
		}
	}
	s.stored, err = db.StoredItems(s.accountID)
	if err != nil {
		failed("load warehouse", err)
		return
	}
	s.points, err = db.Points(s.accountID)
	if err != nil {
		failed("load points", err)
		return
	}
	if err := s.loadQuests(); err != nil {
		failed("load quests", err)
		return
	}
	s.mounted, s.dead, s.cast, s.autoAttack = false, false, nil, false
	s.reviveOnExit(s.deathExpLoss()) // an older save left at 0 HP
	s.statusItemProtection = 0
	s.lastRegenAt, s.regenMode = time.Now(), 1 // InitEquipment
	s.refreshMapCurses(false)
	st := s.stats()
	s.log.Printf("GX Server World selected character id=%d items=%d skills=%d buffs=%d hp=%d/%d", id, len(s.items), len(s.skills), len(s.buffs), st.CurrentHP, st.MaxHP)
	if err := s.initPlayerMovement(time.Now()); err != nil {
		s.log.Printf("GX Server collision initialization failed: %v", err)
		s.char = nil
		s.send(buildCharResult(packet.SelectCharacter, false, id), "collision resources unavailable")
		return
	}
	if originalMap != ch.MapID || relocatedHeight {
		if err := db.SaveCharacter(&ch); err != nil {
			failed("relocate character", err)
			return
		}
	}
	s.awaitingMapLoad = true
	s.send(buildCharResult(packet.SelectCharacter, true, id), "0x0104 select character ok")

	s.sendInitialState(&st)
	s.sendGiftBox()
}

// sendInitialState is the captured character-select sequence; afterwards
// the server->client direction switches to the expanded XOR scheme.
func (s *Session) sendInitialState(st *CharacterStats) {
	s.sendEnchantRules()
	s.send(activeEconomy().packet(false), "economy settings")
	s.sendOJLimits()
	s.sendChestCloakFits()
	s.socialFriends = nil
	s.socialReloadAt = time.Time{}
	s.socialRefresh.Store(false)
	s.sendSocial()
	ch := s.char
	for _, p := range []labeled{
		{buildWorldDay(), "0x0404 world day"},
		{buildCharacterDetails(ch, st), "0x0105 character details"},
		{buildKillCountUpdate(0, ch.Kills), "0x020E kills"},
		{buildKillCountUpdate(1, ch.Deaths), "0x020E deaths"},
		{buildKillCountUpdate(2, ch.Victories), "0x020E victories"},
		{buildKillCountUpdate(3, ch.Defeats), "0x020E defeats"},
		{buildKCStatus(s.faction), "0x0211 country kill count"},
		{buildAdditionalStats(st), "0x0526 additional stats"},
		{buildMaxHPMPSP(ch, st), "0x051F max hp/mp/sp"},
		{buildMaxEnergy(ch, 0, st.MaxHP), "0x050B max hp"},
		{buildMaxEnergy(ch, 1, st.MaxSP), "0x050B max sp"},
		{buildMaxEnergy(ch, 2, st.MaxMP), "0x050B max mp"},
		{buildCurrentHitpoints(st), "0x0521 current hp/mp/sp"},
		{buildSetMoney(ch.Money), "0x0213 set money"},
	} {
		s.send(p.Packet, p.Label)
	}
	for _, p := range buildInventoryPackets(s.items) {
		s.send(p, "0x0106 inventory")
	}
	for _, p := range []labeled{
		{buildCharacterSkills(s.skills), "0x0108 skills"},
	} {
		s.send(p.Packet, p.Label)
	}
	for _, p := range s.questListPackets() {
		s.send(p, "0x0901/0x0906 quests")
	}
	for _, p := range []labeled{
		{buildCharacterActiveBuffs(s.buffs), "0x010A active buffs"},
		{buildZeroCount(packet.BankItemList), "0xB101 bank"},
		{buildWarehouseItemList(s.stored), "0x0711 warehouse"},
		{buildZeroCount(packet.TeleportSavePositionList), "0x0225 saved teleports"},
		{buildAccountPoints(s.points), "0x2601 account points"},
		{buildCharacterSkillBar(s.quickSlots), "0x010B skill bar"},
	} {
		s.send(p.Packet, p.Label)
	}
	s.expanded = true
}

// handleMapLoaded (0x0201 from the client): send the map state, register
// the player and exchange visibility with nearby players.
func (s *Session) handleMapLoaded() {
	if !s.awaitingMapLoad {
		return
	}
	s.awaitingMapLoad = false
	ch := s.char
	if ch == nil {
		return
	}
	s.send(buildCharacterMapPosition(ch), "0x020B set map position")
	s.send(buildCharacterEnteredMap(ch, s.isAdmin, true), "0x0201 entered map")
	s.gmReply("")
	s.sendPhysicsClock()
	s.send(buildCharacterShape(ch, s.faction, s.items, 0), "0x0303 character shape")
	s.send(buildCharacterMove(ch, 0, 0), "0x0501 character move")
	if s.player == nil {
		// Register and send the entry snapshot before live blessing broadcasts
		// can reach this character. This also closes the map-load interval.
		kcSendMu.Lock()
		s.player = RegisterPlayer(*ch, s.items, s.isAdmin, s.faction, s.out, s.logger, s)
		if snapshot := buildBlessingSnapshot(s.faction); snapshot != nil {
			s.send(snapshot, "GX blessing snapshot")
		}
		kcSendMu.Unlock()
		s.guildConnect()
		s.activity("player.enter", ch.Name+" entered World")
	} else {
		s.activity("player.map", ch.Name+" entered map")
		s.clearView(false)
		s.player.SetMap(ch.MapID)
		s.player.UpdateCharacter(*ch)
		s.player.UpdateItems(s.items)
	}
	s.equipment()
	s.syncPlayerSpeed(true)
	s.syncVisibility()
	s.send(s.player.visibilityPacket(), "GX visibility")
	s.sendGMModes()
	s.send(buildVehicleState(s.player.vehicleState()), "GX own vehicle state")
	s.sendRelicList()
	s.updateWeather(time.Now(), true)
	// CZone::EnterUser + SendCellInfoView: the 5x5 cells around the user.
	s.placeView()
	s.sendPlayerFeedback(proto.PlayerSpawnFeedback)
	s.syncQuestNPCCatalog(true)
	s.syncQuestMobCatalog(true)
	near := EntitiesInView(ch.MapID, cellFor(ch.PosX, ch.PosZ))
	others := s.player.knownIDs(viewUser)
	s.log.Printf("GX Server World map loaded map=%d npcs=%d mobs=%d players=%d", ch.MapID, len(near.NPCs), len(near.Mobs), len(others))
}

// handleMotion (0x0506): sit/stand persist; other motions are one-shot.
// The sender gets its own echo, which the client needs to play it.
func (s *Session) handleMotion(p []byte) {
	if len(p) < 3 || s.char == nil {
		return
	}
	motion := p[2]
	if motion == 2 && s.player != nil && s.player.vehicleState().State == proto.VehiclePassenger {
		s.dismount()
		return
	}
	if s.ridingVehicle() && motion != proto.MotionJump {
		return
	}
	s.cancelVehicleCast()
	if motion == proto.MotionJump {
		// A swing or cast has just been committed: no jump for that instant.
		if time.Now().Before(s.attackMoveUntil) {
			return
		}
		if !s.startMovementJump(time.Now()) {
			return
		}
		s.autoAttack, s.pendingCombo, s.pendingSkill = false, nil, nil
		if s.cast != nil {
			s.cast = nil
			s.syncPlayerSpeed(false)
		}
	}
	if s.player != nil {
		switch motion {
		case MotionSitDown:
			s.player.SetSitting(true)
		case MotionStandUp:
			s.player.SetSitting(false)
		}
		if motion != MotionSitDown {
			s.endMyShop() // standing, jumping or any other motion ends the market
		}
	}
	s.broadcastNear(buildCharacterMotion(s.char.ID, motion), "0x0506 character motion", true)
}

// handleLogout (0x0107) returns to character select on the same
// connection; the account stays held.
func (s *Session) handleLogout() {
	id := uint32(0)
	if s.char != nil {
		id = s.char.ID
	}
	if !s.leaveWorld(false) {
		if s.conn != nil {
			netx.Abort(s.conn)
		}
		return
	}
	s.awaitingMapLoad = false
	s.buffs, s.cast, s.mounted, s.dead, s.autoAttack = nil, nil, false, false, false
	s.items, s.skills, s.quickSlots, s.stored, s.points = nil, nil, nil, nil, 0
	s.send(buildLeaveWorldAck(id), "0x0107 leave world ack")
	s.expanded = false
	s.weatherOverrideUntil = time.Time{}
	s.weatherCheck = time.Time{}
	s.sendCharacterSelect()
}
