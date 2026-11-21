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

// TestLogicalChainsMultiline checks -split-logical: a chain of two or more operators that is already on several lines is split, and one on a single line is not, even when it starts with true.
func TestLogicalChainsMultiline(t *testing.T) {
	cases := []struct {
		name string
		src string
		flag bool
		want string
	}{
		{name: "a multi-line chain is split with the flag", src: "local a = x and\n\ty and\n\tz\n", flag: true, want: "local a = x\n\tand y\n\tand z"},
		{name: "a multi-line chain is kept without the flag", src: "local a = x and\n\ty and\n\tz\n", flag: false, want: "local a = x and y and z"},
		{name: "a chain of two operands is not split", src: "local a = x and\n\ty\n", flag: true, want: "local a = x and y"},
		{name: "a single-line chain is not split", src: "local a = x or y or z\n", flag: true, want: "local a = x or y or z"},
	}

	for _, tc := range cases {
		splitMultilineLogical = tc.flag

		got, err := formatSource(tc.src)

		splitMultilineLogical = false

		if err != nil {
			t.Errorf("%s: error formatting: %v", tc.name, err)
			continue
		}

		if got != tc.want {
			t.Errorf("%s: expected %q, got %q", tc.name, tc.want, got)
		}
	}
}
