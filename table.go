package main

import "strings"

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

var libtable = NewTable([][2]any{
	MakeFn("clear", table_clear),
	MakeFn("clone", table_clone),
	MakeFn("concat", table_concat),
})
