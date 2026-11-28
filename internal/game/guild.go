package game

import "github.com/Spelunkern/Shaiya-GX/internal/proto"

// GuildState is populated only by server packets; roster batches can arrive
// before the country-wide guild directory at world entry.
type GuildState struct {
	ID                 uint32
	Name               string
	Rank               uint8
	Directory          map[uint32]proto.GuildInfo
	DirectoryReady     bool
	Ranking            map[uint32]proto.GuildRankingEntry
	RankingReady       bool
	RankingNotice      proto.GuildRankingNoticeInfo
	RankingPoints      proto.GuildRankingPointsInfo
	RankingPointsKnown bool
	Etin               uint32
	Npcs               map[uint8]proto.GuildNpc // house NPC slots (0x0D24, 0x0D25)
	NpcResult          proto.GuildNpcUpgradeResult
	NpcSerial          uint64
	HouseResult        uint8 // 0x0D34
	HouseBought        bool
	HouseGM            bool // GX: GM-founded guild (no upkeep)
	// GX 0x0D33: the guild's current house bonuses as whole percentages: shop
	// discount by type (weapon, armor, general, potion), teleport fee discount,
	// gem link success bonus and gem link cost discount.
	HouseShopDiscount                                    [4]uint8
	HouseGateDiscount, HouseLinkBonus, HouseLinkDiscount uint8
	HouseSerial                                          uint64
	Donated                                              int // items consumed by the last 0x0D35
	DonateSerial                                         uint64
	EtinKnown                                            bool
	Remark                                               string
	RemarkKnown                                          bool
	RemarkResult                                         uint8
	RemarkSerial                                         uint64
	Members                                              map[uint32]proto.GuildMember
	Applicants                                           map[uint32]proto.GuildCandidate
	Bank                                                 map[uint8]proto.InventoryItem
	BankResult                                           proto.GuildBankResultInfo
	BankResultSerial                                     uint64
	ApplyResult                                          uint8
	ApplySerial                                          uint64
	JoinResult                                           uint8
	JoinSerial                                           uint64
	CreateResult                                         uint8
	CreateSerial                                         uint64
	CreateVote                                           proto.GuildCreateVote
	VoteSerial                                           uint64
	CreateFailure                                        uint8
	FailureSerial                                        uint64
	LeaveResult                                          uint8
	LeaveSerial                                          uint64
	DisbandResult                                        uint8
	DisbandSerial                                        uint64
	DisbandMaster                                        uint32
	DisbandVoteSerial                                    uint64
	DisbandedSerial                                      uint64
}

func (s *State) clearVisibleGuild(id uint32) {
	if id == 0 {
		return
	}
	for _, e := range s.Entities {
		if e.Kind == KindUser && e.GuildID == id {
			e.GuildID = 0
		}
	}
}

func (s *State) setVisibleGuildMember(id, guild uint32) {
	if guild == 0 {
		return
	}
	if e := s.Entities[id]; e != nil && e.Kind == KindUser {
		e.GuildID = guild
	}
}

