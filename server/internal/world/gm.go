package world

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/gm"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

const gxGMState packet.Opcode = 0x7F01
const gmMaxSpawn = 100
const gxGMModes packet.Opcode = 0x7F29

func (s *Session) turboActive() bool { return s.isAdmin && s.gmTurbo }

func (s *Session) gmReply(message string) {
	if s.char == nil {
		return
	}
	w := packet.New(gxGMState)
	w.U32(s.char.ID)
	role, turbo := uint8(0), uint8(0)
	if s.isAdmin {
		role = 1
		if s.turboActive() {
			turbo = 1
		}
	}
	w.U8(role)
	w.U8(turbo)
	w.U16(uint16(len(message)))
	w.Bytes([]byte(message))
	s.send(w.MustFinish(), "GX GM state")
}

// sendGMModes publishes the three GM switches to the GM's own client.
func (s *Session) sendGMModes() {
	if s.char == nil || !s.isAdmin {
		return
	}
	b := func(v bool) uint8 {
		if v {
			return 1
		}
		return 0
	}
	w := packet.New(gxGMModes)
	w.U32(s.char.ID)
	w.U8(b(s.gmConcealed))
	w.U8(b(s.gmImmune))
	w.U8(b(s.turboActive()))
	s.send(w.MustFinish(), "GX GM modes")
}

// setGMChar and setGMAttack are the original admin commands (CUser::
// PacketAdminCmdE 0xFA01/0xFA02 bVisible and 0xFA05/0xFA06 bAttackable, chat
// "/char on|off" and "/attack on|off" at Game.exe 0x471c20). "/char on" makes
// the character transparent; "/attack on" makes it unattackable.
// gmModeGuard rejects a /char or /attack change that follows another within
// 750 ms. Each change re-evaluates who sees the GM, so a client that loops (or
// an input burst) must not be able to flood the world with them.
func (s *Session) gmModeGuard() bool {
	now := time.Now()
	if now.Sub(s.lastGMModeAt) < 750*time.Millisecond {
		s.gmReply("GM: wait a moment before changing this again.")
		return false
	}
	s.lastGMModeAt = now
	return true
}

func (s *Session) setGMChar(on bool) {
	s.gmConcealed = on
	s.syncVisibility()
	if on {
		s.gmReply("Character's transparent setting is on")
	} else {
		s.gmReply("Character's transparent setting is off")
	}
	s.sendGMModes()
}

func (s *Session) setGMAttack(on bool) {
	s.gmImmune = on
	if on {
		s.removeEffects(func(b *ActiveBuff) bool { return b.Effect != typeEffectBuff }, "GM immunity")
	}
	s.syncVisibility()
	if on {
		s.gmReply("Won't be attacked anymore")
	} else {
		s.gmReply("Will be attacked")
	}
	s.sendGMModes()
}

func gmNumbers(args []string, bounds ...uint64) ([]uint64, error) {
	if len(args) != len(bounds) {
		return nil, fmt.Errorf("argument count")
	}
	out := make([]uint64, len(args))
	for i, a := range args {
		n, err := strconv.ParseUint(a, 10, 32)
		if err != nil || n == 0 || n > bounds[i] {
			return nil, fmt.Errorf("argument value")
		}
		out[i] = n
	}
	return out, nil
}

