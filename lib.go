package main

import "fmt"

func invalidNumArgs(fn string, nx int, tx ...string) (res string) {
	res += fmt.Sprintf("missing argument #%d to '%s'", nx, fn)
	if len(tx) > 0 {
		res += fmt.Sprintf(" (%s expected)", tx[0])
	}
	return
}

func invalidArgType(i int, fn string, tx, tg string) string {
	return fmt.Sprintf("invalid argument #%d to '%s' (%s expected, got %s)", i, fn, luautype[tx], luautype[tg])
}

func invalidArg(i int, fn, msg string) string {
	return fmt.Sprintf("invalid argument #%d to '%s' (%s)", i, fn, msg)
}

type Args struct {
	args []any
	name string
	co   *Coroutine
	pos  int
}

type (
	Ret  any
	Rets []any
)

func getArg[T any](args *Args, optionalValue []T, tx ...string) T {
	args.pos++
	if args.pos > len(args.args) {
		if len(optionalValue) == 0 {
			panic(invalidNumArgs(args.name, args.pos, tx...))
		}
		return optionalValue[0]
	}

	possibleArg := args.args[args.pos-1]

	arg, ok := possibleArg.(T)
	if !ok {
		panic(invalidArgType(args.pos, args.name, typeOf(arg), typeOf(possibleArg)))
	}
	return arg
}

func (a *Args) CheckNextArg(tx ...string) {
	if a.pos >= len(a.args) {
		panic(invalidNumArgs(a.name, a.pos, tx...))
	}
}

func (a *Args) GetNumber(optionalValue ...float64) float64 {
	return getArg(a, optionalValue, "number")
}

func (a *Args) GetString(optionalValue ...string) string {
	return getArg(a, optionalValue, "string")
}

func (a *Args) GetBool(optionalValue ...bool) bool {
	return getArg(a, optionalValue, "boolean")
}

func (a *Args) GetTable(optionalValue ...*Table) *Table {
	return getArg(a, optionalValue, "table")
}

func (a *Args) GetFunction(optionalValue ...*Function) *Function {
	return getArg(a, optionalValue, "function")
}

func (a *Args) GetCoroutine(optionalValue ...*Coroutine) *Coroutine {
	return getArg(a, optionalValue, "thread")
}

func (a *Args) GetBuffer(optionalValue ...*Buffer) *Buffer {
	return getArg(a, optionalValue, "buffer")
}

func (a *Args) GetVector(optionalValue ...Vector) Vector {
	return getArg(a, optionalValue, "vector")
}

func (a *Args) GetAny(optionalValue ...any) (arg any) {
	a.pos++
	if a.pos > len(a.args) {
		if len(optionalValue) == 0 {
			panic(invalidNumArgs(a.name, a.pos))
		}
		return optionalValue[0]
	}

	return a.args[a.pos-1]
}

// Reflection don't scale
func MakeFn(name string, fn func(args Args) Rets) [2]any {
	fn2 := Function(func(co *Coroutine, vargs ...any) []any {
		return fn(Args{vargs, name, co, 0})
	})
	return [2]any{name, &fn2}
}

func MakeFn1(name string, fn func(args Args) Ret) [2]any {
	fn2 := Function(func(co *Coroutine, vargs ...any) []any {
		return []any{fn(Args{vargs, name, co, 0})}
	})
	return [2]any{name, &fn2}
}

func MakeFn0(name string, fn func(args Args)) [2]any {
	fn2 := Function(func(co *Coroutine, vargs ...any) []any {
		fn(Args{vargs, name, co, 0})
		return []any{}
	})
	return [2]any{name, &fn2}
}
