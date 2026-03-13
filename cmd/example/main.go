package main

import (
	"fmt"
	"log"

	"github.com/trojvn/rcvgo"
	"gocv.io/x/gocv"
)

func main() {
	// 1. Загрузка скриншота экрана
	screen := gocv.IMRead("screen.png", gocv.IMReadColor)
	if screen.Empty() {
		log.Fatal("Could not read screen.png")
	}
	defer screen.Close()

	// 2. Инициализация шаблона
	// По умолчанию ищет в папке "images/" и добавляет расширение ".png"
	tpl := rcvgo.NewTemplate("button")
	tpl.Threshold = 0.8
	tpl.RGB = true
	tpl.OptimizationLevel = rcvgo.OptimizationBalanced

	// 3. Поиск шаблона на экране
	point, err := tpl.MatchIn(screen)
	if err != nil {
		fmt.Printf("Template not found or error: %v\n", err)
	} else {
		fmt.Printf("Found template at point: %v\n", point)
	}

	// 4. Если нужен детальный результат (уверенность, прямоугольник)
	result, err := tpl.CVMatch(screen)
	if err == nil && result != nil {
		fmt.Printf("Confidence: %.2f\n", result.Confidence)
		fmt.Printf("Rectangle: %v\n", result.Rectangle)
	}
}