func (s *State) applyGuild(p []byte) error {
	g := &s.Guild
	switch proto.Op(p) {
	case proto.GuildEtin:
		v, err := proto.DecodeU32(p)
		if err != nil {
			return err
		}
		g.Etin, g.EtinKnown = v, true
	case proto.GuildNpcList:
		v, err := proto.DecodeGuildNpcList(p)
		if err != nil {
			return err
		}
		if g.Npcs == nil {
			g.Npcs = map[uint8]proto.GuildNpc{}
		}
		for _, n := range v {
			if n.Slot < 8 {
				g.Npcs[n.Slot] = n
			}
		}
	case proto.GuildNpcUpgrade:
		v, err := proto.DecodeGuildNpcUpgrade(p)
		if err != nil {
			return err
		}
		g.NpcResult = v
		g.NpcSerial++
		if v.Result == 0 && v.Npc.Slot < 8 {
			if g.Npcs == nil {
				g.Npcs = map[uint8]proto.GuildNpc{}
			}
			g.Npcs[v.Npc.Slot] = v.Npc
		}
	case proto.GuildHouseKeepEtin:
		// GX: sent only for a guild that owns its house (0x0D33 has no EP5.4 reader).
		v, err := proto.DecodeU32(p)
		if err != nil {
			return err
		}
		g.HouseBought, g.HouseGM = true, v&(1<<31) != 0
		if len(p) >= 6+7 {
			copy(g.HouseShopDiscount[:], p[6:10])
			g.HouseGateDiscount, g.HouseLinkBonus, g.HouseLinkDiscount = p[10], p[11], p[12]
		}
	case proto.GuildHouseBuy:
		v, err := proto.DecodeGuildHouseBuy(p)
		if err != nil {
			return err
		}
		s.Details.Money = v.Money
		g.HouseResult = v.Result
		g.HouseBought = g.HouseBought || v.Result == 0 || v.Result == 1
		g.HouseSerial++
	case proto.GuildDonate:
		v, err := proto.DecodeGuildDonation(p)
		if err != nil {
			return err
		}
		for _, pos := range v {
			delete(s.Inventory, pos)
		}
		g.Donated = len(v)
		g.DonateSerial++
	case proto.GuildRemark:
		v, err := proto.DecodeGuildRemark(p)
		if err != nil {
			return err
		}
		g.RemarkResult = v.Result
		g.RemarkSerial++
		if v.Result == 1 {
			g.Remark, g.RemarkKnown = v.Remark, true
			if info, ok := g.Directory[g.ID]; ok {
				info.Remark = v.Remark
				g.Directory[g.ID] = info
			}
		}
	case proto.GuildRankingNotice:
		v, err := proto.DecodeGuildRankingNotice(p)
		if err != nil {
			return err
		}
		g.RankingNotice = v
	case proto.GuildRankingPoints:
		v, err := proto.DecodeGuildRankingPoints(p)
		if err != nil {
			return err
		}
		g.RankingPoints, g.RankingPointsKnown = v, true
	case proto.GuildRankingList:
		rows, err := proto.DecodeGuildRankingList(p)
		if err != nil {
			return err
		}
		g.Ranking = make(map[uint32]proto.GuildRankingEntry, len(rows))
		for _, row := range rows {
			g.Ranking[row.Guild] = row
			if info, ok := g.Directory[row.Guild]; ok {
				info.Rank, info.Points = row.Rank, row.Points
				g.Directory[row.Guild] = info
			}
		}
		g.RankingReady = true
	case proto.GuildBankList:
		items, err := proto.DecodeGuildBankList(p)
		if err != nil {
			return err
		}
		if g.Bank == nil {
			g.Bank = make(map[uint8]proto.InventoryItem)
		}
		for _, item := range items {
			g.Bank[item.Slot] = item
		}
	case proto.GuildBankPlaced, proto.GuildBankTaken:
		event, err := proto.DecodeGuildBankEvent(p)
		if err != nil {
			return err
		}
		if proto.Op(p) == proto.GuildBankTaken {
			delete(g.Bank, event.Item.Slot)
		} else {
			if g.Bank == nil {
				g.Bank = make(map[uint8]proto.InventoryItem)
			}
			g.Bank[event.Item.Slot] = event.Item
		}
	case proto.GuildBankResult:
		result, err := proto.DecodeGuildBankResult(p)
		if err != nil {
			return err
		}
		g.BankResult = result
		g.BankResultSerial++
	case proto.GuildListBegin:
		g.Directory = make(map[uint32]proto.GuildInfo)
		g.DirectoryReady = false
	case proto.GuildListEnd:
		g.DirectoryReady = true
	case proto.GuildList:
		list, err := proto.DecodeGuildList(p)
		if err != nil {
			return err
		}
		if g.Directory == nil {
			g.Directory = make(map[uint32]proto.GuildInfo)
		}
		for _, v := range list {
			g.Directory[v.ID] = v
			if v.ID == g.ID {
				g.Remark, g.RemarkKnown = v.Remark, true
			}
		}
	case proto.GuildNew:
		v, err := proto.DecodeGuildInfo(p)
		if err != nil {
			return err
		}
		if g.Directory == nil {
			g.Directory = make(map[uint32]proto.GuildInfo)
		}
		g.Directory[v.ID] = v
		if v.ID == g.ID {
			g.Remark, g.RemarkKnown = v.Remark, true
		}
	case proto.GuildDisbanded:
		id, err := proto.DecodeU32(p)
		if err != nil {
			return err
		}
		delete(g.Directory, id)
		s.clearVisibleGuild(id)
	case proto.GuildOnlineList, proto.GuildOfflineList:
		members, err := proto.DecodeGuildMembers(p, proto.Op(p) == proto.GuildOnlineList)
		if err != nil {
			return err
		}
		if g.Members == nil {
			g.Members = make(map[uint32]proto.GuildMember)
		}
		for _, m := range members {
			g.Members[m.ID] = m
			s.setVisibleGuildMember(m.ID, g.ID)
			if m.ID == s.SelfID {
				g.Rank = m.Rank
			}
		}
	case proto.GuildMemberAdded:
		m, err := proto.DecodeGuildMemberAdded(p)
		if err != nil {
			return err
		}
		if g.Members == nil {
			g.Members = make(map[uint32]proto.GuildMember)
		}
		g.Members[m.ID] = m
		s.setVisibleGuildMember(m.ID, g.ID)
		if m.ID == s.SelfID {
			g.Rank = m.Rank
		}
	case proto.GuildApplicantList:
		applicants, err := proto.DecodeGuildApplicants(p)
		if err != nil {
			return err
		}
		if g.Applicants == nil {
			g.Applicants = make(map[uint32]proto.GuildCandidate)
		}
		for _, candidate := range applicants {
			g.Applicants[candidate.ID] = candidate
		}
	case proto.GuildApplicant:
		candidate, err := proto.DecodeGuildApplicant(p)
		if err != nil {
			return err
		}
		if g.Applicants == nil {
			g.Applicants = make(map[uint32]proto.GuildCandidate)
		}
		g.Applicants[candidate.ID] = candidate
	case proto.GuildApplicantRemoved:
		id, err := proto.DecodeU32(p)
		if err != nil {
			return err
		}
		delete(g.Applicants, id)
	case proto.GuildApply:
		result, err := proto.DecodeU8(p)
		if err != nil {
			return err
		}
		g.ApplyResult = result
		g.ApplySerial++
	case proto.GuildStatus:
		r := proto.NewReader(p)
		mode, id := r.U8(), r.U32()
		if err := r.Err(); err != nil {
			return err
		}
		if mode == 101 {
			g.DisbandedSerial++
			s.clearVisibleGuild(g.ID)
			g.ID, g.Name, g.Rank, g.Members = 0, "", 0, nil
			g.Bank = nil
			g.Etin, g.EtinKnown, g.Remark, g.RemarkKnown, g.Npcs, g.HouseBought, g.HouseGM, g.HouseShopDiscount, g.HouseGateDiscount, g.HouseLinkBonus, g.HouseLinkDiscount = 0, false, "", false, nil, false, false, [4]uint8{}, 0, 0, 0
			g.Applicants = nil
			s.Details.Guild = ""
			return nil
		}
		if mode == 102 || mode == 103 {
			if e := s.Entities[id]; e != nil && e.Kind == KindUser && e.GuildID == g.ID {
				e.GuildID = 0
			}
		}
		m, ok := g.Members[id]
		if !ok {
			return nil
		}
		switch {
		case mode == 102 || mode == 103:
			delete(g.Members, id)
			return nil
		case mode == 104:
			m.Online = true
		case mode == 105:
			m.Online = false
		case mode > 200:
			m.Rank = mode - 200
		case mode > 0 && mode <= 100:
			m.Level = uint16(mode)
		}
		g.Members[id] = m
		if id == s.SelfID {
			g.Rank = m.Rank
		}
	case proto.GuildJoinResult:
		v, err := proto.DecodeGuildJoin(p)
		if err != nil {
			return err
		}
		g.JoinResult = v.Result
		g.JoinSerial++
		if v.Result == 1 {
			g.ID, g.Name, g.Rank, g.Members = v.ID, v.Name, v.Rank, nil
			g.Etin, g.EtinKnown, g.Remark, g.RemarkKnown, g.Npcs, g.HouseBought, g.HouseGM, g.HouseShopDiscount, g.HouseGateDiscount, g.HouseLinkBonus, g.HouseLinkDiscount = 0, false, "", false, nil, false, false, [4]uint8{}, 0, 0, 0
			if info, ok := g.Directory[v.ID]; ok {
				g.Remark, g.RemarkKnown = info.Remark, true
			}
			g.Bank = nil
			g.Applicants = nil
			s.Details.Guild = v.Name
		}
	case proto.GuildCreate:
		v, remark, err := proto.DecodeGuildCreate(p)
		if err != nil {
			return err
		}
		g.CreateResult = v.Result
		g.CreateSerial++
		if v.Result == 1 {
			g.ID, g.Name, g.Rank, g.Members = v.ID, v.Name, v.Rank, nil
			g.Etin, g.EtinKnown, g.Remark, g.RemarkKnown = 0, false, remark, true
			g.Bank = nil
			g.Applicants = nil
			s.Details.Guild = v.Name
		}
	case proto.GuildCreateAgree:
		vote, err := proto.DecodeGuildCreateVote(p)
		if err != nil {
			return err
		}
		g.CreateVote = vote
		g.VoteSerial++
	case proto.GuildCreateFailed:
		result, err := proto.DecodeGuildCreateFailure(p)
		if err != nil {
			return err
		}
		g.CreateFailure = result
		g.FailureSerial++
	case proto.GuildLeave:
		v, err := proto.DecodeU8(p)
		if err != nil {
			return err
		}
		g.LeaveResult = v
		g.LeaveSerial++
		if v == 1 {
			g.ID, g.Name, g.Rank, g.Members = 0, "", 0, nil
			g.Etin, g.EtinKnown, g.Remark, g.RemarkKnown, g.Npcs, g.HouseBought, g.HouseGM, g.HouseShopDiscount, g.HouseGateDiscount, g.HouseLinkBonus, g.HouseLinkDiscount = 0, false, "", false, nil, false, false, [4]uint8{}, 0, 0, 0
			g.Bank = nil
			g.Applicants = nil
			s.Details.Guild = ""
		}
	case proto.GuildDisband:
		v, err := proto.DecodeU8(p)
		if err != nil {
			return err
		}
		g.DisbandResult = v
		g.DisbandSerial++
	case proto.GuildDisbandAgree:
		id, err := proto.DecodeU32(p)
		if err != nil {
			return err
		}
		g.DisbandMaster = id
		g.DisbandVoteSerial++
	case proto.GuildExpel:
		reply, err := proto.DecodePartyReply(p)
		if err != nil {
			return err
		}
		if reply.Result == 1 && reply.ID == s.SelfID {
			g.ID, g.Name, g.Rank, g.Members = 0, "", 0, nil
			g.Etin, g.EtinKnown, g.Remark, g.RemarkKnown, g.Npcs, g.HouseBought, g.HouseGM, g.HouseShopDiscount, g.HouseGateDiscount, g.HouseLinkBonus, g.HouseLinkDiscount = 0, false, "", false, nil, false, false, [4]uint8{}, 0, 0, 0
			g.Bank = nil
			g.Applicants = nil
			s.Details.Guild = ""
		}
	}
	return nil
}
