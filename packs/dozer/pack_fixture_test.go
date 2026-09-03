package dozerpack_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"permitdenied/internal/contract"
)

type packLocator struct {
	ID            contract.ID `json:"id"`
	Map           string      `json:"map"`
	Mission       string      `json:"mission"`
	Vehicles      []string    `json:"vehicles"`
	Destructibles []string    `json:"destructibles"`
	AssetRoots    struct {
		Sprites []string `json:"sprites"`
	} `json:"asset_roots"`
}

func TestDozerPackFixtureValidatesBundle(t *testing.T) {
	packDir := "."
	locatorBytes := mustReadFile(t, filepath.Join(packDir, "pack.json"))
	var locator packLocator
	decodeStrictJSON(t, locatorBytes, &locator)
	if err := locator.ID.Validate("pack.id"); err != nil {
		t.Fatal(err)
	}
	if len(locator.Vehicles) == 0 {
		t.Fatal("pack.vehicles must not be empty")
	}
	if len(locator.Destructibles) == 0 {
		t.Fatal("pack.destructibles must not be empty")
	}
	if len(locator.AssetRoots.Sprites) == 0 {
		t.Fatal("pack.asset_roots.sprites must not be empty")
	}

	mapPath := resolvePackPath(t, packDir, locator.Map)
	missionPath := resolvePackPath(t, packDir, locator.Mission)

	mapDoc, err := contract.ParseMapJSON(mustReadFile(t, mapPath))
	if err != nil {
		t.Fatalf("parse map: %v", err)
	}
	missionDoc, err := contract.ParseMissionJSON(mustReadFile(t, missionPath))
	if err != nil {
		t.Fatalf("parse mission: %v", err)
	}

	vehicles := make([]contract.Vehicle, 0, len(locator.Vehicles))
	for _, rel := range locator.Vehicles {
		abs := resolvePackPath(t, packDir, rel)
		doc, err := contract.ParseVehicleJSON(mustReadFile(t, abs))
		if err != nil {
			t.Fatalf("parse vehicle %s: %v", rel, err)
		}
		vehicles = append(vehicles, doc)
	}

	destructibles := make([]contract.Destructible, 0, len(locator.Destructibles))
	for _, rel := range locator.Destructibles {
		abs := resolvePackPath(t, packDir, rel)
		doc, err := contract.ParseDestructibleJSON(mustReadFile(t, abs))
		if err != nil {
			t.Fatalf("parse destructible %s: %v", rel, err)
		}
		destructibles = append(destructibles, doc)
	}

	if err := (contract.Bundle{
		Map:           mapDoc,
		Vehicles:      vehicles,
		Destructibles: destructibles,
		Mission:       missionDoc,
	}).Validate(); err != nil {
		t.Fatalf("bundle validation failed: %v", err)
	}

	assertFileExists(t, filepath.Join(filepath.Dir(mapPath), mapDoc.Tileset))

	seenSprites := map[contract.ID]struct{}{}
	for _, vehicle := range vehicles {
		seenSprites[vehicle.SpriteSet] = struct{}{}
	}
	for _, destructible := range destructibles {
		for _, state := range destructible.States {
			seenSprites[state.Sprite] = struct{}{}
		}
	}
	for spriteID := range seenSprites {
		assertSpriteExists(t, packDir, locator.AssetRoots.Sprites, spriteID)
	}
}

func decodeStrictJSON(t *testing.T, src []byte, dst any) {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(src))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		t.Fatalf("decode strict json: %v", err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("decode strict json: expected single document, got %v", err)
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}

func resolvePackPath(t *testing.T, packDir, rel string) string {
	t.Helper()
	if !filepath.IsLocal(rel) {
		t.Fatalf("pack path %q must be relative and stay within the pack directory", rel)
	}
	abs := filepath.Join(packDir, rel)
	assertFileExists(t, abs)
	return abs
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func assertSpriteExists(t *testing.T, packDir string, roots []string, spriteID contract.ID) {
	t.Helper()
	_, local, ok := strings.Cut(spriteID.String(), ":")
	if !ok || local == "" {
		t.Fatalf("sprite id %q is not namespaced", spriteID)
	}
	for _, root := range roots {
		if !filepath.IsLocal(root) {
			t.Fatalf("asset root %q must be local to the pack", root)
		}
		path := filepath.Join(packDir, root, local+".png")
		if _, err := os.Stat(path); err == nil {
			return
		}
	}
	t.Fatalf("no sprite file found for %q under asset roots %v", spriteID, roots)
}
