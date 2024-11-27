package main

import "fmt"

func invalidNumArgs(fn string, nx int, tx string) string {
	return fmt.Sprintf("missing argument #%d to '%s' (%s expected)", nx, fn, luautype[tx])
}

func invalidArgType(i int, fn string, tx, tg string) string {
	return fmt.Sprintf("invalid argument #%d to '%s' (%s expected, got %s)", i, fn, luautype[tx], luautype[tg])
}

type Args struct {
	args []any
	name string
	pos  int
}

func getArg[T any](args *Args, optionalValue []T) T {
	var possibleArg any

	if args.pos >= len(args.args) {
		if len(optionalValue) == 0 {
			panic(invalidNumArgs(args.name, args.pos, typeOf(args.args[args.pos-1])))
		}
		possibleArg = optionalValue[0]
	} else {
		possibleArg = args.args[args.pos]
	}

	args.pos++
	arg, ok := possibleArg.(T)
	if !ok {
		panic(invalidArgType(args.pos, args.name, typeOf(arg), typeOf(possibleArg)))
	}
	return arg
}

func (a *Args) CheckNextArg() {
	if a.pos >= len(a.args) {
		panic(invalidNumArgs(a.name, a.pos, typeOf(a.args[a.pos-1])))
	}
}

func (a *Args) GetNumber(optionalValue ...float64) float64 {
	return getArg(a, optionalValue)
}

func (a *Args) GetString(optionalValue ...string) string {
	return getArg(a, optionalValue)
}

func (a *Args) GetTable(optionalValue ...*Table) *Table {
	return getArg(a, optionalValue)
}

func (args *Args) GetAny(optionalValue ...any) any {
	return getArg(args, optionalValue)
}

// Reflection don't scale
func MakeFn(name string, fn func(args *Args) []any) [2]any {
	fn2 := Function(func(vargs ...any) []any {
		return fn(&Args{
			args: vargs,
			name: name,
		})
	})

	return [2]any{name, &fn2}
}
