package vm

import (
	"crypto/sha3"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

const Ext = ".luau"

func luauCompile(path string, o uint8) (bytecode []byte, err error) {
	cmd := exec.Command("luau-compile", "--binary", fmt.Sprintf("-O%d", o), path)
	return cmd.Output()
}

// simpler compilation, deserialisation, and loading API
type compiled struct {
	deserialised
	filepath       string
	dbgpath        string
	compiler       *Compiler
	requireHistory []string
}

type deserpath struct {
	deserialised
	dbgpath string
}

// Compiler allows programs to be compiled and deserialised with a cache and given optimisation level.
type Compiler struct {
	cache map[[32]byte]deserpath
	o     uint8
}

// NewCompiler creates a new compiler with the given optimisation level.
func NewCompiler(o uint8) Compiler {
	return Compiler{
		cache: make(map[[32]byte]deserpath),
		o:     o,
	}
}

func (c Compiler) deserialise(b []byte, path, dbgpath string) (compiled, error) {
	d, err := deserialise(b)
	if err != nil {
		return compiled{}, fmt.Errorf("error deserialising bytecode: %w", err)
	}

	return compiled{
		deserialised: d,
		filepath:     path,
		dbgpath:      dbgpath,
		compiler:     &c,
	}, nil
}

// Compile compiles a program at a specific path to bytecode and returns its deserialised form.
func (c Compiler) Compile(path string) (p compiled, err error) {
	// hash path instead of bytecode
	hash := sha3.Sum256([]byte(path))
	if d, ok := c.cache[hash]; ok {
		return compiled{
			deserialised: d.deserialised,
			filepath:     path,
			dbgpath:      d.dbgpath,
			compiler:     &c,
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

	b, err := luauCompile(pathext, c.o)
	if err != nil {
		return compiled{}, fmt.Errorf("error compiling file: %w", err)
	}

	// dbgpath has the extension and all
	p, err = c.deserialise(b, path, pathext)
	if err != nil {
		return
	}

	c.cache[hash] = deserpath{
		deserialised: p.deserialised,
		dbgpath:      pathext,
	}
	return
}

func (p compiled) Load(env Env, args ProgramArgs) (co Coroutine, cancel func()) {
	return loadmodule(p, env, map[string]Vals{}, nil, args)
}
