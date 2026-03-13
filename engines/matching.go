package engines

import (
	"image"
	"sync"

	"github.com/trojvn/rcvgo/core"
	"github.com/trojvn/rcvgo/utils"
	"gocv.io/x/gocv"
)

// CalRGBConfidence вычисляет уверенность по RGB каналам параллельно.
func CalRGBConfidence(imgSrc, imgSch gocv.Mat) float32 {
	srcChannels := gocv.Split(imgSrc)
	schChannels := gocv.Split(imgSch)
	defer func() {
		for i := 0; i < len(srcChannels); i++ {
			srcChannels[i].Close()
			schChannels[i].Close()
		}
	}()

	confidences := make([]float32, len(srcChannels))
	var wg sync.WaitGroup
	mask := gocv.NewMat()
	defer mask.Close()

	for i := 0; i < len(srcChannels); i++ {
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

func (tm *TemplateMatching) FindBestResult() (*core.MatchResult, error) {
	if err := utils.CheckSourceLargerThanSearch(tm.ImSource, tm.ImSearch); err != nil {
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

	return utils.NewMatchResult(maxLoc, w, h, confidence), nil
}

func (tm *TemplateMatching) getTemplateResultMatrix() gocv.Mat {
	srcGray := utils.ImgMatToGray(tm.ImSource)
	schGray := utils.ImgMatToGray(tm.ImSearch)
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
			rect := image.Rect(maxLoc.X, maxLoc.Y, maxLoc.X+w, maxLoc.Y+h)
			imgCrop, err := utils.CropImage(tm.ImSource, rect)
			if err == nil {

				defer imgCrop.Close()
				confidence = CalRGBConfidence(imgCrop, tm.ImSearch)
			}
		}
	}
	return confidence
}
