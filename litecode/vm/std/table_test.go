package std

import (
	"fmt"
	"strconv"
	"testing"

	. "github.com/Heliodex/coputer/litecode/types"
)

var tableSink Val

func BenchmarkEmptyTableClone(b *testing.B) {
	t := &Table{}

	for b.Loop() {
		clone(t)
	}
}

func BenchmarkTableCloneList(b *testing.B) {
	t := &Table{}
	t.SetInt(1, Str("a"))

	for b.Loop() {
		clone(t)
	}
}

func BenchmarkTableCloneListBig(b *testing.B) {
	t := &Table{}
	for i := range 100 {
		t.SetInt(i+1, Num(float64(i)))
	}

	for b.Loop() {
		clone(t)
	}
}

func BenchmarkTableCloneHash(b *testing.B) {
	t := &Table{}
	t.Set(Str("key"), Str("value"))

	for b.Loop() {
		clone(t)
	}
}

func BenchmarkTableCloneHashBig(b *testing.B) {
	t := &Table{}
	for i := range 100 {
		t.Set(Str(fmt.Sprintf("key%d", i)), Num(float64(i)))
	}

	for b.Loop() {
		clone(t)
	}
}

func BenchmarkTableCloneMixed(b *testing.B) {
	t := &Table{}
	for i := range 100 {
		t.SetInt(i+1, Num(float64(i)))
	}
	for i := range 100 {
		t.Set(Str(fmt.Sprintf("key%d", i)), Num(float64(i)))
	}

	for b.Loop() {
		clone(t)
	}
}

const tableBenchN = 1024

// BenchmarkTableSetIntList hits SetInt directly, the way SETTABLEN and list appends do.
func BenchmarkTableSetIntList(b *testing.B) {
	t := &Table{List: make([]Val, tableBenchN)}
	i := 0
	for b.Loop() {
		i = i%tableBenchN + 1
		t.SetInt(i, Num(float64(i)))
	}
}

// BenchmarkTableGetIntList hits GetInt directly.
func BenchmarkTableGetIntList(b *testing.B) {
	t := &Table{List: make([]Val, tableBenchN)}
	i := 0
	for b.Loop() {
		i = i % tableBenchN
		tableSink = t.GetInt(i)
	}
}

// BenchmarkTableSetNumKey goes through Set and the listKey path, the way SETTABLE with a number key does.
func BenchmarkTableSetNumKey(b *testing.B) {
	t := &Table{List: make([]Val, tableBenchN)}
	i := 0
	for b.Loop() {
		i = i%tableBenchN + 1
		t.Set(Num(float64(i)), Num(float64(i)))
	}
}

// BenchmarkTableGetNumKey goes through Get and the listKey path.
func BenchmarkTableGetNumKey(b *testing.B) {
	t := &Table{List: make([]Val, tableBenchN)}
	i := 0
	for b.Loop() {
		i = i % tableBenchN
		tableSink = t.Get(Num(float64(i)))
	}
}

func BenchmarkTableSetHash(b *testing.B) {
	const n = 256
	keys := make([]Val, n)
	for i := range keys {
		keys[i] = Str(strconv.Itoa(i))
	}
	t := &Table{Hash: make(map[Val]Val, n)}
	i := 0
	for b.Loop() {
		i = i % n
		t.Set(keys[i], Num(float64(i)))
	}
}

func BenchmarkTableGetHash(b *testing.B) {
	const n = 256
	keys := make([]Val, n)
	for i := range keys {
		keys[i] = Str(strconv.Itoa(i))
	}
	t := &Table{Hash: make(map[Val]Val, n)}
	for i := range n {
		t.Set(keys[i], Num(float64(i)))
	}
	i := 0
	for b.Loop() {
		i = i % n
		tableSink = t.Get(keys[i])
	}
}

// BenchmarkTableCreate fills a fresh table the way table.create(size, value) does.
func BenchmarkTableCreate(b *testing.B) {
	const n = 1 << 20
	args := Args{List: []Val{Num(n), Num(0)}}
	for b.Loop() {
		args.pos = 0
		r, err := table_create(args)
		if err != nil {
			b.Fatal(err)
		}
		tableSink = r[0]
	}
}
