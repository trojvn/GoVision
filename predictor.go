package rcvgo

import (
	"image"
	"math"
)

// Predictor реализует логику предсказания координат на основе относительного смещения.
type Predictor struct {
	Deviation int
}

func NewPredictor() *Predictor {
	return &Predictor{
		Deviation: 100,
	}
}

// CountRecordPos вычисляет относительное смещение точки от центра экрана.
func (p *Predictor) CountRecordPos(pos image.Point, resolution image.Point) (float64, float64) {
	w, _ := float64(resolution.X), float64(resolution.Y)
	deltaX := (float64(pos.X) - w*0.5) / w
	deltaY := (float64(pos.Y) - float64(resolution.Y)*0.5) / w

	return math.Round(deltaX*1000) / 1000, math.Round(deltaY*1000) / 1000
}

// GetPredictPoint вычисляет абсолютные координаты на основе относительного смещения и текущего разрешения.
func (p *Predictor) GetPredictPoint(recordPos [2]float64, screenResolution image.Point) (float64, float64) {
	deltaX, deltaY := recordPos[0], recordPos[1]
	w, h := float64(screenResolution.X), float64(screenResolution.Y)

	targetX := deltaX*w + w*0.5
	targetY := deltaY*w + h*0.5

	return targetX, targetY
}

// GetPredictArea вычисляет область поиска вокруг предсказанной точки.
func (p *Predictor) GetPredictArea(recordPos [2]float64, imageWH image.Point, imageResolution, screenResolution image.Point) [4]int {
	x, y := p.GetPredictPoint(recordPos, screenResolution)

	var predictXRadius, predictYRadius int

	if imageResolution.X > 0 {
		predictXRadius = int(float64(imageWH.X)*float64(screenResolution.X)/(2*float64(imageResolution.X))) + p.Deviation
		predictYRadius = int(float64(imageWH.Y)*float64(screenResolution.Y)/(2*float64(imageResolution.Y))) + p.Deviation
	} else {
		predictXRadius = imageWH.X/2 + p.Deviation
		predictYRadius = imageWH.Y/2 + p.Deviation
	}

	return [4]int{
		int(x) - predictXRadius,
		int(y) - predictYRadius,
		int(x) + predictXRadius,
		int(y) + predictYRadius,
	}
}
