package rcvgo

import (
	"errors"
	"image"

	"gocv.io/x/gocv"
)

// GetResolution возвращает ширину и высоту изображения.
func GetResolution(img gocv.Mat) (int, int) {
	return img.Cols(), img.Rows()
}

// CropImage обрезает изображение по заданному прямоугольнику [xMin, yMin, xMax, yMax].
func CropImage(img gocv.Mat, rect []int) (gocv.Mat, error) {
	if len(rect) != 4 {
		return gocv.NewMat(), errors.New("rect should be [xMin, yMin, xMax, yMax]")
	}

	height, width := img.Rows(), img.Cols()
	xMin, yMin, xMax, yMax := rect[0], rect[1], rect[2], rect[3]

	// Валидация границ
	xMin = clamp(xMin, 0, width-1)
	yMin = clamp(yMin, 0, height-1)
	xMax = clamp(xMax, 0, width-1)
	yMax = clamp(yMax, 0, height-1)

	if xMin >= xMax || yMin >= yMax {
		return gocv.NewMat(), errors.New("invalid crop area")
	}

	region := image.Rect(xMin, yMin, xMax, yMax)
	return img.Region(region), nil
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

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
