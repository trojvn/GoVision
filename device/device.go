package device

import (
	"image"
)

// Device - интерфейс для взаимодействия с устройством (iOS/Android).
// Позволяет модулю оставаться независимым от конкретной реализации (WDA, ADB и т.д.).
type Device interface {
	Screenshot() ([]byte, error)
	Tap(x, y int) error
	Swipe(x1, y1, x2, y2 int, duration float64) error
	WindowSize() (image.Point, error)
}
