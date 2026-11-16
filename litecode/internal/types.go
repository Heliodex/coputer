// Package internal contains code shared between the vm and types packages.
package internal

import "fmt"

type (
	// CoError is a custom error type used in coroutines that includes debugging information.
	CoError struct {
		Line uint32
		Dbgname, Path string
		Sub error
	}

	// VM internals types

	OpInfo struct {
		Mode, KMode uint8
		HasAux bool
	}

	// Coroutine types

	// Status represents the status of a coroutine.
	Status uint8

	// Val represents any possible VM stack/register value.
	//
	// It is a tagged union: Num carries a number inline (no boxing), and Ref is the discriminant plus payload for every other kind. Ref is nil for nil, a marker for numbers and booleans, and the boxed payload otherwise. The zero value is Luau nil. Use the constructors and accessors in the types package rather than the fields directly.
	Val struct {
		Num float64
		Ref any
	}

	Inst struct {
		K Val
		// KI holds the K0, K1, K2 import path components (up to 3 layers deep), used only by GETIMPORT
		KI *[3]string
		KC, Opcode, KMode, B, C uint8
		A, D int32
		Aux uint32
		KN bool
	}

	Proto struct {
		Dbgname string
		Code []Inst
		InstLineInfo, Protos []uint32
		MaxStackSize, NumParams, Nups uint8
	}

	// simpler compilation, deserialisation, and loading API

	Deserialised struct {
		MainProto *Proto
		ProtoList []*Proto
	}

	Deserpath struct {
		Deserialised

		Dbgpath string
	}

	// Yield represents a coroutine yield, containing the return values or error if one occurred.
	Yield struct {
		Rets []Val
		Err error
	}
)

// Coroutine stati
const (
	CoNotStarted Status = iota
	CoSuspended
	CoRunning
	CoNormal
	CoDead
)

func (e *CoError) Error() string {
	// MUCH better than previous
	return fmt.Sprintf("%s:%d: function %s\n%s", e.Path, e.Line, e.Dbgname, e.Sub.Error())
}
