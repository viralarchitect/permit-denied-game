package render

import (
	"fmt"
	"math"
	"testing"

	"permitdenied/assets"
	"permitdenied/internal/dozer"
	"permitdenied/internal/fx"
	"permitdenied/internal/lot"
	"permitdenied/internal/threats"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestEmbedFilesExist(t *testing.T) {
	for _, path := range []string{
		"usable/lot.json",
		"usable/tileset.png",
		"usable/sprites.png",
		"usable/sprites.json",
	} {
		b, err := assets.FS.ReadFile(path)
		if err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
		if len(b) == 0 {
			t.Fatalf("%s is empty", path)
		}
	}
}

func TestAtlasLoad(t *testing.T) {
	a, err := ensureAtlas()
	if err != nil {
		t.Fatal(err)
	}
	if len(a.ground) != 80 || len(a.decal) != 80 {
		t.Fatalf("lot height: ground=%d decal=%d want 80", len(a.ground), len(a.decal))
	}
	if len(a.ground[0]) != 40 || len(a.decal[0]) != 40 {
		t.Fatalf("lot width: ground=%d decal=%d want 40", len(a.ground[0]), len(a.decal[0]))
	}
	for id := 0; id < tileCount; id++ {
		if a.tile(id) == nil {
			t.Fatalf("missing tile %d", id)
		}
	}
	required := []string{
		"dozer_up_00", "dozer_up_15", "dozer_down_00", "dozer_down_15",
		"cruiser_00", "cruiser_15", "dump_00", "jersey", "jersey_broken",
		"excavator_00", "excavator_14", "chopper_0", "chopper_3",
		"ped", "pip_amber", "pip_rust", "pip_cyan",
		"building_intact", "building_cracked", "building_rubble", "dollar",
		"boom_00", "boom_04", "spark_00", "spark_03",
		"fire_00", "fire_14", "wagon_00", "wagon_14",
	}
	for _, name := range required {
		if _, ok := a.frames[name]; !ok {
			t.Fatalf("missing frame %s", name)
		}
	}
	for i := 0; i < 16; i++ {
		up := fmt.Sprintf("dozer_up_%02d", i)
		down := fmt.Sprintf("dozer_down_%02d", i)
		if _, ok := a.frames[up]; !ok {
			t.Fatalf("missing %s", up)
		}
		if _, ok := a.frames[down]; !ok {
			t.Fatalf("missing %s", down)
		}
	}
}

func TestFacingIndexCardinals(t *testing.T) {
	if facingIndex(0) != 0 {
		t.Fatalf("north: %d", facingIndex(0))
	}
	if facingIndex(math.Pi/2) != 4 {
		t.Fatalf("east: %d", facingIndex(math.Pi/2))
	}
	if facingIndex(math.Pi) != 8 {
		t.Fatalf("south: %d", facingIndex(math.Pi))
	}
	if facingIndex(3*math.Pi/2) != 12 {
		t.Fatalf("west: %d", facingIndex(3*math.Pi/2))
	}
}

func TestDrawWorldSmoke(t *testing.T) {
	dst := ebiten.NewImage(screenW, screenH)
	l := lot.New(2)
	d := dozer.Spawn(320, 1180)
	DrawWorld(dst, View{
		CamX:      160,
		CamY:      1068,
		Dozer:     d,
		Buildings: l.Buildings,
		Heavies: []threats.Heavy{
			threats.FireTruck(300, 1100, 28, 16, 22),
			threats.Wagon(340, 1120, 48, 20, 36),
		},
	})
	DrawTitle(dst, 0, 0)
}

func TestDrawBurstMissingFrameIsSilent(t *testing.T) {
	dst := ebiten.NewImage(screenW, screenH)
	a, err := ensureAtlas()
	if err != nil {
		t.Fatal(err)
	}
	drawBursts(dst, View{
		CamX: 160, CamY: 1068,
		Bursts: []fx.Burst{{X: 320, Y: 1180, Kind: fx.BurstBoom}},
	}, a)
}

func TestDrawBoomOnSheriffCrop(t *testing.T) {
	dst := ebiten.NewImage(screenW, screenH)
	l := lot.New(2)
	d := dozer.Spawn(224, 700)
	v := View{
		CamX: 224 - 160, CamY: 680 - 112,
		Dozer:     d,
		Buildings: l.Buildings,
		Bursts:    []fx.Burst{{X: 224, Y: 680, Kind: fx.BurstBoom}},
	}
	DrawWorld(dst, v)
}

func TestLockedFrameRects(t *testing.T) {
	a, err := ensureAtlas()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]frame{
	"dozer_up_00": {X: 0, Y: 0, W: 32, H: 32},
	"dozer_up_01": {X: 32, Y: 0, W: 32, H: 32},
	"dozer_up_02": {X: 64, Y: 0, W: 32, H: 32},
	"dozer_up_03": {X: 96, Y: 0, W: 32, H: 32},
	"dozer_up_04": {X: 128, Y: 0, W: 32, H: 32},
	"dozer_up_05": {X: 160, Y: 0, W: 32, H: 32},
	"dozer_up_06": {X: 192, Y: 0, W: 32, H: 32},
	"dozer_up_07": {X: 224, Y: 0, W: 32, H: 32},
	"dozer_up_08": {X: 256, Y: 0, W: 32, H: 32},
	"dozer_up_09": {X: 288, Y: 0, W: 32, H: 32},
	"dozer_up_10": {X: 320, Y: 0, W: 32, H: 32},
	"dozer_up_11": {X: 352, Y: 0, W: 32, H: 32},
	"dozer_up_12": {X: 384, Y: 0, W: 32, H: 32},
	"dozer_up_13": {X: 416, Y: 0, W: 32, H: 32},
	"dozer_up_14": {X: 448, Y: 0, W: 32, H: 32},
	"dozer_up_15": {X: 480, Y: 0, W: 32, H: 32},
	"dozer_down_00": {X: 0, Y: 32, W: 32, H: 32},
	"dozer_down_01": {X: 32, Y: 32, W: 32, H: 32},
	"dozer_down_02": {X: 64, Y: 32, W: 32, H: 32},
	"dozer_down_03": {X: 96, Y: 32, W: 32, H: 32},
	"dozer_down_04": {X: 128, Y: 32, W: 32, H: 32},
	"dozer_down_05": {X: 160, Y: 32, W: 32, H: 32},
	"dozer_down_06": {X: 192, Y: 32, W: 32, H: 32},
	"dozer_down_07": {X: 224, Y: 32, W: 32, H: 32},
	"dozer_down_08": {X: 256, Y: 32, W: 32, H: 32},
	"dozer_down_09": {X: 288, Y: 32, W: 32, H: 32},
	"dozer_down_10": {X: 320, Y: 32, W: 32, H: 32},
	"dozer_down_11": {X: 352, Y: 32, W: 32, H: 32},
	"dozer_down_12": {X: 384, Y: 32, W: 32, H: 32},
	"dozer_down_13": {X: 416, Y: 32, W: 32, H: 32},
	"dozer_down_14": {X: 448, Y: 32, W: 32, H: 32},
	"dozer_down_15": {X: 480, Y: 32, W: 32, H: 32},
	"dozer_plate_paint": {X: 0, Y: 64, W: 32, H: 32},
	"dozer_plate_paint_down": {X: 128, Y: 64, W: 32, H: 32},
	"dozer_plate_primer": {X: 32, Y: 64, W: 32, H: 32},
	"dozer_plate_primer_down": {X: 160, Y: 64, W: 32, H: 32},
	"dozer_plate_rust": {X: 64, Y: 64, W: 32, H: 32},
	"dozer_plate_rust_down": {X: 192, Y: 64, W: 32, H: 32},
	"dozer_plate_frame": {X: 96, Y: 64, W: 32, H: 32},
	"dozer_plate_frame_down": {X: 224, Y: 64, W: 32, H: 32},
	"dozer_heat": {X: 256, Y: 64, W: 32, H: 32},
	"cruiser_00": {X: 0, Y: 96, W: 16, H: 16},
	"cruiser_01": {X: 16, Y: 96, W: 16, H: 16},
	"cruiser_02": {X: 32, Y: 96, W: 16, H: 16},
	"cruiser_03": {X: 48, Y: 96, W: 16, H: 16},
	"cruiser_04": {X: 64, Y: 96, W: 16, H: 16},
	"cruiser_05": {X: 80, Y: 96, W: 16, H: 16},
	"cruiser_06": {X: 96, Y: 96, W: 16, H: 16},
	"cruiser_07": {X: 112, Y: 96, W: 16, H: 16},
	"cruiser_08": {X: 128, Y: 96, W: 16, H: 16},
	"cruiser_09": {X: 144, Y: 96, W: 16, H: 16},
	"cruiser_10": {X: 160, Y: 96, W: 16, H: 16},
	"cruiser_11": {X: 176, Y: 96, W: 16, H: 16},
	"cruiser_12": {X: 192, Y: 96, W: 16, H: 16},
	"cruiser_13": {X: 208, Y: 96, W: 16, H: 16},
	"cruiser_14": {X: 224, Y: 96, W: 16, H: 16},
	"cruiser_15": {X: 240, Y: 96, W: 16, H: 16},
	"dump_00": {X: 24, Y: 112, W: 24, H: 24},
	"dump_04": {X: 72, Y: 112, W: 24, H: 24},
	"dump_08": {X: 120, Y: 112, W: 24, H: 24},
	"dump_12": {X: 168, Y: 112, W: 24, H: 24},
	"dump_02": {X: 48, Y: 112, W: 24, H: 24},
	"dump_06": {X: 96, Y: 112, W: 24, H: 24},
	"dump_10": {X: 144, Y: 112, W: 24, H: 24},
	"dump_14": {X: 192, Y: 112, W: 24, H: 24},
	"jersey": {X: 0, Y: 136, W: 16, H: 16},
	"jersey_broken": {X: 16, Y: 136, W: 16, H: 16},
	"excavator_00": {X: 0, Y: 152, W: 32, H: 32},
	"excavator_02": {X: 32, Y: 152, W: 32, H: 32},
	"excavator_04": {X: 64, Y: 152, W: 32, H: 32},
	"excavator_06": {X: 96, Y: 152, W: 32, H: 32},
	"excavator_08": {X: 128, Y: 152, W: 32, H: 32},
	"excavator_10": {X: 160, Y: 152, W: 32, H: 32},
	"excavator_12": {X: 192, Y: 152, W: 32, H: 32},
	"excavator_14": {X: 224, Y: 152, W: 32, H: 32},
	"chopper_0": {X: 256, Y: 96, W: 16, H: 16},
	"chopper_1": {X: 272, Y: 96, W: 16, H: 16},
	"chopper_2": {X: 288, Y: 96, W: 16, H: 16},
	"chopper_3": {X: 304, Y: 96, W: 16, H: 16},
	"spotlight": {X: 320, Y: 88, W: 48, H: 48},
	"ped": {X: 368, Y: 96, W: 8, H: 8},
	"building_intact": {X: 0, Y: 184, W: 16, H: 16},
	"building_cracked": {X: 16, Y: 184, W: 16, H: 16},
	"building_rubble": {X: 32, Y: 184, W: 16, H: 16},
	"pip_amber": {X: 48, Y: 184, W: 4, H: 4},
	"pip_rust": {X: 56, Y: 184, W: 4, H: 4},
	"pip_cyan": {X: 64, Y: 184, W: 4, H: 4},
	"dollar": {X: 72, Y: 184, W: 8, H: 8},
	"boom_00": {X: 0, Y: 208, W: 48, H: 48},
	"boom_01": {X: 48, Y: 208, W: 48, H: 48},
	"boom_02": {X: 96, Y: 208, W: 48, H: 48},
	"boom_03": {X: 144, Y: 208, W: 48, H: 48},
	"boom_04": {X: 192, Y: 208, W: 48, H: 48},
	"boom_05": {X: 240, Y: 208, W: 48, H: 48},
	"spark_00": {X: 288, Y: 208, W: 16, H: 16},
	"spark_01": {X: 304, Y: 208, W: 16, H: 16},
	"spark_02": {X: 320, Y: 208, W: 16, H: 16},
	"spark_03": {X: 336, Y: 208, W: 16, H: 16},
	"fire_00": {X: 0, Y: 256, W: 32, H: 16},
	"fire_14": {X: 224, Y: 256, W: 32, H: 16},
	"wagon_00": {X: 0, Y: 272, W: 48, H: 24},
	"wagon_14": {X: 336, Y: 272, W: 48, H: 24},
	}
	for name, w := range want {
		got, ok := a.frames[name]
		if !ok {
			t.Fatalf("missing locked frame %s", name)
		}
		if got.X != w.X || got.Y != w.Y || got.W != w.W || got.H != w.H {
			t.Fatalf("%s moved: got %d,%d %dx%d want %d,%d %dx%d", name, got.X, got.Y, got.W, got.H, w.X, w.Y, w.W, w.H)
		}
	}
	if got, ok := a.frames["dozer_up_00"]; !ok || got.X != 0 || got.Y != 0 || got.W != 32 || got.H != 32 {
		t.Fatalf("dozer_up_00 must stay at 0,0 32x32: %+v", got)
	}
}
