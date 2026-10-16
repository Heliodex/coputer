package std

import (
	"strings"
	"testing"

	. "github.com/Heliodex/coputer/litecode/types"
)

var findSink []Val

// benchFind runs stringFindAux (the engine behind string.find) over a 1 MiB string of 'a's, which is what the bigstring script stresses.
func benchFind(b *testing.B, p string) {
	s := strings.Repeat("a", 1<<20)
	b.SetBytes(int64(len(s)))
	for b.Loop() {
		r, err := stringFindAux(s, p, 1, false, true)
		if err != nil {
			b.Fatal(err)
		}
		findSink = r
	}
}

// Anchored greedy star that succeeds: expands to the whole run, then backs off once.
func BenchmarkFindAnchoredStar(b *testing.B) { benchFind(b, "^a*.?$") }

// Anchored greedy star that fails: expands to the whole run, then backs off over every position.
func BenchmarkFindAnchoredStarMiss(b *testing.B) { benchFind(b, "^a*.?b$") }

// Anchored lazy star: expands one byte at a time from the start.
func BenchmarkFindAnchoredLazy(b *testing.B) { benchFind(b, "^a-.?$") }

// No special characters, so it takes the plain search path.
func BenchmarkFindPlain(b *testing.B) { benchFind(b, "b") }

// A capture forces the pattern engine and pushCaptures.
func BenchmarkFindCaptures(b *testing.B) {
	s := strings.Repeat("ab", 1<<16)
	for b.Loop() {
		r, err := stringFindAux(s, "a(b)a", 1, false, true)
		if err != nil {
			b.Fatal(err)
		}
		findSink = r
	}
}
