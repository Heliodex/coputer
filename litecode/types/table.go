package types

import (
	"fmt"
	"iter"
	"slices"
	"strings"
)

func listKey(k Val) (int, bool) {
	fk, ok := k.(float64)
	if !ok {
		return 0, false
	}

	ik := int(fk)
	return ik, 1 <= ik && float64(ik) == fk
}

func mapKeySort(a, b Val) int {
	// It doesn't have to be pretty for map keys
	// (in fact, the reference implementation of Luau has a rather insane sort order)
	// It just has to be DETERMINISTIC
	return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
}

func iterList(l []Val, y func(Val, Val) bool) {
	for i, v := range l {
		if v != nil && !y(float64(i+1), v) {
			return
		}
	}
}

func iterHash(m map[Val]Val, y func(Val, Val) bool) {
	// order keys in map
	keys := make([]Val, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, mapKeySort)
	for _, k := range keys {
		if !y(k, m[k]) {
			return
		}
	}
}

// Q why are tables like this
// A:
// 1: the reference implementation of tables is too complex: rehashing and resizing is a pain but not too bad, list boundaries are worse and I don't want 1.5k lines of code just for that, and Go does a resizing-like thing automatically with slices anyway
// 2: the way nodes are implemented works well in C++ and not in Go (plus I don't know if it's actually O(1) for node lookups??)
// 3: rehashing etc is slower than just using a slice... somehow. most of this program is between 10-20x slower than the reference implementation, but the tables (which were previously like 50x slower) are now only like 2-3x slower for large allocations (bench/largealloc.luau)
// 4: having a list part is actually nice for iteration and for large tables (as opposed to the lua4 way, where it's *just* a hash part), the way it's done here is simpler though we have to move stuff around and between the list and node parts more explicitly
// 5: very weird quirks arise from table length implementations etc. the nil stuff can easily be forgiven, it's the stuff with creating a table and getting a length afterwards (see tests/clear.luau) that is fucking devilish; this is one of the few parts that puts Luau, as the language at the top of my favourites list, in jeopardy
// 6: we don't actually break *that* much compatibility doing it this way, right??
// 7: if anyone tells you tables are simple THEY ARE LYING, CALL THEM OUT ON THEIR SHIT

// Table represents a Luau table, with resizeable list and hash parts. Luau type `table`
// As tables are compared by reference, this type must always be used as a pointer.
type Table struct {
	List     []Val
	Hash     map[Val]Val
	Readonly bool
}

// Len returns the length of the list part of the table (the length of the list up until the first nil).
func (t *Table) Len() int {
	if t.List == nil {
		return 0
	}
	return len(t.List)
}

// setHash updates or deletes a key-value pair in the hash part of the table.
func (t *Table) setHash(k Val, v Val) {
	if t.Hash == nil {
		if v == nil {
			return
		}
		t.Hash = map[Val]Val{k: v}
	} else if v == nil {
		delete(t.Hash, k)
	} else {
		t.Hash[k] = v
	}
}

// check if we can move some stuff from the hash part to the list part
func (t *Table) moveToList(l int) {
	if t.Hash == nil {
		return
	}

	for f2 := float64(l + 2); ; f2++ {
		v2, ok := t.Hash[f2]
		if !ok {
			break
		}
		t.List = append(t.List, v2)
		delete(t.Hash, f2)
	}
}

// SetInt sets a value at an integer index, placing it into the List part or the Hash part and resizing each as appropriate.
func (t *Table) SetInt(i int, v Val) {
	// fmt.Println("SetInt", i, v)

	if t.List == nil {
		if i == 1 {
			t.List = []Val{v}

			t.moveToList(0)
			return
		}
	} else if l := len(t.List); i < l+1 {
		if v != nil {
			// set in the list portion
			t.List[i-1] = v
			return
		}

		// cut the list portion
		after := t.List[i:]
		t.List = t.List[:i-1]

		// move the rest to the hash part
		for i2, v2 := range after {
			t.setHash(float64(i+i2), v2)
		}
		return
	} else if i == l+1 {
		// append to the end
		t.List = append(t.List, v)

		t.moveToList(l)
		return
	}

	// add to the hash part instead
	t.setHash(float64(i), v)
}

// Set sets a table value at a key. Make sure to check if the table is readonly beforehand.
func (t *Table) Set(k Val, v Val) {
	if ak, ok := listKey(k); ok {
		t.SetInt(ak, v)
		return
	}
	t.setHash(k, v)
}

// GetHash returns a value at a key, only searching the hash part of the table.
func (t *Table) GetHash(k Val) (v Val) {
	if t.Hash == nil {
		return
	}
	return t.Hash[k]
}

// Get returns a value at a key in the table.
func (t *Table) Get(k Val) Val {
	if ak, ok := listKey(k); ok && ak <= t.Len() {
		return t.List[ak-1]
	}
	return t.GetHash(k)
}

// Iter returns an iterator over the table, yielding key-value pairs in a deterministic order.
func (t *Table) Iter() iter.Seq2[Val, Val] {
	return func(y func(Val, Val) bool) {
		if t.List != nil {
			iterList(t.List, y)
		}
		if t.Hash != nil {
			iterHash(t.Hash, y)
		}
	}
}
