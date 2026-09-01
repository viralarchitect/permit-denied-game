package meta

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "meta.json")
	in := Save{EngineTier: 2, ArmorTier: 1, BestCash: 4400, HighestTier: 3}
	if !in.Grant(Engine1) {
		t.Fatal("grant engine_1")
	}
	if !in.Grant(Ripper) {
		t.Fatal("grant ripper")
	}
	if err := SaveFile(path, in); err != nil {
		t.Fatal(err)
	}
	out, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if out.EngineTier != 2 || out.ArmorTier != 1 || out.BestCash != 4400 || out.HighestTier != 3 {
		t.Fatalf("scalars: %+v", out)
	}
	if !out.Has(Engine1) || !out.Has(Ripper) {
		t.Fatalf("unlocks=%v", out.Unlocks)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestLoadMissingIsEmpty(t *testing.T) {
	s, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if s.BestCash != 0 || len(s.Unlocks) != 0 {
		t.Fatalf("got %+v", s)
	}
}
