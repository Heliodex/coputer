package vm

import (
	"crypto/sha3"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/Heliodex/coputer/litecode/types"
)

const Ext = ".luau"

func luauCompile(path string, o uint8) (bytecode []byte, err error) {
	cmd := exec.Command("luau-compile", "--binary", fmt.Sprintf("-O%d", o), path)
	return cmd.Output()
}

type compiled struct {
	types.Deserpath
	Filepath       string
	Compiler       types.Compiler
	RequireHistory []string
}

// NewCompiler creates a new compiler with the given optimisation level.
func NewCompiler(o uint8) types.Compiler {
	return types.Compiler{
		Cache: make(map[[32]byte]types.Deserpath),
		O:     o,
	}
}

// Compile compiles a program at a specific path to bytecode and returns its deserialised form.
func Compile(c types.Compiler, path string) (p compiled, err error) {
	// hash path instead of bytecode
	hash := sha3.Sum256([]byte(path))
	if dp, ok := c.Cache[hash]; ok {
		return compiled{
			Deserpath: dp,
			Filepath:        path,
			Compiler:        c,
		}, nil
	}

	pathext := path + Ext
	// find if file at path exists
	if _, err := os.Stat(pathext); err != nil {
		if _, err := os.Stat(path); err != nil {
			return compiled{}, errors.New("error finding file")
		}
		// init.luau directory
		pathext = path + "/init" + Ext
	}

	b, err := luauCompile(pathext, c.O)
	if err != nil {
		return compiled{}, fmt.Errorf("error compiling file: %w", err)
	}

	// dbgpath has the extension and all
	d, err := deserialise(b)
	if err != nil {
		return compiled{}, fmt.Errorf("error deserialising bytecode: %w", err)
	}

	c.Cache[hash] = p.Deserpath
	return compiled{
		Deserpath: types.Deserpath{d, pathext},
		Filepath:        path,
		Compiler:        c,
	}, nil
}

func (p compiled) Load(env types.Env, args types.ProgramArgs) (co types.Coroutine, cancel func()) {
	return loadmodule(p, env, map[string]types.Val{}, args)
}
