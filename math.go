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

	newFn := Function(func(args ...any) []any {
		l := len(args)
		if (!variadic && l < argCount) || (variadic && l < argCount-1) {
			missingArg := tfn.In(l).String()
			fmt.Println("missingArg", missingArg)
			panic(invalidNumArgs(name, argCount, missingArg))
		}

		ml := min(l, argCount)

		vargs := make([]reflect.Value, ml)
		for i := range ml {
			a := args[i]

			var ta, ra reflect.Type
			if variadic && i >= argCount-1 {
				fmt.Println("in varidaci", args, args[i])
				vargs[i], ta, ra = reflect.ValueOf(args[i]), reflect.TypeOf(args[i]), tfn.In(argCount-1).Elem()
			} else {
				vargs[i], ta, ra = reflect.ValueOf(a), reflect.TypeOf(a), tfn.In(i)
			}

			if ta != ra {
				panic(invalidArgType(i+1, name, ra.String(), ta.String()))
			}
		}

		ret := vfn.Call(vargs)
		return []any{ret[0].Interface()}
	})
	return &newFn
}

// functions

var math_max = MakeFunction("max", func(first float64, rest ...float64) float64 {
	v := first
	for _, arg := range rest {
		v = math.Max(v, arg)
	}
	return v
})

var math_min = MakeFunction("min", func(first float64, rest ...float64) float64 {
	v := first
	for _, arg := range rest {
		v = math.Min(v, arg)
	}
	return v
})

var libmath = NewTable(map[any]any{
	"abs":   MakeFunction("abs", math.Abs),
	"floor": MakeFunction("floor", math.Floor),
	"max":   math_max,
	"min":   math_min,
	"pi":   math.Pi,
})
