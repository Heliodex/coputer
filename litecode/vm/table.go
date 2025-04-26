package vm

import (
	"errors"
	"strings"
)

func table_clear(args Args) (r []Val, err error) {
	t := args.GetTable()
	if t.readonly {
		return nil, errors.New("attempt to modify a readonly table")
	}

	t.Array = nil
	t.Hash = nil
	return
}

func table_clone(args Args) (r []Val, err error) {
	t := args.GetTable()

	nt := &Table{}

	if t.Array != nil {
		a := make([]Val, len(t.Array))
		copy(a, t.Array)
		nt.Array = a
	}

	if t.Hash != nil {
		h := make(map[Val]Val, len(t.Hash))
		for k, v := range t.Hash {
			h[k] = v
		}
		nt.Hash = h
	}

	return []Val{nt}, nil
}

func table_concat(args Args) (r []Val, err error) {
	t := args.GetTable()
	sep := args.GetString("")
	i := args.GetNumber(1)
	j := args.GetNumber(float64(t.Len()))

	if i > j {
		return []Val{""}, nil
	}

	b := strings.Builder{}
	for ; i <= j; i++ {
		v, ok := t.Get(i).(string)
		if !ok {
			return nil, errors.New("attempt to concatenate non-string value")
		}

		b.WriteString(v)
		if i < j {
			b.WriteString(sep)
		}
	}

	return []Val{b.String()}, nil
}

func table_create(args Args) (r []Val, err error) {
	s := int(args.GetNumber())
	if s < 0 {
		return nil, errors.New("index out of range")
	}

	if len(args.List) == 1 {
		a := make([]Val, 0, s)
		return []Val{&Table{Array: a}}, nil
	}

	value := args.GetAny()
	a := make([]Val, s)
	for i := range a {
		a[i] = value
	}

	return []Val{&Table{Array: a}}, nil
}

func table_find(args Args) (r []Val, err error) {
	haystack := args.GetTable()
	needle := args.GetAny()
	init := args.GetNumber(1)
	if init < 1 {
		return nil, errors.New("index out of range")
	}

	if haystack.Array != nil {
		for i := int(init) - 1; i < len(haystack.Array); i++ {
			if needle == haystack.Array[i] {
				return []Val{float64(i + 1)}, nil
			}
		}
	}
	if haystack.Hash != nil {
		for k, v := range haystack.Hash {
			if needle == v {
				return []Val{k}, nil
			}
		}
	}

	return []Val{nil}, nil
}

func table_freeze(args Args) (r []Val, err error) {
	t := args.GetTable()

	t.readonly = true
	return []Val{t}, nil
}

func bumpelements(t *Table, start int) {
	// fmt.Println("BEFORE", start)
	// fmt.Println(t)
	// fmt.Println()

	fstart := float64(start)
	var keys float64
	for i := fstart; ; i++ {
		// fmt.Println("starting with", i)
		if t.Get(i) == nil {
			keys = i - 1
			break
		}
	}

	// fmt.Println("keys", keys)

	for k := keys; k >= fstart; k-- {
		// fmt.Println("moving key", k+1, "=", t.Get(k))
		t.ForceSet(k+1, t.Get(k))
	}

	// fmt.Println("AFTER")
	// fmt.Println(t)
	// fmt.Println()
}

func table_insert(args Args) (r []Val, err error) {
	t := args.GetTable()
	if t.readonly {
		return nil, errors.New("attempt to modify a readonly table")
	}

	n := t.Len()

	var pos int

	switch len(args.List) {
	case 2:
		pos = n + 1
	case 3:
		pos = int(args.GetNumber()) // 2nd argument is the position

		// fmt.Println("bumping elements up", pos)
		// move elements up if necessary
		if n > 0 && 1 <= pos && pos <= n {
			bumpelements(t, pos)
		}
	default:
		return nil, errors.New("wrong number of arguments to 'insert'")
	}

	v := args.GetAny()
	t.ForceSet(float64(pos), v)

	return
}

