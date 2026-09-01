package render

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
	"permitdenied/internal/attach"
	"permitdenied/internal/dozer"
	"permitdenied/internal/fx"
	"permitdenied/internal/lot"
	"permitdenied/internal/meta"
	"permitdenied/internal/threats"
)

const (
	screenW = 320
	screenH = 224
	lotW    = 640
	lotH    = 1280
	tile    = 16
)

var (
	hudFace = text.NewGoXFace(basicfont.Face7x13)
)

type Rect struct {
	X, Y, W, H float64
}

type View struct {
	CamX, CamY, ShakeX, ShakeY float64
	Tick                       int
	Dozer                      dozer.Dozer
	Buildings                  []lot.Building
	Rubble                     []lot.Rubble
	Cruisers                   []threats.Cruiser
	Blockers                   []threats.Blocker
	Heavies                    []threats.Heavy
	Pickups                    []attach.Pickup
	Streets                    []Rect
	Excavator                  threats.Excavator
	Chopper                    threats.Chopper
	Peds                       []threats.Ped
	Dollars                    []fx.Dollar
	Bursts                     []fx.Burst
	Banner                     string
	BannerT                    float64
	RunTick                    int
	StructCash, VehicleCash    int
	Targets                    int
	NamedDown, NamedTotal      int
	PIT, Dump, Set, YardDown   bool
	Heat                       float64
	Plates                     int
	BladeDown                  bool
	Speed                      float64
	Debug                      bool
	Time                       float64
	HideStance                 bool
	MapW, MapH                 float64
	Procedural                 bool
	Tier                       int
	KitCount                   int
}

type Tally struct {
	T           float64
	Death       string
	StructCash  int
	VehicleCash int
	Time        float64
	Targets     int
	Mult        float64
	Total       int
}

func DrawWorld(dst *ebiten.Image, v View) {
	a, err := ensureAtlas()
	if err != nil {
		panic(err)
	}
	fill(dst, 0, 0, screenW, screenH, ColBG)
	drawGround(dst, v, a)
	drawConcrete(dst, v, a)
	drawBuildingShadows(dst, v)
	drawBuildings(dst, v, a)
	drawRubble(dst, v, a)
	drawBlockers(dst, v, a)
	drawHeavies(dst, v)
	drawPickups(dst, v)
	drawPeds(dst, v, a)
	drawCruisers(dst, v, a)
	drawExcavator(dst, v, a)
	drawDozer(dst, v, a)
	drawChopper(dst, v, a)
	drawBursts(dst, v, a)
	drawDollars(dst, v, a)
	drawBanner(dst, v)
}

func DrawHUD(dst *ebiten.Image, v View) {
	elapsed := int(float64(v.RunTick) / 60.0)
	if elapsed < 0 {
		elapsed = 0
	}
	mm, ss := elapsed/60, elapsed%60
	drawText(dst, fmt.Sprintf("%d:%02d", mm, ss), 4, 2, ColHUD)

	cash := v.StructCash + v.VehicleCash
	right := fmt.Sprintf("$%d", cash)
	if v.Targets > 0 {
		right = fmt.Sprintf("$%d ×%s", cash, multLabel(v.Targets))
	}
	drawText(dst, right, float64(screenW-4-len(right)*7), 2, ColMoney)

	drawPips(dst, v)
	if v.NamedTotal > 0 && v.Tier > 0 {
		drawText(dst, fmt.Sprintf("%d/%d", v.NamedDown, v.NamedTotal), 4, 14, ColHUD)
	}

	if !v.HideStance {
		stance := "BLADE UP"
		sc := ColStanceUp
		if v.BladeDown {
			stance = "BLADE DOWN"
			sc = ColStanceDn
		}
		drawText(dst, stance, 4, float64(screenH-14), sc)
	}

	if v.Debug {
		line := fmt.Sprintf("T=%.1f PIT=%d DUMP=%d SET=%d YARD=%d",
			v.Time, b2i(v.PIT), b2i(v.Dump), b2i(v.Set), b2i(v.YardDown))
		drawText(dst, line, 4, 16, ColHUD)
		line2 := fmt.Sprintf("HEAT=%.0f PLATES=%d BLADE=%s SPD=%.0f",
			v.Heat, v.Plates, bladeWord(v.BladeDown), v.Speed)
		drawText(dst, line2, 4, 28, ColHUD)
	}
}

