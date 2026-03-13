package rcvgo

import (
	"image"
	"image/color"
	"sync"

	"gocv.io/x/gocv"
)

// CalRGBConfidence вычисляет уверенность по RGB каналам параллельно.
func CalRGBConfidence(imgSrc, imgSch gocv.Mat) float32 {
	src := imgSrc.Clone()
	defer src.Close()
	sch := imgSch.Clone()
	defer sch.Close()

	gocv.CvtColor(src, &src, gocv.ColorBGRToHSV)
	gocv.CvtColor(sch, &sch, gocv.ColorBGRToHSV)

	// Исправлено: использование color.RGBA вместо gocv.Scalar
	gocv.CopyMakeBorder(src, &src, 10, 10, 10, 10, gocv.BorderReplicate, color.RGBA{0, 0, 0, 0})
	if src.Cols() > 1 {
		src.SetUCharAt(0, 0, 0)
		src.SetUCharAt(0, 1, 255)
	}

	srcChannels := gocv.Split(src)
	schChannels := gocv.Split(sch)
	defer func() {
		for i := 0; i < 3; i++ {
			srcChannels[i].Close()
			schChannels[i].Close()
		}
	}()

	confidences := make([]float32, 3)
	var wg sync.WaitGroup

	mask := gocv.NewMat()
	defer mask.Close()

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			res := gocv.NewMat()
			defer res.Close()
			gocv.MatchTemplate(srcChannels[idx], schChannels[idx], &res, gocv.TmCcoeffNormed, mask)
			_, maxVal, _, _ := gocv.MinMaxLoc(res)
			confidences[idx] = maxVal
		}(i)
	}
	wg.Wait()

	minConf := confidences[0]
	for _, c := range confidences {
		if c < minConf {
			minConf = c
		}
	}
	return minConf
}

// TemplateMatching инкапсулирует логику поиска по шаблону.
type TemplateMatching struct {
	ImSource           gocv.Mat
	ImSearch           gocv.Mat
	Threshold          float32
	RGB                bool
	SmartRGB           bool
	RGBThresholdOffset float32
	MaxResultCount     int
}

func NewTemplateMatching(imSearch, imSource gocv.Mat, threshold float32) *TemplateMatching {
	return &TemplateMatching{
		ImSource:           imSource,
		ImSearch:           imSearch,
		Threshold:          threshold,
		RGB:                true,
		SmartRGB:           true,
		RGBThresholdOffset: 0.15,
		MaxResultCount:     10,
	}
}

func (tm *TemplateMatching) FindBestResult() (*MatchResult, error) {
	if err := CheckSourceLargerThanSearch(tm.ImSource, tm.ImSearch); err != nil {
		return nil, err
	}

	res := tm.getTemplateResultMatrix()
	defer res.Close()

	_, maxVal, _, maxLoc := gocv.MinMaxLoc(res)
	h, w := tm.ImSearch.Rows(), tm.ImSearch.Cols()
	confidence := tm.getConfidenceFromMatrix(maxLoc, maxVal, w, h)

	if confidence < tm.Threshold {
		return nil, nil
	}

	middlePoint, rectangle := tm.getTargetRectangle(maxLoc, w, h)
	return &MatchResult{
		Result:     middlePoint,
		Rectangle:  rectangle,
		Confidence: confidence,
	}, nil
}

func (tm *TemplateMatching) getTemplateResultMatrix() gocv.Mat {
	srcGray := ImgMatToGray(tm.ImSource)
	schGray := ImgMatToGray(tm.ImSearch)
	defer srcGray.Close()
	defer schGray.Close()

	mask := gocv.NewMat()
	defer mask.Close()

	res := gocv.NewMat()
	gocv.MatchTemplate(srcGray, schGray, &res, gocv.TmCcoeffNormed, mask)
	return res
}

func (tm *TemplateMatching) getConfidenceFromMatrix(maxLoc image.Point, maxVal float32, w, h int) float32 {
	confidence := maxVal

	if tm.SmartRGB && tm.RGB {
		thresholdWithOffset := tm.Threshold + tm.RGBThresholdOffset
		if confidence < thresholdWithOffset {
			rect := []int{maxLoc.X, maxLoc.Y, maxLoc.X + w, maxLoc.Y + h}
			imgCrop, err := CropImage(tm.ImSource, rect)
			if err == nil {
				defer imgCrop.Close()
				confidence = CalRGBConfidence(imgCrop, tm.ImSearch)
			}
		}
	}
	return confidence
}

func (tm *TemplateMatching) getTargetRectangle(leftTop image.Point, w, h int) (image.Point, []image.Point) {
	middle := image.Pt(leftTop.X+w/2, leftTop.Y+h/2)
	rect := []image.Point{
		leftTop,
		image.Pt(leftTop.X, leftTop.Y+h),
		image.Pt(leftTop.X+w, leftTop.Y+h),
		image.Pt(leftTop.X+w, leftTop.Y),
	}
	return middle, rect
}
