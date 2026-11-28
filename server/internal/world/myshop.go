package world

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Spelunkern/Shaiya-GX/internal/chatguard"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/textwire"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

// Personal markets (setup market), after CUser::PacketMyShop 0x478bd0 and
// stMyShop: a seated player sells items from the bags at a price per unit in
// a map where markets are allowed. Buyers browse the list and choose how many
// units to take. Moving, standing, dying, leaving the map or logging out ends
// the market.
const (
	opMyShopBegin  packet.Opcode = 0x2301 // enter setup
	opMyShopCancel packet.Opcode = 0x2302 // cancel / market closed
	opMyShopOpen   packet.Opcode = 0x2303 // start selling with a message
	opMyShopStop   packet.Opcode = 0x2304 // stop selling, back to setup
	opMyShopAdd    packet.Opcode = 0x2305
	opMyShopRemove packet.Opcode = 0x2306
	opMyShopPrice  packet.Opcode = 0x2307
	opMyShopSold   packet.Opcode = 0x2308 // to the seller: a sale
	opMyShopEnter  packet.Opcode = 0x2309 // a buyer opens a market
	opMyShopLeave  packet.Opcode = 0x230a
	opMyShopList   packet.Opcode = 0x230b
	opMyShopClosed packet.Opcode = 0x230c // to buyers: the market ended
	opMyShopBuy    packet.Opcode = 0x230d
	opMyShopStock  packet.Opcode = 0x230e // to buyers: an item's new count
	opShopBubble   packet.Opcode = 0x0523 // the market's bubble over the seller

	myShopSlots     = proto.MyShopSlots
	myShopBuyersMax = 16
)

// Maps where a market can be set up (GX Server configuration for now).
var myShopMaps = map[uint16]bool{35: true, 36: true, 42: true}

// Results of the begin request (client message in the reply).
const (
	myShopBeginOK       = 0
	myShopBeginWrongMap = 1
	myShopBeginBusy     = 2
)

// Results of a purchase.
const (
	myShopBuyOK         = 1
	myShopBuyInvalid    = 0
	myShopBuyNoMoney    = 2
	myShopBuySellerCap  = 3
	myShopBuyNoRoom     = 4
	myShopEnterOK       = 1
	myShopEnterFailed   = 0
	myShopEnterTooMany  = 2
	myShopStatusNone    = 0
	myShopStatusSetup   = 1
	myShopStatusSelling = 2
)

type shopSlot struct {
	used  bool
	item  db.Item
	price uint32
}

type myShop struct {
	status int
	slots  [myShopSlots]shopSlot
	buyers []*ConnectedPlayer
}

// Match the offered item even when its bag position is occupied by a new item.
func (sl shopSlot) itemIndex(items []db.Item) int {
	if !sl.used {
		return -1
	}
	i := findItem(items, sl.item.Bag, sl.item.Slot)
	if i < 0 {
		return -1
	}
	it := items[i]
	if it.Count == 0 || it.UID != sl.item.UID || it.Type != sl.item.Type || it.TypeID != sl.item.TypeID ||
		it.Quality != sl.item.Quality || it.Gems != sl.item.Gems || it.CraftName != sl.item.CraftName {
		return -1
	}
	return i
}

func shopAck(op packet.Opcode, body func(w *packet.Writer)) []byte {
	w := packet.New(op)
	if body != nil {
		body(w)
	}
	return w.MustFinish()
}

// shopBubbleFrame is 0x0523: the market message over the seller (off: no text).
func shopBubbleFrame(id uint32, title string) []byte {
	w := packet.New(opShopBubble)
	w.U32(id)
	if title == "" {
		w.U8(0)
		return w.MustFinish()
	}
	w.U8(1)
	w.U8(uint8(len(title)))
	w.Bytes([]byte(title))
	return w.MustFinish()
}

func (p *ConnectedPlayer) setShopTitle(title string) {
	p.shopMu.Lock()
	p.shopTitle = title
	p.shopMu.Unlock()
}

