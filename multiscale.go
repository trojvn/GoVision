package rcvgo

import (
	"image"
	"sync"

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
	var wg sync.WaitGroup
	results := make(chan MultiScaleResult)

	// Ограничиваем количество одновременно работающих горутин для стабильности
	semaphore := make(chan struct{}, 8)

	for r := ratioMin; r <= ratioMax; r += step {
		wg.Add(1)
		go func(ratio float64) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			src := orgSrc.Clone()
			defer src.Close()
			templ := orgTempl.Clone()
			defer templ.Close()

			newW := int(float64(templ.Cols()) * ratio)
			newH := int(float64(templ.Rows()) * ratio)

			if newW < 10 || newH < 10 || newW > src.Cols() || newH > src.Rows() {
				return
			}

			gocv.Resize(templ, &templ, image.Pt(newW, newH), 0, 0, gocv.InterpolationLinear)

			res := gocv.NewMat()
			defer res.Close()

			mask := gocv.NewMat()
			defer mask.Close()

			// Используем Grayscale для скорости в многомасштабном поиске
			srcGray := ImgMatToGray(src)
			defer srcGray.Close()
			schGray := ImgMatToGray(templ)
			defer schGray.Close()

			gocv.MatchTemplate(srcGray, schGray, &res, gocv.TmCcoeffNormed, mask)

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
