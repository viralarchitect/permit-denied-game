package render

import (
	"fmt"
	"math"
	"testing"

	"permitdenied/assets"
	"permitdenied/internal/dozer"
	"permitdenied/internal/fx"
	"permitdenied/internal/lot"

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
	})
	DrawTitle(dst)
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
