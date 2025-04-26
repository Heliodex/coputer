// package vm implements Litecode's deterministic Luau virtual machine and standard library toolkit.
package vm

import (
	"errors"
	"fmt"
	"iter"
	"maps"
	"math"
	"net/url"
	"reflect"
	"slices"
	"strings"
)

func arrayKey(k Val) (int, bool) {
	fk, ok := k.(float64)
	if !ok {
		return 0, false
	}

	ik := int(fk)
	return ik, float64(ik) == fk && 1 <= ik
}

func mapKeySort(a, b Val) int {
	// It doesn't have to be pretty for map keys
	// (in fact, the reference implementation of Luau has a rather insane sort order)
	// It just has to be DETERMINISTIC
	return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
}

func iterArray(array []Val, y func(Val, Val) bool) {
	for i, v := range array {
		if v != nil && !y(float64(i+1), v) {
			return
		}
	}
}

func iterHash(hash map[Val]Val, y func(Val, Val) bool) {
	// order keys in map
	keys := make([]Val, 0, len(hash))
	for k := range hash {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, mapKeySort)
	for _, k := range keys {
		if !y(k, hash[k]) {
			return
		}
	}
}

// Q why are tables like this
// A:
// 1: the reference implementation of tables is too complex: rehashing and resizing is a pain but not too bad, array boundaries are worse and I don't want 1.5k lines of code just for that, and Go does a resizing-like thing automatically with slices anyway
// 2: the way nodes are implemented works well in C++ and not in Go (plus I don't know if it's actually O(1) for node lookups??)
// 3: rehashing etc is slower than just using a slice... somehow. most of this program is between 10-20x slower than the reference implementation, but the tables (which were previously like 50x slower) are now only like 2-3x slower for large allocations (bench/largealloc.luau)
// 4: having an array part is actually nice for iteration and for large tables (as opposed to the lua4 way, where it's *just* a hash part), the way it's done here is simpler though we have to move stuff around and between the array and node parts more explicitly
// 5: very weird quirks arise from table length implementations etc. the nil stuff can easily be forgiven, it's the stuff with creating a table and getting a length afterwards (see tests/clear.luau) that is fucking devilish; this is one of the few parts that puts Luau, as the language at the top of my favourites list, in jeopardy
// 6: we don't actually break *that* much compatibility doing it this way, right??
// 7: if anyone tells you tables are simple THEY ARE LYING, CALL THEM OUT ON THEIR SHIT

// Table represents a Luau table, with resizeable array and hash parts. Luau type `table`
type Table struct {
	Array    []Val
	Hash     map[Val]Val
	readonly bool
}

// Len returns the length of the array part of the table (the length of the array up until the first nil).
func (t *Table) Len() int {
	if t.Array == nil {
		return 0
	}
	return len(t.Array)
}

// SetHash updates or deletes a key-value pair in the hash part of the table.
func (t *Table) SetHash(k, v Val) {
	if t.Hash == nil {
		if v == nil {
			return
		}
		t.Hash = map[Val]Val{k: v}
	} else if v == nil {
		delete(t.Hash, k)
	} else {
		t.Hash[k] = v
	}
}

// SetArray sets a value at an integer index, plcing it into the Array part or the Hash part and resizing each as appropriate.
func (t *Table) SetArray(i int, v Val) {
	if t.Array == nil {
		if i == 1 {
			t.Array = []Val{v}
			return
		}
	} else if l := len(t.Array); i < l+1 {
		if v != nil {
			// set in the array portion
			t.Array[i-1] = v
			return
		}

		// cut the array portion
		after := t.Array[i:]
		t.Array = t.Array[:i-1]

		// move the rest to the hash part
		for i2, v2 := range after {
			t.SetHash(float64(i+i2), v2)
		}
		return
	} else if i == l+1 {
		// append to the end
		t.Array = append(t.Array, v)

		// check if we can move some stuff from the hash part to the array part
		if t.Hash == nil {
			return
		}

		for f2 := float64(l + 2); ; f2++ {
			v2, ok := t.Hash[f2]
			if !ok {
				break
			}
			t.Array = append(t.Array, v2)
			delete(t.Hash, f2)
		}
		return
	}
	// add to the hash part instead
	t.SetHash(float64(i), v)
}

// ForceSet sets a table value at a key, regardless of whether the table is readonly.
func (t *Table) ForceSet(k, v Val) {
	if ak, ok := arrayKey(k); ok {
		t.SetArray(ak, v)
		return
	}
	t.SetHash(k, v)
}

// Set sets a table value at a key, returning an error if the table is readonly.
func (t *Table) Set(k, v Val) error {
	if t.readonly {
		return errors.New("attempt to modify a readonly table")
	}
	t.ForceSet(k, v)
	return nil
}

// GetHash returns a value at a key, only searching the hash part of the table.
func (t *Table) GetHash(k Val) Val {
	if t.Hash == nil {
		return nil
	}
	return t.Hash[k]
}

// Get returns a value at a key in the table.
func (t *Table) Get(k Val) Val {
	if ak, ok := arrayKey(k); ok && ak <= t.Len() {
		return t.Array[ak-1]
	}
	return t.GetHash(k)
}

// Iter returns an iterator over the table, yielding key-value pairs in a deterministic order.
func (t *Table) Iter() iter.Seq2[Val, Val] {
	return func(y func(Val, Val) bool) {
		if t.Array != nil {
			iterArray(t.Array, y)
		}
		if t.Hash != nil {
			iterHash(t.Hash, y)
		}
	}
}

// Vector represents a 3-wide or 4-wide vector value. Luau type `vector`
type Vector [4]float32

// bit32 extraction
func extract(n, field, width int) uint32 {
	return uint32(n>>field) & uint32(math.Pow(2, float64(width))-1)
}

// opList contains information about the instruction, each instruction is defined in this format:
// { Name, Mode, KMode, HasAux }
// Mode specifies what type of registers the instruction uses if any
//		0 = NONE
//		1 = A
//		2 = AB
//		3 = ABC
//		4 = AD
//		5 = AE
// KMode specifies if the instruction has a register that holds a constant table index, which will be directly converted to the constant in the 2nd pass
//		0 = NONE
//		1 = AUX
//		2 = C
//		3 = D
//		4 = AUX import
//		5 = AUX boolean low 1 bit
//		6 = AUX number low 24 bits
// HasAux boolean specifies whether the instruction is followed up with an AUX word, which may be used to execute the instruction.

type opInfo struct {
	mode, kMode uint8
	hasAux      bool
}

