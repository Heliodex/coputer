package net

// communication with execution server

import (
	"bytes"
	"crypto/sha3"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	. "github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/wallflower/keys"
)

const (
	addr      = "http://localhost:2505"
	storeAddr = addr + "/store"
	// Maximum response body size (10MB)
	maxResponseSize = 10 << 20
)

// httpClient is a shared HTTP client with connection pooling and timeouts
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

func GetProfile(pk keys.PK) (programs []string, err error) {
	res, err := httpClient.Get(addr + "/" + pk.EncodeNoPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get programs: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status from execution server while getting programs: %s", res.Status)
	}

	// Limit response size to prevent memory exhaustion
	b, err := io.ReadAll(io.LimitReader(res.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body while getting programs: %v", err)
	}

	bprograms := bytes.Split(bytes.TrimSpace(b), []byte{'\n'})
	programs = make([]string, len(bprograms))
	for i, v := range bprograms {
		programs[i] = string(v)
	}
	return
}

func StoreProgram(pk keys.PK, name string, b []byte) (hash [32]byte, err error) {
	// fmt.Println("Storing program", pk.Encode(), name)
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

	res, err := httpClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusConflict {
		return
	}

	// read body into byte arr with size limit
	body, err := io.ReadAll(io.LimitReader(res.Body, maxResponseSize))
	if err != nil {
		return
	}

	return [32]byte{}, fmt.Errorf("bad status from execution server while storing web program: %s, %s", res.Status, string(body))
}

func StartWebProgram(pk keys.PK, name string, args WebArgs) (rets WebRets, err error) {
	res, err := httpClient.Post(addr+"/web/"+pk.EncodeNoPrefix()+"/"+url.PathEscape(name), "", bytes.NewReader(args.Encode()))
	if err != nil {
		return WebRets{}, fmt.Errorf("failed to start web program: %v", err)
	}
	defer res.Body.Close()

	// we need the body either way, but limit size to prevent memory exhaustion
	b, err := io.ReadAll(io.LimitReader(res.Body, maxResponseSize))
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
