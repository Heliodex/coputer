package main

import "testing"

// TestStringStyles checks which strings keep their quotes when formatted. A raw string stays raw only when its value spans lines, since a quoted string can hold any value. Every other string is double quoted, unless its value has a double quote and no single quote.
func TestStringStyles(t *testing.T) {
	cases := []struct {
		name string
		src string
		want string
	}{
		{name: "a single-line raw string is quoted", src: "local a = [==[z]==]\n", want: "local a = \"z\""},
		{name: "a single-line raw string with brackets is quoted", src: "local a = [=[x]]y]=]\n", want: "local a = \"x]]y\""},
		{name: "single quotes become double", src: "local a = 'single'\n", want: "local a = \"single\""},
		{name: "a value with a double quote keeps single quotes", src: "local a = 'say \"hi\"'\n", want: "local a = 'say \"hi\"'"},
		{name: "interpolated strings become double quoted", src: "local a = `tick`\n", want: "local a = \"tick\""},
		{name: "a multi-line raw string stays raw", src: "local a = [[\nline one\nline two]]\n", want: "local a = [[\nline one\nline two]]"},
		{name: "a multi-line raw string keeps the depth its value needs", src: "local a = [=[x\n]]y]=]\n", want: "local a = [=[\nx\n]]y]=]"},
		{name: "a multi-line raw string keeps its lines unindented in a block", src: "local function f()\n\tlocal s = [[\nline one\nline two]]\nend\n", want: "local function f()\n\tlocal s = [[\nline one\nline two]]\nend"},
		{name: "a multi-line raw string that starts with a newline keeps it", src: "local a = [[\n\nabc\ndef]]\n", want: "local a = [[\n\nabc\ndef]]"},
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
