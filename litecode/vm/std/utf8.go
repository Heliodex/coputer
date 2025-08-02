package std

import (
	"errors"
	"strings"
	"unicode/utf8"

	. "github.com/Heliodex/coputer/litecode/types"
	"golang.org/x/text/unicode/norm"
)

func iscont(c byte) bool {
	return !utf8.RuneStart(c)
}

// relative string position: negative means back from end
func utf8_posrelat(pos, len int) int {
	if pos >= 0 {
		return pos
	}
	if 0-pos > len {
		return 0
	}
	return len + pos + 1
}

func utf8_char(args Args) (r []Val, err error) {
	var b strings.Builder
	for range args.List {
		a := args.GetNumber()
		b.WriteRune(rune(a))
	}
	return []Val{b.String()}, nil
}

func iter_aux(args Args) (cps []Val, err error) {
	s := args.GetString()
	n := args.GetNumber() - 1

	len := float64(len(s))
	if n < 0 { // first iteration?
		n = 0 // start from here
	} else if n < len {
		n++ // skip current byte
		for n < len && iscont(s[uint(n)]) {
			n++ // and its continuations
		}
	}
	if n >= len {
		return // no more codepoints
	}

	r, _ := utf8.DecodeRuneInString(s[uint(n):]) // thx go stdlib
	if r == utf8.RuneError {
		return nil, errors.New("invalid UTF-8 code")
	}
	return []Val{float64(n + 1), float64(r)}, nil
}

func utf8_codes(args Args) (r []Val, err error) {
	str := args.GetString()

	return []Val{MakeFn("codes", iter_aux), str, float64(0)}, nil
}

const int_max = int(^uint(0) >> 1)

func utf8_codepoint(args Args) (cps []Val, err error) {
	s := args.GetString()
	i := args.GetNumber(1)
	j := args.GetNumber(i)

	posi, pose := utf8_posrelat(int(i), len(s)), utf8_posrelat(int(j), len(s))

	// luaL_argcheck(L, posi >= 1, 2, "out of range");
	// luaL_argcheck(L, pose <= (int)len, 3, "out of range");
	if posi > pose {
		return // empty interval; return no values
	}
	if pose-posi >= int_max { // (int -> int) overflow?
		return nil, errors.New("string slice too long")
	}
	// n := pose - posi + 1
	// luaL_checkstack(L, n, "string slice too long");
	// n := 0

	si := posi - 1

	for si < pose {
		s, size := utf8.DecodeRuneInString(s[si:])
		if s == utf8.RuneError {
			return nil, errors.New("invalid UTF-8 code")
		}
		si += size
		cps = append(cps, float64(s))
	}

	return
}

// roblox docs says this returns 1 number (incorrect)
func utf8_len(args Args) (r []Val, err error) {
	s := args.GetString()
	i := args.GetNumber(1)
	j := args.GetNumber(-1)

	posi, pose := utf8_posrelat(int(i), len(s)), utf8_posrelat(int(j), len(s))

	sl := s[max(posi-1, 0):min(pose, len(s))]
	n := utf8.RuneCountInString(sl)

	return []Val{float64(n)}, nil
}

func utf8_offset(args Args) (r []Val, err error) {
	s := args.GetString()
	n := args.GetNumber()

	var posi int
	if n >= 0 {
		posi = int(args.GetNumber(1))
	} else {
		posi = int(args.GetNumber(float64(len(s) + 1)))
	}
	posi = utf8_posrelat(posi, len(s)) - 1
	// luaL_argcheck(L, 1 <= posi && --posi <= (int)len, 3, "position out of range");

	if n == 0 {
		// find beginning of current byte sequence
		for posi > 0 && iscont(s[posi]) {
			posi--
		}
		return []Val{float64(posi + 1)}, nil
	}
	if iscont(s[posi]) {
		return nil, errors.New("initial position is a continuation byte")
	}
	if n < 0 {
		for ; n < 0 && posi > 0; posi-- { // move back
			// find beginning of previous character
			for posi > 0 && iscont(s[posi]) {
				posi--
			}
			n++
		}
	} else {
		n-- // do not move for 1st character
		for ; n > 0 && posi < len(s); posi++ {
			// find beginning of next character
			for iscont(s[posi]) {
				posi++
			}
			n--
		}
	}

	if n != 0 { // did it find a given character?
		// no such character
		return
	}
	return []Val{float64(posi + 1)}, nil
}

// func utf8_graphemes(args Args) (r Rets, err error) {
// 	panic("not implemented")
// }

func utf8_nfcnormalize(args Args) (r []Val, err error) {
	s := args.GetString()

	return []Val{norm.NFC.String(s)}, nil
}

func utf8_nfdnormalize(args Args) (r []Val, err error) {
	s := args.GetString()

	return []Val{norm.NFD.String(s)}, nil
}

var Libutf8 = NewLib([]Function{
	MakeFn("char", utf8_char),
	MakeFn("codes", utf8_codes),
	MakeFn("codepoint", utf8_codepoint),
	MakeFn("len", utf8_len),
	MakeFn("offset", utf8_offset),
	// MakeFn("graphemes", utf8_graphemes), // we can't actually test this, mainly due to the fact it... doesn't exist in the reference implementation..?
	MakeFn("nfcnormalize", utf8_nfcnormalize), // these are also untestable but they're so trivial here
	MakeFn("nfdnormalize", utf8_nfdnormalize),
}, map[string]Val{
	"charpattern": "[\x00-\x7F\xC2-\xF4][\x80-\xBF]*",
})