func (p *ConnectedPlayer) currentShopTitle() string {
	p.shopMu.Lock()
	defer p.shopMu.Unlock()
	return p.shopTitle
}

func (p *ConnectedPlayer) setViewedShop(seller *ConnectedPlayer) {
	p.shopMu.Lock()
	p.viewedShop = seller
	p.shopMu.Unlock()
}

func (p *ConnectedPlayer) currentViewedShop() *ConnectedPlayer {
	p.shopMu.Lock()
	defer p.shopMu.Unlock()
	return p.viewedShop
}

// Order a seller's notifications against the buyer switching or leaving a shop.
func (p *ConnectedPlayer) sendShopUpdate(seller *ConnectedPlayer, frame []byte, closing bool) {
	p.shopMu.Lock()
	defer p.shopMu.Unlock()
	if p.viewedShop != seller {
		return
	}
	if closing {
		p.viewedShop = nil
	}
	p.Send(frame, "personal market update")
}

func (s *Session) closeMyShopView() {
	p := s.player
	if p == nil {
		return
	}
	p.shopMu.Lock()
	defer p.shopMu.Unlock()
	if p.viewedShop != nil {
		p.viewedShop = nil
		s.send(shopAck(opMyShopClosed, nil), "0x230c market view closed")
	}
}

func (s *Session) tickMyShop() {
	if s.player == nil {
		return
	}
	if s.shop.status != myShopStatusNone {
		if s.dead || s.ridingVehicle() || tradeFor(s.char.ID) != nil {
			s.endMyShop()
		} else {
			s.refreshMyShop()
		}
	}
	if seller := s.player.currentViewedShop(); seller != nil && !s.canBrowseShop(seller) {
		s.closeMyShopView()
	}
}

// canOpenMyShop is the setup precondition.
func (s *Session) canOpenMyShop() int {
	ch := s.char
	if !myShopMaps[ch.MapID] {
		return myShopBeginWrongMap
	}
	if s.dead || ch.CurrentHP == 0 || s.player == nil || tradeFor(ch.ID) != nil || s.player.vehicleState().State != proto.VehicleOff || s.cast != nil || s.dash != nil ||
		s.inCombat() || s.actionBlocked() || s.movement.ladder.Active || s.movement.jumping {
		return myShopBeginBusy
	}
	return myShopBeginOK
}

// handleMyShop covers the seller's requests 0x2301-0x2307 and the buyer's
// 0x2309, 0x230a and 0x230d.
func (s *Session) handleMyShop(op packet.Opcode, p []byte) {
	if s.char == nil || s.player == nil {
		return
	}
	switch op {
	case opMyShopBegin:
		if s.shop.status != myShopStatusNone {
			s.endMyShop() // starting again discards the previous market
		}
		result := s.canOpenMyShop()
		s.send(shopAck(opMyShopBegin, func(w *packet.Writer) { w.U8(uint8(result)) }), "0x2301 market setup")
		if result == myShopBeginOK {
			s.closeMyShopView()
			s.shop = myShop{status: myShopStatusSetup}
			// The seller sits as soon as the setup starts.
			s.cancelMove()
			if !s.player.IsSitting() {
				s.player.SetSitting(true)
				s.broadcastNear(buildCharacterMotion(s.char.ID, MotionSitDown), "0x0506 market sit", true)
			}
		}
	case opMyShopCancel:
		s.endMyShop()
	case opMyShopOpen:
		s.openMyShop(p)
	case opMyShopStop:
		if s.shop.status != myShopStatusSelling {
			return
		}
		s.stopMyShop()
		s.send(shopAck(opMyShopStop, nil), "0x2304 market stopped")
	case opMyShopAdd:
		if len(p) < 9 || s.shop.status != myShopStatusSetup {
			return
		}
		idx, bag, slot := p[2], p[3], p[4]
		price, _ := packet.U32(p[5:])
		if int(idx) >= myShopSlots || s.shop.slots[idx].used || price == 0 || price > db.MoneyCap || bag == 0 || !isValidItemPosition(bag, slot) {
			return
		}
		i := findItem(s.items, bag, slot)
		if i < 0 || realType(s.items[i].Type) == 27 || s.items[i].Count == 0 {
			return
		}
		for _, o := range s.shop.slots {
			if o.used && o.item.UID == s.items[i].UID {
				return
			}
		}
		s.shop.slots[idx] = shopSlot{used: true, item: s.items[i], price: price}
		s.send(shopAck(opMyShopAdd, func(w *packet.Writer) { w.U8(idx); w.U8(bag); w.U8(slot); w.U32(price) }), "0x2305 market item added")
	case opMyShopRemove:
		if len(p) < 3 || s.shop.status != myShopStatusSetup || int(p[2]) >= myShopSlots || !s.shop.slots[p[2]].used {
			return
		}
		s.shop.slots[p[2]] = shopSlot{}
		s.send(shopAck(opMyShopRemove, func(w *packet.Writer) { w.U8(p[2]) }), "0x2306 market item removed")
	case opMyShopPrice:
		if len(p) < 7 || s.shop.status != myShopStatusSetup || int(p[2]) >= myShopSlots || !s.shop.slots[p[2]].used {
			return
		}
		price, _ := packet.U32(p[3:])
		if price == 0 || price > db.MoneyCap {
			return
		}
		s.shop.slots[p[2]].price = price
		s.send(shopAck(opMyShopPrice, func(w *packet.Writer) { w.U8(p[2]); w.U32(price) }), "0x2307 market price changed")
	case opMyShopEnter:
		if len(p) >= 6 {
			id, _ := packet.U32(p[2:])
			s.enterMyShop(id)
		}
	case opMyShopLeave:
		s.player.setViewedShop(nil)
		s.send(shopAck(opMyShopLeave, nil), "0x230a market left")
	case opMyShopBuy:
		if len(p) >= 4 {
			s.buyFromMyShop(p[2], p[3])
		}
	}
}

