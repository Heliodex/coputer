package main

import "testing"

// TestUnionLeadingSymbol checks how a union or intersection is laid out. A leading | or & splits it, one operand to a line. Otherwise it is on one line, and -leading-symbol also splits one that is over several lines in the source.
func TestUnionLeadingSymbol(t *testing.T) {
	cases := []struct {
		name string
		src string
		flag bool
		want string
	}{
		{name: "a union on one line stays on one line", src: "type U = number | string\n", want: "type U = number | string"},
		{name: "a leading bar splits a union", src: "type U = | number | string\n", want: "type U =\n\t| number\n\t| string"},
		{name: "a leading ampersand splits an intersection", src: "type I = & A & B\n", want: "type I =\n\t& A\n\t& B"},
		{name: "an intersection on one line stays on one line", src: "type I = { a: number } & { b: string }\n", want: "type I = { a: number } & { b: string }"},
		{name: "an optional union is on one line", src: "type B = (number | string)?\n", want: "type B = number | string?"},
		{name: "a union over several lines without a leading bar is kept on one line", src: "type U = number |\n\tstring\n", want: "type U = number | string"},
		{name: "-leading-symbol splits a union over several lines", src: "type U = number |\n\tstring\n", flag: true, want: "type U =\n\t| number\n\t| string"},
		{name: "-leading-symbol leaves a union on one line alone", src: "type U = number | string\n", flag: true, want: "type U = number | string"},
	}

	for _, tc := range cases {
		splitMultilineTypes = tc.flag

		got, err := formatSource(tc.src)

		splitMultilineTypes = false

		if err != nil {
			t.Errorf("%s: error formatting: %v", tc.name, err)
			continue
		}

		if got != tc.want {
			t.Errorf("%s: expected %q, got %q", tc.name, tc.want, got)
		}
	}
}
