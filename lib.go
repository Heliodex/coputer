package litecode

import "fmt"

func invalidNumArgs(fn string, nx int, tx ...string) error {
	if len(tx) > 0 {
		return fmt.Errorf("missing argument #%d to '%s' (%s expected)", nx, fn, tx[0])
	}
	return fmt.Errorf("missing argument #%d to '%s'", nx, fn)
}

func invalidArgType(i int, fn, tx, tg string) error {
	return fmt.Errorf("invalid argument #%d to '%s' (%s expected, got %s)", i, fn, luautype[tx], luautype[tg])
}

func invalidArg(i int, fn, msg string) error {
	return fmt.Errorf("invalid argument #%d to '%s' (%s)", i, fn, msg)
}

type Args struct {
	Co   *Coroutine
	List []any
	name string
	pos  int
}

type Rets []any

func getArg[T any](a *Args, optV []T, tx ...string) (g T) {
	a.pos++
	if a.pos > len(a.List) {
		if len(optV) == 0 {
			a.Co.Error(invalidNumArgs(a.name, a.pos, tx...))
			return
		}
		return optV[0]
	}

	possibleArg := a.List[a.pos-1]

	arg, ok := possibleArg.(T)
	if !ok {
		a.Co.Error(invalidArgType(a.pos, a.name, typeOf(arg), typeOf(possibleArg)))
		return
	}
	return arg
}

func (a *Args) CheckNextArg() {
	if a.pos >= len(a.List) {
		a.Co.Error(invalidNumArgs(a.name, a.pos+1))
	}
}

func (a *Args) GetNumber(optV ...float64) float64 {
	return getArg(a, optV, "number")
}

func (a *Args) GetString(optV ...string) string {
	return getArg(a, optV, "string")
}

func (a *Args) GetBool(optV ...bool) bool {
	return getArg(a, optV, "boolean")
}

func (a *Args) GetTable(optV ...*Table) *Table {
	return getArg(a, optV, "table")
}

func (a *Args) GetFunction(optV ...Function) Function {
	return getArg(a, optV, "function")
}

func (a *Args) GetCoroutine(optV ...*Coroutine) *Coroutine {
	return getArg(a, optV, "thread")
}

func (a *Args) GetBuffer(optV ...*Buffer) *Buffer {
	return getArg(a, optV, "buffer")
}

func (a *Args) GetVector(optV ...Vector) Vector {
	return getArg(a, optV, "vector")
}

func (a *Args) GetAny(optV ...any) (arg any) {
	a.pos++
	if a.pos > len(a.List) {
		if len(optV) == 0 {
			a.Co.Error(invalidNumArgs(a.name, a.pos))
			return
		}
		return optV[0]
	}

	return a.List[a.pos-1]
}

func MakeFn(name string, fn func(args Args) (r Rets, err error)) [2]any {
	return [2]any{name, Fn(func(co *Coroutine, vargs ...any) (r Rets, err error) {
		return fn(Args{Co: co, List: vargs, name: name})
	})}
}
