// Package internal contains code shared between the vm and types packages.
package internal

import "fmt"

// Val represents any possible VM stack/register value.
type Val any

// VM internals types

type OpInfo struct {
	Mode, KMode uint8
	HasAux      bool
}

type Inst struct {
	OpInfo

	K Val
	// K0, K1, K2 for imports (up to 3 lay.ers.deep)
	K0, K1, K2          string
	KC, Opcode, A, B, C uint8
	D, E                int32
	Aux                 uint32
	KN                  bool
}

type Proto struct {
	Dbgname              string
	K                    []Val
	Code                 []*Inst
	InstLineInfo, Protos []uint32
	Dbgcode              []uint8

	// LineDefined uint32
	MaxStackSize, NumParams, Nups uint8
	LineInfoEnabled               bool
}

// simpler compilation, deserialisation, and loading API

type Deserialised struct {
	MainProto *Proto
	ProtoList []*Proto
}

type Deserpath struct {
	Deserialised
	Dbgpath string
}

// Coroutine types

// Status represents the status of a coroutine.
type Status uint8

// Coroutine stati
const (
	CoNotStarted Status = iota
	CoSuspended
	CoRunning
	CoNormal
	CoDead
)

// CoError is a custom error type used in coroutines that includes debugging information.
type CoError struct {
	Line          uint32
	Dbgname, Path string
	Sub           error
}

func (e *CoError) Error() string {
	// MUCH better than previous
	return fmt.Sprintf("%s:%d: function %s\n%s", e.Path, e.Line, e.Dbgname, e.Sub.Error())
}

// Yield represents a coroutine yield, containing the return values or error if one occurred.
type Yield struct {
	Rets []Val
	Err  error
}
