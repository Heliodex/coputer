package vm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Heliodex/coputer/litecode/types"
)

func string_posrelat(pos, l int) int {
	// relative string position: negative means back from end
	if pos < 0 {
		pos += l + 1
	}
	return max(0, pos)
}

func string_byte(args Args) (bytes []types.Val, err error) {
	s := args.GetString()
	l := len(s)

	i := args.GetNumber(1)
	posi := string_posrelat(int(i), l)
	j := args.GetNumber(float64(posi))
	pose := string_posrelat(int(j), l)

	posi, pose = max(posi, 1), min(pose, l)

	n := pose - posi + 1
	if n < 0 {
		return // empty interval; return no values
	}
	if posi+n <= pose { // overflow?
		return nil, errors.New("string slice too long")
	}

	bytes = make([]types.Val, n)
	for b := range bytes {
		bytes[b] = float64(s[posi+b-1])
	}
	return
}

func string_char(args Args) (r []types.Val, err error) {
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

	return []types.Val{string(bytes)}, nil
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

// stole from ctype
func isalpha(c byte) bool  { return c|32-'a' < 26 }
func iscntrl(c byte) bool  { return c < ' ' || c == 127 }
func isdigit(c byte) bool  { return c-'0' < 10 }
func isgraph(c byte) bool  { return c-'!' < 94 }
func islower(c byte) bool  { return c-'a' < 26 }
func isspace(c byte) bool  { return c == ' ' || c-'\t' < 5 }
func isupper(c byte) bool  { return c-'A' < 26 }
func isalnum(c byte) bool  { return isalpha(c) || isdigit(c) }
func ispunct(c byte) bool  { return isgraph(c) && !isalnum(c) }
func isxdigit(c byte) bool { return isdigit(c) || c|32-'a' < 6 }

func string_find(args Args) (r []types.Val, err error) {
	s, p := args.GetString(), args.GetString()
	i := int(args.GetNumber(1))
	plain := args.GetBool(false)

	return stringFindAux(s, p, i, plain, true)
}

func addquoted(args Args, b *strings.Builder) {
	s := args.GetString()

	b.WriteByte('"')
	for i, l := uint(0), len(s); l > 0; l-- {
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
		return 0, "", 0, errors.New("invalid format (repeated flags)")
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
		return 0, "", 0, errors.New("invalid format (width or precision too long)")
	}

	return strfrmt[p], strfrmt[:p], p, nil
}

func format(form string, formatIndicator byte, sub any) string {
	f := fmt.Sprintf("%%%s%c", form, formatIndicator) // wow, real meta
	return fmt.Sprintf(f, sub)
}

func formatItem(args *Args, formatIndicator byte, form string, b *strings.Builder) error {
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
		addquoted(*args, b)
	case 's':
		s := args.GetString()
		// no precision and string is too long to be formatted, or no format necessary to begin with
		if len(form) <= 1 || !strings.ContainsRune(form, '.') && len(s) > 100 {
			b.WriteString(s)
			break
		}

		b.WriteString(format(form, 's', s))
	case '*':
		//  %* is parsed above, so if we got here we must have a %...*
		return errors.New("%* does not take a form")
	default: // also treat cases `pnLlh'
		return fmt.Errorf("invalid option '%%%c' to 'format'", formatIndicator)
	}
	return nil
}

