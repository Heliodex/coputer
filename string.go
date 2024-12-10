package main

import (
	"fmt"
	"slices"
	"strings"
)

func posrelat(pos, len int) int {
	// relative string position: negative means back from end
	if pos < 0 {
		pos += len + 1
	}
	if pos >= 0 {
		return min(pos, len) - 1
	}
	return 0
}

func string_byte(args Args) (bytes Rets) {
	s := args.GetString()
	i := args.GetNumber(1)
	j := args.GetNumber(i)

	ls := len(s)

	ui := posrelat(int(i), ls)
	uj := posrelat(int(j), ls)

	cap := uj - ui + 1
	if cap < 0 {
		return
	}

	bytes = make(Rets, cap)
	chars := []byte(s)
	for i := range bytes {
		bytes[i] = chars[i+ui]
	}
	return
}

func string_char(args Args) Ret {
	l := len(args.args)

	bytes := make([]byte, l)
	for i := range bytes {
		a := args.GetNumber()
		ba := byte(a)
		if float64(ba) != a {
			panic(fmt.Sprintf("invalid argument #%d to 'char' (invalid value)", i+1))
		}

		bytes[i] = ba
	}

	return string(bytes)
}

var L_ESC = byte('%')

func string_find(args Args) Rets {
	s := args.GetString()
	p := args.GetString()
	i := args.GetNumber(1)
	// plain := args.GetBool(false)

	init := posrelat(int(i), len(s))
	pos := strings.Index(s[init:], p)
	if pos == -1 {
		return Rets{nil} // one nil
	}
	return Rets{pos + init + 1, pos + init + len(p)}
}

var FLAGS = "-+ #0"

func isdigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func addquoted(args *Args, b *strings.Builder, arg int) {
	s := args.GetString()
	l := len(s)
	i := 0

	b.WriteByte('"')
	for l > 0 {
		l--
		switch w := s[i]; w {
		case '"', '\\', '\n':
			b.WriteByte('\\')
			b.WriteByte(w)
		case '\r':
			b.WriteString("\\r") // TODO: line endings
		case 0:
			b.WriteString("\\000")
		default:
			b.WriteByte(w)
		}
		i++
	}
	b.WriteByte('"')
}

func scanformat(strfrmt string, form [32]byte, size *int) string {
	p := 0
	for strfrmt[p] != 0 && strings.ContainsRune(FLAGS, rune(strfrmt[p])) {
		p++ // skip flags
	}
	if p > len(FLAGS) {
		panic("invalid format (repeated flags)")
	}
	if isdigit(strfrmt[p]) {
		p++ // skip width
		if isdigit(strfrmt[p]) {
			p++ // (2 digits at most)
		}
	}
	if strfrmt[p] == '.' {
		p++
		if isdigit(strfrmt[p]) {
			p++ // skip precision
			if isdigit(strfrmt[p]) {
				p++ // (2 digits at most)
			}
		}
	}
	if isdigit(strfrmt[p]) {
		panic("invalid format (width or precision too long)")
	}

	form[0] = '%'
	*size = p + 1
	copy(form[:], strfrmt[:*size])
	form[len(form)-1] = 0 // implementation detailllllllllllllllllllllllllllllll
	return strfrmt[p:]
}

func addInt64Format(form [32]byte, formatIndicator byte, formatItemSize int) {
	// LUAU_ASSERT((formatItemSize + 3) <= 32)
	// LUAU_ASSERT(form[0] == '%')
	// LUAU_ASSERT(form[formatItemSize] != 0)
	// LUAU_ASSERT(form[formatItemSize+1] == 0)
	form[formatItemSize] = 'l'
	form[formatItemSize+1] = 'l'
	form[formatItemSize+2] = formatIndicator
	form[formatItemSize+3] = 0
}

func string_format(args Args) Ret {
	strfrmt := args.GetString()
	// top := len(args)

	sfl := len(strfrmt)
	arg := 1
	b := strings.Builder{}

	for i := 0; i < sfl; {
		if strfrmt[i] != L_ESC {
			b.WriteByte(strfrmt[i])
			i++
		} else if i += 1; strfrmt[i] == L_ESC {
			b.WriteByte(L_ESC) // %%
			i++
		} else if strfrmt[i] == '*' {
			i++
			a := args.GetAny()
			b.WriteString(fmt.Sprintf("%v", a)) // we'd really do tostring() later
		} else { // format item
			var form [32]byte         // to store the format (`%...')
			buff := strings.Builder{} // to store the formatted item
			args.CheckNextArg()
			formatItemSize := 0
			strfrmt = scanformat(strfrmt, form, &formatItemSize)
			formatIndicator := strfrmt[i]
			i++

			switch formatIndicator {
			case 'c':
				n := args.GetNumber()
				buff.WriteString(fmt.Sprintf("%c", byte(n))) // TODO: %?
			case 'd', 'i':
				addInt64Format(form, formatIndicator, formatItemSize)
				n := args.GetNumber()
				b.WriteString(fmt.Sprintf("%d", int(n))) // TODO: %?
			case 'o', 'u', 'x', 'X':
				argValue := args.GetNumber()
				addInt64Format(form, formatIndicator, formatItemSize)
				var v uint64
				if argValue < 0 {
					v = uint64(int64(argValue))
				} else {
					v = uint64(argValue)
				}

				if formatIndicator == 'u' {
					formatIndicator = 'd'
				}
				buff.WriteString(fmt.Sprintf("%"+string(formatIndicator), v)) // TODO: %?
			case 'e', 'E', 'f', 'g', 'G':
				n := args.GetNumber()
				buff.WriteString(fmt.Sprintf("%f", n)) // TODO: %?
			case 'q':
				addquoted(&args, &b, arg)
				continue // skip adding the string at the end
			case 's':
				s := args.GetString()
				// no precision and string is too long to be formatted, or no format necessary to begin with
				if form[2] == 0 || (!slices.Contains(form[:], '.') && len(s) > 100) {
					b.WriteString(s)
					continue // skip adding the string at the end
				}
				buff.WriteString(fmt.Sprintf("%v", s)) // TODO: %?
			case '*':
				//  %* is parsed above, so if we got here we must have a %...*
				panic("%* does not take a form")
			default: // also treat cases `pnLlh'
				panic(fmt.Sprintf("invalid option '%%%c' to 'format'", formatIndicator))
			}
			b.WriteString(buff.String())
		}
	}

	return b.String()
}

// func string_match(args Args) Rets {
// 	s := args.GetString()
// 	p := args.GetString()
// 	init := args.GetNumber(1)

// 	var start, end int
// 	pos := posrelat(int(init), len(s))
// }

var libstring = NewTable([][2]any{
	MakeFn("byte", string_byte),
	MakeFn1("char", string_char),
	MakeFn("find", string_find),
	MakeFn1("format", string_format),
})
