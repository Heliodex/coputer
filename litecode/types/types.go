// package types holds type definitions for the Litecode VM.
package types

type (
	// Val represents any possible Luau value. Luau type `any`
	Val    any
	ValMap map[Val]Val

	// Buffer represents a Luau byte buffer. Luau type`buffer`
	Buffer []byte

	// Vector represents a 3-wide or 4-wide vector value. Luau type `vector`
	Vector [4]float32
)
