package main

func table_clear(t *Table) {
	for i := range *t.array {
		(*t.array)[i] = nil
	}
	for k := range *t.hash {
		(*t.hash)[k] = nil
	}
}

var libtable = NewTable([][2]any{
	MakeFn("clear", table_clear),
})
