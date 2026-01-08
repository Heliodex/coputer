package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	. "github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/wallflower/keys"
)

var addr = "http://localhost:" + strconv.Itoa(commPort)

// god not another proxy
func GetProfile(pk keys.PK) (programs []string, err error) {
	res, err := http.Get(addr + "/" + pk.EncodeNoPrefix())
	if err != nil {
		return nil, fmt.Errorf("get programs: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status from communication server while getting programs: %s", res.Status)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body while getting programs: %v", err)
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
	res, err := http.Post(addr+"/web/"+pk.EncodeNoPrefix()+"/"+url.PathEscape(name), "", bytes.NewReader(args.Encode()))
	if err != nil {
		return WebRets{}, fmt.Errorf("start web program: %v", err)
	}

	// we need the body either way
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return WebRets{}, fmt.Errorf("read response body while starting web program: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return WebRets{}, fmt.Errorf("bad status from communication server while starting web program: %s, %s", res.Status, b)
	}

	// deserialise it
	if rets, err = DecodeRets[WebRets](b); err != nil {
		return WebRets{}, fmt.Errorf("decode response body while starting web program: %v", err)
	}
	return
}
