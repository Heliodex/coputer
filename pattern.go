package litecode

import (
	"errors"
	"fmt"
	"strings"
)

var depth = 0

func dlog(a ...any) {
	fmt.Print(strings.Repeat("  ", depth))
	fmt.Println(a...)
}

func classend(p string, pi int) (int, error) {
	// dlog("CLASSEND", p, pi, string(p[pi]))

	depth++
	defer func() { depth-- }()

	switch p[pi] {
	case l_esc:
		pi++

		if pi == len(p) {
			return 0, fmt.Errorf("malformed pattern (ends with '%c')", l_esc)
		}
	case '[':
		pi++

		if p[pi] == '^' {
			pi++
		}

		for {
			// look for a ']'
			if pi == len(p) {
				return 0, errors.New("malformed pattern (missing ']')")
			} else if p[pi] == l_esc && pi+1 < len(p) {
				pi++
			}
			pi++

			if p[pi] == ']' {
				break
			}
		}
	}

	return pi + 1, nil
}

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

func matchbracketclass(c byte, p string, pi, eci int) bool {
	sig := true
	if p[pi+1] == '^' {
		sig = false
		pi++ // skip the '^'
	}

	for {
		pi++
		if pi >= eci {
			break
		}

		if p[pi] == l_esc {
			pi++
			if match_class(c, p[pi]) {
				return sig
			}
		} else if p[pi+1] == '-' && pi+2 < eci {
			pi += 2
			if p[pi-2] <= c && c <= p[pi] {
				return sig
			}
		} else if p[pi] == c {
			return sig
		}
	}

	return !sig
}

func singlematch(s, p string, si, pi, epi int) bool {
	dlog("SINGLEMATCH", s, p, si, pi, epi)

	if si >= len(s) {
		return false
	}

	c := s[si]
	switch p[pi] {
	case '.': // matches any char
		return true
	case l_esc:

		m := match_class(c, p[pi+1])
		dlog("  match class", string(c), p[pi+1], m)
		return m
	case '[':
		return matchbracketclass(c, p, pi, epi-1)
	}
	return p[pi] == c
}

func push_onecapture(s string, start, end, i int, caps captures) (any, error) {
	if i >= caps.level {
		if i == 0 { // caps.level == 0, too
			return s[start:end], nil // add whole match
		}
		return nil, errors.New("invalid capture index")
	}

	cap := caps.captures[i]
	l, i := cap.len, cap.init
	if l == cap_unfinished {
		return nil, errors.New("unfinished capture")
	} else if l == cap_position {
		return float64(caps.captures[i].init + 1), nil
	}
	return s[i : i+l], nil
}

func push_captures(s string, start, end int, caps captures) (r Rets, err error) {
	nlevels := caps.level
	if nlevels == 0 && start != -1 { // -1? hmm
		nlevels = 1
	}

	r = make(Rets, nlevels)
	for i := range nlevels {
		if r[i], err = push_onecapture(s, start, end, i, caps); err != nil {
			return nil, err
		}
	}

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

func maxExpand(s, p string, si, pi, epi int, caps captures) (si2 int, err error) {
	i := 0 // counts maximum expand for item
	for singlematch(s, p, si+i, pi, epi) {
		i++
	}

	// keeps trying to match with the maximum repetitions
	dlog("xpanding", i)
	for i >= 0 {
		if si, err = matchPos(s, p, si+i, epi+1, caps); err != nil {
			return 0, err
		} else if si != -1 {
			dlog("xpandmatched", si)
			return si, nil
		}
		i-- // else didn't match; reduce 1 repetition to try again
	}

	dlog("xpandfailed")
	return -1, nil
}

func minExpand(s, p string, si, pi, epi int, caps captures) (int, error) {
	for {
		if res, err := matchPos(s, p, si, epi+1, caps); err != nil {
			return 0, err
		} else if res != -1 {
			return res, nil
		} else if singlematch(s, p, si, pi, epi) {
			si++ // try with one more repetition
		} else {
			return -1, nil
		}
	}
}

type capture struct {
	init, len int
}

type captures struct {
	level    int
	captures []capture
}

func captureToClose(caps captures) (int, error) {
	for level := caps.level - 1; level >= 0; level-- {
		fmt.Println(" cap level", level, caps.captures[level])
		if caps.captures[level].len == cap_unfinished {
			return level, nil
		}
	}
	return 0, errors.New("invalid pattern capture")
}

func startCapture(s, p string, si, pi, what int, caps captures) (si2 int, err error) {
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
	} else if si == -1 { // match failed?
		caps.level-- // undo capture
	}

	return si, nil
}

func endCapture(s, p string, si, pi int, caps captures) (si2 int, err error) {
	l, err := captureToClose(caps)
	if err != nil {
		return 0, err
	}

	fmt.Println("endcapture", l, caps.captures[l])

	caps.captures[l].len = si - caps.captures[l].init // close capture
	if si, err = matchPos(s, p, si, pi, caps); err != nil {
		return 0, err
	} else if si == -1 { // match failed?
		fmt.Println("undo capture")
		caps.captures[l].len = cap_unfinished // undo capture
	}

	return si, nil
}

