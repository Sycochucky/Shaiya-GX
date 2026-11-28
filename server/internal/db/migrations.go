package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/blessingrules"
	"github.com/Spelunkern/Shaiya-GX/internal/enchantrules"
	"github.com/Spelunkern/Shaiya-GX/internal/itemboxes"
	"github.com/Spelunkern/Shaiya-GX/internal/maprules"
	"github.com/Spelunkern/Shaiya-GX/internal/mobaimessages"
	"github.com/Spelunkern/Shaiya-GX/internal/npcquest"
	"github.com/Spelunkern/Shaiya-GX/internal/teleportrunes"
	"github.com/Spelunkern/Shaiya-GX/internal/worlddrops"
)

type migration struct {
	Version   int
	Name, SQL string
}

//go:embed experience.sql
var experienceMigration string

//go:embed npcskills.sql
var npcSkillsMigration string

//go:embed auction.sql
var auctionMigration string

//go:embed prune_legacy.sql
var pruneLegacyMigration string

//go:embed percentage_skills.sql
var percentageSkillsMigration string

//go:embed restore_consumables.sql
var restoreConsumablesMigration string

//go:embed remove_growth_modes.sql
var removeGrowthModesMigration string

//go:embed consumable_stacks.sql
var consumableStacksMigration string

//go:embed forge_hammers.sql
var forgeHammersMigration string

