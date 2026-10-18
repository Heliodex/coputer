package std

import (
	"fmt"

	. "github.com/Heliodex/coputer/litecode/types"
)

func invalidNumArgsExpected(f string, nx int, tx string) error {
	return fmt.Errorf("missing argument #%d to '%s' (%s expected)", nx, f, tx)
}

func invalidNumArgs(f string, nx int) error {
	return fmt.Errorf("missing argument #%d to '%s'", nx, f)
}

func invalidArgType(i int, fn, tx, tg string) error {
	return fmt.Errorf("invalid argument #%d to '%s' (%s expected, got %s)", i, fn, tx, tg)
}

// Args represents the arguments passed to a user-defined native function.
//
// A number of helper functions are provided to extract arguments from the list. If these functions fail to extract the argument, the coroutine yields an invalid/missing argument error.
type Args struct {
	// Co is the coroutine that the function is running.
	Co *Coroutine
	// List is the list of all arguments passed to the function.
	List []Val
	Name string
	pos  int
}

// Rets returns a reusable buffer containing vals, so a native function can return a small result list without allocating. The caller must consume it before the next native call.
func (a *Args) Rets(vals ...Val) []Val { return a.Co.Rets(vals...) }

// nextArg advances to the next argument, reporting whether one exists.
func (a *Args) nextArg() (Val, bool) {
	a.pos++
	if a.pos > len(a.List) {
		return NilVal, false
	}
	return a.List[a.pos-1], true
}

// CheckNextArg ensures that there is at least one more argument to be read. If there isn't, the coroutine yields a missing argument error.
func (a *Args) CheckNextArg() {
	if a.pos >= len(a.List) {
		a.Co.Error(invalidNumArgs(a.Name, a.pos+1))
	}
}

// GetNumber returns the next argument as a float64 number value. An optional value can be passed if the argument is not required.
func (a *Args) GetNumber(optV ...float64) float64 {
	v, ok := a.nextArg()
	if !ok {
		if len(optV) == 0 {
			a.Co.Error(invalidNumArgsExpected(a.Name, a.pos, "number"))
		}
		return optV[0]
	}

	f, ok := AsNum(v)
	if !ok {
		a.Co.Error(invalidArgType(a.pos, a.Name, "number", TypeOf(v)))
	}
	return f
}

// GetString returns the next argument as a string value. An optional value can be passed if the argument is not required.
func (a *Args) GetString(optV ...string) string {
	v, ok := a.nextArg()
	if !ok {
		if len(optV) == 0 {
			a.Co.Error(invalidNumArgsExpected(a.Name, a.pos, "string"))
		}
		return optV[0]
	}

	s, ok := AsString(v)
	if !ok {
		a.Co.Error(invalidArgType(a.pos, a.Name, "string", TypeOf(v)))
	}
	return s
}

// GetBool returns the next argument as a boolean value. An optional value can be passed if the argument is not required.
func (a *Args) GetBool(optV ...bool) bool {
	v, ok := a.nextArg()
	if !ok {
		if len(optV) == 0 {
			a.Co.Error(invalidNumArgsExpected(a.Name, a.pos, "boolean"))
		}
		return optV[0]
	}

	b, ok := AsBool(v)
	if !ok {
		a.Co.Error(invalidArgType(a.pos, a.Name, "boolean", TypeOf(v)))
	}
	return b
}

// GetTable returns the next argument as a table value. An optional value can be passed if the argument is not required.
func (a *Args) GetTable(optV ...*Table) *Table {
	v, ok := a.nextArg()
	if !ok {
		if len(optV) == 0 {
			a.Co.Error(invalidNumArgsExpected(a.Name, a.pos, "table"))
		}
		return optV[0]
	}

	t, ok := AsTable(v)
	if !ok {
		a.Co.Error(invalidArgType(a.pos, a.Name, "table", TypeOf(v)))
	}
	return t
}

// GetFunction returns the next argument as a function value. An optional value can be passed if the argument is not required.
func (a *Args) GetFunction(optV ...Function) Function {
	v, ok := a.nextArg()
	if !ok {
		if len(optV) == 0 {
			a.Co.Error(invalidNumArgsExpected(a.Name, a.pos, "function"))
		}
		return optV[0]
	}

	f, ok := AsFunction(v)
	if !ok {
		a.Co.Error(invalidArgType(a.pos, a.Name, "function", TypeOf(v)))
	}
	return f
}

// GetCoroutine returns the next argument as a coroutine value. An optional value can be passed if the argument is not required.
func (a *Args) GetCoroutine(optV ...*Coroutine) *Coroutine {
	v, ok := a.nextArg()
	if !ok {
		if len(optV) == 0 {
			a.Co.Error(invalidNumArgsExpected(a.Name, a.pos, "thread"))
		}
		return optV[0]
	}

	c, ok := AsCoroutine(v)
	if !ok {
		a.Co.Error(invalidArgType(a.pos, a.Name, "thread", TypeOf(v)))
	}
	return c
}

// GetBuffer returns the next argument as a buffer value. An optional value can be passed if the argument is not required.
func (a *Args) GetBuffer(optV ...*Buffer) *Buffer {
	v, ok := a.nextArg()
	if !ok {
		if len(optV) == 0 {
			a.Co.Error(invalidNumArgsExpected(a.Name, a.pos, "buffer"))
		}
		return optV[0]
	}

	b, ok := AsBuffer(v)
	if !ok {
		a.Co.Error(invalidArgType(a.pos, a.Name, "buffer", TypeOf(v)))
	}
	return b
}

// GetVector returns the next argument as a vector value. An optional value can be passed if the argument is not required.
func (a *Args) GetVector(optV ...Vector) Vector {
	v, ok := a.nextArg()
	if !ok {
		if len(optV) == 0 {
			a.Co.Error(invalidNumArgsExpected(a.Name, a.pos, "vector"))
		}
		return optV[0]
	}

	x, ok := AsVector(v)
	if !ok {
		a.Co.Error(invalidArgType(a.pos, a.Name, "vector", TypeOf(v)))
	}
	return x
}

// GetAny returns the next argument.
func (a *Args) GetAny(optV ...Val) (arg Val) {
	a.pos++
	if a.pos > len(a.List) {
		if len(optV) == 0 {
			a.Co.Error(invalidNumArgs(a.Name, a.pos))
		}
		return optV[0]
	}

	return a.List[a.pos-1]
}

// NewLib creates a new library with a given table of functions and other values, such as constants. Functions can be created using MakeFn.
func NewLib(functions []Function, other ...map[string]Val) *Table {
	// remember, no duplicates
	hash := make(map[Val]Val, len(functions)+len(other))
	for _, f := range functions {
		hash[Str(f.Name)] = FnVal(f)
	}
	if len(other) > 0 {
		for k, v := range other[0] {
			hash[Str(k)] = v
		}
	}

	return &Table{
		Readonly: true,
		Hash:     hash,
	}
}

// MakeFn creates a new function with a given name and body. Functions created by MakeFn can be added to a library using NewLib.
func MakeFn(name string, f func(Args) ([]Val, error)) Function {
	return fn(name, func(co *Coroutine, vargs ...Val) ([]Val, error) {
		return f(Args{
			Co:   co,
			List: vargs,
			Name: name,
		})
	})
}
