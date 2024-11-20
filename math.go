package main

import (
	"math"
)

// functions

func math_clamp(x, min, max float64) float64 {
	if x < min {
		return min
	} else if x > max {
		return max
	}
	return x
}

func math_deg(x float64) float64 {
	return x * 180 / math.Pi
}

func math_frexp(x float64) (float64, float64) {
	frac, exp := math.Frexp(x)
	return frac, float64(exp)
}

func math_ldexp(x, e float64) float64 {
	return math.Ldexp(x, int(e))
}

func math_log(x float64, base ...float64) float64 {
	if len(base) == 0 {
		return math.Log(x)
	}
	return math.Log(x) / math.Log(base[0])
}

func math_map(x, inmin, inmax, outmin, outmax float64) float64 {
	return outmin + (x-inmin)*(outmax-outmin)/(inmax-inmin)
}

func math_max(first float64, rest ...float64) float64 {
	for _, arg := range rest {
		first = math.Max(first, arg)
	}
	return first
}

func math_min(first float64, rest ...float64) float64 {
	for _, arg := range rest {
		first = math.Min(first, arg)
	}
	return first
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

func math_noise(x float64, rest ...float64) float64 {
	var y, z float64
	if len(rest) > 0 {
		y = rest[0]
	}
	if len(rest) > 1 {
		z = rest[1]
	}

	return perlin(float32(x), float32(y), float32(z))
}

func math_rad(x float64) float64 {
	return x * math.Pi / 180
}

func math_sign(x float64) float64 {
	if x > 0 {
		return 1
	} else if x < 0 {
		return -1
	}
	return 0
}

var libmath = NewTable([][2]any{
	MakeFn("abs", math.Abs),
	MakeFn("acos", math.Acos),
	MakeFn("asin", math.Asin),
	MakeFn("atan", math.Atan),
	MakeFn("atan2", math.Atan2),
	MakeFn("ceil", math.Ceil),
	MakeFn("clamp", math_clamp),
	MakeFn("cos", math.Cos),
	MakeFn("cosh", math.Cosh),
	MakeFn("deg", math_deg),
	MakeFn("exp", math.Exp),
	MakeFn("floor", math.Floor),
	MakeFn("fmod", math.Mod),
	MakeFn("frexp", math_frexp),
	MakeFn("ldexp", math_ldexp),
	MakeFn("log", math_log),
	MakeFn("log10", math.Log10),
	MakeFn("map", math_map), // w00t
	MakeFn("max", math_max),
	MakeFn("min", math_min),
	MakeFn("modf", math.Modf),
	MakeFn("noise", math_noise),
	MakeFn("pow", math.Pow),
	MakeFn("rad", math_rad),
	// math.random and randomseed removed because we want determinism
	MakeFn("round", math.Round),
	MakeFn("sign", math_sign),
	MakeFn("sin", math.Sin),
	MakeFn("sinh", math.Sinh),
	MakeFn("sqrt", math.Sqrt),
	MakeFn("tan", math.Tan),
	MakeFn("tanh", math.Tanh),

	{"huge", math.Inf(1)},
	{"pi", math.Pi},
})
