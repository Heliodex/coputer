package main

import "testing"

// TestLogicalChains checks that an `and` chain starting with true, or an `or` chain starting with false, is split with one operand on each line and the operator leading. Any other chain stays on one line, and so does a chain nested inside another operator.
func TestLogicalChains(t *testing.T) {
	cases := []struct {
		name string
		src string
		want string
	}{
		{name: "an and chain that starts with true is split", src: "local a = true and x and y\n", want: "local a = true\n\tand x\n\tand y"},
		{name: "an or chain that starts with false is split", src: "local b = false or x or y\n", want: "local b = false\n\tor x\n\tor y"},
		{name: "an and chain that does not start with true stays on one line", src: "local c = x and y\n", want: "local c = x and y"},
		{name: "an or chain that does not start with false stays on one line", src: "local d = x or false or y\n", want: "local d = x or false or y"},
		{name: "a chain inside another operator stays on one line", src: "local e = (true and x) or y\n", want: "local e = true and x or y"},
		{name: "a chain in a condition is split", src: "if true and ready and not busy then\n\tprint(1)\nend\n", want: "if true\n\tand ready\n\tand not busy then\n\tprint(1)\nend"},
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
