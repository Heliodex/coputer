package std

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/Heliodex/coputer/litecode/types"
)

func falsy(v Val) bool {
	if IsNil(v) {
		return true
	}
	b, ok := AsBool(v)
	return ok && !b
}

func fn(name string, f func(*Coroutine, ...Val) ([]Val, error)) Function {
	return Function{
		Run:  &f,
		Name: name,
	}
}

// ToString returns a string representation of any value.
func ToString(a Val) string {
	if f, ok := AsNum(a); ok {
		return num2str(f)
	}
	if i, ok := AsInt(a); ok {
		return strconv.Itoa(int(i))
	}
	if s, ok := AsString(a); ok {
		return strings.ReplaceAll(s, "\n", "\r\n") // bruh
	}
	if IsNil(a) {
		return "nil"
	}
	if b, ok := AsBool(a); ok {
		if b {
			return "true"
		}
		return "false"
	}
	if v, ok := AsVector(a); ok {
		// just 3-wide 4-now
		return fmt.Sprintf("%s, %s, %s", num2str(float64(v[0])), num2str(float64(v[1])), num2str(float64(v[2])))
	}
	if t, ok := AsTable(a); ok {
		return fmt.Sprintf("table: %v", *t)
	}
	if f, ok := AsFunction(a); ok {
		return fmt.Sprintf("function: %v", f)
	}
	if c, ok := AsCoroutine(a); ok {
		return fmt.Sprintf("thread: %v", *c)
	}
	if b, ok := AsBuffer(a); ok {
		return fmt.Sprintf("buffer: %v", *b)
	}
	// panic("tostring bad type")
	return "userdata"
}

// TypeOf returns the underlying VM datatype of a value as a string.
// This does not return the Luau type, as type() does.
func TypeOf(v Val) string {
	if IsNum(v) {
		return "number"
	}
	if IsString(v) {
		return "string"
	}
	if IsBool(v) {
		return "boolean"
	}
	if IsNil(v) {
		return "nil"
	}
	if IsTable(v) {
		return "table"
	}
	if IsFunction(v) {
		return "function"
	}
	if IsCoroutine(v) {
		return "thread"
	}
	if IsBuffer(v) {
		return "buffer"
	}
	if IsVector(v) {
		return "vector"
	}
	// panic("typeof bad type")
	return "userdata"
}
