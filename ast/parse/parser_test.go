package main

import (
	"sync"
	"testing"
)

// TestConcurrentParse verifies that Parse is reentrant: all mutable parser
// state lives on *Parser, so independent parses can run concurrently.
func TestConcurrentParse(t *testing.T) {
	sources := []string{
		`local x = 1 + 2`,
		`if local y = 3 then print(y) end`,
		`export const z = 4`,
		`local n = 0xFFi`,
		`local s = ` + "`hello {1 + 1}`",
	}

	var wg sync.WaitGroup
	for i := range 32 {
		src := sources[i%len(sources)]
		wg.Go(func() {
			ok, res := Parse(src, Options{})
			if !ok {
				t.Errorf("parse failed for %q: %v", src, res.Errors)
			}
		})
	}
	wg.Wait()
}
