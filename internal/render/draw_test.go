package render

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func TestDrawTallyLinesFitOnScreen(t *testing.T) {
	_, h := text.Measure("TARGETS    2 \u00d71.6", hudFace, 0)
	if h <= 0 {
		t.Fatal("font height is 0")
	}

	targetsBottom := tallyStatY(3) + h
	if targetsBottom > screenH {
		t.Fatalf("TARGETS clipped: y=%.0f..%.0f screenH=%d", tallyStatY(3), targetsBottom, screenH)
	}

	totalBottom := tallyStatY(4) + h
	if totalBottom > screenH {
		t.Fatalf("TOTAL clipped: y=%.0f..%.0f screenH=%d", tallyStatY(4), totalBottom, screenH)
	}
}

func TestTallyCopy(t *testing.T) {
	cases := map[string]string{
		"cooked": "ENGINE COOKED",
		"track":  "TRACK THROWN",
		"buzzer": "COUNTY CLOCK",
	}
	for death, want := range cases {
		if got := deathLine(death); got != want {
			t.Fatalf("death %q: got %q want %q", death, got, want)
		}
	}
	if againPrompt() != "SPACE / TAP — AGAIN" {
		t.Fatalf("again=%q", againPrompt())
	}
}
