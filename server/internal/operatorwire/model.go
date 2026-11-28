// Package operatorwire defines the private local operator protocol.
package operatorwire

import (
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
)

type Account struct {
	ID        uint32
	Name      string
	GM        bool
	Points    uint32
	Banned    bool
	BanReason string
}

type IPBan struct {
	Address, Reason string
	CreatedAt       time.Time
}

type Character struct {
	ID      uint32
	Name    string
	Level   uint16
	Deleted bool
	Online  bool
}

type Stats struct {
	Strength, Dexterity, Rec, Intelligence, Wisdom, Luck uint16
	StatPoints, SkillPoints                              uint16
	Gold, EXP, Kills                                     uint32
}

type Request struct {
	Economy                      *config.Economy         `json:"economy,omitempty"`
	ScopedDrop                   *config.ScopedDropEdit  `json:"scoped_drop,omitempty"`
	MapTuning                    *config.MapTuning       `json:"map_tuning,omitempty"`
	RewardEvent                  *config.RewardEventEdit `json:"reward_event,omitempty"`
	Action                       string
	MapID                        uint16
	MapReward                    *config.MapRewardOverrides `json:"map_reward,omitempty"`
	Query                        string
	AccountID, CharacterID       uint32
	Name, Password               string
	IPAddress, BanReason         string
	GM                           bool
	Points                       uint32
	ExpectedAccount              *Account
	Stats, ExpectedStats         *Stats
	ConfirmName                  string
	MobRate, QuestRate, KillRate string
	DropRate                     *string `json:"drop_rate,omitempty"`
	GoldRate                     *string `json:"gold_rate,omitempty"`
	WorldDropRate                *string `json:"world_drop_rate,omitempty"`
}

type BossRespawn struct {
	MonsterID uint16
	Alive     bool
	Next      time.Time
}

type MapActivity struct {
	BossRespawns                                              []BossRespawn
	Players, Instances                                        int
	MonsterKills, PvPKills, Drops, ItemsCollected, GoldGained uint64
}

type Response struct {
	MapActivity    *MapActivity        `json:"map_activity,omitempty"`
	EffectiveRates *config.RewardRates `json:"effective_rates,omitempty"`
	Accounts       []Account
	IPBans         []IPBan
	More           bool
	Account        *Account
	Characters     []Character
	Stats          *Stats
	Online         bool
	MinEXP, MaxEXP uint32
}
