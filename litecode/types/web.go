package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
)

// WebArgsUrl represents a parsed URL and its properties.
type WebArgsUrl struct {
	Rawpath  string              `json:"rawpath"`
	Path     string              `json:"path"`
	Rawquery string              `json:"rawquery"`
	Query    map[string][]string `json:"query"`
}

// WebArgs stores the arguments passed to a web program.
type WebArgs struct {
	Url     WebArgsUrl        `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

// Type returns WebProgramType.
func (WebArgs) Type() ProgramType {
	return WebProgramType
}

func (args WebArgs) Encode() []byte {
	b, _ := json.Marshal(args)
	return b
}

// WebRets stores the response returned from a web program.
type WebRets struct {
	StatusCode int `json:"statuscode"`
	// StatusMessage string            `json:"statusmessage"` // removed 3 now
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

func (r1 WebRets) Equal(r2 WebRets) (err error) {
	if r1.StatusCode != r2.StatusCode {
		err = fmt.Errorf("Expected StatusCode %d, got %d", r1.StatusCode, r2.StatusCode)
	}
	if !maps.Equal(r1.Headers, r2.Headers) {
		err = errors.Join(err, fmt.Errorf("Expected Headers %v, got %v", r1.Headers, r2.Headers))
	}
	if !slices.Equal(r1.Body, r2.Body) {
		err = errors.Join(err, fmt.Errorf("Expected Body %q, got %q", string(r1.Body), string(r2.Body)))
	}

	return
}

// Type returns WebProgramType.
func (WebRets) Type() ProgramType {
	return WebProgramType
}

func (rets WebRets) Encode() []byte {
	b, _ := json.Marshal(rets)
	return b
}
