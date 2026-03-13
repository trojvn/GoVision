package core

import (
	"image"
	"testing"
)

func TestTargetPos_GetXY(t *testing.T) {
	// Arrange
	res := MatchResult{
		Result: image.Pt(50, 50),
		Rectangle: []image.Point{
			image.Pt(10, 10), // LT
			image.Pt(10, 90), // LB
			image.Pt(90, 90), // RB
			image.Pt(90, 10), // RT
		},
		Confidence: 0.9,
	}

	tests := []struct {
		name     string
		tp       TargetPos
		expected image.Point
	}{
		{"MID", MID, image.Pt(50, 50)},
		{"LT", LT, image.Pt(10, 10)},
		{"LB", LB, image.Pt(10, 90)},
		{"RB", RB, image.Pt(90, 90)},
		{"RT", RT, image.Pt(90, 10)},
		{"Default (MID)", TargetPos(99), image.Pt(50, 50)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := tt.tp.GetXY(res)

			// Assert
			if got != tt.expected {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, got)
			}
		})
	}
}

func TestOptimizationMap(t *testing.T) {
	// Проверяем, что карта оптимизаций содержит все ожидаемые уровни
	levels := []OptimizationLevel{
		OptimizationMaximum,
		OptimizationBalanced,
		OptimizationFast,
		OptimizationAccurate,
	}

	for _, level := range levels {
		config, ok := OptimizationMap[level]
		if !ok {
			t.Errorf("OptimizationLevel %s not found in OptimizationMap", level)
			continue
		}

		// Базовая проверка конфига
		if config.Threshold <= 0 {
			t.Errorf("Config for %s has invalid threshold: %f", level, config.Threshold)
		}
	}
}
