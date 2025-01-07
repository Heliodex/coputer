package main

import (
	"fmt"
	"strings"
)

func string_posrelat(pos, len int) int {
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

	ui := string_posrelat(int(i), ls)
	uj := string_posrelat(int(j), ls)

	cap := uj - ui + 1
	if cap < 0 {
		return
	}

	bytes = make(Rets, cap)
	chars := []byte(s)
	for n := range bytes {
		bytes[n] = float64(chars[n+ui])
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

func string_find(args Args) Rets {
	s := args.GetString()
	p := args.GetString()
	i := args.GetNumber(1)
	// plain := args.GetBool(false)

	init := string_posrelat(int(i), len(s))
	pos := strings.Index(s[init:], p)
	if pos == -1 {
		return Rets{nil} // one nil
	}
	return Rets{
		float64(pos + init + 1),
		float64(pos + init + len(p)),
	}
}

var (
	L_ESC = byte('%')
	FLAGS = "-+ #0"
)

func isdigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func addquoted(args *Args, b *strings.Builder) {
	s := args.GetString()
	l := len(s)
	var i uint

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
	var p int
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

func fmtstring(strfrmt string, args Args) string {
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
			b.WriteString(fmt.Sprintf("%c", byte(n)))
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
			b.WriteString(fmt.Sprintf(format, n))
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

func string_format(args Args) Ret {
	strfrmt := args.GetString()

	return fmtstring(strfrmt, args)
}

func string_gmatch(args Args) Ret {
	panic("not implemented")
}

func string_gsub(args Args) Rets {
	panic("not implemented")
}

func string_len(args Args) Ret {
	s := args.GetString()

	return float64(len(s))
}

func string_lower(args Args) Ret {
	s := args.GetString()

	return strings.ToLower(s)
}

func string_match(args Args) Ret {
	panic("not implemented")
}

func string_rep(args Args) Ret {
	s := args.GetString()
	n := args.GetNumber()

	return strings.Repeat(s, max(int(n), 0))
}

func string_reverse(args Args) Ret {
	s := args.GetString()

	r := []byte(s) // []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}

	return string(r)
}

func string_split(args Args) Ret {
	s := args.GetString()
	separator := args.GetString(",")

	split := strings.Split(s, separator)
	a := make([]any, len(split))
	for i, v := range split {
		a[i] = v
	}

	return &Table{
		array: &a,
		asize: uint(len(a)),
	}
}

func string_sub(args Args) Ret {
	s := args.GetString()
	i := args.GetNumber(1)
	j := args.GetNumber(-1)

	l := len(s)
	ui := string_posrelat(int(i), l)
	uj := string_posrelat(int(j), l)

	if uj < ui {
		return ""
	}
	return s[ui:min(uj+1, l)]
}

func string_upper(args Args) Ret {
	s := args.GetString()

	return strings.ToUpper(s)
}

var libstring = NewTable([][2]any{
	MakeFn("byte", string_byte),
	MakeFn1("char", string_char),
	MakeFn("find", string_find),
	MakeFn1("format", string_format),
	MakeFn1("gmatch", string_gmatch),
	MakeFn("gsub", string_gsub),
	MakeFn1("len", string_len),
	MakeFn1("lower", string_lower),
	MakeFn1("match", string_match),
	// buffer should be used instead of pack/packsize
	MakeFn1("rep", string_rep),
	MakeFn1("reverse", string_reverse),
	MakeFn1("split", string_split),
	MakeFn1("sub", string_sub),
	// buffer should be used instead of unpack
	MakeFn1("upper", string_upper),
})
