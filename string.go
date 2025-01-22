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

type strPos struct {
	s   string
	pos int
}

func (sp *strPos) String() string {
	return string(sp.s[:sp.pos])
}

func (sp *strPos) Len() int {
	return len(sp.s) - sp.pos
}

func (sp *strPos) End() bool {
	return sp.pos == len(sp.s)
}

func (sp *strPos) Inc(a ...int) (b byte) {
	if !sp.End() {
		b = sp.s[sp.pos]
	}

	if len(a) > 0 {
		sp.pos += a[0]
	} else {
		sp.pos++
	}
	return
}

func (sp *strPos) Get(i int) byte {
	if sp.End() {
		return 0
	}
	return sp.s[sp.pos+i]
}

type matchState struct {
	matchdepth, level int    // control for recursive depth (to avoid stack overflow), total number of captures (finished or unfinished)
	s, p              strPos // source string, pattern
	capture           [max_captures]struct {
		init strPos
		len  int
	}
}

func check_capture(ms *matchState, l int) (int, error) {
	l -= '1'
	if l < 0 || l >= ms.level || ms.capture[l].len == cap_unfinished {
		return 0, fmt.Errorf("invalid capture index %%%d", l+1)
	}
	return l, nil
}

func capture_to_close(ms *matchState) (int, error) {
	level := ms.level
	for level--; level >= 0; level-- {
		if ms.capture[level].len == cap_unfinished {
			return level, nil
		}
	}
	return 0, errors.New("invalid pattern capture")
}

func classend(p strPos) (strPos, error) {
	switch p.Inc() {
	case l_esc:
		if p.End() {
			return strPos{}, errors.New("malformed pattern (ends with '%')")
		}
		return strPos{p.s, p.pos + 1}, nil
	case '[':
		if p.Get(0) == '^' {
			p.Inc()
		}

		for { // look for a ']'
			if p.End() {
				return strPos{}, errors.New("malformed pattern (missing ']')")
			}

			p.Inc()
			if p.Get(-1) == l_esc && !p.End() {
				p.Inc() // skip escapes (e.g. '%]')
			}

			if p.Get(0) == ']' {
				break
			}
		}

		return strPos{p.s, p.pos + 1}, nil
	}
	return p, nil
}

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

func match_class(c, cl byte) (res bool) {
	switch tolower(cl) {
	case 'a':
		res = isalpha(c)
	case 'c':
		res = iscntrl(c)
	case 'd':
		res = isdigit(c)
	case 'g':
		res = isgraph(c)
	case 'l':
		res = islower(c)
	case 'p':
		res = ispunct(c)
	case 's':
		res = isspace(c)
	case 'u':
		res = isupper(c)
	case 'w':
		res = isalnum(c)
	case 'x':
		res = isxdigit(c)
	case 'z': // deprecated option
		res = c == 0
	default:
		return cl == c
	}

	if islower(cl) {
		return
	}
	return !res
}

func matchbracketclass(c byte, p, ec strPos) bool {
	sig := true
	if p.Get(1) == '^' {
		sig = false
		p.Inc() // skip the '^'
	}

	for {
		p.Inc()
		if p.pos >= ec.pos {
			break
		}

		if p.Get(0) == l_esc {
			p.Inc()
			if match_class(c, p.Get(0)) {
				return sig
			}
		} else if (p.Get(1) == '-') && (p.pos+2 < ec.pos) {
			p.Inc(2)
			if p.Get(-2) <= c && c <= p.Get(0) {
				return sig
			}
		} else if p.Get(0) == c {
			return sig
		}
	}

	return !sig
}

// hot single pattern matching groups in your area
func singlematch(s, p, ep strPos) bool {
	if s.End() {
		return false
	}

	// fmt.Println("    SINGLEMATCH", fmt.Sprintf("%c", s.Get(0)), fmt.Sprintf("%c", p.Get(0)))

	c := s.Get(0)
	switch p.Get(0) {
	case '.':
		return true // matches any char
	case l_esc:
		return match_class(c, p.Get(1))
	case '[':
		return matchbracketclass(c, p, strPos{ep.s, ep.pos - 1})
	}
	return p.Get(0) == c
}