var opList = [83]opInfo{
	{0, 0, false}, // NOP
	{0, 0, false}, // BREAK
	{1, 0, false}, // LOADNIL
	{3, 0, false}, // LOADB
	{4, 0, false}, // LOADN
	{4, 3, false}, // LOADK
	{2, 0, false}, // MOVE
	{1, 1, true},  // GETGLOBAL
	{1, 1, true},  // SETGLOBAL
	{2, 0, false}, // GETUPVAL
	{2, 0, false}, // SETUPVAL
	{1, 0, false}, // CLOSEUPVALS
	{4, 4, true},  // GETIMPORT
	{3, 0, false}, // GETTABLE
	{3, 0, false}, // SETTABLE
	{3, 1, true},  // GETTABLEKS
	{3, 1, true},  // SETTABLEKS
	{3, 0, false}, // GETTABLEN
	{3, 0, false}, // SETTABLEN
	{4, 0, false}, // NEWCLOSURE
	{3, 1, true},  // NAMECALL
	{3, 0, false}, // CALL
	{2, 0, false}, // RETURN
	{4, 0, false}, // JUMP
	{4, 0, false}, // JUMPBACK
	{4, 0, false}, // JUMPIF
	{4, 0, false}, // JUMPIFNOT
	{4, 0, true},  // JUMPIFEQ
	{4, 0, true},  // JUMPIFLE
	{4, 0, true},  // JUMPIFLT
	{4, 0, true},  // JUMPIFNOTEQ
	{4, 0, true},  // JUMPIFNOTLE
	{4, 0, true},  // JUMPIFNOTLT
	{3, 0, false}, // ADD
	{3, 0, false}, // SUB
	{3, 0, false}, // MUL
	{3, 0, false}, // DIV
	{3, 0, false}, // MOD
	{3, 0, false}, // POW
	{3, 2, false}, // ADDK
	{3, 2, false}, // SUBK
	{3, 2, false}, // MULK
	{3, 2, false}, // DIVK
	{3, 2, false}, // MODK
	{3, 2, false}, // POWK
	{3, 0, false}, // AND
	{3, 0, false}, // OR
	{3, 2, false}, // ANDK
	{3, 2, false}, // ORK
	{3, 0, false}, // CONCAT
	{2, 0, false}, // NOT
	{2, 0, false}, // MINUS
	{2, 0, false}, // LENGTH
	{2, 0, true},  // NEWTABLE
	{4, 3, false}, // DUPTABLE
	{3, 0, true},  // SETLIST
	{4, 0, false}, // FORNPREP
	{4, 0, false}, // FORNLOOP
	{4, 8, true},  // FORGLOOP
	{4, 0, false}, // FORGPREP_INEXT
	{3, 1, true},  // FASTCALL3
	{4, 0, false}, // FORGPREP_NEXT
	{0, 0, false}, // FORGLOOP_NEXT (deprecated)
	{2, 0, false}, // GETVARARGS
	{4, 3, false}, // DUPCLOSURE
	{1, 0, false}, // PREPVARARGS
	{1, 1, true},  // LOADKX
	{5, 0, false}, // JUMPX
	{3, 0, false}, // FASTCALL
	{5, 0, false}, // COVERAGE
	{2, 0, false}, // CAPTURE
	{3, 7, false}, // SUBRK
	{3, 7, false}, // DIVRK
	{3, 0, false}, // FASTCALL1
	{3, 0, true},  // FASTCALL2
	{3, 1, true},  // FASTCALL2K
	{4, 0, false}, // FORGPREP
	{4, 5, true},  // JUMPXEQKNIL
	{4, 5, true},  // JUMPXEQKB
	{4, 6, true},  // JUMPXEQKN
	{4, 6, true},  // JUMPXEQKS
	{3, 0, false}, // IDIV
	{3, 2, false}, // IDIVK
}

// Functions and Tables are used as pointers normally, as they need to be hashed

// Function represents a native or wrapped Luau function. Luau type `function`
type Function struct {
	// Run is the native body of the function. Its coroutine argument is used to run the function in a coroutine.
	Run  *func(co *Coroutine, args ...Val) (r []Val, err error)
	name string
}

func fn(name string, f func(co *Coroutine, args ...Val) (r []Val, err error)) Function {
	return Function{&f, name}
}

// Status represents the status of a coroutine.
type Status uint8

// Coroutine stati
const (
	CoSuspended Status = iota
	CoRunning
	CoNormal
	CoDead
)

type yield struct {
	rets []Val
	err  error
}

type debugging struct {
	// top     int
	line    uint32
	enabled bool
	opcode  uint8
	dbgname string
}

// Env represents a global Luau environment.
type Env map[Val]Val

// AddFn adds a function to the environment.
func (e *Env) AddFn(f Function) {
	if *e == nil {
		*e = Env{f.name: f}
	} else {
		(*e)[f.name] = f
	}
}

// ProgramType represents the type of a program.
type ProgramType uint8

const (
	// TestProgramType represents the type of a test program.
	// Test programs are to be used for debugging and testing purposes only.
	TestProgramType ProgramType = iota
	// WebProgramType represents the type of a web program.
	WebProgramType
)

// ProgramArgs represents the arguments passed to a program.
type ProgramArgs interface {
	Type() ProgramType
}

// ProgramRets represents the response returned from a program.
type ProgramRets interface {
	Type() ProgramType
}

// Test programs

// Type returns WebProgramType.
func (WebRets) Type() ProgramType {
	return WebProgramType
}

// TestArgs stores the arguments passed to a test program.
type TestArgs struct{}

// Type returns TestProgramType.
func (TestArgs) Type() ProgramType {
	return TestProgramType
}

// TestRets stores the response returned from a test program.
type TestRets struct{}

func (r1 TestRets) Equal(r2 TestRets) error {
	return nil
}

// Type returns TestProgramType.
func (TestRets) Type() ProgramType {
	return TestProgramType
}

// Web programs

// WebUrl represents a parsed URL and its properties.
type WebUrl struct {
	Rawpath  string            `json:"rawpath"`
	Path     string            `json:"path"`
	Rawquery string            `json:"rawquery"`
	Query    map[string]string `json:"query"`
}

func queryToMap(q url.Values) (m map[string]string) {
	m = make(map[string]string, len(q))
	for k, v := range q {
		m[k] = strings.Join(v, "")
	}

	return
}

func WebUrlFromString(s string) (wurl WebUrl, err error) {
	url, err := url.Parse(s)
	if err != nil {
		return
	}

	wurl.Rawpath = s
	wurl.Path = url.Path
	wurl.Rawquery = url.RawQuery
	wurl.Query = queryToMap(url.Query())
	return
}

