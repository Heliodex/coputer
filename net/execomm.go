package net

// communication with exec

import (
	"bytes"
	"crypto/sha3"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
)

const (
	addr      = "http://localhost:2505"
	storeAddr = addr + "/store"
)

func StoreProgram(data []byte) (err error) {
	hash := sha3.Sum256(data)
	hexhash := hex.EncodeToString(hash[:])

	res, err := http.Get(addr + "/" + hexhash)
	if err != nil || res.StatusCode == http.StatusOK {
		return
	}

	req, err := http.NewRequest(http.MethodPut, storeAddr, bytes.NewReader(data))
	if err != nil {
		return
	}

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	} else if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		// read body into byte arr
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("bad status: %s, %s", res.Status, string(b))
	}

	fmt.Println("Stored program with hash:", hexhash)
	return
}
