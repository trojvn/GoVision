package device

import (
	"fmt"
	"image"
	"time"
)

// MultiEngine расширяет Engine для работы с несколькими шаблонами одновременно.
// Аналог класса RMulti из Python.
type MultiEngine struct {
	*Engine
	RaiseInProgress bool // Если true, выбрасывает ошибку при таймауте во время поиска
	LastMatched     TemplateInterface
}

func NewMultiEngine(device Device) *MultiEngine {
	return &MultiEngine{
		Engine: NewEngine(device),
	}
}

// LoopFindMulti ищет один из переданных шаблонов в цикле до истечения таймаута.
func (me *MultiEngine) LoopFindMulti(templates []TemplateInterface, timeout time.Duration, interval time.Duration) (image.Point, error) {
	startTime := time.Now()
	for {
		screen, err := me.GetSnapshot()
		if err == nil && !screen.Empty() {
			for _, tpl := range templates {
				// Пробуем найти каждый шаблон (2 попытки как в оригинале)
				for range 2 {
					pos, err := tpl.MatchIn(screen)
					if err == nil {
						me.LastMatched = tpl
						screen.Close()
						return pos, nil
					}
				}

				// Если RaiseInProgress активен, проверяем таймаут внутри цикла по шаблонам
				if me.RaiseInProgress && time.Since(startTime) > timeout {
					screen.Close()
					return image.Point{}, fmt.Errorf("multi-find failed: timeout during template iteration")
				}
			}
			screen.Close()
		}

		if time.Since(startTime) > timeout {
			return image.Point{}, fmt.Errorf("none of the %d templates found: timeout", len(templates))
		}
		time.Sleep(interval)
	}
}

// ExistsMulti проверяет существование любого из шаблонов.
func (me *MultiEngine) ExistsMulti(templates []TemplateInterface, timeout time.Duration) (*image.Point, error) {
	pos, err := me.LoopFindMulti(templates, timeout, 100*time.Millisecond)
	if err != nil {
		return nil, err
	}
	return &pos, nil
}

// TapMulti находит один из шаблонов и выполняет нажатие.
func (me *MultiEngine) TapMulti(templates []TemplateInterface, timeout time.Duration) error {
	pos, err := me.ExistsMulti(templates, timeout)
	if err != nil {
		return err
	}
	return me.PTap(*pos)
}
