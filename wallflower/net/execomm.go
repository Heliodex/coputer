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

	"github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/wallflower/keys"
)

const (
	addr      = "http://localhost:2505"
	storeAddr = addr + "/store"
)

func StoreProgram(pk keys.PK, name string, b []byte) (hash [32]byte, err error) {
	hash = sha3.Sum256(b)

	res, err := http.Get(addr + "/" + hex.EncodeToString(hash[:]))
	if err != nil || res.StatusCode == http.StatusOK {
		return
	}

	p := storeAddr + "/" + pk.EncodeNoPrefix() + "/" + url.PathEscape(name)
	req, err := http.NewRequest(http.MethodPut, p, bytes.NewReader(b))
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

func StartWebProgramHash(hash [32]byte, args types.WebArgs) (output types.WebRets, err error) {
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
		return types.WebRets{}, fmt.Errorf("bad status from execution server while starting web program: %s, %s", res.Status, b)
	}

	// deserialise it
	return output, json.Unmarshal(b, &output)
}

func StartWebProgramName(pk keys.PK, name string, args types.WebArgs) (output types.WebRets, err error) {
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
		return types.WebRets{}, fmt.Errorf("bad status from execution server while starting web program: %s, %s", res.Status, b)
	}

	// deserialise it
	return output, json.Unmarshal(b, &output)
}
