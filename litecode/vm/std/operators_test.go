package std

import (
	"testing"

	. "github.com/Heliodex/coputer/litecode/types"
)

func BenchmarkF32Floor(b *testing.B) {
	v := float32(1.5)
	for b.Loop() {
		f32Floor(v)
	}
}

func BenchmarkF32Sqrt(b *testing.B) {
	v := float32(1.5)
	for b.Loop() {
		f32Sqrt(v)
	}
}

func BenchmarkAddFF(b *testing.B) {
	v1 := float64(1)
	v2 := float64(2)
	for b.Loop() {
		Add(v1, v2)
	}
}

func BenchmarkAddVV(b *testing.B) {
	v1 := Vector{1, 2, 3, 0}
	v2 := Vector{5, 6, 7, 0}
	for b.Loop() {
		Add(v1, v2)
	}
}

func BenchmarkIdivFF(b *testing.B) {
	v1 := float64(1)
	v2 := float64(2)
	for b.Loop() {
		Idiv(v1, v2)
	}
}

func BenchmarkIdivVV(b *testing.B) {
	v1 := Vector{1, 2, 3}
	v2 := Vector{5, 6, 7}
	for b.Loop() {
		Idiv(v1, v2)
	}
}

func BenchmarkIdivFV(b *testing.B) {
	v1 := float64(1)
	v2 := Vector{5, 6, 7}
	for b.Loop() {
		Idiv(v1, v2)
	}
}

func BenchmarkIdivVF(b *testing.B) {
	v1 := Vector{1, 2, 3}
	v2 := float64(2)
	for b.Loop() {
		Idiv(v1, v2)
	}
}
