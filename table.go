package main

import (
	"fmt"
	"strings"
)

func table_clear(args *Args) []any {
	t := args.GetTable()

	for i := range *t.array {
		(*t.array)[i] = nil
	}
	for k := range *t.hash {
		(*t.hash)[k] = nil
	}

	return []any{}
}

func table_clone(args *Args) []any {
	t := args.GetTable()

	a2 := make([]any, len(*t.array))
	copy(a2, *t.array)

	h2 := map[any]any{}
	for k, v := range *t.hash {
		h2[k] = v
	}

	return []any{&Table{
		array: &a2,
		hash:  &h2,
		asize: t.asize,
	}}
}

func table_concat(args *Args) []any {
	t := args.GetTable()
	sep := args.GetString("")
	i := args.GetNumber(1)
	j := args.GetNumber(t.Len())

	if i > j {
		return []any{""}
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

	return []any{b.String()}
}

func table_create(args *Args) []any {
	count := uint(args.GetNumber())

	array := make([]any, count)

	t := &Table{
		array: &array,
		hash:  &map[any]any{},
		asize: count, // not ^2?
	}

	if len(args.args) > 1 {
		value := args.GetAny()
		for i := range array {
			array[i] = value
		}
	}

	return []any{t}
}

func table_find(args *Args) []any {
	haystack := args.GetTable()
	needle := args.GetAny()
	init := args.GetNumber(1)

	for i := init; i < haystack.Len(); i++ {
		if haystack.Get(i) == needle {
			return []any{i}
		}
	}
	return []any{nil}
}

func table_insert(args *Args) []any {
	t := args.GetTable()
	args.CheckNextArg()

	l := t.Len()
	if len(args.args) == 2 {
		value := args.GetAny()
		fmt.Println("2args")
		t.Set(l+1, value)
	} else {
		pos, value := int(args.GetNumber()), args.GetAny()
		fmt.Println("3args")

		arr := t.array
		// bump array indices after pos
		for i := int(l); i >= pos; i-- {
			(*arr)[i] = (*arr)[i-1]
		}
		(*arr)[pos-1] = value
	}

	return []any{}
}

var libtable = NewTable([][2]any{
	MakeFn("clear", table_clear),
	MakeFn("clone", table_clone),
	MakeFn("concat", table_concat),
	MakeFn("create", table_create),
	MakeFn("find", table_find),
	MakeFn("insert", table_insert),
})