var migrations = []migration{
	{1, "atomic_session_handoff", `CREATE TABLE IF NOT EXISTS GX_SessionHandoff(UserUID INTEGER PRIMARY KEY,Identity BLOB NOT NULL,AESMaterial BLOB NOT NULL,SessionID INTEGER NOT NULL,ExpiresAt INTEGER NOT NULL)`},
	{2, "indexed_account_lookup", `CREATE INDEX IF NOT EXISTS GX_Users_UserID ON PS_UserData_Users_Master(UserID)`},
	{3, "credential_revocation", `CREATE TABLE GX_AccountSecurity(UserUID INTEGER PRIMARY KEY,Version INTEGER NOT NULL DEFAULT 1); ALTER TABLE GX_SessionHandoff ADD COLUMN CredentialVersion INTEGER NOT NULL DEFAULT 1`},
	{4, "warehouse_lookup", `CREATE INDEX GX_Warehouse_User ON PS_GameData_UserStoredItems(ServerID,UserUID,ItemUID)`},
	{5, "active_account_ownership", `CREATE TABLE GX_ActiveSessions(UserUID INTEGER PRIMARY KEY,SessionID INTEGER NOT NULL)`},
	{6, "character_teleport_points", `CREATE TABLE GX_TeleportPoints(
		CharID INTEGER NOT NULL REFERENCES PS_GameData_Chars(CharID) ON DELETE CASCADE,
		Slot INTEGER NOT NULL CHECK(Slot BETWEEN 0 AND 3),
		Map INTEGER NOT NULL CHECK(Map BETWEEN 0 AND 65535),
		X REAL NOT NULL, Y REAL NOT NULL, Z REAL NOT NULL,
		PRIMARY KEY(CharID, Slot))`},
	{7, "skill_catalog_cleanup", `-- Dead skill levels (placeholders such as "...ED" / "Not Available.") of skills that have
		-- working levels. Fully dead skills and every working level stay.
		DELETE FROM PS_GameDefs_Skills WHERE SkillLevel BETWEEN 1 AND 9 AND (SkillID,SkillLevel) IN (VALUES
		(174,3),(201,2),(201,3),(202,2),(202,3),(203,2),(203,3),(204,2),(204,3),(205,2),
		(205,3),(206,2),(206,3),(210,2),(210,3),(211,2),(211,3),(212,2),(212,3),(213,2),
		(213,3),(214,2),(214,3),(215,2),(215,3),(222,2),(222,3),(223,2),(223,3),(224,2),
		(224,3),(225,2),(225,3),(226,2),(226,3),(227,2),(227,3),(228,2),(228,3),(229,2),
		(229,3),(230,2),(230,3),(231,2),(231,3),(232,2),(232,3),(233,2),(233,3),(234,2),
		(234,3),(235,2),(235,3),(236,2),(236,3),(237,2),(237,3),(238,2),(238,3),(239,2),
		(239,3),(240,2),(240,3),(241,2),(241,3),(242,2),(242,3),(243,2),(243,3),(244,2),
		(244,3),(245,2),(245,3),(246,2),(246,3),(247,2),(247,3),(248,2),(248,3),(249,2),
		(249,3),(250,3),(251,2),(251,3),(252,2),(252,3),(253,2),(253,3),(254,2),(254,3),
		(255,2),(255,3),(261,2),(261,3),(262,2),(262,3),(266,3),(267,2),(267,3),(268,3),
		(269,3),(270,3),(271,2),(271,3),(272,2),(272,3),(273,3),(302,3),(309,3),(310,3),
		(320,2),(320,3),(321,3),(322,3),(323,3),(324,3));
		-- The chain mechanic is not used: its previous-skill relation goes (Shadow Attack and
		-- Shadow Sleep keep the Stealth requirement in their own previous-skill field).
		UPDATE PS_GameDefs_Skills SET PrevSkillID=0 WHERE SkillLevel<100 AND SkillID IN (61,64,79,94,104,135,145,150,161,302,311,312,313);
		-- Aggravation's time damage is not part of the skill.
		UPDATE PS_GameDefs_Skills SET TimeDamageHP=0,TimeDamageSP=0,TimeDamageMP=0 WHERE SkillID=150 AND SkillLevel<100;
		-- Detail 25 was a common attack with extra HP/MP/SP damage; AddDamage is read from its own
		-- fields, so the standard attack detail 21 behaves identically.
		UPDATE PS_GameDefs_Skills SET TypeDetail=21 WHERE TypeDetail=25 AND TypeEffect=1 AND SkillLevel<100`},
	{8, "skill_catalog_dead_levels", `-- More dead levels, recognised by their description ("Not Available.", "Skill not available
		-- currently"): placeholders that still carry copied effect fields. Levels that consumables
		-- reference (Remedy Lv3, Mini Sun EXP Stone 2, Little Elkid Clone) are working and stay.
		DELETE FROM PS_GameDefs_Skills WHERE SkillLevel BETWEEN 1 AND 9 AND (SkillID,SkillLevel) IN (VALUES
		(15,3),(17,3),(19,3),(20,3),(21,3),(22,3),(100,2),(100,3),(109,3),(125,2),
		(125,3),(173,2),(173,3),(175,3),(176,3),(177,3),(178,3),(179,3),(180,3),(181,3),
		(182,3),(183,3),(184,3),(185,3),(316,3),(317,3),(318,3),(319,3))`},
	{9, "skill_previous_requirement", `-- PrevSkillID is back: the skill it names must be cooling down for this one to be usable.
		UPDATE PS_GameDefs_Skills SET PrevSkillID=CASE SkillID
			WHEN 61 THEN 104 WHEN 64 THEN 150 WHEN 79 THEN 135 WHEN 94 THEN 79 WHEN 104 THEN 134
			WHEN 135 THEN 134 WHEN 145 THEN 134 WHEN 150 THEN 145 WHEN 161 THEN 61 WHEN 302 THEN 39
			WHEN 311 THEN 150 WHEN 312 THEN 61 WHEN 313 THEN 79 END
		WHERE SkillLevel<100 AND SkillID IN (61,64,79,94,104,135,145,150,161,302,311,312,313)`},
	{10, "item_enchant_rules", enchantrules.LegacyCreateSQL},
	{11, "enchant_target_levels", enchantrules.UpgradeSQL},
	{12, "mob_ai_messages", mobaimessages.MigrationSQL},
	{13, "goddess_blessing", blessingrules.MigrationSQL},
	{14, "universal_goddess_blessing", blessingrules.MergeSQL},
	{15, "world_drops", worlddrops.MigrationSQL},
	{16, "item_boxes", itemboxes.MigrationSQL},
	{17, "npc_quest", npcquest.MigrationSQL},
	{18, "map_properties", maprules.MigrationSQL},
	{19, "unified_experience", experienceMigration},
	{20, "separate_npc_skills", npcSkillsMigration},
	{21, "server_notices", `CREATE TABLE PS_GameDefs_ServerNotices (
		"Index" INTEGER PRIMARY KEY CHECK ("Index" BETWEEN 1 AND 4294967295),
		Message TEXT NOT NULL CHECK (length(trim(Message)) > 0 AND length(CAST(Message AS BLOB)) <= 127),
		Frequency INTEGER NOT NULL DEFAULT 0 CHECK (Frequency BETWEEN 0 AND 5256000),
		FixedHour1 TEXT, FixedHour2 TEXT, FixedHour3 TEXT, FixedHour4 TEXT, FixedHour5 TEXT
	) STRICT;
	INSERT INTO PS_GameDefs_ServerNotices ("Index", Message, Frequency)
	VALUES (1, 'GX Automatic Notice - Test', 60);`},
	{22, "teleport_runes", teleportrunes.MigrationSQL},
	{23, "auction_house", auctionMigration},
	{24, "runtime_write_indexes", `
		CREATE INDEX GX_Handoff_Expiry ON GX_SessionHandoff(ExpiresAt);
		CREATE INDEX GX_CharEvents_Char ON PS_GameData_CharEvents(CharID);
		CREATE INDEX GX_GuildBank_Guild ON PS_GameData_GuildStoredItems(GuildID,ItemUID);
		CREATE INDEX GX_Friends_Reverse ON PS_GameData_FriendChars(FriendID);
		CREATE INDEX GX_Blocked_Reverse ON PS_GameData_BanChars(BanID);
		CREATE INDEX GX_Chars_LiveName ON PS_GameData_Chars(CharName COLLATE NOCASE) WHERE Del=0;
	`},
	{25, "indexed_legacy_row_identity", `
		-- The imported RowID triggers preserve the exporter identity values.
		-- Their MAX(RowID) lookup must not scan all previous player rows.
		CREATE INDEX GX_Identity_Users_Master ON PS_UserData_Users_Master(RowID);
		CREATE INDEX GX_Identity_UserOnlineTimeLog ON PS_UserData_UserOnlineTimeLog(RowID);
		CREATE INDEX GX_Identity_BanChars ON PS_GameData_BanChars(RowID);
		CREATE INDEX GX_Identity_CharApplySkills ON PS_GameData_CharApplySkills(RowID);
		CREATE INDEX GX_Identity_CharItems ON PS_GameData_CharItems(RowID);
		CREATE INDEX GX_Identity_CharQuests ON PS_GameData_CharQuests(RowID);
		CREATE INDEX GX_Identity_CharQuickSlots ON PS_GameData_CharQuickSlots(RowID);
		CREATE INDEX GX_Identity_CharSkills ON PS_GameData_CharSkills(RowID);
		CREATE INDEX GX_Identity_FriendChars ON PS_GameData_FriendChars(RowID);
		CREATE INDEX GX_Identity_Guilds ON PS_GameData_Guilds(RowID);
		CREATE INDEX GX_Identity_GuildStoredItems ON PS_GameData_GuildStoredItems(RowID);
		CREATE INDEX GX_Identity_Keeps ON PS_GameData_Keeps(RowID);
		CREATE INDEX GX_Identity_UserMaxGrow ON PS_GameData_UserMaxGrow(RowID);
		CREATE INDEX GX_Identity_UserStoredItems ON PS_GameData_UserStoredItems(RowID);
		CREATE INDEX GX_Identity_UserStoredMoney ON PS_GameData_UserStoredMoney(RowID);
	`},
	{pruneLegacyVersion, "remove_inactive_sqlite_tables", pruneLegacyMigration},
	{27, "quest_timing", `
		ALTER TABLE PS_GameData_CharQuests ADD COLUMN StartedAt INTEGER NOT NULL DEFAULT 0;
		CREATE TABLE GX_QuestStartTerms (
			CharID INTEGER NOT NULL REFERENCES PS_GameData_Chars(CharID) ON DELETE CASCADE,
			QuestID INTEGER NOT NULL,
			AvailableAt INTEGER NOT NULL,
			PRIMARY KEY(CharID, QuestID)
		) WITHOUT ROWID;
	`},
	{28, "retire_interpretation_and_backpack_upgrade", `
		DELETE FROM PS_GameData_CharQuickSlots AS q WHERE q.Bag=100
		AND EXISTS (SELECT 1 FROM PS_GameData_CharSkills AS s
			WHERE s.CharID=q.CharID AND s.Number=q.Number AND s.SkillID IN (3,4))
		AND NOT EXISTS (SELECT 1 FROM PS_GameData_CharSkills AS s
			WHERE s.CharID=q.CharID AND s.Number=q.Number AND s.Del=0 AND s.SkillID NOT IN (3,4));
		DELETE FROM PS_GameData_CharSkills WHERE SkillID IN (3,4);
		DELETE FROM PS_GameData_CharApplySkills WHERE SkillID IN (3,4) AND SkillLevel<>100;
		DELETE FROM PS_GameDefs_Skills WHERE SkillID IN (3,4);
	`},
	{29, "enable_soul_burn", `
		UPDATE PS_GameDefs_Skills
		SET Defensemage=1, SkillPoint=3, TypeShow=4, TypeEffect=6, SuccessType=1, SuccessValue=100, KeepTime=0
		WHERE SkillID=125 AND SkillLevel=1 AND TypeDetail=122;
		UPDATE PS_GameData_CharSkills SET SkillLevel=1 WHERE SkillID=125 AND SkillLevel IN (2,3)
		AND EXISTS (SELECT 1 FROM PS_GameDefs_Skills WHERE SkillID=125 AND SkillLevel=1 AND TypeDetail=122);
		DELETE FROM PS_GameDefs_Skills WHERE SkillID=125 AND SkillLevel IN (2,3) AND TypeDetail=122;
	`},
	{30, "repair_special_npc_skills", `
		UPDATE PS_GameDefs_NpcSkills SET ChangeType=101
		WHERE SkillID=100 AND Rank=1 AND SkillName='Porcine Party' AND TypeEffect=3 AND TypeDetail=90 AND ChangeType=7 AND ChangeLevel=60;
		UPDATE PS_GameDefs_NpcSkills SET TypeEffect=3,TypeDetail=64,DamageType=4,DamageHP=25,TimeDamageType=0,TimeDamageHP=0
		WHERE SkillID=172 AND Rank=1 AND SkillName='Ice Stun' AND TypeEffect=7 AND TypeDetail=61;
		UPDATE PS_GameDefs_NpcSkills SET TypeEffect=3,KeepTime=20
		WHERE SkillID=344 AND Rank=1 AND SkillName='Wide Death Sentence' AND TypeEffect=1 AND TypeDetail=63 AND KeepTime=0;
	`},
	{31, "retire_warehouse_and_drop_protection_skills", `
		DELETE FROM PS_GameData_CharQuickSlots AS q WHERE q.Bag=100
		AND EXISTS (SELECT 1 FROM PS_GameData_CharSkills AS s
			WHERE s.CharID=q.CharID AND s.Number=q.Number AND s.SkillID IN (234,236,253,254))
		AND NOT EXISTS (SELECT 1 FROM PS_GameData_CharSkills AS s
			WHERE s.CharID=q.CharID AND s.Number=q.Number AND s.Del=0 AND s.SkillID NOT IN (234,236,253,254));
		DELETE FROM PS_GameData_CharSkills WHERE SkillID IN (234,236,253,254);
		DELETE FROM PS_GameData_CharApplySkills WHERE SkillLevel<>100 AND SkillID IN (77,107,127,234,236,253,254);
		DELETE FROM PS_GameDefs_Skills WHERE SkillID IN (234,236,253,254);
	`},
	{32, "retire_obsolete_items_and_preserve_reset_cooldowns", `
		CREATE TABLE IF NOT EXISTS GX_ResetSkillCooldowns (
			CharID INTEGER NOT NULL, SkillID INTEGER NOT NULL, ExpiresAt INTEGER NOT NULL,
			PRIMARY KEY(CharID, SkillID)
		) WITHOUT ROWID;
		DELETE FROM PS_GameDefs_Items WHERE ItemID IN (30071,30072,30073,30074,30075,30076,30077,100041,100064,100066,100069,100075,100077,100139,100174);
		WITH retired(ItemID) AS (VALUES (30071),(30072),(30073),(30074),(30075),(30076),(30077),(100041),(100064),(100066),(100069),(100075),(100077),(100139),(100174))
		DELETE FROM PS_GameDefs_ProductList
		WHERE EXISTS (SELECT 1 FROM retired WHERE retired.ItemID IN (
			ItemID1,ItemID2,ItemID3,ItemID4,ItemID5,ItemID6,ItemID7,ItemID8,ItemID9,ItemID10,ItemID11,ItemID12,
			ItemID13,ItemID14,ItemID15,ItemID16,ItemID17,ItemID18,ItemID19,ItemID20,ItemID21,ItemID22,ItemID23,ItemID24
		));
	`},
	{33, "explicit_experience_and_combat_chance_percentages", percentageSkillsMigration},
	{34, "repair_consumable_links_and_retire_resurrection_items", restoreConsumablesMigration},
	{35, "remove_growth_modes", removeGrowthModesMigration},
	{36, "stack_all_consumables", consumableStacksMigration},
	{37, "enhance_wind_spin_spear", `-- The description lists spears; the weapon flags did not (real type 6).
		UPDATE PS_GameDefs_Skills SET NeedWeapon6=1 WHERE SkillID=156 AND SkillLevel<100`},
	{38, "repair_remaining_consumable_skill_links", `-- Expand known native byte aliases without changing other Range fields.
		UPDATE PS_GameDefs_Items SET Range=Range+256
		WHERE (ItemID, Range) IN (VALUES
			(44089,12),(44090,13),(44091,14),(44152,12),(44153,14),(44154,13),
			(44213,17),(44214,17),(44217,18),(44218,18),(44219,19),(44220,19),(44221,18),(44222,19),
			(100114,6),(100118,7),(100119,7),(100120,7),(100121,8),(100122,8),
			(100123,8),(100124,9),(100125,9),(100126,9),(100134,15),(100200,18),(100201,19))
		  AND AttackTime BETWEEN 1 AND 3 AND Special IN (0,50);
		-- Item 44/242 still references the original 24-hour, 50% EXP buff.
		UPDATE PS_GameDefs_Skills
		SET SkillName='GM EXP Stone', TypeEffect=4, TypeDetail=82,
		    KeepTime=24, FixRange=1, AbilityType1=35, AbilityValue1=50
		WHERE SkillID=250 AND SkillLevel=2 AND SkillName='Mini Sun EXP StoneED'
		  AND TypeEffect=0 AND TypeDetail=0 AND KeepTime=0 AND FixRange=0
		  AND AbilityType1=0 AND AbilityValue1=0 AND AbilityType2=0 AND AbilityType3=0
		  AND EXISTS (SELECT 1 FROM PS_GameDefs_Items
		              WHERE Type=44 AND TypeID=242 AND Range=250 AND AttackTime=2 AND Special=0)`},
	{39, "rebalance_lapis_hammers", forgeHammersMigration},
	{40, "operator_bans", `CREATE TABLE IF NOT EXISTS GX_OperatorAccountBans (UserUID INTEGER PRIMARY KEY, Reason TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS GX_OperatorIPBans (Address TEXT PRIMARY KEY NOT NULL, Reason TEXT NOT NULL, CreatedAt INTEGER NOT NULL DEFAULT (unixepoch()))`},
}

