package std

import (
	"testing"

	. "github.com/Heliodex/coputer/litecode/types"
)

func BenchmarkVectorMag(b *testing.B) {
	v := Vector{1, 2, 3}
	for b.Loop() {
		mag(v)
	}
}

func BenchmarkVectorCross(b *testing.B) {
	v1 := Vector{1, 2, 3}
	v2 := Vector{5, 6, 7}
	for b.Loop() {
		cross(v1, v2)
	}
}

func BenchmarkVectorDot(b *testing.B) {
	v1 := Vector{1, 2, 3}
	v2 := Vector{5, 6, 7}
	for b.Loop() {
		dot(v1, v2)
	}
}

func BenchmarkVectorAngle(b *testing.B) {
	v1 := Vector{1, 2, 3}
	v2 := Vector{5, 6, 7}
	axis := Vector{0, 0, 1}
	for b.Loop() {
		angle(v1, v2, axis)
	}
}

func BenchmarkVectorSign(b *testing.B) {
	v := Vector{1, -2}
	for b.Loop() {
		_ = Vector{sign(v[0]), sign(v[1]), sign(v[2]), sign(v[3])}
	}
}

func BenchmarkVectorClamp(b *testing.B) {
	v := Vector{1, -2, 3, -4}
	vmin := Vector{-1, -1, -1, -1}
	vmax := Vector{1, 1, 1, 1}
	for b.Loop() {
		_ = Vector{
			clamp(v[0], vmin[0], vmax[0]),
			clamp(v[1], vmin[1], vmax[1]),
			clamp(v[2], vmin[2], vmax[2]),
			clamp(v[3], vmin[3], vmax[3]),
		}
	}
}
