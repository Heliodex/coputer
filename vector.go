package litecode

import "math"

const wide4 = false

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
	return float32(math.Sqrt(float64(v[0]*v[0] + v[1]*v[1] + v[2]*v[2] + v[3]*v[3])))
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

func fFloor(v float32) float32 {
	return float32(math.Floor(float64(v)))
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

	return Rets{Vector{fFloor(v[0]), fFloor(v[1]), fFloor(v[2]), fFloor(v[3])}}, nil
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

	for range len(args.Args) - 1 {
		b := args.GetVector()
		if b[0] > result[0] {
			result[0] = b[0]
		}
		if b[1] > result[1] {
			result[1] = b[1]
		}
		if b[2] > result[2] {
			result[2] = b[2]
		}
		if b[3] > result[3] {
			result[3] = b[3]
		}
	}

	return Rets{result}, nil
}

func vector_min(args Args) (r Rets, err error) {
	first := args.GetVector()

	result := Vector{first[0], first[1], first[2], first[3]}

	for range len(args.Args) - 1 {
		b := args.GetVector()
		if b[0] < result[0] {
			result[0] = b[0]
		}
		if b[1] < result[1] {
			result[1] = b[1]
		}
		if b[2] < result[2] {
			result[2] = b[2]
		}
		if b[3] < result[3] {
			result[3] = b[3]
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
