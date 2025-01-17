package main

import "math"

func math_abs(args Args) Ret {
	x := args.GetNumber()

	return math.Abs(x)
}

func math_acos(args Args) Ret {
	x := args.GetNumber()

	return math.Acos(x)
}

func math_asin(args Args) Ret {
	x := args.GetNumber()

	return math.Asin(x)
}

func math_atan(args Args) Ret {
	x := args.GetNumber()
	return math.Atan(x)
}

func math_atan2(args Args) Ret {
	y, x := args.GetNumber(), args.GetNumber()

	return math.Atan2(y, x)
}

func math_ceil(args Args) Ret {
	x := args.GetNumber()

	return math.Ceil(x)
}

func math_clamp(args Args) Ret {
	x, min, max := args.GetNumber(), args.GetNumber(), args.GetNumber()

	if x < min {
		return min
	} else if x > max {
		return max
	}
	return x
}

func math_cos(args Args) Ret {
	x := args.GetNumber()

	return math.Cos(x)
}

func math_cosh(args Args) Ret {
	x := args.GetNumber()
	return math.Cosh(x)
}

func math_deg(args Args) Ret {
	x := args.GetNumber()

	return x * 180 / math.Pi
}

func math_exp(args Args) Ret {
	x := args.GetNumber()

	return math.Exp(x)
}

func math_floor(args Args) Ret {
	x := args.GetNumber()

	return math.Floor(x)
}

func math_fmod(args Args) Ret {
	x, y := args.GetNumber(), args.GetNumber()

	return math.Mod(x, y)
}

func math_frexp(args Args) Rets {
	x := args.GetNumber()

	frac, exp := math.Frexp(x)
	return Rets{frac, float64(exp)}
}

func math_ldexp(args Args) Ret {
	x, e := args.GetNumber(), args.GetNumber()

	return math.Ldexp(x, int(e))
}

func math_lerp(args Args) Ret {
	a, b, t := args.GetNumber(), args.GetNumber(), args.GetNumber()

	if t == 1 {
		return b
	}
	return a + (b-a)*t
}

func math_log(args Args) Ret {
	x := args.GetNumber()

	if len(args.args) > 1 {
		base := args.GetNumber()
		return math.Log(x) / math.Log(base)
	}
	return math.Log(x)
}

// Deprecated
// func math_log10(args Args) Ret {
// 	x := args.GetNumber()
// 	return math.Log10(x)
// }

func math_map(args Args) Ret {
	x, inmin, inmax, outmin, outmax := args.GetNumber(), args.GetNumber(), args.GetNumber(), args.GetNumber(), args.GetNumber()
	return outmin + (x-inmin)*(outmax-outmin)/(inmax-inmin)
}

// Go builtin math.Min and math.Max funciions don't handle nans and infs the same way
func math_max(args Args) Ret {
	first := args.GetNumber()
	for range len(args.args) - 1 {
		if n := args.GetNumber(); n > first {
			first = n
		}
	}
	return first
}

func math_min(args Args) Ret {
	first := args.GetNumber()
	for range len(args.args) - 1 {
		if n := args.GetNumber(); n < first {
			first = n
		}
	}
	return first
}

func math_modf(args Args) Rets {
	x := args.GetNumber()
	int, frac := math.Modf(x)
	return Rets{int, frac}
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

func math_noise(args Args) Ret {
	x, y, z := args.GetNumber(), args.GetNumber(0), args.GetNumber(0)

	return perlin(float32(x), float32(y), float32(z))
}

func math_pow(args Args) Ret {
	x, y := args.GetNumber(), args.GetNumber()

	return math.Pow(x, y)
}

func math_rad(args Args) Ret {
	x := args.GetNumber()

	return x * math.Pi / 180
}

func math_round(args Args) Ret {
	x := args.GetNumber()

	return math.Round(x)
}

func math_sign(args Args) Ret {
	x := args.GetNumber()

	if x > 0 {
		return float64(1)
	} else if x < 0 {
		return float64(-1)
	}
	return float64(0)
}

func math_sin(args Args) Ret {
	x := args.GetNumber()

	return math.Sin(x)
}

func math_sinh(args Args) Ret {
	x := args.GetNumber()

	return math.Sinh(x)
}

func math_sqrt(args Args) Ret {
	x := args.GetNumber()

	return math.Sqrt(x)
}

func math_tan(args Args) Ret {
	x := args.GetNumber()

	return math.Tan(x)
}

func math_tanh(args Args) Ret {
	x := args.GetNumber()

	return math.Tanh(x)
}

var libmath = NewTable([][2]any{
	MakeFn1("abs", math_abs),
	MakeFn1("acos", math_acos),
	MakeFn1("asin", math_asin),
	MakeFn1("atan", math_atan),
	MakeFn1("atan2", math_atan2),
	MakeFn1("ceil", math_ceil),
	MakeFn1("clamp", math_clamp),
	MakeFn1("cos", math_cos),
	MakeFn1("cosh", math_cosh),
	MakeFn1("deg", math_deg),
	MakeFn1("exp", math_exp),
	MakeFn1("floor", math_floor),
	MakeFn1("fmod", math_fmod),
	MakeFn("frexp", math_frexp),
	MakeFn1("ldexp", math_ldexp),
	MakeFn1("lerp", math_lerp),
	MakeFn1("log", math_log),
	// MakeFn("log10", math_log10), // deprecated
	MakeFn1("map", math_map), // w00t
	MakeFn1("max", math_max),
	MakeFn1("min", math_min),
	MakeFn("modf", math_modf),
	MakeFn1("noise", math_noise),
	MakeFn1("pow", math_pow),
	MakeFn1("rad", math_rad),
	// math.random and randomseed removed because we want determinism
	MakeFn1("round", math_round),
	MakeFn1("sign", math_sign),
	MakeFn1("sin", math_sin),
	MakeFn1("sinh", math_sinh),
	MakeFn1("sqrt", math_sqrt),
	MakeFn1("tan", math_tan),
	MakeFn1("tanh", math_tanh),

	{"huge", math.Inf(1)},
	{"pi", math.Pi},
})
