// package types holds type definitions used for interfacing with the Litecode VM.
package types

import "github.com/Heliodex/coputer/litecode/internal"

// Compiler allows programs to be compiled and deserialised with a cache and given optimisation level.
type Compiler struct {
	Cache map[[32]byte]internal.Deserpath
	O     uint8
}

// Luau types
type (
	// Val represents any possible Luau value. Luau type `any`
	Val = internal.Val

	// Function represents a native or wrapped Luau function. Luau type `function`
	Function struct {
		// Run is the native body of the function. Its coroutine argument is used to run the function in a coroutine.
		Run  *func(co *Coroutine, args ...Val) (r []Val, err error)
		Name string
		Co *Coroutine // if in a different coroutine
	}

	// Buffer represents a Luau byte buffer. Luau type`buffer`
	// As buffers are compared by reference, this type must always be used as a pointer.
	Buffer []byte

	// Vector represents a 3-wide or 4-wide vector value. Luau type `vector`
	Vector [4]float32
)

// Env represents a global Luau environment.
type Env map[string]Val

// AddFn adds a function to the environment.
func (e *Env) AddFn(f Function) {
	if *e != nil {
		(*e)[f.Name] = f
		return
	}
	*e = Env{f.Name: f}
}
