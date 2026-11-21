package main

import "testing"

// TestCastOperands checks that a cast is wrapped in parentheses only where it needs it: as a unary operand, a postfix base, or the left operand of <. Elsewhere it stays bare.
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
		{name: "a cast as the right operand stays bare", src: "local a = second + (tonumber(readerLine :: string) :: number)\n", want: "local a = second + tonumber(readerLine :: string) :: number"},
		{name: "a cast as the left operand of an operator other than < stays bare", src: "local a = (a :: number) + 1\n", want: "local a = a :: number + 1"},
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
