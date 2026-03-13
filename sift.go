package rcvgo

import (
	"errors"
	"image"

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

func (s *SIFTMatching) FindBestResult() (*MatchResult, error) {
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

	// 3. Находим матрицу гомографии
	srcPointsMat := gocv.NewMatFromPoint2fVector(gocv.NewPoint2fVectorFromPoints(srcPts), false)
	defer srcPointsMat.Close()

	schPointsMat := gocv.NewMatFromPoint2fVector(gocv.NewPoint2fVectorFromPoints(schPts), false)
	defer schPointsMat.Close()

	maskHomography := gocv.NewMat()
	defer maskHomography.Close()

	// RANSAC для отсеивания выбросов
	homography := gocv.FindHomography(schPointsMat, srcPointsMat, gocv.HomographyMethodRANSAC, 5.0, &maskHomography, 2000, 0.995)
	if homography.Empty() {
		return nil, errors.New("homography matrix not found")
	}
	defer homography.Close()

	// 4. Трансформируем углы шаблона в координаты источника
	h, w := s.ImSearch.Rows(), s.ImSearch.Cols()
	ptsSch := []gocv.Point2f{
		{X: 0, Y: 0},
		{X: 0, Y: float32(h - 1)},
		{X: float32(w - 1), Y: float32(h - 1)},
		{X: float32(w - 1), Y: 0},
	}
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

	// Центр масс как результат
	middlePoint := image.Pt(int(sumX/4), int(sumY/4))

	// Рассчитываем уверенность на основе количества инлайеров (inliers)
	inliers := 0
	for i := 0; i < maskHomography.Rows(); i++ {
		if maskHomography.GetUCharAt(i, 0) > 0 {
			inliers++
		}
	}
	confidence := float32(inliers) / float32(len(goodMatches))

	return &MatchResult{
		Result:     middlePoint,
		Rectangle:  rectangle,
		Confidence: confidence,
	}, nil
}
