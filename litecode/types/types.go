// package types holds type definitions used for interfacing with the Litecode VM.
package types

import "github.com/Heliodex/coputer/litecode/internal"

// debugging stores debugging information for a coroutine.
type debugging struct {
	// top,
	Line uint32
	// enabled bool
	// opcode  uint8
	Name string
}

// types.Compiler allows programs to be compiled and deserialised with a cache and given optimisation level.
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
	}

	// Buffer represents a Luau byte buffer. Luau type`buffer`
	Buffer []byte

	// Vector represents a 3-wide or 4-wide vector value. Luau type `vector`
	Vector [4]float32

	// Coroutine represents a Luau coroutine, including the main coroutine. Luau type `thread`
	Coroutine struct {
		Body              Function
		Env               Env
		Filepath, Dbgpath string   // actually does well here
		RequireHistory    []string // prevents cyclic module dependencies
		YieldChan         chan internal.Yield
		ResumeChan        chan []Val
		Dbg               debugging
		Compiler          Compiler // for require()
		Status            internal.Status
		ProgramArgs       ProgramArgs // idk how
	}
)

// Error yields an error to the coroutine, killing it shortly after.
func (co *Coroutine) Error(err error) {
	co.YieldChan <- internal.Yield{
		Rets: nil,
		Err: &internal.CoError{
			Line:    co.Dbg.Line,
			Dbgname: co.Dbg.Name,
			Path:    co.Dbgpath,
			Sub:     err,
		},
	}
}

func startCoroutine(co *Coroutine, args []Val) {
	// fmt.Println(" RG calling coroutine body with", args)
	r, err := (*co.Body.Run)(co, args...)

	co.Status = internal.CoDead
	// fmt.Println("RG  yielding", r)
	co.YieldChan <- internal.Yield{
		Rets: r,
		Err:  err,
	}
	// fmt.Println("RG  yielded", r)
}

// Resume executes the coroutine with the provided arguments, starting it with the given arguments if it is not already started, otherwise resuming it and passing the argument values back to the yielded function.
func (co *Coroutine) Resume(args ...Val) (r []Val, err error) {
	if co.Status == internal.CoNotStarted {
		// fmt.Println("RM  starting", args)
		co.Status = internal.CoRunning

		go startCoroutine(co, args)
	} else {
		co.Status = internal.CoRunning
		// fmt.Println("RM  resuming", args)
		co.ResumeChan <- args
		// fmt.Println("RM  resumed", args)
	}
	// fmt.Println("RM  waiting for yield")
	y := <-co.YieldChan
	// fmt.Println("RM  waited for yield", y.rets)
	return y.Rets, y.Err
}

// Env represents a global Luau environment.
type Env map[Val]Val

// AddFn adds a function to the environment.
func (e *Env) AddFn(f Function) {
	if *e != nil {
		(*e)[f.Name] = f
		return
	}
	*e = Env{f.Name: f}
}
