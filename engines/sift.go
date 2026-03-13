package engines

import (
	"errors"

	"github.com/trojvn/rcvgo/core"
	"github.com/trojvn/rcvgo/utils"
	"gocv.io/x/gocv"
)

// SIFTMatching реализует поиск по ключевым точкам SIFT с использованием гомографии.
type SIFTMatching struct {
	ImSource  gocv.Mat
	ImSearch  gocv.Mat
	Threshold float32
}

func NewSIFTMatching(imSearch, imSource gocv.Mat, threshold float32) *SIFTMatching {
	return &SIFTMatching{
		ImSource:  imSource,
		ImSearch:  imSearch,
		Threshold: threshold,
	}
}

func (s *SIFTMatching) FindBestResult() (*core.MatchResult, error) {
	if s.ImSource.Empty() || s.ImSearch.Empty() {
		return nil, errors.New("empty images")
	}

	// 1. Находим ключевые точки и дескрипторы
	detector := gocv.NewSIFT()
	defer detector.Close()

	mask := gocv.NewMat()
	defer mask.Close()

	kpSch, desSch := detector.DetectAndCompute(s.ImSearch, mask)
	defer desSch.Close()
	kpSrc, desSrc := detector.DetectAndCompute(s.ImSource, mask)
	defer desSrc.Close()

	if len(kpSch) < 4 || len(kpSrc) < 4 {
		return nil, errors.New("not enough keypoints found")
	}

	// 2. Сопоставление дескрипторов (KNN Match)
	matcher := gocv.NewBFMatcherWithParams(gocv.NormL2, false)
	defer matcher.Close()

	knnMatches := matcher.KnnMatch(desSch, desSrc, 2)

	var goodMatches []gocv.DMatch
	var srcPts []gocv.Point2f
	var schPts []gocv.Point2f

	// Фильтр Лоу (Lowe's ratio test)
	for _, m := range knnMatches {
		if len(m) == 2 && m[0].Distance < 0.7*m[1].Distance {
			goodMatches = append(goodMatches, m[0])
			schPts = append(schPts, gocv.Point2f{X: float32(kpSch[m[0].QueryIdx].X), Y: float32(kpSch[m[0].QueryIdx].Y)})
			srcPts = append(srcPts, gocv.Point2f{X: float32(kpSrc[m[0].TrainIdx].X), Y: float32(kpSrc[m[0].TrainIdx].Y)})
		}
	}

	if len(goodMatches) < 4 {
		return nil, errors.New("not enough good matches after ratio test")
	}

	h, w := s.ImSearch.Rows(), s.ImSearch.Cols()
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
