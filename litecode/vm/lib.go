package vm

import
(
	"fmt"
"github.com/Heliodex/coputer/litecode/types"
)

func invalidNumArgs(f string, nx int, tx ...string) error {
	if len(tx) > 0 {
		return fmt.Errorf("missing argument #%d to '%s' (%s expected)", nx, f, tx[0])
	}
	return fmt.Errorf("missing argument #%d to '%s'", nx, f)
}

func invalidArgType(i int, fn, tx, tg string) error {
	return fmt.Errorf("invalid argument #%d to '%s' (%s expected, got %s)", i, fn, luautype[tx], luautype[tg])
}

func invalidArg(i int, fn, msg string) error {
	return fmt.Errorf("invalid argument #%d to '%s' (%s)", i, fn, msg)
}

// Args represents the arguments passed to a user-defined native function.
//
// A number of helper functions are provided to extract arguments from the list. If these functions fail to extract the argument, the coroutine yields an invalid/missing argument error.
type Args struct {
	// Co is the coroutine that the function is running.
	Co *Coroutine
	// List is the list of all arguments passed to the function.
	List []types.Val
	name string
	pos  int
}

func getArg[T types.Val](a *Args, optV []T, tx string) (g T) {
	a.pos++
	if a.pos > len(a.List) {
		if len(optV) == 0 {
			a.Co.Error(invalidNumArgs(a.name, a.pos, tx))
			return
		}
		return optV[0]
	}

	possibleArg := a.List[a.pos-1]
	arg, ok := possibleArg.(T)
	if !ok {
		a.Co.Error(invalidArgType(a.pos, a.name, TypeOf(arg), TypeOf(possibleArg)))
		return
	}
	return arg
}

// CheckNextArg ensures that there is at least one more argument to be read. If there isn't, the coroutine yields a missing argument error.
func (a *Args) CheckNextArg() {
	if a.pos >= len(a.List) {
		a.Co.Error(invalidNumArgs(a.name, a.pos+1))
	}
}

// GetNumber returns the next argument as a float64 number value. An optional value can be passed if the argument is not required.
func (a *Args) GetNumber(optV ...float64) float64 {
	return getArg(a, optV, "number")
}

// GetString returns the next argument as a string value. An optional value can be passed if the argument is not required.
func (a *Args) GetString(optV ...string) string {
	return getArg(a, optV, "string")
}

// GetBool returns the next argument as a boolean value. An optional value can be passed if the argument is not required.
func (a *Args) GetBool(optV ...bool) bool {
	return getArg(a, optV, "boolean")
}

// GetTable returns the next argument as a table value. An optional value can be passed if the argument is not required.
func (a *Args) GetTable(optV ...*Table) *Table {
	return getArg(a, optV, "table")
}

// GetFunction returns the next argument as a function value. An optional value can be passed if the argument is not required.
func (a *Args) GetFunction(optV ...Function) Function {
	return getArg(a, optV, "function")
}

// GetCoroutine returns the next argument as a coroutine value. An optional value can be passed if the argument is not required.
func (a *Args) GetCoroutine(optV ...*Coroutine) *Coroutine {
	return getArg(a, optV, "thread")
}

// GetBuffer returns the next argument as a buffer value. An optional value can be passed if the argument is not required.
func (a *Args) GetBuffer(optV ...*types.Buffer) *types.Buffer {
	return getArg(a, optV, "buffer")
}

// GetVector returns the next argument as a vector value. An optional value can be passed if the argument is not required.
func (a *Args) GetVector(optV ...types.Vector) types.Vector {
	return getArg(a, optV, "vector")
}

// GetAny returns the next argument.
func (a *Args) GetAny(optV ...types.Val) (arg types.Val) {
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

// NewLib creates a new library with a given table of functions and other values, such as constants. Functions can be created using MakeFn.
func NewLib(functions []Function, other ...map[string]types.Val) *Table {
	// remember, no duplicates
	hash := make(types.ValMap, len(functions)+len(other))
	for _, f := range functions {
		hash[f.name] = f
	}
	if len(other) > 0 {
		for k, v := range other[0] {
			hash[k] = v
		}
	}

	return &Table{
		readonly: true,
		Hash:     hash,
	}
}

// MakeFn creates a new function with a given name and body. Functions created by MakeFn can be added to a library using NewLib.
func MakeFn(name string, f func(args Args) (r []types.Val, err error)) Function {
	return fn(name, func(co *Coroutine, vargs ...types.Val) (r []types.Val, err error) {
		return f(Args{
			Co:   co,
			List: vargs,
			name: name,
		})
	})
}
