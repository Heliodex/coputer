package bundle

import "testing"

const testpath = "../test/web1"

func TestBundle(t *testing.T) {
	b, err := Bundle(testpath)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Bundle:", len(b))
	ub, err := Unbundle(b)
	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}

	// t.Log("Unbundle:")
	for _, f := range ub {
		t.Log("  ", f.path, len(f.data))
	}

	// rebundle
	b2, err := Bundle(testpath)
	if err != nil {
		t.Fatal(err)
	}

	if len(b) != len(b2) {
		t.Fatal("rebundled bundle is different")
	}
}
