package predictor

import (
	"image"
	"testing"
)

func TestPredictor(t *testing.T) {
	p := NewPredictor()
	p.Deviation = 10

	res := image.Pt(1000, 1000)
	pos := image.Pt(600, 400)

	// Act
	dx, dy := p.CountRecordPos(pos, res)

	// Assert
	// (600 - 500) / 1000 = 0.1
	// (400 - 500) / 1000 = -0.1
	if dx != 0.1 || dy != -0.1 {
		t.Errorf("Expected (0.1, -0.1), got (%f, %f)", dx, dy)
	}

	// Act: GetPredictPoint
	pt := p.GetPredictPoint([2]float64{0.1, -0.1}, res)
	if pt.X != 600 || pt.Y != 400 {
		t.Errorf("Expected (600, 400), got %v", pt)
	}

	// Act: GetPredictArea
	// imageWH = (20, 20), imageRes = (1000, 1000), screenRes = (1000, 1000)
	// radius = 20 * 1000 / (2 * 1000) + 10 = 10 + 10 = 20
	area := p.GetPredictArea([2]float64{0.1, -0.1}, image.Pt(20, 20), res, res)
	expectedArea := image.Rect(580, 380, 620, 420)
	if area != expectedArea {
		t.Errorf("Expected area %v, got %v", expectedArea, area)
	}
}
