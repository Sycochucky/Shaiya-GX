package world

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
)

// Instanced zones (Map.ini CreateType R guild ranking, H guild house, P
// party) are copies of a map with their own mobs, NPCs and players
// (CWorld::GetZoneByMapID / CZone). Rapture keys the world by a uint16 zone
// id: real maps keep their map id, instances get a virtual id from
// instanceZoneBase up that aliases the real map for its data. The client and
// the database only ever see the real map (realMap).

const instanceZoneBase = 1000

// realMap is the map a zone shows (itself for real maps).
func realMap(zone uint16) uint16 {
	if zone < instanceZoneBase {
		return zone
	}
	realm.instanceMu.Lock()
	defer realm.instanceMu.Unlock()
	if m, ok := realm.instanceMaps[zone]; ok {
		return m
	}
	return zone
}

func isInstanceZone(zone uint16) bool { return zone >= instanceZoneBase }

func init() { db.MapForSave = realMap }

// mapCreateType is the Map.ini CreateType of a zone's map ("" if none).
func mapCreateType(zone uint16) string {
	if z, ok := mapConfig.Zones[realMap(zone)]; ok {
		return z.CreateType
	}
	return ""
}

// createInstance builds a fresh copy of mapID from its svmap.
func createInstance(mapID uint16, owner uint32) (uint16, error) {
	load, err := loadSvmapFile(mapID, filepath.Join(mapDataDir(), fmt.Sprintf("%d.svmap", mapID)), db.MustDefs())
	if err != nil {
		return 0, err
	}
	timing := configuredDungeonRun(mapID, time.Now())
	realm.instanceMu.Lock()
	zone := realm.nextInstance
	for {
		if _, used := realm.instanceMaps[zone]; !used {
			break
		}
		zone++
		if zone < instanceZoneBase {
			zone = instanceZoneBase
		}
	}
	realm.nextInstance = zone + 1
	if realm.nextInstance < instanceZoneBase {
		realm.nextInstance = instanceZoneBase
	}
	realm.instanceMaps[zone], realm.instanceOwner[zone] = mapID, owner
	realm.instanceTiming[zone] = timing
	realm.instanceMu.Unlock()
	RegisterMapGrid(zone, load.mapSize, load.mask)
	nav, err := loadMobNavigation(zone)
	if err != nil {
		destroyInstance(zone)
		return 0, err
	}
	if nav != nil {
		valid := load.entities.Mobs[:0]
		for i := range load.entities.Mobs {
			m := load.entities.Mobs[i]
			if m.placeOnNavigation(nav, zone) {
				valid = append(valid, m)
			} else if m.group != nil {
				m.group.alive--
			}
		}
		load.entities.Mobs = valid
	}
	AddEntities(zone, load.entities)
	initializeInstanceBosses(zone, mapID)
	return zone, nil
}

// destroyInstance drops an (empty) instance's entities and grid.
func destroyInstance(zone uint16) {
	if !isInstanceZone(zone) {
		return
	}
	relics.Lock()
	delete(relics.bosses, zone)
	relics.Unlock()
	realm.instanceMu.Lock()
	delete(realm.instanceMaps, zone)
	delete(realm.instanceOwner, zone)
	delete(realm.instanceTiming, zone)
	realm.instanceMu.Unlock()
	realm.shardsMu.Lock()
	delete(realm.shards, zone)
	realm.shardsMu.Unlock()
	realm.gridsMu.Lock()
	delete(realm.grids, zone)
	realm.gridsMu.Unlock()
	clearMapItems(zone)
}

// instanceZones lists the live instances of mapID.
func instanceZones(mapID uint16) []uint16 {
	realm.instanceMu.Lock()
	defer realm.instanceMu.Unlock()
	var out []uint16
	for z, m := range realm.instanceMaps {
		if m == mapID {
			out = append(out, z)
		}
	}
	return out
}

// instanceOwnerOf is the owner id (guild, party) an instance was created for.
func instanceOwnerOf(zone uint16) uint32 {
	realm.instanceMu.Lock()
	defer realm.instanceMu.Unlock()
	return realm.instanceOwner[zone]
}
