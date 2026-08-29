package render

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
	"permitdenied/internal/dozer"
	"permitdenied/internal/fx"
	"permitdenied/internal/lot"
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
	hudFace  = text.NewGoXFace(basicfont.Face7x13)
	whiteImg *ebiten.Image
)

func white() *ebiten.Image {
	if whiteImg == nil {
		whiteImg = ebiten.NewImage(1, 1)
		whiteImg.Fill(color.White)
	}
	return whiteImg
}

type View struct {
	CamX, CamY, ShakeX, ShakeY float64
	Tick                       int
	Dozer                      dozer.Dozer
	Buildings                  []lot.Building
	Rubble                     []lot.Rubble
	Cruisers                   []threats.Cruiser
	Blockers                   []threats.Blocker
	Excavator                  threats.Excavator
	Chopper                    threats.Chopper
	Peds                       []threats.Ped
	Dollars                    []fx.Dollar
	Banner                     string
	BannerT                    float64
	RunTick                    int
	StructCash, VehicleCash    int
	Targets                    int
	PIT, Dump, Set, YardDown   bool
	Heat                       float64
	Plates                     int
	BladeDown                  bool
	Speed                      float64
	Debug                      bool
	Time                       float64
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
	fill(dst, 0, 0, screenW, screenH, ColBG)
	drawGround(dst, v)
	drawConcrete(dst, v)
	drawBuildingShadows(dst, v)
	drawBuildings(dst, v)
	drawBlockers(dst, v)
	drawPeds(dst, v)
	drawCruisers(dst, v)
	drawExcavator(dst, v)
	drawDozer(dst, v)
	drawChopper(dst, v)
	drawDollars(dst, v)
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

	stance := "BLADE UP"
	sc := ColStanceUp
	if v.BladeDown {
		stance = "BLADE DOWN"
		sc = ColStanceDn
	}
	drawText(dst, stance, 4, float64(screenH-14), sc)

	if v.Debug {
		line := fmt.Sprintf("T=%.1f PIT=%d DUMP=%d SET=%d YARD=%d",
			v.Time, b2i(v.PIT), b2i(v.Dump), b2i(v.Set), b2i(v.YardDown))
		drawText(dst, line, 4, 16, ColHUD)
		line2 := fmt.Sprintf("HEAT=%.0f PLATES=%d BLADE=%s SPD=%.0f",
			v.Heat, v.Plates, bladeWord(v.BladeDown), v.Speed)
		drawText(dst, line2, 4, 28, ColHUD)
	}
}

func hudMult(targets int) float64 {
	switch {
	case targets <= 0:
		return 1.0
	case targets == 1:
		return 1.25
	case targets == 2:
		return 1.6
	default:
		return 2.0
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

func DrawTitle(dst *ebiten.Image) {
	fill(dst, 0, 0, screenW, screenH, ColBG)
	// dirt strip suggestion
	fill(dst, 0, 140, screenW, 84, ColDirt)
	fill(dst, 120, 0, 80, screenH, ColAsphalt)

	drawRotated(dst, screenW/2, 118, 18, 28, 0.15, ColPaint)
	drawRotated(dst, screenW/2, 118-16, 32, 8, 0.15, ColBlade)

	drawTextScaled(dst, "PERMIT DENIED", 52, 36, 2, ColPaint)
	drawText(dst, "THE COUNTY SAID NO.", 86, 72, ColHUD)
	drawText(dst, "A/D STEER  W GO  S BACK", 68, 168, ColHUD)
	drawText(dst, "SPACE BLADE", 112, 180, ColHUD)
	drawText(dst, "PRESS SPACE", 112, 200, ColPaint)
}

func DrawTally(dst *ebiten.Image, t Tally) {
	fill(dst, 40, 28, 240, 168, ColPanel)
	death := "COUNTY CLOCK"
	switch t.Death {
	case "cooked":
		death = "ENGINE COOKED"
	case "track":
		death = "TRACK THROWN"
	}
	drawText(dst, death, 86, 36, ColRust)

	y := 56.0
	if t.T >= 0.2 {
		drawText(dst, fmt.Sprintf("STRUCTURE  $%d", t.StructCash), 70, y, ColMoney)
	}
	y += 16
	if t.T >= 0.45 {
		drawText(dst, fmt.Sprintf("VEHICLE    $%d", t.VehicleCash), 70, y, ColMoney)
	}
	y += 16
	if t.T >= 0.7 {
		drawText(dst, fmt.Sprintf("TIME       %d", int(t.Time)), 70, y, ColHUD)
	}
	y += 16
	if t.T >= 0.95 {
		drawText(dst, fmt.Sprintf("TARGETS    %d ×%s", t.Targets, multLabel(t.Targets)), 70, y, ColPlantPip)
	}
	y += 20
	if t.T >= 1.2 {
		drawText(dst, fmt.Sprintf("TOTAL      %d", t.Total), 70, y, ColPaint)
	}
	if t.T >= 1.4 {
		drawText(dst, "SPACE / TAP — AGAIN", 78, 176, ColHUD)
	}
}

func drawGround(dst *ebiten.Image, v View) {
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
			n := tileHash(tx, ty)
			dirt := ColDirt
			if n&3 == 0 {
				dirt = shade(dirt, -8)
			} else if n&3 == 1 {
				dirt = shade(dirt, 8)
			}
			col := dirt
			// drag asphalt x=240..400
			if wx >= 240 && wx < 400 {
				col = ColAsphalt
				if n&7 == 0 {
					col = shade(col, 10)
				}
			}
			// plant pad 200,48,240,200
			if wx >= 200 && wx < 440 && wy >= 48 && wy < 248 {
				col = ColPad
			}
			fill(dst, sx, sy, tile, tile, col)
		}
	}
	// rail spur y=200..216, x=400..640
	rx, ry := world(v, 400, 200)
	fill(dst, rx, ry, 240, 16, ColRail)
	for i := 0; i < 15; i++ {
		tx, ty := world(v, 400+float64(i*16), 200)
		fill(dst, tx, ty, 4, 16, ColTie)
	}
}

