// Pattern matching seems to be some sort of "great filter" for Lua/Luau VM implementations in other languages.
// Where I now stand, many have fallen before me.

package std

import (
	"errors"
	"fmt"
	"strings"

	. "github.com/Heliodex/coputer/litecode/types"
)

// loggin shit for login' shit
// var depth = 0

// func dlog(a ...any) {
// 	fmt.Print(strings.Repeat("  ", depth))
// 	fmt.Println(a...)
// }

const max_captures = 32

var (
	errMalformedEsc     = fmt.Errorf("malformed pattern (ends with '%c')", l_esc)
	errMalformedBracket = errors.New("malformed pattern (missing ']')")
)

func classend(p string, pi int) (int, error) {
	// dlog("CLASSEND", p, pi, string(p[pi]))

	// depth++
	// defer func() { depth-- }()

	switch p[pi] {
	case l_esc:
		pi++

		if pi == len(p) {
			return 0, errMalformedEsc
		}
	case '[':
		pi++

		if pi == len(p) {
			return 0, errMalformedBracket
		}

		if p[pi] == '^' {
			pi++
		}

		for {
			// look for a ']'
			if pi+1 >= len(p) {
				return 0, errMalformedBracket
			}
			if p[pi] == l_esc {
				pi++ // skip escapes (eg. '%]')
			}
			pi++

			if p[pi] == ']' {
				break
			}
		}
	}

	return pi + 1, nil
}

func matchclass(c, cl byte) (res bool) {
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

	return res == islower(cl)
}

func matchbracketclass(c byte, p string, pi, eci int) (sig bool) {
	// fmt.Println("MBC", p[pi:eci], pi, eci)

	sig = true // nock mont 3
	if p[pi+1] == '^' {
		sig = false
		pi++ // skip the '^'
	}

	for ; pi < eci; pi++ {
		// fmt.Printf("  %d %c, %d\n", c, p[pi], pi)
		if p[pi] == l_esc {
			// fmt.Println("it's a %")
			pi++
			if matchclass(c, p[pi]) {
				// fmt.Println("MATCHEDClass")
				return
			}
		} else if p[pi+1] == '-' && pi+2 < eci {
			pi += 2
			if p[pi-2] <= c && c <= p[pi] {
				return
			}
		} else if p[pi] == c {
			return
		}
	}

	return !sig
}

func singlematch(s, p string, si, pi, epi int) bool {
	// dlog("SINGLEMATCH", s, p, si, pi, epi)

	if si >= len(s) {
		return false
	}

	c := s[si]
	switch p[pi] {
	case '.': // matches any char
		return true
	case l_esc:
		// dlog("  match class", string(c), p[pi+1], m)
		return matchclass(c, p[pi+1])
	case '[':
		return matchbracketclass(c, p, pi, epi-1)
	}
	return p[pi] == c
}

func matchbalance(s, p string, si, pi int) (int, error) {
	if pi >= len(p)-1 {
		return 0, errors.New("malformed pattern (missing arguments to '%b')")
	}

	b := p[pi]
	if s[si] != b {
		return -1, nil
	}

	e := p[pi+1]
	for cont := 1; ; {
		si++
		if si >= len(s) {
			break
		}

		switch s[si] {
		case e:
			cont--
			if cont == 0 {
				return si + 1, nil
			}
		case b:
			cont++
		}
	}

	return -1, nil // string ends out of balance
}

func pushCapture(caps *captures, s string, start, end, i int) (Val, error) {
	// fmt.Println("    push", start, end, i, caps)

	if i >= caps.level {
		if i != 0 {
			return nil, errors.New("invalid capture index")
		}
		if start == -1 {
			// caps.level == 0, too
			// fmt.Println("    adding whole match")
			return s, nil // add whole string
		}
		return s[start:end], nil // add whole match
	}

	c := caps.captures[i]
	l, i := c.len, c.init
	if l == cap_unfinished {
		return nil, errors.New("unfinished capture")
	}
	if l == cap_position {
		// fmt.Println("    adding cap position")
		return float64(i + 1), nil
	}
	// fmt.Println("    adding normal")
	return s[i:][:l], nil
}

