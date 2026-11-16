package std

import (
	"testing"

	. "github.com/Heliodex/coputer/litecode/types"
)

func BenchmarkAddFF(b *testing.B) {
	for v1, v2 := Num(float64(1)), Num(float64(2)); b.Loop(); {
		Add(v1, v2)
	}
}

func BenchmarkAddVV(b *testing.B) {
	for v1, v2 := VecVal(Vector{1, 2, 3, 0}), VecVal(Vector{5, 6, 7, 0}); b.Loop(); {
		Add(v1, v2)
	}
}

func BenchmarkF32Floor(b *testing.B) {
	for v := float32(1.5); b.Loop(); {
		f32Floor(v)
	}
}

func BenchmarkF32Sqrt(b *testing.B) {
	for v := float32(1.5); b.Loop(); {
		f32Sqrt(v)
	}
}

func BenchmarkIdivFF(b *testing.B) {
	for v1, v2 := Num(float64(1)), Num(float64(2)); b.Loop(); {
		Idiv(v1, v2)
	}
}

func BenchmarkIdivFV(b *testing.B) {
	for v1, v2 := Num(float64(1)), VecVal(Vector{5, 6, 7}); b.Loop(); {
		Idiv(v1, v2)
	}
}

func BenchmarkIdivVF(b *testing.B) {
	for v1, v2 := VecVal(Vector{1, 2, 3}), Num(float64(2)); b.Loop(); {
		Idiv(v1, v2)
	}
}

func BenchmarkIdivVV(b *testing.B) {
	for v1, v2 := VecVal(Vector{1, 2, 3}), VecVal(Vector{5, 6, 7}); b.Loop(); {
		Idiv(v1, v2)
	}
}
