package net

import (
	"testing"

	"github.com/Heliodex/coputer/bundle"
	"github.com/Heliodex/coputer/wallflower/keys"
)

func TestExec(t *testing.T) {
	for _, test := range webTests {
		t.Log("-- Testing", test.Name)

		b, err := bundle.Bundle(testProgramPath + "/" + test.Name)
		if err != nil {
			t.Fatal(err)
		}

		// lel nil pk
		if _, err = StoreProgram(keys.PK{}, test.Name, b); err != nil {
			t.Fatal(err)
		}

		res, err := StartWebProgram(keys.PK{}, test.Name, test.Args)
		if err != nil {
			t.Fatal(err)
		}

		if err := test.Rets.Equal(res); err != nil {
			t.Fatal("unexpected response:", err)
		}

		t.Log(string(res.Body))
	}
}
