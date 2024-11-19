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
	b := base[0]
	return math.Log(x) / math.Log(b)
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
	"max":   math_max,
	"min":   math_min,
	"modf":  MakeFunction("modf", math.Modf),

	"pi":  math.Pi,
	"sin": MakeFunction("sin", math.Sin),
	"tan": MakeFunction("tan", math.Tan),
})