// Known commands are consumed even when refused; they never become public chat.
func (s *Session) handleGMCommand(text string) bool {
	command, args, known := gm.Parse(text)
	if !known {
		return false
	}
	if s.char == nil {
		return true
	}
	if !s.isAdmin {
		s.activity("staff.command", fmt.Sprintf("%s denied GM command %s", s.char.Name, command), "command", command, "outcome", "denied")
		s.gmReply("GM: access denied.")
		return true
	}
	if tradeFor(s.char.ID) != nil { // Inventory must not change underneath an exchange.
		s.activity("staff.command", fmt.Sprintf("%s blocked GM command %s while trading", s.char.Name, command), "command", command, "outcome", "blocked")
		s.gmReply("GM: finish trading before using commands.")
		return true
	}
	var err error
	switch command {
	case "/bots":
		err = s.gmBots(args)
	case "/mera":
		err = s.gmEraseMobs(args)
	case "/nera":
		err = s.gmEraseNPC(args)
	case "/kill":
		err = s.gmKill(args)
	case "/notice":
		err = s.gmNotice(strings.TrimSpace(strings.TrimSpace(text)[len(command):]))
	case "/weather":
		err = s.gmWeather(args)
	case "/set":
		err = s.gmSet(args)
	case "/char", "/attack":
		if len(args) != 1 || (!strings.EqualFold(args[0], "on") && !strings.EqualFold(args[0], "off")) {
			s.gmReply("Usage: " + command + " on|off")
			break
		}
		if !s.gmModeGuard() {
			break
		}
		if command == "/char" {
			s.setGMChar(strings.EqualFold(args[0], "on"))
		} else {
			s.setGMAttack(strings.EqualFold(args[0], "on"))
		}
	case "/turbo":
		if len(args) != 0 {
			s.gmReply("Usage: /turbo")
			break
		}
		s.gmTurbo = !s.gmTurbo
		defer s.sendGMModes()
		s.movement.budgetAt, s.movement.credit = time.Now(), 0
		s.movement.speedCorrectionPending = false
		if s.gmTurbo {
			s.gmReply("GM: turbo enabled. Fall damage disabled. Server speed limits disabled.")
		} else {
			s.gmReply("GM: turbo disabled. Fall damage follows server settings. Server speed limits restored.")
		}
	case "/amove", "/asummon":
		if len(args) != 1 {
			s.gmReply("Usage: " + command + " charname")
			break
		}
		if command == "/amove" {
			err = s.gmGoTo(args[0])
		} else {
			err = s.gmBring(args[0])
		}
	case "/cure":
		if len(args) != 0 {
			s.gmReply("Usage: /cure")
			break
		}
		err = s.gmCure()
	case "/iclear":
		if len(args) != 0 {
			s.gmReply("Usage: /iclear")
			break
		}
		err = s.gmClearInventory()
	case "/getitem":
		if len(args) == 2 {
			args = append(args, "1")
		}
		n, e := gmNumbers(args, 255, 255, 5*24*255)
		if e != nil {
			s.gmReply("Usage: /getitem type id [quantity=1]")
			break
		}
		err = s.gmGiveItem(uint8(n[0]), uint8(n[1]), int(n[2]))
	case "/mmake":
		if len(args) == 1 {
			args = append(args, "1")
		}
		n, e := gmNumbers(args, 65535, gmMaxSpawn)
		if e != nil {
			s.gmReply("Usage: /mmake mobid [quantity=1] (1-100)")
			break
		}
		err = s.gmSpawn(uint16(n[0]), int(n[1]), 0)
	case "/nmake":
		n, e := gmNumbers(args, npcTypeCount, 65535)
		if e != nil {
			s.gmReply("Usage: /nmake type id")
			break
		}
		err = s.gmSpawn(uint16(n[1]), 1, uint8(n[0]))
	case "/cmove":
		if len(args) != 1 {
			s.gmReply("Usage: /cmove mapid")
			break
		}
		n, e := strconv.ParseUint(args[0], 10, 16)
		if e != nil {
			s.gmReply("Usage: /cmove mapid")
			break
		}
		err = s.gmMove(uint16(n))
	}
	if err != nil {
		s.gmReply("GM: " + err.Error())
	}
	outcome := "handled"
	if err != nil {
		outcome = "error"
	}
	// Only the known verb is logged. Arguments may contain private staff text.
	s.activity("staff.command", fmt.Sprintf("%s used %s (%s)", s.char.Name, command, outcome), "command", command, "outcome", outcome)
	return true
}

// Clear carried inventory only. Equipped items and warehouse contents survive.
func (s *Session) gmClearInventory() error {
	var removed, kept []db.Item
	var ids []int64
	for _, it := range s.items {
		if it.Bag >= InventoryFirstBag && it.Bag <= InventoryLastBag {
			removed = append(removed, it)
			ids = append(ids, it.UID)
		} else {
			kept = append(kept, it)
		}
	}
	if err := db.DeleteCharacterItems(s.char.ID, ids...); err != nil {
		return err
	}
	s.items = kept
	if s.player != nil {
		s.player.UpdateItems(s.items)
	}
	for i := range removed {
		s.send(buildRemoveItem(&removed[i], true), "GM inventory clear")
	}
	s.gmReply(fmt.Sprintf("GM: inventory cleared (%d carried items).", len(removed)))
	return nil
}

