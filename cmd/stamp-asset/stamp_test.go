package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("go.mod not found from ", wd)
	return ""
}

func usable(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "assets", "usable", name)
}

func writeTestPNG(path string, w, h int, c color.NRGBA) error {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

func pixelEq(a, b image.Image, r Rect) bool {
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			ar, ag, ab, aa := a.At(x, y).RGBA()
			br, bg, bb, ba := b.At(x, y).RGBA()
			if ar != br || ag != bg || ab != bb || aa != ba {
				return false
			}
		}
	}
	return true
}

func TestOverlapsEdgeAdjacentOK(t *testing.T) {
	a := Rect{0, 0, 32, 32}
	b := Rect{32, 0, 32, 32}
	if Overlaps(a, b) {
		t.Fatal("edge-adjacent should not overlap")
	}
	c := Rect{31, 0, 32, 32}
	if !Overlaps(a, c) {
		t.Fatal("overlapping interiors should collide")
	}
}

func TestReserveOverlapRejected(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "sprites.json")
	if err := copyFile(usable(t, "sprites.json"), jsonPath); err != nil {
		t.Fatal(err)
	}
	paths := Paths{SpritesJSON: jsonPath}
	if err := ReserveFrame(paths, "overlap_test", Rect{0, 0, 16, 16}); err == nil {
		t.Fatal("expected overlap reject")
	}
}

func TestReserveAppendNoOverlap(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "sprites.json")
	before, err := os.ReadFile(usable(t, "sprites.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(jsonPath, before, 0o644); err != nil {
		t.Fatal(err)
	}
	paths := Paths{SpritesJSON: jsonPath}
	r := Rect{400, 256, 16, 16}
	if err := ReserveFrame(paths, "slot_test_00", r); err != nil {
		t.Fatal(err)
	}
	doc, after, err := loadFrames(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	fr, ok := doc.Frames["slot_test_00"]
	if !ok {
		t.Fatal("missing reserved frame")
	}
	if fr.X != r.X || fr.Y != r.Y || fr.W != r.W || fr.H != r.H {
		t.Fatalf("reserved rect %+v", fr)
	}
	up, ok := doc.Frames["dozer_up_00"]
	if !ok || up.X != 0 || up.Y != 0 || up.W != 32 || up.H != 32 {
		t.Fatalf("dozer_up_00 moved: %+v", up)
	}
	idx := bytes.Index(before, []byte(`"frames"`))
	if idx < 0 || !bytes.Equal(before[:idx], after[:idx]) {
		t.Fatal("reserve rewrote file prefix")
	}
}

func TestStampFrameIsolatesPixels(t *testing.T) {
	dir := t.TempDir()
	pngPath := filepath.Join(dir, "sprites.png")
	jsonPath := filepath.Join(dir, "sprites.json")
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(usable(t, "sprites.png"), pngPath); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(usable(t, "sprites.json"), jsonPath); err != nil {
		t.Fatal(err)
	}
	paths := Paths{SpritesPNG: pngPath, SpritesJSON: jsonPath, SrcSprites: srcDir}
	if err := ReserveFrame(paths, "slot_test_00", Rect{400, 256, 16, 16}); err != nil {
		t.Fatal(err)
	}
	before, err := decodePNGFile(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	master := color.NRGBA{R: 255, G: 0, B: 128, A: 255}
	if err := writeTestPNG(filepath.Join(srcDir, "slot_test_00.png"), 16, 16, master); err != nil {
		t.Fatal(err)
	}
	if err := StampFrame(paths, "slot_test_00"); err != nil {
		t.Fatal(err)
	}
	after, err := decodePNGFile(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	doc, _, err := loadFrames(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	for name, fr := range doc.Frames {
		if name == "slot_test_00" {
			continue
		}
		r := Rect{fr.X, fr.Y, fr.W, fr.H}
		if !pixelEq(before, after, r) {
			t.Fatalf("locked frame %s pixels changed", name)
		}
	}
	r := Rect{400, 256, 16, 16}
	ar, ag, ab, aa := after.At(r.X, r.Y).RGBA()
	if uint8(ar>>8) != master.R || uint8(ag>>8) != master.G || uint8(ab>>8) != master.B || uint8(aa>>8) != master.A {
		t.Fatalf("stamped pixel not master color: %d %d %d %d", ar, ag, ab, aa)
	}
}

func TestStampTileIsolatesLockedCells(t *testing.T) {
	dir := t.TempDir()
	pngPath := filepath.Join(dir, "tileset.png")
	srcDir := filepath.Join(dir, "tiles")
	ledgerPath := filepath.Join(dir, "LEDGER.md")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(usable(t, "tileset.png"), pngPath); err != nil {
		t.Fatal(err)
	}
	ledger := "| name | sheet | x | y | w | h | status |\n|---|---|---:|---:|---:|---:|---|\n| 19_test | tileset.png | 48 | 32 | 16 | 16 | reserved |\n"
	if err := os.WriteFile(ledgerPath, []byte(ledger), 0o644); err != nil {
		t.Fatal(err)
	}
	before, err := decodePNGFile(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	master := color.NRGBA{R: 10, G: 200, B: 30, A: 255}
	if err := writeTestPNG(filepath.Join(srcDir, "19_test.png"), 16, 16, master); err != nil {
		t.Fatal(err)
	}
	paths := Paths{TilesetPNG: pngPath, SrcTiles: srcDir, Ledger: ledgerPath}
	if err := StampTile(paths, 19); err != nil {
		t.Fatal(err)
	}
	after, err := decodePNGFile(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	for id := 0; id <= lockedTileMax; id++ {
		r := tileRect(id)
		if !pixelEq(before, after, r) {
			t.Fatalf("locked tile %d pixels changed", id)
		}
	}
}

func TestStampFrameMissingName(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "sprites.json")
	pngPath := filepath.Join(dir, "sprites.png")
	srcDir := filepath.Join(dir, "src")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = copyFile(usable(t, "sprites.json"), jsonPath)
	_ = copyFile(usable(t, "sprites.png"), pngPath)
	paths := Paths{SpritesPNG: pngPath, SpritesJSON: jsonPath, SrcSprites: srcDir}
	if err := StampFrame(paths, "no_such_frame"); err == nil {
		t.Fatal("expected missing name error")
	}
}

func TestStampFrameSizeMismatch(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "sprites.json")
	pngPath := filepath.Join(dir, "sprites.png")
	srcDir := filepath.Join(dir, "src")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = copyFile(usable(t, "sprites.json"), jsonPath)
	_ = copyFile(usable(t, "sprites.png"), pngPath)
	_ = writeTestPNG(filepath.Join(srcDir, "dozer_up_00.png"), 8, 8, color.NRGBA{A: 255})
	paths := Paths{SpritesPNG: pngPath, SpritesJSON: jsonPath, SrcSprites: srcDir}
	if err := StampFrame(paths, "dozer_up_00"); err == nil {
		t.Fatal("expected size mismatch")
	}
}
