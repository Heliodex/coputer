package ast

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
