package vm

import (
	"math"

	"github.com/Heliodex/coputer/litecode/types"
)

const wide4 = false

func mag(v types.Vector) float32 {
	return f32Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2] + v[3]*v[3])
}

func vector_create(args Args) (r []types.Val, err error) {
	x := float32(args.GetNumber())
	y := float32(args.GetNumber())
	z := float32(args.GetNumber())
	var w float32
	if wide4 {
		w = float32(args.GetNumber())
	}

	return []types.Val{types.Vector{x, y, z, w}}, nil
}

func vector_magnitude(args Args) (r []types.Val, err error) {
	v := args.GetVector()

	return []types.Val{float64(mag(v))}, nil
}

func vector_normalize(args Args) (r []types.Val, err error) {
	v := args.GetVector()

	invSqrt := 1 / mag(v)
	return []types.Val{types.Vector{v[0] * invSqrt, v[1] * invSqrt, v[2] * invSqrt, v[3] * invSqrt}}, nil
}

func cross(a, b types.Vector) types.Vector {
	return types.Vector{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

func vector_cross(args Args) (r []types.Val, err error) {
	a, b := args.GetVector(), args.GetVector()

	return []types.Val{cross(a, b)}, nil
}

func vector_dot(args Args) (r []types.Val, err error) {
	a, b := args.GetVector(), args.GetVector()

	return []types.Val{float64(a[0]*b[0] + a[1]*b[1] + a[2]*b[2] + a[3]*b[3])}, nil
}

func vector_angle(args Args) (r []types.Val, err error) {
	a, b := args.GetVector(), args.GetVector()
	axis := args.GetVector(types.Vector{})

	c := cross(a, b)

	sinA := math.Sqrt(float64(c[0]*c[0] + c[1]*c[1] + c[2]*c[2]))
	cosA := float64(a[0]*b[0] + a[1]*b[1] + a[2]*b[2])
	angle := math.Atan2(sinA, cosA)

	if c[0]*axis[0]+c[1]*axis[1]+c[2]*axis[2] < 0 {
		return []types.Val{-angle}, nil
	}
	return []types.Val{angle}, nil
}

func vector_floor(args Args) (r []types.Val, err error) {
	v := args.GetVector()

	return []types.Val{types.Vector{f32Floor(v[0]), f32Floor(v[1]), f32Floor(v[2]), f32Floor(v[3])}}, nil
}

func vector_ceil(args Args) (r []types.Val, err error) {
	v := args.GetVector()

	return []types.Val{types.Vector{f32Ceil(v[0]), f32Ceil(v[1]), f32Ceil(v[2]), f32Ceil(v[3])}}, nil
}

func vector_abs(args Args) (r []types.Val, err error) {
	v := args.GetVector()

	return []types.Val{types.Vector{f32Abs(v[0]), f32Abs(v[1]), f32Abs(v[2]), f32Abs(v[3])}}, nil
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

func vector_sign(args Args) (r []types.Val, err error) {
	v := args.GetVector()

	return []types.Val{types.Vector{sign(v[0]), sign(v[1]), sign(v[2]), sign(v[3])}}, nil
}

func clamp(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func vector_clamp(args Args) (r []types.Val, err error) {
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

	return []types.Val{types.Vector{
		clamp(v[0], vmin[0], vmax[0]),
		clamp(v[1], vmin[1], vmax[1]),
		clamp(v[2], vmin[2], vmax[2]),
		clamp(v[3], vmin[3], vmax[3]),
	}}, nil
}

func vector_max(args Args) (r []types.Val, err error) {
	first := args.GetVector()

	result := types.Vector{first[0], first[1], first[2], first[3]}

	for range len(args.List) - 1 {
		b := args.GetVector()

		for i, v := range b {
			result[i] = max(result[i], v)
		}
	}

	return []types.Val{result}, nil
}

func vector_min(args Args) (r []types.Val, err error) {
	first := args.GetVector()

	result := types.Vector{first[0], first[1], first[2], first[3]}

	for range len(args.List) - 1 {
		b := args.GetVector()

		for i, v := range b {
			result[i] = min(result[i], v)
		}
	}

	return []types.Val{result}, nil
}

var libvector = NewLib([]types.Function{
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
}, map[string]types.Val{
	"one":  types.Vector{1, 1, 1, 0}, // 3-wide otherwise it breaks
	"zero": types.Vector{0, 0, 0, 0},
})
