package audio

import "testing"

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
	a.Stop()
	for i, p := range a.shots {
		if p != nil {
			t.Fatalf("shot %d still live after Stop", i)
		}
	}
}
