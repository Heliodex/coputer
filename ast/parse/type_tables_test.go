package main

import "testing"

// TestTypeTableSeparators checks how a type table is laid out. A type table stays on one line unless it has a comment, an entry over several lines, or a trailing comma, which splits it: then one entry to a line, with a semicolon keeping the next entry on its line.
func TestTypeTableSeparators(t *testing.T) {
	cases := []struct {
		name, src, want string
	}{
		{
			name: "semicolons on one line",
			src: "type A = {a: number; b: string}\n",
			want: "type A = { a: number; b: string }",
		},
		{
			name: "commas keep one entry per line",
			src: "type B = {a: number, b: string}\n",
			want: "type B = { a: number, b: string }",
		},
		{
			name: "a trailing comma splits a mixed type table, with a semicolon joining the next entry",
			src: "type C = {a: number; b: string, c: boolean,}\n",
			want: "type C = {\n\ta: number; b: string,\n\tc: boolean,\n}",
		},
		{
			name: "an indexer is on the line like any entry",
			src: "type D = {[string]: number; x: boolean}\n",
			want: "type D = { [string]: number; x: boolean }",
		},
		{
			name: "a single entry stays on one line",
			src: "type H = {[string]: boolean}\n",
			want: "type H = { [string]: boolean }",
		},
		{
			name: "a trailing semicolon is dropped on one line",
			src: "type G = {a: number;}\n",
			want: "type G = { a: number }",
		},
		{
			name: "a trailing comma keeps the table split",
			src: "type J = {a: number,}\n",
			want: "type J = {\n\ta: number,\n}",
		},
		{
			name: "a comment keeps the next entry on its own line",
			src: "type F = {\n\ta: number; -- one\n\tb: string\n}\n",
			want: "type F = {\n\ta: number; -- one\n\tb: string,\n}",
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
