package engines

import (
	"image"
	"image/color"
	"testing"

	"gocv.io/x/gocv"
)

func TestMultiScaleSearch(t *testing.T) {
	// Arrange
	src := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer src.Close()
	src.SetTo(gocv.Scalar{Val1: 0, Val2: 0, Val3: 0, Val4: 0})

	// Рисуем белый квадрат 20x20 в центре
	rect := image.Rect(40, 40, 60, 60)
	white := gocv.NewMatWithSize(20, 20, gocv.MatTypeCV8UC3)
	defer white.Close()
	white.SetTo(gocv.Scalar{Val1: 255, Val2: 255, Val3: 255, Val4: 0})
	region := src.Region(rect)
	white.CopyTo(&region)
	region.Close()

	// Шаблон - тот же белый квадрат, но другого размера (10x10)
	templ := gocv.NewMatWithSize(10, 10, gocv.MatTypeCV8UC3)
	defer templ.Close()
	templ.SetTo(gocv.Scalar{Val1: 255, Val2: 255, Val3: 255, Val4: 0})
	gocv.Circle(&templ, image.Pt(5, 5), 3, color.RGBA{R: 255, G: 0, B: 0, A: 255}, -1)

	// Рисуем увеличенный шаблон на источнике
	rect = image.Rect(40, 40, 60, 60)
	region = src.Region(rect)
	gocv.Resize(templ, &region, image.Pt(20, 20), 0, 0, gocv.InterpolationLinear)
	region.Close()

	// Act
	// Ищем шаблон 10x10 на изображении, где есть квадрат 20x20.
	// При масштабе 2.0 он должен идеально совпасть.
	res := MultiScaleSearch(src, templ, 1.0, 3.0, 0.1, 0.8)

	// Assert
	if res == nil {
		t.Fatal("Expected to find result, got nil")
	}
	if res.Confidence < 0.8 {
		t.Errorf("Expected high confidence, got %f", res.Confidence)
	}
	// Проверяем, что найденный масштаб близок к 2.0
	if res.Ratio < 1.9 || res.Ratio > 2.1 {
		t.Errorf("Expected ratio around 2.0, got %f", res.Ratio)
	}
}
