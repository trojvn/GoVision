package GoVision

import (
	"github.com/trojvn/GoVision/engines"
	"github.com/trojvn/GoVision/utils"
	"gocv.io/x/gocv"
)

// Recognition - основной интерфейс для работы с модулем.
type Recognition struct {
	Threshold float32
}

func New(threshold float32) *Recognition {
	return &Recognition{
		Threshold: threshold,
	}
}

// FindTemplate ищет шаблон на экране, используя стандартный Template Matching.
func (r *Recognition) FindTemplate(screen, template gocv.Mat) (*MatchResult, error) {
	tm := engines.NewTemplateMatching(template, screen, r.Threshold)
	return tm.FindBestResult()
}

// FindMultiScale ищет шаблон на экране, перебирая масштабы.
func (r *Recognition) FindMultiScale(screen, template gocv.Mat, ratioMin, ratioMax, step float64) (*MatchResult, error) {
	res := engines.MultiScaleSearch(screen, template, ratioMin, ratioMax, step, r.Threshold)
	if res == nil {
		return nil, nil
	}

	return utils.NewMatchResult(res.MaxLoc, res.Width, res.Height, res.Confidence), nil
}
