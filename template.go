package GoVision

import (
	"errors"
	"image"
	"path/filepath"
	"strings"

	"github.com/trojvn/GoVision/core"
	"github.com/trojvn/GoVision/engines"
	"github.com/trojvn/GoVision/utils"
	"gocv.io/x/gocv"
)

// Реэкспорт основных типов для удобства пользователя
type MatchResult = core.MatchResult
type TargetPos = core.TargetPos

const (
	MID = core.MID
	LT  = core.LT
	LB  = core.LB
	RB  = core.RB
	RT  = core.RT
)

type OptimizationLevel = core.OptimizationLevel

const (
	OptimizationMaximum  = core.OptimizationMaximum
	OptimizationBalanced = core.OptimizationBalanced
	OptimizationFast     = core.OptimizationFast
	OptimizationAccurate = core.OptimizationAccurate
)

// Template - аналог класса TP из Python.
type Template struct {
	Filename          string
	BaseFolder        string
	Threshold         float32
	TargetPos         TargetPos
	RecordPos         *[2]float64
	Resolution        image.Point
	RGB               bool
	ScaleMax          float64
	ScaleStep         float64
	OptimizationLevel OptimizationLevel
	Config            core.Config
}

func (t *Template) GetFilename() string {
	return t.Filename
}

func NewTemplate(filename string) *Template {
	return &Template{
		Filename:          filename,
		BaseFolder:        "images",
		Threshold:         0.9,
		TargetPos:         MID,
		RGB:               false,
		ScaleMax:          1000,
		ScaleStep:         0.005,
		OptimizationLevel: OptimizationAccurate,
		Config:            core.OptimizationMap[core.OptimizationAccurate],
	}
}

// Filepath возвращает полный путь к файлу шаблона.
func (t *Template) Filepath() string {
	filename := t.Filename
	if !strings.HasSuffix(strings.ToLower(filename), ".png") {
		filename += ".png"
	}
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(t.BaseFolder, filename)
}

// MatchIn ищет шаблон на экране и возвращает координаты целевой точки.
func (t *Template) MatchIn(screen gocv.Mat) (image.Point, error) {
	res, err := t.CVMatch(screen)
	if err != nil {
		return image.Point{}, err
	}
	if res == nil {
		return image.Point{}, errors.New("template not found")
	}
	return t.TargetPos.GetXY(*res), nil
}

// CVMatch реализует логику перебора методов (mstpl, gmstpl, sift).
func (t *Template) CVMatch(screen gocv.Mat) (*MatchResult, error) {
	if screen.Empty() {
		return nil, errors.New("screen image is empty")
	}

	path := t.Filepath()
	templateImg := gocv.IMRead(path, gocv.IMReadColor)
	if templateImg.Empty() {
		return nil, errors.New("could not read template image: " + path)
	}
	defer templateImg.Close()

	// 1. Подготовка изображения (Resize по стратегии)
	resizedTemplate := t.resizeImage(templateImg, screen)
	defer resizedTemplate.Close()

	var lastErr error

	// 2. Попытка MultiScaleTemplateMatching (mstpl/gmstpl)
	if t.Config.PyramidSearch || t.Config.AdaptiveStep {
		res := engines.MultiScaleSearch(screen, templateImg, 0.5, 1.5, t.ScaleStep, t.Threshold)
		if res != nil {
			return utils.NewMatchResult(res.MaxLoc, res.Width, res.Height, res.Confidence), nil
		}
	}

	// 3. Попытка обычного TemplateMatching
	tm := engines.NewTemplateMatching(resizedTemplate, screen, t.Threshold)
	tm.RGB = t.RGB
	tm.SmartRGB = t.Config.SmartRGB
	res, err := tm.FindBestResult()
	if err != nil {
		lastErr = err
	}
	if res != nil {
		return res, nil
	}

	// 4. Попытка SIFT (если ничего не помогло)
	sift := engines.NewSIFTMatching(resizedTemplate, screen, t.Threshold)
	res, err = sift.FindBestResult()
	if err != nil {
		lastErr = err
	}

	if res != nil {
		return res, nil
	}

	return nil, lastErr
}

func (t *Template) resizeImage(img, screen gocv.Mat) gocv.Mat {
	if t.Resolution.X == 0 || t.Resolution.Y == 0 {
		return img.Clone()
	}

	screenRes := image.Pt(screen.Cols(), screen.Rows())
	if t.Resolution == screenRes {
		return img.Clone()
	}

	// Реализация cocos_min_strategy
	designRes := image.Pt(960, 640)

	scaleSch := mathMin(
		float64(t.Resolution.X)/float64(designRes.X),
		float64(t.Resolution.Y)/float64(designRes.Y),
	)
	scaleSrc := mathMin(
		float64(screenRes.X)/float64(designRes.X),
		float64(screenRes.Y)/float64(designRes.Y),
	)
	scale := scaleSrc / scaleSch

	newW := int(float64(img.Cols()) * scale)
	newH := int(float64(img.Rows()) * scale)

	resized := gocv.NewMat()
	gocv.Resize(img, &resized, image.Pt(newW, newH), 0, 0, gocv.InterpolationLinear)
	return resized
}

func mathMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