func drawConcrete(dst *ebiten.Image, v View) {
	for _, b := range v.Blockers {
		if b.Kind != threats.BlockerConcrete || !b.Alive {
			continue
		}
		sx, sy := world(v, b.X, b.Y)
		c := ColWet
		if b.Set {
			c = ColSet
		}
		fill(dst, sx, sy, b.W, b.H, c)
	}
}

func drawBuildingShadows(dst *ebiten.Image, v View) {
	for _, b := range v.Buildings {
		sx, sy := world(v, b.X+2, b.Y+2)
		fill(dst, sx, sy, b.W, b.H, ColShadow)
	}
}

func drawBuildings(dst *ebiten.Image, v View) {
	for _, b := range v.Buildings {
		sx, sy := world(v, b.X, b.Y)
		switch b.State {
		case lot.Intact:
			fill(dst, sx, sy, b.W, b.H, ColBuilding)
		case lot.Cracked:
			fill(dst, sx, sy, b.W, b.H, shade(ColBuilding, -24))
			vector.StrokeLine(dst, float32(sx+2), float32(sy+2), float32(sx+b.W-2), float32(sy+b.H-2), 1, ColCrack, false)
			vector.StrokeLine(dst, float32(sx+b.W-2), float32(sy+2), float32(sx+2), float32(sy+b.H-2), 1, ColCrack, false)
		case lot.InRubble:
			fill(dst, sx, sy, b.W, b.H, ColRubble)
			scatter(dst, v, b.X, b.Y, b.W, b.H)
		}
		if b.Label != "" && b.State != lot.InRubble {
			drawText(dst, b.Label, sx+4, sy+4, ColLabel)
		}
	}
}

func scatter(dst *ebiten.Image, v View, x, y, w, h float64) {
	n := int(w*h) / 80
	if n < 3 {
		n = 3
	}
	if n > 12 {
		n = 12
	}
	for i := 0; i < n; i++ {
		hsh := tileHash(int(x)+i*13, int(y)+i*7)
		ox := float64(hsh % uint32(w-4))
		oy := float64((hsh >> 8) % uint32(h-4))
		sx, sy := world(v, x+ox, y+oy)
		fill(dst, sx, sy, 4, 3, shade(ColRubble, int(hsh%20)-10))
	}
}

func drawBlockers(dst *ebiten.Image, v View) {
	for _, b := range v.Blockers {
		if !b.Alive || b.Kind == threats.BlockerConcrete {
			continue
		}
		sx, sy := world(v, b.X, b.Y)
		switch b.Kind {
		case threats.BlockerJersey:
			fill(dst, sx, sy, b.W, b.H, ColJersey)
			fill(dst, sx, sy, 4, b.H, ColPaint)
		case threats.BlockerDump:
			fill(dst, sx, sy, b.W, b.H, ColDump)
			fill(dst, sx+4, sy+2, b.W-8, 6, shade(ColDump, 20))
		}
	}
}

func drawPeds(dst *ebiten.Image, v View) {
	for _, p := range v.Peds {
		if !p.Alive {
			continue
		}
		sx, sy := world(v, p.X-2, p.Y-3)
		fill(dst, sx, sy, 4, 6, ColPed)
	}
}

