package main

import "testing"

// TestTableSeparators checks how a table is laid out from the separators written in it: a table of only semicolons goes on one line, a comma keeps the table on one item per line, and a semicolon in a split table keeps the next item on its line unless a comment comes between them.
func TestTableSeparators(t *testing.T) {
	cases := []struct {
		name string
		src string
		want string
	}{
		{
			name: "semicolons on one line",
			src: "local t = {1; 2; 3}\n",
			want: "local t = {1; 2; 3}",
		},
		{
			name: "a trailing semicolon is dropped on one line",
			src: "local t = {1; 2;}\n",
			want: "local t = {1; 2}",
		},
		{
			name: "commas keep one item per line",
			src: "local t = {1, 2, 3}\n",
			want: "local t = {\n\t1,\n\t2,\n\t3,\n}",
		},
		{
			name: "a trailing comma keeps the table split, with the semicolon joining the item before it",
			src: "local t = {1; 2,}\n",
			want: "local t = {\n\t1; 2,\n}",
		},
		{
			name: "a mixed table splits, with a semicolon joining the next item",
			src: "local t = {1, 2; 3, 4,}\n",
			want: "local t = {\n\t1,\n\t2; 3,\n\t4,\n}",
		},
		{
			name: "a comment after a semicolon keeps the next item on its own line",
			src: "local t = {\n\t1; -- one\n\t2\n}\n",
			want: "local t = {\n\t1; -- one\n\t2,\n}",
		},
		{
			name: "a comment before the next item keeps it on its own line",
			src: "local t = {\n\t1;\n\t-- two\n\t2;\n\t3\n}\n",
			want: "local t = {\n\t1;\n\t-- two\n\t2; 3,\n}",
		},
	}

	for _, tc := range cases {
		got, err := formatSource(tc.src)

		if err != nil {
			t.Errorf("%s: error formatting: %v", tc.name, err)
			continue
		}

		if got != tc.want {
			t.Errorf("%s: expected\n%s\ngot\n%s", tc.name, tc.want, got)
		}
	}
}
