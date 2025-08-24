package ast

import "strings"

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
	containsSquote := strings.ContainsRune(str, '\'')
	containsDquote := strings.ContainsRune(str, '"')

	if containsNewline {
		// [[string]]
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

	if !containsSquote {
		// "string"
		return "'" + str + "'"
	}

	if !containsDquote {
		// 'string'
		return "\"" + str + "\""
	}

	// string contains both single and double quotes
	squoteCount := strings.Count(str, "'")
	dquoteCount := strings.Count(str, "\"")

	// prefer double quotes in cases which each are equal
	if dquoteCount <= squoteCount {
		replaced := strings.ReplaceAll(str, "\"", "\\\"")
		return "\"" + replaced + "\""
	}
	replaced := strings.ReplaceAll(str, "'", "\\'")
	return "'" + replaced + "'"
}
