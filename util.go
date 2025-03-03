package litecode

import (
	"crypto/sha3"
	"fmt"
	"os/exec"
)

func compile(path string, o uint8) (bytecode []byte, err error) {
	cmd := exec.Command("luau-compile", "--binary", fmt.Sprintf("-O%d", o), path)
	return cmd.Output()
}

// simpler compilation, deserialisation, and loading API
type compiled struct {
	deserialised
	filepath string
	compiler *Compiler
}

// Compiler allows programs to be compiled and deserialised with a cache and given optimisation level.
type Compiler struct {
	cache map[[32]byte]deserialised
	o     uint8
}

// NewCompiler creates a new compiler with the given optimisation level.
func NewCompiler(o uint8) Compiler {
	return Compiler{
		cache: make(map[[32]byte]deserialised),
		o:     o,
	}
}

func (c Compiler) deserialise(b []byte, filepath string) (compiled, error) {
	hash := sha3.Sum256(b)

	d, ok := c.cache[hash]
	if ok {
		return compiled{d, filepath, &c}, nil
	}

	d, err := deserialise(b)
	if err != nil {
		return compiled{}, fmt.Errorf("error deserialising bytecode: %w", err)
	}

	c.cache[hash] = d
	return compiled{d, filepath, &c}, nil
}

func (c Compiler) CompileAndDeserialise(path string) (p compiled, err error) {
	b, err := compile(path, c.o)
	if err != nil {
		return compiled{}, fmt.Errorf("error compiling file: %w", err)
	}

	return c.deserialise(b, path)
}

func (p compiled) Load(env Env) (co Coroutine, cancel func()) {
	return loadmodule(p, env, map[string]Rets{})
}
