package net

import (
	"testing"

	"github.com/Heliodex/coputer/bundle"
)

func TestExec(t *testing.T) {
	for _, test := range webTests {
		t.Log("-- Testing", test.Name)

		b, err := bundle.Bundle(testProgramPath + "/" + test.Name)
		if err != nil {
			t.Fatal(err)
		}

		hash, err := StoreProgram(b)
		if err != nil {
			t.Fatal(err)
		}

		res, err := StartWebProgram(hash, test.Args)
		if err != nil {
			t.Fatal(err)
		} else if err := test.Rets.Equal(res); err != nil {
			t.Fatal("unexpected response:", err)
		}

		t.Log(string(res.Body))
	}
}
