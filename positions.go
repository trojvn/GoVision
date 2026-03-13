package rcvgo

import (
	"image"
)

// TargetPos определяет, какую точку возвращать из найденной области.
type TargetPos int

const (
	MID TargetPos = iota // Центр
	LT                   // Левый верхний угол
	LB                   // Левый нижний угол
	RB                   // Правый нижний угол
	RT                   // Правый верхний угол
)

// GetXY возвращает координаты точки в зависимости от выбранной позиции.
func (tp TargetPos) GetXY(res MatchResult) image.Point {
	switch tp {
	case LT:
		return res.Rectangle[0]
	case LB:
		return res.Rectangle[1]
	case RB:
		return res.Rectangle[2]
	case RT:
		return res.Rectangle[3]
	case MID:
		fallthrough
	default:
		return res.Result
	}
}
