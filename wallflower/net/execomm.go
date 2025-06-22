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
	"net/url"

	. "github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/wallflower/keys"
)

const (
	addr      = "http://localhost:2505"
	storeAddr = addr + "/store"
)

func StoreProgram(pk keys.PK, name string, b []byte) (hash [32]byte, err error) {
	// fmt.Println("Storing program", pk.Encode(), name)
	hash = sha3.Sum256(b)

	hashPath := "/" + hex.EncodeToString(hash[:])
	res1, err := http.Get(addr + hashPath)
	if err != nil {
		return
	}

	namePath := "/" + pk.EncodeNoPrefix() + "/" + url.PathEscape(name)
	res2, err := http.Get(addr + namePath)
	if err != nil {
		return
	}

	// Make sure that the program is accessible by both hash and pk/name
	if res1.StatusCode == http.StatusOK && res2.StatusCode == http.StatusOK {
		return
	}

	req, err := http.NewRequest(http.MethodPut, storeAddr+namePath, bytes.NewReader(b))
	if err != nil {
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	} else if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusConflict {
		// read body into byte arr
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return [32]byte{}, err
		}

		return [32]byte{}, fmt.Errorf("bad status from execution server while storing web program: %s, %s", res.Status, string(b))
	}

	return
}

func StartWebProgramHash(hash [32]byte, args WebArgs) (output WebRets, err error) {
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
		return
	} else if res.StatusCode != http.StatusOK {
		return WebRets{}, fmt.Errorf("bad status from execution server while starting web program: %s, %s", res.Status, b)
	}

	// deserialise it
	return output, json.Unmarshal(b, &output)
}

func StartWebProgramName(pk keys.PK, name string, args WebArgs) (output WebRets, err error) {
	// encode to json
	jsonargs, err := json.Marshal(args)
	if err != nil {
		return
	}

	res, err := http.Post(addr+"/web/"+pk.EncodeNoPrefix()+"/"+url.PathEscape(name), "", bytes.NewReader(jsonargs))
	if err != nil {
		return
	}

	// we need the body either way
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return
	} else if res.StatusCode != http.StatusOK {
		return WebRets{}, fmt.Errorf("bad status from execution server while starting web program: %s, %s", res.Status, b)
	}

	// deserialise it
	return output, json.Unmarshal(b, &output)
}