// WebArgs stores the arguments passed to a web program.
type WebArgs struct {
	Url     WebUrl            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

// Type returns WebProgramType.
func (WebArgs) Type() ProgramType {
	return WebProgramType
}

// WebRets stores the response returned from a web program.
type WebRets struct {
	StatusCode    int               `json:"statuscode"`
	StatusMessage string            `json:"statusmessage"`
	Headers       map[string]string `json:"headers"`
	Body          []byte            `json:"body"`
}

func (r1 WebRets) Equal(r2 WebRets) error {
	if r1.StatusCode != r2.StatusCode {
		return fmt.Errorf("Expected StatusCode %b, got %b", r1.StatusCode, r2.StatusCode)
	} else if r1.StatusMessage != r2.StatusMessage {
		return fmt.Errorf("Expected StatusMessage %s, got %s", r1.StatusMessage, r2.StatusMessage)
	} else if !maps.Equal(r1.Headers, r2.Headers) {
		return fmt.Errorf("Expected Headers %v, got %v", r1.Headers, r2.Headers)
	} else if !slices.Equal(r1.Body, r2.Body) {
		return fmt.Errorf("Expected Body %q, got %q", string(r1.Body), string(r2.Body))
	}

	return nil
}

// Coroutine represents a Luau coroutine, including the main coroutine. Luau type `thread`
type Coroutine struct {
	body              Function
	env               Env
	filepath, dbgpath string   // actually does well here
	requireHistory    []string // prevents cyclic module dependencies
	yield             chan yield
	resume            chan []Val
	dbg               *debugging
	compiler          *Compiler
	status            Status
	started           bool
	programArgs       ProgramArgs // idk how
}

// Error is a custom error type that includes debugging information.
type Error struct {
	dbg  *debugging
	path string
	sub  error
}

func (e *Error) Error() string {
	var eb strings.Builder

	for e != nil {
		err, ok := e.sub.(*Error)
		if !ok {
			eb.WriteString(
				fmt.Sprintf("%s:%d: function %s\n%s", e.path, e.dbg.line, e.dbg.dbgname, e.sub))
			break
		}
		eb.WriteString(
			fmt.Sprintf("%s:%d: function %s\n", e.path, e.dbg.line, e.dbg.dbgname))
		e = err
	}

	return eb.String()
}

func createCoroutine(body Function, currentCo *Coroutine) *Coroutine {
	// first time i actually ran into the channel axiom issues
	return &Coroutine{
		body:     body,
		filepath: currentCo.filepath,
		dbgpath:  currentCo.dbgpath,
		yield:    make(chan yield, 1),
		resume:   make(chan []Val, 1),
	}
}

// Error yields an error to the coroutine, killing it shortly after.
func (co *Coroutine) Error(err error) {
	co.yield <- yield{nil, &Error{co.dbg, co.dbgpath, err}}

	// ostensibly blocks forever, but the coroutine is dead/to be killed very soon so it doesn't matter
	select {}
}

func startCoroutine(co *Coroutine, args []Val) {
	// fmt.Println(" RG calling coroutine body with", args)
	r, err := (*co.body.Run)(co, args...)

	// fmt.Println("RG  yielding", r)
	co.yield <- yield{r, err}
	// fmt.Println("RG  yielded", r)

	co.status = CoDead
	if len(co.yield) == 0 {
		// finish up
		// fmt.Println("RG  yielding, finishing up")
		co.yield <- yield{}
		// fmt.Println("RG  yielding, finished up")
	}
}

// Resume executes the coroutine with the provided arguments, starting it with the given arguments if it is not already started, otherwise resuming it and passing the argument values back to the yielded function.
func (co *Coroutine) Resume(args ...Val) (r []Val, err error) {
	if !co.started {
		// fmt.Println("RM  starting", args)
		co.started = true
		co.status = CoRunning

		go startCoroutine(co, args)
	} else {
		co.status = CoRunning
		// fmt.Println("RM  resuming", args)
		co.resume <- args
		// fmt.Println("RM  resumed", args)
	}
	// fmt.Println("RM  waiting for yield")
	y := <-co.yield
	// fmt.Println("RM  waited for yield", y.rets)
	return y.rets, y.err
}

const luau_multret = -1

func vectorCtor(x, y, z, w float32) Vector {
	return Vector{x, y, z, w}
}

func namecallHandler(co *Coroutine, kv string, stack *[]Val, c1, c2 int) (ok bool, retList []Val, err error) {
	switch kv {
	case "format":
		str := (*stack)[c1].(string)
		args := (*stack)[c1+1 : c2+1]

		f, err := fmtstring(str, Args{Co: co, List: args, name: "format"})
		if err != nil {
			return false, nil, err
		}
		return true, []Val{f}, nil
	}
	return
}

var exts = Env{
	"math":      libmath,
	"table":     libtable,
	"string":    libstring,
	"coroutine": libcoroutine,
	"bit32":     libbit32,
	"utf8":      libutf8,
	// fuck os
	// and debug
	"buffer": libbuffer,
	"vector": libvector,

	// globals
	"type": MakeFn("type", global_type),
	// "typeof":   MakeFn("typeof", global_type)[1], // same because no metatables
	"ipairs":   MakeFn("ipairs", global_ipairs),
	"pairs":    MakeFn("pairs", global_pairs),
	"next":     MakeFn("next", global_next),
	"tonumber": MakeFn("tonumber", global_tonumber),
	"tostring": MakeFn("tostring", global_tostring),
	"_VERSION": "Luau", // todo: custom

	"require": MakeFn("require", global_require),
	"args":    libargs,
}

// var VectorSize = 4
// var AllowProxyErrors = false

type inst struct {
	opInfo

	K, K0, K1, K2          any
	A, B, C, D, E, KC, aux int
	opcode                 uint8
	KN                     bool
}

type proto struct {
	dbgname              string
	k                    []Val
	code                 []*inst
	instlineinfo, protos []uint32
	dbgcode              []uint8

	// linedefined uint32
	maxstacksize, numparams, nups uint8
	lineinfoenabled               bool
}

type deserialised struct {
	mainProto proto
	protoList []proto
}

func checkkmode(i *inst, k []Val) {
	switch i.kMode {
	case 1: // AUX
		if i.aux < len(k) { // sometimes huge for some reason
			i.K = k[i.aux]
		}
	case 2: // C
		i.K = k[i.C]
		// fmt.Println("SET K TO", i.K, "FROM", i.C)
	case 3: // D
		i.K = k[i.D]
	case 4: // AUX import
		extend := i.aux
		count := extend >> 30
		i.KC = count

		id0 := extend >> 20 & 0x3FF
		i.K0 = k[id0] // maybe can .(string) this
		// fmt.Println("AUX", i.K0)

		if count >= 2 {
			id1 := extend >> 10 & 0x3FF
			i.K1 = k[id1]
		}
		if count == 3 { // should never be 3
			id2 := extend & 0x3FF
			i.K2 = k[id2]
		}
	case 5: // AUX boolean low 1 bit
		i.K = extract(i.aux, 0, 1) == 1
		i.KN = extract(i.aux, 31, 1) == 1
	case 6: // AUX number low 24 bits
		i.K = k[extract(i.aux, 0, 24)]
		i.KN = extract(i.aux, 31, 1) == 1
	case 7: // B
		i.K = k[i.B]
	case 8: // AUX number low 16 bits
		i.K = i.aux & 0xF
	}
}

type stream struct {
	data []byte
	pos  uint32
}

func (s *stream) rByte() (b byte) {
	b = s.data[s.pos]
	s.pos++
	return
}

func (s *stream) rBool() bool {
	return s.rByte() != 0
}

func (s *stream) rWord() (w uint32) {
	w = uint32(s.data[s.pos]) |
		uint32(s.data[s.pos+1])<<8 |
		uint32(s.data[s.pos+2])<<16 |
		uint32(s.data[s.pos+3])<<24
	s.pos += 4
	return
}

func (s *stream) rFloat32() float32 {
	return math.Float32frombits(s.rWord())
}

func (s *stream) rFloat64() float64 {
	return math.Float64frombits(uint64(s.rWord()) | uint64(s.rWord())<<32)
}

func (s *stream) rVarInt() (r uint32) {
	for i := range 4 {
		v := uint32(s.rByte())
		r |= v & 0x7F << (i * 7)
		if v&0x80 == 0 {
			return
		}
	}
	return
}

func (s *stream) rString() string {
	size := s.rVarInt()
	if size == 0 {
		return ""
	}

	s.pos += size

	return string(s.data[s.pos-size : s.pos])
}

func (s *stream) CheckEnd() error {
	if s.pos != uint32(len(s.data)) {
		return errors.New("deserialiser position mismatch")
	}
	return nil
}

// reads either 1 or 2 words
func readInst(codeList *[]*inst, s *stream) bool {
	value := s.rWord()
	opcode := uint8(value & 0xFF)

	opinfo := opList[opcode]

	i := &inst{
		opInfo: opinfo,
		opcode: opcode,
	}

	*codeList = append(*codeList, i)

	switch opinfo.mode {
	case 1: // A
		i.A = int(value>>8) & 0xFF
	case 2: // AB
		i.A = int(value>>8) & 0xFF
		i.B = int(value>>16) & 0xFF
	case 3: // ABC
		i.A = int(value>>8) & 0xFF
		i.B = int(value>>16) & 0xFF
		i.C = int(value>>24) & 0xFF
	case 4: // AD
		i.A = int(value>>8) & 0xFF
		i.D = int(value>>16) & 0xFFFF
		if i.D >= 0x8000 {
			i.D -= 0x10000
		}
	case 5: // AE
		i.E = int(value>>8) & 0xFFFFFF
		if i.E >= 0x800000 {
			i.E -= 0x1000000
		}
	}

	if opinfo.hasAux {
		i.aux = int(s.rWord())

		*codeList = append(*codeList, &inst{})
	}

	return opinfo.hasAux
}

func readProto(stringList []string, s *stream) (p proto, err error) {
	p.maxstacksize, p.numparams, p.nups = s.rByte(), s.rByte(), s.rByte()

	s.rBool()            // isvararg
	s.rByte()            // -- flags
	s.pos += s.rVarInt() // typesize

	sizecode := s.rVarInt()

	for i := uint32(0); i < sizecode; i++ {
		if readInst(&p.code, s) {
			i++
		}
	}

	p.dbgcode = make([]uint8, sizecode)
	for i := range sizecode {
		p.dbgcode[i] = p.code[i].opcode
	}

	sizek := s.rVarInt()
	p.k = make([]Val, sizek) // crazy

	for i := range sizek {
		switch kt := s.rByte(); kt {
		case 0: // Nil
			p.k[i] = nil
		case 1: // Bool
			p.k[i] = s.rBool()
		case 2: // Number
			p.k[i] = s.rFloat64()
		case 3: // String
			p.k[i] = stringList[s.rVarInt()-1]
		case 4: // Import
			p.k[i] = float64(s.rWord()) // strange
		case 5: // Table
			dataLength := s.rVarInt()
			t := make([]uint32, dataLength)

			for j := range dataLength {
				t[j] = s.rVarInt() // whatever
			}

			p.k[i] = t
		case 6: // Closure
			p.k[i] = s.rVarInt()
		case 7: // Vector
			p.k[i] = vectorCtor(s.rFloat32(), s.rFloat32(), s.rFloat32(), s.rFloat32())
		default:
			return proto{}, fmt.Errorf("unknown ktype %d", kt)
		}
	}

	// -- 2nd pass to replace constant references in the instruction
	for i := range sizecode {
		checkkmode(p.code[i], p.k)
	}

	sizep := s.rVarInt()
	p.protos = make([]uint32, sizep)
	for i := range sizep {
		p.protos[i] = s.rVarInt() + 1
	}

	// p.linedefined =
	s.rVarInt()

	if dbgnamei := s.rVarInt(); dbgnamei == 0 {
		p.dbgname = "(??)"
	} else {
		p.dbgname = stringList[dbgnamei-1]
	}

	// -- lineinfo
	if p.lineinfoenabled = s.rBool(); p.lineinfoenabled {
		linegaplog2 := s.rByte()
		intervals := uint32((sizecode-1)>>linegaplog2) + 1

		lineinfo := make([]uint8, sizecode)
		var lastoffset uint8
		for j := range sizecode {
			lastoffset += s.rByte() // yeah this is how the reference implementation does it, whatever hpppens happens
			lineinfo[j] = lastoffset
		}

		abslineinfo := make([]uint32, intervals)
		var lastline uint32
		for i := range intervals {
			lastline += s.rWord()
			// fmt.Println("lastline", lastline)
			abslineinfo[i] = lastline // overflow babyy (faster than % (1 << 32))
		}

		p.instlineinfo = make([]uint32, sizecode)
		for i := range sizecode {
			// -- p->abslineinfo[pc >> p->linegaplog2] + p->lineinfo[pc];
			p.instlineinfo[i] = abslineinfo[i>>linegaplog2] + uint32(lineinfo[i])
		}
	}

	// -- debuginfo
	if s.rBool() {
		// fmt.Println("DEBUGINFO")
		for range s.rVarInt() { // sizel
			s.rVarInt()
			s.rVarInt()
			s.rVarInt()
			s.rByte()
		}
		for range s.rVarInt() { // sizeupvalues
			s.rVarInt()
		}
	}

	return
}

func deserialise(data []byte) (deserialised, error) {
	s := &stream{data: data}

	if luauVersion := s.rByte(); luauVersion == 0 {
		return deserialised{}, errors.New("the provided bytecode is an error message")
	} else if luauVersion != 6 {
		return deserialised{}, errors.New("the version of the provided bytecode is unsupported")
	} else if s.rByte() != 3 { // types version
		return deserialised{}, errors.New("the types version of the provided bytecode is unsupported")
	}

	stringCount := s.rVarInt()
	stringList := make([]string, stringCount)
	for i := range stringCount {
		stringList[i] = s.rString()
	}

	// userdataRemapping (not used in VM, left unused)
	for i := s.rBool(); i; i = s.rBool() {
		s.rVarInt()
	}

	protoCount := s.rVarInt()
	protoList := make([]proto, protoCount)
	for i := range protoCount {
		p, err := readProto(stringList, s)
		if err != nil {
			return deserialised{}, err
		}
		protoList[i] = p
	}

	mainProto := protoList[s.rVarInt()]
	mainProto.dbgname = "(main)"

	return deserialised{mainProto, protoList}, s.CheckEnd()
}

type iterator struct {
	args    chan *Table
	resume  chan *[]Val
	running bool
}

type upval struct {
	value   Val
	store   []Val
	index   int
	selfRef bool
}

func truthy(v Val) bool {
	return v != nil && v != false
}

const typeprefix = "vm."

var luautype = map[string]string{
	"nil":                          "nil",
	"float64":                      "number",
	"string":                       "string",
	"bool":                         "boolean",
	"*" + typeprefix + "Table":     "table",
	typeprefix + "Function":        "function",
	"*" + typeprefix + "Coroutine": "thread",
	"*" + typeprefix + "Buffer":    "buffer",
	typeprefix + "Vector":          "vector",
}

func invalidCompare(op, ta, tb string) error {
	return fmt.Errorf("attempt to compare %s %s %s", luautype[ta], op, luautype[tb])
}

func uncallableType(v string) error {
	return fmt.Errorf("attempt to call a %s value", luautype[v])
}

func invalidArithmetic(op, ta, tb string) error {
	return fmt.Errorf("attempt to perform arithmetic (%s) on %s and %s", op, luautype[ta], luautype[tb])
}

func invalidUnm(t string) error {
	return fmt.Errorf("attempt to perform arithmetic (unm) on %s", luautype[t])
}

func invalidFor(pos, t string) error {
	return fmt.Errorf("invalid 'for' %s (number expected, got %s)", pos, luautype[t])
}

func invalidLength(t string) error {
	return fmt.Errorf("attempt to get length of a %s value", luautype[t])
}

func invalidConcat(t1, t2 string) error {
	return fmt.Errorf("attempt to concatenate %s with %s", luautype[t1], luautype[t2])
}

func invalidIndex(ta string, val any) error {
	tb := luautype[TypeOf(val)]
	if tb == "string" {
		tb = fmt.Sprintf("'%v'", val)
	}

	return fmt.Errorf("attempt to index %v with %v", luautype[ta], tb)
}

// TypeOf returns the underlying VM datatype of a value as a string.
// This does not return the Luau type, as type() does.
func TypeOf(v Val) string {
	if v == nil { // prevent nil pointer dereference
		return "nil"
	}
	return reflect.TypeOf(v).String()
}

func aAdd(a, b Val) (Val, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa + fb, nil
	}

	va, ok3 := a.(Vector)
	vb, ok4 := b.(Vector)
	if ok3 && ok4 {
		return Vector{va[0] + vb[0], va[1] + vb[1], va[2] + vb[2], va[3] + vb[3]}, nil
	}

	return nil, invalidArithmetic("add", TypeOf(a), TypeOf(b))
}

