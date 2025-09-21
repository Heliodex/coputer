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
	t.SetInt(1, "a")

	for b.Loop() {
		clone(t)
	}
}

func BenchmarkTableCloneListBig(b *testing.B) {
	t := &Table{}
	for i := range 100 {
		t.SetInt(i+1, i)
	}

	for b.Loop() {
		clone(t)
	}
}

func BenchmarkTableCloneHash(b *testing.B) {
	t := &Table{}
	t.Set("key", "value")

	for b.Loop() {
		clone(t)
	}
}

func BenchmarkTableCloneHashBig(b *testing.B) {
	t := &Table{}
	for i := range 100 {
		t.Set(fmt.Sprintf("key%d", i), i)
	}

	for b.Loop() {
		clone(t)
	}
}

func BenchmarkTableCloneMixed(b *testing.B) {
	t := &Table{}
	for i := range 100 {
		t.SetInt(i+1, i)
	}
	for i := range 100 {
		t.Set(fmt.Sprintf("key%d", i), i)
	}

	for b.Loop() {
		clone(t)
	}
}
