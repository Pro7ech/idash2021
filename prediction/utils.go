package prediction

import (
	"math"
)

func SoftMax(vec []float64) {
	max := 0.0
	for i := range vec {
		max = math.Max(max, vec[i])
	}

	sumExp := 0.0
	for i := range vec {
		sumExp += math.Exp(vec[i] - max)
	}

	for i := range vec {
		vec[i] = math.Exp(vec[i]-max) / sumExp
	}
}

func Sigmoid(vec []float64) {
	for i := range vec {
		vec[i] = 1 / (math.Exp(-vec[i]) + 1)
	}
}

func MaxIndex(vec []float64) int {
	max := -4294967295.0
	idx := 0
	for i, v := range vec {
		if v > max {
			max = v
			idx = i
		}
	}
	return idx
}