func matchbalance(s, p strPos) (sp strPos, err error) {
	if p.pos >= len(p.s)-1 {
		return strPos{}, errors.New("malformed pattern (missing arguments to '%%b')")
	} else if s.Get(0) != p.Get(0) {
		return
	}

	b := p.Get(0)
	e := p.Get(1)
	cont := 1
	for {
		s.Inc()
		if s.End() {
			break
		}

		if s.Get(0) == e {
			cont--
			if cont == 0 {
				return strPos{s.s, s.pos + 1}, nil
			}
		} else if s.Get(0) == b {
			cont++
		}
	}

	return // string ends out of balance
}

func max_expand(ms *matchState, s, p, ep strPos) (sp strPos, err error) {
	i := 0 // counts maximum expand for item
	for singlematch(strPos{s.s, s.pos + i}, p, ep) {
		i++
	}

	// keeps trying to match with the maximum repetitions
	for i >= 0 {
		res, err := match(ms, strPos{s.s, s.pos + i}, strPos{ep.s, ep.pos + 1})
		if err != nil {
			return strPos{}, err
		} else if res != (strPos{}) {
			return res, nil
		}
		i-- // else didn't match; reduce 1 repetition to try again
	}
	return
}

func min_expand(ms *matchState, s, p, ep strPos) (strPos, error) {
	for {
		if res, err := match(ms, s, strPos{ep.s, ep.pos + 1}); err != nil {
			return strPos{}, err
		} else if res != (strPos{}) {
			return res, nil
		} else if singlematch(s, p, ep) {
			s.Inc() // try with one more repetition
		} else {
			return strPos{}, nil
		}
	}
}

func start_capture(ms *matchState, s, p strPos, what int) (strPos, error) { // what
	level := ms.level
	if level >= max_captures {
		return strPos{}, errors.New("too many captures")
	}

	ms.capture[level].init = s
	ms.capture[level].len = what
	ms.level = level + 1
	res, err := match(ms, s, p)
	if err != nil {
		return strPos{}, err
	} else if res == (strPos{}) { // match failed?
		ms.level-- // undo capture
	}

	return res, nil
}

func end_capture(ms *matchState, s, p strPos) (strPos, error) {
	l, err := capture_to_close(ms)
	if err != nil {
		return strPos{}, err
	}

	ms.capture[l].len = s.pos - ms.capture[l].init.pos // close capture
	res, err := match(ms, s, p)
	if err != nil {
		return strPos{}, err
	} else if res == (strPos{}) { // match failed?
		ms.capture[l].len = cap_unfinished // undo capture
	}

	return res, nil
}

// oh god
func memcmp(a, b strPos, n int) bool {
	return strings.Compare(a.s[a.pos:a.pos+n], b.s[b.pos:b.pos+n]) == 0
}

func match_capture(ms *matchState, s strPos, l int) (strPos, error) {
	l, err := check_capture(ms, l)
	if err != nil {
		return strPos{}, err
	}

	len := ms.capture[l].len
	if (s.Len()-s.pos) >= len && memcmp(ms.capture[l].init, s, len) {
		return strPos{s.s, s.pos + len}, nil
	}

	return strPos{}, nil
}

