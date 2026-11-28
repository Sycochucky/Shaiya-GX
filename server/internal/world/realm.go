package world

import (
	"sync"
	"sync/atomic"
)

// realmState is the process's authoritative realm. Map shards own entity state
// and Lua VMs; sessions own characters. These registries own identities/lifetimes,
// not gameplay mutations. One realm per process keeps deployment lightweight.
// Separate realms run in separate processes with separate SQLite runtimes.
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
	activeAccounts       map[uint32]*Session
	instanceMu           sync.Mutex
	instanceMaps         map[uint16]uint16
	instanceTiming       map[uint16]dungeonRun
	instanceOwner        map[uint16]uint32
	nextInstance         uint16
}

func newRealm() *realmState {
	r := &realmState{shards: map[uint16]*mapShard{}, grids: map[uint16]mapGrid{}, registry: map[uint16]map[uint32]*ConnectedPlayer{}, activeAccounts: map[uint32]*Session{}, instanceMaps: map[uint16]uint16{}, instanceOwner: map[uint16]uint32{}, instanceTiming: map[uint16]dungeonRun{}, nextInstance: instanceZoneBase}
	r.nextNpcID.Store(0x70000000)
	r.nextMobID.Store(0x71000000)
	return r
}

var realm = newRealm()
