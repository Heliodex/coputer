package main

import "strings"

func table_clear(args Args) {
	t := args.GetTable()

	if t.readonly {
		panic("attempt to modify a readonly table")
	}

	if t.array != nil {
		for i := range *t.array {
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
	count := uint(args.GetNumber())

	array := make([]any, count)
	if len(args.args) > 1 {
		value := args.GetAny()
		for i := range array {
			array[i] = value
		}
	}

	return &Table{
		array: &array,
		hash:  &map[any]any{},
		asize: count, // not ^2?
	}
}

func table_find(args Args) Ret {
	haystack := args.GetTable()
	needle := args.GetAny()
	init := args.GetNumber(1)

	for i := init; i < haystack.Len(); i++ {
		if haystack.Get(i) == needle {
			return i
		}
	}
	return nil
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
		t.Set(l+1, value)
		return
	}

	pos, value := uint(args.GetNumber()), args.GetAny()

	if t.array == nil {
		arr := make([]any, pos)
		arr[pos-1] = value
		t.array = &arr
		t.asize = pos
		return
	}

	if pos > t.asize {
		for j := uint(l); j >= pos; j-- {
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
	} else if t.hash == nil {
		t.hash = &map[any]any{float64(pos): value}
	} else {
		(*t.hash)[float64(pos)] = value
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
			if v == nil {
				continue
			}
			nentries[float64(i+1)] = true
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
})