// planGMGrant keeps the entire request atomic, including quantities above 255.
func planGMGrant(items []db.Item, def *db.ItemDef, quantity int) ([]db.Item, []db.Item, []db.Item, error) {
	if def == nil || quantity <= 0 || quantity > 5*24*255 {
		return nil, nil, nil, fmt.Errorf("invalid item or quantity")
	}
	sim := append([]db.Item(nil), items...)
	tpl := db.Item{Type: def.Type, TypeID: def.TypeID, Quality: def.Quality, MakeType: "G"}
	var changed, added []db.Item
	stack := def.MaxStack()
	for quantity > 0 {
		if i := findStackable(sim, &tpl, stack); i >= 0 {
			n := min(quantity, int(stack-sim[i].Count))
			sim[i].Count += uint16(n)
			quantity -= n
			changed = append(changed, sim[i])
			continue
		}
		free := freeInventorySlots(sim)
		if len(free) == 0 {
			return nil, nil, nil, fmt.Errorf("not enough space for the entire quantity")
		}
		it := tpl
		it.UID = db.NewItemUID()
		it.MakeTime = time.Now()
		it.Bag, it.Slot = free[0][0], free[0][1]
		it.Count = uint16(min(quantity, int(stack)))
		quantity -= int(it.Count)
		sim = append(sim, it)
		added = append(added, it)
	}
	return sim, changed, added, nil
}

func (s *Session) gmGiveItem(itemType, id uint8, quantity int) error {
	def := db.MustDefs().Item(itemType, id)
	sim, changed, added, err := planGMGrant(s.items, def, quantity)
	if err != nil {
		return err
	}
	if err = db.ApplyGMItemGrant(s.char.ID, changed, added); err != nil {
		return err
	}
	s.items = sim
	for i := range changed {
		s.send(buildAddItem(&changed[i]), "GM item stack")
	}
	for i := range added {
		s.send(buildAddItem(&added[i]), "GM item grant")
	}
	if s.player != nil {
		s.player.UpdateItems(s.items)
	}
	s.gmReply(fmt.Sprintf("GM: received %d x %s (%d/%d).", quantity, def.Name, itemType, id))
	return nil
}

// Spawn positions form expanding rings; occupied points and blocked cells
// are rejected, including entities created by earlier commands.
func gmSpawnPositions(mapID uint16, x, z float32, count int, entities *MapEntities) ([][2]float32, error) {
	size, exists := mapSize(mapID)
	if !exists || size <= 0 {
		return nil, fmt.Errorf("map unavailable")
	}
	var result [][2]float32
	for ring := 1; ring <= 16 && len(result) < count; ring++ {
		radius := float64(3 * ring)
		points := max(6, 6*ring)
		for j := 0; j < points && len(result) < count; j++ {
			angle := 2 * math.Pi * float64(j) / float64(points)
			px, pz := x+float32(radius*math.Cos(angle)), z+float32(radius*math.Sin(angle))
			if px < 1 || pz < 1 || px >= size-1 || pz >= size-1 || !IsMapWalkable(mapID, px, pz) {
				continue
			}
			occupied := false
			for _, n := range entities.NPCs {
				if DistanceXZ(px, pz, n.PosX, n.PosZ) < 2 {
					occupied = true
					break
				}
			}
			if !occupied {
				for _, m := range entities.Mobs {
					if m.CurrentHP > 0 && DistanceXZ(px, pz, m.PosX, m.PosZ) < 2 {
						occupied = true
						break
					}
				}
			}
			if !occupied {
				for _, p := range result {
					if DistanceXZ(px, pz, p[0], p[1]) < 2 {
						occupied = true
						break
					}
				}
			}
			if !occupied {
				result = append(result, [2]float32{px, pz})
			}
		}
	}
	if len(result) != count {
		return nil, fmt.Errorf("no nearby free space")
	}
	return result, nil
}

