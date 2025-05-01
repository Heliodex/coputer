// package types holds type definitions for the Litecode VM.
package types

import "fmt"

// Status represents the status of a coroutine.
type Status uint8

// Debugging stores debugging information for a coroutine.
type Debugging struct {
	// top,
	Line uint32
	// enabled bool
	// opcode  uint8
	Name string
}

// graah vm internals
// this package is getting out of hand
type OpInfo struct {
	Mode, KMode uint8
	HasAux      bool
}

type Inst struct {
	OpInfo

	// K0, K1, K2 for imports (up to 3 lay.ers.deep)
	K, K0, K1, K2       Val
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

// types.Compiler allows programs to be compiled and deserialised with a cache and given optimisation level.
type Compiler struct {
	Cache map[[32]byte]Deserpath
	O     uint8
}

// Luau types
type (
	// Val represents any possible Luau value. Luau type `any`
	Val any

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
		YieldChan         chan Yield
		ResumeChan        chan []Val
		Dbg               *Debugging
		Compiler          Compiler // for require()
		Status            Status
		Started           bool
		ProgramArgs       ProgramArgs // idk how
	}
)

// Coroutine stati
const (
	CoSuspended Status = iota
	CoRunning
	CoNormal
	CoDead
)

// Error yields an error to the coroutine, killing it shortly after.
func (co *Coroutine) Error(err error) {
	co.YieldChan <- Yield{nil, &CoError{
		Line:    co.Dbg.Line,
		Dbgname: co.Dbg.Name,
		Path:    co.Dbgpath,
		Sub:     err,
	}}
}

func startCoroutine(co *Coroutine, args []Val) {
	// fmt.Println(" RG calling coroutine body with", args)
	r, err := (*co.Body.Run)(co, args...)

	co.Status = CoDead
	// fmt.Println("RG  yielding", r)
	co.YieldChan <- Yield{r, err}
	// fmt.Println("RG  yielded", r)
}

// Resume executes the coroutine with the provided arguments, starting it with the given arguments if it is not already started, otherwise resuming it and passing the argument values back to the yielded function.
func (co *Coroutine) Resume(args ...Val) (r []Val, err error) {
	if !co.Started {
		// fmt.Println("RM  starting", args)
		co.Started = true
		co.Status = CoRunning

		go startCoroutine(co, args)
	} else {
		co.Status = CoRunning
		// fmt.Println("RM  resuming", args)
		co.ResumeChan <- args
		// fmt.Println("RM  resumed", args)
	}
	// fmt.Println("RM  waiting for yield")
	y := <-co.YieldChan
	// fmt.Println("RM  waited for yield", y.rets)
	return y.Rets, y.Err
}

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
