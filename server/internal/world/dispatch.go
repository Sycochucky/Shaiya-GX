package world

import (
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/netx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

// dispatch is the protocol adapter. Application methods own validation and
// state changes; guarded storage operations own persistence and durability.
func (s *Session) dispatch(op packet.Opcode, p []byte) {
	if !s.phaseAllows(op) {
		return
	}
	if economicOpcode(op) && s.economy != nil && !s.economy.Allow("") {
		if s.conn != nil {
			netx.Abort(s.conn)
		}
		return
	}
	switch op {
	case packet.CharacterMobAutoAttack, packet.CharacterUserAutoAttack, packet.SkillUse, packet.SkillUseNoTarget,
		packet.UseItem, packet.UseVehicle, packet.CharacterMotion, packet.ReturnToTown:
		s.cancelTravelCast()
	}
	switch op {
	case packet.Opcode(proto.GXEvolutionCast):
		s.handleEvolutionCast(p)
	case packet.Opcode(proto.GXPartyAssist):
		s.handlePartyAssist(p)
	case packet.Opcode(proto.GXTargetEffectsQuery):
		s.handleTargetEffects(p)
	case packet.Opcode(proto.GXAuction):
		s.handleAuction(p)
	case packet.Opcode(proto.GXLadder):
		s.handleLadder(p)
	case 0x2202, 0x2203, 0x2205, 0x2206, 0x2209, 0x220a, 0x220b, 0x220c:
		s.handleSocial(op, p)
	case packet.AccountFaction:
		if len(p) >= 3 && p[2] <= FactionDark && (s.faction == FactionNone || s.faction == p[2]) {
			if s.faction == p[2] {
				s.send(buildAccountFaction(s.faction), "0x0109 account faction selected")
				return
			}
			if err := db.SaveCountry(int32(s.accountID), p[2]); err != nil {
				s.logErr("choose faction", err)
				return
			}
			s.faction = p[2]
			s.send(buildAccountFaction(s.faction), "0x0109 account faction selected")
		}
	case packet.CreateCharacter:
		s.handleCreateCharacter(p)
	case packet.DeleteCharacter:
		if len(p) >= 6 {
			id, _ := packet.U32(p[2:])
			ok, err := deleteCharacter(s.accountID, id)
			s.logErr("delete character", err)
			s.send(buildCharResult(packet.DeleteCharacter, ok, id), "0x0103 delete character")
		}
	case packet.RestoreCharacter:
		if len(p) >= 6 {
			id, _ := packet.U32(p[2:])
			ok, err := db.RestoreCharacter(s.accountID, id)
			s.logErr("restore character", err)
			s.send(buildCharResult(packet.RestoreCharacter, ok, id), "0x010F restore character")
		}
	case packet.RenameCharacter:
		s.handleRename(p)
	case packet.SelectCharacter:
		if len(p) >= 6 {
			id, _ := packet.U32(p[2:])
			s.selectCharacter(id)
		}
	case packet.ChatNormal, packet.ChatWorld, packet.ChatMap, packet.ChatNormalAdmin, packet.ChatWorldAdmin,
		packet.ChatShout, packet.ChatWorldShout:
		s.handleChat(op, p)
	case packet.ChatWhisper:
		s.handleWhisper(p)
	case packet.Opcode(proto.GXSummonRune):
		s.handleSummonRune(p)
	case packet.Opcode(proto.GXTeleportStone):
		s.handleTeleportStone(p)
	case packet.UseItem:
		s.handleUseItem(p)
	case packet.ChangeAppearance:
		s.handleChangeAppearance(p)
	case packet.EnchantAdd:
		s.handleEnchantAdd(p)
	case packet.GemAdd:
		s.handleGemAdd(p)
	case packet.GemRemove:
		s.handleGemRemove(p)
	case packet.CloakAdd:
		s.handleCloakAdd(p)
	case packet.CloakRemove:
		s.handleCloakRemove(p)
	case packet.ItemCompose, packet.ShopSell, packet.ShopBuy, packet.Opcode(proto.GXEconomyAction):
		s.economyAction(op, p)
	case packet.RemoveItem:
		s.handleDropItem(p)
	case packet.AddItem:
		s.handlePickupItem(p)
	case packet.ItemBagToBag:
		s.handleItemBagToBag(p)
	case packet.QuestStart, packet.QuestStartAlt:
		s.handleQuestStart(p)
	case packet.QuestEnd:
		s.handleQuestEnd(p)
	case packet.QuestChoose:
		s.handleQuestChoose(p)
	case packet.QuestAbandon:
		s.handleQuestAbandon(p)
	case packet.NpcTalk:
		if len(p) >= 6 {
			id, _ := packet.U32(p[2:])
			s.useNpc(id)
		}
	case packet.Opcode(proto.PvPRewardStatus), packet.Opcode(proto.PvPRewardKill), packet.Opcode(proto.PvPRewardDeath):
		s.handlePvPReward(op, p)
	case packet.PartyInvite, packet.PartyInviteName:
		s.handlePartyInvite(op, p)
	case packet.GXPartyInviteName:
		s.handlePartyInviteGlobal(p)
	case packet.PartyAnswer:
		s.handlePartyAnswer(p)
	case packet.PartyLeft, packet.PartyLeaveAlt:
		s.handlePartyLeave()
	case packet.PartyExpelled:
		s.handlePartyExpel(p)
	case packet.PartyLeader:
		s.handlePartyLeader(p)
	case packet.RaidConvert, packet.RaidDivision, packet.RaidAutoJoin, packet.RaidDisband:
		s.handleRaidSettings(op, p)
	case packet.RaidInvite, packet.RaidInviteName:
		s.handleRaidInvite(op, p)
	case packet.RaidAnswer:
		s.handleRaidAnswer(p)
	case packet.RaidAutoJoinName:
		s.handleRaidAutoJoin(p)
	case packet.RaidSwap, packet.RaidSubLeader, packet.RaidLeader:
		s.handleRaidRoles(op, p)
	case packet.RaidExpel:
		s.handleRaidExpel(p)
	case packet.ChatParty:
		if s.char != nil {
			if msg, err := readChatMessage(p); err == nil {
				if items, ok := s.chatItemSnapshots(p, 2); ok {
					s.partyChat(msg, items)
				}
			}
		}
	case packet.ChatRaid:
		if s.char != nil && s.player != nil {
			if msg, err := readChatMessage(p); err == nil {
				s.raidChat(msg)
			}
		}
	case packet.ChatGuild:
		if s.char != nil && s.player != nil && len(p) >= 3 && p[2] > 1 && p[2] <= 0x80 {
			if msg, err := readChatMessage(p); err == nil {
				if items, ok := s.chatItemSnapshots(p, 2); ok {
					s.guildChat(msg, items)
				}
			}
		}
	case packet.GuildCreate, packet.GuildCreateAgree, packet.GuildDisband, packet.GuildDisbandAgree,
		packet.GuildApply, packet.GuildJoinResult,
		packet.GuildLeave, packet.GuildExpel, packet.GuildStatus, packet.GuildRemark, packet.GuildEtin,
		packet.GuildNpcUpgrade, packet.GuildHouseBuy, packet.GuildDonate:
		s.handleGuild(op, p)
	case opMyShopBegin, opMyShopCancel, opMyShopOpen, opMyShopStop, opMyShopAdd, opMyShopRemove, opMyShopPrice, opMyShopEnter, opMyShopLeave, opMyShopBuy:
		s.handleMyShop(op, p)
	case packet.Gatekeeper:
		s.handleGatekeeper(p)
	case packet.ExchangeRequest:
		if len(p) >= 6 {
			id, _ := packet.U32(p[2:])
			s.handleTradeRequest(id)
		}
	case packet.ExchangeAnswer:
		if len(p) >= 3 {
			s.handleTradeAnswer(p[2] == 0)
		}
	case packet.ExchangeReady:
		if len(p) >= 3 {
			s.handleTradeReady(p[2])
		}
	case packet.ExchangeAdd:
		if len(p) >= 6 {
			r := packet.NewReader(p)
			bag, slot, count, dest := r.U8(), r.U8(), r.Count(), r.U8()
			if r.Err() == nil {
				s.handleTradeAdd(bag, slot, count, dest)
			}
		}
	case packet.ExchangeRemove:
		if len(p) >= 3 {
			s.handleTradeRemove(p[2])
		}
	case packet.ExchangeAddMoney:
		if len(p) >= 6 {
			v, _ := packet.U32(p[2:])
			s.handleTradeMoney(v)
		}
	case packet.ExchangeConfirm:
		if len(p) >= 3 {
			s.handleTradeConfirm(p[2] != 0)
		}
	case packet.CharacterMove:
		s.handleCharacterMove(p)
	case packet.CharacterMotion:
		s.handleMotion(p)
	case packet.CharacterEnteredPortal:
		if len(p) >= 3 {
			s.handleEnteredPortal(p[2])
		}
	case packet.CharacterEnteredMap:
		s.handleMapLoaded()
	case packet.LearnNewSkill:
		s.handleLearnSkill(p)
	case packet.SetStatPoint:
		s.handleSetStatPoint(p)
	case packet.CharacterSkillBar:
		s.handleSkillBar(p)
	case packet.AutoAttackStop, packet.TargetClear:
		s.autoAttack = false
	case 0x2401, 0x2402, 0x2406, 0x2409, 0x240A, 0x240B, 0x240C:
		s.handleDuel(op, p)
	case 0x2501, 0x2502:
		s.handleGuildDuel(op, p)
	case packet.ResurrectRequest:
		s.handleResurrect(p)
	case packet.ItemRepair:
		// Legacy repair requests are ignored; GX has no item durability.
	case packet.ReturnToTown:
		s.handleReturnToTown()
	case packet.UseVehicle:
		s.handleUseVehicle()
	case packet.Opcode(proto.VehicleInvite):
		s.handleVehicleInvite(p)
	case packet.Opcode(proto.VehicleAnswer):
		s.handleVehicleAnswer(p)
	case packet.Opcode(proto.VehicleLeave):
		s.dismount()
	case packet.CharacterMobAutoAttack:
		s.cancelMove()
		if len(p) >= 6 && s.char != nil && s.attackMovementAllowed() {
			id, _ := packet.U32(p[2:])
			s.autoAttack, s.autoAttackTarget, s.autoAttackUser = true, id, false
			if !s.autoAttackSwing(id) {
				s.autoAttack = false
			}
		}
	case packet.CharacterUserAutoAttack:
		// PacketPC 0x0502: only a hostile user starts the attack.
		s.cancelMove()
		if len(p) >= 6 && s.char != nil && s.attackMovementAllowed() {
			id, _ := packet.U32(p[2:])
			if s.hostileUser(id) != nil {
				s.autoAttack, s.autoAttackTarget, s.autoAttackUser = true, id, true
				if !s.userAttackSwing(id) {
					s.autoAttack = false
				}
			}
		}
	case packet.SkillUse, packet.SkillUseNoTarget:
		s.handleUseSkill(op, p)
	case packet.InspectEquipment:
		s.handleInspectEquipment(p)
	case packet.TargetMobHpUpdate, packet.TargetMobGetState:
		if len(p) >= 6 && s.char != nil {
			id, _ := packet.U32(p[2:])
			if m, ok := MobByID(s.char.MapID, id); ok {
				if op == packet.TargetMobHpUpdate {
					s.send(buildMobTargetHP(&m), "0x0305 mob target hp")
				} else {
					// Visibility already publishes movement. A target snapshot would interrupt the active segment.
					s.send(buildMobState(&m), "0x0304 mob state")
				}
			}
		}
	case packet.PointBuy:
		s.handlePointBuy(p)
	case packet.PointGiftSend:
		s.handleGiftSend(p)
	case packet.PointStoredItemGet:
		s.handleGiftClaim(p)
	case 0x7F09:
		s.handleGMMapTeleport(p)
	case packet.GXForgeQuote:
		s.handleForgeQuote(p)
	case packet.CashPoint:
		s.send(buildCashPoints(s.points), "0x2605 cash points")
	case packet.Ping:
		s.send(buildPing(), "0xA303 ping")
	case packet.Logout:
		s.handleLogout()
	}
}

// Phase admission is enforced before reaching application handlers. Character
// management cannot replace an actor that is already registered in the world.
func (s *Session) phaseAllows(op packet.Opcode) bool {
	if op == packet.Ping || op == packet.Logout {
		return true
	}
	management := op == packet.AccountFaction || op == packet.CreateCharacter || op == packet.DeleteCharacter || op == packet.RestoreCharacter || op == packet.RenameCharacter || op == packet.SelectCharacter
	if s.char == nil {
		return management
	}
	if management {
		return false
	}
	if s.awaitingMapLoad {
		return op == packet.CharacterEnteredMap || op == packet.CharacterSkillBar
	}
	return op != packet.CharacterEnteredMap
}

func economicOpcode(op packet.Opcode) bool {
	switch op {
	case packet.AccountFaction, packet.CreateCharacter, packet.DeleteCharacter, packet.RestoreCharacter, packet.RenameCharacter, packet.SelectCharacter,
		packet.Opcode(proto.GXAuction), packet.Opcode(proto.GXEconomyAction), packet.Opcode(proto.GXTeleportStone), packet.UseItem, packet.ChangeAppearance, packet.EnchantAdd, packet.GemAdd, packet.GemRemove, packet.CloakAdd, packet.CloakRemove, packet.ItemCompose, packet.RemoveItem, packet.AddItem, packet.ItemBagToBag, packet.ShopSell, packet.ShopBuy,
		packet.QuestStart, packet.QuestStartAlt, packet.QuestEnd, packet.QuestChoose, packet.QuestAbandon, packet.LearnNewSkill, packet.SetStatPoint, packet.PointBuy, packet.PointGiftSend, packet.PointStoredItemGet:
		return true
	}
	if op >= opMyShopBegin && op <= opMyShopBuy {
		return true // personal market purchases move items and gold
	}
	return op >= 0x0d00 && op <= 0x0dff // guild changes include permissions and shared storage
}
