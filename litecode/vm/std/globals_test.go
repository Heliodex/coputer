package std

import (
	"math"
	"strconv"
	"testing"

	. "github.com/Heliodex/coputer/litecode/types"
)

func TestMul128(t *testing.T) {
	const x = 100
	const y = 0x11111111_11111111

	rx, ry := mul128(x, y)

	if rx != 6 || ry != 0xaaaaaaaa_aaaaaaa4 {
		t.Fatalf("mul128 failed, got %x, %x", rx, ry)
	}
}

func BenchmarkMul128hi(b *testing.B) {
	const x = 100
	const y = 0x11111111_11111111

	for b.Loop() {
		mul128hi(x, y)
	}
}

func BenchmarkMul128(b *testing.B) {
	const x = 100
	const y = 0x11111111_11111111

	for b.Loop() {
		mul128(x, y)
	}
}

func TestMul192(t *testing.T) {
	const xhi = 100
	const xlo = 0x11111111_11111111
	const y = 0x22222222_22222222

	rx, ry := mul192hi(xhi, xlo, y)

	if rx != 0xd || ry != 0x579be024_68acf128 {
		t.Fatalf("mul128 failed, got %x, %x", rx, ry)
	}
}

func BenchmarkMul192hi(b *testing.B) {
	const xhi = 100
	const xlo = 0x11111111_11111111
	const y = 0x22222222_22222222

	for b.Loop() {
		mul192hi(xhi, xlo, y)
	}
}

func BenchmarkSchubfach1(b *testing.B) {
	for b.Loop() {
		schubfach(1230, 2163113714887186)
	}
}

func BenchmarkNum2Str1(b *testing.B) {
	for n := float64(1234567890123456789); b.Loop(); {
		num2str(n)
	}
}

func BenchmarkNum2Str2(b *testing.B) {
	for n := math.Pow(11, 60); b.Loop(); {
		num2str(n)
	}
}

func BenchmarkNum2Str3(b *testing.B) {
	for n := math.Inf(1); b.Loop(); {
		num2str(n)
	}
}

var nextSink []Val

// BenchmarkNextFirst measures next(t, nil): one step to the first pair.
func BenchmarkNextFirst(b *testing.B) {
	t := &Table{List: make([]Val, 256)}
	for i := range t.List {
		t.List[i] = Num(float64(i + 1))
	}
	args := Args{Name: "next", List: []Val{TableVal(t), NilVal}}

	for b.Loop() {
		args.pos = 0
		r, err := global_next(args)
		if err != nil {
			b.Fatal(err)
		}
		nextSink = r
	}
}

// BenchmarkNextLast measures next(t, last) on the list part.
func BenchmarkNextLast(b *testing.B) {
	t := &Table{List: make([]Val, 256)}
	for i := range t.List {
		t.List[i] = Num(float64(i + 1))
	}
	args := Args{Name: "next", List: []Val{TableVal(t), Num(256)}}

	for b.Loop() {
		args.pos = 0
		r, err := global_next(args)
		if err != nil {
			b.Fatal(err)
		}
		nextSink = r
	}
}

// BenchmarkNextHash measures next(t, key) on the hash part, which has to locate the key in the ordered hash.
func BenchmarkNextHash(b *testing.B) {
	t := &Table{Hash: make(map[Val]Val, 256)}
	for i := range 256 {
		t.Set(Str(strconv.Itoa(i)), Num(float64(i)))
	}
	args := Args{Name: "next", List: []Val{TableVal(t), Str("128")}}

	for b.Loop() {
		args.pos = 0
		r, err := global_next(args)
		if err != nil {
			b.Fatal(err)
		}
		nextSink = r
	}
}
