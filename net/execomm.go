package net

// communication with execution server

import (
	"bytes"
	"crypto/sha3"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Heliodex/coputer/litecode/vm"
)

const (
	addr      = "http://localhost:2505"
	storeAddr = addr + "/store"
)

func StoreProgram(data []byte) (hash [32]byte, err error) {
	hash = sha3.Sum256(data)

	res, err := http.Get(addr + "/" + hex.EncodeToString(hash[:]))
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
			return [32]byte{}, err
		}

		return [32]byte{}, fmt.Errorf("bad status from execution server while storing web program: %s, %s", res.Status, string(b))
	}

	return
}

func StartWebProgram(hash [32]byte, args vm.WebArgs) (output vm.WebRets, err error) {
	// encode to json
	jsonargs, err := json.Marshal(args)
	if err != nil {
		return
	}

	res, err := http.Post(addr+"/web/"+hex.EncodeToString(hash[:]), "", bytes.NewReader(jsonargs))
	if err != nil {
		return
	}

	// we need the body either way
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return vm.WebRets{}, err
	} else if res.StatusCode != http.StatusOK {
		return vm.WebRets{}, fmt.Errorf("bad status from execution server while starting web program: %s, %s", res.Status, b)
	}

	// deserialise it
	return output, json.Unmarshal(b, &output)
}