// openMyShop is 0x2303: the market starts with its message and the seller sits.
func (s *Session) openMyShop(p []byte) {
	if s.shop.status != myShopStatusSetup || len(p) < 3 {
		return
	}
	n := int(p[2])
	if n > proto.MyShopTitleBytes || n != len(p)-3 {
		return
	}
	title := cleanShopTitle(decodeText(p[3:]))
	if !s.refreshMyShop() || title == "" || s.canOpenMyShop() != myShopBeginOK {
		return
	}
	s.shop.status = myShopStatusSelling
	s.cancelMove()
	s.player.setShopTitle(title)
	if !s.player.IsSitting() {
		s.player.SetSitting(true)
		s.broadcastNear(buildCharacterMotion(s.char.ID, MotionSitDown), "0x0506 market sit", true)
	}
	s.send(shopAck(opMyShopOpen, nil), "0x2303 market opened")
	s.broadcastNear(shopBubbleFrame(s.char.ID, title), "0x0523 market bubble", true)
}

func cleanShopTitle(title string) string {
	filtered := chatguard.Filter(title, false)
	if filtered.Impersonation {
		return "[Message hidden]"
	}
	runes := []rune(strings.TrimSpace(filtered.Text))
	return textwire.Prefix(string(runes[:min(len(runes), proto.MyShopTitleMax)]), proto.MyShopTitleBytes)
}

// stopMyShop goes back to setup: buyers are told, the bubble goes away.
func (s *Session) stopMyShop() {
	if s.shop.status != myShopStatusSelling {
		return
	}
	s.shop.status = myShopStatusSetup
	s.closeMyShopBuyers()
	s.player.setShopTitle("")
	s.broadcastNear(shopBubbleFrame(s.char.ID, ""), "0x0523 market bubble off", true)
}

func (s *Session) closeMyShopBuyers() {
	frame := shopAck(opMyShopClosed, nil)
	for _, buyer := range s.shop.buyers {
		buyer.sendShopUpdate(s.player, frame, true)
	}
	s.shop.buyers = nil
}

