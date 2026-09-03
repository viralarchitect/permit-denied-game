package dozerpack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
	"sync"

	"permitdenied/internal/contract"
	"permitdenied/internal/lot"
	"permitdenied/packs"
)

const Root = "packs/dozer"
const embeddedRoot = "dozer"

type BlockerKind string

const (
	BlockerJersey   BlockerKind = "jersey"
	BlockerDump     BlockerKind = "dump"
	BlockerConcrete BlockerKind = "concrete"
)

type BlockerSpec struct {
	Kind       BlockerKind
	X, Y, W, H float64
}

type SpawnPoint struct {
	X, Y float64
}

type locator struct {
	ID            contract.ID `json:"id"`
	Map           string      `json:"map"`
	Mission       string      `json:"mission"`
	Vehicles      []string    `json:"vehicles"`
	Destructibles []string    `json:"destructibles"`
	AssetRoots    struct {
		Sprites []string `json:"sprites"`
	} `json:"asset_roots"`
}

type Scenario struct {
	gameMap       contract.Map
	mission       contract.Mission
	lot           lot.Lot
	spawnX        float64
	spawnY        float64
	spawnHeading  float64
	startBlockers []BlockerSpec
	cruiserSpawns []SpawnPoint
}

var (
	embeddedOnce sync.Once
	embeddedPack Scenario
	embeddedErr  error
)

func LoadEmbedded() (Scenario, error) {
	embeddedOnce.Do(func() {
		embeddedPack, embeddedErr = LoadFS(packs.FS, embeddedRoot)
	})
	return embeddedPack, embeddedErr
}

func LoadFS(fsys fs.FS, root string) (Scenario, error) {
	locatorPath := path.Join(root, "pack.json")
	locatorBytes, err := fs.ReadFile(fsys, locatorPath)
	if err != nil {
		return Scenario{}, fmt.Errorf("read %s: %w", locatorPath, err)
	}
	var loc locator
	if err := decodeStrict(locatorBytes, &loc); err != nil {
		return Scenario{}, fmt.Errorf("parse %s: %w", locatorPath, err)
	}
	if err := loc.ID.Validate("pack.id"); err != nil {
		return Scenario{}, err
	}

	mapPath, err := resolve(root, loc.Map)
	if err != nil {
		return Scenario{}, fmt.Errorf("pack map: %w", err)
	}
	missionPath, err := resolve(root, loc.Mission)
	if err != nil {
		return Scenario{}, fmt.Errorf("pack mission: %w", err)
	}

	gameMap, err := parseMap(fsys, mapPath)
	if err != nil {
		return Scenario{}, err
	}
	mission, err := parseMission(fsys, missionPath)
	if err != nil {
		return Scenario{}, err
	}

	vehicles := make([]contract.Vehicle, 0, len(loc.Vehicles))
	for _, rel := range loc.Vehicles {
		vehiclePath, err := resolve(root, rel)
		if err != nil {
			return Scenario{}, fmt.Errorf("pack vehicle: %w", err)
		}
		doc, err := parseVehicle(fsys, vehiclePath)
		if err != nil {
			return Scenario{}, err
		}
		vehicles = append(vehicles, doc)
	}

	destructibles := make([]contract.Destructible, 0, len(loc.Destructibles))
	destructibleSet := make(map[contract.ID]contract.Destructible, len(loc.Destructibles))
	for _, rel := range loc.Destructibles {
		destructiblePath, err := resolve(root, rel)
		if err != nil {
			return Scenario{}, fmt.Errorf("pack destructible: %w", err)
		}
		doc, err := parseDestructible(fsys, destructiblePath)
		if err != nil {
			return Scenario{}, err
		}
		destructibles = append(destructibles, doc)
		destructibleSet[doc.ID] = doc
	}

	bundle := contract.Bundle{
		Map:           gameMap,
		Vehicles:      vehicles,
		Destructibles: destructibles,
		Mission:       mission,
	}
	if err := bundle.Validate(); err != nil {
		return Scenario{}, fmt.Errorf("validate dozer pack bundle: %w", err)
	}

	spawnObj, ok := gameMap.ObjectByID(mission.PlayerSpawnID)
	if !ok {
		return Scenario{}, fmt.Errorf("player spawn %q not found", mission.PlayerSpawnID)
	}

	scenario := Scenario{
		gameMap:       gameMap,
		mission:       mission,
		spawnX:        spawnObj.X,
		spawnY:        spawnObj.Y,
		spawnHeading:  spawnObj.Heading,
		startBlockers: blockersFromMap(gameMap),
		cruiserSpawns: cruiserSpawnsFromMap(gameMap, mission),
	}
	scenario.lot, err = lotFromBundle(bundle, destructibleSet)
	if err != nil {
		return Scenario{}, err
	}
	return scenario, nil
}

