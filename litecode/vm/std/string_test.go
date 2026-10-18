package std

import (
	"strings"
	"testing"

	. "github.com/Heliodex/coputer/litecode/types"
)

var stringSink []Val

// These call the native string functions directly. Each returns a freshly allocated []Val, which is the per-call allocation the VM then copies into registers.

func BenchmarkStringByte(b *testing.B) {
	s := strings.Repeat("abcdefghij", 100)
	args := Args{Co: &Coroutine{}, Name: "byte", List: []Val{Str(s), Num(1), Num(10)}}

	for b.Loop() {
		args.pos = 0
		r, err := string_byte(args)
		if err != nil {
			b.Fatal(err)
		}
		stringSink = r
	}
}

func BenchmarkStringChar(b *testing.B) {
	args := Args{Co: &Coroutine{}, Name: "char", List: []Val{Num(104), Num(105), Num(33)}}

	for b.Loop() {
		args.pos = 0
		r, err := string_char(args)
		if err != nil {
			b.Fatal(err)
		}
		stringSink = r
	}
}

func BenchmarkStringFormat(b *testing.B) {
	args := Args{Co: &Coroutine{}, Name: "format", List: []Val{Str("%d:%s:%.3f"), Num(42), Str("x"), Num(1.5)}}

	for b.Loop() {
		args.pos = 0
		r, err := string_format(args)
		if err != nil {
			b.Fatal(err)
		}
		stringSink = r
	}
}

func BenchmarkStringSub(b *testing.B) {
	s := strings.Repeat("abcdefghij", 100)
	args := Args{Co: &Coroutine{}, Name: "sub", List: []Val{Str(s), Num(10), Num(20)}}

	for b.Loop() {
		args.pos = 0
		r, err := string_sub(args)
		if err != nil {
			b.Fatal(err)
		}
		stringSink = r
	}
}
