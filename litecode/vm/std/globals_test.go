package std

import "testing"

func TestMul128(t *testing.T) {
	const x = 100
	const y = 0x11111111_11111111

	rx, ry := mul128(x, y)

	if rx != 6 || ry != 0xaaaaaaaa_aaaaaaa4 {
		t.Fatalf("mul128 failed, got %x, %x", rx, ry)
	}
}

func BenchmarkMul128(b *testing.B) {
	const x = 100
	const y = 0x11111111_11111111

	for b.Loop() {
		mul128(x, y)
	}
}

func TestMul192(t *testing.T) {
	const xhi = 100
	const xlo = 0x11111111_11111111
	const y = 0x22222222_22222222

	rx, ry := mul192hi(xhi, xlo, y)

	if rx != 0xd || ry != 0x579be024_68acf128 {
		t.Fatalf("mul128 failed, got %x, %x", rx, ry)
	}
}

func BenchmarkMul192(b *testing.B) {
	const xhi = 100
	const xlo = 0x11111111_11111111
	const y = 0x22222222_22222222
	
	for b.Loop() {
		mul192hi(xhi, xlo, y)
	}
}
