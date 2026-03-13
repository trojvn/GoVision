package rcvgo

import "image"

// MatchResult представляет результат поиска шаблона на изображении.
type MatchResult struct {
	Result     image.Point   `json:"result"`     // Центральная точка найденной области
	Rectangle  []image.Point `json:"rectangle"`  // Четыре угла найденной области (LT, LB, RB, RT)
	Confidence float32       `json:"confidence"` // Уровень уверенности (0.0 - 1.0)
}

// Point - вспомогательная структура для координат (если не хватает стандартной image.Point)
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Rect - вспомогательная структура для области
type Rect struct {
	XMin int
	YMin int
	XMax int
	YMax int
}
