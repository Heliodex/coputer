package litecode

import (
	"crypto/sha256"
	"fmt"
	"os/exec"
)

func Compile(path string, o uint8) (bytecode []byte, err error) {
	cmd := exec.Command("luau-compile", "--binary", fmt.Sprintf("-O%d", o), path)
	return cmd.Output()
}

// simpler compilation, deserialisation, and loading API
type compiled struct {
	deserialised
	filepath string
	compiler *compiler
}

type compiler struct {
	cache map[[32]byte]deserialised
	o     uint8
}

func NewCompiler(o uint8) compiler {
	return compiler{
		cache: make(map[[32]byte]deserialised),
		o:     o,
	}
}

func (c compiler) deserialise(b []byte, filepath string) (compiled, error) {
	hash := sha256.Sum256(b)

	d, ok := c.cache[hash]
	if ok {
		return compiled{d, filepath, &c}, nil
	}

	d, err := Deserialise(b)
	if err != nil {
		return compiled{}, fmt.Errorf("error deserialising bytecode: %w", err)
	}

	c.cache[hash] = d
	return compiled{d, filepath, &c}, nil
}

func (c compiler) CompileAndDeserialise(path string) (d compiled, err error) {
	b, err := Compile(path, c.o)
	if err != nil {
		return compiled{}, fmt.Errorf("error compiling file: %w", err)
	}

	return c.deserialise(b, path)
}

func (p compiled) Load(env Env) (co Coroutine, cancel func()) {
	return LoadModule(p, env, map[string]Rets{})
}
