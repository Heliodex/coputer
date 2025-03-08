package exec

import (
	"crypto/sha3"
	"encoding/hex"
	"os"
	"path/filepath"
)

const (
	// one entrypoint to rule them all
	// one entrypoint to find them
	// one entrypoint to bring them all
	// and in the darkness require() them
	EntrypointFilename = "init.luau"
	ProgramsDir        = "./data/programs"
)

func UnbundleToDir(b []byte) (entrypath string, err error) {
	hash := sha3.Sum256(b)
	hexhash := hex.EncodeToString(hash[:])

	path := filepath.Join(ProgramsDir, hexhash)
	entrypath = filepath.Join(path, EntrypointFilename)

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

func BundleStored(hexhash string) (b bool) {
	path := filepath.Join(ProgramsDir, hexhash)
	_, err := os.Stat(path)

	return err == nil
}
