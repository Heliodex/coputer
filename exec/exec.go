package exec

import (
	"crypto/sha3"
	"encoding/hex"
	"os"
	"path/filepath"

	lc "github.com/Heliodex/coputer/litecode/vm"
)

const (
	// one entrypoint to rule them all
	// one entrypoint to find them
	// one entrypoint to bring them all
	// and in the darkness require() them
	entrypointFilename = "init.luau"
	programsDir        = "data/programs"
)

func unbundleToDir(b []byte, d string) (entrypath string, err error) {
	hash := sha3.Sum256(b)
	hexhash := hex.EncodeToString(hash[:])

	path := filepath.Join(d, hexhash)
	entrypath = filepath.Join(path, entrypointFilename)

	// if dir exists, return
	if _, err = os.Stat(entrypath); err == nil {
		return
	}

	ub, err := Unbundle(b)
	if err != nil {
		return
	}

	for _, f := range ub {
		p := filepath.Join(path, f.path)
		if err = os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return
		} else if err = os.WriteFile(p, f.data, 0o644); err != nil {
			return
		}
	}

	return
}

func Execute(c lc.Compiler, b []byte, env lc.Env) (lc.Coroutine, error) {
	entrypoint, err := unbundleToDir(b, programsDir)
	if err != nil {
		return lc.Coroutine{}, err
	}

	p, err := c.Compile(entrypoint)
	if err != nil {
		return lc.Coroutine{}, err
	}

	co, _ := p.Load(env)
	return co, nil
}