func multLabel(targets int) string {
	switch targets {
	case 1:
		return "1.25"
	case 2:
		return "1.6"
	default:
		return "2"
	}
}

func DrawTitle(dst *ebiten.Image, bestCash, bestTier int) {
	a, err := ensureAtlas()
	if err != nil {
		panic(err)
	}
	fill(dst, 0, 0, screenW, screenH, ColBG)
	fill(dst, 0, 140, screenW, 84, ColDirt)
	fill(dst, 120, 0, 80, screenH, ColAsphalt)

	blitFrame(dst, a, "dozer_up_00", screenW/2, 118)

	drawTextScaled(dst, "PERMIT DENIED", 52, 36, 2, ColPaint)
	drawText(dst, "THE COUNTY SAID NO.", 86, 72, ColHUD)
	drawText(dst, fmt.Sprintf("BEST $%d  REACH %s", bestCash, tierName(bestTier)), 70, 88, ColMoney)
	drawText(dst, "A/D STEER  W GO  S BACK", 68, 168, ColHUD)
	drawText(dst, "SPACE BLADE", 112, 180, ColHUD)
	drawText(dst, "PRESS SPACE", 112, 200, ColPaint)
}

func tierName(t int) string {
	switch t {
	case 1:
		return "COUNTY"
	case 2:
		return "TOWN"
	case 3:
		return "CITY"
	case 4:
		return "CAPITOL"
	default:
		return "—"
	}
}

func DrawMeta(dst *ebiten.Image, s meta.Save, news []string) {
	fill(dst, 0, 0, screenW, screenH, ColBG)
	drawText(dst, "FILED AMENDMENTS", 8, 8, ColPaint)
	drawText(dst, "THE COUNTY SAID NO.", 8, 22, ColHUD)
	newset := map[string]bool{}
	for _, n := range news {
		newset[n] = true
	}
	y := 40.0
	for _, id := range meta.Order {
		mark := "[ ]"
		col := ColHUD
		if s.Has(id) {
			mark = "[X]"
			col = ColMoney
		}
		if newset[id] {
			col = ColPaint
		}
		drawText(dst, mark+" "+s.Label(id), 8, y, col)
		y += 14
	}
	drawText(dst, "SPACE — FILE CLOSED", 8, 208, ColPaint)
}

const (
	tallyStripY = 136
	tallyStripH = 88
	tallyFirstY = 16
	tallyLineH  = 14
)

func tallyStatY(i int) float64 {
	return float64(tallyStripY + tallyFirstY + i*tallyLineH)
}

func DrawTally(dst *ebiten.Image, t Tally) {
	fill(dst, 0, tallyStripY, screenW, tallyStripH, ColPanel)
	death := "COUNTY CLOCK"
	switch t.Death {
	case "cooked":
		death = "ENGINE COOKED"
	case "track":
		death = "TRACK THROWN"
	case "pinned":
		death = "PINNED"
	case "buried":
		death = "BURIED"
	case "cleared":
		death = "CLEARED"
	case "buzzer":
		death = "COUNTY CLOCK"
	}
	drawText(dst, death, 8, tallyStripY+2, ColRust)

	if t.T >= 0.2 {
		drawText(dst, fmt.Sprintf("STRUCTURE  $%d", t.StructCash), 8, tallyStatY(0), ColMoney)
	}
	if t.T >= 0.45 {
		drawText(dst, fmt.Sprintf("VEHICLE    $%d", t.VehicleCash), 8, tallyStatY(1), ColMoney)
	}
	if t.T >= 0.7 {
		drawText(dst, fmt.Sprintf("TIME       %d", int(t.Time)), 8, tallyStatY(2), ColHUD)
	}
	if t.T >= 0.95 {
		drawText(dst, fmt.Sprintf("TARGETS    %d ×%s", t.Targets, multLabel(t.Targets)), 8, tallyStatY(3), ColPlantPip)
	}
	if t.T >= 1.2 {
		drawText(dst, fmt.Sprintf("TOTAL      %d", t.Total), 8, tallyStatY(4), ColPaint)
	}
	if t.T >= 1.4 {
		drawText(dst, "THE COUNTY SAID NO.", 148, float64(tallyStripY+2), ColHUD)
		drawText(dst, "SPACE / TAP — AGAIN", 148, tallyStatY(4), ColHUD)
	}
}

