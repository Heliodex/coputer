package exec

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var gzheader = []byte{0x1f, 0x8b, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff}

type File struct {
	path string
	data []byte
}

func Compress(n string, f []byte) (cf File, err error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)

	w.Name = n
	if _, err = w.Write(f); err != nil {
		return
	} else if err = w.Close(); err != nil {
		return
	}

	return File{n, b.Bytes()[len(gzheader):]}, nil // remove the header
}

func Decompress(c []byte) (f File, err error) {
	// add the header
	c = append(gzheader, c...)

	var b bytes.Buffer
	r, err := gzip.NewReader(bytes.NewReader(c))
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

	return Compress(np, f)
}

func Bundle(path string) (b []byte, err error) {
	var cFiles []File

	// walk through the directory
	if err = filepath.WalkDir(path, func(p string, info os.DirEntry, err error) error {
		if err != nil || info.IsDir() {
			return err
		} else if bf, err := bundleFile(p); err != nil {
			return err
		} else if bf.path == entrypointFilename {
			cFiles = append([]File{bf}, cFiles...) // entrypoint goes first
		} else {
			cFiles = append(cFiles, bf)
		}

		return nil
	}); err != nil {
		return
	}

	if cFiles[0].path != entrypointFilename {
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
	for i, lb := uint64(0), uint64(len(b)); i < lb; {
		// read file length
		l, n := binary.Uvarint(b[i:])
		i += uint64(n)

		// read file
		c := b[i:][:l]
		i += l

		f, err := Decompress(c)
		if err != nil {
			return nil, err
		}

		fs = append(fs, f)
	}

	return
}