func match(ms *matchState, s, p strPos) (sp strPos, err error) {
	if ms.matchdepth == 0 {
		return strPos{}, errors.New("pattern too complex")
	}
	ms.matchdepth--

	// interrupt stuff, hmm going to ignore this

	// fmt.Println("pos init", s.pos)

	// THERE IS NEVER A REASONABLE USE CASE FOR A GOTO!!!!!!!!!!!!!!!! xkcd292
init:
	if p.End() { // end of pattern?
		ms.matchdepth++
		// fmt.Println("pos end", s.pos)
		return s, nil
	}

	dflt := func() (goback bool, err error) {
		// fmt.Println("  default case", p.pos, fmt.Sprintf("%c", p.Get(0)))

		// pattern class plus optional suffix
		ep, err := classend(p) // points to optional suffix
		if err != nil {
			return false, err
		}

		// does not match at least once?
		if !singlematch(s, p, ep) {
			if ep.Get(0) == '*' || ep.Get(0) == '?' || ep.Get(0) == '-' {
				// accept empty?
				p.pos = ep.pos + 1
				return true, nil
			}

			// '+' or no suffix
			// fmt.Println("  + or no suffix")
			s = strPos{} // fail
		} else { // matched once
			// matched? I was matched once

			switch ep.Get(0) { // handle optional suffix
			case '?': // optional
				res, err := match(ms, strPos{s.s, s.pos + 1}, strPos{ep.s, ep.pos + 1})
				if err != nil {
					return false, err
				} else if res != (strPos{}) {
					s = res
				} else {
					p = strPos{ep.s, ep.pos + 1}
					return true, nil
				}
			case '+': // 1 or more repetitions
				s.Inc() // 1 match already done
				fallthrough
			case '*': // 0 or more repetitions
				if s, err = max_expand(ms, s, p, ep); err != nil {
					return false, err
				}
			case '-': // 0 or more repetitions (minimum)
				if s, err = min_expand(ms, s, p, ep); err != nil {
					return false, err
				}
			default: // no suffix
				p = ep
				return true, nil
			}
		}
		return false, nil
	}

	// fmt.Println("gettin'", fmt.Sprintf("%c", p.Get(0)))
	switch p.Get(0) {
	case '(': // start capture
		if p.Get(1) == ')' { // position capture?
			if s, err = start_capture(ms, s, strPos{p.s, p.pos + 2}, cap_position); err != nil {
				return
			}
		} else if s, err = start_capture(ms, s, strPos{p.s, p.pos + 1}, cap_unfinished); err != nil {
			return
		}
	case ')': // end capture
		if s, err = end_capture(ms, s, strPos{p.s, p.pos + 1}); err != nil {
			return
		}
	case '$':
		if p.pos+1 != len(p.s) { // is the `$' the last char in pattern?
			if goback, err := dflt(); err != nil {
				return strPos{}, err
			} else if goback {
				goto init
			}
		} else if !s.End() {
			s = strPos{} // check end of string
		}
	case l_esc: // escaped sequences not in the format class[*+?-]?
		switch p.Get(1) {
		case 'b': // balanced string?
			if s, err = matchbalance(s, strPos{p.s, p.pos + 2}); err != nil {
				return
			} else if s != (strPos{}) {
				p.Inc(4)
				goto init // return match(ms, s, &strPos{p.s, p.pos + 4})
			} // else fail (s == nil)
		case 'f': // frontier?
			p.Inc(2)
			if p.Get(0) != '[' {
				return strPos{}, errors.New("missing '[' after '%f' in pattern")
			}

			ep, err := classend(p) // points to what is next
			if err != nil {
				return strPos{}, err
			}

			var previous byte
			if s.pos == 0 {
				previous = '\000'
			} else {
				previous = s.Get(-1)
			}

			if !matchbracketclass(previous, p, strPos{ep.s, ep.pos - 1}) &&
				matchbracketclass(s.Get(0), p, strPos{ep.s, ep.pos - 1}) {
				p = ep
				goto init // return match(ms, s, ep)
			}
			s = strPos{} // match failed
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // capture results (%0-%9)?
			if s, err = match_capture(ms, s, int(p.Get(1))); err != nil {
				return
			} else if s == (strPos{}) {
				p.Inc(2)
				goto init // return match(ms, s, &strPos{p.s, p.pos + 2})
			}
		default:
			if goback, err := dflt(); err != nil {
				return strPos{}, err
			} else if goback {
				goto init
			}
		}
	default:
		if goback, err := dflt(); err != nil {
			return strPos{}, err
		} else if goback {
			goto init
		}
	}

	ms.matchdepth++
	// fmt.Println("pos end", s.pos)
	return s, nil
}