// Base user_version=1 remains compatible with the distributed development seed.
// GX migrations have their own transactional, checksummed ledger.
func migrate(ctx context.Context, pool *sql.DB, steps []migration) error {
	tx, err := pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS GX_SchemaMigrations(Version INTEGER PRIMARY KEY,Name TEXT NOT NULL,Checksum TEXT NOT NULL,AppliedAt TEXT NOT NULL)`); err != nil {
		return err
	}
	var highest int
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(Version),0) FROM GX_SchemaMigrations`).Scan(&highest); err != nil {
		return err
	}
	if highest > len(steps) {
		return fmt.Errorf("runtime migration %d is newer than this server", highest)
	}
	for i, step := range steps {
		if step.Version != i+1 {
			return fmt.Errorf("non-contiguous migration version")
		}
		// Embedded SQL may have been checked out with CRLF on Windows.
		// Only line endings are canonicalized; every other byte is checked.
		sqlText := strings.ReplaceAll(step.SQL, "\r\n", "\n")
		checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(sqlText)))
		var oldName, oldChecksum string
		err = tx.QueryRowContext(ctx, `SELECT Name,Checksum FROM GX_SchemaMigrations WHERE Version=?`, step.Version).Scan(&oldName, &oldChecksum)
		if err == nil {
			if oldName != step.Name {
				return fmt.Errorf("migration %d checksum mismatch", step.Version)
			}
			if oldChecksum != checksum {
				legacy := strings.ReplaceAll(sqlText, "\n", "\r\n")
				if oldChecksum != fmt.Sprintf("%x", sha256.Sum256([]byte(legacy))) {
					return fmt.Errorf("migration %d checksum mismatch", step.Version)
				}
			}
			continue
		}
		if err != sql.ErrNoRows {
			return err
		}
		if step.Version <= highest {
			return fmt.Errorf("migration history contains gaps")
		}
		if _, err = tx.ExecContext(ctx, sqlText); err != nil {
			return fmt.Errorf("migration %d: %w", step.Version, err)
		}
		if step.Version == 11 {
			if err := initializeEnchant11(ctx, tx); err != nil {
				return fmt.Errorf("migration %d: %w", step.Version, err)
			}
		}
		if step.Version == 17 {
			if err := initializeNPCQuest(ctx, tx); err != nil {
				return fmt.Errorf("migration %d: %w", step.Version, err)
			}
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO GX_SchemaMigrations VALUES(?,?,?,?)`, step.Version, step.Name, checksum, time.Now().UTC().Format(time.RFC3339)); err != nil {
			return err
		}
	}
	return tx.Commit()
}
