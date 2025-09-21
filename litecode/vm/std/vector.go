package std

import (
	"math"

	. "github.com/Heliodex/coputer/litecode/types"
)

const wide4 = false

func mag(v Vector) float32 {
	return f32Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2] + v[3]*v[3])
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

func vector_magnitude(args Args) (r []Val, err error) {
	v := args.GetVector()

	return []Val{float64(mag(v))}, nil
}

func vector_normalize(args Args) (r []Val, err error) {
	v := args.GetVector()

	invSqrt := 1 / mag(v)
	return []Val{Vector{v[0] * invSqrt, v[1] * invSqrt, v[2] * invSqrt, v[3] * invSqrt}}, nil
}

func cross(a, b Vector) Vector {
	return Vector{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

func vector_cross(args Args) (r []Val, err error) {
	a, b := args.GetVector(), args.GetVector()

	return []Val{cross(a, b)}, nil
}

func dot(a, b Vector) float32 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] + a[3]*b[3]
}

func vector_dot(args Args) (r []Val, err error) {
	a, b := args.GetVector(), args.GetVector()

	return []Val{float64(dot(a, b))}, nil
}

func angle(a, b, axis Vector) float64 {
	// cross(a, b)
	c := Vector{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}

	sinA := f32Sqrt(c[0]*c[0] + c[1]*c[1] + c[2]*c[2])
	// dot(a, b)
	cosA := a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
	angle := math.Atan2(float64(sinA), float64(cosA))

	// dot(c, axis)
	if c[0]*axis[0]+c[1]*axis[1]+c[2]*axis[2] < 0 {
		return -angle
	}
	return angle
}

func vector_angle(args Args) (r []Val, err error) {
	a, b := args.GetVector(), args.GetVector()
	axis := args.GetVector(Vector{})

	return []Val{angle(a, b, axis)}, nil
}

func vector_floor(args Args) (r []Val, err error) {
	v := args.GetVector()

	return []Val{Vector{f32Floor(v[0]), f32Floor(v[1]), f32Floor(v[2]), f32Floor(v[3])}}, nil
}

func vector_ceil(args Args) (r []Val, err error) {
	v := args.GetVector()

	return []Val{Vector{f32Ceil(v[0]), f32Ceil(v[1]), f32Ceil(v[2]), f32Ceil(v[3])}}, nil
}

func vector_abs(args Args) (r []Val, err error) {
	v := args.GetVector()

	return []Val{Vector{f32Abs(v[0]), f32Abs(v[1]), f32Abs(v[2]), f32Abs(v[3])}}, nil
}

func sign(v float32) float32 {
	if v > 0 {
		return 1
	}
	if v < 0 {
		return -1
	}
	return 0
}

func vector_sign(args Args) (r []Val, err error) {
	v := args.GetVector()

	return []Val{Vector{sign(v[0]), sign(v[1]), sign(v[2]), sign(v[3])}}, nil
}

func clamp(v, minv, maxv float32) float32 {
	if v < minv {
		return minv
	}
	if v > maxv {
		return maxv
	}
	return v
}

func vector_clamp(args Args) (r []Val, err error) {
	v := args.GetVector()
	vmin, vmax := args.GetVector(), args.GetVector()

	if vmin[0] > vmax[0] {
		// return nil, errors.New("max.x must be greater than or equal to min.x")
		vmin[0], vmax[0] = vmax[0], vmin[0]
	} else if vmin[1] > vmax[1] {
		// return nil, errors.New("max.y must be greater than or equal to min.y")
		vmin[1], vmax[1] = vmax[1], vmin[1]
	} else if vmin[2] > vmax[2] {
		// return nil, errors.New("max.z must be greater than or equal to min.z")
		vmin[2], vmax[2] = vmax[2], vmin[2]
	}

	return []Val{Vector{
		clamp(v[0], vmin[0], vmax[0]),
		clamp(v[1], vmin[1], vmax[1]),
		clamp(v[2], vmin[2], vmax[2]),
		clamp(v[3], vmin[3], vmax[3]),
	}}, nil
}

func vector_max(args Args) (r []Val, err error) {
	first := args.GetVector()

	result := Vector{first[0], first[1], first[2], first[3]}

	for range len(args.List) - 1 {
		b := args.GetVector()

		for i, v := range b {
			result[i] = max(result[i], v)
		}
	}

	return []Val{result}, nil
}

func vector_min(args Args) (r []Val, err error) {
	first := args.GetVector()

	result := Vector{first[0], first[1], first[2], first[3]}

	for range len(args.List) - 1 {
		b := args.GetVector()

		for i, v := range b {
			result[i] = min(result[i], v)
		}
	}

	return []Val{result}, nil
}

var Libvector = NewLib([]Function{
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
}, map[string]Val{
	"one":  Vector{1, 1, 1, 0}, // 3-wide otherwise it breaks
	"zero": Vector{0, 0, 0, 0},
})
