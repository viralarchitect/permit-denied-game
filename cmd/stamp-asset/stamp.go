package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	tilesetCols   = 8
	tileSize      = 16
	lockedTileMax = 18
)

type Paths struct {
	SpritesPNG  string
	SpritesJSON string
	TilesetPNG  string
	Ledger      string
	SrcSprites  string
	SrcTiles    string
}

func DefaultPaths() Paths {
	return Paths{
		SpritesPNG:  "assets/usable/sprites.png",
		SpritesJSON: "assets/usable/sprites.json",
		TilesetPNG:  "assets/usable/tileset.png",
		Ledger:      "assets/LEDGER.md",
		SrcSprites:  "assets/src/sprites",
		SrcTiles:    "assets/src/tiles",
	}
}

type Rect struct {
	X, Y, W, H int
}

type frameDoc struct {
	Frames map[string]struct {
		X      int       `json:"x"`
		Y      int       `json:"y"`
		W      int       `json:"w"`
		H      int       `json:"h"`
		Anchor []float64 `json:"anchor"`
	} `json:"frames"`
}

type ledgerSlot struct {
	Name   string
	Sheet  string
	Rect   Rect
	Status string
}

func Overlaps(a, b Rect) bool {
	return !(a.X+a.W <= b.X || b.X+b.W <= a.X || a.Y+a.H <= b.Y || b.Y+b.H <= a.Y)
}

func loadFrames(path string) (frameDoc, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return frameDoc{}, nil, err
	}
	var doc frameDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return frameDoc{}, nil, err
	}
	if doc.Frames == nil {
		doc.Frames = map[string]struct {
			X      int       `json:"x"`
			Y      int       `json:"y"`
			W      int       `json:"w"`
			H      int       `json:"h"`
			Anchor []float64 `json:"anchor"`
		}{}
	}
	return doc, raw, nil
}

func decodePNGFile(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func toNRGBA(src image.Image) *image.NRGBA {
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)
	return dst
}

func growPad(src *image.NRGBA, minW, minH int) *image.NRGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w >= minW && h >= minH {
		return src
	}
	if minW > w {
		w = minW
	}
	if minH > h {
		h = minH
	}
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, b, src, b.Min, draw.Src)
	return dst
}

// blitIntoSheet decodes sheetPath, blits master into r (growing with transparent pad if needed), writeReplace.
func blitIntoSheet(sheetPath string, master image.Image, r Rect) error {
	sheet, err := decodePNGFile(sheetPath)
	if err != nil {
		return err
	}
	dst := toNRGBA(sheet)
	mb := master.Bounds()
	if mb.Dx() != r.W || mb.Dy() != r.H {
		return fmt.Errorf("master size %dx%d != slot %dx%d", mb.Dx(), mb.Dy(), r.W, r.H)
	}
	needW, needH := r.X+r.W, r.Y+r.H
	dst = growPad(dst, needW, needH)
	draw.Draw(dst, image.Rect(r.X, r.Y, r.X+r.W, r.Y+r.H), master, mb.Min, draw.Src)
	out, err := encodePNG(dst)
	if err != nil {
		return err
	}
	return writeReplace(sheetPath, out)
}

func writeReplace(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+"-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	cleanup := func() { _ = os.Remove(name) }
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(name, path); err == nil {
		return nil
	}
	if _, statErr := os.Stat(path); statErr != nil {
		cleanup()
		return err
	}
	if err := os.Remove(path); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(name, path); err != nil {
		cleanup()
		return err
	}
	return nil
}

func StampFrame(paths Paths, name string) error {
	doc, _, err := loadFrames(paths.SpritesJSON)
	if err != nil {
		return err
	}
	fr, ok := doc.Frames[name]
	if !ok {
		return fmt.Errorf("frame %q not in sprites.json", name)
	}
	masterPath := filepath.Join(paths.SrcSprites, name+".png")
	master, err := decodePNGFile(masterPath)
	if err != nil {
		return fmt.Errorf("master %s: %w", masterPath, err)
	}
	r := Rect{X: fr.X, Y: fr.Y, W: fr.W, H: fr.H}
	return blitIntoSheet(paths.SpritesPNG, master, r)
}

func tileRect(id int) Rect {
	return Rect{
		X: (id % tilesetCols) * tileSize,
		Y: (id / tilesetCols) * tileSize,
		W: tileSize,
		H: tileSize,
	}
}

func findTileMaster(srcTiles string, id int) (string, error) {
	prefix := strconv.Itoa(id) + "_"
	entries, err := os.ReadDir(srcTiles)
	if err != nil {
		return "", err
	}
	var match string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if !strings.HasSuffix(strings.ToLower(n), ".png") {
			continue
		}
		if strings.HasPrefix(n, prefix) {
			match = filepath.Join(srcTiles, n)
			break
		}
	}
	if match == "" {
		cand := filepath.Join(srcTiles, strconv.Itoa(id)+".png")
		if _, err := os.Stat(cand); err == nil {
			match = cand
		}
	}
	if match == "" {
		return "", fmt.Errorf("no master for tile %d under %s", id, srcTiles)
	}
	return match, nil
}

