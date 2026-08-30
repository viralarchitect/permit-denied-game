package audio

import (
	"bytes"
	"testing"
)

func TestLoadWreckWAV(t *testing.T) {
	b := loadSample("wreck.wav")
	if len(b) == 0 {
		t.Fatal("wreck.wav missing or decode failed")
	}
}

func TestLoadPeelAndBurstWAV(t *testing.T) {
	if len(loadSample("peel.wav")) == 0 {
		t.Fatal("peel.wav")
	}
	if len(loadSample("burst.wav")) == 0 {
		t.Fatal("burst.wav")
	}
}

func TestMissingWAVFailsOpen(t *testing.T) {
	if loadSample("no-such.wav") != nil {
		t.Fatal("missing file should return nil")
	}
}

func TestStopClearsOneShots(t *testing.T) {
	a := &Audio{}
	a.Wreck()
	a.DuckWreck(true)
	a.Stop()
	for i, p := range a.shots {
		if p != nil {
			t.Fatalf("shot %d still live after Stop", i)
		}
	}
	if a.chip != nil && a.chip.IsPlaying() {
		t.Fatal("chip still playing after Stop")
	}
}

func TestCrunchNotWreckSample(t *testing.T) {
	a := &Audio{}
	a.ensure()
	if len(a.crunchB) == 0 || len(a.wreckB) == 0 {
		t.Fatal("missing crunch or wreck PCM")
	}
	if bytes.Equal(a.crunchB, a.wreckB) {
		t.Fatal("Crunch and Wreck share the same sample")
	}
}

func TestTallyDuckWinsOverWreck(t *testing.T) {
	a := &Audio{}
	a.StartChase()
	a.DuckWreck(true)
	if v := a.ChaseVolume(); v >= 0.20 {
		t.Fatalf("wreck duck ChaseVolume=%v want < 0.20", v)
	}
	a.Duck(true)
	if v := a.ChaseVolume(); v != 0.12 {
		t.Fatalf("tally duck ChaseVolume=%v want 0.12", v)
	}
}
