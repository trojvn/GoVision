package rcvgo

type OptimizationLevel string

const (
	OptimizationMaximum  OptimizationLevel = "maximum"
	OptimizationBalanced OptimizationLevel = "balanced"
	OptimizationFast     OptimizationLevel = "fast"
	OptimizationAccurate OptimizationLevel = "accurate"
)

type Config struct {
	SmartRGB             bool
	SmartCascade         bool
	AdaptiveStep         bool
	TargetRatio          bool
	PyramidSearch        bool
	Threshold            float32
	RGBThresholdOffset   float32
	CascadeSkipThreshold float32
}

var OptimizationMap = map[OptimizationLevel]Config{
	OptimizationMaximum: {
		SmartRGB:           true,
		SmartCascade:       true,
		AdaptiveStep:       true,
		TargetRatio:        true,
		PyramidSearch:      true,
		Threshold:          0.9,
		RGBThresholdOffset: 0.15,
	},
	OptimizationBalanced: {
		SmartRGB:           true,
		SmartCascade:       true,
		AdaptiveStep:       true,
		TargetRatio:        true,
		PyramidSearch:      false,
		Threshold:          0.9,
		RGBThresholdOffset: 0.15,
	},
	OptimizationFast: {
		SmartRGB:           true,
		SmartCascade:       true,
		AdaptiveStep:       false,
		TargetRatio:        false,
		PyramidSearch:      false,
		Threshold:          0.9,
		RGBThresholdOffset: 0.15,
	},
	OptimizationAccurate: {
		SmartRGB:           true,
		SmartCascade:       false,
		AdaptiveStep:       false,
		TargetRatio:        false,
		PyramidSearch:      false,
		Threshold:          0.9,
		RGBThresholdOffset: 0.15,
	},
}