func table_isfrozen(args Args) (r []Val, err error) {
	t := args.GetTable()

	return []Val{t.readonly}, nil
}

func table_maxn(args Args) (r []Val, err error) {
	t := args.GetTable()

	var maxn float64

	// array kvs
	if t.Array != nil {
		for i, v := range t.Array {
			if v == nil {
				continue
			} else if fi := float64(i + 1); fi > maxn {
				maxn = fi
			}
		}
	}

	// hash kvs
	if t.Hash != nil {
		for k, v := range t.Hash {
			if v == nil {
				continue
			} else if fk, ok := k.(float64); ok && fk > maxn {
				maxn = fk
			}
		}
	}

	return []Val{maxn}, nil
}

func table_move(args Args) (r []Val, err error) {
	src := args.GetTable()
	a, b, t := args.GetNumber(), args.GetNumber(), args.GetNumber()
	dst := args.GetTable(src)
	if dst.readonly {
		return nil, errors.New("attempt to modify a readonly table")
	}

	for i := a; i <= b; i++ {
		dst.ForceSet(t+i-a, src.Get(i))
	}

	return []Val{dst}, nil
}

func table_pack(args Args) (r []Val, err error) {
	l := len(args.List)
	a := make([]Val, l)
	copy(a, args.List)

	return []Val{&Table{
		Hash:  map[Val]Val{"n": float64(l)},
		Array: a,
	}}, nil
}

func table_remove(args Args) (r []Val, err error) {
	t := args.GetTable()
	if t.readonly {
		return nil, errors.New("attempt to modify a readonly table")
	}

	l := float64(t.Len())
	pos := args.GetNumber(l)

	p := t.Get(pos)
	if uint(pos) == uint(l) {
		t.ForceSet(pos, nil)
	} else if 0 < pos && pos < l {
		for i := pos; i < l; i++ {
			t.ForceSet(i, t.Get(i+1))
		}
		t.ForceSet(l, nil)
	}
	return []Val{p}, nil
}

// ltablib.cpp
type comp func(a, b Val) (bool, error) // ton, compton, aint no city quite like miiine

func sort_swap(t *Table, i, j int) {
	a := t.Array
	// LUAU_ASSERT(unsigned(i) < unsigned(n) && unsigned(j) < unsigned(n)) // contract maintained in sort_less after predicate call

	// no barrier required because both elements are in the array before and after the swap
	a[i], a[j] = a[j], a[i]
}

func sort_less(t *Table, i, j int, c comp) (res bool, err error) {
	a, n := t.Array, len(t.Array)
	// LUAU_ASSERT(unsigned(i) < unsigned(n) && unsigned(j) < unsigned(n)) // contract maintained in sort_less after predicate call

	res, err = c(a[i], a[j])

	// predicate call may resize the table, which is invalid
	if len(a) != n {
		return false, errors.New("table modified during sorting")
	}
	return
}

func sort_siftheap(t *Table, l, u int, c comp, root int) (err error) {
	// LUAU_ASSERT(l <= u)
	count := u - l + 1

	// process all elements with two children
	for root*2+2 < count {
		left, right := root*2+1, root*2+2
		next := root
		if r, err := sort_less(t, l+next, l+left, c); err != nil {
			return err
		} else if r {
			next = left
		}
		if r, err := sort_less(t, l+next, l+left, c); err != nil {
			return err
		} else if r {
			next = left
		}
		if r, err := sort_less(t, l+next, l+right, c); err != nil {
			return err
		} else if r {
			next = right
		}

		if next == root {
			break
		}

		sort_swap(t, l+root, l+next)
		root = next
	}

	// process last element if it has just one child
	if lastleft := root*2 + 1; lastleft != count-1 {
		return
	} else if r, err := sort_less(t, l+root, l+lastleft, c); err != nil {
		return err
	} else if r {
		sort_swap(t, l+root, l+lastleft)
	}
	return
}