func memchr(s *strPos, c byte, n int) *strPos {
	for i := 0; i < n; i++ {
		if s.Get(i) == c {
			return &strPos{s.s, s.pos + i}
		}
	}
	return nil
}

// check whether pattern has no special characters
func nospecials(p string) bool {
	for _, c := range p {
		if strings.ContainsRune(specials, c) {
			return false // pattern has a special character
		}
	}
	return true // no special characters found
}

func newstate(s, p strPos) *matchState {
	ms := matchState{
		matchdepth: maxcalls,
		s:          s,
		p:          p,
	}
	return &ms
}

func reprepstate(ms *matchState) {
	ms.level = 0
	// LUAU_ASSERT(ms.matchdepth == maxcalls)
}

func push_onecapture(ms *matchState, i int, s strPos) (r any, err error) {
	if i >= ms.level {
		if i == 0 { // ms.level == 0, too
			return s.String(), nil
		}
		return nil, errors.New("invalid capture index")
	}
	l := ms.capture[i].len
	if l == cap_unfinished {
		return nil, errors.New("unfinished capture")
	} else if l == cap_position {
		return float64(ms.capture[i].init.pos + 1), nil
	}
	return ms.capture[i].init.String(), nil
}

func push_captures(ms *matchState, s strPos) (r Rets, err error) {
	nlevels := 1
	if ms.level != 0 || s == (strPos{}) {
		nlevels = ms.level
	}
	// luaL_checkstack(ms->L, nlevels, "too many captures"); // idk
	r = make(Rets, nlevels)
	for i := range nlevels {
		res, err := push_onecapture(ms, i, s)
		if err != nil {
			return nil, err
		}
		r = append(r, res)
	}
	return
}

func str_find_aux(args Args, find bool) (r Rets, err error) {
	s, p := args.GetString(), args.GetString()
	i := args.GetNumber(1)
	plain := args.GetBool(false)
	ls := len(s)

	// fmt.Println("FIND", s, p, i, plain)
	init := string_posrelat(int(i), ls)
	if init < 1 {
		init = 1
	} else if init > ls+1 { // start after string's end?
		return Rets{nil}, nil // cannot find anything
	}

	// explicit request or no special characters?

	if find && (plain || nospecials(p)) {
		// do a plain search, good-style
		pos := strings.Index(s[init-1:], p)
		if pos == -1 {
			return Rets{nil}, nil // one nil
		}
		return Rets{
			float64(pos + init),
			float64(pos + init + len(p) - 1),
		}, nil
	}

	s1 := strPos{s: s, pos: init - 1}
	ss, pp := strPos{s: s}, strPos{s: p}

	anchor := pp.Get(0) == '^'
	if anchor {
		pp.Inc() // skip anchor character
	}
	ms := newstate(ss, pp)
	for {
		reprepstate(ms)
		if res, err := match(ms, s1, pp); err != nil {
			return nil, err
		} else if res != (strPos{}) {
			if !find {
				return push_captures(ms, s1)
			}

			// fmt.Println("s1pos", s1.pos)

			r := Rets{float64(s1.pos + 1), float64(res.pos)} // start, end
			caps, err := push_captures(ms, strPos{})
			if err != nil {
				return nil, err
			}

			return append(r, caps...), nil
		}

		if s1.End() || anchor {
			s1.Inc()
			break
		}
		s1.Inc()
	}

	return Rets{nil}, nil // not found
}

func string_find(args Args) (r Rets, err error) {
	return str_find_aux(args, true)
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
