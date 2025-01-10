package main

import (
	"errors"
	"fmt"
)

func invalidNumArgs(fn string, nx int, tx ...string) error {
	res := fmt.Sprintf("missing argument #%d to '%s'", nx, fn)
	if len(tx) > 0 {
		res += fmt.Sprintf(" (%s expected)", tx[0])
	}
	return errors.New(res)
}

func invalidArgType(i int, fn string, tx, tg string) error {
	return fmt.Errorf("invalid argument #%d to '%s' (%s expected, got %s)", i, fn, luautype[tx], luautype[tg])
}

func invalidArg(i int, fn, msg string) error {
	return fmt.Errorf("invalid argument #%d to '%s' (%s)", i, fn, msg)
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

func getArg[T any](a *Args, optionalValue []T, tx ...string) T {
	a.pos++
	if a.pos > len(a.args) {
		if len(optionalValue) == 0 {
			panic(invalidNumArgs(a.name, a.pos, tx...))
		}
		return optionalValue[0]
	}

	possibleArg := a.args[a.pos-1]

	arg, ok := possibleArg.(T)
	if !ok {
		panic(invalidArgType(a.pos, a.name, typeOf(arg), typeOf(possibleArg)))
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

func (a *Args) GetFunction(optionalValue ...Function) Function {
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
	return [2]any{name, Fn(func(co *Coroutine, vargs ...any) (Rets, error) {
		return fn(Args{vargs, name, co, 0}), nil
	})}
}

func MakeFn1(name string, fn func(args Args) Ret) [2]any {
	return [2]any{name, Fn(func(co *Coroutine, vargs ...any) (Rets, error) {
		return Rets{fn(Args{vargs, name, co, 0})}, nil
	})}
}

func MakeFn0(name string, fn func(args Args)) [2]any {
	return [2]any{name, Fn(func(co *Coroutine, vargs ...any) (Rets, error) {
		fn(Args{vargs, name, co, 0})
		return Rets{}, nil
	})}
}

// ...and neither do panics
func MakeFnE(name string, fn func(args Args) (Rets, error)) [2]any {
	return [2]any{name, Fn(func(co *Coroutine, vargs ...any) (Rets, error) {
		return fn(Args{vargs, name, co, 0})
	})}
}

func MakeFn1E(name string, fn func(args Args) (Ret, error)) [2]any {
	return [2]any{name, Fn(func(co *Coroutine, vargs ...any) (Rets, error) {
		r, err := fn(Args{vargs, name, co, 0})
		if err != nil {
			return nil, err
		}
		return Rets{r}, nil
	})}
}

func MakeFn0E(name string, fn func(args Args) error) [2]any {
	return [2]any{name, Fn(func(co *Coroutine, vargs ...any) (Rets, error) {
		err := fn(Args{vargs, name, co, 0})
		return nil, err
	})}
}
