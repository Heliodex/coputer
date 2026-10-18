package std

import (
	"strings"
	"testing"

	. "github.com/Heliodex/coputer/litecode/types"
)

var findSink []Val

// renderResults formats a result list the way a Luau caller would see it, for compact assertions.
func renderResults(r []Val) string {
	parts := make([]string, len(r))
	for i, v := range r {
		parts[i] = ToString(v)
	}
	return strings.Join(parts, " ")
}

func findResults(s, p string) (string, error) {
	r, err := stringFindAux(s, p, 1, false, true)
	if err != nil {
		return "", err
	}
	return renderResults(r), nil
}

// TestPatternFindUnit pins the matcher's behaviour without needing to compile and diff against reference Luau, so regressions surface in a plain go test.
func TestPatternFindUnit(t *testing.T) {
	cases := []struct{ s, p, want string }{
		{"hello world", "o w", "5 7"},
		{"hello", "l+", "3 4"},
		{"hello", "(l+)", "3 4 ll"},
		{"hello", "()ll()", "3 4 3 5"},
		{"abc123", "%d+", "4 6"},
		{"abc", "^b", "nil"},
		{"abc", "c$", "3 3"},
		{"abc", "b$", "nil"},
		{"abc", "[a-c]+", "1 3"},
		{"abc", "a.c", "1 3"},
		{"(abc)", "%(abc%)", "1 5"},
		{"a%b", "%%", "2 2"},
		{"(nested (x))", "%b()", "1 12"},
		{"hello world", "%f[%w]world", "7 11"},
		{"abcabc", "(abc)%1", "1 6 abc"},
		{"hello", ".", "1 1"},
		{"a.b", "%.", "2 2"},
		{"abc", "", "1 0"},
		{"", "", "1 0"},
		{"aaa", "^a*$", "1 3"},
		{"aaa", "b*", "1 0"},
		{"aaab", "a*", "1 3"},
		{"aaab", "a-", "1 0"},
		{"abc", "%a+", "1 3"},
		{"abc", "%A+", "nil"},
		{"a b", "%s", "2 2"},
		{"Hello", "%u", "1 1"},
		{"Hello", "%l", "2 2"},
		{"abc", "[^b]+", "1 1"},
		{"abc", "b-", "1 0"},
		// '?' on a literal, both the taken and skipped paths of the inlined optional
		{"a", "a?", "1 1"},
		{"b", "a?", "1 0"},
		{"ab", "^a?b$", "1 2"},
		{"b", "^a?b$", "1 1"},
		{"abc", "^a?b?c$", "1 3"},
		{"ac", "^a?b?c$", "1 2"},
		{"bc", "^a?b?c$", "1 2"},
		{"xbc", "^a?b?c$", "nil"},
		// '$' as a literal rather than an anchor, and as the end anchor
		{"a$b", "a$", "nil"},
		{"a$a", "$", "4 3"},
		{"a$a", ".$", "3 3"},
		{"a$a", ".%$", "1 2"},
		{"a$a", ".$.", "1 3"},
		{"a$a", "$$", "nil"},
		// balanced match against an empty subject, which used to index out of range
		{"", "a*%b()", "nil"},
		{"()", "%b()", "1 2"},
		{"a(b)c", "a%b()c", "1 5"},
		{"deadbeef", "%x+", "1 8"},
		{"hello", "%z", "nil"},
		{"12345712301-helloworld8yeah", "([0-7%l%-]+)", "1 22 12345712301-helloworld"},
	}

	for _, c := range cases {
		got, err := findResults(c.s, c.p)
		if err != nil {
			t.Errorf("find(%q, %q): unexpected error %v", c.s, c.p, err)
			continue
		}
		if got != c.want {
			t.Errorf("find(%q, %q) = %q, want %q", c.s, c.p, got, c.want)
		}
	}
}

func gsubResults(s, p, repl string) (string, error) {
	r, err := string_gsub(Args{Name: "gsub", List: []Val{Str(s), Str(p), Str(repl)}})
	if err != nil {
		return "", err
	}
	return renderResults(r), nil
}

func TestPatternGsubUnit(t *testing.T) {
	cases := []struct{ s, p, repl, want string }{
		{"hello", "l+", "L", "heLo 1"},
		{"hello", "l", "L", "heLLo 2"},
		{"abc", "%a", "x", "xxx 3"},
		{"hello world", "(%w+)", "<%1>", "<hello> <world> 2"},
		{"abc", ".", "%0%0", "aabbcc 3"},
	}

	for _, c := range cases {
		got, err := gsubResults(c.s, c.p, c.repl)
		if err != nil {
			t.Errorf("gsub(%q, %q, %q): unexpected error %v", c.s, c.p, c.repl, err)
			continue
		}
		if got != c.want {
			t.Errorf("gsub(%q, %q, %q) = %q, want %q", c.s, c.p, c.repl, got, c.want)
		}
	}
}

func TestPatternErrors(t *testing.T) {
	cases := []struct{ s, p, want string }{
		{"abc", "[", "malformed pattern (missing ']')"},
		{"abc", "%", "malformed pattern (ends with '%')"},
		{"abc", "(a", "unfinished capture"},
		{"abc", "%2", "invalid capture index %2"},
	}

	for _, c := range cases {
		_, err := stringFindAux(c.s, c.p, 1, false, true)
		if err == nil {
			t.Errorf("find(%q, %q): expected error, got nil", c.s, c.p)
			continue
		}
		if err.Error() != c.want {
			t.Errorf("find(%q, %q) error = %q, want %q", c.s, c.p, err, c.want)
		}
	}
}

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
