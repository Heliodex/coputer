package main

import "strings"

func table_clear(args Args) {
	t := args.GetTable()
	if t.readonly {
		panic("attempt to modify a readonly table")
	}

	if t.array != nil {
		for i := range t.asize {
			(*t.array)[i] = nil
		}
	}
	if t.hash != nil {
		for k := range *t.hash {
			(*t.hash)[k] = nil
		}
	}
}

func table_clone(args Args) Ret {
	t := args.GetTable()

	a2 := make([]any, len(*t.array))
	copy(a2, *t.array)

	h2 := map[any]any{}
	for k, v := range *t.hash {
		h2[k] = v
	}

	return &Table{
		array: &a2,
		hash:  &h2,
		asize: t.asize,
	}
}

func table_concat(args Args) Ret {
	t := args.GetTable()
	sep := args.GetString("")
	i := args.GetNumber(1)
	j := args.GetNumber(t.Len())

	if i > j {
		return ""
	}

	b := strings.Builder{}
	for ; i <= j; i++ {
		v, ok := t.Get(i).(string)
		if !ok {
			panic("attempt to concatenate non-string value")
		}

		b.WriteString(v)
		if i < j {
			b.WriteString(sep)
		}
	}

	return b.String()
}

func table_create(args Args) Ret {
	s := args.GetNumber()
	if s < 0 {
		panic("index out of range")
	}

	asize := uint(s)
	array := make([]any, asize)
	if len(args.args) > 1 {
		value := args.GetAny()
		for i := range array {
			array[i] = value
		}
	}

	return &Table{
		array: &array,
		hash:  &map[any]any{},
		asize: asize, // not ^2?
	}
}

func table_find(args Args) Ret {
	haystack := args.GetTable()
	needle := args.GetAny()
	init := args.GetNumber(1)
	if init < 1 {
		panic("index out of range")
	}

	arr := *haystack.array
	for i := uint(init - 1); ; i++ {
		v := arr[i]
		if v == nil {
			return nil
		} else if v == needle {
			return float64(i + 1)
		}
	}
}

func table_freeze(args Args) Ret {
	t := args.GetTable()
	t.readonly = true
	return t
}

func table_insert(args Args) {
	t := args.GetTable()
	args.CheckNextArg()
	if t.readonly {
		panic("attempt to modify a readonly table")
	}

	l := t.Len()
	if len(args.args) == 2 {
		value := args.GetAny()
		if t.array == nil {
			t.array = &[]any{value}
			t.asize = 1
			return
		}

		t.SetArray(uint(l+1), value)
		return
	}

	pos, value := int(args.GetNumber()), args.GetAny()

	if t.array == nil {
		arr := make([]any, pos)
		arr[pos-1] = value
		t.array = &arr
		t.asize = uint(pos)
	} else if pos > int(t.asize) {
		for j := int(l); j >= pos; j-- {
			(*t.array)[j] = (*t.array)[j-1]
		}

		t.Rehash(float64(pos), value)
	} else if 1 <= pos {
		for j := int(l); j >= int(pos); j-- {
			if j >= len(*t.array) {
				// we may ought to rehash here
				(*t.array) = append(*t.array, (*t.array)[j-1])
			}
			(*t.array)[j] = (*t.array)[j-1]
		}

		(*t.array)[pos-1] = value
	} else {
		t.SetHash(float64(pos), value)
	}
}

func table_isfrozen(args Args) Ret {
	t := args.GetTable()
	return t.readonly
}

func table_maxn(args Args) Ret {
	t := args.GetTable()

	var lenArray, lenHash uint
	arrayExists, hashExists := t.array != nil, t.hash != nil

	if arrayExists {
		lenArray = uint(len(*t.array))
	}
	if hashExists {
		lenHash = uint(len(*t.hash))
	}

	nentries := make(map[float64]bool, lenArray+lenHash)

	// array kvs
	if arrayExists {
		for i, v := range *t.array {
			if v != nil {
				nentries[float64(i+1)] = true
			}
		}
	}

	// hash kvs
	if hashExists {
		for k, v := range *t.hash {
			if fk, ok := k.(float64); ok && v != nil {
				nentries[fk] = true
			}
		}
	}

	var maxn float64
	for k := range nentries {
		if k > maxn {
			maxn = k
		}
	}

	return maxn
}

func table_move(args Args) Ret {
	src := args.GetTable()
	a, b, t := args.GetNumber(), args.GetNumber(), args.GetNumber()
	dst := args.GetTable(src)
	if dst.readonly {
		panic("attempt to modify a readonly table")
	}

	for i := a; i <= b; i++ {
		dst.ForceSet(t+i-a, src.Get(i))
	}

	return dst
}

func table_pack(args Args) Ret {
	n := float64(len(args.args))
	t := &Table{
		array: &[]any{},
		hash:  &map[any]any{"n": n},
	}
	for i, v := range args.args {
		t.SetArray(uint(i)+1, v)
	}

	return t
}

func table_remove(args Args) (r Ret) {
	t := args.GetTable()
	if t.readonly {
		panic("attempt to modify a readonly table")
	}

	l := t.Len()
	pos := args.GetNumber(l)

	r = t.Get(pos)
	if uint(pos) == uint(l) {
		t.ForceSet(pos, nil)
	} else if 0 < pos && pos < l {
		for i := pos; i < l; i++ {
			t.ForceSet(i, t.Get(i+1))
		}
		t.ForceSet(l, nil)
	}
	return
}

// ltablib.cpp
type Comp func(a, b any) bool

