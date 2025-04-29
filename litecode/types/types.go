// package types holds type definitions for the Litecode VM.
package types

type (
	// Val represents any possible Luau value. Luau type `any`
	Val any

	// Function represents a native or wrapped Luau function. Luau type `function`
	Function[co any] struct {
		// Run is the native body of the function. Its coroutine argument is used to run the function in a coroutine.
		Run  *func(co co, args ...Val) (r []Val, err error)
		Name string
	}

	// Buffer represents a Luau byte buffer. Luau type`buffer`
	Buffer []byte

	// Vector represents a 3-wide or 4-wide vector value. Luau type `vector`
	Vector [4]float32
)
