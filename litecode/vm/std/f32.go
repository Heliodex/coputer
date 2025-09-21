package std

import "math"

// this is at least 2x as fast as f32-native floor function
func f32Floor(v float32) float32 {
	return float32(math.Floor(float64(v)))
}

// ...and this one's at least 10x faster, because machine instructions...
func f32Sqrt(v float32) float32 {
	return float32(math.Sqrt(float64(v)))
}

func f32Ceil(v float32) float32 {
	return float32(math.Ceil(float64(v)))
}

func f32Abs(v float32) float32 {
	return float32(math.Abs(float64(v)))
}
