package std

import (
	"testing"
)

func BenchmarkBitmask(b *testing.B) {
	for b.Loop() {
		for n := 1; n <= 32; n++ {
			bitmask(n)
		}
	}
}

func BenchmarkBshift(b *testing.B) {
	r := uint32(0)
	for b.Loop() {
		b_shift(r, -5)
		b_shift(r, -50)
		b_shift(r, 5)
		b_shift(r, 50)
	}
}
