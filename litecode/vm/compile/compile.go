package compile

import (
	"crypto/sha3"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/Heliodex/coputer/litecode/internal"
	. "github.com/Heliodex/coputer/litecode/types"
)

const Ext = ".luau"

func luauCompile(path string, o uint8) (bytecode []byte, err error) {
	cmd := exec.Command("luau-compile", "--binary", fmt.Sprintf("-O%d", o), path)
	return cmd.Output()
}

type Program struct {
	internal.Deserpath
	Filepath       string
	Compiler       Compiler
	RequireHistory []string
}

// NewCompiler creates a new compiler with the given optimisation level.
func NewCompiler(o uint8) Compiler {
	return Compiler{
		Cache: make(map[[32]byte]internal.Deserpath),
		O:     o,
	}
}

// Compile compiles a program at a specific path to bytecode and returns its deserialised form.
func Compile(c Compiler, path string) (p Program, err error) {
	// hash path instead of bytecode
	hash := sha3.Sum256([]byte(path))
	if dp, ok := c.Cache[hash]; ok {
		return Program{
			Deserpath: dp,
			Filepath:  path,
			Compiler:  c,
		}, nil
	}

	pathext := path + Ext
	// find if file at path exists
	if _, err := os.Stat(pathext); err != nil {
		if _, err := os.Stat(path); err != nil {
			return Program{}, errors.New("error finding file")
		}
		// main.luau directory
		pathext = path + "/main" + Ext
	}

	b, err := luauCompile(pathext, c.O)
	if err != nil {
		return Program{}, fmt.Errorf("error compiling file: %w", err)
	}

	// dbgpath has the extension and all
	d, err := Deserialise(b)
	if err != nil {
		return Program{}, fmt.Errorf("error deserialising bytecode: %w", err)
	}

	dp := internal.Deserpath{
		Deserialised: d,
		Dbgpath:      pathext,
	}
	c.Cache[hash] = dp

	return Program{
		Deserpath: dp,
		Filepath:  path,
		Compiler:  c,
	}, nil
}
