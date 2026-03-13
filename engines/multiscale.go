package engines

import (
	"image"
	"sync"

	"github.com/trojvn/rcvgo/utils"
	"gocv.io/x/gocv"
)

// MultiScaleResult содержит данные о лучшем совпадении на определенном масштабе.
type MultiScaleResult struct {
	Confidence float32
	MaxLoc     image.Point
	Width      int
	Height     int
	Ratio      float64
}

// MultiScaleSearch выполняет параллельный поиск шаблона в разных масштабах.
func MultiScaleSearch(orgSrc, orgTempl gocv.Mat, ratioMin, ratioMax, step float64, threshold float32) *MultiScaleResult {
	srcGray := utils.ImgMatToGray(orgSrc)
	defer srcGray.Close()
	templGray := utils.ImgMatToGray(orgTempl)
	defer templGray.Close()

	var wg sync.WaitGroup
	results := make(chan MultiScaleResult)
	semaphore := make(chan struct{}, 8)

	mask := gocv.NewMat()
	defer mask.Close()

	for r := ratioMin; r <= ratioMax; r += step {
		wg.Add(1)
		go func(ratio float64) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			newW := int(float64(templGray.Cols()) * ratio)
			newH := int(float64(templGray.Rows()) * ratio)

			if newW < 10 || newH < 10 || newW > srcGray.Cols() || newH > srcGray.Rows() {
				return
			}

			resizedTempl := gocv.NewMat()
			defer resizedTempl.Close()
			gocv.Resize(templGray, &resizedTempl, image.Pt(newW, newH), 0, 0, gocv.InterpolationLinear)

			res := gocv.NewMat()
			defer res.Close()

			gocv.MatchTemplate(srcGray, resizedTempl, &res, gocv.TmCcoeffNormed, mask)

			_, maxVal, _, maxLoc := gocv.MinMaxLoc(res)

			if maxVal >= threshold {
				results <- MultiScaleResult{
					Confidence: maxVal,
					MaxLoc:     maxLoc,
					Width:      newW,
					Height:     newH,
					Ratio:      ratio,
				}
			}
		}(r)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var best *MultiScaleResult
	for res := range results {
		if best == nil || res.Confidence > best.Confidence {
			temp := res
			best = &temp
		}
	}

	return best
}
