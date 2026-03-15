package engines

import (
	"errors"

	"github.com/trojvn/GoVision/core"
	"github.com/trojvn/GoVision/utils"
	"gocv.io/x/gocv"
)

// KAZEMatching реализует поиск по ключевым точкам KAZE.
type KAZEMatching struct {
	ImSource  gocv.Mat
	ImSearch  gocv.Mat
	Threshold float32
}

func NewKAZEMatching(imSearch, imSource gocv.Mat, threshold float32) *KAZEMatching {
	return &KAZEMatching{
		ImSource:  imSource,
		ImSearch:  imSearch,
		Threshold: threshold,
	}
}

func (k *KAZEMatching) FindBestResult() (*core.MatchResult, error) {
	if k.ImSource.Empty() || k.ImSearch.Empty() {
		return nil, errors.New("empty images")
	}

	// В GoCV KAZE доступен через основной пакет
	detector := gocv.NewKAZE()
	defer detector.Close()

	mask := gocv.NewMat()
	defer mask.Close()

	kpSch, desSch := detector.DetectAndCompute(k.ImSearch, mask)
	defer desSch.Close()
	kpSrc, desSrc := detector.DetectAndCompute(k.ImSource, mask)
	defer desSrc.Close()

	if len(kpSch) < 4 || len(kpSrc) < 4 {
		return nil, errors.New("not enough keypoints")
	}

	matcher := gocv.NewBFMatcherWithParams(gocv.NormL2, false)
	defer matcher.Close()

	matches := matcher.KnnMatch(desSch, desSrc, 2)

	var goodMatches []gocv.DMatch
	var srcPts []gocv.Point2f
	var schPts []gocv.Point2f

	for _, m := range matches {
		if len(m) == 2 && m[0].Distance < 0.7*m[1].Distance {
			goodMatches = append(goodMatches, m[0])
			schPts = append(schPts, gocv.Point2f{X: float32(kpSch[m[0].QueryIdx].X), Y: float32(kpSch[m[0].QueryIdx].Y)})
			srcPts = append(srcPts, gocv.Point2f{X: float32(kpSrc[m[0].TrainIdx].X), Y: float32(kpSrc[m[0].TrainIdx].Y)})
		}
	}

	if len(goodMatches) < 4 {
		return nil, errors.New("not enough good matches")
	}

	h, w := k.ImSearch.Rows(), k.ImSearch.Cols()
	middle, rect, confidence, err := utils.FindHomographyResult(srcPts, schPts, h, w)
	if err != nil {
		return nil, err
	}

	return &core.MatchResult{
		Result:     *middle,
		Rectangle:  rect,
		Confidence: confidence,
	}, nil
}
