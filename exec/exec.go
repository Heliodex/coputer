package exec

import (
	"crypto/sha3"
	"encoding/hex"
	"os"
	"path/filepath"

	lc "github.com/Heliodex/litecode"
)

type Unbundler struct {
	cache    map[[32]byte]string // hash -> entrypoint
	compiler lc.Compiler
}

func NewUnbundler() Unbundler {
	return Unbundler{
		cache:    make(map[[32]byte]string),
		compiler: lc.NewCompiler(1),
	}
}

func (u *Unbundler) unbundleToDir(b []byte, d string) (entrypoint string, err error) {
	hash := sha3.Sum256(b)
	if entrypoint, ok := u.cache[hash]; ok {
		return entrypoint, nil
	}

	hexhash := hex.EncodeToString(hash[:])

	ub, err := Unbundle(b)
	if err != nil {
		return
	}

	path := filepath.Join(d, hexhash)
	entrypoint = filepath.Join(path, ub[0].path)

	if err = os.MkdirAll(path, 0o755); err != nil {
		return
	}

	for _, f := range ub {
		if err = os.WriteFile(filepath.Join(path, f.path), f.data, 0o644); err != nil {
			return
		}
	}

	return
}

func (u *Unbundler) Execute(b []byte, env lc.Env) (lc.Coroutine, error) {
	entrypoint, err := u.unbundleToDir(b, "data")
	if err != nil {
		return lc.Coroutine{}, err
	}

	p, err := u.compiler.CompileAndDeserialise(entrypoint)
	if err != nil {
		return lc.Coroutine{}, err
	}

	co, _ := p.Load(env)
	return co, nil
}