func parseLedger(path string) ([]ledgerSlot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []ledgerSlot
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") || strings.Contains(line, "---") {
			continue
		}
		cols := strings.Split(line, "|")
		// leading empty, name, sheet, x, y, w, h, status, trailing empty
		if len(cols) < 8 {
			continue
		}
		name := strings.TrimSpace(cols[1])
		sheet := strings.TrimSpace(cols[2])
		if name == "name" || sheet == "sheet" {
			continue
		}
		x, err1 := strconv.Atoi(strings.TrimSpace(cols[3]))
		y, err2 := strconv.Atoi(strings.TrimSpace(cols[4]))
		w, err3 := strconv.Atoi(strings.TrimSpace(cols[5]))
		h, err4 := strconv.Atoi(strings.TrimSpace(cols[6]))
		status := strings.TrimSpace(cols[7])
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			continue
		}
		out = append(out, ledgerSlot{
			Name:   name,
			Sheet:  sheet,
			Rect:   Rect{X: x, Y: y, W: w, H: h},
			Status: status,
		})
	}
	return out, nil
}

func ledgerTileSlot(slots []ledgerSlot, id int) (ledgerSlot, bool) {
	prefix := strconv.Itoa(id) + "_"
	idStr := strconv.Itoa(id)
	for _, s := range slots {
		if s.Sheet != "tileset.png" {
			continue
		}
		if s.Name == idStr || strings.HasPrefix(s.Name, prefix) {
			return s, true
		}
	}
	return ledgerSlot{}, false
}

func StampTile(paths Paths, id int) error {
	if id < 0 {
		return fmt.Errorf("tile id %d < 0", id)
	}
	r := tileRect(id)
	if id > lockedTileMax {
		slots, err := parseLedger(paths.Ledger)
		if err != nil {
			return fmt.Errorf("ledger: %w", err)
		}
		slot, ok := ledgerTileSlot(slots, id)
		if !ok {
			return fmt.Errorf("tile id %d not reserved in LEDGER", id)
		}
		if slot.Status != "reserved" && slot.Status != "locked" {
			return fmt.Errorf("tile id %d LEDGER status %q", id, slot.Status)
		}
		r = slot.Rect
	}
	masterPath, err := findTileMaster(paths.SrcTiles, id)
	if err != nil {
		return err
	}
	master, err := decodePNGFile(masterPath)
	if err != nil {
		return fmt.Errorf("master %s: %w", masterPath, err)
	}
	return blitIntoSheet(paths.TilesetPNG, master, r)
}

func ParseReserve(spec string) (name string, r Rect, err error) {
	parts := strings.SplitN(spec, "=", 2)
	if len(parts) != 2 {
		return "", Rect{}, fmt.Errorf("reserve want name=x,y,w,h got %q", spec)
	}
	name = strings.TrimSpace(parts[0])
	if name == "" {
		return "", Rect{}, fmt.Errorf("empty reserve name")
	}
	nums := strings.Split(parts[1], ",")
	if len(nums) != 4 {
		return "", Rect{}, fmt.Errorf("reserve want name=x,y,w,h got %q", spec)
	}
	vals := make([]int, 4)
	for i, s := range nums {
		v, e := strconv.Atoi(strings.TrimSpace(s))
		if e != nil {
			return "", Rect{}, fmt.Errorf("reserve parse: %w", e)
		}
		vals[i] = v
	}
	r = Rect{X: vals[0], Y: vals[1], W: vals[2], H: vals[3]}
	if r.W <= 0 || r.H <= 0 {
		return "", Rect{}, fmt.Errorf("reserve size must be positive")
	}
	return name, r, nil
}

func ReserveFrame(paths Paths, name string, r Rect) error {
	doc, raw, err := loadFrames(paths.SpritesJSON)
	if err != nil {
		return err
	}
	if _, exists := doc.Frames[name]; exists {
		return fmt.Errorf("frame %q already exists", name)
	}
	for other, fr := range doc.Frames {
		or := Rect{X: fr.X, Y: fr.Y, W: fr.W, H: fr.H}
		if Overlaps(r, or) {
			return fmt.Errorf("rect overlaps existing frame %q", other)
		}
	}
	out, err := appendFrameJSON(raw, name, r)
	if err != nil {
		return err
	}
	return writeReplace(paths.SpritesJSON, out)
}

// appendFrameJSON surgically inserts one new frames entry without re-marshaling the file.
func appendFrameJSON(raw []byte, name string, r Rect) ([]byte, error) {
	const key = `"frames"`
	idx := bytes.Index(raw, []byte(key))
	if idx < 0 {
		return nil, fmt.Errorf("sprites.json missing frames key")
	}
	i := idx + len(key)
	for i < len(raw) && (raw[i] == ' ' || raw[i] == '\t' || raw[i] == '\n' || raw[i] == '\r' || raw[i] == ':') {
		i++
	}
	if i >= len(raw) || raw[i] != '{' {
		return nil, fmt.Errorf("sprites.json frames is not an object")
	}
	start := i
	depth := 0
	end := -1
	for j := start; j < len(raw); j++ {
		switch raw[j] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = j
				j = len(raw)
			}
		}
	}
	if end < 0 {
		return nil, fmt.Errorf("sprites.json frames object unclosed")
	}
	inner := bytes.TrimSpace(raw[start+1 : end])
	entry := fmt.Sprintf(`"%s": { "x": %d, "y": %d, "w": %d, "h": %d, "anchor": [%d, %d] }`,
		name, r.X, r.Y, r.W, r.H, r.W/2, r.H/2)
	var mid []byte
	if len(inner) == 0 {
		mid = []byte("\n    " + entry + "\n  ")
	} else {
		if inner[len(inner)-1] == ',' {
			mid = append(append([]byte{}, raw[start+1:end]...), []byte("\n    "+entry+"\n  ")...)
		} else {
			// keep original inner bytes, add comma after last entry
			mid = append(append([]byte{}, raw[start+1:end]...), []byte(",\n    "+entry+"\n  ")...)
		}
	}
	out := append([]byte{}, raw[:start+1]...)
	out = append(out, mid...)
	out = append(out, raw[end:]...)
	return out, nil
}