func fmtstring(strfrmt string, args Args) (string, error) {
	b := strings.Builder{}

	for i, sfl := 0, len(strfrmt); i < sfl; {
		if strfrmt[i] != l_esc {
			b.WriteByte(strfrmt[i])
			i++
			continue
		}

		i++
		if strfrmt[i] == l_esc {
			b.WriteByte(l_esc) // %%
			i++
			continue
		}
		if strfrmt[i] == '*' {
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

		if err := formatItem(&args, formatIndicator, form, &b); err != nil {
			return "", err
		}
	}

	return b.String(), nil
}

func string_format(args Args) (r []types.Val, err error) {
	strfrmt := args.GetString()

	res, err := fmtstring(strfrmt, args)
	if err != nil {
		return
	}
	return []types.Val{res}, nil
}

func string_gmatch(args Args) (r []types.Val, err error) {
	s, p := args.GetString(), args.GetString()
	ls := len(s)

	var start int
	gmatch := func(args Args) (r []types.Val, err error) {
		for caps := (&captures{}); start <= ls; start++ {
			caps.level = 0

			end, err := matchPos(s, p, start, 0, caps)
			if err != nil {
				return nil, err
			}
			if end == -1 {
				continue
			}

			// we got a match
			r, err = pushCaptures(caps, s, start, end, false)
			start = max(start+1, end)
			return r, err
		}

		return
	}

	return []types.Val{MakeFn("gmatch", gmatch)}, nil
}

func add_s(caps *captures, b *strings.Builder, s string, si, ei int, news string) error {
	for i := 0; i < len(news); i++ {
		if news[i] != l_esc {
			b.WriteByte(news[i])
			continue
		}

		i++ // skip ESC
		if !isdigit(news[i]) {
			if news[i] != l_esc {
				return fmt.Errorf("invalid use of '%c' in replacement string", l_esc)
			}
			b.WriteByte(l_esc) // always %?
		} else if news[i] == '0' {
			b.WriteString(s[si:ei])
		} else {
			r, err := pushCapture(caps, s, si, ei, int(news[i]-'1'))
			if err != nil {
				return err
			}

			b.WriteString(ToString(r)) // add capture to accumulated result
		}
	}

	return nil
}

func add_value(caps *captures, b *strings.Builder, co *types.Coroutine, s string, si, ei int, next types.Val) (err error) {
	var val types.Val

	switch n := next.(type) {
	case string:
		return add_s(caps, b, s, si, ei, n)
	case types.Function:
		rets, err := (*n.Run)(co, s[si:ei])
		if err != nil {
			return err
		}
		if len(rets) != 0 {
			val = rets[0]
		}
	case *Table:
		r, err := pushCapture(caps, s, si, ei, 0) // at least pretty sure this'll always return a string
		if err != nil {
			return err
		}

		// fmt.Println("typeofr", TypeOf(r))

		val = n.GetHash(ToString(r)) // we'll tostring here to be safe
	}

	if !truthy(val) { // nil or false?
		b.WriteString(s[si:ei]) // keep original text
		return
	}
	if _, ok := val.(string); !ok {
		return fmt.Errorf("invalid replacement value (a %s)", TypeOf(val))
	}

	// add result to accumulator?
	b.WriteString(ToString(val))
	return
}

func string_gsub(args Args) (r []types.Val, err error) {
	src := args.GetString()
	p := args.GetString()
	next := args.GetAny()

	switch next.(type) {
	case string, types.Function, *Table:
	default:
		return nil, fmt.Errorf("invalid argument #3 to 'gsub' (string/function/table expected, got %s)", TypeOf(next))
	}

	srcl := len(src)
	max_s := args.GetNumber(float64(srcl + 1))

	m, pis, sis := int(max_s), 0, 0
	var anchor bool
	if len(p) > 0 && p[0] == '^' {
		anchor = true
		pis = 1
	}

	var n int // number of replacements
	caps := &captures{}
	b := strings.Builder{}

	for n < m {
		caps.level = 0
		e, err := matchPos(src, p, sis, pis, caps)
		if err != nil {
			return nil, err
		}
		if e != -1 {
			n++
			add_value(caps, &b, args.Co, src, sis, e, next)
		}

		if e != -1 && e > sis { // non-empty match?
			sis = e // skip it
		} else if sis < srcl {
			b.WriteByte(src[sis])
			sis++
		} else {
			break
		}

		if anchor {
			break
		}
	}

	if sis < srcl {
		b.WriteString(src[sis:])
	}

	return []types.Val{b.String(), float64(n)}, nil
}

func string_len(args Args) (r []types.Val, err error) {
	s := args.GetString()

	return []types.Val{float64(len(s))}, nil
}

func string_lower(args Args) (r []types.Val, err error) {
	s := args.GetString()

	return []types.Val{strings.ToLower(s)}, nil
}

func string_match(args Args) (r []types.Val, err error) {
	s, p := args.GetString(), args.GetString()
	i := int(args.GetNumber(1))

	return stringFindAux(s, p, i, false, false)
}

func string_rep(args Args) (r []types.Val, err error) {
	s := args.GetString()
	n := args.GetNumber()

	return []types.Val{strings.Repeat(s, max(int(n), 0))}, nil
}

func string_reverse(args Args) (r []types.Val, err error) {
	s := args.GetString()

	rs := []byte(s) // []rune(s)
	for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
		rs[i], rs[j] = rs[j], rs[i]
	}
	return []types.Val{string(rs)}, nil
}

func string_split(args Args) (r []types.Val, err error) {
	s := args.GetString()
	separator := args.GetString(",")

	split := strings.Split(s, separator)

	// can't copy (or copy()) []string to []types.Val
	a := make([]types.Val, len(split))
	for i, v := range split {
		a[i] = v
	}

	return []types.Val{&Table{List: a}}, nil
}

func string_sub(args Args) (r []types.Val, err error) {
	s := args.GetString()
	i, j := args.GetNumber(1), args.GetNumber(-1)

	l := len(s)
	start, end := string_posrelat(int(i), l), string_posrelat(int(j), l)
	start, end = max(start, 1), min(end, l)

	if end < start {
		return []types.Val{""}, nil
	}
	return []types.Val{s[start-1 : end]}, nil
}

func string_upper(args Args) (r []types.Val, err error) {
	s := args.GetString()

	return []types.Val{strings.ToUpper(s)}, nil
}

var libstring = NewLib([]types.Function{
	MakeFn("byte", string_byte),
	MakeFn("char", string_char),
	MakeFn("find", string_find),
	MakeFn("format", string_format),
	MakeFn("gmatch", string_gmatch),
	MakeFn("gsub", string_gsub),
	MakeFn("len", string_len),
	MakeFn("lower", string_lower),
	MakeFn("match", string_match),
	// buffer should be used instead of pack/packsize
	MakeFn("rep", string_rep),
	MakeFn("reverse", string_reverse),
	MakeFn("split", string_split),
	MakeFn("sub", string_sub),
	// buffer should be used instead of unpack
	MakeFn("upper", string_upper),
})
