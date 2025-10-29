package bundle

import (
	"bytes"
	"compress/gzip"
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
	// Pre-allocate buffer with estimated size (typically compressed data is ~30-50% of original)
	var b bytes.Buffer
	b.Grow(len(f) / 2)
	
	w := gzip.NewWriter(&b)
	w.Name = n
	
	if _, err = w.Write(f); err != nil {
		return
	}

	if err = w.Close(); err != nil {
		return
	}

	return File{n, b.Bytes()[gzheaderLen:]}, nil // remove the header
}

func decompress(c []byte) (f File, err error) {
	// Pre-allocate buffer with estimated decompressed size (typically 2-3x compressed size)
	var b bytes.Buffer
	b.Grow(len(c) * 2)
	
	r, err := gzip.NewReader(bytes.NewReader(append(gzheader[:], c...))) // add the header
	if err != nil {
		return
	}

	if _, err = b.ReadFrom(r); err != nil {
		return
	}

	if err = r.Close(); err != nil {
		return
	}

	return File{r.Name, b.Bytes()}, nil
}

func bundleFile(p string, depth int) (bf File, err error) {
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
	np := strings.Join(s[depth:], "/")

	return compress(np, f)
}
