package types

import "github.com/Heliodex/coputer/litecode/internal"

// debugging stores debugging information for a coroutine.
type debugging struct {
	Line uint32
	// enabled bool
	// opcode  uint8
	Name string
}

// Coroutine represents a Luau coroutine, including the main coroutine. Luau type `thread`
// As coroutines are compared by reference, this type must always be used as a pointer.
type Coroutine struct {
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

	select {} // yeah, we still need this because async things
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