func (s *Session) canBrowseShop(seller *ConnectedPlayer) bool {
	if seller == nil || seller == s.player || seller.owner == nil || s.dead || s.char.CurrentHP == 0 || s.shop.status != myShopStatusNone || tradeFor(s.char.ID) != nil ||
		seller.MapID() != s.char.MapID || (seller.Faction() != s.faction && s.char.MapID != 42) || !s.player.canSee(seller) {
		return false
	}
	x, z := seller.PositionXZ()
	dx, dz := s.char.PosX-x, s.char.PosZ-z
	return dx*dx+dz*dz <= proto.MyShopReach*proto.MyShopReach
}

// Reconcile offers when a stack changes or the original item leaves its slot.
// An empty, published market closes and disconnects its buyers.
func (s *Session) refreshMyShop() bool {
	listed := false
	for index, sl := range s.shop.slots {
		if !sl.used {
			continue
		}
		i := sl.itemIndex(s.items)
		if i < 0 {
			s.shop.slots[index] = shopSlot{}
			s.send(shopAck(opMyShopRemove, func(w *packet.Writer) { w.U8(uint8(index)) }), "unavailable market item removed")
		} else {
			listed = true
			s.shop.slots[index].item.Count = s.items[i].Count
		}
		if count := s.shop.slots[index].item.Count; s.shop.status == myShopStatusSelling && count != sl.item.Count {
			frame := shopAck(opMyShopStock, func(w *packet.Writer) { w.U8(uint8(index)); w.U8(uint8(min(count, 255))) })
			for _, buyer := range s.shop.buyers {
				buyer.sendShopUpdate(s.player, frame, false)
			}
		}
	}
	if !listed && s.shop.status == myShopStatusSelling {
		s.endMyShop()
	}
	return listed
}

// endMyShop closes the market for good (stand up, move, die, leave...).
func (s *Session) endMyShop() {
	if s.shop.status == myShopStatusNone || s.player == nil || s.char == nil {
		return
	}
	selling := s.shop.status == myShopStatusSelling
	if selling {
		s.closeMyShopBuyers()
		s.player.setShopTitle("")
		s.broadcastNear(shopBubbleFrame(s.char.ID, ""), "0x0523 market bubble off", true)
	}
	s.shop = myShop{}
	s.send(shopAck(opMyShopCancel, nil), "0x2302 market ended")
}

// enterMyShop is 0x2309: a buyer opens the seller's market.
func (s *Session) enterMyShop(id uint32) {
	reply := func(result uint8) {
		s.send(shopAck(opMyShopEnter, func(w *packet.Writer) { w.U8(result); w.U32(id) }), "0x2309 market enter")
	}
	seller := FindPlayerByID(id)
	if !s.canBrowseShop(seller) {
		reply(myShopEnterFailed)
		return
	}
	other := seller.owner
	if !other.stateMu.TryLock() {
		reply(myShopEnterFailed)
		return
	}
	defer other.stateMu.Unlock()
	if other.player != seller || other.char == nil || other.shop.status != myShopStatusSelling || !s.canBrowseShop(seller) || tradeFor(other.char.ID) != nil || !other.refreshMyShop() {
		reply(myShopEnterFailed)
		return
	}
	other.shop.buyers = slices.DeleteFunc(other.shop.buyers, func(buyer *ConnectedPlayer) bool { return buyer.currentViewedShop() != seller })
	if !slices.Contains(other.shop.buyers, s.player) {
		if len(other.shop.buyers) >= myShopBuyersMax {
			reply(myShopEnterTooMany)
			return
		}
		other.shop.buyers = append(other.shop.buyers, s.player)
	}
	s.player.setViewedShop(seller)
	reply(myShopEnterOK)
	s.send(other.myShopList(), "0x230b market list")
}

// myShopList is 0x230b: count and, per listed item, its index, price and record.
func (s *Session) myShopList() []byte {
	count := 0
	for _, sl := range s.shop.slots {
		if sl.used {
			count++
		}
	}
	w := packet.New(opMyShopList)
	w.U8(uint8(count))
	for idx, sl := range s.shop.slots {
		if !sl.used {
			continue
		}
		it := sl.item
		w.U8(uint8(idx))
		w.U32(sl.price)
		w.U8(it.Type)
		w.U8(it.TypeID)
		w.U8(uint8(min(it.Count, 255)))
		w.U16(it.Quality)
		w.Bytes(it.Gems[:])
		w.FixedString(it.CraftName, 21)
	}
	return w.MustFinish()
}

