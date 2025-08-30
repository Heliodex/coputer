package ast

import (
	"fmt"
	"strings"
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

// we're not gonna use `string`s here, to reserve those for only interpolated strings
func StringToSource(str string) string {
	containsNewline := strings.ContainsRune(str, '\n')
	containsDquote := strings.ContainsRune(str, '"')
	containsSquote := strings.ContainsRune(str, '\'')

	if containsNewline {
		// [[string]] or [=...[string]=...]
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
func NumberToSource(n float64) string {
	replaced := fmt.Sprintf("%g", n)
}