func drawGround(dst *ebiten.Image, v View, a *atlas) {
	if v.Procedural {
		drawProceduralGround(dst, v)
		return
	}
	x0 := int(v.CamX)/tile - 1
	y0 := int(v.CamY)/tile - 1
	x1 := int(v.CamX+screenW)/tile + 2
	y1 := int(v.CamY+screenH)/tile + 2
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > lotW/tile {
		x1 = lotW / tile
	}
	if y1 > lotH/tile {
		y1 = lotH / tile
	}
	for ty := y0; ty < y1; ty++ {
		for tx := x0; tx < x1; tx++ {
			wx := float64(tx * tile)
			wy := float64(ty * tile)
			sx, sy := world(v, wx, wy)
			blitTile(dst, a, a.ground[ty][tx], sx, sy)
			decal := a.decal[ty][tx]
			if decal != 0 {
				blitTile(dst, a, decal, sx, sy)
			}
		}
	}
}

func drawConcrete(dst *ebiten.Image, v View, a *atlas) {
	// Wet concrete (id 14) is already on the decal layer.
	// When a patch sets, overwrite those cells with concrete_set (id 15).
	for _, b := range v.Blockers {
		if b.Kind != threats.BlockerConcrete || !b.Alive || !b.Set {
			continue
		}
		tx0 := int(b.X) / tile
		ty0 := int(b.Y) / tile
		tx1 := int(b.X+b.W-1)/tile + 1
		ty1 := int(b.Y+b.H-1)/tile + 1
		for ty := ty0; ty < ty1; ty++ {
			for tx := tx0; tx < tx1; tx++ {
				if ty < 0 || ty >= len(a.decal) || tx < 0 || tx >= len(a.decal[ty]) {
					continue
				}
				if a.decal[ty][tx] != tileWet {
					continue
				}
				sx, sy := world(v, float64(tx*tile), float64(ty*tile))
				blitTile(dst, a, tileSet, sx, sy)
			}
		}
	}
}

func drawBuildingShadows(dst *ebiten.Image, v View) {
	for _, b := range v.Buildings {
		if b.State == lot.InRubble {
			continue
		}
		sx, sy := world(v, b.X+2, b.Y+2)
		fill(dst, sx, sy, b.W, b.H, ColShadow)
	}
}

func drawRubble(dst *ebiten.Image, v View, a *atlas) {
	for _, r := range v.Rubble {
		if v.Procedural {
			sx, sy := world(v, r.X, r.Y)
			c := ColRubble
			if r.Ramp {
				c = shade(ColRubble, 24)
			}
			fill(dst, sx, sy, r.W, r.H, c)
			continue
		}
		stampAABB(dst, v, a, "building_rubble", r.X, r.Y, r.W, r.H)
	}
}

func drawProceduralGround(dst *ebiten.Image, v View) {
	fill(dst, 0, 0, screenW, screenH, ColDirt)
	for _, s := range v.Streets {
		sx, sy := world(v, s.X, s.Y)
		fill(dst, sx, sy, s.W, s.H, ColAsphalt)
	}
	if len(v.Streets) == 0 {
		fill(dst, 0, 0, screenW, screenH, ColAsphalt)
	}
}

