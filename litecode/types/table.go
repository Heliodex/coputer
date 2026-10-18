package types

import (
	"fmt"
	"iter"
	"slices"
	"strings"
)

func listKeyFloat(fk float64) (int, bool) {
	ik := int(fk)
	return ik, 1 <= ik && float64(ik) == fk
}

func listKey(k Val) (ik int, ok bool) {
	fk, ok := AsNum(k)
	if !ok {
		return
	}

	return listKeyFloat(fk)
}

// orderedKey is a hash key paired with its printed form, so sorting compares strings rather than formatting on every comparison.
type orderedKey struct {
	k Val
	s string
}

func newOrderedKey(k Val) orderedKey {
	return orderedKey{k: k, s: fmt.Sprint(k)}
}

func keyKind(v Val) int {
	switch {
	case IsNil(v):
		return 0
	case IsBool(v):
		return 1
	case IsNum(v):
		return 2
	case IsInt(v):
		return 3
	case IsString(v):
		return 4
	}
	return 5
}

func compareOrderedKeys(a, b orderedKey) int {
	// It doesn't have to be pretty for map keys
	// (in fact, the reference implementation of Luau has a rather insane sort order)
	// It just has to be DETERMINISTIC
	if c := strings.Compare(a.s, b.s); c != 0 {
		return c
	}
	// Printed forms can collide (the number 1 and the string "1"), so break ties by kind to keep the order total for binary search.
	return keyKind(a.k) - keyKind(b.k)
}

// orderedKeys returns the hash part's deterministic key order, rebuilding it only after the hash changes.
func (t *Table) orderedKeys() []orderedKey {
	if t.Hash == nil {
		return nil
	}
	if t.orderedHash != nil {
		return t.orderedHash
	}

	keys := make([]orderedKey, 0, len(t.Hash))
	for k := range t.Hash {
		keys = append(keys, newOrderedKey(k))
	}
	slices.SortFunc(keys, compareOrderedKeys)

	t.orderedHash = keys
	return keys
}

func iterList(l []Val, y func(Val, Val) bool) (completed bool) {
	for i, v := range l {
		if !IsNil(v) && !y(Num(float64(i+1)), v) {
			return false
		}
	}
	return true
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
	// orderedHash caches the hash part's deterministic key order. It is nil when it must be rebuilt.
	orderedHash []orderedKey
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
	t.orderedHash = nil
	if t.Hash == nil {
		if IsNil(v) {
			return
		}
		t.Hash = map[Val]Val{k: v}
	} else if IsNil(v) {
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
	t.orderedHash = nil

	for f2 := float64(l + 2); ; f2++ {
		v2, ok := t.Hash[Num(f2)]
		if !ok {
			break
		}
		t.List = append(t.List, v2)
		delete(t.Hash, Num(f2))
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
	} else if l := len(t.List); 0 < i && i < l+1 {
		if !IsNil(v) {
			// set in the list portion
			t.List[i-1] = v
			return
		}

		// cut the list portion
		after := t.List[i:]
		t.List = t.List[:i-1]

		// move the rest to the hash part
		for i2, v2 := range after {
			t.setHash(Num(float64(i+i2)), v2)
		}
		return
	} else if i == l+1 {
		// append to the end
		t.List = append(t.List, v)

		t.moveToList(l)
		return
	}

	// add to the hash part instead
	t.setHash(Num(float64(i)), v)
}

// Set sets a table value at a key. Make sure to check if the table is readonly beforehand.
func (t *Table) Set(k Val, v Val) {
	if fk, ok := AsNum(k); ok {
		if ik := int(fk); 1 <= ik && float64(ik) == fk {
			if l := len(t.List); ik <= l && !IsNil(v) {
				// fast path: overwrite an existing list element in place
				t.List[ik-1] = v
				return
			}
			t.SetInt(ik, v)
			return
		}
	}
	t.setHash(k, v)
}

// Clone returns a shallow copy of the table, mirroring the reference implementation's luaH_clone used to instantiate DUPTABLE templates.
func (t *Table) Clone() *Table {
	n := &Table{}
	if t.List != nil {
		n.List = slices.Clone(t.List)
	}
	if t.Hash != nil {
		n.Hash = make(map[Val]Val, len(t.Hash))
		for k, v := range t.Hash {
			n.Hash[k] = v
		}
	}
	return n
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
	if fk, ok := AsNum(k); ok {
		if ik := int(fk); 1 <= ik && float64(ik) == fk {
			if ik <= len(t.List) {
				// fast path: list element
				return t.List[ik-1]
			}
		}
	}
	return t.GetHash(k)
}

// GetFloat returns a value at a float key in the table.
func (t *Table) GetFloat(k float64) Val {
	if ak, ok := listKeyFloat(k); ok && ak <= t.Len() {
		return t.List[ak-1]
	}
	return t.GetHash(Num(k))
}

// GetFloat returns a value at a float key in the table.
func (t *Table) GetInt(k int) Val {
	if ok := 1 <= k; ok && k <= t.Len() {
		return t.List[k-1]
	}
	return t.GetHash(Num(float64(k)))
}

// Iter returns an iterator over the table, yielding key-value pairs in a deterministic order.
func (t *Table) Iter() iter.Seq2[Val, Val] {
	return func(y func(Val, Val) bool) {
		if t.List != nil && !iterList(t.List, y) {
			return
		}
		if t.Hash != nil {
			for _, hk := range t.orderedKeys() {
				if !y(hk.k, t.Hash[hk.k]) {
					return
				}
			}
		}
	}
}

// Next returns the pair after fk in the table's iteration order, starting from the first pair when fk is nil. found reports whether fk was nil or an actual key, and ok whether a pair was returned.
func (t *Table) Next(fk Val) (k, v Val, found, ok bool) {
	from := 0
	found = IsNil(fk)
	if !found {
		if i, isList := listKey(fk); isList && i <= len(t.List) && !IsNil(t.List[i-1]) {
			found = true
			from = i // 0-based index i is the element after the key at i-1
		} else {
			return t.nextHash(fk)
		}
	}

	for j := from; j < len(t.List); j++ {
		if !IsNil(t.List[j]) {
			return Num(float64(j + 1)), t.List[j], true, true
		}
	}
	return t.nextHash(NilVal)
}

// nextHash returns the pair after fk in the sorted hash part, or the first pair when fk is nil.
func (t *Table) nextHash(fk Val) (k, v Val, found, ok bool) {
	keys := t.orderedKeys()
	if IsNil(fk) {
		if len(keys) == 0 {
			return NilVal, NilVal, true, false
		}
		k = keys[0].k
		return k, t.Hash[k], true, true
	}

	idx, exact := slices.BinarySearchFunc(keys, newOrderedKey(fk), compareOrderedKeys)
	if !exact {
		return NilVal, NilVal, false, false
	}
	if idx+1 == len(keys) {
		return NilVal, NilVal, true, false
	}
	k = keys[idx+1].k
	return k, t.Hash[k], true, true
}