func drawCruisers(dst *ebiten.Image, v View) {
	for _, c := range v.Cruisers {
		if !c.Alive {
			continue
		}
		sx, sy := world(v, c.X, c.Y)
		drawRotated(dst, sx, sy, 10, 16, c.Heading, ColCruiser)
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

func drawExcavator(dst *ebiten.Image, v View) {
	ex := v.Excavator
	if ex.Announced && !ex.Arrived {
		// dust puff at yard
		sx, sy := world(v, 448, 720)
		fill(dst, sx+20, sy+40, 10, 6, shade(ColDirt, 20))
		fill(dst, sx+36, sy+48, 14, 8, shade(ColDirt, 10))
	}
	if !ex.Arrived || !ex.Alive {
		return
	}
	sx, sy := world(v, ex.X, ex.Y)
	drawRotated(dst, sx, sy, 20, 36, ex.Heading, ColEx)
	ang := (ex.BoomPhase - 0.5) * 1.2
	hdg := ex.Heading + ang
	fx, fy := math.Sin(hdg), -math.Cos(hdg)
	x1 := sx + fx*48
	y1 := sy + fy*48
	vector.StrokeLine(dst, float32(sx), float32(sy), float32(x1), float32(y1), 3, ColBoom, false)
	fill(dst, x1-3, y1-3, 6, 6, ColFrame)
}

func drawDozer(dst *ebiten.Image, v View) {
	d := v.Dozer
	sx, sy := world(v, d.X, d.Y)
	base := PlateColor(d.Plates)
	t := d.Heat / 100
	col := mix(base, ColHeat, t*0.85)
	if d.Heat > 90 && v.Tick%12 < 6 {
		col = mix(col, ColHeat, 0.5)
	}
	if d.IFrames > 0 && v.Tick%4 < 2 {
		col = mix(col, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}, 0.35)
	}
	// tracks
	rx, ry := math.Cos(d.Heading), math.Sin(d.Heading)
	drawRotated(dst, sx+rx*8, sy+ry*8, 5, 26, d.Heading, ColTrack)
	drawRotated(dst, sx-rx*8, sy-ry*8, 5, 26, d.Heading, ColTrack)
	drawRotated(dst, sx, sy, 16, 24, d.Heading, col)
	bh := 6.0
	if d.BladeDown {
		bh = 10
	}
	drawRotatedOffset(dst, sx, sy, 32, bh, d.Heading, 0, -18, ColBlade)
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

func drawChopper(dst *ebiten.Image, v View) {
	c := v.Chopper
	if !c.Active {
		return
	}
	sx, sy := world(v, c.X, c.Y)
	// spotlight under chopper sprite (order: spotlight then diamond)
	vector.FillCircle(dst, float32(sx), float32(sy), float32(c.SpotR), ColSpot, true)
	drawRotated(dst, sx, sy, 16, 10, 0, ColChopper)
	spin := float64(v.Tick) * 0.7
	vector.StrokeLine(dst,
		float32(sx+math.Cos(spin)*10), float32(sy+math.Sin(spin)*10),
		float32(sx-math.Cos(spin)*10), float32(sy-math.Sin(spin)*10),
		1, ColHUD, false)
}

func drawDollars(dst *ebiten.Image, v View) {
	for _, d := range v.Dollars {
		rise := (0.7 - d.Life) / 0.7 * 20
		if rise < 0 {
			rise = 0
		}
		sx, sy := world(v, d.X, d.Y-rise)
		drawText(dst, fmt.Sprintf("+$%d", d.Amt), sx-12, sy-8, ColMoney)
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
	for _, b := range v.Buildings {
		if b.ID == lot.TargetNone || b.State == lot.InRubble {
			continue
		}
		cx, cy := b.Center()
		var col color.RGBA
		switch b.ID {
		case lot.TargetSheriff:
			col = ColAmber
		case lot.TargetYard:
			col = ColRust
		default:
			col = ColPlantPip
		}
		px, py := rimPoint(v, cx, cy)
		fill(dst, px-2, py-2, 4, 4, col)
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

func drawRotated(dst *ebiten.Image, cx, cy, w, h, heading float64, c color.Color) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w, h)
	op.GeoM.Translate(-w/2, -h/2)
	op.GeoM.Rotate(heading)
	op.GeoM.Translate(cx, cy)
	op.Filter = ebiten.FilterNearest
	op.ColorScale.ScaleWithColor(c)
	dst.DrawImage(white(), op)
}

func drawRotatedOffset(dst *ebiten.Image, cx, cy, w, h, heading, ox, oy float64, c color.Color) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w, h)
	op.GeoM.Translate(-w/2, -h/2)
	op.GeoM.Translate(ox, oy)
	op.GeoM.Rotate(heading)
	op.GeoM.Translate(cx, cy)
	op.Filter = ebiten.FilterNearest
	op.ColorScale.ScaleWithColor(c)
	dst.DrawImage(white(), op)
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

func tileHash(tx, ty int) uint32 {
	n := uint32(tx)*374761393 ^ uint32(ty)*668265263
	n = (n << 13) ^ n
	n *= 1274126177
	return n
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