func drawProceduralBuildings(dst *ebiten.Image, v View) {
	for _, b := range v.Buildings {
		if b.State == lot.InRubble {
			continue
		}
		sx, sy := world(v, b.X, b.Y)
		c := matColor(b.Material)
		if b.State == lot.Cracked {
			c = shade(c, -30)
		}
		fill(dst, sx, sy, b.W, b.H, c)
		vector.StrokeRect(dst, float32(sx), float32(sy), float32(b.W), float32(b.H), 1, ColFrame, false)
		if b.Label != "" {
			drawText(dst, b.Label, sx+3, sy+3, ColLabel)
		}
	}
}

func matColor(m lot.Material) color.RGBA {
	switch m {
	case lot.MatWood:
		return ColManila
	case lot.MatBrick:
		return ColRust
	case lot.MatConcrete:
		return ColBuilding
	case lot.MatSteel:
		return ColSteel
	default:
		return ColBuilding
	}
}

func drawHeavies(dst *ebiten.Image, v View) {
	for _, h := range v.Heavies {
		if !h.Alive {
			continue
		}
		x, y, w, hh := h.Body()
		sx, sy := world(v, x, y)
		c := ColFireTruck
		if h.Kind == threats.HeavyWagon {
			c = ColWagon
		}
		fill(dst, sx, sy, w, hh, c)
	}
}

func drawPickups(dst *ebiten.Image, v View) {
	for _, p := range v.Pickups {
		if !p.Alive {
			continue
		}
		sx, sy := world(v, p.X, p.Y)
		fill(dst, sx, sy, p.W, p.H, ColPaint)
		drawText(dst, p.Kind.Label()[:1], sx+4, sy+2, ColFrame)
	}
}

func drawBuildings(dst *ebiten.Image, v View, a *atlas) {
	if v.Procedural {
		drawProceduralBuildings(dst, v)
		return
	}
	for _, b := range v.Buildings {
		if b.State == lot.InRubble {
			continue
		}
		name := "building_intact"
		switch b.State {
		case lot.Intact:
			name = "building_intact"
		case lot.Cracked:
			name = "building_cracked"
		default:
			name = "building_intact"
		}
		if b.State == lot.Cracked {
			stampCrackedAABB(dst, v, a, name, b.X, b.Y, b.W, b.H)
		} else {
			stampAABB(dst, v, a, name, b.X, b.Y, b.W, b.H)
		}
		if b.Label != "" {
			sx, sy := world(v, b.X, b.Y)
			drawText(dst, b.Label, sx+4, sy+4, ColLabel)
		}
	}
}

func stampAABB(dst *ebiten.Image, v View, a *atlas, name string, x, y, w, h float64) {
	for oy := 0.0; oy < h; oy += float64(tile) {
		for ox := 0.0; ox < w; ox += float64(tile) {
			sx, sy := world(v, x+ox, y+oy)
			blitFrameTL(dst, a, name, sx, sy)
		}
	}
}

func stampCrackedAABB(dst *ebiten.Image, v View, a *atlas, name string, x, y, w, h float64) {
	const dim = 0.62
	for oy := 0.0; oy < h; oy += float64(tile) {
		for ox := 0.0; ox < w; ox += float64(tile) {
			sx, sy := world(v, x+ox, y+oy)
			blitFrameTLScale(dst, a, name, sx, sy, dim, dim, dim)
		}
	}
	sx, sy := world(v, x, y)
	vector.StrokeLine(dst, float32(sx), float32(sy), float32(sx+w), float32(sy+h), 1, ColCrack, false)
	vector.StrokeLine(dst, float32(sx+w), float32(sy), float32(sx), float32(sy+h), 1, ColCrack, false)
}

func drawBlockers(dst *ebiten.Image, v View, a *atlas) {
	for _, b := range v.Blockers {
		if !b.Alive || b.Kind == threats.BlockerConcrete {
			continue
		}
		switch b.Kind {
		case threats.BlockerJersey:
			name := "jersey"
			if v.YardDown || b.HP <= 1 {
				name = "jersey_broken"
			}
			stampAABB(dst, v, a, name, b.X, b.Y, b.W, b.H)
		case threats.BlockerDump:
			sx, sy := world(v, b.X+b.W/2, b.Y+b.H/2)
			blitFrame(dst, a, "dump_00", sx, sy)
		}
	}
}

