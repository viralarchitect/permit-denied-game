package fx

import "testing"

func TestBoomAgesOut(t *testing.T) {
	var f FX
	f.SpawnBoom(10, 20)
	if len(f.Bursts) != 1 || f.Bursts[0].FrameName() != "boom_00" {
		t.Fatalf("spawn: %+v", f.Bursts)
	}
	for i := 0; i < boomLife; i++ {
		f.Step(1.0/60, 8)
	}
	if len(f.Bursts) != 0 {
		t.Fatalf("still live after %d ticks: %+v", boomLife, f.Bursts)
	}
}

func TestSparkFrameAndLife(t *testing.T) {
	var f FX
	f.SpawnSpark(1, 2)
	if f.Bursts[0].FrameName() != "spark_00" {
		t.Fatalf("frame %s", f.Bursts[0].FrameName())
	}
	for i := 0; i < sparkLife; i++ {
		f.Step(1.0/60, 8)
	}
	if len(f.Bursts) != 0 {
		t.Fatalf("spark still live")
	}
}

func TestBoom05IsLastLiveFrame(t *testing.T) {
	b := Burst{Kind: BurstBoom, Age: boomLife - 1}
	if b.Dead() {
		t.Fatal("last tick should still draw")
	}
	if b.FrameName() != "boom_05" {
		t.Fatalf("got %s", b.FrameName())
	}
}
