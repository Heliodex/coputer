package main

import "testing"

// TestCastOperands checks that a cast used as an operand is wrapped in parentheses, so that its type does not run into the operator after it, and that a cast written as a whole expression stays bare.
func TestCastOperands(t *testing.T) {
	cases := []struct {
		name string
		src string
		want string
	}{
		{name: "a comparison", src: "local a = (index :: number) < 0\n", want: "local a = (index :: number) < 0"},
		{name: "a unary operand", src: "local a = -(index :: number)\n", want: "local a = -(index :: number)"},
		{name: "a length operand", src: "local a = #(x :: {number})\n", want: "local a = #(x :: { number })"},
		{name: "a whole value stays bare", src: "local a = index :: number\n", want: "local a = index :: number"},
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
