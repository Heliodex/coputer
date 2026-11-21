package main

import "testing"

// TestStringStyles checks which strings keep their quotes when formatted. A raw string stays raw, at the shortest bracket depth its value allows. Other strings are double quoted, unless their value has a double quote and no single quote.
func TestStringStyles(t *testing.T) {
	cases := []struct {
		name string
		src string
		want string
	}{
		{name: "raw string at the shortest depth", src: "local a = [==[z]==]\n", want: "local a = [[z]]"},
		{name: "raw string keeps the depth its value needs", src: "local a = [=[x]]y]=]\n", want: "local a = [=[x]]y]=]"},
		{name: "single quotes become double", src: "local a = 'single'\n", want: "local a = \"single\""},
		{name: "a value with a double quote keeps single quotes", src: "local a = 'say \"hi\"'\n", want: "local a = 'say \"hi\"'"},
		{name: "interpolated strings become double quoted", src: "local a = `tick`\n", want: "local a = \"tick\""},
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
