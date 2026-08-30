package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"permitdenied/assets"
)

const (
	tilesetCols = 8
	tileCount   = 19
	tileWet     = 14
	tileSet     = 15
)

type frame struct {
	X, Y, W, H int
	AnchorX    float64
	AnchorY    float64
}

type atlas struct {
	tileset *ebiten.Image
	sprites *ebiten.Image
	frames  map[string]frame
	ground  [][]int
	decal   [][]int
	tileImg [tileCount]*ebiten.Image
}

var (
	atlasOnce sync.Once
	atlasInst *atlas
	atlasErr  error
)

func ensureAtlas() (*atlas, error) {
	atlasOnce.Do(func() {
		atlasInst, atlasErr = loadAtlas()
	})
	return atlasInst, atlasErr
}

func loadAtlas() (*atlas, error) {
	a := &atlas{frames: make(map[string]frame)}

	tilesetImg, err := decodePNG("usable/tileset.png")
	if err != nil {
		return nil, err
	}
	a.tileset = ebiten.NewImageFromImage(tilesetImg)

	spritesImg, err := decodePNG("usable/sprites.png")
	if err != nil {
		return nil, err
	}
	a.sprites = ebiten.NewImageFromImage(spritesImg)

	if err := a.loadFrames(); err != nil {
		return nil, err
	}
	if err := a.loadLot(); err != nil {
		return nil, err
	}
	a.cacheTiles()
	return a, nil
}

func decodePNG(path string) (image.Image, error) {
	b, err := assets.FS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return img, nil
}

type spritesJSON struct {
	Frames map[string]struct {
		X      int       `json:"x"`
		Y      int       `json:"y"`
		W      int       `json:"w"`
		H      int       `json:"h"`
		Anchor []float64 `json:"anchor"`
	} `json:"frames"`
}

func (a *atlas) loadFrames() error {
	b, err := assets.FS.ReadFile("usable/sprites.json")
	if err != nil {
		return fmt.Errorf("read sprites.json: %w", err)
	}
	var doc spritesJSON
	if err := json.Unmarshal(b, &doc); err != nil {
		return fmt.Errorf("parse sprites.json: %w", err)
	}
	for name, f := range doc.Frames {
		fr := frame{X: f.X, Y: f.Y, W: f.W, H: f.H}
		if len(f.Anchor) >= 2 {
			fr.AnchorX = f.Anchor[0]
			fr.AnchorY = f.Anchor[1]
		} else {
			fr.AnchorX = float64(f.W) / 2
			fr.AnchorY = float64(f.H) / 2
		}
		a.frames[name] = fr
	}
	return nil
}

type lotJSON struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Layers struct {
		Ground [][]int `json:"ground"`
		Decal  [][]int `json:"decal"`
	} `json:"layers"`
}

func (a *atlas) loadLot() error {
	b, err := assets.FS.ReadFile("usable/lot.json")
	if err != nil {
		return fmt.Errorf("read lot.json: %w", err)
	}
	var doc lotJSON
	if err := json.Unmarshal(b, &doc); err != nil {
		return fmt.Errorf("parse lot.json: %w", err)
	}
	if doc.Width != lotW/tile || doc.Height != lotH/tile {
		return fmt.Errorf("lot size %dx%d want %dx%d", doc.Width, doc.Height, lotW/tile, lotH/tile)
	}
	if len(doc.Layers.Ground) != doc.Height || len(doc.Layers.Decal) != doc.Height {
		return fmt.Errorf("lot layers height mismatch")
	}
	for ty := 0; ty < doc.Height; ty++ {
		if len(doc.Layers.Ground[ty]) != doc.Width || len(doc.Layers.Decal[ty]) != doc.Width {
			return fmt.Errorf("lot row %d width mismatch", ty)
		}
	}
	a.ground = doc.Layers.Ground
	a.decal = doc.Layers.Decal
	return nil
}

func (a *atlas) cacheTiles() {
	for id := 0; id < tileCount; id++ {
		x := (id % tilesetCols) * tile
		y := (id / tilesetCols) * tile
		a.tileImg[id] = a.tileset.SubImage(image.Rect(x, y, x+tile, y+tile)).(*ebiten.Image)
	}
}

func (a *atlas) tile(id int) *ebiten.Image {
	if id < 0 || id >= tileCount {
		return nil
	}
	return a.tileImg[id]
}

func (a *atlas) frameImg(name string) (*ebiten.Image, frame, bool) {
	f, ok := a.frames[name]
	if !ok {
		return nil, frame{}, false
	}
	sub := a.sprites.SubImage(image.Rect(f.X, f.Y, f.X+f.W, f.Y+f.H)).(*ebiten.Image)
	return sub, f, true
}

// facingIndex snaps continuous heading to 16 facings (draw only).
// Matches game.FacingIndex: 0 = north, 4 = east, 8 = south, 12 = west.
func facingIndex(heading float64) int {
	const step = 2 * math.Pi / 16
	h := math.Mod(heading+step/2, 2*math.Pi)
	if h < 0 {
		h += 2 * math.Pi
	}
	return int(h / step)
}

func evenFacing(heading float64) int {
	i := facingIndex(heading)
	return (i / 2) * 2
}

func blitTile(dst *ebiten.Image, a *atlas, id int, sx, sy float64) {
	if id == 0 {
		return
	}
	img := a.tile(id)
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(sx, sy)
	op.Filter = ebiten.FilterNearest
	dst.DrawImage(img, op)
}

func blitFrame(dst *ebiten.Image, a *atlas, name string, cx, cy float64) {
	blitFrameTint(dst, a, name, cx, cy, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
}

func blitFrameTint(dst *ebiten.Image, a *atlas, name string, cx, cy float64, tint color.RGBA) {
	img, f, ok := a.frameImg(name)
	if !ok {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(cx-f.AnchorX, cy-f.AnchorY)
	op.Filter = ebiten.FilterNearest
	if tint.R != 255 || tint.G != 255 || tint.B != 255 || tint.A != 255 {
		op.ColorScale.ScaleWithColor(tint)
	}
	dst.DrawImage(img, op)
}

func blitFrameTL(dst *ebiten.Image, a *atlas, name string, sx, sy float64) {
	blitFrameTLScale(dst, a, name, sx, sy, 1, 1, 1)
}

func blitFrameTLScale(dst *ebiten.Image, a *atlas, name string, sx, sy, sr, sg, sb float64) {
	img, _, ok := a.frameImg(name)
	if !ok {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(sx, sy)
	op.Filter = ebiten.FilterNearest
	if sr != 1 || sg != 1 || sb != 1 {
		op.ColorScale.Scale(float32(sr), float32(sg), float32(sb), 1)
	}
	dst.DrawImage(img, op)
}
