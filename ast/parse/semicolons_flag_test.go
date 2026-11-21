package main

import "testing"

// TestCommasWithoutFlag checks that without -semicolons a table on one line keeps its commas, and stays on one line.
func TestCommasWithoutFlag(t *testing.T) {
	const want string = "local t = { 1, 2, 3 }"

	got, err := formatSource("local t = {1, 2, 3}\n")
	if err != nil {
		t.Fatal("error formatting:", err)
	}

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

// TestSemicolonsOnOneLine checks the -semicolons option: a table on one line has its commas written as semicolons, so it is still on one line, and a table over several lines keeps its commas.
func TestSemicolonsOnOneLine(t *testing.T) {
	semicolonsOnOneLine = true

	defer func() {
		semicolonsOnOneLine = false
	}()

	cases := []struct {
		name, src, want string
	}{
		{
			name: "a value table on one line",
			src: "local t = {1, 2, 3}\n",
			want: "local t = { 1; 2; 3 }",
		},
		{
			name: "a trailing comma is dropped",
			src: "local t = {1, 2,}\n",
			want: "local t = { 1; 2 }",
		},
		{
			name: "a nested table on one line",
			src: "local t = {1, {2, 3}}\n",
			want: "local t = { 1; { 2; 3 } }",
		},
		{
			name: "a type table on one line",
			src: "type A = {a: number, b: string}\n",
			want: "type A = { a: number; b: string }",
		},
		{
			name: "a table over several lines keeps its commas",
			src: "local t = {\n\t1,\n\t2,\n}\n",
			want: "local t = {\n\t1,\n\t2,\n}",
		},
	}

	for _, tc := range cases {
		got, err := formatSource(tc.src)
		if err != nil {
			t.Errorf("%s: error formatting: %v", tc.name, err)
			continue
		}

		if got != tc.want {
			t.Errorf("%s: expected %q, got %q", tc.name, tc.want, got)
		}
	}
}
