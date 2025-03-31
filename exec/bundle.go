package exec

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const gzheaderLen = 10

var gzheader = [gzheaderLen]byte{0x1f, 0x8b, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff}

type File struct {
	path string
	data []byte
}

func compress(n string, f []byte) (cf File, err error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)

	w.Name = n
	if _, err = w.Write(f); err != nil {
		return
	} else if err = w.Close(); err != nil {
		return
	}

	return File{n, b.Bytes()[gzheaderLen:]}, nil // remove the header
}

func decompress(c []byte) (f File, err error) {
	var b bytes.Buffer
	r, err := gzip.NewReader(bytes.NewReader(append(gzheader[:], c...))) // add the header
	if err != nil {
		return
	} else if _, err = b.ReadFrom(r); err != nil {
		return
	} else if err = r.Close(); err != nil {
		return
	}

	return File{r.Name, b.Bytes()}, nil
}

func bundleFile(p string) (bf File, err error) {
	// read
	f, err := os.ReadFile(p)
	if err != nil {
		return
	}

	// trim ../ or ..\ from p
	fps := string(filepath.Separator)
	p = strings.TrimPrefix(p, ".."+fps)
	// remove walked directory from p (everything before first /)
	s := strings.Split(p, fps)
	np := strings.Join(s[1:], "/")

	return compress(np, f)
}

func Bundle(path string) (b []byte, err error) {
	var cFiles []File

	// walk through the directory
	if err = filepath.WalkDir(path, func(p string, info os.DirEntry, err error) error {
		if err != nil || info.IsDir() {
			return err
		} else if bf, err := bundleFile(p); err != nil {
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