func sort_heap(t *Table, l, u int, c comp) {
	// LUAU_ASSERT(l <= u)
	count := u - l + 1

	for i := count/2 - 1; i >= 0; i-- {
		sort_siftheap(t, l, u, c, i)
	}

	for i := count - 1; i > 0; i-- {
		sort_swap(t, l, l+i)
		sort_siftheap(t, l, l+i-1, c, 0)
	}
}

func sort_rec(t *Table, l, u, limit int, c comp) (err error) {
	// sort range [l..u] (inclusive, 0-based)
	for l < u {
		// if the limit has been reached, quick sort is going over the permitted nlogn complexity, so we fall back to heap sort
		if limit == 0 {
			sort_heap(t, l, u, c)
			return
		} else if r, err := sort_less(t, u, l, c); err != nil {
			return err
		} else if r { // a[u] < a[l]?
			// sort elements a[l], a[(l+u)/2] and a[u]
			// note: this simultaneously acts as a small sort and a median selector
			sort_swap(t, u, l) // swap a[l] - a[u]
		}
		if u-l == 1 {
			break // only 2 elements
		}

		m := l + (u-l)>>1 // midpoint
		if r, err := sort_less(t, m, l, c); err != nil {
			return err
		} else if r { // a[m]<a[l]?
			sort_swap(t, m, l)
		} else if r, err := sort_less(t, u, m, c); err != nil {
			return err
		} else if r { // a[u]<a[m]?
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
			for {
				if r, err := sort_less(t, i, p, c); err != nil {
					return err
				} else if !r {
					break
				} else if i >= u {
					return errors.New("invalid order function for sorting")
				}
				i++
			}

			// repeat --j until a[j] <= P
			j--
			for {
				if r, err := sort_less(t, p, j, c); err != nil {
					return err
				} else if !r {
					break
				} else if j <= l {
					return errors.New("invalid order function for sorting")
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
		limit = limit>>1 + limit>>2

		// a[l..i-1] <= a[i] == P <= a[i+1..u]
		// sort smaller half recursively; the larger half is sorted in the next loop iteration
		if i-l < u-i {
			err := sort_rec(t, l, i-1, limit, c)
			if err != nil {
				return err
			}
			l = i + 1
		} else {
			err := sort_rec(t, i+1, u, limit, c)
			if err != nil {
				return err
			}
			u = i - 1
		}
	}

	return
}

func table_sort(args Args) (r []Val, err error) {
	t := args.GetTable()
	if t.readonly {
		return nil, errors.New("attempt to modify a readonly table")
	}

	var c comp
	if len(args.List) == 1 {
		c = jumpLt
	} else {
		f := args.GetFunction()
		c = func(a, b Val) (bool, error) {
			res, err := (*f.Run)(args.Co, a, b)
			if err != nil {
				return false, err
			}
			return res[0].(bool), nil
		}
	}

	if n := t.Len(); n > 0 {
		return nil, sort_rec(t, 0, n-1, n, c)
	}
	return
}

func table_unpack(args Args) (r []Val, err error) {
	list := args.GetTable()
	i := args.GetNumber(1)
	e := args.GetNumber(float64(list.Len()))
	if i > e {
		return // empty range
	}

	ui, uj := int(i), int(e)
	if uj <= len(list.Array) {
		return list.Array[ui-1 : uj], nil
	}

	r = make([]Val, uj-ui+1)
	for k := i; k <= e; k++ {
		r[int(k)-ui] = list.Get(k)
	}

	return
}

var libtable = NewLib([]Function{
	MakeFn("clear", table_clear),
	MakeFn("clone", table_clone),
	MakeFn("concat", table_concat),
	MakeFn("create", table_create),
	MakeFn("find", table_find),
	MakeFn("freeze", table_freeze),
	MakeFn("insert", table_insert),
	MakeFn("isfrozen", table_isfrozen),
	MakeFn("maxn", table_maxn),
	MakeFn("move", table_move),
	MakeFn("pack", table_pack),
	MakeFn("remove", table_remove),
	MakeFn("sort", table_sort),
	MakeFn("unpack", table_unpack),
})
