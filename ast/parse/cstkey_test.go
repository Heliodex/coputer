package main

import "testing"

// TestCstTableSeparators checks that the CST of a table records each item's separator, a comma or a semicolon, and that a table parses without a CST panicking when no CST data is asked for.
func TestCstTableSeparators(t *testing.T) {
	ok, res := Parse("local t = {1, 2; 3,}\n", Options{StoreCstData: true})

	if !ok {
		t.Fatal("expected the table to parse")
	}

	var tables []CstExprTable

	for _, node := range res.CstNodeMap {
		if table, isTable := node.(CstExprTable); isTable {
			tables = append(tables, table)
		}
	}

	if len(tables) != 1 {
		t.Fatalf("expected 1 table in the CST, got %d", len(tables))
	}

	items := tables[0].Items
	want := []*int{&pzero, &pone, &pzero}

	if len(items) != len(want) {
		t.Fatalf("expected %d items, got %d", len(want), len(items))
	}

	for i, item := range items {
		if item.Separator != want[i] {
			t.Errorf("item %d: expected separator %v, got %v", i, *want[i], item.Separator)
		}
	}

	if !items[0].SeparatorIsComma() || items[0].SeparatorIsSemicolon() {
		t.Error("expected the first item to be followed by a comma")
	}

	if !items[1].SeparatorIsSemicolon() || items[1].SeparatorIsComma() {
		t.Error("expected the second item to be followed by a semicolon")
	}

	if ok, _ := Parse("local t = {1, 2}\n", Options{}); !ok {
		t.Error("expected a table to parse without CST data")
	}
}
