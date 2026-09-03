package dozerpack

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"permitdenied/internal/contract"
	"permitdenied/packs"
)

func TestLoadEmbeddedScenario(t *testing.T) {
	scenario, err := LoadEmbedded()
	if err != nil {
		t.Fatal(err)
	}
	if scenario.Map().Width != 20 || scenario.Map().Height != 18 {
		t.Fatalf("embedded map size=%dx%d want 20x18", scenario.Map().Width, scenario.Map().Height)
	}
	if got := len(scenario.Lot().Buildings); got < 5 {
		t.Fatalf("embedded lot buildings=%d want at least 5", got)
	}
}

func TestRemovingBuildingFromMapChangesScenarioLot(t *testing.T) {
	root := t.TempDir()
	copyEmbeddedPack(t, root)

	original, err := LoadEmbedded()
	if err != nil {
		t.Fatal(err)
	}

	mapPath := filepath.Join(root, "dozer", "map.json")
	mapBytes, err := os.ReadFile(mapPath)
	if err != nil {
		t.Fatal(err)
	}
	mapDoc, err := contract.ParseMapJSON(mapBytes)
	if err != nil {
		t.Fatal(err)
	}

	filtered := make([]contract.MapObject, 0, len(mapDoc.Objects)-1)
	removed := false
	for _, obj := range mapDoc.Objects {
		if !removed && obj.ID == contract.ID("permitdenied:building.shack_west") {
			removed = true
			continue
		}
		filtered = append(filtered, obj)
	}
	if !removed {
		t.Fatal("expected to remove shack_west from map.json")
	}
	mapDoc.Objects = filtered

	rewritten, err := json.MarshalIndent(mapDoc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mapPath, rewritten, 0o644); err != nil {
		t.Fatal(err)
	}

	scenario, err := LoadFS(os.DirFS(root), "dozer")
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(scenario.Lot().Buildings), len(original.Lot().Buildings)-1; got != want {
		t.Fatalf("building count=%d want %d after deleting map object", got, want)
	}
}

func copyEmbeddedPack(t *testing.T, root string) {
	t.Helper()
	if err := fs.WalkDir(packs.FS, "dozer", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(root, filepath.FromSlash(name))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(packs.FS, name)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	}); err != nil {
		t.Fatal(err)
	}
}