func (s Scenario) Map() contract.Map {
	return s.gameMap
}

func (s Scenario) Mission() contract.Mission {
	return s.mission
}

func (s Scenario) Spawn() (x, y, heading float64) {
	return s.spawnX, s.spawnY, s.spawnHeading
}

func (s Scenario) Lot() lot.Lot {
	buildings := append([]lot.Building(nil), s.lot.Buildings...)
	rubble := append([]lot.Rubble(nil), s.lot.Rubble...)
	return lot.Lot{
		Buildings: buildings,
		Rubble:    rubble,
	}
}

func (s Scenario) StartBlockers() []BlockerSpec {
	return append([]BlockerSpec(nil), s.startBlockers...)
}

func (s Scenario) CruiserSpawns() []SpawnPoint {
	return append([]SpawnPoint(nil), s.cruiserSpawns...)
}

func lotFromBundle(bundle contract.Bundle, destructibleSet map[contract.ID]contract.Destructible) (lot.Lot, error) {
	buildings := make([]lot.Building, 0, len(bundle.Map.Objects))
	for _, obj := range bundle.Map.Objects {
		if obj.Type != contract.MapObjectBuilding {
			continue
		}
		destructibleID, _, err := obj.PropertyID("destructible_id")
		if err != nil {
			return lot.Lot{}, err
		}
		desc, ok := destructibleSet[destructibleID]
		if !ok {
			return lot.Lot{}, fmt.Errorf("map building %q references unknown destructible %q", obj.ID, destructibleID)
		}
		target, label, err := targetAndLabel(obj)
		if err != nil {
			return lot.Lot{}, err
		}
		rubbleState, err := desc.StateForHealthFraction(0)
		if err != nil {
			return lot.Lot{}, fmt.Errorf("destructible %q rubble state: %w", desc.ID, err)
		}
		buildings = append(buildings, lot.Building{
			ID:             target,
			Label:          label,
			X:              obj.X,
			Y:              obj.Y,
			W:              obj.W,
			H:              obj.H,
			HP:             desc.Health,
			MaxHP:          desc.Health,
			State:          lot.Intact,
			Value:          desc.YieldDollars,
			Material:       materialFromContract(desc.Material),
			Role:           roleForTarget(target),
			AuthoredRubble: true,
			SpawnsRubble:   desc.RubbleRule.SpawnCollision,
			RubbleInset:    desc.RubbleRule.Inset,
			RubbleRamp:     desc.RubbleRule.Ramp || rubbleState.Collision == contract.CollisionRamp,
		})
	}
	return lot.Lot{Buildings: buildings}, nil
}

func blockersFromMap(gameMap contract.Map) []BlockerSpec {
	out := make([]BlockerSpec, 0)
	for _, obj := range gameMap.Objects {
		if obj.Type != contract.MapObjectBlocker {
			continue
		}
		when, ok, err := obj.PropertyString("when")
		if err != nil || (ok && when != "start") {
			continue
		}
		kind, ok, err := obj.PropertyString("kind")
		if err != nil || !ok {
			continue
		}
		switch BlockerKind(kind) {
		case BlockerJersey, BlockerDump, BlockerConcrete:
			out = append(out, BlockerSpec{
				Kind: BlockerKind(kind),
				X:    obj.X,
				Y:    obj.Y,
				W:    obj.W,
				H:    obj.H,
			})
		}
	}
	return out
}