// buyFromMyShop is 0x230d: take count units of the listed item idx.
func (s *Session) buyFromMyShop(idx, count uint8) {
	reply := func(result uint8) {
		s.send(shopAck(opMyShopBuy, func(w *packet.Writer) { w.U8(result) }), fmt.Sprintf("0x230d market buy result %d", result))
	}
	seller := s.player.currentViewedShop()
	if !s.canBrowseShop(seller) || count == 0 || int(idx) >= myShopSlots {
		reply(myShopBuyInvalid)
		return
	}
	other := seller.owner
	if !other.stateMu.TryLock() {
		reply(myShopBuyInvalid)
		return
	}
	defer other.stateMu.Unlock()
	if other.player != seller || other.char == nil || other.shop.status != myShopStatusSelling || !s.canBrowseShop(seller) || tradeFor(other.char.ID) != nil {
		reply(myShopBuyInvalid)
		return
	}
	sl := other.shop.slots[idx]
	si := sl.itemIndex(other.items)
	if si < 0 || other.items[si].Count < uint16(count) {
		other.refreshMyShop()
		reply(myShopBuyInvalid)
		return
	}
	total := uint64(sl.price) * uint64(count)
	switch {
	case total > uint64(s.char.Money):
		reply(myShopBuyNoMoney)
		return
	case uint64(other.char.Money)+total > db.MoneyCap:
		reply(myShopBuySellerCap)
		return
	}
	a := db.CharacterChange{Before: *s.char, After: *s.char, ItemsBefore: s.items, ItemsAfter: append([]db.Item(nil), s.items...)}
	b := db.CharacterChange{Before: *other.char, After: *other.char, ItemsBefore: other.items, ItemsAfter: append([]db.Item(nil), other.items...)}
	a.After.Money = uint32(uint64(s.char.Money) - total)
	b.After.Money = uint32(uint64(other.char.Money) + total)
	tpl := b.ItemsAfter[si]
	b.ItemsAfter[si].Count -= uint16(count)
	sellerPacket := buildRemoveItem(&b.ItemsAfter[si], b.ItemsAfter[si].Count == 0)
	remaining := b.ItemsAfter[si].Count
	if remaining == 0 {
		b.ItemsAfter = append(b.ItemsAfter[:si], b.ItemsAfter[si+1:]...)
	}
	adds, left := planGiveItem(&a.ItemsAfter, &tpl, uint16(count), maxStackOf(tpl.Type, tpl.TypeID))
	if left != 0 {
		reply(myShopBuyNoRoom)
		return
	}
	if err := db.CommitGameplay([]db.CharacterChange{a, b}, nil); err != nil {
		s.logErr("market purchase transaction", err)
		reply(myShopBuyInvalid)
		return
	}
	*s.char, *other.char = a.After, b.After
	s.items, other.items = a.ItemsAfter, b.ItemsAfter
	for _, o := range []*Session{s, other} {
		o.syncPlayer()
		o.player.UpdateItems(o.items)
	}
	for _, pk := range adds {
		s.send(pk, "0x0205 market purchase item")
	}
	s.send(buildSetMoney(s.char.Money), "0x0213 market purchase money")
	other.send(sellerPacket, "0x0204 market sale item")
	other.send(buildSetMoney(other.char.Money), "0x0213 market sale money")
	other.goldActivity(b.Before.Money, "personal market sale")
	reply(myShopBuyOK)
	other.send(shopAck(opMyShopSold, func(w *packet.Writer) { w.U8(idx); w.U8(count); w.U32(uint32(total)) }), "0x2308 market sale")
	s.log.Printf("GX Server World market sale seller=%d buyer=%d item=%d/%d count=%d total=%d", other.char.ID, s.char.ID, tpl.Type, tpl.TypeID, count, total)
	other.refreshMyShop()
}
