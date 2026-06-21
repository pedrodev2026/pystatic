package runtime

import (
	"math"
)

func MathSqrt(x float64) float64 {
	return math.Sqrt(x)
}

func MathAbs(x float64) float64 {
	return math.Abs(x)
}

func MathPow(x, y float64) float64 {
	return math.Pow(x, y)
}
