package litecode

import (
	"errors"
	"fmt"
	"strings"
)

func string_posrelat(pos, l int) int {
	// relative string position: negative means back from end
	if pos < 0 {
		pos += l + 1
	}
	if pos >= 0 {
		return pos
	}
	return 0
}

func string_byte(args Args) (bytes Rets, err error) {
	s := args.GetString()
	l := len(s)

	i := args.GetNumber(1)
	posi := string_posrelat(int(i), l)
	j := args.GetNumber(float64(posi))
	pose := string_posrelat(int(j), l)

	if posi <= 0 {
		posi = 1
	}
	if pose > l {
		pose = l
	}

	n := pose - posi + 1
	if n < 0 {
		return // empty interval; return no values
	} else if posi+n <= pose { // overflow?
		return nil, errors.New("string slice too long")
	}

	bytes = make(Rets, n)
	chars := []byte(s)
	for b := range bytes {
		bytes[b] = float64(chars[posi+b-1])
	}
	return
}

func string_char(args Args) (r Rets, err error) {
	l := len(args.List)

	bytes := make([]byte, l)
	for i := range bytes {
		a := args.GetNumber()
		ba := byte(a)
		if float64(ba) != a {
			return nil, fmt.Errorf("invalid argument #%d to 'char' (invalid value)", i+1)
		}

		bytes[i] = ba
	}

	return Rets{string(bytes)}, nil
}

// matching yeyyy

const (
	cap_unfinished = -1
	cap_position   = -2
	max_captures   = 32
	l_esc          = byte('%')
	specials       = "^$*+?.([%-"
	maxcalls       = 200
)

func tolower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 'a' - 'A'
	}
	return c
}

func isalpha(c byte) bool  { return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') }
func iscntrl(c byte) bool  { return c < ' ' || c == 127 }
func isdigit(c byte) bool  { return c >= '0' && c <= '9' }
func isgraph(c byte) bool  { return c > ' ' && c <= '~' }
func islower(c byte) bool  { return c >= 'a' && c <= 'z' }
func ispunct(c byte) bool  { return isgraph(c) && !isalnum(c) }
func isspace(c byte) bool  { return c == ' ' || (c >= '\t' && c <= '\r') }
func isupper(c byte) bool  { return c >= 'A' && c <= 'Z' }
func isalnum(c byte) bool  { return isalpha(c) || isdigit(c) }
func isxdigit(c byte) bool { return isdigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') }

func string_find(args Args) (r Rets, err error) {
	s, p := args.GetString(), args.GetString()
	i := int(args.GetNumber(1))
	plain := args.GetBool(false)

	return str_find_aux(s, p, i, plain, true)
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

func scanformat(strfrmt string) (byte, string, int, error) {
	const flags = "-+ #0"

	var p int
	for strings.ContainsRune(flags, rune(strfrmt[p])) {
		p++ // skip flags
	}
	if p > len(flags) {
		return 0, "", 0, fmt.Errorf("invalid format (repeated flags)")
	} else if isdigit(strfrmt[p]) {
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
		return 0, "", 0, fmt.Errorf("invalid format (width or precision too long)")
	}

	return strfrmt[p], strfrmt[:p], p, nil
}

func format(form string, formatIndicator byte, sub any) string {
	f := fmt.Sprintf("%%%s%c", form, formatIndicator) // wow, real meta
	return fmt.Sprintf(f, sub)
}

func fmtstring(strfrmt string, args *Args) (string, error) {
	b := strings.Builder{}

	for i, sfl := 0, len(strfrmt); i < sfl; {
		if strfrmt[i] != l_esc {
			b.WriteByte(strfrmt[i])
			i++
			continue
		} else if i++; strfrmt[i] == l_esc {
			b.WriteByte(l_esc) // %%
			i++
			continue
		} else if strfrmt[i] == '*' {
			a := args.GetAny()
			b.WriteString(ToString(a))
			i++
			continue
		}

		// format item
		args.CheckNextArg()
		formatIndicator, form, p, err := scanformat(strfrmt[i:])
		if err != nil {
			return "", err
		}
		i += p + 1

		switch formatIndicator {
		case 'c':
			n := args.GetNumber()

			b.WriteByte(byte(n))
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
			addquoted(args, &b)
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
			return "", errors.New("%* does not take a form")
		default: // also treat cases `pnLlh'
			return "", fmt.Errorf("invalid option '%%%c' to 'format'", formatIndicator)
		}
	}

	return b.String(), nil
}

func string_format(args Args) (r Rets, err error) {
	strfrmt := args.GetString()

	res, err := fmtstring(strfrmt, &args)
	if err != nil {
		return
	}
	return Rets{res}, nil
}

// func string_gmatch(args Args) (r Rets, err error) {
// 	panic("not implemented")
// }

// func string_gsub(args Args) (r Rets, err error) {
// 	panic("not implemented")
// }

func string_len(args Args) (r Rets, err error) {
	s := args.GetString()

	return Rets{float64(len(s))}, nil
}

func string_lower(args Args) (r Rets, err error) {
	s := args.GetString()

	return Rets{strings.ToLower(s)}, nil
}

// func string_match(args Args) (r Rets, err error) {
// 	panic("not implemented")
// }

func string_rep(args Args) (r Rets, err error) {
	s := args.GetString()
	n := args.GetNumber()

	return Rets{strings.Repeat(s, max(int(n), 0))}, nil
}

func string_reverse(args Args) (r Rets, err error) {
	s := args.GetString()

	rs := []byte(s) // []rune(s)
	for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
		rs[i], rs[j] = rs[j], rs[i]
	}
	return Rets{string(rs)}, nil
}

func string_split(args Args) (r Rets, err error) {
	s := args.GetString()
	separator := args.GetString(",")

	// can't copy (or copy()) []string to []any
	split := strings.Split(s, separator)
	a := make([]any, len(split))
	for i, v := range split {
		a[i] = v
	}

	return Rets{&Table{
		Array: a,
	}}, nil
}

func string_sub(args Args) (r Rets, err error) {
	s := args.GetString()
	i := args.GetNumber(1)
	j := args.GetNumber(-1)

	l := len(s)
	start := string_posrelat(int(i), l)
	end := string_posrelat(int(j), l)
	if start < 1 {
		start = 1
	}
	if end > l {
		end = l
	}

	if end < start {
		return Rets{""}, nil
	}
	return Rets{s[start-1 : end]}, nil
}

func string_upper(args Args) (r Rets, err error) {
	s := args.GetString()

	return Rets{strings.ToUpper(s)}, nil
}

var libstring = NewTable([][2]any{
	MakeFn("byte", string_byte),
	MakeFn("char", string_char),
	MakeFn("find", string_find),
	MakeFn("format", string_format),
	// MakeFn("gmatch", string_gmatch),
	// MakeFn("gsub", string_gsub),
	MakeFn("len", string_len),
	MakeFn("lower", string_lower),
	// MakeFn("match", string_match),
	// buffer should be used instead of pack/packsize
	MakeFn("rep", string_rep),
	MakeFn("reverse", string_reverse),
	MakeFn("split", string_split),
	MakeFn("sub", string_sub),
	// buffer should be used instead of unpack
	MakeFn("upper", string_upper),
})
