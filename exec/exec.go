package exec

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
)

// var header = []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00}

func Compress(f []byte) (c []byte, err error) {
	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return
	}

	w.Write(f)

	w.Close()
	return b.Bytes(), nil // remove the 3 byte header
}

func Decompress(f []byte) (c []byte, err error) {
	// add the 3 byte header
	// f = append(header, f...)

	var b bytes.Buffer
	r, err := gzip.NewReader(bytes.NewReader(f))
	if err != nil {
		return
	}

	b.ReadFrom(r)

	r.Close()
	return b.Bytes(), nil
}

type File struct {
	path string
	data []byte
}

func Bundle(path, entrypoint string) (b []byte, err error) {
	var compressedFiles []File

	// walk through the directory
	if err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if info.IsDir() {
			return nil
		}

		// read
		f, err := os.ReadFile(p)
		if err != nil {
			return err
		}

		// remove walked directory from p (everything before first /)
		// warning: can't use path as it is longer than p
		s := strings.Split(p, string(filepath.Separator))
		p = strings.Join(s[1:], "/")

		c, err := Compress(f)
		if err != nil {
			return err
		}

		// enrypoint goes first
		if p == entrypoint {
			compressedFiles = append([]File{{p, c}}, compressedFiles...)
		} else {
			compressedFiles = append(compressedFiles, File{p, c})
		}

		return nil
	}); err != nil {
		return
	}

	// write compressed files
	for _, c := range compressedFiles {
		// add file path length as uint8
		b = append(b, byte(len(c.path)))
		b = append(b, c.path...)

		// add file length as uvarint
		b = binary.AppendUvarint(b, uint64(len(c.data)))
		b = append(b, c.data...)
	}

	return
}

func Unbundle(b []byte) (fs []File, err error) {
	for i, lb := uint64(0), uint64(len(b)); i < lb; {
		// read file path length
		l := uint64(b[i])
		i++

		// read file path
		p := b[i:][:l]
		i += l

		// read file length
		l, n := binary.Uvarint(b[i:])
		i += uint64(n)

		// read file
		f := b[i:][:l]
		i += l

		dec, err := Decompress(f)
		if err != nil {
			return nil, err
		}

		fs = append(fs, File{string(p), dec})
	}

	return
}
