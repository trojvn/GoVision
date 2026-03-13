package main

import (
	"fmt"
	"image"
	"log"
	"time"

	"github.com/trojvn/rcvgo"
	"github.com/trojvn/rcvgo/device"
)

// MockDevice — пример реализации интерфейса device.Device.
// В реальном проекте здесь будет интеграция с ADB (Android) или WDA (iOS).
type MockDevice struct{}

func (m *MockDevice) Screenshot() ([]byte, error) {
	// Здесь должен быть код получения скриншота с девайса
	return nil, fmt.Errorf("not implemented in mock")
}

func (m *MockDevice) Tap(x, y int) error {
	fmt.Printf("Выполнен тап по координатам: [%d, %d]\n", x, y)
	return nil
}

func (m *MockDevice) Swipe(x1, y1, x2, y2 int, duration float64) error {
	fmt.Printf("Выполнен свайп: [%d, %d] -> [%d, %d] за %.2f сек\n", x1, y1, x2, y2, duration)
	return nil
}

func (m *MockDevice) WindowSize() (image.Point, error) {
	// Логическое разрешение экрана (например, в "точках" для iOS)
	return image.Pt(375, 667), nil
}

func ExampleDeviceUsage() {
	// 1. Инициализируем устройство
	myDevice := &MockDevice{}

	// 2. Создаем движок управления
	engine := device.NewEngine(myDevice)

	// 3. Подготавливаем шаблоны для поиска
	// Допустим, мы ищем кнопку "Принять"
	btnAccept := rcvgo.NewTemplate("accept_button")
	btnAccept.Threshold = 0.85

	// 4. Пример: Ожидание появления кнопки и нажатие на нее
	// Engine автоматически сделает скриншот, найдет шаблон и нажмет с учетом масштабирования
	fmt.Println("Ищем кнопку 'Принять'...")
	err := engine.Tap(btnAccept, 10*time.Second)
	if err != nil {
		fmt.Printf("Не удалось нажать на кнопку: %v\n", err)
	}

	// 5. Пример: Работа с несколькими шаблонами (MultiEngine)
	// Полезно, когда на экране может появиться одно из нескольких окон/кнопок
	multi := device.NewMultiEngine(myDevice)

	templates := []device.TemplateInterface{
		rcvgo.NewTemplate("error_popup"),
		rcvgo.NewTemplate("success_popup"),
	}

	fmt.Println("Ожидаем результат операции...")
	pos, err := multi.ExistsMulti(templates, 15*time.Second)
	if err != nil {
		log.Printf("Ни один шаблон не найден: %v", err)
	} else {
		fmt.Printf("Найден шаблон! Координаты на скриншоте: %v\n", pos)
	}

	// 6. Пример: Прямой тап по координатам (с автоматическим масштабированием)
	// Если TouchFactor еще не рассчитан, UpdateDisplayInfo сделает это автоматически
	err = engine.PTap(image.Pt(500, 1000)) // Координаты в разрешении скриншота
	if err != nil {
		log.Fatal(err)
	}
}
