package litecode

import "math"

func math_abs(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{math.Abs(x)}, nil
}

func math_acos(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{math.Acos(x)}, nil
}

func math_asin(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{math.Asin(x)}, nil
}

func math_atan(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{math.Atan(x)}, nil
}

func math_atan2(args Args) (r Rets, err error) {
	y, x := args.GetNumber(), args.GetNumber()

	return Rets{math.Atan2(y, x)}, nil
}

func math_ceil(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{math.Ceil(x)}, nil
}

func math_clamp(args Args) (r Rets, err error) {
	x, min, max := args.GetNumber(), args.GetNumber(), args.GetNumber()

	if x < min {
		return Rets{min}, nil
	} else if x > max {
		return Rets{max}, nil
	}
	return Rets{x}, nil
}

func math_cos(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{math.Cos(x)}, nil
}

func math_cosh(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{math.Cosh(x)}, nil
}

func math_deg(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{x * 180 / math.Pi}, nil
}

func math_exp(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{math.Exp(x)}, nil
}

func math_floor(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{math.Floor(x)}, nil
}

func math_fmod(args Args) (r Rets, err error) {
	x, y := args.GetNumber(), args.GetNumber()

	return Rets{math.Mod(x, y)}, nil
}

func math_frexp(args Args) (r Rets, err error) {
	x := args.GetNumber()

	frac, exp := math.Frexp(x)
	return Rets{frac, float64(exp)}, nil
}

func math_ldexp(args Args) (r Rets, err error) {
	x, e := args.GetNumber(), args.GetNumber()

	return Rets{math.Ldexp(x, int(e))}, nil
}

func math_lerp(args Args) (r Rets, err error) {
	a, b, t := args.GetNumber(), args.GetNumber(), args.GetNumber()

	if t == 1 {
		return Rets{b}, nil
	}
	return Rets{a + (b-a)*t}, nil
}

func math_log(args Args) (r Rets, err error) {
	x := args.GetNumber()

	if len(args.List) > 1 {
		base := args.GetNumber()
		return Rets{math.Log(x) / math.Log(base)}, nil
	}
	return Rets{math.Log(x)}, nil
}

// Deprecated
// func math_log10(args Args) (r Rets, err error) {
// 	x := args.GetNumber()

// 	return Rets{math.Log10(x)}, nil
// }

func math_map(args Args) (r Rets, err error) {
	x, inmin, inmax, outmin, outmax := args.GetNumber(), args.GetNumber(), args.GetNumber(), args.GetNumber(), args.GetNumber()

	return Rets{outmin + (x-inmin)*(outmax-outmin)/(inmax-inmin)}, nil
}

// Go builtin math.Min and math.Max funciions don't handle nans and infs the same way
func math_max(args Args) (r Rets, err error) {
	first := args.GetNumber()

	for range len(args.List) - 1 {
		if n := args.GetNumber(); n > first {
			first = n
		}
	}
	return Rets{first}, nil
}

func math_min(args Args) (r Rets, err error) {
	first := args.GetNumber()

	for range len(args.List) - 1 {
		if n := args.GetNumber(); n < first {
			first = n
		}
	}
	return Rets{first}, nil
}

func math_modf(args Args) (r Rets, err error) {
	x := args.GetNumber()

	i, frac := math.Modf(x)
	return Rets{i, frac}, nil
}

// lmathlib.cpp
var kPerlinHash = [257]int{
	151, 160, 137, 91, 90, 15, 131, 13, 201, 95, 96, 53, 194, 233, 7, 225, 140, 36, 103, 30, 69, 142, 8, 99, 37, 240, 21, 10, 23,
	190, 6, 148, 247, 120, 234, 75, 0, 26, 197, 62, 94, 252, 219, 203, 117, 35, 11, 32, 57, 177, 33, 88, 237, 149, 56, 87, 174, 20,
	125, 136, 171, 168, 68, 175, 74, 165, 71, 134, 139, 48, 27, 166, 77, 146, 158, 231, 83, 111, 229, 122, 60, 211, 133, 230, 220, 105, 92,
	41, 55, 46, 245, 40, 244, 102, 143, 54, 65, 25, 63, 161, 1, 216, 80, 73, 209, 76, 132, 187, 208, 89, 18, 169, 200, 196, 135, 130,
	116, 188, 159, 86, 164, 100, 109, 198, 173, 186, 3, 64, 52, 217, 226, 250, 124, 123, 5, 202, 38, 147, 118, 126, 255, 82, 85, 212, 207,
	206, 59, 227, 47, 16, 58, 17, 182, 189, 28, 42, 223, 183, 170, 213, 119, 248, 152, 2, 44, 154, 163, 70, 221, 153, 101, 155, 167, 43,
	172, 9, 129, 22, 39, 253, 19, 98, 108, 110, 79, 113, 224, 232, 178, 185, 112, 104, 218, 246, 97, 228, 251, 34, 242, 193, 238, 210, 144,
	12, 191, 179, 162, 241, 81, 51, 145, 235, 249, 14, 239, 107, 49, 192, 214, 31, 181, 199, 106, 157, 184, 84, 204, 176, 115, 121, 50, 45,
	127, 4, 150, 254, 138, 236, 205, 93, 222, 114, 67, 29, 24, 72, 243, 141, 128, 195, 78, 66, 215, 61, 156, 180, 151,
}

