package main

import (
	"fmt"
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

func addquoted(args *Args, b *strings.Builder) {
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

func scanformat(strfrmt string) (byte, string, int) {
	p := 0
	for strings.ContainsRune(FLAGS, rune(strfrmt[p])) {
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

	return strfrmt[p], strfrmt[:p], p
}

func format(form string, formatIndicator byte, sub any) string {
	f := fmt.Sprintf("%%%s%c", form, formatIndicator) // wow, real meta
	return fmt.Sprintf(f, sub)
}

func string_format(args Args) Ret {
	strfrmt := args.GetString()

	b := strings.Builder{}

	for i, sfl := 0, len(strfrmt); i < sfl; {
		if strfrmt[i] != L_ESC {
			b.WriteByte(strfrmt[i])
			i++
			continue
		} else if i++; strfrmt[i] == L_ESC {
			b.WriteByte(L_ESC) // %%
			i++
			continue
		} else if strfrmt[i] == '*' {
			a := args.GetAny()
			b.WriteString(fmt.Sprintf("%v", a)) // we'd really do tostring() later
			i++
			continue
		}

		// format item
		args.CheckNextArg()
		formatIndicator, form, p := scanformat(strfrmt[i:])
		i += p + 1

		switch formatIndicator {
		case 'c':
			n := args.GetNumber()
			b.WriteString(fmt.Sprintf("%c", byte(n))) // TODO: %?
		case 'd', 'i':
			n := args.GetNumber()

			b.WriteString(format(form, 'd', int(n)))
		case 'o', 'u', 'x', 'X':
			n := args.GetNumber()

			var v uint64
			if n < 0 {
				v = uint64(int64(n))
			} else {
				v = uint64(n)
			}

			if formatIndicator == 'u' {
				formatIndicator = 'd'
			}
			b.WriteString(format(form, formatIndicator, v))
		case 'e', 'E', 'f', 'g', 'G':
			n := args.GetNumber()

			format := fmt.Sprintf("%%%s%c", form, formatIndicator)
			b.WriteString(fmt.Sprintf(format, n)) // TODO: %?
		case 'q':
			addquoted(&args, &b)
			continue // skip adding the string at the end
		case 's':
			s := args.GetString()
			// no precision and string is too long to be formatted, or no format necessary to begin with
			if len(form) <= 1 || (!strings.ContainsRune(form, '.') && len(s) > 100) {
				b.WriteString(s)
				break
			}

			b.WriteString(format(form, 's', s))
		case '*':
			//  %* is parsed above, so if we got here we must have a %...*
			panic("%* does not take a form")
		default: // also treat cases `pnLlh'
			panic(fmt.Sprintf("invalid option '%%%c' to 'format'", formatIndicator))
		}
	}

	return b.String()
}

var libstring = NewTable([][2]any{
	MakeFn("byte", string_byte),
	MakeFn1("char", string_char),
	MakeFn("find", string_find),
	MakeFn1("format", string_format),
})
