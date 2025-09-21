package ast

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

func BinopToSource(name string) string {
	switch name {
	case "Add":
		return "+"
	case "Sub":
		return "-"
	case "Mul":
		return "*"
	case "Div":
		return "/"
	case "Mod":
		return "%"
	case "Pow":
		return "^"
	case "FloorDiv":
		return "//"
	case "CompareLt":
		return "<"
	case "CompareGt":
		return ">"
	case "CompareLe":
		return "<="
	case "CompareGe":
		return ">="
	case "CompareEq":
		return "=="
	case "CompareNe":
		return "~="
	case "And":
		return "and"
	case "Or":
		return "or"
	case "Concat":
		return ".."
	}

	panic("bad binop: " + name)
}

func UnopToSource(name string) string {
	switch name {
	case "Not":
		return "not "
	case "Minus":
		return "-"
	case "Len":
		return "#"
	}

	panic("bad unop: " + name)
}

var escapes = map[byte][]byte{
	'\a': []byte("\\a"),
	'\b': []byte("\\b"),
	'\f': []byte("\\f"),
	'\n': []byte("\\n"),
	'\r': []byte("\\r"),
	'\t': []byte("\\t"),
	'\v': []byte("\\v"),
	// idk about \z
}

// we're not gonna use `string`s here, to reserve those for only interpolated strings
func StringToSource(str string) string {
	containsNewline := strings.ContainsRune(str, '\n')
	containsOtherControl := strings.ContainsAny(str, "\a\b\f\r\t\v")
	containsDquote := strings.ContainsRune(str, '"')
	containsSquote := strings.ContainsRune(str, '\'')

	if containsNewline && !containsOtherControl {
		// [[string]] or [=...[string]=...]
		// these types of strings can't hold escapes
		for eqCount := 0; ; eqCount++ {
			eqs := strings.Repeat("=", eqCount)
			endDelimiter := "]" + eqs + "]"

			if strings.Contains(str, endDelimiter) {
				continue
			}

			startDelimiter := "[" + eqs + "["
			return startDelimiter + str + endDelimiter
		}
	}

	str = strings.ReplaceAll(str, "\\", "\\\\")
	str = strings.ReplaceAll(str, "\r", "\\r")
	str = strings.ReplaceAll(str, "\t", "\\t")

	// fmt.Println([]byte(str))

	// UTF-8 MANGLING!!!!

	// fmt.Println("StringToSource", str)

	sbs := []byte(str)
	var bs []byte
	for i := 0; i < len(sbs); i++ {
		b := sbs[i]
		// fmt.Println(str, "b", b, string(b))

		if esc, ok := escapes[b]; ok {
			bs = append(bs, esc...)
			continue
		}

		r := []rune(string(sbs[i:]))[0]
		if unicode.IsControl(r) {
			bs = append(bs, []byte(fmt.Sprintf("\\%d", r))...)
			i += len(string(r)) - 1
			continue
		}

		if r == unicode.ReplacementChar && unicode.IsControl(rune(b)) {
			// invalid utf-8 byte
			bs = append(bs, []byte(fmt.Sprintf("\\%d", b))...)
			continue
		}

		bs = append(bs, b)
	}
	str = string(bs)

	if !containsDquote {
		// 'string'
		return "\"" + str + "\""
	}

	if !containsSquote {
		// "string"
		return "'" + str + "'"
	}

	// string contains both single and double quotes

	// prefer double quotes in cases which each are equal
	if strings.Count(str, "\"") <= strings.Count(str, "'") {
		replaced := strings.ReplaceAll(str, "\"", "\\\"")
		return "\"" + replaced + "\""
	}
	replaced := strings.ReplaceAll(str, "'", "\\'")
	return "'" + replaced + "'"
}

// oh god
func NumberToSource(n Number) string {
	if math.IsInf(float64(n), 1) {
		return "math.huge" // luau's max number is 1e308, so 1e309 is inf
	}

	if math.IsInf(float64(n), -1) {
		return "-math.huge" // shouldn't ever happen because negative numbers aren't possible in the AST, but whatever
	}

	rep := fmt.Sprintf("%g", n)
	rep = strings.Replace(rep, "e+", "e", 1)

	for strings.Contains(rep, "e0") {
		rep = strings.Replace(rep, "e0", "e", 1)
	}

	for strings.Contains(rep, "e-0") {
		rep = strings.Replace(rep, "e-0", "e-", 1)
	}

	if strings.Contains(rep, "e-") || !strings.Contains(rep, "e") {
		return rep
	}

	// check if the exponent matches the number of decimal places, in which case we can just remove it
	eSplit := strings.Split(rep, "e")
	if len(eSplit) == 1 {
		return rep
	}
	dotSplit := strings.Split(eSplit[0], ".")
	if len(dotSplit) == 1 {
		return rep
	}

	start, decimal, exponent := dotSplit[0], dotSplit[1], eSplit[1]

	if exponent == fmt.Sprintf("%d", len(decimal)) {
		return start + decimal
	}
	return rep
}
