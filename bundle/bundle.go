package bundle

import (
	"crypto/sha3"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// one entrypoint to rule them all
	// one entrypoint to find them
	// one entrypoint to bring them all
	// and in the darkness require() them
	Entrypoint         = "init"
	EntrypointFilename = Entrypoint + ".luau"
	DataDir            = "./data"
	ProgramsDir        = DataDir + "/programs"
)

func Bundle(path string) (b []byte, err error) {
	var cFiles []File

	// count the number of file separators in path
	fps := string(filepath.Separator)
	depth := strings.Count(path, fps)

	// walk through the directory
	if err = filepath.WalkDir(path, func(p string, info os.DirEntry, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		if bf, err := bundleFile(p, depth); err != nil {
			return err
		} else if bf.path == EntrypointFilename {
			cFiles = append([]File{bf}, cFiles...) // entrypoint goes first
		} else {
			cFiles = append(cFiles, bf)
		}

		return nil
	}); err != nil {
		return
	}

	if cFiles[0].path != EntrypointFilename {
		return nil, errors.New("entrypoint (init.luau) not found")
	}

	// write compressed files
	for _, c := range cFiles {
		// add file length as uvarint
		b = binary.AppendUvarint(b, uint64(len(c.data)))
		b = append(b, c.data...)
	}

	return
}

func Unbundle(b []byte) (fs []File, err error) {
	if len(b) == 0 {
		return nil, errors.New("bad bundle: 0 length")
	}

	for i, lb := uint64(0), uint64(len(b)); i < lb; {
		// read file length
		l, n := binary.Uvarint(b[i:])
		i += uint64(n)

		if i+l > lb {
			return nil, errors.New("bad bundle: file length exceeds bundle length")
		}

		// read file
		c := b[i:][:l]
		i += l

		f, err := decompress(c)
		if err != nil {
			return nil, fmt.Errorf("bad bundle: %w", err)
		}

		fs = append(fs, f)
	}

	return
}

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
