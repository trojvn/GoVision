package utils

import (
	"image"
	"testing"

	"gocv.io/x/gocv"
)

func TestNewMatchResult(t *testing.T) {
	// Arrange
	maxLoc := image.Pt(10, 20)
	w, h := 100, 50
	confidence := float32(0.95)

	// Act
	res := NewMatchResult(maxLoc, w, h, confidence)

	// Assert
	if res.Confidence != confidence {
		t.Errorf("Expected confidence %f, got %f", confidence, res.Confidence)
	}
	expectedCenter := image.Pt(60, 45) // 10 + 100/2, 20 + 50/2
	if res.Result != expectedCenter {
		t.Errorf("Expected center %v, got %v", expectedCenter, res.Result)
	}
	if len(res.Rectangle) != 4 {
		t.Errorf("Expected 4 points in rectangle, got %d", len(res.Rectangle))
	}
	if res.Rectangle[0] != maxLoc {
		t.Errorf("Expected LT %v, got %v", maxLoc, res.Rectangle[0])
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		val, min, max, expected int
	}{
		{5, 0, 10, 5},
		{-1, 0, 10, 0},
		{11, 0, 10, 10},
		{0, 0, 10, 0},
		{10, 0, 10, 10},
	}

	for _, tt := range tests {
		result := clamp(tt.val, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("clamp(%d, %d, %d) = %d; want %d", tt.val, tt.min, tt.max, result, tt.expected)
		}
	}
}

func TestImgMatToGray(t *testing.T) {
	// Arrange
	img := gocv.NewMatWithSize(10, 10, gocv.MatTypeCV8UC3)
	defer img.Close()

	// Act
	gray := ImgMatToGray(img)
	defer gray.Close()

	// Assert
	if gray.Channels() != 1 {
		t.Errorf("Expected 1 channel for grayscale image, got %d", gray.Channels())
	}
}

func TestCheckSourceLargerThanSearch(t *testing.T) {
	// Arrange
	src := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC1)
	defer src.Close()
	small := gocv.NewMatWithSize(50, 50, gocv.MatTypeCV8UC1)
	defer small.Close()
	big := gocv.NewMatWithSize(150, 150, gocv.MatTypeCV8UC1)
	defer big.Close()

	// Act & Assert
	if err := CheckSourceLargerThanSearch(src, small); err != nil {
		t.Errorf("Expected no error for smaller search image, got %v", err)
	}
	if err := CheckSourceLargerThanSearch(src, big); err == nil {
		t.Error("Expected error for larger search image, got nil")
	}
}
