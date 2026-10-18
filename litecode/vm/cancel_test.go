package vm

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/litecode/vm/compile"
)

// A cancelled program must stop even if it never calls anything. The cancellation check lives at loop back-edges, so this covers each loop form the compiler emits: JUMPBACK (while/repeat), FORNLOOP (numeric for) and FORGLOOP (generic for).
func TestCancel(t *testing.T) {
	scripts := map[string]string{
		"while":  "local n = 0\nwhile true do n += 1 end\n",
		"repeat": "local n = 0\nrepeat n += 1 until false\n",
		"fornum": "local n = 0\nfor i = 1, math.huge do n += 1 end\n",
		"forgen": "local function it() return 1 end\nlocal n = 0\nfor v in it do n += v end\n",
	}

	dir := t.TempDir()
	for name, src := range scripts {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(dir, name)
			if err := os.WriteFile(path+compile.Ext, []byte(src), 0o644); err != nil {
				t.Fatal(err)
			}

			p, err := compile.Compile(compile.MakeCompiler(1), path)
			if err != nil {
				t.Fatal(err)
			}

			co, cancel := Load(p, nil, TestArgs{})
			done := make(chan struct{})
			go func() {
				co.Resume()
				close(done)
			}()

			time.Sleep(20 * time.Millisecond)
			cancel()

			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("program did not stop after cancel")
			}
		})
	}
}