func (s *Session) gmSpawn(id uint16, count int, npcType uint8) error {
	if count < 1 || count > gmMaxSpawn {
		return fmt.Errorf("quantity out of range")
	}
	var nd *npcDef
	var md *db.MobDef
	if npcType != 0 {
		if count != 1 {
			return fmt.Errorf("nmake summons exactly one NPC")
		}
		nd = findNpcDef(npcType, id)
		if nd == nil {
			return fmt.Errorf("unknown NPC; use /nmake type id (1 1 = merchant 1)")
		}
	} else {
		md = db.MustDefs().Mobs[id]
		if md == nil {
			return fmt.Errorf("unknown mob")
		}
	}
	sh := shard(s.char.MapID)
	nav, navErr := loadMobNavigation(s.char.MapID)
	if navErr != nil {
		return navErr
	}
	sh.mu.Lock()
	positions, err := gmSpawnPositions(s.char.MapID, s.char.PosX, s.char.PosZ, count, &sh.entities)
	if err != nil {
		sh.mu.Unlock()
		return err
	}
	var npcs []MapNpc
	var mobs []MapMob
	for _, pos := range positions {
		if npcType != 0 {
			n := MapNpc{ID: realm.nextNpcID.Add(1) - 1, Type: nd.Type, TypeID: nd.TypeID, PosX: pos[0], PosY: s.char.PosY, PosZ: pos[1], Summoned: true, cell: cellFor(pos[0], pos[1]), hasCell: true}
			if n.Type == npcTypeGuard {
				if nav != nil {
					if ground, ok := nav.Project(n.point(), .6, 5); ok {
						n.PosY = ground[1]
					}
				}
				// One point anchors pursuit and return without starting a patrol.
				n.patrol = []NpcPatrolPoint{{X: n.PosX, Y: n.PosY, Z: n.PosZ, Angle: n.Angle}}
			}
			sh.entities.NPCs = append(sh.entities.NPCs, n)
			sh.npcRevision.Add(1)
			npcs = append(npcs, n)
		} else {
			m := NewStaticMob(md, id, pos[0], pos[1], nil, 0)
			m.PosY = s.char.PosY
			if !m.placeOnNavigation(nav, s.char.MapID) {
				continue
			}
			sh.entities.Mobs = append(sh.entities.Mobs, m)
			mobs = append(mobs, m)
		}
	}
	sh.mu.Unlock()
	for _, n := range npcs {
		for _, p := range PlayersViewing(s.char.MapID, n.cell) {
			viewEnter(p, viewNpc, n.ID, buildNpcEnter(&n), "GM NPC spawn")
		}
	}
	for _, m := range mobs {
		for _, p := range PlayersViewing(s.char.MapID, m.cell) {
			viewEnter(p, viewMob, m.ID, buildMobEnter(&m, false), "GM mob spawn")
		}
	}
	spawned := len(npcs) + len(mobs)
	if spawned == 0 {
		return fmt.Errorf("no valid navigable spawn surface for id %d", id)
	}
	s.gmReply(fmt.Sprintf("GM: summoned %d (id %d).", spawned, id))
	return nil
}

func (s *Session) gmMove(mapID uint16) error {
	size, ok := mapSize(mapID)
	if !ok || size <= 0 || isInstanceZone(mapID) {
		return fmt.Errorf("unknown map")
	}
	// The server owns bind heights, not terrain geometry. Choose the closest
	// bind to the equivalent relative X/Z rather than guessing an unsafe Y.
	x, z := s.char.PosX, s.char.PosZ
	if old, ok := mapSize(s.char.MapID); ok && old > 0 {
		x = x / old * size
		z = z / old * size
	}
	entities := EntitiesOnMap(mapID)
	var target *SpawnBox
	best := float32(math.MaxFloat32)
	for _, a := range entities.SpawnAreas {
		for i := range a.Boxes {
			b := a.Boxes[i]
			bx, bz := (b.LowerX+b.UpperX)/2, (b.LowerZ+b.UpperZ)/2
			d := DistanceXZ(x, z, bx, bz)
			if d < best && IsMapWalkable(mapID, bx, bz) {
				copy := b
				target = &copy
				best = d
			}
		}
	}
	if target == nil {
		// Dungeons can have no rebirth boxes; use an authored incoming portal
		// or gatekeeper destination, with its explicit server-side height.
		consider := func(px, py, pz float32) {
			d := DistanceXZ(x, z, px, pz)
			if d < best && px >= 0 && pz >= 0 && px < size && pz < size && IsMapWalkable(mapID, px, pz) {
				target = &SpawnBox{LowerX: px, UpperX: px, LowerY: py - .9, LowerZ: pz, UpperZ: pz}
				best = d
			}
		}
		for _, id := range shardIDs() {
			for _, portal := range EntitiesOnMap(id).Portals {
				if portal.TargetMapID == mapID {
					consider(portal.TargetX, portal.TargetY, portal.TargetZ)
				}
			}
		}
		for _, list := range npcDefs {
			for _, n := range list {
				if n.Type == npcTypeGatekeeper {
					for _, g := range n.Gates {
						if g.MapID == mapID {
							consider(g.X, g.Y, g.Z)
						}
					}
				}
			}
		}
	}
	var arrival MapPos
	if target != nil {
		arrival = MapPos{mapID, (target.LowerX + target.UpperX) / 2, target.LowerY + .9, (target.LowerZ + target.UpperZ) / 2}
	} else {
		// Event maps can have a faction bind without an incoming portal.
		if mapID == 39 { // Secret Arena has no local bind.
			arrival = MapPos{39, 83, 2.13, 66}
		} else {
			var found bool
			arrival, found = rebirthPosition(mapID, s.faction)
			if !found || arrival.MapID != mapID {
				return fmt.Errorf("map has no safe arrival point")
			}
		}
		if !IsMapWalkable(mapID, arrival.X, arrival.Z) {
			return fmt.Errorf("map has no safe arrival point")
		}
	}
	s.cast = nil
	s.autoAttack = false
	s.useNpc(0)
	if !s.applyTeleport(mapID, arrival.X, arrival.Y, arrival.Z, "GM cmove", true) {
		return fmt.Errorf("destination unavailable")
	}
	s.gmReply(fmt.Sprintf("GM: map %d.", mapID))
	return nil
}

