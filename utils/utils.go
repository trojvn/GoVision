package utils

import (
	"errors"
	"image"

	"github.com/trojvn/rcvgo/core"
	"gocv.io/x/gocv"
)

// GetResolution возвращает ширину и высоту изображения.
func GetResolution(img gocv.Mat) (int, int) {
	return img.Cols(), img.Rows()
}

// FindHomographyResult находит матрицу гомографии и трансформирует углы шаблона.
func FindHomographyResult(srcPts, schPts []gocv.Point2f, h, w int) (*image.Point, []image.Point, float32, error) {
	if len(srcPts) < 4 {
		return nil, nil, 0, errors.New("not enough points for homography")
	}

	srcPointsMat := gocv.NewMatFromPoint2fVector(gocv.NewPoint2fVectorFromPoints(srcPts), false)
	defer srcPointsMat.Close()
	schPointsMat := gocv.NewMatFromPoint2fVector(gocv.NewPoint2fVectorFromPoints(schPts), false)
	defer schPointsMat.Close()

	maskHomography := gocv.NewMat()
	defer maskHomography.Close()

	homography := gocv.FindHomography(schPointsMat, srcPointsMat, gocv.HomographyMethodRANSAC, 5.0, &maskHomography, 2000, 0.995)
	if homography.Empty() {
		return nil, nil, 0, errors.New("homography matrix not found")
	}
	defer homography.Close()

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

	middlePoint := image.Pt(int(sumX/4), int(sumY/4))

	inliers := 0
	for i := 0; i < maskHomography.Rows(); i++ {
		if maskHomography.GetUCharAt(i, 0) > 0 {
			inliers++
		}
	}
	confidence := float32(inliers) / float32(len(srcPts))

	return &middlePoint, rectangle, confidence, nil
}

// CropImage обрезает изображение по заданному прямоугольнику.
func CropImage(img gocv.Mat, rect image.Rectangle) (gocv.Mat, error) {
	height, width := img.Rows(), img.Cols()

	xMin := clamp(rect.Min.X, 0, width-1)
	yMin := clamp(rect.Min.Y, 0, height-1)
	xMax := clamp(rect.Max.X, 0, width-1)
	yMax := clamp(rect.Max.Y, 0, height-1)

	if xMin >= xMax || yMin >= yMax {
		return gocv.NewMat(), errors.New("invalid crop area")
	}

	return img.Region(image.Rect(xMin, yMin, xMax, yMax)), nil
}

// ImgMatToGray конвертирует изображение в оттенки серого.
func ImgMatToGray(img gocv.Mat) gocv.Mat {
	gray := gocv.NewMat()
	if img.Channels() == 1 {
		img.CopyTo(&gray)
		return gray
	}
	gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	return gray
}

// CheckSourceLargerThanSearch проверяет, что исходное изображение больше шаблона.
func CheckSourceLargerThanSearch(imSource, imSearch gocv.Mat) error {
	if imSearch.Rows() > imSource.Rows() || imSearch.Cols() > imSource.Cols() {
		return errors.New("imSearch is bigger than imSource")
	}
	return nil
}

// NewMatchResult создает MatchResult и рассчитывает прямоугольник и центр.
func NewMatchResult(maxLoc image.Point, w, h int, confidence float32) *core.MatchResult {
	middle := image.Pt(maxLoc.X+w/2, maxLoc.Y+h/2)
	rect := []image.Point{
		maxLoc,
		image.Pt(maxLoc.X, maxLoc.Y+h),
		image.Pt(maxLoc.X+w, maxLoc.Y+h),
		image.Pt(maxLoc.X+w, maxLoc.Y),
	}
	return &core.MatchResult{
		Result:     middle,
		Rectangle:  rect,
		Confidence: confidence,
	}
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