func optSuffix(s, p string, si, pi, epi int, caps captures) (cont bool, si2, pi2 int, err error) {
	switch p[epi] {
	case '?': // optional
		dlog("optional", s[si+1:], p[epi+1:])

		si2, err = matchPos(s, p, si+1, epi+1, caps)
		if err != nil {
			return
		}

		dlog("optional done", si, si2)

		if si2 != -1 {
			si = si2
		} else {
			dlog("pi is", pi, epi+1)
			pi = epi
			return true, si, pi, nil
		}
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
		dlog("no suffix")
		si++
		pi = epi - 1
		return true, si, pi, nil
	}

	return false, si, pi, nil
}

func defaultCase(s, p string, si, pi, epi int, caps captures) (cont bool, si2, pi2 int, err error) {
	dlog("default", s, "[", si, pi, "]")

	// pattern class plus optional suffix
	epi, err = classend(p, pi) // points to optional suffix
	if err != nil {
		return false, 0, 0, err
	} else if !singlematch(s, p, si, pi, epi) {
		// does not match at least once?
		dlog("  nomatch", epi)

		if epi < len(p) && (p[epi] == '*' || p[epi] == '?' || p[epi] == '-') {
			// accept empty?
			dlog("  accept empty")
			pi = epi
			// dlog("  pi is", pi, epi+1, string(p[epi+1]))
			return true, si, pi, nil // get out of default free card
		}

		// '+' or no suffix
		dlog("  '+' or no suffix")
		return false, -1, pi, nil // fail
	}

	// matched once
	dlog("matched once")

	if epi >= len(p) {
		dlog("after end")
		si++
		pi = epi - 1
		return true, si, pi, nil
	}

	// handle optional suffix
	return optSuffix(s, p, si, pi, epi, caps)
}

func matchPos(s, p string, si, pis int, caps captures) (si2 int, err error) {
	dlog("MATCHPOS", s, p, si, pis)

	depth++
	defer func() { depth-- }()

	for pi := pis; pi < len(p); pi++ {
		dlog("ITER", pi, string(p[pi]))

		switch p[pi] {
		case '(': // start capture
			what := cap_unfinished
			if p[pi+1] == ')' { // position capture?
				dlog("POSITION CAPTURE")
				what = cap_position
			}

			dlog("start", what, pi, pi-what)
			return startCapture(s, p, si, pi-what, what, caps)
		case ')': // end capture
			dlog("end", s, p, si, pi)
			return endCapture(s, p, si, pi+1, caps)
		case '$':
			if pi+1 != len(p) { // s the '$' the last char in pattern?
				// no; go to default
				var cont bool
				if cont, si, pi, err = defaultCase(s, p, si, pi, len(p), caps); !cont {
					return si, err
				}
			} else if si != len(s) { // check end of string
				dlog("not end of string", len(s)-si)
				return -1, nil
			}
		case l_esc: // escaped sequences not in the format class[*+?-]?
			switch p[pi+1] {
			case 'b': // balanced string?
			case 'f': // frontier?
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // capture results (%0-%9)?
			default:
				// go to default
				var cont bool
				if cont, si, pi, err = defaultCase(s, p, si, pi, len(p), caps); !cont {
					return si, err
				}
			}
		default:
			// go to default; do not pass go, do not collect £200
			var cont bool
			if cont, si, pi, err = defaultCase(s, p, si, pi, len(p), caps); !cont {
				return si, err
			}
		}
	}

	dlog("MATCHPOS DONE", si, pis)
	return si, nil
}

func match(s, p string, caps captures) (start, end int, err error) {
	dlog("MATCH", s, p, len(s), len(p))

	m, pis := 0, 1
	if p[0] != '^' {
		m, pis = len(s), 0
	}

	for ; start <= m; start++ {
		// reprep state
		caps.level = 0

		depth++
		e, err := matchPos(s[start:], p, 0, pis, caps)
		if err != nil {
			return -1, -1, err
		}
		depth--

		dlog("MATCH DONE", start, e)

		if e == -1 {
			continue
		}
		return start, start + e, nil
	}

	return -1, -1, nil
}

func str_find_aux(s, p string, i int, plain, find bool) (r Rets, err error) {
	ls := len(s)

	init := string_posrelat(i, ls)
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

	// pattern matching
	caps := captures{}

	start, end, err := match(s[init-1:], p, caps)
	if err != nil {
		return nil, err
	} else if start == -1 {
		return Rets{nil}, nil // not found
	}

	if !find {
		return push_captures(s, start, end, caps)
	}

	rs, err := push_captures(s, -1, 0, caps)
	if err != nil {
		return nil, err
	}

	fmt.Println("pushed", rs)

	return append(Rets{
		float64(start + init),
		float64(end + init - 1),
	}, rs...), nil
}
