package utils

import "math/rand"

func RandRangeInt(min, max int) int {
	if min == max {
		return min
	}
	return min + rand.Intn(max-min)
}

func RandRangeFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