func cruiserSpawnsFromMap(gameMap contract.Map, mission contract.Mission) []SpawnPoint {
	markerIDs := make(map[contract.ID]struct{})
	for _, entries := range mission.SpawnTables {
		for _, entry := range entries {
			markerIDs[entry.SpawnMarkerID] = struct{}{}
		}
	}
	if len(markerIDs) == 0 {
		return nil
	}
	seen := make(map[contract.ID]struct{}, len(markerIDs))
	out := make([]SpawnPoint, 0, len(markerIDs))
	for _, obj := range gameMap.Objects {
		if obj.Type != contract.MapObjectMarker {
			continue
		}
		if _, ok := markerIDs[obj.ID]; !ok {
			continue
		}
		if _, ok := seen[obj.ID]; ok {
			continue
		}
		seen[obj.ID] = struct{}{}
		out = append(out, SpawnPoint{
			X: obj.X + obj.W/2,
			Y: obj.Y + obj.H/2,
		})
	}
	return out
}

func targetAndLabel(obj contract.MapObject) (lot.TargetID, string, error) {
	target, ok, err := obj.PropertyString("target")
	if err != nil {
		return lot.TargetNone, "", err
	}
	if !ok || target == "" || target == "none" {
		return lot.TargetNone, "", nil
	}
	switch target {
	case "sheriff":
		return lot.TargetSheriff, "SHERIFF", nil
	case "yard":
		return lot.TargetYard, "YARD", nil
	case "plant":
		return lot.TargetPlant, "PLANT", nil
	default:
		return lot.TargetNone, "", fmt.Errorf("map object %q has unknown target %q", obj.ID, target)
	}
}

func roleForTarget(target lot.TargetID) lot.Role {
	if target == lot.TargetNone {
		return lot.RoleMundane
	}
	return lot.RoleTarget
}

func materialFromContract(material contract.DestructibleMaterial) lot.Material {
	switch material {
	case contract.MaterialWood:
		return lot.MatWood
	case contract.MaterialConcrete:
		return lot.MatConcrete
	case contract.MaterialSteel:
		return lot.MatSteel
	default:
		return lot.MatBrick
	}
}

func parseMap(fsys fs.FS, file string) (contract.Map, error) {
	b, err := fs.ReadFile(fsys, file)
	if err != nil {
		return contract.Map{}, fmt.Errorf("read %s: %w", file, err)
	}
	doc, err := contract.ParseMapJSON(b)
	if err != nil {
		return contract.Map{}, fmt.Errorf("parse %s: %w", file, err)
	}
	return doc, nil
}

func parseMission(fsys fs.FS, file string) (contract.Mission, error) {
	b, err := fs.ReadFile(fsys, file)
	if err != nil {
		return contract.Mission{}, fmt.Errorf("read %s: %w", file, err)
	}
	doc, err := contract.ParseMissionJSON(b)
	if err != nil {
		return contract.Mission{}, fmt.Errorf("parse %s: %w", file, err)
	}
	return doc, nil
}

func parseVehicle(fsys fs.FS, file string) (contract.Vehicle, error) {
	b, err := fs.ReadFile(fsys, file)
	if err != nil {
		return contract.Vehicle{}, fmt.Errorf("read %s: %w", file, err)
	}
	doc, err := contract.ParseVehicleJSON(b)
	if err != nil {
		return contract.Vehicle{}, fmt.Errorf("parse %s: %w", file, err)
	}
	return doc, nil
}

func parseDestructible(fsys fs.FS, file string) (contract.Destructible, error) {
	b, err := fs.ReadFile(fsys, file)
	if err != nil {
		return contract.Destructible{}, fmt.Errorf("read %s: %w", file, err)
	}
	doc, err := contract.ParseDestructibleJSON(b)
	if err != nil {
		return contract.Destructible{}, fmt.Errorf("parse %s: %w", file, err)
	}
	return doc, nil
}

func resolve(root, rel string) (string, error) {
	clean := path.Clean(rel)
	if clean == "." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
		return "", fmt.Errorf("path %q must stay within %q", rel, root)
	}
	return path.Join(root, clean), nil
}

func decodeStrict(src []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(src))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("expected a single JSON document")
	}
	return nil
}