const gmSetUsage = "usage: /set money|kill|str|dex|rec|int|wis|luc|statpoint|skillpoint|level amount"

// gmSet changes only the issuing character. Claimed rewards remain claimed.
func (s *Session) gmSet(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("%s", gmSetUsage)
	}
	field := strings.ToLower(args[0])
	if field == "level" {
		return s.gmSetLevel(args[1:])
	}
	var cap uint64
	switch field {
	case "money":
		cap = db.MoneyCap
	case "kill":
		cap = pvpKillCap
	case "str", "dex", "rec", "int", "wis", "luc", "statpoint", "skillpoint":
		cap = db.BaseStatCap
	default:
		return fmt.Errorf("%s", gmSetUsage)
	}
	// Saturate while parsing so even an arbitrarily large decimal cannot overflow.
	var value uint64
	if args[1] == "" {
		return fmt.Errorf("amount must be a non-negative integer")
	}
	for _, digit := range args[1] {
		if digit < '0' || digit > '9' {
			return fmt.Errorf("amount must be a non-negative integer")
		}
		value = min(value*10+uint64(digit-'0'), cap)
	}
	next := *s.char
	switch field {
	case "money":
		next.Money = uint32(value)
	case "kill":
		next.Kills = uint32(value)
	case "str":
		next.Str = uint16(value)
	case "dex":
		next.Dex = uint16(value)
	case "rec":
		next.Rec = uint16(value)
	case "int":
		next.Int = uint16(value)
	case "wis":
		next.Wis = uint16(value)
	case "luc":
		next.Luc = uint16(value)
	case "statpoint":
		next.StatPoint = uint16(value)
	case "skillpoint":
		next.SkillPoint = uint16(value)
	}
	points := field == "statpoint" || field == "skillpoint"
	stat := field != "money" && field != "kill" && !points
	if stat {
		normalize(&next)
	}
	if err := db.SaveCharacter(&next); err != nil {
		return err
	}
	*s.char = next
	s.syncPlayer()
	if field == "money" {
		s.send(buildSetMoney(next.Money), "GM set money")
	} else if points {
		// Unspent points are shown by the level-up record, which carries both.
		s.send(buildLevelUp(packet.LevelUpSelf, &next), "GM set points")
	} else if field == "kill" {
		s.send(buildKillCountUpdate(0, next.Kills), "GM set kills")
		w := packet.New(0x0522)
		w.U32(next.ID)
		w.U32(next.Kills)
		frame := w.MustFinish()
		s.send(frame, "GM set rank")
		s.sendToViewers(frame, "GM set rank (view)")
	} else {
		s.send(buildStatsReset(next.StatPoint, next.Str, next.Rec, next.Int, next.Wis, next.Dex, next.Luc), "GM set base stats")
		// 0214 resets the client's displayed claim stage; restore the actual
		// unchanged stages immediately without reopening any reward.
		s.send(buildPvPRewardStatus(&next), "GM set unchanged PvP rewards")
		s.sendStatPackets()
		s.partyPools()
	}
	s.gmReply(fmt.Sprintf("GM: %s set to %d.", field, value))
	return nil
}

// gmCure clears all debuffs and cooldowns and restores HP, MP and SP.
func (s *Session) gmCure() error {
	if s.dead {
		return fmt.Errorf("you are dead")
	}
	s.removeEffectsInternal(func(b *ActiveBuff) bool { return b.Effect == typeEffectDebuff }, "GM cure", true)
	s.refillPools()
	s.persist()
	st := s.stats()
	s.send(buildCurrentHitpoints(&st), "0x0521 GM cure")
	for id := range s.skillCooldowns {
		s.sendSkillCooldown(id, 0, 0)
	}
	clear(s.skillCooldowns)
	clear(s.itemCooldowns)
	vehicleMu.Lock()
	delete(vehicleCooldowns, s.char.ID)
	vehicleMu.Unlock()
	s.send(packet.New(packet.Opcode(proto.GXItemCooldownReset)).MustFinish(), "GM item cooldown reset")
	s.persistSkillDelays()
	s.gmReply("GM: restored HP, MP and SP and cleared all debuffs and cooldowns.")
	return nil
}
