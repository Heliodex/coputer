package types

import "testing"

var keySink bool

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