func drawPeds(dst *ebiten.Image, v View, a *atlas) {
	for _, p := range v.Peds {
		if !p.Alive {
			continue
		}
		sx, sy := world(v, p.X, p.Y)
		blitFrame(dst, a, "ped", sx, sy)
	}
}

func drawCruisers(dst *ebiten.Image, v View, a *atlas) {
	for _, c := range v.Cruisers {
		if !c.Alive {
			continue
		}
		sx, sy := world(v, c.X, c.Y)
		name := fmt.Sprintf("cruiser_%02d", facingIndex(c.Heading))
		blitFrame(dst, a, name, sx, sy)
		on := v.Tick%20 < 10
		if on {
			fill(dst, sx-3, sy-6, 2, 2, ColSiren)
			fill(dst, sx+1, sy-6, 2, 2, color.RGBA{0x3A, 0x6F, 0xBF, 0xFF})
		} else {
			fill(dst, sx-3, sy-6, 2, 2, ColCruiser)
			fill(dst, sx+1, sy-6, 2, 2, ColSiren)
		}
	}
}

func drawExcavator(dst *ebiten.Image, v View, a *atlas) {
	ex := v.Excavator
	if ex.Announced && !ex.Arrived {
		sx, sy := world(v, 448, 720)
		fill(dst, sx+20, sy+40, 10, 6, shade(ColDirt, 20))
		fill(dst, sx+36, sy+48, 14, 8, shade(ColDirt, 10))
	}
	if !ex.Arrived || !ex.Alive {
		return
	}
	sx, sy := world(v, ex.X, ex.Y)
	name := fmt.Sprintf("excavator_%02d", evenFacing(ex.Heading))
	blitFrame(dst, a, name, sx, sy)
	ang := (ex.BoomPhase - 0.5) * 1.2
	hdg := ex.Heading + ang
	fx, fy := math.Sin(hdg), -math.Cos(hdg)
	x1 := sx + fx*48
	y1 := sy + fy*48
	vector.StrokeLine(dst, float32(sx), float32(sy), float32(x1), float32(y1), 3, ColBoom, false)
	fill(dst, x1-3, y1-3, 6, 6, ColFrame)
}

func drawDozer(dst *ebiten.Image, v View, a *atlas) {
	d := v.Dozer
	sx, sy := world(v, d.X, d.Y)
	stance := "up"
	if d.BladeDown {
		stance = "down"
	}
	name := fmt.Sprintf("dozer_%s_%02d", stance, facingIndex(d.Heading))

	base := PlateColor(d.Plates)
	t := d.Heat / 100
	col := mix(base, ColHeat, t*0.85)
	if d.Heat > 90 && v.Tick%12 < 6 {
		col = mix(col, ColHeat, 0.5)
	}
	if d.IFrames > 0 && v.Tick%4 < 2 {
		col = mix(col, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}, 0.35)
	}

	img, f, ok := a.frameImg(name)
	if !ok {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(sx-f.AnchorX, sy-f.AnchorY)
	op.Filter = ebiten.FilterNearest
	// Remap paint-yellow atlas pixels toward plate/heat tint.
	paint := ColPaint
	op.ColorScale.Scale(
		float32(col.R)/float32(paint.R),
		float32(col.G)/float32(paint.G),
		float32(col.B)/float32(paint.B),
		1,
	)
	dst.DrawImage(img, op)

	if d.Heat > 70 && v.Tick%12 < 6 {
		fx, fy := math.Sin(d.Heading), -math.Cos(d.Heading)
		px := sx - fy*6 - fx*4
		py := sy + fx*6 - fy*4
		fill(dst, px, py, 1, 3, ColHeat)
		px2 := sx + fy*6 - fx*4
		py2 := sy - fx*6 - fy*4
		fill(dst, px2, py2, 1, 3, ColHeat)
	}
}

