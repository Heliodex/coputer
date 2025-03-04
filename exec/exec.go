package exec

import (
	"crypto/sha3"
	"encoding/hex"
	"os"
	"path/filepath"

	lc "github.com/Heliodex/litecode"
)

const (
	entrypointFilename = "entrypoint.txt"
	programsDir        = "data/programs"
)

func unbundleToDir(b []byte, d string) (entrypath string, err error) {
	hash := sha3.Sum256(b)
	hexhash := hex.EncodeToString(hash[:])

	path := filepath.Join(d, hexhash)

	// if dir exists, return
	if e, err := os.ReadFile(filepath.Join(d, hexhash, entrypointFilename)); err == nil {
		return filepath.Join(path, string(e)), nil
	}

	ub, err := Unbundle(b)
	if err != nil {
		return
	}

	entrypoint := ub[0].path
	entrypath = filepath.Join(path, entrypoint)

	if err = os.MkdirAll(path, 0o755); err != nil {
		return
	}

	for _, f := range ub {
		if err = os.WriteFile(filepath.Join(path, f.path), f.data, 0o644); err != nil {
			return
		}
	}

	// write entrypoint to file
	// (maybe just have 1 entrypoint for all programs? likely main.luau)
	return entrypath, os.WriteFile(filepath.Join(path, entrypointFilename), []byte(entrypoint), 0o644)
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
