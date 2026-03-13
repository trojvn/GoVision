package engines

import (
	"image"
	"image/color"
	"testing"

	"gocv.io/x/gocv"
)

func TestTemplateMatching(t *testing.T) {
	// Arrange
	src := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer src.Close()
	src.SetTo(gocv.Scalar{Val1: 0, Val2: 0, Val3: 0, Val4: 0})

	rect := image.Rect(10, 10, 30, 30)
	templ := gocv.NewMatWithSize(20, 20, gocv.MatTypeCV8UC3)
	defer templ.Close()
	// Рисуем что-то более сложное, чем просто сплошной цвет
	templ.SetTo(gocv.Scalar{Val1: 255, Val2: 255, Val3: 255, Val4: 0})
	gocv.Circle(&templ, image.Pt(10, 10), 5, color.RGBA{R: 255, G: 0, B: 0, A: 255}, -1)

	region := src.Region(rect)
	templ.CopyTo(&region)
	region.Close()

	tm := NewTemplateMatching(templ, src, 0.8)
	tm.SmartRGB = false // Отключаем для базового теста

	// Act
	res, err := tm.FindBestResult()

	// Assert
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if res == nil {
		t.Fatal("Expected result, got nil")
	}
	if res.Confidence < 0.99 {
		t.Errorf("Expected high confidence, got %f", res.Confidence)
	}
	if res.Result != image.Pt(20, 20) {
		t.Errorf("Expected center (20,20), got %v", res.Result)
	}
}

func TestCalRGBConfidence(t *testing.T) {
	// Arrange
	img1 := gocv.NewMatWithSize(20, 20, gocv.MatTypeCV8UC3)
	defer img1.Close()
	img1.SetTo(gocv.Scalar{Val1: 255, Val2: 0, Val3: 0, Val4: 0}) // Blue in BGR
	gocv.Circle(&img1, image.Pt(10, 10), 5, color.RGBA{R: 255, G: 0, B: 0, A: 255}, -1)

	img2 := img1.Clone()
	defer img2.Close()

	img3 := gocv.NewMatWithSize(20, 20, gocv.MatTypeCV8UC3)
	defer img3.Close()
	img3.SetTo(gocv.Scalar{Val1: 0, Val2: 0, Val3: 255, Val4: 0}) // Red in BGR
	gocv.Circle(&img3, image.Pt(10, 10), 5, color.RGBA{R: 0, G: 255, B: 0, A: 255}, -1)

	// Act
	confSame := CalRGBConfidence(img1, img2)
	confDiff := CalRGBConfidence(img1, img3)

	// Assert
	if confSame < 0.99 {
		t.Errorf("Expected high confidence for same colors, got %f", confSame)
	}
	if confDiff > 0.5 {
		t.Errorf("Expected low confidence for different colors, got %f", confDiff)
	}
}
