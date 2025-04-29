package vm

import "math"

// float32-optimised maths, without converting from float32 to float64
// float32s are only used for vectors at the moment
const (
	uvnan = 0x7fe00000
	mask  = 0xff
	shift = 32 - 8 - 1
	bias  = 0x7f
)

var f32NaN = math.Float32frombits(uvnan)

func f32IsNaN(f float32) bool {
	return f != f
}

func f32IsInf(f float32, sign int) bool {
	return sign >= 0 && f > math.MaxFloat32 || sign <= 0 && f < -math.MaxFloat32
}

func f32Modf(f float32) (i, frac float32) {
	if f < 1 {
		if f < 0 {
			i, frac = f32Modf(-f)
			return -i, -frac
		}
		if f == 0 {
			return f, f // Return -0, -0 when f == -0
		}
		return 0, f
	}

	x := math.Float32bits(f)

	// Keep the top 9+e bits, the integer part; clear the rest.
	if e := uint(x>>shift)&mask - bias; e < 32-9 {
		x &^= 1<<(32-9-e) - 1
	}
	i = math.Float32frombits(x)
	frac = f - i
	return
}

func f32Floor(x float32) float32 {
	if x == 0 || f32IsNaN(x) || f32IsInf(x, 0) {
		return x
	}
	if x < 0 {
		d, fract := f32Modf(-x)
		if fract != 0 {
			d += 1
		}
		return -d
	}

	d, _ := f32Modf(x)
	return d
}

func f32Sqrt(x float32) float32 {
	// special cases
	if x == 0 || f32IsNaN(x) || f32IsInf(x, 1) {
		return x
	}
	if x < 0 {
		return f32NaN
	}
	ix := math.Float32bits(x)

	// normalize x
	exp := int(ix >> shift & mask)
	if exp == 0 { // subnormal x
		for ix&(1<<shift) == 0 {
			ix <<= 1
			exp--
		}
		exp++
	}
	exp -= bias // unbias exponent
	ix &^= mask << shift
	ix |= 1 << shift
	if exp&1 == 1 { // odd exp, double x to make it even
		ix <<= 1
	}
	exp >>= 1 // exp = exp/2, exponent of square root
	// generate sqrt(x) bit by bit
	ix <<= 1
	var q, s uint32 // q = sqrt(x)

	// r = moving bit from MSB to LSB
	for r := uint32(1 << (shift + 1)); r != 0; {
		if t := s + r; t <= ix {
			s = t + r
			ix -= t
			q += r
		}
		ix <<= 1
		r >>= 1
	}

	// final rounding
	if ix != 0 { // remainder, result not exact
		q += q & 1 // round according to extra bit
	}
	ix = q>>1 + uint32(exp-1+bias)<<shift // significand + biased exponent
	return math.Float32frombits(ix)
}

func f32Ceil(v float32) float32 {
	return float32(math.Ceil(float64(v)))
}

func f32Abs(v float32) float32 {
	return float32(math.Abs(float64(v)))
}

func vector_create(args Args) (r []Val, err error) {
	x := float32(args.GetNumber())
	y := float32(args.GetNumber())
	z := float32(args.GetNumber())
	var w float32
	if wide4 {
		w = float32(args.GetNumber())
	}

	return []Val{Vector{x, y, z, w}}, nil
}
