package render

import "testing"

func TestPlateColorMetaArmorStaysPaint(t *testing.T) {
	// Armor I adds a fifth plate; that must still read as paint yellow, not bare frame.
	if PlateColor(5) != ColPaint {
		t.Fatalf("PlateColor(5)=%v want ColPaint %v", PlateColor(5), ColPaint)
	}
	if PlateColor(4) != ColPaint {
		t.Fatalf("PlateColor(4)=%v want ColPaint", PlateColor(4))
	}
	if PlateColor(1) != ColFrame {
		t.Fatalf("PlateColor(1)=%v want ColFrame", PlateColor(1))
	}
}
