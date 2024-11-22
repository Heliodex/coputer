package main

import "strings"

func table_clear(t *Table) {
	for i := range *t.array {
		(*t.array)[i] = nil
	}
	for k := range *t.hash {
		(*t.hash)[k] = nil
	}
}

func table_clone(t *Table) *Table {
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

func table_concat(t *Table, args ...any) string {
	sep, i, j := "", 1., t.Len()

	switch len(args) {
	case 3:
		j = args[2].(float64)
		fallthrough
	case 2:
		i = args[1].(float64)
		fallthrough
	case 1:
		sep = args[0].(string)
	}

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

var libtable = NewTable([][2]any{
	MakeFn("clear", table_clear),
	MakeFn("clone", table_clone),
	MakeFn("concat", table_concat),
})