func pushCaptures(caps *captures, s string, start, end int, find bool) (r []Val, err error) {
	// fmt.Println("  PUSHING CAPS", start, end, caps)

	nlevels := caps.level
	if nlevels == 0 { // hmm
		if find {
			// fmt.Println("  no nlevels")
			return
		}
		nlevels = 1
	}

	// fmt.Println("  nlevels", nlevels)

	r = make([]Val, nlevels)
	for i := range nlevels {
		if r[i], err = pushCapture(caps, s, start, end, i); err != nil {
			return
		}
		// fmt.Println("  captured", r[i])
	}

	// fmt.Println("  PUSHED", r)

	return
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

func maxExpand(s, p string, si, pi, epi int, caps *captures) (res int, err error) {
	var i int // counts maximum expand for item
	for singlematch(s, p, si+i, pi, epi) {
		i++
	}

	// keeps trying to match with the maximum repetitions
	// dlog("xpanding", si, i)
	for i >= 0 {
		// dlog("reduced", si, i)
		if res, err = matchPos(s, p, si+i, epi+1, caps); err != nil {
			return
		}
		if res != -1 {
			// dlog("xpandmatched", si, caps.captures)
			return res, nil
		}
		i-- // else didn't match; reduce 1 repetition to try again
	}

	// dlog("xpandfailed")
	return -1, nil
}

func minExpand(s, p string, si, pi, epi int, caps *captures) (res int, err error) {
	for ; ; si++ {
		if res, err = matchPos(s, p, si, epi+1, caps); err != nil || res != -1 {
			return
		}
		if !singlematch(s, p, si, pi, epi) {
			return -1, nil
		}
		// try with one more repetition
	}
}

type capture struct {
	init, len int
}

type captures struct {
	level    int
	captures []capture
}

func checkCapture(l int, caps *captures) (int, error) {
	l -= '1'
	if l < 0 || l >= caps.level || caps.captures[l].len == cap_unfinished {
		return 0, fmt.Errorf("invalid capture index %%%d", l+1)
	}

	return l, nil
}

func captureToClose(caps *captures) (int, error) {
	for level := caps.level - 1; level >= 0; level-- {
		// fmt.Println(" cap level", level, caps.captures, caps.captures[level])
		if caps.captures[level].len == cap_unfinished {
			return level, nil
		}
	}
	return 0, errors.New("invalid pattern capture")
}

func startCapture(s, p string, si, pi, what int, caps *captures) (si2 int, err error) {
	level := caps.level
	if level >= max_captures {
		return 0, errors.New("too many captures")
	}

	for len(caps.captures) <= level {
		caps.captures = append(caps.captures, capture{})
	}

	caps.captures[level].init = si
	caps.captures[level].len = what
	caps.level++

	if si, err = matchPos(s, p, si, pi, caps); err != nil {
		return 0, err
	}
	if si == -1 { // match failed?
		caps.level-- // undo capture
	}

	return si, nil
}

func endCapture(s, p string, si, pi int, caps *captures) (si2 int, err error) {
	l, err := captureToClose(caps)
	if err != nil {
		return 0, err
	}

	// fmt.Println("endcapture", l, caps.captures, caps.captures[l])

	caps.captures[l].len = si - caps.captures[l].init // close capture
	if si, err = matchPos(s, p, si, pi, caps); err != nil {
		return 0, err
	}
	if si == -1 { // match failed?
		// fmt.Println("undo capture")
		caps.captures[l].len = cap_unfinished // undo capture
	}

	// fmt.Println("endcapture done", si, caps.captures)

	return si, nil
}

func matchCapture(s string, si, l int, caps *captures) (i int, err error) {
	if l, err = checkCapture(l, caps); err != nil {
		return
	}
	// fmt.Println("matching catching", s, si, l)
	c := caps.captures[l]

	if ll := c.len; len(s) >= ll {
		if s2 := s[si:]; s[c.init:][:ll] == s2[:min(len(s2), ll)] {
			return si + ll, nil
		}
	}
	return -1, nil
}

func optSuffix(s, p string, si, pi, epi int, caps *captures) (cont bool, si2, pi2 int, err error) {
	switch p[epi] {
	case '?': // optional
		// dlog("optional", s[si+1:], p[epi+1:])
		if si2, err = matchPos(s, p, si+1, epi+1, caps); err != nil {
			return
		}
		if si2 == -1 {
			// dlog("pi is", pi, epi+1)
			return true, si, epi, nil
		}
		// dlog("optional done", si, si2)
		si = si2
	case '+': // 1 or more repetitions
		si++        // 1 match already done
		fallthrough // go through
	case '*': // 0 or more repetitions
		if si, err = maxExpand(s, p, si, pi, epi, caps); err != nil {
			return
		}
	case '-': // 0 or more repetitions (minimum)
		if si, err = minExpand(s, p, si, pi, epi, caps); err != nil {
			return
		}
	default: // no suffix
		// dlog("no suffix")
		return true, si + 1, epi - 1, nil
	}

	return false, si, pi, nil
}

func defaultCase(s, p string, si, pi int, caps *captures) (cont bool, si2, pi2 int, err error) {
	// dlog("default", s, "[", si, pi, "]")

	// pattern class plus optional suffix
	epi, err := classend(p, pi) // points to optional suffix
	if err != nil {
		return
	}
	if !singlematch(s, p, si, pi, epi) {
		// does not match at least once?
		// dlog("  nomatch", epi)

		if epi < len(p) && (p[epi] == '*' || p[epi] == '?' || p[epi] == '-') {
			// accept empty?
			// dlog("  accept empty")
			// dlog("  pi is", pi, epi, string(p[epi]))
			return true, si, epi, nil // get out of default free card
		}

		// '+' or no suffix
		// dlog("  '+' or no suffix")
		return false, -1, pi, nil // fail
	}
	if epi >= len(p) {
		// matched once
		// dlog("matched once")
		// dlog("after end")
		return true, si + 1, epi - 1, nil
	}

	// handle optional suffix
	return optSuffix(s, p, si, pi, epi, caps)
}

func matchPos(s, p string, si, pi int, caps *captures) (si2 int, err error) {
	// dlog("MATCHPOS", s, p, si, pis)

	// depth++
	// defer func() { depth-- }()

	for ; pi < len(p); pi++ {
		// dlog("ITER", pi, string(p[pi]), s[si:])

		switch p[pi] {
		case '(': // start capture
			what := cap_unfinished
			if p[pi+1] == ')' { // position capture?
				// dlog("POSITION CAPTURE")
				what = cap_position
			}

			// dlog("start", what, pi, pi-what)
			return startCapture(s, p, si, pi-what, what, caps)
		case ')': // end capture
			// dlog("end", s, p, si, pi, caps.captures)
			return endCapture(s, p, si, pi+1, caps)
		case '$':
			if pi+1 != len(p) { // s the '$' the last char in pattern?
				// no; go to default
				var cont bool
				if cont, si, pi, err = defaultCase(s, p, si, pi, caps); !cont {
					return si, err
				}
			} else if si != len(s) { // check end of string
				// dlog("not end of string", len(s)-si)
				return -1, nil
			}
		case l_esc: // escaped sequences not in the format class[*+?-]?
			if pi+1 == len(p) {
				return 0, errMalformedEsc
			}

			switch p[pi+1] {
			case 'b': // balanced string?
				if si, err = matchbalance(s, p, si, pi+2); err != nil {
					return 0, err
				}
				if si != -1 {
					pi += 3 // 1 less beacuse increment
					// dlog("matched balanced", pi < len(p))
					continue
				} // else fail (s == -1)
				return si, nil
			case 'f': // frontier?
				// dlog("the final frontier", si)

				pi += 2
				if pi == len(p) || p[pi] != '[' {
					return 0, errors.New("missing '[' after '%f' in pattern")
				}

				ep, err := classend(p, pi) // points to what is next
				if err != nil {
					return 0, err
				}
				// dlog(" P", pi)
				// dlog("EP", ep)

				// dlog("do it equal", si)
				var previous byte // \0
				if si != 0 {
					previous = s[si-1]
				}
				// dlog("previous", previous, fmt.Sprint(previous))

				// the frontier thing
				// dlog("m1", m1)
				if !matchbracketclass(previous, p, pi, ep-1) {
					var sc byte
					if si != len(s) {
						sc = s[si]
					}

					// dlog("m2", m2)
					if matchbracketclass(sc, p, pi, ep-1) {
						// dlog("  FRONTiER MATCHED")
						pi = ep - 1
						continue
					}
				}
				// dlog("  frontIer failed")
				return -1, nil // match failed
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // capture results (%0-%9)?
				if si, err = matchCapture(s, si, int(p[pi+1]), caps); err != nil {
					return 0, err
				}
				if si != -1 {
					// dlog("matched catched")
					pi += 1
					continue
				}
				return si, nil
			}
			// go to default (escaped character)
			fallthrough
		default:
			// go to default; do not pass go, do not collect £200
			var cont bool
			if cont, si, pi, err = defaultCase(s, p, si, pi, caps); !cont {
				return si, err
			}
		}
	}

	// dlog("MATCHPOS DONE", si, pis)
	return si, nil
}

func match(s, p string, caps *captures) (start, end int, err error) {
	// dlog("MATCH", s, p, len(s), len(p))

	m, pis := 0, 1
	if len(p) > 0 && p[0] != '^' {
		m, pis = len(s), 0
	}

	for ; start <= m; start++ {
		// reprep state
		caps.level = 0

		// depth++
		if end, err = matchPos(s, p, start, pis, caps); err != nil || end != -1 {
			return
		}
		// depth--

		// dlog("MATCH DONE", start, end)
	}

	return -1, -1, nil
}

func stringFindAux(s, p string, i int, plain, find bool) (r []Val, err error) {
	ls := len(s)

	init := max(1, string_posrelat(i, ls))
	if init > ls+1 { // start after string's end?
		return []Val{nil}, nil // cannot find anything
	}

	// explicit request or no special characters?
	is := s[init-1:]
	if find && (plain || nospecials(p)) {
		// do a plain search, good-style
		pos := strings.Index(is, p)
		if pos == -1 {
			return []Val{nil}, nil // one nil
		}
		return []Val{
			float64(pos + init),
			float64(pos + init + len(p) - 1),
		}, nil
	}

	// pattern matching
	caps := &captures{}

	start, end, err := match(is, p, caps)
	if err != nil {
		return
	}
	if start == -1 {
		return []Val{nil}, nil // not found
	}

	// fmt.Println("DONE!", caps)
	if r, err = pushCaptures(caps, is, start, end, find); err != nil || !find {
		return
	}

	// fmt.Println("pushed", rs)

	return append([]Val{
		float64(start + init),
		float64(end + init - 1),
	}, r...), nil
}