func aSub(a, b Val) (Val, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa - fb, nil
	}

	va, ok3 := a.(Vector)
	vb, ok4 := b.(Vector)
	if ok3 && ok4 {
		return Vector{va[0] - vb[0], va[1] - vb[1], va[2] - vb[2], va[3] - vb[3]}, nil
	}

	return nil, invalidArithmetic("sub", TypeOf(a), TypeOf(b))
}

func aMul(a, b Val) (Val, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa * fb, nil
	}

	va, ok3 := a.(Vector)
	vb, ok4 := b.(Vector)
	if ok3 && ok4 {
		return Vector{va[0] * vb[0], va[1] * vb[1], va[2] * vb[2], va[3] * vb[3]}, nil
	} else if ok1 && ok4 {
		f := float32(fa)
		return Vector{f * vb[0], f * vb[1], f * vb[2], f * vb[3]}, nil
	} else if ok3 && ok2 {
		f := float32(fb)
		return Vector{va[0] * f, va[1] * f, va[2] * f, va[3] * f}, nil
	}

	return nil, invalidArithmetic("mul", TypeOf(a), TypeOf(b))
}

func aDiv(a, b Val) (Val, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa / fb, nil
	}

	va, ok3 := a.(Vector)
	vb, ok4 := b.(Vector)
	if ok3 && ok4 {
		return Vector{va[0] / vb[0], va[1] / vb[1], va[2] / vb[2], va[3] / vb[3]}, nil
	} else if ok1 && ok4 {
		f := float32(fa)
		return Vector{f / vb[0], f / vb[1], f / vb[2], f / vb[3]}, nil
	} else if ok3 && ok2 {
		f := float32(fb)
		return Vector{va[0] / f, va[1] / f, va[2] / f, va[3] / f}, nil
	}

	return nil, invalidArithmetic("div", TypeOf(a), TypeOf(b))
}