func drawChopper(dst *ebiten.Image, v View, a *atlas) {
	c := v.Chopper
	if !c.Active {
		return
	}
	sx, sy := world(v, c.X, c.Y)
	vector.FillCircle(dst, float32(sx), float32(sy), float32(c.SpotR), ColSpot, true)
	frame := (v.Tick / 2) % 4
	blitFrame(dst, a, fmt.Sprintf("chopper_%d", frame), sx, sy)
}

func drawBursts(dst *ebiten.Image, v View, a *atlas) {
	for _, b := range v.Bursts {
		name := b.FrameName()
		if name == "boom_05" {
			continue
		}
		sx, sy := world(v, b.X, b.Y)
		blitFrame(dst, a, name, sx, sy)
	}
}

func drawDollars(dst *ebiten.Image, v View, a *atlas) {
	for _, d := range v.Dollars {
		rise := (0.7 - d.Life) / 0.7 * 20
		if rise < 0 {
			rise = 0
		}
		sx, sy := world(v, d.X, d.Y-rise)
		blitFrame(dst, a, "dollar", sx, sy)
		drawText(dst, fmt.Sprintf("+$%d", d.Amt), sx-12, sy-12, ColMoney)
	}
}

func drawBanner(dst *ebiten.Image, v View) {
	if v.Banner == "" || v.BannerT <= 0 {
		return
	}
	fill(dst, 8, 40, screenW-16, 16, ColPanel)
	drawText(dst, v.Banner, 14, 42, ColPaint)
}

func drawPips(dst *ebiten.Image, v View) {
	a, err := ensureAtlas()
	if err != nil {
		return
	}
	for _, b := range v.Buildings {
		if b.ID == lot.TargetNone || b.State == lot.InRubble {
			continue
		}
		cx, cy := b.Center()
		name := "pip_cyan"
		switch b.ID {
		case lot.TargetSheriff:
			name = "pip_amber"
		case lot.TargetYard:
			name = "pip_rust"
		case lot.TargetPlant:
			name = "pip_cyan"
		}
		px, py := rimPoint(v, cx, cy)
		blitFrame(dst, a, name, px, py)
	}
}

func rimPoint(v View, wx, wy float64) (float64, float64) {
	sx := wx - v.CamX + v.ShakeX
	sy := wy - v.CamY + v.ShakeY
	cx, cy := screenW/2.0, screenH/2.0
	dx, dy := sx-cx, sy-cy
	if dx == 0 && dy == 0 {
		return cx, 3
	}
	const m = 3.0
	hw, hh := cx-m, cy-m
	tx, ty := 1e9, 1e9
	if dx != 0 {
		tx = hw / math.Abs(dx)
	}
	if dy != 0 {
		ty = hh / math.Abs(dy)
	}
	t := math.Min(tx, ty)
	return cx + dx*t, cy + dy*t
}

func world(v View, x, y float64) (float64, float64) {
	return x - v.CamX + v.ShakeX, y - v.CamY + v.ShakeY
}

func fill(dst *ebiten.Image, x, y, w, h float64, c color.Color) {
	vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), c, false)
}

func drawText(dst *ebiten.Image, s string, x, y float64, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	op.Filter = ebiten.FilterNearest
	text.Draw(dst, s, hudFace, op)
}

func drawTextScaled(dst *ebiten.Image, s string, x, y, scale float64, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	op.Filter = ebiten.FilterNearest
	text.Draw(dst, s, hudFace, op)
}

func shade(c color.RGBA, d int) color.RGBA {
	add := func(v uint8) uint8 {
		n := int(v) + d
		if n < 0 {
			n = 0
		}
		if n > 255 {
			n = 255
		}
		return uint8(n)
	}
	return color.RGBA{add(c.R), add(c.G), add(c.B), c.A}
}

func b2i(v bool) int {
	if v {
		return 1
	}
	return 0
}

func bladeWord(down bool) string {
	if down {
		return "DOWN"
	}
	return "UP"
}
