package litecode

import (
	"errors"
	"fmt"
	"iter"
	"math"
	"reflect"
	"slices"
	"strings"
)

func arrayKey(k any) (int, bool) {
	fk, ok := k.(float64)
	if !ok {
		return 0, false
	}

	ik := int(fk)
	return ik, float64(ik) == fk && 1 <= ik
}

// Q why are tables like this
// A
// 1: the reference implementation of tables is too complex: rehashing and resizing is a pain but not too bad, array boundaries are worse and I don't want 1.5k lines of code just for that, and Go does a resizing-like thing automatically with slices anyway
// 2: the way nodes are implemented works well in C++ and not in Go (plus I don't know if it's actually O(1) for node lookups??)
// 3: rehashing etc is slower than just using a slice... somehow. most of this program is between 10-20x slower than the reference implementation, but the tables (which were previously like 50x slower) are now only like 2-3x slower for large allocations (bench/largealloc.luau)
// 4: having an array part is actually nice for iteration and for large tables (as opposed to the lua4 way, where it's *just* a hash part), the way it's done here is simpler though we have to move stuff around and between the array and node parts more explicitly
// 5: very weird quirks arise from table length implementations etc. the nil stuff can easily be forgiven, it's the stuff with creating a table and getting a length afterwards (see tests/clear.luau) that is fucking devilish; this is one of the few parts that puts Luau, as the language at the top of my favourites list, in jeopardy
// 6: we don't actually break *that* much compatibility doing it this way, right??
// 7: if anyone tells you tables are simple THEY ARE LYING, CALL THEM OUT ON THEIR SHIT
type Table struct {
	array    *[]any
	node     *map[any]any
	readonly bool
}

func (t *Table) String() (s string) {
	if t.array == nil {
		s += "  array: nil"
	} else {
		s += fmt.Sprintf("  array: %v\n", *t.array)
	}

	if t.node == nil {
		s += "  node: nil"
	} else {
		s += fmt.Sprintf("  node:  %v", *t.node)
	}
	return
}

func NewTable(toHash [][2]any) *Table {
	// remember, no duplicates
	hash := make(map[any]any, len(toHash))
	for _, v := range toHash {
		hash[v[0]] = v[1]
	}
	return &Table{
		readonly: true,
		node:     &hash,
	}
}

// O(1) length, bitchesss
func (t *Table) Len() int {
	if t.array == nil {
		return 0
	}
	return len(*t.array)
}

func (t *Table) SetHash(i, v any) {
	if t.node == nil {
		if v == nil {
			return
		}
		t.node = &map[any]any{i: v}
		return
	}

	if v == nil {
		delete(*t.node, i)
	} else {
		(*t.node)[i] = v
	}
}

func (t *Table) SetArray(i int, v any) {
	if t.array == nil {
		if i == 1 {
			t.array = &[]any{v}
			return
		}

		t.SetHash(float64(i), v)
	} else if l := len(*t.array); i < l+1 {
		if v != nil {
			// set in the array portion
			(*t.array)[i-1] = v
			return
		}

		// cut the array portion
		after := (*t.array)[i:]
		*t.array = (*t.array)[:i-1]

		// move the rest to the hash part
		for i2, v2 := range after {
			t.SetHash(float64(i+i2), v2)
		}
	} else if i == l+1 {
		// append to the end
		*t.array = append(*t.array, v)

		// check if we can move some stuff from the hash part to the array part
		if t.node == nil {
			return
		}

		for i2 := l + 2; ; i2++ {
			if v2, ok := (*t.node)[float64(i2)]; ok {
				*t.array = append(*t.array, v2)
				delete(*t.node, float64(i2))
			} else {
				break
			}
		}
	} else {
		// add to the hash part instead
		t.SetHash(float64(i), v)
	}
}

func (t *Table) ForceSet(i, v any) {
	if ak, ok := arrayKey(i); ok {
		t.SetArray(ak, v)
		return
	}
	t.SetHash(i, v)
}

func (t *Table) Set(i, v any) error {
	if t.readonly {
		return errors.New("attempt to modify a readonly table")
	}
	t.ForceSet(i, v)
	return nil
}

func (t *Table) GetHash(i any) any {
	if t.node == nil {
		return nil
	}
	return (*t.node)[i]
}

func (t *Table) Get(i any) any {
	if ak, ok := arrayKey(i); ok && ak <= t.Len() {
		return (*t.array)[ak-1]
	}
	return t.GetHash(i)
}

func mapKeySort(a, b any) int {
	// It doesn't have to be pretty for map keys
	// (in fact, the reference implementation of Luau has a rather insane sort order)
	// It just has to be DETERMINISTIC
	return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
}

