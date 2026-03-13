package device

import (
	"fmt"
	"image"
	"time"

	"gocv.io/x/gocv"
)

// TemplateInterface - интерфейс для шаблона, чтобы избежать циклической зависимости.
type TemplateInterface interface {
	MatchIn(screen gocv.Mat) (image.Point, error)
	GetFilename() string
}

// Engine - аналог класса R из Python. Управляет процессом распознавания на устройстве.
type Engine struct {
	Device      Device
	TouchFactor float64
	WindowSize  image.Point
	ScreenSize  image.Point
}

func NewEngine(device Device) *Engine {
	return &Engine{
		Device: device,
	}
}

// GetSnapshot получает скриншот с устройства и декодирует его в gocv.Mat.
func (e *Engine) GetSnapshot() (gocv.Mat, error) {
	data, err := e.Device.Screenshot()
	if err != nil {
		return gocv.NewMat(), err
	}
	img, err := gocv.IMDecode(data, gocv.IMReadColor)
	if err != nil {
		return gocv.NewMat(), err
	}
	return img, nil
}

// UpdateDisplayInfo обновляет информацию о разрешении экрана и коэффициенте касания.
func (e *Engine) UpdateDisplayInfo() error {
	winSize, err := e.Device.WindowSize()
	if err != nil {
		return err
	}
	e.WindowSize = winSize

	snapshot, err := e.GetSnapshot()
	if err != nil {
		return err
	}
	defer snapshot.Close()

	e.ScreenSize = image.Pt(snapshot.Cols(), snapshot.Rows())

	if e.ScreenSize.Y > 0 {
		e.TouchFactor = float64(e.WindowSize.Y) / float64(e.ScreenSize.Y)
	}
	return nil
}

// LoopFind ищет шаблон в цикле до истечения таймаута.
func (e *Engine) LoopFind(tpl TemplateInterface, timeout time.Duration, interval time.Duration) (image.Point, error) {
	startTime := time.Now()
	consecutiveErrors := 0
	maxConsecutiveErrors := 5

	for {
		screen, err := e.GetSnapshot()
		if err != nil {
			consecutiveErrors++
			if consecutiveErrors >= maxConsecutiveErrors {
				return image.Point{}, fmt.Errorf("failed to get snapshot after %d attempts: %v", maxConsecutiveErrors, err)
			}
		} else if !screen.Empty() {
			consecutiveErrors = 0
			// Пробуем найти 2 раза (как в оригинале)
			for i := 0; i < 2; i++ {
				pos, err := tpl.MatchIn(screen)
				if err == nil {
					screen.Close()
					return pos, nil
				}
			}
			screen.Close()
		}

		if time.Since(startTime) > timeout {
			return image.Point{}, fmt.Errorf("target %s not found: timeout", tpl.GetFilename())
		}
		time.Sleep(interval)
	}
}

// Exists проверяет существование шаблона на экране.
func (e *Engine) Exists(tpl TemplateInterface, timeout time.Duration) (*image.Point, error) {
	pos, err := e.LoopFind(tpl, timeout, 100*time.Millisecond)
	if err != nil {
		return nil, err
	}
	return &pos, nil
}

// Tap находит шаблон и выполняет нажатие.
func (e *Engine) Tap(tpl TemplateInterface, timeout time.Duration) error {
	pos, err := e.Exists(tpl, timeout)
	if err != nil {
		return err
	}
	return e.PTap(*pos)
}

// PTap выполняет нажатие по координатам с учетом TouchFactor.
func (e *Engine) PTap(pos image.Point) error {
	if e.TouchFactor == 0 {
		if err := e.UpdateDisplayInfo(); err != nil {
			return err
		}
	}

	// Масштабируем координаты из разрешения скриншота в логическое разрешение WDA
	x := int(float64(pos.X) * e.TouchFactor)
	y := int(float64(pos.Y) * e.TouchFactor)

	return e.Device.Tap(x, y)
}

// Swipe выполняет свайп между точками.
func (e *Engine) Swipe(p1, p2 image.Point, duration float64) error {
	if e.TouchFactor == 0 {
		if err := e.UpdateDisplayInfo(); err != nil {
			return err
		}
	}

	x1 := int(float64(p1.X) * e.TouchFactor)
	y1 := int(float64(p1.Y) * e.TouchFactor)
	x2 := int(float64(p2.X) * e.TouchFactor)
	y2 := int(float64(p2.Y) * e.TouchFactor)

	return e.Device.Swipe(x1, y1, x2, y2, duration)
}
