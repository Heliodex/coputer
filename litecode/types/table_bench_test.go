package types

import (
	"strconv"
	"testing"
)

var keySink bool
var iterSum float64

// BenchmarkListKeyFloat measures the check every number-keyed table access runs to decide whether the key is a list index.
func BenchmarkListKeyFloat(b *testing.B) {
	i := 0
	for b.Loop() {
		i = i%1024 + 1
		_, keySink = listKeyFloat(float64(i))
	}
}

func BenchmarkListKeyNum(b *testing.B) {
	i := 0
	for b.Loop() {
		i = i%1024 + 1
		_, keySink = listKey(Num(float64(i)))
	}
}

// BenchmarkTableIterList measures a full walk of the list part, which pairs/next and generic-for rely on.
func BenchmarkTableIterList(b *testing.B) {
	t := &Table{List: make([]Val, 1024)}
	for i := range t.List {
		t.List[i] = Num(float64(i + 1))
	}

	for b.Loop() {
		sum := 0.0
		for _, v := range t.Iter() {
			f, _ := AsNum(v)
			sum += f
		}
		iterSum = sum
	}
}

// BenchmarkTableIterHash measures a full walk of the hash part, including the per-iteration key sort.
func BenchmarkTableIterHash(b *testing.B) {
	const n = 256
	t := &Table{Hash: make(map[Val]Val, n)}
	for i := range n {
		t.Set(Str(strconv.Itoa(i)), Num(float64(i)))
	}

	for b.Loop() {
		sum := 0.0
		for _, v := range t.Iter() {
			f, _ := AsNum(v)
			sum += f
		}
		iterSum = sum
	}
}