var kPerlinGrad = [16][3]float32{
	{1, 1, 0},
	{-1, 1, 0},
	{1, -1, 0},
	{-1, -1, 0},
	{1, 0, 1},
	{-1, 0, 1},
	{1, 0, -1},
	{-1, 0, -1},
	{0, 1, 1},
	{0, -1, 1},
	{0, 1, -1},
	{0, -1, -1},
	{1, 1, 0},
	{0, -1, 1},
	{-1, 1, 0},
	{0, -1, -1},
}

func fade(t float32) float32 {
	return t * t * t * (t*(t*6-15) + 10)
}

func lerp(t, a, b float32) float32 {
	return a + t*(b-a)
}

func grad(hash int, x, y, z float32) float32 {
	g := kPerlinGrad[hash&15]
	return g[0]*x + g[1]*y + g[2]*z
}

func floor32(x float32) float32 {
	return float32(math.Floor(float64(x)))
}

func perlin(x, y, z float32) float64 {
	xflr, yflr, zflr := floor32(x), floor32(y), floor32(z)
	xi, yi, zi := int(xflr)&255, int(yflr)&255, int(zflr)&255
	xf, yf, zf := x-xflr, y-yflr, z-zflr
	u, v, w := fade(xf), fade(yf), fade(zf)

	p := kPerlinHash

	a := (p[xi] + yi) & 255
	aa, ab := (p[a]+zi)&255, (p[a+1]+zi)&255

	b := (p[xi+1] + yi) & 255
	ba, bb := (p[b]+zi)&255, (p[b+1]+zi)&255

	la := lerp(u, grad(p[aa], xf, yf, zf), grad(p[ba], xf-1, yf, zf))
	lb := lerp(u, grad(p[ab], xf, yf-1, zf), grad(p[bb], xf-1, yf-1, zf))
	la1 := lerp(u, grad(p[aa+1], xf, yf, zf-1), grad(p[ba+1], xf-1, yf, zf-1))
	lb1 := lerp(u, grad(p[ab+1], xf, yf-1, zf-1), grad(p[bb+1], xf-1, yf-1, zf-1))

	return float64(lerp(w, lerp(v, la, lb), lerp(v, la1, lb1)))
}

func math_noise(args Args) (r Rets, err error) {
	x, y, z := args.GetNumber(), args.GetNumber(0), args.GetNumber(0)

	return Rets{perlin(float32(x), float32(y), float32(z))}, nil
}

func math_pow(args Args) (r Rets, err error) {
	x, y := args.GetNumber(), args.GetNumber()

	return Rets{math.Pow(x, y)}, nil
}

func math_rad(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{x * math.Pi / 180}, nil
}

func math_round(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{math.Round(x)}, nil
}

func math_sign(args Args) (r Rets, err error) {
	x := args.GetNumber()

	if x > 0 {
		return Rets{float64(1)}, nil
	} else if x < 0 {
		return Rets{float64(-1)}, nil
	}
	return Rets{float64(0)}, nil
}

func math_sin(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{math.Sin(x)}, nil
}

func math_sinh(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{math.Sinh(x)}, nil
}

func math_sqrt(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{math.Sqrt(x)}, nil
}

func math_tan(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{math.Tan(x)}, nil
}

func math_tanh(args Args) (r Rets, err error) {
	x := args.GetNumber()

	return Rets{math.Tanh(x)}, nil
}

var libmath = NewTable([][2]any{
	MakeFn("abs", math_abs),
	MakeFn("acos", math_acos),
	MakeFn("asin", math_asin),
	MakeFn("atan", math_atan),
	MakeFn("atan2", math_atan2),
	MakeFn("ceil", math_ceil),
	MakeFn("clamp", math_clamp),
	MakeFn("cos", math_cos),
	MakeFn("cosh", math_cosh),
	MakeFn("deg", math_deg),
	MakeFn("exp", math_exp),
	MakeFn("floor", math_floor),
	MakeFn("fmod", math_fmod),
	MakeFn("frexp", math_frexp),
	MakeFn("ldexp", math_ldexp),
	MakeFn("lerp", math_lerp),
	MakeFn("log", math_log),
	// MakeFn("log10", math_log10), // deprecated
	MakeFn("map", math_map), // w00t
	MakeFn("max", math_max),
	MakeFn("min", math_min),
	MakeFn("modf", math_modf),
	MakeFn("noise", math_noise),
	MakeFn("pow", math_pow),
	MakeFn("rad", math_rad),
	// math.random and randomseed removed because we want determinism
	MakeFn("round", math_round),
	MakeFn("sign", math_sign),
	MakeFn("sin", math_sin),
	MakeFn("sinh", math_sinh),
	MakeFn("sqrt", math_sqrt),
	MakeFn("tan", math_tan),
	MakeFn("tanh", math_tanh),

	{"huge", math.Inf(1)},
	{"pi", math.Pi},
})
