package bundle

import (
	"os"
	"testing"
)

const testpath = "../test/programs"

func TestBundle(t *testing.T) {
	files, err := os.ReadDir(testpath)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}

		path := testpath + "/" + f.Name()

		b, err := Bundle(path)
		if err != nil {
			t.Fatal(err)
		}
	
		t.Log("Bundle:", len(b))
		ub, err := Unbundle(b)
		if err != nil {
			t.Fatal(err)
			t.FailNow()
		}
	
		t.Log("Unbundle:")
		for _, f := range ub {
			t.Log("  ", f.path, len(f.data))
		}
	
		// rebundle
		b2, err := Bundle(path)
		if err != nil {
			t.Fatal(err)
		}
	
		if len(b) != len(b2) {
			t.Fatal("rebundled bundle is different")
		}
	}
}
