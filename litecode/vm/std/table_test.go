package std

import (
	"fmt"
	"testing"

	. "github.com/Heliodex/coputer/litecode/types"
)

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
