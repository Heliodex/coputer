package net

// communication with execution server

import (
	"bytes"
	"crypto/sha3"
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
	fmt.Println("Storing program", pk.Encode(), name)
	hash = sha3.Sum256(b)

	namePath := "/" + pk.EncodeNoPrefix() + "/" + url.PathEscape(name)
	// doesn't update an existing program
	// if resn, err := http.Get(addr + namePath); err != nil || resn.StatusCode == http.StatusOK {
	// 	return hash, err
	// }

	req, err := http.NewRequest(http.MethodPut, storeAddr+namePath, bytes.NewReader(b))
	if err != nil {
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusConflict {
		return
	}

	// read body into byte arr
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}

	return [32]byte{}, fmt.Errorf("bad status from execution server while storing web program: %s, %s", res.Status, string(body))
}

func StartWebProgram(pk keys.PK, name string, args WebArgs) (rets WebRets, err error) {
	res, err := http.Post(addr+"/web/"+pk.EncodeNoPrefix()+"/"+url.PathEscape(name), "", bytes.NewReader(args.Encode()))
	if err != nil {
		return WebRets{}, fmt.Errorf("failed to start web program: %v", err)
	}

	// we need the body either way
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return WebRets{}, fmt.Errorf("failed to read response body while starting web program: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return WebRets{}, fmt.Errorf("bad status from execution server while starting web program: %s, %s", res.Status, b)
	}

	// deserialise it
	if rets, err = DecodeRets[WebRets](b); err != nil {
		return WebRets{}, fmt.Errorf("failed to decode response body while starting web program: %v", err)
	}
	return
}
