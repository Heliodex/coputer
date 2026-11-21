package main

import "testing"

// TestComparisonOperands checks that a comparison under another comparison is wrapped in parentheses, since Luau's linter flags two comparisons in a row without them. Other operators are not affected.
func TestComparisonOperands(t *testing.T) {
	cases := []struct {
		name string
		src string
		want string
	}{
		{name: "a comparison on the left of a comparison", src: "local x = r < 0 ~= (b < 0)\n", want: "local x = (r < 0) ~= (b < 0)"},
		{name: "comparisons chain to the left, each wrapped", src: "local x = a < b == c < d\n", want: "local x = ((a < b) == c) < d"},
		{name: "a single comparison is not wrapped", src: "local x = a == b\n", want: "local x = a == b"},
		{name: "arithmetic under a comparison is not wrapped", src: "local x = a + 1 < b\n", want: "local x = a + 1 < b"},
		{name: "a comparison under and is not wrapped", src: "local x = a < b and c == d\n", want: "local x = a < b and c == d"},
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
