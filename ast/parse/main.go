package main

import (
	"fmt"
	"io"
	"os"
)

// main formats the Luau source read from stdin and writes it to stdout.
func main() {
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading stdin:", err)
		os.Exit(1)
	}

	ok, res := Parse(string(content), Options{})
	if !ok {
		fmt.Fprintln(os.Stderr, "parse failed with errors:")
		for _, err := range res.Errors {
			fmt.Fprintf(os.Stderr, "- %s at %s\n", err.Message, err.Location)
		}
		os.Exit(1)
	}

	fmt.Println(res.Root.Source())
}