func aMod(a, b Val) (float64, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa - fb*math.Floor(fa/fb), nil
	}

	return 0, invalidArithmetic("mod", TypeOf(a), TypeOf(b))
}

func aPow(a, b Val) (float64, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return math.Pow(fa, fb), nil
	}

	return 0, invalidArithmetic("pow", TypeOf(a), TypeOf(b))
}

func aIdiv(a, b Val) (Val, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return math.Floor(fa / fb), nil
	}

	va, ok3 := a.(Vector)
	vb, ok4 := b.(Vector)
	if ok3 && ok4 {
		return Vector{
			f32Floor(va[0] / vb[0]),
			f32Floor(va[1] / vb[1]),
			f32Floor(va[2] / vb[2]),
			f32Floor(va[3] / vb[3]),
		}, nil
	} else if ok1 && ok4 {
		f := float32(fa)
		return Vector{
			f32Floor(f / vb[0]),
			f32Floor(f / vb[1]),
			f32Floor(f / vb[2]),
			f32Floor(f / vb[3]),
		}, nil
	} else if ok3 && ok2 {
		f := float32(fb)
		return Vector{
			f32Floor(va[0] / f),
			f32Floor(va[1] / f),
			f32Floor(va[2] / f),
			f32Floor(va[3] / f),
		}, nil
	}

	return nil, invalidArithmetic("idiv", TypeOf(a), TypeOf(b))
}

// vectors dont have these comparisons
func jumpLe(a, b Val) (bool, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa <= fb, nil
	}

	sa, ok1 := a.(string)
	sb, ok2 := b.(string)
	if ok1 && ok2 {
		return sa <= sb, nil
	}

	return false, invalidCompare("<=", TypeOf(a), TypeOf(b))
}

func jumpLt(a, b Val) (bool, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa < fb, nil
	}

	sa, ok1 := a.(string)
	sb, ok2 := b.(string)
	if ok1 && ok2 {
		return sa < sb, nil
	}

	return false, invalidCompare("<", TypeOf(a), TypeOf(b))
}

func jumpGt(a, b Val) (bool, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa > fb, nil
	}

	sa, ok1 := a.(string)
	sb, ok2 := b.(string)
	if ok1 && ok2 {
		return sa > sb, nil
	}

	return false, invalidCompare(">", TypeOf(a), TypeOf(b))
}

func jumpGe(a, b Val) (bool, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa >= fb, nil
	}

	sa, ok1 := a.(string)
	sb, ok2 := b.(string)
	if ok1 && ok2 {
		return sa >= sb, nil
	}

	return false, invalidCompare(">=", TypeOf(a), TypeOf(b))
}

func gettable(index, v Val) (Val, error) {
	switch t := v.(type) {
	case *Table:
		return t.Get(index), nil
	case Vector: // direction,,, and mmmagnitude!! oh yeah!!11!!
		switch index {
		case "x":
			return t[0], nil
		case "y":
			return t[1], nil
		case "z":
			return t[2], nil
			// case "w":
			// 	(*stack)[i.A] = t[3]
		}
		return nil, invalidIndex(typeprefix+"Vector", index)
	}
	return nil, invalidIndex(TypeOf(v), index)
}

type toWrap struct {
	proto        proto
	upvals       []*upval
	alive        *bool
	protolist    []proto
	env          Env
	requireCache map[string][]Val
}

func iterate(c *iterator) {
	args := *<-c.args
	c.args = nil // we're done here
	c.running = true
	// fmt.Println("-2- generating iterator", args)

	for i, v := range args.Iter() {
		if !c.running {
			return
		}
		// fmt.Println("-2- yielding", i, v)
		c.resume <- &[]Val{i, v}
		// fmt.Println("-2- yielded!")
	}

	c.resume <- nil
}

