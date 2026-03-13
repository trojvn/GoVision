package device

import (
	"fmt"
	"image"
	"testing"
	"time"

	"gocv.io/x/gocv"
)

// MockDevice для тестирования
type MockDevice struct {
	TappedPoints   []image.Point
	WindowSizePt   image.Point
	ScreenshotData []byte
}

func (m *MockDevice) Screenshot() ([]byte, error) {
	return m.ScreenshotData, nil
}

func (m *MockDevice) Tap(x, y int) error {
	m.TappedPoints = append(m.TappedPoints, image.Pt(x, y))
	return nil
}

func (m *MockDevice) Swipe(x1, y1, x2, y2 int, duration float64) error {
	return nil
}

func (m *MockDevice) WindowSize() (image.Point, error) {
	return m.WindowSizePt, nil
}

// MockTemplate для тестирования
type MockTemplate struct {
	FoundPoint image.Point
	ShouldFind bool
}

func (m *MockTemplate) MatchIn(screen gocv.Mat) (image.Point, error) {
	if m.ShouldFind {
		return m.FoundPoint, nil
	}
	return image.Point{}, fmt.Errorf("not found")
}

func (m *MockTemplate) GetFilename() string {
	return "mock_template"
}

func TestEngine_PTap(t *testing.T) {
	// Arrange
	mock := &MockDevice{
		WindowSizePt: image.Pt(375, 667), // iPhone-like
	}
	e := NewEngine(mock)

	// Имитируем скриншот 750x1334 (Retina)
	// TouchFactor должен стать 0.5
	e.ScreenSize = image.Pt(750, 1334)
	e.TouchFactor = 0.5

	// Act
	target := image.Pt(100, 200)
	err := e.PTap(target)

	// Assert
	if err != nil {
		t.Fatalf("PTap failed: %v", err)
	}
	if len(mock.TappedPoints) != 1 {
		t.Fatal("Expected 1 tap")
	}
	expected := image.Pt(50, 100) // 100 * 0.5, 200 * 0.5
	if mock.TappedPoints[0] != expected {
		t.Errorf("Expected tap at %v, got %v", expected, mock.TappedPoints[0])
	}
}

func TestEngine_UpdateDisplayInfo(t *testing.T) {
	// Создаем пустой черный скриншот 100x200
	img := gocv.NewMatWithSize(200, 100, gocv.MatTypeCV8UC3)
	defer img.Close()
	buf, _ := gocv.IMEncode(".png", img)

	mock := &MockDevice{
		WindowSizePt:   image.Pt(50, 100),
		ScreenshotData: buf.GetBytes(),
	}
	e := NewEngine(mock)

	// Act
	err := e.UpdateDisplayInfo()

	// Assert
	if err != nil {
		t.Fatalf("UpdateDisplayInfo failed: %v", err)
	}
	if e.ScreenSize != image.Pt(100, 200) {
		t.Errorf("Expected ScreenSize (100,200), got %v", e.ScreenSize)
	}
	if e.TouchFactor != 0.5 {
		t.Errorf("Expected TouchFactor 0.5, got %f", e.TouchFactor)
	}
}

func TestMultiEngine_ExistsMulti(t *testing.T) {
	mock := &MockDevice{}
	me := NewMultiEngine(mock)

	// Создаем фейковый скриншот для GetSnapshot
	img := gocv.NewMatWithSize(10, 10, gocv.MatTypeCV8UC3)
	defer img.Close()
	buf, _ := gocv.IMEncode(".png", img)
	mock.ScreenshotData = buf.GetBytes()

	t1 := &MockTemplate{ShouldFind: false}
	t2 := &MockTemplate{ShouldFind: true, FoundPoint: image.Pt(5, 5)}

	// Act
	pos, err := me.ExistsMulti([]TemplateInterface{t1, t2}, 1*time.Second)

	// Assert
	if err != nil {
		t.Fatalf("ExistsMulti failed: %v", err)
	}
	if pos == nil || *pos != image.Pt(5, 5) {
		t.Errorf("Expected point (5,5), got %v", pos)
	}
	if me.LastMatched != t2 {
		t.Error("Expected LastMatched to be t2")
	}
}
