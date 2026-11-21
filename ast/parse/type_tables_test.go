package main

import "testing"

// TestTypeTableSeparators checks how a type table is laid out from the separators written in it, by the same rules as a table value: one line when every separator is a semicolon, and otherwise one entry to a line with a semicolon keeping the next entry on its line.
func TestTypeTableSeparators(t *testing.T) {
	cases := []struct {
		name string
		src string
		want string
	}{
		{name: "semicolons on one line", src: "type A = {a: number; b: string}\n", want: "type A = { a: number; b: string }"},
		{name: "commas keep one entry per line", src: "type B = {a: number, b: string}\n", want: "type B = {\n\ta: number,\n\tb: string,\n}"},
		{name: "a mixed type table splits, with a semicolon joining the next entry", src: "type C = {a: number; b: string, c: boolean}\n", want: "type C = {\n\ta: number; b: string,\n\tc: boolean,\n}"},
		{name: "an indexer is on the line like any entry", src: "type D = {[string]: number; x: boolean}\n", want: "type D = { [string]: number; x: boolean }"},
		{name: "a comment keeps the next entry on its own line", src: "type F = {\n\ta: number; -- one\n\tb: string\n}\n", want: "type F = {\n\ta: number; -- one\n\tb: string,\n}"},
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