func execute(towrap toWrap, stack *[]Val, co *Coroutine, vargsList []Val, vargsLen uint8) (r []Val, err error) {
	p, upvals := towrap.proto, towrap.upvals
	pc, top, openUpvals, generalisedIterators := 1, -1, []*upval{}, map[inst]*iterator{}

	moveStack := func(src []Val, b, t int) {
		for t+b >= len(*stack) { // graah stack expansion
			*stack = append(*stack, nil)
		}

		for i := range b {
			if i >= len(src) {
				(*stack)[t+i] = nil
				continue
			}
			(*stack)[t+i] = src[i]
		}
	}

	var handlingBreak bool
	var i inst
	var op uint8

	// a a a a
	// stayin' alive
	// fmt.Println("starting with upvals", upvals)
	for *towrap.alive {
		if !handlingBreak {
			i = *p.code[pc-1]
			op = i.opcode
		}
		handlingBreak = false

		co.dbg.line = p.instlineinfo[pc-1]
		// co.dbg.top = top
		co.dbg.enabled = p.lineinfoenabled
		co.dbg.opcode = i.opcode
		co.dbg.dbgname = p.dbgname

		// if len(upvals) > 0 {
		// 	fmt.Println("upval", upvals[0])
		// }
		// fmt.Printf("OP %-2d PC %-3d UV %d\n", op, pc+1, len(upvals))

	mainswitch: // dw not using this 4 gotos
		switch op {
		case 0: // NOP
			pc++
			// -- Do nothing
		case 1: // BREAK
			op = p.dbgcode[pc]
			handlingBreak = true
		case 2: // LOADNIL
			pc++
			(*stack)[i.A] = nil
		case 3: // LOADB
			pc++
			(*stack)[i.A] = i.B == 1
			pc += i.C
		case 4: // LOADN
			pc++
			(*stack)[i.A] = float64(i.D) // never put an int on the stack
		case 5: // LOADK
			pc++
			// fmt.Println("LOADK", i.A, i.K)
			(*stack)[i.A] = i.K
		case 6: // MOVE
			pc++
			// we should (ALMOST) never have to change the size of the stack (p.maxstacksize)
			(*stack)[i.A] = (*stack)[i.B]
		case 7: // GETGLOBAL
			kv := i.K

			if e, ok := exts[kv]; ok {
				(*stack)[i.A] = e
			} else {
				(*stack)[i.A] = towrap.env[kv]
			}

			pc += 2 // -- adjust for aux
		case 8: // SETGLOBAL
			// LOL
			if kv, ok := i.K.(string); ok {
				if _, ok := exts[kv]; ok {
					return nil, fmt.Errorf("attempt to redefine global '%s'", kv)
				}
				return nil, fmt.Errorf("attempt to set global '%s'", kv)
			}
		case 9: // GETUPVAL
			pc++
			if uv := upvals[i.B]; uv.selfRef {
				(*stack)[i.A] = uv.value
			} else {
				// fmt.Println("GETTING UPVAL", uv)
				// fmt.Println("Setting stacka to", uv.store[uv.index])

				(*stack)[i.A] = uv.store[uv.index]
			}
		case 10: // SETUPVAL
			pc++
			if uv := upvals[i.B]; !uv.selfRef {
				uv.store[uv.index] = (*stack)[i.A]
			} else {
				uv.value = (*stack)[i.A]
			}
		case 11: // CLOSEUPVALS
			pc++
			for n, uv := range openUpvals {
				if uv == nil || uv.selfRef || uv.index < i.A {
					continue
				}
				// fmt.Println("closing upvalue", uv)
				uv.value = uv.store[uv.index]
				uv.store = nil
				uv.selfRef = true
				openUpvals[n] = nil
				// fmt.Println("closed", uv)
			}
		case 12: // GETIMPORT
			k0 := i.K0
			imp := exts[k0]
			if imp == nil {
				imp = towrap.env[k0]
			}

			count := i.KC

			if count >= 2 {
				t, ok := imp.(*Table)
				if !ok {
					return nil, invalidIndex("nil", i.K1)
				}

				imp = t.Get(i.K1)
				// fmt.Println("GETIMPORT2", i.A, (*stack)[i.A])

				if count == 3 {
					t1, ok := imp.(*Table)
					if !ok {
						return nil, invalidIndex(TypeOf(imp), i.K2)
					}

					imp = t1.Get(i.K2)
					// fmt.Println("GETIMPORT3", i.A, (*stack)[i.A])
				}
			}

			(*stack)[i.A] = imp

			pc += 2 // -- adjust for aux
		case 13: // GETTABLE
			pc++

			v, err := gettable((*stack)[i.C], (*stack)[i.B])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = v
		case 14: // SETTABLE
			pc++
			index := (*stack)[i.C]
			t, ok := (*stack)[i.B].(*Table) // SETTABLE or SETTABLEKS on a Vector actually does return "attempt to index vector with 'whatever'"
			if !ok {
				return nil, invalidIndex(TypeOf((*stack)[i.B]), index)
			}

			// fmt.Println("SETTABLE", index, (*stack)[i.A])
			if err := t.Set(index, (*stack)[i.A]); err != nil {
				return nil, err
			}
		case 15: // GETTABLEKS
			v, err := gettable(i.K, (*stack)[i.B])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = v

			pc += 2 // -- adjust for aux
		case 16: // SETTABLEKS
			index := i.K
			t, ok := (*stack)[i.B].(*Table)
			if !ok {
				// fmt.Println("indexing", typeOf((*stack)[i.B]), "with", index)
				return nil, invalidIndex(TypeOf((*stack)[i.B]), index)
			}

			if err := t.Set(index, (*stack)[i.A]); err != nil {
				return nil, err
			}

			pc += 2 // -- adjust for aux
		case 17: // GETTABLEN
			t := (*stack)[i.B].(*Table)
			idx := int(i.C + 1)

			(*stack)[i.A] = t.Get(float64(idx))

			pc++
		case 18: // SETTABLEN
			t := (*stack)[i.B].(*Table)
			if t.readonly {
				return nil, errors.New("attempt to modify a readonly table")
			} else if i, v := int(i.C+1), (*stack)[i.A]; 1 <= i || i > len(t.Array) {
				t.SetArray(i, v)
			} else {
				t.SetHash(float64(i), v)
			}

			pc++
		case 19: // NEWCLOSURE
			newProto := towrap.protolist[p.protos[i.D]-1]

			nups := newProto.nups
			towrap.upvals = make([]*upval, nups)

			// wrap is reused for closures
			towrap.proto = newProto

			(*stack)[i.A] = wrapclosure(towrap)
			// fmt.Println("WRAPPING WITH", uvs)

			// fmt.Println("nups", nups)
			for n := range nups {
				switch pseudo := p.code[pc]; pseudo.A {
				case 0: // -- value
					uv := &upval{
						value:   (*stack)[pseudo.B],
						selfRef: true,
					}
					uv.store = nil

					towrap.upvals[n] = uv
				case 1: // -- reference
					index := pseudo.B
					// fmt.Println("index", index, len(openUpvals))
					// for si, sv := range *stack {
					// 	fmt.Printf("  [%d] = %v\n", si, sv)
					// }

					var prev *upval
					if index < len(openUpvals) {
						prev = openUpvals[index]
					}

					if prev == nil {
						prev = &upval{
							store: *stack,
							index: index,
						}

						for len(openUpvals) <= index {
							openUpvals = append(openUpvals, nil)
						}
						openUpvals[index] = prev
					}

					towrap.upvals[n] = prev
					// fmt.Println("set upvalue", i, "to", prev)
				case 2: // -- upvalue
					// fmt.Println("moving", i, pseudo.B)
					towrap.upvals[n] = upvals[pseudo.B]
				}
				pc++
			}
			pc++
		case 20: // NAMECALL
			pc++
			// fmt.Println("NAMECALL")

			A, B := i.A, i.B
			kv := i.K.(string)
			// fmt.Println("kv", kv)

			(*stack)[A+1] = (*stack)[B]

			// -- Special handling for native namecall behaviour
			callInst := p.code[pc]
			callOp := callInst.opcode

			// -- Copied from the CALL handler
			callA, callB, callC := callInst.A, callInst.B, callInst.C

			var params int
			if callB == 0 {
				params = top - callA
			} else {
				params = callB - 1
			}

			ok, retList, err := namecallHandler(co, kv, stack, callA+1, callA+params)
			if err != nil {
				return nil, err
			} else if !ok {
				t := (*stack)[B].(*Table)

				if t.Hash == nil {
					(*stack)[A] = nil
				} else {
					(*stack)[A] = t.GetHash(kv)
				}
				break
			}

			pc += 2 // -- adjust for aux, Skip next CALL instruction

			i = *callInst
			op = callOp

			co.dbg.line = p.instlineinfo[pc-1]
			co.dbg.opcode = i.opcode

			retCount := len(retList)

			if callC == 0 {
				top = callA + retCount - 1
			} else {
				retCount = callC - 1
			}

			moveStack(retList, retCount, callA)
		case 21: // CALL
			pc++
			A, B, C := i.A, i.B, i.C

			var params int
			if B == 0 {
				params = top - A
			} else {
				params = B - 1
			}

			// fmt.Println(A, B, C, (*stack)[A], params)

			f := (*stack)[A]
			fn, ok := f.(Function)
			// fmt.Println("calling with", (*stack)[A+1:][:params])
			if !ok {
				return nil, uncallableType(TypeOf(f))
			}

			// fmt.Println("upvals1", len(upvals))
			retList, err := (*fn.Run)(co, (*stack)[A+1:][:params]...) // not inclusive
			// fmt.Println("upvals2", len(upvals))
			if err != nil {
				return nil, err
			}
			// fmt.Println("resultt", retList)
			retCount := len(retList)

			// fmt.Println("COUNT", retCount)
			if retCount == 1 { // requires should return only 1 value anyway
				if lc, ok := retList[0].(compiled); ok {
					// it's a require
					// fmt.Println("REQUIRE", lc.filepath)

					if c, ok := towrap.requireCache[lc.filepath]; ok {
						retList = c[len(c)-1:]
					} else {
						// since environments only store global libraries etc, using the same env here should be fine??
						c2, _ := loadmodule(lc, co.env, towrap.requireCache, lc.requireHistory, co.programArgs)
						reqrets, err := c2.Resume()
						if err != nil {
							return nil, err
						} else if len(reqrets) == 0 {
							return nil, errors.New("module must return a value")
						}

						// only the last return value (weird luau behaviour...)
						ret := reqrets[len(reqrets)-1]
						switch ret.(type) {
						case *Table, Function:
						default:
							return nil, errors.New("module must return a table or function")
						}

						retList = []Val{ret}
						towrap.requireCache[lc.filepath] = retList
					}
				}
			}

			// development checking lelell
			for _, v := range retList {
				switch v.(type) {
				case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
					panic(fmt.Sprintf("Hey idiot YOU RETURNED AN INTEGER INSTEAD OFA  FLOAT FROM YUR FUNCTION O MY GOD %v", v))
				case float32:
					panic(fmt.Sprintf("u  dun fukt up %v", v))
				}
			}

			if C == 0 {
				top = A + retCount - 1
			} else {
				retCount = C - 1
			}

			moveStack(retList, retCount, A)
		case 22: // RETURN
			pc++
			A, B := i.A, i.B
			b := B - 1

			// nresults
			if b == luau_multret {
				b = top - A + 1
			}

			return (*stack)[A:max(A+b, 0)], nil
		case 23, 24: // JUMP, JUMPBACK
			pc += i.D + 1
		case 25: // JUMPIF
			if truthy((*stack)[i.A]) {
				pc += i.D + 1
			} else {
				pc++
			}
		case 26: // JUMPIFNOT
			if !truthy((*stack)[i.A]) {
				pc += i.D + 1
			} else {
				pc++
			}
		case 27: // jump
			if (*stack)[i.A] == (*stack)[i.aux] {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 28:
			if j, err := jumpLe((*stack)[i.A], (*stack)[i.aux]); err != nil {
				return nil, err
			} else if j {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 29:
			if j, err := jumpLt((*stack)[i.A], (*stack)[i.aux]); err != nil {
				return nil, err
			} else if j {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 30:
			if (*stack)[i.A] != (*stack)[i.aux] {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 31:
			if j, err := jumpGt((*stack)[i.A], (*stack)[i.aux]); err != nil {
				return nil, err
			} else if j {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 32:
			if j, err := jumpGe((*stack)[i.A], (*stack)[i.aux]); err != nil {
				return nil, err
			} else if j {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 33: // arithmetic
			j, err := aAdd((*stack)[i.B], (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
			pc++
		case 34:
			j, err := aSub((*stack)[i.B], (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
			pc++
		case 35:
			j, err := aMul((*stack)[i.B], (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
			pc++
		case 36:
			j, err := aDiv((*stack)[i.B], (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
			pc++
		case 37:
			j, err := aMod((*stack)[i.B], (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
			pc++
		case 38:
			j, err := aPow((*stack)[i.B], (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
			pc++
		case 81:
			j, err := aIdiv((*stack)[i.B], (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
			pc++
		case 39: // arithmetik
			pc++
			j, err := aAdd((*stack)[i.B], i.K)
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 40:
			j, err := aSub((*stack)[i.B], i.K)
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
			pc++
		case 41:
			j, err := aMul((*stack)[i.B], i.K)
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
			pc++
		case 42:
			pc++
			j, err := aDiv((*stack)[i.B], i.K)
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 43:
			j, err := aMod((*stack)[i.B], i.K)
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
			pc++
		case 44:
			j, err := aPow((*stack)[i.B], i.K)
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
			pc++
		case 82:
			j, err := aIdiv((*stack)[i.B], i.K)
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
			pc++

		case 45: // logic AND
			pc++
			a := (*stack)[i.B]
			b := (*stack)[i.C]

			if truthy(a) {
				(*stack)[i.A] = b
			} else {
				(*stack)[i.A] = a
			}
		case 46: // logic OR
			pc++
			a := (*stack)[i.B]
			b := (*stack)[i.C]

			if truthy(a) {
				(*stack)[i.A] = a
			} else if truthy(b) {
				(*stack)[i.A] = b
			} else {
				(*stack)[i.A] = false
			}
		case 47: // logik AND
			pc++
			// fmt.Println("LOGIK")
			a := (*stack)[i.B]
			b := i.K

			if truthy(a) {
				(*stack)[i.A] = b
			} else {
				(*stack)[i.A] = a
			}
		case 48: // logik OR
			pc++
			// fmt.Println("LOGIK")
			a := (*stack)[i.B]
			b := i.K

			if truthy(a) {
				(*stack)[i.A] = a
			} else if truthy(b) {
				(*stack)[i.A] = b
			} else {
				(*stack)[i.A] = false
			}
		case 49: // CONCAT
			pc++
			s := strings.Builder{}

			var first int
			for n := i.B; n <= i.C; n++ {
				toWrite, ok := (*stack)[n].(string)
				if !ok {
					// ensure correct order of operands in error message
					return nil, invalidConcat(TypeOf((*stack)[n+first]), TypeOf((*stack)[n+1+first]))
				}
				s.WriteString(toWrite)
				first = -1
			}
			(*stack)[i.A] = s.String()
		case 50: // NOT
			pc++
			(*stack)[i.A] = !truthy((*stack)[i.B])
		case 51: // MINUS
			pc++
			a, ok := (*stack)[i.B].(float64)
			if !ok {
				return nil, invalidUnm(TypeOf((*stack)[i.B]))
			}

			(*stack)[i.A] = -a
		case 52: // LENGTH
			pc++
			switch t := (*stack)[i.B].(type) {
			case *Table:
				(*stack)[i.A] = float64(t.Len())
			case string:
				(*stack)[i.A] = float64(len(t))
			default:
				return nil, invalidLength(TypeOf(t))
			}
		case 53: // NEWTABLE
			(*stack)[i.A] = &Table{}

			pc += 2 // -- adjust for aux
		case 54: // DUPTABLE
			pc++
			serialised := &Table{}
			fmt.Println("TEMPLATING")
			for _, id := range i.K.([]uint32) { // template
				if err := serialised.Set(p.k[id], nil); err != nil { // constants
					return nil, err
				}
			}
			(*stack)[i.A] = serialised
		case 55: // SETLIST
			A, B := i.A, i.B
			c := i.C - 1

			if c == luau_multret {
				c = top - B + 1
			}

			s := (*stack)[A].(*Table)
			if s.readonly {
				return nil, errors.New("attempt to modify a readonly table")
			}

			// one-indexed lol
			for n, v := range (*stack)[B:min(B+c, len(*stack))] {
				s.SetArray(n+i.aux, v)
			}
			// (*stack)[A] = s

			pc += 2 // -- adjust for aux
		case 56: // FORNPREP
			pc++
			A := i.A

			index, ok := (*stack)[A+2].(float64)
			if !ok {
				return nil, invalidFor("initial value", TypeOf((*stack)[A+2]))
			}

			limit, ok := (*stack)[A].(float64)
			if !ok {
				return nil, invalidFor("limit", TypeOf((*stack)[A]))
			}

			step, ok := (*stack)[A+1].(float64)
			if !ok {
				return nil, invalidFor("step", TypeOf((*stack)[A+1]))
			}

			if step > 0 {
				if index > limit {
					pc += i.D
				}
			} else if limit > index {
				pc += i.D
			}
		case 57: // FORNLOOP
			pc++
			A := i.A
			limit := (*stack)[A].(float64)
			step := (*stack)[A+1].(float64)
			init := (*stack)[A+2].(float64) + step

			(*stack)[A+2] = init

			if step > 0 {
				if init <= limit {
					pc += i.D
				}
			} else if limit <= init {
				pc += i.D
			}
		case 58: // FORGLOOP
			A := i.A
			res := i.K.(int)

			top = A + 6

			switch it := (*stack)[A].(type) {
			case Function:
				// fmt.Println("IT func", fn, (*stack)[A+1], (*stack)[A+2])
				vals, err := (*it.Run)(co, (*stack)[A+1], (*stack)[A+2])
				if err != nil {
					return nil, err
				}

				moveStack(vals, res, A+3)
				// fmt.Println(A+3, (*stack)[A+3])

				if (*stack)[A+3] == nil {
					pc += 2
					break mainswitch
				}

				(*stack)[A+2] = (*stack)[A+3]
				pc += i.D + 1
				break mainswitch
			case *Table:
				// fmt.Println("GETTING GENITER", typeOf(it))
				iter := *generalisedIterators[i]

				if !iter.running {
					// fmt.Println("-1- sending thru the wire")
					// fmt.Println((*stack)[A+1], (*stack)[A+2]) // <nil> <nil>
					iter.args <- it
					// fmt.Println("-1- sent")
				}

				vals := <-iter.resume
				// fmt.Println("-1- received!", vals)

				if vals == nil {
					delete(generalisedIterators, i)
					pc += 2
					break mainswitch
				}

				moveStack(*vals, res, A+3)

				(*stack)[A+2] = (*stack)[A+3]
				pc += i.D + 1
			default:
				return nil, fmt.Errorf("attempt to iterate over a %s value", TypeOf(it))
			}
		case 59, 61: // FORGPREP_INEXT, FORGPREP_NEXT
			if _, ok := (*stack)[i.A].(Function); !ok {
				return nil, fmt.Errorf("attempt to iterate over a %s value", TypeOf((*stack)[i.A])) // -- encountered non-function value
			}
			pc += i.D + 1
		case 60: // FASTCALL3
			// Skipped
			pc += 2 // adjust for aux
		case 63: // GETVARARGS
			pc++
			A := i.A
			b := i.B - 1

			// fmt.Println("MULTRET", b, vargsLen)
			if b == luau_multret {
				b = int(vargsLen)
				top = A + b - 1
			}

			// stack may get expanded here
			// (MAX STACK SIZE IS A LIE!!!!!!!!!!!!!!!!!!!!!!!)
			moveStack(vargsList, b, A)
		case 64: // DUPCLOSURE
			newProto := towrap.protolist[i.K.(uint32)]

			nups := newProto.nups
			towrap.upvals = make([]*upval, nups)

			// reusing wrapping again bcause we're eco friendly
			towrap.proto = newProto

			(*stack)[i.A] = wrapclosure(towrap)

			for i := range nups {
				switch pseudo := p.code[pc]; pseudo.A {
				case 0: // value
					towrap.upvals[i] = &upval{
						value:   (*stack)[pseudo.B],
						selfRef: true,
					}

				// -- references dont get handled by DUPCLOSURE
				case 2: // upvalue
					towrap.upvals[i] = upvals[pseudo.B]
				}

				pc++
			}
			pc++
		case 65: // PREPVARARGS
			pc++
			// Handled by wrapper
		case 66: // LOADKX
			(*stack)[i.A] = float64(i.K.(uint32)) // kv (graah)

			pc += 2 // -- adjust for aux
		case 67: // JUMPX
			pc += i.E + 1
		case 68: // FASTCALL
			pc++
			// Skipped
		case 69: // COVERAGE
			pc++
			i.E++
		case 70: // CAPTURE
			// Handled by CLOSURE
			panic("encountered unhandled CAPTURE")
		case 71: // SUBRK
			pc++
			j, err := aSub(i.K, (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 72: // DIVRK
			pc++
			j, err := aDiv(i.K, (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 73: // FASTCALL1
			pc++
			// Skipped
		case 74, 75: // FASTCALL2, FASTCALL2K
			// Skipped
			pc += 2 // adjust for aux
		case 76: // FORGPREP
			pc += i.D + 1
			if _, ok := (*stack)[i.A].(Function); ok {
				break
			}

			loopInst := *p.code[pc-1]
			if generalisedIterators[loopInst] != nil {
				break
			}

			c := &iterator{
				args:   make(chan *Table),
				resume: make(chan *[]Val),
			}
			go iterate(c)
			// fmt.Println("SETTING GENITER", loopInst)
			generalisedIterators[loopInst] = c
		case 77: // JUMPXEQKNIL
			ra := (*stack)[i.A]

			if ra == nil != i.KN {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 78: //  JUMPXEQKB
			kv := i.K.(bool)
			ra := (*stack)[i.A].(bool)

			if ra == kv != i.KN {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 79: // JUMPXEQKN
			kv := i.K.(float64)
			ra := (*stack)[i.A].(float64)

			if ra == kv != i.KN {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 80: // JUMPXEQKS
			kv := i.K.(string)
			ra := (*stack)[i.A].(string)

			if ra == kv != i.KN {
				pc += i.D + 1
			} else {
				pc += 2
			}
		default:
			return nil, fmt.Errorf("unsupported opcode: %d", op)
		}
	}

	for _, uv := range openUpvals {
		if uv.selfRef {
			continue
		}
		uv.value = uv.store[uv.index]
		uv.store = nil
		uv.selfRef = true
	}

	for _, v := range generalisedIterators {
		v.running = false
	}

	if !*towrap.alive {
		// program was killed
		return nil, errors.New("program execution cancelled: timeout")
	}
	return
}

func wrapclosure(towrap toWrap) Function {
	proto := towrap.proto

	return fn("", func(co *Coroutine, args ...Val) (r []Val, err error) {
		maxs, np := proto.maxstacksize, proto.numparams // maxs 2 lel

		la := uint8(len(args)) // we can't have more than 255 args anyway right?
		// fmt.Println("MAX STACK SIZE", maxs)
		stack := make([]Val, maxs)
		copy(stack, args[:min(np, la)])

		var list []Val
		if np < la {
			list = args[np:]
		}

		dbg := &debugging{enabled: proto.lineinfoenabled, opcode: 255}
		co.dbg = dbg

		result, err := execute(towrap, &stack, co, list, max(la-np, 0))
		if !*towrap.alive {
			return
		} else if err != nil {
			return nil, &Error{dbg, co.dbgpath, err}
		}

		return result, nil
	})
}

func loadmodule(m compiled, env Env, requireCache map[string][]Val, requireHistory []string, args ProgramArgs) (co Coroutine, cancel func()) {
	alive := true

	towrap := toWrap{
		m.mainProto,
		[]*upval{},
		&alive,
		m.protoList,
		env,
		requireCache,
	}

	return Coroutine{
		body:           wrapclosure(towrap),
		env:            env,
		filepath:       m.filepath,
		dbgpath:        m.dbgpath,
		requireHistory: requireHistory,
		yield:          make(chan yield, 1),
		resume:         make(chan []Val, 1),
		dbg:            &debugging{opcode: 255},
		compiler:       m.compiler,
		programArgs:    args,
	}, func() { alive = false }
}