func iterArray(array []any, yield func(any, any) bool) {
	for i, v := range array {
		if v != nil && !yield(float64(i+1), v) {
			return
		}
	}
}

func iterHash(hash map[any]any, y func(any, any) bool) {
	// order keys in map
	keys := make([]any, 0, len(hash))
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

// 1.23 goes hard
func (t *Table) Iter() iter.Seq2[any, any] {
	return func(y func(any, any) bool) {
		if t.array != nil {
			iterArray(*t.array, y)
		}
		if t.node != nil {
			iterHash(*t.node, y)
		}
	}
}

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

var opList = [83]struct {
	name        string
	mode, kMode uint8
	hasAux      bool
}{
	{"NOP", 0, 0, false},
	{"BREAK", 0, 0, false},
	{"LOADNIL", 1, 0, false},
	{"LOADB", 3, 0, false},
	{"LOADN", 4, 0, false},
	{"LOADK", 4, 3, false},
	{"MOVE", 2, 0, false},
	{"GETGLOBAL", 1, 1, true},
	{"SETGLOBAL", 1, 1, true},
	{"GETUPVAL", 2, 0, false},
	{"SETUPVAL", 2, 0, false},
	{"CLOSEUPVALS", 1, 0, false},
	{"GETIMPORT", 4, 4, true},
	{"GETTABLE", 3, 0, false},
	{"SETTABLE", 3, 0, false},
	{"GETTABLEKS", 3, 1, true},
	{"SETTABLEKS", 3, 1, true},
	{"GETTABLEN", 3, 0, false},
	{"SETTABLEN", 3, 0, false},
	{"NEWCLOSURE", 4, 0, false},
	{"NAMECALL", 3, 1, true},
	{"CALL", 3, 0, false},
	{"RETURN", 2, 0, false},
	{"JUMP", 4, 0, false},
	{"JUMPBACK", 4, 0, false},
	{"JUMPIF", 4, 0, false},
	{"JUMPIFNOT", 4, 0, false},
	{"JUMPIFEQ", 4, 0, true},
	{"JUMPIFLE", 4, 0, true},
	{"JUMPIFLT", 4, 0, true},
	{"JUMPIFNOTEQ", 4, 0, true},
	{"JUMPIFNOTLE", 4, 0, true},
	{"JUMPIFNOTLT", 4, 0, true},
	{"ADD", 3, 0, false},
	{"SUB", 3, 0, false},
	{"MUL", 3, 0, false},
	{"DIV", 3, 0, false},
	{"MOD", 3, 0, false},
	{"POW", 3, 0, false},
	{"ADDK", 3, 2, false},
	{"SUBK", 3, 2, false},
	{"MULK", 3, 2, false},
	{"DIVK", 3, 2, false},
	{"MODK", 3, 2, false},
	{"POWK", 3, 2, false},
	{"AND", 3, 0, false},
	{"OR", 3, 0, false},
	{"ANDK", 3, 2, false},
	{"ORK", 3, 2, false},
	{"CONCAT", 3, 0, false},
	{"NOT", 2, 0, false},
	{"MINUS", 2, 0, false},
	{"LENGTH", 2, 0, false},
	{"NEWTABLE", 2, 0, true},
	{"DUPTABLE", 4, 3, false},
	{"SETLIST", 3, 0, true},
	{"FORNPREP", 4, 0, false},
	{"FORNLOOP", 4, 0, false},
	{"FORGLOOP", 4, 8, true},
	{"FORGPREP_INEXT", 4, 0, false},
	{"FASTCALL3", 3, 1, true},
	{"FORGPREP_NEXT", 4, 0, false},
	{"DEP_FORGLOOP_NEXT", 0, 0, false},
	{"GETVARARGS", 2, 0, false},
	{"DUPCLOSURE", 4, 3, false},
	{"PREPVARARGS", 1, 0, false},
	{"LOADKX", 1, 1, true},
	{"JUMPX", 5, 0, false},
	{"FASTCALL", 3, 0, false},
	{"COVERAGE", 5, 0, false},
	{"CAPTURE", 2, 0, false},
	{"SUBRK", 3, 7, false},
	{"DIVRK", 3, 7, false},
	{"FASTCALL1", 3, 0, false},
	{"FASTCALL2", 3, 0, true},
	{"FASTCALL2K", 3, 1, true},
	{"FORGPREP", 4, 0, false},
	{"JUMPXEQKNIL", 4, 5, true},
	{"JUMPXEQKB", 4, 5, true},
	{"JUMPXEQKN", 4, 6, true},
	{"JUMPXEQKS", 4, 6, true},
	{"IDIV", 3, 0, false},
	{"IDIVK", 3, 2, false},
}

type (
	Function *func(co *Coroutine, args ...any) (r Rets, err error)
	Status   uint8
)

func Fn(f func(co *Coroutine, args ...any) (r Rets, err error)) Function {
	return Function(&f)
}

const (
	CoSuspended Status = iota
	CoRunning
	CoNormal
	CoDead
)

type yield struct {
	rets Rets
	err  error
}

type Coroutine struct {
	body     Function
	env      map[any]any
	filepath string // lel nowhere else to put this
	yield    chan yield
	resume   chan Rets
	dbg      *debugging
	o        uint8
	status   Status
	started  bool
}

func createCoroutine(body Function) *Coroutine {
	// first time i actually ran into the channel axiom issues
	return &Coroutine{
		body:   body,
		yield:  make(chan yield, 1),
		resume: make(chan Rets, 1),
	}
}

func errorfmt(err error, d *debugging) error {
	op := "NONE"
	if d.opcode != 255 {
		op = opList[d.opcode].name
	}

	if d.enabled {
		// fmt.Println(d.instlineinfo)

		// PC removed for determinism between O levels
		return fmt.Errorf(
			"Opcode: %s\n%s:%d: %w",
			op,
			d.debugname,
			d.instlineinfo[d.pc-1],
			err)
	}

	return fmt.Errorf(
		"Opcode: %s\n%s: %w",
		op,
		d.debugname,
		err)
}

func (co *Coroutine) Error(err error) {
	co.yield <- yield{nil, errorfmt(err, co.dbg)}
}

func (co *Coroutine) Resume(args ...any) (r Rets, err error) {
	if !co.started {
		// fmt.Println("RM  starting", args)
		co.started = true
		co.status = CoRunning

		go func() {
			// fmt.Println(" RG calling coroutine body with", args)
			r, err := (*co.body)(co, args...)

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
		}()
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

func namecallHandler(co *Coroutine, kv string, stack *[]any, c1, c2 int) (ok bool, retList []any, err error) {
	switch kv {
	case "format":
		str := (*stack)[c1].(string)
		args := (*stack)[c1+1 : c2+1]

		f, err := fmtstring(str, &Args{Co: co, List: args, name: "format"})
		if err != nil {
			return false, nil, err
		}
		return true, []any{f}, nil
	}
	return
}

var exts = map[any]any{
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
	"type": MakeFn("type", global_type)[1],
	// "typeof":   MakeFn("typeof", global_type)[1], // same because no metatables
	"ipairs":   MakeFn("ipairs", global_ipairs)[1],
	"pairs":    MakeFn("pairs", global_pairs)[1],
	"next":     MakeFn("next", global_next)[1],
	"tonumber": MakeFn("tonumber", global_tonumber)[1],
	"tostring": MakeFn("tostring", global_tostring)[1],
	"_VERSION": "Luau", // todo: custom

	"require": MakeFn("require", global_require)[1],
}

// var VectorSize = 4
// var AllowProxyErrors = false

type inst struct {
	K, K0, K1, K2          any
	opname                 string
	A, B, C, D, E, KC, aux int
	value                  uint32
	kmode, opcode, opmode  uint8
	KN, usesAux            bool
}

type proto struct {
	debugname            string
	k                    []any
	code                 []*inst
	instlineinfo, protos []uint32
	debugcode            []uint8

	linedefined, sizecode, sizek, sizep, bytecodeid uint32
	maxstacksize, numparams, nups                   uint8
	isvararg, lineinfoenabled                       bool
}

type deserialised struct {
	mainProto proto
	protoList []proto
}

func checkkmode(i *inst, k []any) {
	switch i.kmode {
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

		id0 := (extend >> 20) & 0x3FF
		i.K0 = k[id0] // maybe can .(string) this
		// fmt.Println("AUX", i.K0)

		if count >= 2 {
			id1 := (extend >> 10) & 0x3FF
			i.K1 = k[id1]
		}
		if count == 3 { // >=?
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
		r |= (v & 0x7F << (i * 7))
		if v&0x80 == 0 {
			break
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

func (s *stream) CheckEnd() {
	if s.pos != uint32(len(s.data)) {
		panic("deserialiser position mismatch")
	}
}

func readInst(codeList *[]*inst, s *stream) (usesAux bool) {
	value := s.rWord()
	opcode := uint8(value & 0xFF)

	opinfo := opList[opcode]
	opmode := opinfo.mode
	usesAux = opinfo.hasAux

	i := &inst{
		opname:  opinfo.name,
		kmode:   opinfo.kMode,
		opcode:  opcode,
		opmode:  opmode,
		usesAux: usesAux,
	}

	*codeList = append(*codeList, i)

	switch opmode {
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

	if usesAux {
		aux := s.rWord()
		i.aux = int(aux)

		*codeList = append(*codeList, &inst{
			opname: "auxvalue",
			value:  aux,
		})
	}
	return
}

func readProto(bytecodeid uint32, stringList []string, s *stream) (proto, error) {
	maxstacksize := s.rByte()
	numparams := s.rByte()
	nups := s.rByte()
	isvararg := s.rByte() != 0

	s.rByte()            // -- flags
	s.pos += s.rVarInt() // typesize

	sizecode := s.rVarInt()
	codelist := new([]*inst)

	var skipnext bool
	for range sizecode {
		if skipnext {
			skipnext = false
			continue
		}
		skipnext = readInst(codelist, s)
	}

	debugcodelist := make([]uint8, sizecode)
	for i := range sizecode {
		debugcodelist[i] = (*codelist)[i].opcode
	}

	sizek := s.rVarInt()
	klist := make([]any, sizek)

	for i := range sizek {
		switch kt := s.rByte(); kt {
		case 0: // Nil
			klist[i] = nil
		case 1: // Bool
			klist[i] = s.rByte() != 0
		case 2: // Number
			klist[i] = s.rFloat64()
		case 3: // String
			klist[i] = stringList[s.rVarInt()-1]
		case 4: // Function
			klist[i] = s.rWord()
		case 5: // Table
			dataLength := s.rVarInt()
			t := make([]uint32, dataLength)

			for j := range dataLength {
				t[j] = s.rVarInt()
			}

			klist[i] = t
		case 6: // Closure
			klist[i] = s.rVarInt()
		case 7: // Vector
			klist[i] = vectorCtor(s.rFloat32(), s.rFloat32(), s.rFloat32(), s.rFloat32())
		default:
			return proto{}, fmt.Errorf("Unknown ktype %d", kt)
		}
	}

	// -- 2nd pass to replace constant references in the instruction
	for i := range sizecode {
		checkkmode((*codelist)[i], klist)
	}

	sizep := s.rVarInt()
	ps := make([]uint32, sizep)
	for i := range sizep {
		ps[i] = s.rVarInt() + 1
	}

	linedefined := s.rVarInt()

	var debugname string
	if debugnameindex := s.rVarInt(); debugnameindex == 0 {
		debugname = "(??)"
	} else {
		debugname = stringList[debugnameindex-1]
	}

	// -- lineinfo
	lineinfoenabled := s.rByte() != 0
	var instlineinfo []uint32

	if lineinfoenabled {
		linegaplog2 := s.rByte()
		intervals := uint32((sizecode-1)>>linegaplog2) + 1

		lineinfo := make([]uint32, sizecode)
		var lastoffset uint32
		for i := range sizecode {
			lastoffset += uint32(s.rByte())
			lineinfo[i] = lastoffset
		}

		abslineinfo := make([]uint32, intervals)
		var lastline uint64
		for i := range intervals {
			lastline += uint64(s.rWord())
			abslineinfo[i] = uint32(lastline % (1 << 32))
		}

		instlineinfo = make([]uint32, sizecode)
		for i := range sizecode {
			// -- p->abslineinfo[pc >> p->linegaplog2] + p->lineinfo[pc];
			instlineinfo[i] = abslineinfo[i>>linegaplog2] + lineinfo[i]
		}
	}

	// -- debuginfo
	if s.rByte() != 0 {
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

	return proto{
		debugname,
		klist,
		*codelist,
		instlineinfo,
		ps,
		debugcodelist,

		linedefined,
		sizecode,
		sizek,
		sizep,
		bytecodeid,

		maxstacksize,
		numparams,
		nups,

		isvararg,
		lineinfoenabled,
	}, nil
}

func Deserialise(data []byte) (deserialised, error) {
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
	for i := s.rByte(); i != 0; i = s.rByte() {
		s.rVarInt()
	}

	protoCount := s.rVarInt()
	protoList := make([]proto, protoCount)
	for i := range protoCount {
		p, err := readProto(i-1, stringList, s)
		if err != nil {
			return deserialised{}, err
		}
		protoList[i] = p
	}

	mainProto := protoList[s.rVarInt()]
	mainProto.debugname = "(main)"
	s.CheckEnd()

	return deserialised{mainProto, protoList}, nil
}

type iterator struct {
	args, resume chan *[]any
	running      bool
}

type upval struct {
	value   any
	store   *[]any
	index   int
	selfRef bool
}

func (u upval) String() string {
	return fmt.Sprintf("{\n  index: %d\n  store: %v\n  value: %v\n  selfRef: %t\n}", u.index, u.store, u.value, u.selfRef)
}

func truthy(v any) bool {
	return v != nil && v != false
}

var luautype = map[string]string{
	"nil":                 "nil",
	"float64":             "number",
	"string":              "string",
	"bool":                "boolean",
	"*litecode.Table":     "table",
	"litecode.Function":   "function",
	"*litecode.Coroutine": "thread",
	"*litecode.Buffer":    "buffer",
	"litecode.Vector":     "vector",
}

func invalidCompare(op string, ta, tb string) error {
	return fmt.Errorf("attempt to compare %s %s %s", luautype[ta], op, luautype[tb])
}

func incomparableType(t string, eq bool) error {
	return fmt.Errorf("type %s cannot be compared; this comparison would always return %t", luautype[t], eq)
}

func uncallableType(v string) error {
	return fmt.Errorf("attempt to call a %s value", luautype[v])
}

func invalidArithmetic(op string, ta, tb string) error {
	return fmt.Errorf("attempt to perform arithmetic (%s) on %s and %s", op, luautype[ta], luautype[tb])
}

func invalidUnm(t string) error {
	return fmt.Errorf("attempt to perform arithmetic (unm) on %s", luautype[t])
}

func invalidCond(t string) error {
	return fmt.Errorf("attempt to compare non-boolean type %s in condition", luautype[t])
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
	tb := luautype[typeOf(val)]
	if tb == "string" {
		tb = fmt.Sprintf("'%v'", val)
	}

	return fmt.Errorf("attempt to index %v with %v", luautype[ta], tb)
}

func typeOf(v any) string {
	if v == nil { // prevent nil pointer dereference
		return "nil"
	}
	return reflect.TypeOf(v).String()
}

func aAdd(a, b any) (any, error) {
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

	return nil, invalidArithmetic("add", typeOf(a), typeOf(b))
}

func aSub(a, b any) (any, error) {
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

	return nil, invalidArithmetic("sub", typeOf(a), typeOf(b))
}

func aMul(a, b any) (any, error) {
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

	return nil, invalidArithmetic("mul", typeOf(a), typeOf(b))
}

func aDiv(a, b any) (any, error) {
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

	return nil, invalidArithmetic("div", typeOf(a), typeOf(b))
}

func aMod(a, b any) (float64, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa - fb*math.Floor(fa/fb), nil
	}

	return 0, invalidArithmetic("mod", typeOf(a), typeOf(b))
}

func aPow(a, b any) (float64, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return math.Pow(fa, fb), nil
	}

	return 0, invalidArithmetic("pow", typeOf(a), typeOf(b))
}

func aIdiv(a, b any) (any, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return math.Floor(fa / fb), nil
	}

	va, ok3 := a.(Vector)
	vb, ok4 := b.(Vector)
	if ok3 && ok4 {
		return Vector{
			fFloor(va[0] / vb[0]),
			fFloor(va[1] / vb[1]),
			fFloor(va[2] / vb[2]),
			fFloor(va[3] / vb[3]),
		}, nil
	} else if ok1 && ok4 {
		f := float32(fa)
		return Vector{
			fFloor(f / vb[0]),
			fFloor(f / vb[1]),
			fFloor(f / vb[2]),
			fFloor(f / vb[3]),
		}, nil
	} else if ok3 && ok2 {
		f := float32(fb)
		return Vector{
			fFloor(va[0] / f),
			fFloor(va[1] / f),
			fFloor(va[2] / f),
			fFloor(va[3] / f),
		}, nil
	}

	return nil, invalidArithmetic("idiv", typeOf(a), typeOf(b))
}

// vectors dont have these comparisons
func jumpLe(a, b any) (bool, error) {
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

	return false, invalidCompare("<=", typeOf(a), typeOf(b))
}

func jumpLt(a, b any) (bool, error) {
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

	return false, invalidCompare("<", typeOf(a), typeOf(b))
}

func jumpGt(a, b any) (bool, error) {
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

	return false, invalidCompare(">", typeOf(a), typeOf(b))
}

func jumpGe(a, b any) (bool, error) {
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

	return false, invalidCompare(">=", typeOf(a), typeOf(b))
}

func gettable(index, v any) (any, error) {
	switch t := v.(type) {
	case *Table:
		return t.Get(index), nil
	case Vector: // direction,,, and mmmagnitude!! oh yeah!!11!!
		si, ok := index.(string)
		if !ok {
			return nil, invalidIndex("litecode.Vector", index)
		}

		switch si {
		case "x":
			return t[0], nil
		case "y":
			return t[1], nil
		case "z":
			return t[2], nil
			// case "w":
			// 	(*stack)[i.A] = t[3]
		}
		return nil, invalidIndex("litecode.Vector", si)
	}
	return nil, invalidIndex(typeOf(v), index)
}

type toWrap struct {
	proto        proto
	upvals       []*upval
	alive        *bool
	protolist    []proto
	env          map[any]any
	requireCache map[string]Rets
}

type debugging struct {
	pc, top      int
	enabled      bool
	opcode       uint8
	debugname    string
	instlineinfo []uint32
}

func execute(towrap toWrap, dbg *debugging, stack *[]any, co *Coroutine, vargsList []any, vargsLen uint8) (r Rets, err error) {
	p, upvals, alive, protolist, env, requireCache := towrap.proto, towrap.upvals, towrap.alive, towrap.protolist, towrap.env, towrap.requireCache
	ps, code := p.protos, p.code
	pc, top, openUpvals, generalisedIterators := 1, -1, []*upval{}, map[inst]*iterator{}

	moveStack := func(src []any, b, t int) {
		for t+b >= len(*stack) {
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
	for *alive {
		if !handlingBreak {
			i = *code[pc-1]
			op = i.opcode
		}
		handlingBreak = false

		dbg.pc = pc
		dbg.top = top
		dbg.enabled = p.lineinfoenabled
		dbg.opcode = i.opcode
		dbg.debugname = p.debugname
		dbg.instlineinfo = p.instlineinfo

		// if len(upvals) > 0 {
		// 	fmt.Println("upval", upvals[0])
		// }
		// fmt.Printf("OP %-2d PC %-3d UV %d\n", op, pc+1, len(upvals))

		switch op {
		case 0: // NOP
			pc++
			// -- Do nothing
		case 1: // BREAK
			op = p.debugcode[pc]
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

			if exts[kv] != nil {
				(*stack)[i.A] = exts[kv]
			} else {
				(*stack)[i.A] = env[kv]
			}

			pc += 2 // -- adjust for aux
		case 8: // SETGLOBAL
			// LOL
			kv := i.K
			if _, ok := kv.(string); ok {
				if exts[kv] != nil {
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
				// fmt.Println("Setting stacka to", (*uv.store)[uv.index])

				(*stack)[i.A] = (*uv.store)[uv.index]
			}
		case 10: // SETUPVAL
			pc++
			if uv := upvals[i.B]; !uv.selfRef {
				(*uv.store)[uv.index] = (*stack)[i.A]
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
				uv.value = (*uv.store)[uv.index]
				uv.store = nil
				uv.selfRef = true
				openUpvals[n] = nil
				// fmt.Println("closed", uv)
			}
		case 12: // GETIMPORT
			k0 := i.K0
			imp := exts[k0]
			if imp == nil {
				imp = env[k0]
			}

			// fmt.Println("IMPORTING", k0)

			switch i.KC { // count
			case 1:
				// fmt.Println("GETIMPORT1", i.A, imp)
				(*stack)[i.A] = imp
			case 2:
				t := imp.(*Table)
				// fmt.Println("GETIMPORT2", i.A, t.Get(i.K1))
				(*stack)[i.A] = t.Get(i.K1)
			case 3:
				t := imp.(*Table)
				// fmt.Println("GETIMPORT3", i.A, t.Get(i.K1).([]any)[i.K2.(uint32)-1])
				(*stack)[i.A] = t.Get(i.K1).([]any)[i.K2.(uint32)-1]
			}

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
				return nil, invalidIndex(typeOf((*stack)[i.B]), index)
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
				return nil, invalidIndex(typeOf((*stack)[i.B]), index)
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
			} else if i, v := int(i.C+1), (*stack)[i.A]; 1 <= i || i > len(*t.array) {
				t.SetArray(i, v)
			} else {
				t.SetHash(float64(i), v)
			}

			pc++
		case 19: // NEWCLOSURE
			newProto := protolist[ps[i.D]-1]

			nups := newProto.nups
			uvs := make([]*upval, nups)

			// wrap is reused for closures
			towrap.proto = newProto
			towrap.upvals = uvs

			(*stack)[i.A] = wrapclosure(towrap)
			// fmt.Println("WRAPPING WITH", uvs)

			// fmt.Println("nups", nups)
			for n := range nups {
				switch pseudo := code[pc]; pseudo.A {
				case 0: // -- value
					uv := &upval{
						value:   (*stack)[pseudo.B],
						selfRef: true,
					}
					uv.store = nil

					uvs[n] = uv
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
							store: stack,
							index: index,
						}

						for len(openUpvals) <= index {
							openUpvals = append(openUpvals, nil)
						}
						openUpvals[index] = prev
					}

					uvs[n] = prev
					// fmt.Println("set upvalue", i, "to", prev)
				case 2: // -- upvalue
					// fmt.Println("moving", i, pseudo.B)
					uvs[n] = upvals[pseudo.B]
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
			callInst := code[pc]
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

				if t.node == nil {
					(*stack)[A] = nil
				} else {
					(*stack)[A] = t.GetHash(kv)
				}
				break
			}

			pc += 2 // -- adjust for aux, Skip next CALL instruction

			i = *callInst
			op = callOp
			dbg.pc = pc
			dbg.opcode = i.opcode

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
			// fmt.Println("calling with", (*stack)[A+1:A+params+1])
			if !ok {
				return nil, uncallableType(typeOf(f))
			}

			// fmt.Println("upvals1", len(upvals))
			retList, err := (*fn)(co, (*stack)[A+1:A+params+1]...) // not inclusive
			// fmt.Println("upvals2", len(upvals))
			if err != nil {
				return nil, err
			}
			// fmt.Println("resultt", retList)
			retCount := len(retList)

			// fmt.Println("COUNT", retCount)
			if retCount == 1 {
				if p, ok := retList[0].(loadParams); ok {
					// it's a require
					if c, ok := requireCache[p.path]; ok {
						retList = c
					} else {
						c2, _ := Load(p.deserialised, p.path, p.o, p.env, requireCache)
						result, err := c2.Resume()
						if err != nil {
							return nil, err
						}

						requireCache[p.path] = result
						retList = result
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
			pc++
			j, err := aAdd((*stack)[i.B], (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 34:
			pc++
			j, err := aSub((*stack)[i.B], (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 35:
			pc++
			j, err := aMul((*stack)[i.B], (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 36:
			pc++
			j, err := aDiv((*stack)[i.B], (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 37:
			pc++
			j, err := aMod((*stack)[i.B], (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 38:
			pc++
			j, err := aPow((*stack)[i.B], (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 81:
			pc++
			j, err := aIdiv((*stack)[i.B], (*stack)[i.C])
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 39: // arithmetik
			pc++
			j, err := aAdd((*stack)[i.B], i.K)
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 40:
			pc++
			j, err := aSub((*stack)[i.B], i.K)
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 41:
			pc++
			j, err := aMul((*stack)[i.B], i.K)
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 42:
			pc++
			j, err := aDiv((*stack)[i.B], i.K)
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 43:
			pc++
			j, err := aMod((*stack)[i.B], i.K)
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 44:
			pc++
			j, err := aPow((*stack)[i.B], i.K)
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j
		case 82:
			pc++
			j, err := aIdiv((*stack)[i.B], i.K)
			if err != nil {
				return nil, err
			}
			(*stack)[i.A] = j

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
					return nil, invalidConcat(typeOf((*stack)[n+first]), typeOf((*stack)[n+1+first]))
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
				return nil, invalidUnm(typeOf((*stack)[i.B]))
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
				return nil, invalidLength(typeOf(t))
			}
		case 53: // NEWTABLE
			(*stack)[i.A] = &Table{}

			pc += 2 // -- adjust for aux
		case 54: // DUPTABLE
			pc++
			serialised := &Table{}
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
				ui := int(n + i.aux)
				if 1 <= ui || ui > len(*s.array) {
					s.SetArray(ui, v)
					continue
				}
				s.SetHash(float64(ui), v)
			}
			// (*stack)[A] = s

			pc += 2 // -- adjust for aux
		case 56: // FORNPREP
			pc++
			A := i.A

			index, ok := (*stack)[A+2].(float64)
			if !ok {
				return nil, invalidFor("initial value", typeOf((*stack)[A+2]))
			}

			limit, ok := (*stack)[A].(float64)
			if !ok {
				return nil, invalidFor("limit", typeOf((*stack)[A]))
			}

			step, ok := (*stack)[A+1].(float64)
			if !ok {
				return nil, invalidFor("step", typeOf((*stack)[A+1]))
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
			it := (*stack)[A]

			if fn, ok := it.(Function); ok {
				// fmt.Println("IT func", fn, (*stack)[A+1], (*stack)[A+2])
				vals, err := (*fn)(co, (*stack)[A+1], (*stack)[A+2])
				if err != nil {
					return nil, err
				}

				moveStack(vals, res, A+3)
				// fmt.Println(A+3, (*stack)[A+3])

				if (*stack)[A+3] == nil {
					pc += 2
					break
				}

				(*stack)[A+2] = (*stack)[A+3]
				pc += i.D + 1
				break
			}

			iter := *generalisedIterators[i]

			if !iter.running {
				// fmt.Println("-1- sending thru the wire")
				iter.args <- &[]any{it, (*stack)[A+1], (*stack)[A+2]}
				// fmt.Println("-1- sent")
			}

			vals := <-iter.resume
			// fmt.Println("-1- received!", vals)

			if vals == nil {
				delete(generalisedIterators, i)
				pc += 2
				break
			}

			moveStack(*vals, res, A+3)

			(*stack)[A+2] = (*stack)[A+3]
			pc += i.D + 1
		case 59, 61: // FORGPREP_INEXT, FORGPREP_NEXT
			if _, ok := (*stack)[i.A].(Function); !ok {
				return nil, fmt.Errorf("attempt to iterate over a %s value", typeOf((*stack)[i.A])) // -- encountered non-function value
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
			newProto := protolist[i.K.(uint32)]

			nups := newProto.nups
			uvs := make([]*upval, nups)

			towrap.proto = newProto
			towrap.upvals = uvs

			(*stack)[i.A] = wrapclosure(towrap)

			for i := range nups {
				switch pseudo := code[pc]; pseudo.A {
				case 0: // value
					uv := &upval{
						value:   (*stack)[pseudo.B],
						selfRef: true,
					}
					uv.store = nil
					uvs[i] = uv

				// -- references dont get handled by DUPCLOSURE
				case 2: // upvalue
					uvs[i] = upvals[pseudo.B]
				}

				pc++
			}
			pc++
		case 65: // PREPVARARGS
			pc++
			// Handled by wrapper
		case 66: // LOADKX
			(*stack)[i.A] = i.K.(uint32) // kv

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

			loopInstruction := *code[pc-1]
			if generalisedIterators[loopInstruction] != nil {
				break
			}

			c := &iterator{
				args:   make(chan *[]any),
				resume: make(chan *[]any),
			}

			go func() {
				args := *<-c.args
				c.args = nil // we're done here
				c.running = true
				// fmt.Println("-2- generating iterator", args)

				for i, v := range args[0].(*Table).Iter() {
					if !c.running {
						return
					}
					// fmt.Println("-2- yielding", i, v)
					c.resume <- &[]any{i, v}
					// fmt.Println("-2- yielded!")
				}

				c.resume <- nil
			}()

			generalisedIterators[loopInstruction] = c
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
			return nil, fmt.Errorf("Unsupported Opcode: %s op: %d", i.opname, op)
		}
	}

	for i, uv := range openUpvals {
		openUpvals[i] = nil
		if uv.selfRef {
			continue
		}
		uv.value = (*uv.store)[uv.index]
		uv.store = nil
		uv.selfRef = true
	}

	for i := range generalisedIterators {
		generalisedIterators[i].running = false
		delete(generalisedIterators, i)
	}
	return
}

func wrapclosure(towrap toWrap) Function {
	proto := towrap.proto

	return Fn(func(co *Coroutine, args ...any) (r Rets, err error) {
		maxs, np := proto.maxstacksize, proto.numparams // maxs 2 lel

		la := uint8(len(args)) // we can't have more than 255 args anyway right?
		// fmt.Println("MAX STACK SIZE", maxs)
		stack := make([]any, maxs)
		copy(stack, args[:min(np, la)])

		var list []any
		if np < la {
			list = args[np:]
		}

		dbg := &debugging{enabled: proto.lineinfoenabled, opcode: 255}
		co.dbg = dbg

		result, err := execute(towrap, dbg, &stack, co, list, max(la-np, 0))
		if err != nil {
			return nil, errorfmt(err, dbg)
		}

		return result, nil
	})
}

func Load(module deserialised, filepath string, o uint8, env map[any]any, requireCache ...map[string]Rets) (co Coroutine, cancel func()) {
	alive := true

	var cache map[string]Rets
	if len(requireCache) > 0 {
		cache = requireCache[0]
	} else {
		cache = map[string]Rets{}
	}

	towrap := toWrap{
		module.mainProto,
		[]*upval{},
		&alive,
		module.protoList,
		env,
		cache,
	}

	return Coroutine{
		body:     wrapclosure(towrap),
		env:      env,
		filepath: filepath,
		yield:    make(chan yield, 1),
		resume:   make(chan Rets, 1),
		dbg:      &debugging{opcode: 255},
		o:        o,
	}, func() { alive = false }
}