func sort_swap(t *Table, i, j int) {
	arr := *t.array
	// n := t.asize
	// LUAU_ASSERT(unsigned(i) < unsigned(n) && unsigned(j) < unsigned(n)); // contract maintained in sort_less after predicate call

	// no barrier required because both elements are in the array before and after the swap
	arr[i], arr[j] = arr[j], arr[i]
}

func sort_less(t *Table, i, j int, comp Comp) (res bool) {
	arr, n := *t.array, t.asize
	// LUAU_ASSERT(unsigned(i) < unsigned(n) && unsigned(j) < unsigned(n)); // contract maintained in sort_less after predicate call

	res = comp(arr[i], arr[j])

	// predicate call may resize the table, which is invalid
	if t.asize != n {
		panic("table modified during sorting")
	}
	return
}

func sort_siftheap(t *Table, l, u int, comp Comp, root int) {
	// LUAU_ASSERT(l <= u);
	count := u - l + 1

	// process all elements with two children
	for root*2+2 < count {
		left, right := root*2+1, root*2+2
		next := root
		if sort_less(t, l+next, l+left, comp) {
			next = left
		}
		if sort_less(t, l+next, l+right, comp) {
			next = right
		}

		if next == root {
			break
		}

		sort_swap(t, l+root, l+next)
		root = next
	}

	// process last element if it has just one child
	if lastleft := root*2 + 1; lastleft == count-1 && sort_less(t, l+root, l+lastleft, comp) {
		sort_swap(t, l+root, l+lastleft)
	}
}

func sort_heap(t *Table, l, u int, comp Comp) {
	// LUAU_ASSERT(l <= u);
	count := u - l + 1

	for i := count/2 - 1; i >= 0; i-- {
		sort_siftheap(t, l, u, comp, i)
	}

	for i := count - 1; i > 0; i-- {
		sort_swap(t, l, l+i)
		sort_siftheap(t, l, l+i-1, comp, 0)
	}
}

func sort_rec(t *Table, l, u, limit int, comp Comp) {
	// sort range [l..u] (inclusive, 0-based)
	for l < u {
		// if the limit has been reached, quick sort is going over the permitted nlogn complexity, so we fall back to heap sort
		if limit == 0 {
			sort_heap(t, l, u, comp)
			return
		} else if sort_less(t, u, l, comp) { // a[u] < a[l]?
			// sort elements a[l], a[(l+u)/2] and a[u]
			// note: this simultaneously acts as a small sort and a median selector
			sort_swap(t, u, l) // swap a[l] - a[u]
		}
		if u-l == 1 {
			break // only 2 elements
		}

		m := l + ((u - l) >> 1)       // midpoint
		if sort_less(t, m, l, comp) { // a[m]<a[l]?
			sort_swap(t, m, l)
		} else if sort_less(t, u, m, comp) { // a[u]<a[m]?
			sort_swap(t, m, u)
		}
		if u-l == 2 {
			break // only 3 elements
		}

		// here l, m, u are ordered; m will become the new pivot
		p := u - 1
		sort_swap(t, m, u-1) // pivot is now (and always) at u-1

		// a[l] <= P == a[u-1] <= a[u], only need to sort from l+1 to u-2
		i := l
		j := u - 1
		for {
			// invariant: a[l..i] <= P <= a[j..u]
			// repeat ++i until a[i] >= P
			i++
			for sort_less(t, i, p, comp) {
				if i >= u {
					panic("invalid order function for sorting")
				}
				i++
			}

			// repeat --j until a[j] <= P
			j--
			for sort_less(t, p, j, comp) {
				if j <= l {
					panic("invalid order function for sorting")
				}
				j--
			}
			if j < i {
				break
			}
			sort_swap(t, i, j)
		}

		// swap pivot a[p] with a[i], which is the new midpoint
		sort_swap(t, p, i)

		// adjust limit to allow 1.5 log2N recursive steps
		limit = (limit >> 1) + (limit >> 2)

		// a[l..i-1] <= a[i] == P <= a[i+1..u]
		// sort smaller half recursively; the larger half is sorted in the next loop iteration
		if i-l < u-i {
			sort_rec(t, l, i-1, limit, comp)
			l = i + 1
		} else {
			sort_rec(t, i+1, u, limit, comp)
			u = i - 1
		}
	}
}

func table_sort(args Args) {
	t := args.GetTable()
	if t.readonly {
		panic("attempt to modify a readonly table")
	}

	var comp Comp
	if len(args.args) == 1 {
		comp = jumpLt
	} else {
		fn := args.GetFunction()
		comp = func(a, b any) bool {
			res := (*fn)(args.co, a, b)
			return res[0].(bool)
		}
	}

	if n := int(t.Len()); n > 0 {
		sort_rec(t, 0, n-1, n, comp)
	}
}

func table_unpack(args Args) (values Rets) {
	list := args.GetTable()
	i := args.GetNumber(1)
	j := args.GetNumber(list.Len())

	ui, uj := uint(i), uint(j)
	if uj <= list.asize {
		return (*list.array)[ui-1 : uj]
	}

	values = make([]any, uj-ui+1)
	for k := i; k <= j; k++ {
		values[uint(k)-ui] = list.Get(k)
	}

	return
}

var libtable = NewTable([][2]any{
	MakeFn0("clear", table_clear),
	MakeFn1("clone", table_clone),
	MakeFn1("concat", table_concat),
	MakeFn1("create", table_create),
	MakeFn1("find", table_find),
	MakeFn1("freeze", table_freeze),
	MakeFn0("insert", table_insert),
	MakeFn1("isfrozen", table_isfrozen),
	MakeFn1("maxn", table_maxn),
	MakeFn1("move", table_move),
	MakeFn1("pack", table_pack),
	MakeFn1("remove", table_remove),
	MakeFn0("sort", table_sort),
	MakeFn("unpack", table_unpack),
})
