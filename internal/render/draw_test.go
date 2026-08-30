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
