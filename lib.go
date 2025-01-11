package main

import "fmt"

func invalidNumArgs(fn string, nx int, tx ...string) error {
	if len(tx) > 0 {
		return fmt.Errorf("missing argument #%d to '%s' (%s expected)", nx, fn, tx[0])
	}
	return fmt.Errorf("missing argument #%d to '%s'", nx, fn)
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

type GotArg[T any] struct {
	arg T
}

func getArg[T any](a *Args, optionalValue []T, tx ...string) GotArg[T] {
	a.pos++
	if a.pos > len(a.args) {
		if len(optionalValue) == 0 {
			a.co.Error(invalidNumArgs(a.name, a.pos, tx...))
			return GotArg[T]{}
		}
		return GotArg[T]{arg: optionalValue[0]}
	}

	possibleArg := a.args[a.pos-1]

	arg, ok := possibleArg.(T)
	if !ok {
		a.co.Error(invalidArgType(a.pos, a.name, typeOf(arg), typeOf(possibleArg)))
		return GotArg[T]{}
	}
	return GotArg[T]{arg: arg}
}

func (a *Args) CheckNextArg() {
	if a.pos >= len(a.args) {
		a.co.Error(invalidNumArgs(a.name, a.pos+1))
	}
}

func (a *Args) GetNumber(optionalValue ...float64) float64 {
	return getArg(a, optionalValue, "number").arg
}

func (a *Args) GetString(optionalValue ...string) string {
	return getArg(a, optionalValue, "string").arg
}

func (a *Args) GetBool(optionalValue ...bool) bool {
	return getArg(a, optionalValue, "boolean").arg
}

func (a *Args) GetTable(optionalValue ...*Table) *Table {
	return getArg(a, optionalValue, "table").arg
}

func (a *Args) GetFunction(optionalValue ...Function) Function {
	return getArg(a, optionalValue, "function").arg
}

func (a *Args) GetCoroutine(optionalValue ...*Coroutine) *Coroutine {
	return getArg(a, optionalValue, "thread").arg
}

func (a *Args) GetBuffer(optionalValue ...*Buffer) *Buffer {
	return getArg(a, optionalValue, "buffer").arg
}

func (a *Args) GetVector(optionalValue ...Vector) Vector {
	return getArg(a, optionalValue, "vector").arg
}

func (a *Args) GetAny(optionalValue ...any) (arg any) {
	a.pos++
	if a.pos > len(a.args) {
		if len(optionalValue) == 0 {
			a.co.Error(invalidNumArgs(a.name, a.pos))
			return nil
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
