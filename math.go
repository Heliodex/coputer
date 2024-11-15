package main

import (
	"fmt"
	"math"
)

func invalidNumArgs(fn string, nx, i int) string {
	return fmt.Sprintf("missing argument #%d to '%s' (%d expected)", nx, fn, i)
}

func invalidArgType(i int, fn string, tx, tg string) string {
	return fmt.Sprintf("invalid argument #%d to '%s' (%s expected, got %s)", i, fn, luautype[tx], luautype[tg])
}

// doesn't exist in reference implementation as it discards all excess arguments, but whatever
func checkArgsLen(fn string, i int, args []any) {
	if len(args) != i {
		panic(invalidNumArgs(fn, i, len(args)))
	}
}

func checkMinArgsLen(fn string, i int, args []any) {
	if len(args) < i {
		panic(invalidNumArgs(fn, i, len(args)))
	}
}

func checkArg[T any](fn string, i int, arg any) T {
	v, ok := arg.(T)
	if !ok {
		panic(invalidArgType(i, fn, typeOf(v), typeOf(arg)))
	}
	return v
}

// functions

var math_abs = Function(func(args ...any) []any {
	checkArgsLen("abs", 1, args)
	v := checkArg[float64]("abs", 1, args[0])

	return []any{math.Abs(v)}
})

var math_floor = Function(func(args ...any) []any {
	checkArgsLen("floor", 1, args)
	v := checkArg[float64]("floor", 1, args[0])

	return []any{math.Floor(v)}
})

var math_max = Function(func(args ...any) []any {
	checkMinArgsLen("max", 1, args)
	v := checkArg[float64]("max", 1, args[0])

	for i := range args[1:] {
		v = math.Max(v, checkArg[float64]("max", i+1, args[i+1]))
	}
	return []any{v}
})

var math_min = Function(func(args ...any) []any {
	checkMinArgsLen("min", 1, args)
	v := checkArg[float64]("min", 1, args[0])

	for i := range args[1:] {
		v = math.Min(v, checkArg[float64]("min", i+1, args[i+1]))
	}
	return []any{v}
})

var libmath = NewTable(map[any]any{
	"abs":   &math_abs,
	"floor": &math_floor,
	"max":   &math_max,
	"min":   &math_min,
})
