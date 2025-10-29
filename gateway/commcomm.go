package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	. "github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/wallflower/keys"
)

const (
	// Maximum response body size (10MB)
	maxResponseSize = 10 << 20
)

var addr = "http://localhost:" + strconv.Itoa(commPort)

// httpClient is a shared HTTP client with connection pooling and timeouts
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

// god not another proxy
func GetProfile(pk keys.PK) (programs []string, err error) {
	res, err := httpClient.Get(addr + "/" + pk.EncodeNoPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get programs: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status from communication server while getting programs: %s", res.Status)
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

// lel
// but seriously, this is different to the StartWebProgram function in the communication system, even though it's identical, because it addresses the communication server instead of the execution server
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
		return WebRets{}, fmt.Errorf("bad status from communication server while starting web program: %s, %s", res.Status, b)
	}

	// deserialise it
	if rets, err = DecodeRets[WebRets](b); err != nil {
		return WebRets{}, fmt.Errorf("failed to decode response body while starting web program: %v", err)
	}
	return
}
