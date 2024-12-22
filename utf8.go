package main

import (
	"strings"
	"unicode/utf8"
)

func iscont(c byte) bool {
	return c&0xC0 == 0x80
}

func utf8_char(args Args) Ret {
	// panic("not implemented")
	b := strings.Builder{}

	for range args.args {
		a := args.GetNumber()
		b.WriteRune(rune(a))
	}
	return b.String()
}

func iter_aux(_ *Coroutine, args ...any) (cps []any) {
	a := args[0].([]any)
	s := a[0].(string)
	n := a[1].(float64) - 1

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
		panic("invalid UTF-8 code")
	}
	return []any{float64(n + 1), float64(r)}
}

func utf8_codes(args Args) Rets {
	// panic("not implemented")
	str := args.GetString()

	fn := Function(iter_aux)
	return Rets{&fn, str, float64(0)}
}

func utf8_codepoint(args Args) Rets {
	panic("not implemented")
}

func utf8_len(args Args) Ret {
	panic("not implemented")
}

func utf8_offset(args Args) Ret {
	panic("not implemented")
}

func utf8_graphemes(args Args) Ret {
	panic("not implemented")
}

func utf8_nfcnormalize(args Args) Ret {
	panic("not implemented")
}

func utf8_nfdnormalize(args Args) Ret {
	panic("not implemented")
}

var libutf8 = NewTable([][2]any{
	MakeFn1("char", utf8_char),
	MakeFn("codes", utf8_codes),
	MakeFn("codepooint", utf8_codepoint),
	MakeFn1("len", utf8_len),
	MakeFn1("offset", utf8_offset),
	MakeFn1("graphemes", utf8_graphemes),
	MakeFn1("nfcnormalize", utf8_nfcnormalize),
	MakeFn1("nfdnormalize", utf8_nfdnormalize),
})
