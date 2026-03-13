# RCvGo

RCvGo — это мощная и производительная библиотека на Go для распознавания образов и поиска шаблонов на изображениях, вдохновленная популярными инструментами компьютерного зрения. Она предоставляет удобный интерфейс для работы с OpenCV (`gocv`), объединяя классические и современные алгоритмы поиска.

## Особенности

- **Template Matching**: Быстрый поиск точных совпадений.
- **Multi-Scale Search**: Поиск шаблонов разных размеров с автоматическим подбором масштаба.
- **SIFT & KAZE**: Поиск по ключевым точкам, устойчивый к поворотам, изменению освещения и частичным перекрытиям.
- **Smart RGB**: Двухуровневая проверка (Grayscale + RGB) для высокой точности при сохранении скорости.
- **Predictor**: Логика предсказания области поиска на основе относительных координат.
- **Оптимизация**: Параллельные вычисления и эффективное управление памятью.

## Установка

Для работы библиотеки требуется установленный [OpenCV](https://opencv.org/) и [GoCV](https://gocv.io/).

```bash
go get github.com/trojvn/rcvgo
```

## Быстрый старт

```go
package main

import (
	"fmt"
	"log"
	"github.com/trojvn/rcvgo"
	"gocv.io/x/gocv"
)

func main() {
	// Загружаем скриншот
	screen := gocv.IMRead("screen.png", gocv.IMReadColor)
	defer screen.Close()

	// Инициализируем шаблон (ищет в папке "images/button.png")
	tpl := rcvgo.NewTemplate("button")
	tpl.Threshold = 0.8
	tpl.RGB = true

	// Ищем шаблон
	point, err := tpl.MatchIn(screen)
	if err != nil {
		log.Fatalf("Ошибка поиска: %v", err)
	}

	fmt.Printf("Шаблон найден в точке: %v\n", point)
}
```

## Продвинутое использование

### Работа с мобильными устройствами (iOS/Android)

Модуль `device` позволяет автоматизировать действия на реальных устройствах. Он абстрагирован от конкретной реализации (ADB, WDA), что позволяет легко интегрировать его в любой стек.

```go
import (
    "github.com/trojvn/rcvgo"
    "github.com/trojvn/rcvgo/device"
)

func main() {
    // 1. Инициализируйте вашу реализацию устройства (ADB/WDA)
    var myDevice device.Device = ... 

    // 2. Создайте движок управления
    engine := device.NewEngine(myDevice)

    // 3. Найдите шаблон и нажмите на него (с автоматическим масштабированием)
    tpl := rcvgo.NewTemplate("login_button")
    err := engine.Tap(tpl, 10 * time.Second)
}
```

### Многомасштабный поиск (Multi-Scale)

Если размер объекта на экране может меняться, используйте `OptimizationBalanced` или `OptimizationMaximum`:

```go
tpl := rcvgo.NewTemplate("icon")
tpl.OptimizationLevel = rcvgo.OptimizationBalanced
tpl.ScaleStep = 0.05 // Шаг изменения масштаба

result, err := tpl.CVMatch(screen)
if err == nil && result != nil {
    fmt.Printf("Уверенность: %.2f, Масштаб: %v\n", result.Confidence, result.Rectangle)
}
```

### Использование Predictor

Позволяет ограничить область поиска на основе предыдущих находок или относительных координат экрана:

```go
p := rcvgo.NewPredictor()
// Относительные координаты (0.1 от центра по X, -0.1 по Y)
recordPos := [2]float64{0.1, -0.1}
screenRes := image.Pt(1920, 1080)

// Получаем область поиска
searchArea := p.GetPredictArea(recordPos, image.Pt(100, 100), screenRes, screenRes)
```

## Структура проекта

- `core/`: Базовые типы, конфигурации и константы.
- `engines/`: Реализация алгоритмов (Matching, SIFT, KAZE, MultiScale).
- `utils/`: Вспомогательные функции для обработки изображений.
- `predictor/`: Логика работы с координатами и областями поиска.
- `device/`: Взаимодействие с устройствами (iOS/Android), управление процессом распознавания.

## Тестирование

Библиотека покрыта Unit-тестами. Для запуска выполните:

```bash
go test -v ./...
```

## Лицензия

MIT
