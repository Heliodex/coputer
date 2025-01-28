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
	if si >= len(s) {
		return false
	}

	c := s[si]
	switch p[pi] {
	case '.': // matches any char
		return true
	case l_esc:
		return match_class(c, p[pi+1])
	case '[':
		return matchbracketclass(c, p, pi, epi-1)
	}
	return p[pi] == c
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

func maxExpand(s, p string, si, pi, epi int) (int, error) {
	i := 0 // counts maximum expand for item
	for singlematch(s, p, si+i, pi, epi) {
		i++
	}

	// keeps trying to match with the maximum repetitions
	for i >= 0 {
		// res, err := match(ms, strPos{s.s, s.pos + i}, strPos{ep.s, ep.pos + 1})
		res, _, err := matchPos(s, p, si+i, epi+1)
		if err != nil {
			return 0, err
		} else if res != -1 {
			return res, nil
		}
		i-- // else didn't match; reduce 1 repetition to try again
	}
	return -1, nil
}

func minExpand(s, p string, si, pi, epi int) (int, error) {
	for {
		if res, _, err := matchPos(s, p, si, epi+1); err != nil {
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
	captures [max_captures]capture
}

func startCapture(s, p string, si, pi *int, caps *captures) {
}

func endCapture(s, p string, si, pi *int, caps *captures) {
}

func matchPos(s, p string, si, pis int) (si2 int, caps captures, err error) {
	dlog("\nMATCHPOS", s, si, pis)

	depth++
	defer func() { depth-- }()

mainloop:
	for pi := pis; pi < len(p); pi++ {
		dlog("ITER", pi, string(p[pi]))

		switch p[pi] {
		/*
			case '(': // start capture
				startCapture(s, p, &si, &pi, &caps)
			case ')': // end capture
				endCapture(s, p, &si, &pi, &caps)
			case '$':
				if pi+1 != len(p) { // '$' at the end of pattern?
					panic("default")
					// if d, err := defaultCase(s, p, &si, &pi); err != nil {
					// 	return 0, captures{}, err
					// } else if !d {
					// 	return -1, captures{}, nil
					// }
				} else if si+1 != len(s) { // at the end of string?
					return -1, captures{}, nil
				}
			case l_esc: // escaped sequences not in the format class[*+?-]?
				switch p[pi+1] {
				case 'b': // balanced string?
				case 'f': // frontier?
				case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // capture results (%0-%9)?
				default:
					// go to default
					panic("default")
					// if d, err := defaultCase(s, p, &si, &pi); err != nil {
					// 	return 0, captures{}, err
					// } else if !d {
					// 	return -1, captures{}, nil
					// }
				}
		*/
		default:
			dlog("default", s, "[", si, pi, "]")

			// pattern class plus optional suffix
			epi, err := classend(p, pi) // points to optional suffix
			if err != nil {
				return 0, captures{}, err
			}

			// does not match at least once?
			if !singlematch(s, p, si, pi, epi) {
				dlog("  nomatch", epi)

				if epi < len(p) && (p[epi] == '*' || p[epi] == '?' || p[epi] == '-') {
					// accept empty?
					dlog("  accept empty")
					pi = epi
					// dlog("  pi is", pi, epi+1, string(p[epi+1]))
					continue
				}

				// '+' or no suffix
				dlog("  '+' or no suffix")
				si = -1 // fail (?)
				break mainloop
			}

			// matched once
			dlog("matched once")

			if epi >= len(p) {
				dlog("after end")
				si++
				pi = epi - 1
				continue
			}

			switch p[epi] { // handle optional suffix
			case '?': // optional
				dlog("optional", s[si+1:], p[epi+1:])

				si2, _, err := matchPos(s, p, si+1, epi+1)
				if err != nil {
					return 0, captures{}, err
				}

				dlog("optional done", si, si2)

				if si2 != -1 {
					si = si2
				} else {
					dlog("pi is", pi, epi+1)
					pi = epi
					continue
				}
			case '+': // 1 or more repetitions
				si++        // 1 match already done
				fallthrough // go through
			case '*': // 0 or more repetitions
				si, err = maxExpand(s, p, si, pi, epi)
				if err != nil {
					return 0, captures{}, err
				}
			case '-': // 0 or more repetitions (minimum)
				si, err = minExpand(s, p, si, pi, epi)
				if err != nil {
					return 0, captures{}, err
				}
			default: // no suffix
				dlog("no suffix")
				si++
				pi = epi - 1
				continue
			}

			break mainloop
		}
	}

	dlog("MATCHPOS DONE", si, pis)
	return si, caps, nil
}

func match(s, p string) (start, end int, caps captures, err error) {
	dlog("MATCH", s, p, len(s), len(p))

	depth++
	defer func() { depth-- }()

	m := 1
	if p[0] != '^' {
		m = len(s)
	}

	for ; start < m; start++ {
		e, caps2, err := matchPos(s[start:], p, 0, 0)
		if err != nil {
			return -1, -1, captures{}, err
		} else if e == -1 {
			continue
		}
		return start, start + e, caps2, nil
	}

	return -1, -1, captures{}, nil
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
	start, end, _, err := match(s[init-1:], p)
	if err != nil {
		return nil, err
	} else if start == -1 {
		return Rets{nil}, nil // not found
	}

	return Rets{
		float64(start + init),
		float64(end + init - 1),
	}, nil
}
