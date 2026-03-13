package rcvgo

import (
	"errors"
	"image"

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

func (k *KAZEMatching) FindBestResult() (*MatchResult, error) {
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

	// Логика гомографии аналогична SIFT
	srcPointsMat := gocv.NewMatFromPoint2fVector(gocv.NewPoint2fVectorFromPoints(srcPts), false)
	defer srcPointsMat.Close()

	schPointsMat := gocv.NewMatFromPoint2fVector(gocv.NewPoint2fVectorFromPoints(schPts), false)
	defer schPointsMat.Close()

	maskHomography := gocv.NewMat()
	defer maskHomography.Close()

	homography := gocv.FindHomography(schPointsMat, srcPointsMat, gocv.HomographyMethodRANSAC, 5.0, &maskHomography, 2000, 0.995)
	if homography.Empty() {
		return nil, errors.New("homography matrix not found")
	}
	defer homography.Close()

	h, w := k.ImSearch.Rows(), k.ImSearch.Cols()
	ptsSch := []gocv.Point2f{{X: 0, Y: 0}, {X: 0, Y: float32(h - 1)}, {X: float32(w - 1), Y: float32(h - 1)}, {X: float32(w - 1), Y: 0}}
	ptsSchMat := gocv.NewMatFromPoint2fVector(gocv.NewPoint2fVectorFromPoints(ptsSch), false)
	defer ptsSchMat.Close()

	ptsDstMat := gocv.NewMat()
	defer ptsDstMat.Close()
	gocv.PerspectiveTransform(ptsSchMat, &ptsDstMat, homography)

	rectangle := make([]image.Point, 4)
	var sumX, sumY float32
	for i := 0; i < 4; i++ {
		px := ptsDstMat.GetFloatAt(i, 0)
		py := ptsDstMat.GetFloatAt(i, 1)
		rectangle[i] = image.Pt(int(px), int(py))
		sumX += px
		sumY += py
	}

	return &MatchResult{
		Result:     image.Pt(int(sumX/4), int(sumY/4)),
		Rectangle:  rectangle,
		Confidence: float32(len(goodMatches)) / 100.0, // Упрощенно
	}, nil
}
