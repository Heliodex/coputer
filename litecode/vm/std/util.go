package std

import (
	"fmt"
	"strings"

	. "github.com/Heliodex/coputer/litecode/types"
)

func falsy(v Val) bool {
	return v == nil || v == false
}

func fn(name string, f func(*Coroutine, ...Val) ([]Val, error)) Function {
	return Function{
		Run:  &f,
		Name: name,
	}
}

// ToString returns a string representation of any value.
func ToString(a Val) string {
	switch v := a.(type) {
	case nil:
		return "nil"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return num2str(v)
	case Vector:
		// just 3-wide 4-now
		return fmt.Sprintf("%s, %s, %s", num2str(float64(v[0])), num2str(float64(v[1])), num2str(float64(v[2])))
	case string:
		return strings.ReplaceAll(v, "\n", "\r\n") // bruh
	}
	// panic("tostring bad type")
	return "userdata"
}

// TypeOf returns the underlying VM datatype of a value as a string.
// This does not return the Luau type, as type() does.
func TypeOf(v Val) string {
	if v == nil {
		return "nil"
	}

	switch v.(type) {
	case float64:
		return "number"
	case string:
		return "string"
	case bool:
		return "boolean"
	case *Table:
		return "table"
	case Function:
		return "function"
	case *Coroutine:
		return "thread"
	case *Buffer:
		return "buffer"
	case Vector:
		return "vector"
	}
	// panic("typeof bad type")
	return "userdata"
}
