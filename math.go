package main

import (
	"fmt"
	"math"
	"reflect"
)

func invalidNumArgs(fn string, nx int, tx string) string {
	return fmt.Sprintf("missing argument #%d to '%s' (%s expected)", nx, fn, luautype[tx])
}

func invalidArgType(i int, fn string, tx, tg string) string {
	return fmt.Sprintf("invalid argument #%d to '%s' (%s expected, got %s)", i, fn, luautype[tx], luautype[tg])
}

func MakeFunction(name string, fn any) *Function {
	vfn, tfn := reflect.ValueOf(fn), reflect.TypeOf(fn)
	if tfn.Kind() != reflect.Func {
		panic("protected item is not a function")
	}

	argCount := tfn.NumIn()
	variadic := tfn.IsVariadic()

	newFn := Function(func(args ...any) (rets []any) {
		l := len(args)
		if (!variadic && l < argCount) || (variadic && l < argCount-1) {
			missingArg := tfn.In(l).String()
			// fmt.Println("missingArg", missingArg)
			panic(invalidNumArgs(name, argCount, missingArg))
		}

		ml := min(l, argCount)

		vargs := make([]reflect.Value, ml)
		for i := range ml {
			a := args[i]

			var ta, ra reflect.Type
			if variadic && i >= argCount-1 {
				// fmt.Println("in varidaci", args, args[i])
				vargs[i], ta, ra = reflect.ValueOf(args[i]), reflect.TypeOf(args[i]), tfn.In(argCount-1).Elem()
			} else {
				vargs[i], ta, ra = reflect.ValueOf(a), reflect.TypeOf(a), tfn.In(i)
			}

			if ta != ra {
				panic(invalidArgType(i+1, name, ra.String(), ta.String()))
			}
		}

		ret := vfn.Call(vargs)
		for _, v := range ret {
			rets = append(rets, v.Interface())
		}
		return
	})
	return &newFn
}

// functions

var math_clamp = MakeFunction("clamp", func(x, min, max float64) float64 {
	if x < min {
		return min
	} else if x > max {
		return max
	}
	return x
})

var math_deg = MakeFunction("deg", func(x float64) float64 {
	return x * 180 / math.Pi
})

var math_frexp = MakeFunction("frexp", func(x float64) (float64, float64) {
	frac, exp := math.Frexp(x)
	return frac, float64(exp)
})

var math_ldexp = MakeFunction("ldexp", func(x, e float64) float64 {
	return math.Ldexp(x, int(e))
})

var math_log = MakeFunction("log", func(x float64, base ...float64) float64 {
	if len(base) == 0 {
		return math.Log(x)
	}
	return math.Log(x) / math.Log(base[0])
})

var math_map = MakeFunction("map", func(x, inmin, inmax, outmin, outmax float64) float64 {
	return outmin + (x-inmin)*(outmax-outmin)/(inmax-inmin)
})

var math_max = MakeFunction("max", func(first float64, rest ...float64) float64 {
	for _, arg := range rest {
		first = math.Max(first, arg)
	}
	return first
})

var math_min = MakeFunction("min", func(first float64, rest ...float64) float64 {
	for _, arg := range rest {
		first = math.Min(first, arg)
	}
	return first
})

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

var math_noise = MakeFunction("noise", func(x float64, rest ...float64) float64 {
	var y, z float64
	if len(rest) > 0 {
		y = rest[0]
	}
	if len(rest) > 1 {
		z = rest[1]
	}

	return perlin(float32(x), float32(y), float32(z))
})

var math_rad = MakeFunction("rad", func(x float64) float64 {
	return x * math.Pi / 180
})

var math_sign = MakeFunction("sign", func(x float64) float64 {
	if x > 0 {
		return 1
	} else if x < 0 {
		return -1
	}
	return 0
})

var libmath = NewTable(map[any]any{
	"abs":   MakeFunction("abs", math.Abs),
	"acos":  MakeFunction("acos", math.Acos),
	"asin":  MakeFunction("asin", math.Asin),
	"atan":  MakeFunction("atan", math.Atan),
	"atan2": MakeFunction("atan2", math.Atan2),
	"ceil":  MakeFunction("ceil", math.Ceil),
	"clamp": math_clamp,
	"cos":   MakeFunction("cos", math.Cos),
	"cosh":  MakeFunction("cosh", math.Cosh),
	"deg":   math_deg,
	"exp":   MakeFunction("exp", math.Exp),
	"floor": MakeFunction("floor", math.Floor),
	"fmod":  MakeFunction("fmod", math.Mod),
	"frexp": math_frexp,
	"ldexp": math_ldexp,
	"log":   math_log,
	"log10": MakeFunction("log10", math.Log10),
	"map":   math_map,
	"max":   math_max,
	"min":   math_min,
	"modf":  MakeFunction("modf", math.Modf),
	"noise": math_noise,
	"pow":   MakeFunction("pow", math.Pow),
	"rad":   math_rad,
	// math.random and randomseed removed because we want determinism
	"round": MakeFunction("round", math.Round),
	"sign":  math_sign,
	"sin":   MakeFunction("sin", math.Sin),
	"sinh":  MakeFunction("sinh", math.Sinh),
	"sqrt":  MakeFunction("sqrt", math.Sqrt),
	"tan":   MakeFunction("tan", math.Tan),
	"tanh":  MakeFunction("tanh", math.Tanh),

	"huge": math.Inf(1),
	"pi":   math.Pi,
})
