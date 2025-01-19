package litecode

import "math"

const wide4 = false

// float32 maths
const (
	uvnan = 0x7FE00000
	mask  = 0xFF
	shift = 32 - 8 - 1
	bias  = 127
)

func f32NaN() float32 {
	return math.Float32frombits(uvnan)
}

func f32IsNaN(f float32) bool {
	return f != f
}

func f32IsInf(f float32, sign int) bool {
	return sign >= 0 && f > math.MaxFloat32 || sign <= 0 && f < -math.MaxFloat32
}

func f32Modf(f float32) (i float32, frac float32) {
	if f < 1 {
		switch {
		case f < 0:
			i, frac = f32Modf(-f)
			return -i, -frac
		case f == 0:
			return f, f // Return -0, -0 when f == -0
		}
		return 0, f
	}

	x := math.Float32bits(f)
	e := uint(x>>shift)&mask - bias

	// Keep the top 9+e bits, the integer part; clear the rest.
	if e < 32-9 {
		x &^= 1<<(32-9-e) - 1
	}
	i = math.Float32frombits(x)
	frac = f - i
	return
}

func f32Floor(x float32) float32 {
	if x == 0 || f32IsNaN(x) || f32IsInf(x, 0) {
		return x
	} else if x < 0 {
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
	switch {
	case x == 0 || f32IsNaN(x) || f32IsInf(x, 1):
		return x
	case x < 0:
		return f32NaN()
	}
	ix := math.Float32bits(x)

	// normalize x
	exp := int((ix >> shift) & mask)
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

func vector_create(args Args) (r Rets, err error) {
	x := float32(args.GetNumber())
	y := float32(args.GetNumber())
	z := float32(args.GetNumber())
	var w float32
	if wide4 {
		w = float32(args.GetNumber())
	}

	return Rets{Vector{x, y, z, w}}, nil
}

func mag(v Vector) float32 {
	return f32Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2] + v[3]*v[3])
}

func vector_magnitude(args Args) (r Rets, err error) {
	v := args.GetVector()

	return Rets{float64(mag(v))}, nil
}

func vector_normalize(args Args) (r Rets, err error) {
	v := args.GetVector()

	invSqrt := 1 / mag(v)
	return Rets{Vector{v[0] * invSqrt, v[1] * invSqrt, v[2] * invSqrt, v[3] * invSqrt}}, nil
}

func cross(a, b Vector) Vector {
	return Vector{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

func vector_cross(args Args) (r Rets, err error) {
	a := args.GetVector()
	b := args.GetVector()

	return Rets{cross(a, b)}, nil
}

func vector_dot(args Args) (r Rets, err error) {
	a := args.GetVector()
	b := args.GetVector()

	return Rets{float64(a[0]*b[0] + a[1]*b[1] + a[2]*b[2] + a[3]*b[3])}, nil
}

func fPow(a, b float32) float32 {
	return float32(math.Pow(float64(a), float64(b)))
}

func fCeil(v float32) float32 {
	return float32(math.Ceil(float64(v)))
}

func fAbs(v float32) float32 {
	return float32(math.Abs(float64(v)))
}

func vector_angle(args Args) (r Rets, err error) {
	a := args.GetVector()
	b := args.GetVector()
	axis := args.GetVector(Vector{})

	c := cross(a, b)

	sinA := math.Sqrt(float64(c[0]*c[0] + c[1]*c[1] + c[2]*c[2]))
	cosA := float64(a[0]*b[0] + a[1]*b[1] + a[2]*b[2])
	angle := math.Atan2(sinA, cosA)

	if (c[0]*axis[0] + c[1]*axis[1] + c[2]*axis[2]) < 0 {
		return Rets{-angle}, nil
	}
	return Rets{angle}, nil
}

func vector_floor(args Args) (r Rets, err error) {
	v := args.GetVector()

	return Rets{Vector{f32Floor(v[0]), f32Floor(v[1]), f32Floor(v[2]), f32Floor(v[3])}}, nil
}

func vector_ceil(args Args) (r Rets, err error) {
	v := args.GetVector()

	return Rets{Vector{fCeil(v[0]), fCeil(v[1]), fCeil(v[2]), fCeil(v[3])}}, nil
}

func vector_abs(args Args) (r Rets, err error) {
	v := args.GetVector()

	return Rets{Vector{fAbs(v[0]), fAbs(v[1]), fAbs(v[2]), fAbs(v[3])}}, nil
}

func sign(v float32) float32 {
	if v > 0 {
		return 1
	} else if v < 0 {
		return -1
	}
	return 0
}

func vector_sign(args Args) (r Rets, err error) {
	v := args.GetVector()

	return Rets{Vector{sign(v[0]), sign(v[1]), sign(v[2]), sign(v[3])}}, nil
}

func clamp(v, min, max float32) float32 {
	if v < min {
		return min
	} else if v > max {
		return max
	}
	return v
}

func vector_clamp(args Args) (r Rets, err error) {
	v := args.GetVector()
	min := args.GetVector()
	max := args.GetVector()

	if min[0] > max[0] {
		// return nil, errors.New("max.x must be greater than or equal to min.x")
		min[0], max[0] = max[0], min[0]
	} else if min[1] > max[1] {
		// return nil, errors.New("max.y must be greater than or equal to min.y")
		min[1], max[1] = max[1], min[1]
	} else if min[2] > max[2] {
		// return nil, errors.New("max.z must be greater than or equal to min.z")
		min[2], max[2] = max[2], min[2]
	}

	return Rets{Vector{
		clamp(v[0], min[0], max[0]),
		clamp(v[1], min[1], max[1]),
		clamp(v[2], min[2], max[2]),
		clamp(v[3], min[3], max[3]),
	}}, nil
}

func vector_max(args Args) (r Rets, err error) {
	first := args.GetVector()

	result := Vector{first[0], first[1], first[2], first[3]}

	for range len(args.List) - 1 {
		b := args.GetVector()
		for i, v := range b {
			if v > result[i] {
				result[i] = v
			}
		}
	}

	return Rets{result}, nil
}

func vector_min(args Args) (r Rets, err error) {
	first := args.GetVector()

	result := Vector{first[0], first[1], first[2], first[3]}

	for range len(args.List) - 1 {
		b := args.GetVector()
		for i, v := range b {
			if v < result[i] {
				result[i] = v
			}
		}
	}

	return Rets{result}, nil
}

var libvector = NewTable([][2]any{
	MakeFn("create", vector_create),
	MakeFn("magnitude", vector_magnitude),
	MakeFn("normalize", vector_normalize),
	MakeFn("cross", vector_cross),
	MakeFn("dot", vector_dot),
	MakeFn("angle", vector_angle),
	MakeFn("floor", vector_floor),
	MakeFn("ceil", vector_ceil),
	MakeFn("abs", vector_abs),
	MakeFn("sign", vector_sign),
	MakeFn("clamp", vector_clamp),
	MakeFn("max", vector_max),
	MakeFn("min", vector_min),

	{"one", Vector{1, 1, 1, 0}}, // 3-wide otherwise it breaks
	{"zero", Vector{0, 0, 0, 0}},
})
