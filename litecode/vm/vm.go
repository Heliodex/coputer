// package vm implements Litecode's deterministic Luau virtual machine and standard library toolkit.
package vm

import (
	"errors"
	"fmt"
	"iter"
	"math"
	"reflect"
	"slices"
	"strings"
	"unsafe"

	"github.com/Heliodex/coputer/litecode/types"
)

const safe = false

func listKey(k types.Val) (int, bool) {
	fk, ok := k.(float64)
	if !ok {
		return 0, false
	}

	ik := int(fk)
	return ik, 1 <= ik && float64(ik) == fk
}

func mapKeySort(a, b types.Val) int {
	// It doesn't have to be pretty for map keys
	// (in fact, the reference implementation of Luau has a rather insane sort order)
	// It just has to be DETERMINISTIC
	return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
}

func iterList(l []types.Val, y func(types.Val, types.Val) bool) {
	for i, v := range l {
		if v != nil && !y(float64(i+1), v) {
			return
		}
	}
}

func iterHash(m map[types.Val]types.Val, y func(types.Val, types.Val) bool) {
	// order keys in map
	keys := make([]types.Val, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, mapKeySort)
	for _, k := range keys {
		if !y(k, m[k]) {
			return
		}
	}
}

// Q why are tables like this
// A:
// 1: the reference implementation of tables is too complex: rehashing and resizing is a pain but not too bad, list boundaries are worse and I don't want 1.5k lines of code just for that, and Go does a resizing-like thing automatically with slices anyway
// 2: the way nodes are implemented works well in C++ and not in Go (plus I don't know if it's actually O(1) for node lookups??)
// 3: rehashing etc is slower than just using a slice... somehow. most of this program is between 10-20x slower than the reference implementation, but the tables (which were previously like 50x slower) are now only like 2-3x slower for large allocations (bench/largealloc.luau)
// 4: having a list part is actually nice for iteration and for large tables (as opposed to the lua4 way, where it's *just* a hash part), the way it's done here is simpler though we have to move stuff around and between the list and node parts more explicitly
// 5: very weird quirks arise from table length implementations etc. the nil stuff can easily be forgiven, it's the stuff with creating a table and getting a length afterwards (see tests/clear.luau) that is fucking devilish; this is one of the few parts that puts Luau, as the language at the top of my favourites list, in jeopardy
// 6: we don't actually break *that* much compatibility doing it this way, right??
// 7: if anyone tells you tables are simple THEY ARE LYING, CALL THEM OUT ON THEIR SHIT

// Table represents a Luau table, with resizeable list and hash parts. Luau type `table`
type Table struct {
	List     []types.Val
	Hash     map[types.Val]types.Val
	Readonly bool
}

// Len returns the length of the list part of the table (the length of the list up until the first nil).
func (t *Table) Len() int {
	if t.List == nil {
		return 0
	}
	return len(t.List)
}

// setHash updates or deletes a key-value pair in the hash part of the table.
func (t *Table) setHash(k types.Val, v types.Val) {
	if t.Hash == nil {
		if v == nil {
			return
		}
		t.Hash = map[types.Val]types.Val{k: v}
	} else if v == nil {
		delete(t.Hash, k)
	} else {
		t.Hash[k] = v
	}
}

// check if we can move some stuff from the hash part to the list part
func (t *Table) moveToList(l int) {
	if t.Hash == nil {
		return
	}

	for f2 := float64(l + 2); ; f2++ {
		v2, ok := t.Hash[f2]
		if !ok {
			break
		}
		t.List = append(t.List, v2)
		delete(t.Hash, f2)
	}
}

// SetInt sets a value at an integer index, placing it into the List part or the Hash part and resizing each as appropriate.
func (t *Table) SetInt(i int, v types.Val) {
	// fmt.Println("SetInt", i, v)

	if t.List == nil {
		if i == 1 {
			t.List = []types.Val{v}

			t.moveToList(0)
			return
		}
	} else if l := len(t.List); i < l+1 {
		if v != nil {
			// set in the list portion
			t.List[i-1] = v
			return
		}

		// cut the list portion
		after := t.List[i:]
		t.List = t.List[:i-1]

		// move the rest to the hash part
		for i2, v2 := range after {
			t.setHash(float64(i+i2), v2)
		}
		return
	} else if i == l+1 {
		// append to the end
		t.List = append(t.List, v)

		t.moveToList(l)
		return
	}

	// add to the hash part instead
	t.setHash(float64(i), v)
}

// Set sets a table value at a key. Make sure to check if the table is readonly beforehand.
func (t *Table) Set(k types.Val, v types.Val) {
	if ak, ok := listKey(k); ok {
		t.SetInt(ak, v)
		return
	}
	t.setHash(k, v)
}

// GetHash returns a value at a key, only searching the hash part of the table.
func (t *Table) GetHash(k types.Val) (v types.Val) {
	if t.Hash == nil {
		return
	}
	return t.Hash[k]
}

// Get returns a value at a key in the table.
func (t *Table) Get(k types.Val) types.Val {
	if ak, ok := listKey(k); ok && ak <= t.Len() {
		return t.List[ak-1]
	}
	return t.GetHash(k)
}

// Iter returns an iterator over the table, yielding key-value pairs in a deterministic order.
func (t *Table) Iter() iter.Seq2[types.Val, types.Val] {
	return func(y func(types.Val, types.Val) bool) {
		if t.List != nil {
			iterList(t.List, y)
		}
		if t.Hash != nil {
			iterHash(t.Hash, y)
		}
	}
}

var errReadonly = errors.New("attempt to modify a readonly table")

// bit32 extraction
func extract(n, field, width uint32) uint32 {
	return n >> field & (1<<width - 1)
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

var opList = [83]types.OpInfo{
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

func fn(name string, f func(co *types.Coroutine, args ...types.Val) (r []types.Val, err error)) types.Function {
	return types.Function{
		Run:  &f,
		Name: name,
	}
}

func createCoroutine(body types.Function, currentCo *types.Coroutine) *types.Coroutine {
	// first time i actually ran into the channel axiom issues
	return &types.Coroutine{
		Body:       body,
		Filepath:   currentCo.Filepath,
		Dbgpath:    currentCo.Dbgpath,
		YieldChan:  make(chan types.Yield, 1),
		ResumeChan: make(chan []types.Val, 1),
	}
}

const luau_multret = -1

func namecallHandler(co *types.Coroutine, kv string, stack *[]types.Val, c1, c2 int32) (ok bool, retList []types.Val, err error) {
	switch kv {
	case "format":
		str := (*stack)[c1].(string)
		args := (*stack)[c1+1 : c2+1]

		f, err := fmtstring(str, Args{Co: co, List: args, name: "format"})
		if err != nil {
			return false, nil, err
		}
		return true, []types.Val{f}, nil
	}
	return
}

var exts = types.Env{
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

func checkkmode(i *types.Inst, k []types.Val) {
	switch i.KMode {
	case 1: // AUX
		if i.Aux < uint32(len(k)) { // sometimes huge for some reason
			i.K = k[i.Aux]
		}
	case 2: // C
		i.K = k[i.C]
		// fmt.Println("SET K TO", i.K, "FROM", i.C)
	case 3: // D
		i.K = k[i.D]
	case 4: // AUX import
		extend := i.Aux
		count := uint8(extend >> 30)
		i.KC = count

		id0 := extend >> 20 & 0x3ff
		i.K0 = k[id0] // maybe can .(string) this
		// fmt.Println("AUX", i.K0)

		if count >= 2 {
			id1 := extend >> 10 & 0x3ff
			i.K1 = k[id1]
		}
		if count == 3 { // should never be 3
			id2 := extend & 0x3ff
			i.K2 = k[id2]
		}
	case 5: // AUX boolean low 1 bit
		i.K = extract(i.Aux, 0, 1) == 1
		i.KN = extract(i.Aux, 31, 1) == 1
	case 6: // AUX number low 24 bits
		i.K = k[extract(i.Aux, 0, 24)]
		i.KN = extract(i.Aux, 31, 1) == 1
	case 7: // B
		i.K = k[i.B]
	case 8: // AUX number low 16 bits
		i.K = i.Aux & 0xf
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
	if safe {
		w = uint32(s.data[s.pos]) |
			uint32(s.data[s.pos+1])<<8 |
			uint32(s.data[s.pos+2])<<16 |
			uint32(s.data[s.pos+3])<<24
	} else {
		w = *(*uint32)(unsafe.Pointer(&s.data[s.pos]))
	}
	s.pos += 4
	return
}

// this is the only thing float32s are ever used for anyway
func (s *stream) rVector() (r types.Vector) {
	if safe {
		for i := range 4 {
			r[i] = math.Float32frombits(s.rWord())
		}
		return
	}

	r = *(*types.Vector)(unsafe.Pointer(&s.data[s.pos]))
	s.pos += 16
	return
}

func (s *stream) rFloat64() (r float64) {
	if safe {
		return math.Float64frombits(uint64(s.rWord()) | uint64(s.rWord())<<32)
	}

	r = *(*float64)(unsafe.Pointer(&s.data[s.pos]))
	s.pos += 8
	return
}

func (s *stream) skipVarInt() {
	for range 4 {
		if s.rByte()&0b1000_0000 == 0 {
			return
		}
	}
}

func (s *stream) rVarInt() (r uint32) {
	for i := range 4 {
		v := uint32(s.rByte())
		r |= v & 0b0111_1111 << (i * 7)
		if v&0b1000_0000 == 0 {
			return
		}
	}
	return
}

func (s *stream) rString() (str string) {
	size := s.rVarInt()
	if size == 0 {
		return ""
	}

	str = string(s.data[s.pos:][:size])
	s.pos += size
	return
}

func (s *stream) CheckEnd() error {
	if s.pos != uint32(len(s.data)) {
		return errors.New("deserialiser position mismatch")
	}
	return nil
}

// reads either 1 or 2 words
func readInst(codeList *[]*types.Inst, s *stream) bool {
	value := s.rWord()

	opcode := uint8(value)
	opinfo := opList[opcode]

	i := types.Inst{
		OpInfo: opinfo,
		Opcode:       opcode,
	}

	value >>= 8
	switch opinfo.Mode {
	case 1: // A
		i.A = uint8(value) // 8 bit
	case 2: // AB
		i.A, i.B = uint8(value), uint8(value>>8)
	case 3: // ABC
		i.A, i.B, i.C = uint8(value), uint8(value>>8), uint8(value>>16)
	case 4: // AD
		i.A = uint8(value)
		i.D = int32(int16(value >> 8))
	case 5: // AE
		if i.E = int32(value); i.E >= 0x800000 { // why no arbitrary width integers, go
			i.E -= 0x1000000
		}
	}

	*codeList = append(*codeList, &i)
	if opinfo.HasAux {
		i.Aux = s.rWord()

		*codeList = append(*codeList, &types.Inst{})
	}

	return opinfo.HasAux
}

func readProto(stringList []string, s *stream) (p *types.Proto, err error) {
	p = &types.Proto{
		MaxStackSize: s.rByte(),
		NumParams:    s.rByte(),
		Nups:         s.rByte(),
	}

	// s.rBool()            // isvararg
	// s.rByte()            // -- flags
	s.pos += 2
	s.pos += s.rVarInt() // typesize

	sizecode := s.rVarInt()

	for i := uint32(0); i < sizecode; i++ {
		if readInst(&p.Code, s) {
			i++
		}
	}

	p.Dbgcode = make([]uint8, sizecode)
	for i := range sizecode {
		p.Dbgcode[i] = p.Code[i].Opcode
	}

	sizek := s.rVarInt()
	p.K = make([]types.Val, sizek) // crazy

	for i := range sizek {
		switch kt := s.rByte(); kt {
		case 0: // Nil
			p.K[i] = nil
		case 1: // Bool
			p.K[i] = s.rBool()
		case 2: // Number
			p.K[i] = s.rFloat64()
		case 3: // String
			p.K[i] = stringList[s.rVarInt()-1]
		case 4: // Import
			p.K[i] = s.rWord() // ⚠️ strange, TODO need something to test this ⚠️
		case 5: // Table
			dataLength := s.rVarInt()
			t := make([]uint32, dataLength)

			for j := range dataLength {
				t[j] = s.rVarInt() // whatever
			}

			p.K[i] = t // ⚠️ not a val ⚠️
		case 6: // Closure
			p.K[i] = s.rVarInt() // ⚠️ not a val ⚠️
		case 7: // types.Vector
			p.K[i] = s.rVector()
		default:
			return nil, fmt.Errorf("unknown ktype %d", kt)
		}
	}

	// -- 2nd pass to replace constant references in the instruction
	for i := range sizecode {
		checkkmode(p.Code[i], p.K)
	}

	sizep := s.rVarInt()
	p.Protos = make([]uint32, sizep)
	for i := range sizep {
		p.Protos[i] = s.rVarInt() + 1
	}

	// p.linedefined = s.rVarInt()
	s.skipVarInt()

	if dbgnamei := s.rVarInt(); dbgnamei == 0 {
		p.Dbgname = "(??)"
	} else {
		p.Dbgname = stringList[dbgnamei-1]
	}

	// -- lineinfo
	if p.LineInfoEnabled = s.rBool(); p.LineInfoEnabled {
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

		p.InstLineInfo = make([]uint32, sizecode)
		for i := range sizecode {
			// -- p->abslineinfo[pc >> p->linegaplog2] + p->lineinfo[pc];
			p.InstLineInfo[i] = abslineinfo[i>>linegaplog2] + uint32(lineinfo[i])
		}
	}

	// -- debuginfo
	if s.rBool() {
		// fmt.Println("DEBUGINFO")
		for range s.rVarInt() { // sizel
			s.skipVarInt()
			s.skipVarInt()
			s.skipVarInt()
			s.pos++
		}
		for range s.rVarInt() { // sizeupvalues
			s.skipVarInt()
		}
	}

	return
}

func deserialise(b []byte) (des types.Deserialised, err error) {
	s := &stream{data: b}

	if luauVersion := s.rByte(); luauVersion == 0 {
		return types.Deserialised{}, errors.New("the provided bytecode is an error message")
	} else if luauVersion != 6 {
		return types.Deserialised{}, errors.New("the version of the provided bytecode is unsupported")
	}
	if s.rByte() != 3 { // types version
		return types.Deserialised{}, errors.New("the types version of the provided bytecode is unsupported")
	}

	stringCount := s.rVarInt()
	stringList := make([]string, stringCount)
	for i := range stringCount {
		stringList[i] = s.rString()
	}

	// userdataRemapping (not used in VM, left unused)
	for s.rBool() {
		s.rVarInt()
	}

	protoCount := s.rVarInt()
	protoList := make([]*types.Proto, protoCount)
	for i := range protoCount {
		protoList[i], err = readProto(stringList, s)
		if err != nil {
			return
		}
	}

	mainProto := protoList[s.rVarInt()]
	mainProto.Dbgname = "(main)"

	return types.Deserialised{mainProto, protoList}, s.CheckEnd()
}

type iterator struct {
	args    chan *Table
	resume  chan *[]types.Val
	running bool
}

type upval struct {
	value   types.Val
	store   []types.Val
	index   uint8
	selfRef bool
}

func truthy(v types.Val) bool {
	return v != nil && v != false
}

var luautype = map[string]string{
	"nil":       "nil",
	"float64":   "number",
	"string":    "string",
	"bool":      "boolean",
	"*vm.Table": "table",
	"types.Function": "function",
	"*types.Coroutine": "thread",
	"*types.Buffer": "buffer",
	"types.Vector":  "vector",
}

func invalidCompare(op, ta, tb string) error {
	return fmt.Errorf("attempt to compare %s %s %s", ta, op, tb)
}

func uncallableType(v string) error {
	return fmt.Errorf("attempt to call a %s value", v)
}

func invalidArithmetic(op, ta, tb string) error {
	return fmt.Errorf("attempt to perform arithmetic (%s) on %s and %s", op, ta, tb)
}

func invalidUnm(t string) error {
	return fmt.Errorf("attempt to perform arithmetic (unm) on %s", t)
}

func invalidFor(pos, t string) error {
	return fmt.Errorf("invalid 'for' %s (number expected, got %s)", pos, t)
}

func invalidLength(t string) error {
	return fmt.Errorf("attempt to get length of a %s value", t)
}

func invalidConcat(t1, t2 string) error {
	return fmt.Errorf("attempt to concatenate %s with %s", t1, t2)
}

func invalidIndex(ta string, v types.Val) error {
	tb := TypeOf(v)
	if tb == "string" {
		tb = fmt.Sprintf("'%v'", v)
	}

	return fmt.Errorf("attempt to index %v with %v", ta, tb)
}

func missingMethod(ta string, v types.Val) error {
	tb := TypeOf(v)
	if tb == "string" {
		tb = fmt.Sprintf("'%v'", v)
	}

	return fmt.Errorf("attempt to call missing method %v of %v", tb, ta)
}

// TypeOf returns the underlying VM datatype of a value as a string.
// This does not return the Luau type, as type() does.
func TypeOf(v types.Val) (t string) {
	if v == nil { // prevent nil pointer dereference
		return "nil"
	}

	t, ok := luautype[reflect.TypeOf(v).String()]
	if !ok {
		return reflect.TypeOf(v).String()
	}
	return
}

func aAdd(a, b types.Val) (nt types.Val, err error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa + fb, nil
	}

	va, ok3 := a.(types.Vector)
	vb, ok4 := b.(types.Vector)
	if ok3 && ok4 {
		return types.Vector{va[0] + vb[0], va[1] + vb[1], va[2] + vb[2], va[3] + vb[3]}, nil
	}

	return nt, invalidArithmetic("add", TypeOf(a), TypeOf(b))
}

func aSub(a, b types.Val) (nt types.Val, err error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa - fb, nil
	}

	va, ok3 := a.(types.Vector)
	vb, ok4 := b.(types.Vector)
	if ok3 && ok4 {
		return types.Vector{va[0] - vb[0], va[1] - vb[1], va[2] - vb[2], va[3] - vb[3]}, nil
	}

	return nt, invalidArithmetic("sub", TypeOf(a), TypeOf(b))
}

func aMul(a, b types.Val) (nt types.Val, err error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa * fb, nil
	}

	va, ok3 := a.(types.Vector)
	vb, ok4 := b.(types.Vector)
	switch {
	case ok3 && ok4:
		return types.Vector{va[0] * vb[0], va[1] * vb[1], va[2] * vb[2], va[3] * vb[3]}, nil
	case ok1 && ok4:
		f := float32(fa)
		return types.Vector{f * vb[0], f * vb[1], f * vb[2], f * vb[3]}, nil
	case ok3 && ok2:
		f := float32(fb)
		return types.Vector{va[0] * f, va[1] * f, va[2] * f, va[3] * f}, nil
	}

	return nt, invalidArithmetic("mul", TypeOf(a), TypeOf(b))
}

func aDiv(a, b types.Val) (nt types.Val, err error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa / fb, nil
	}

	va, ok3 := a.(types.Vector)
	vb, ok4 := b.(types.Vector)
	switch {
	case ok3 && ok4:
		return types.Vector{va[0] / vb[0], va[1] / vb[1], va[2] / vb[2], va[3] / vb[3]}, nil
	case ok1 && ok4:
		f := float32(fa)
		return types.Vector{f / vb[0], f / vb[1], f / vb[2], f / vb[3]}, nil
	case ok3 && ok2:
		f := float32(fb)
		return types.Vector{va[0] / f, va[1] / f, va[2] / f, va[3] / f}, nil
	}

	return nt, invalidArithmetic("div", TypeOf(a), TypeOf(b))
}

func aMod(a, b types.Val) (nt types.Val, err error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa - fb*math.Floor(fa/fb), nil
	}

	return nt, invalidArithmetic("mod", TypeOf(a), TypeOf(b))
}

func aPow(a, b types.Val) (nt types.Val, err error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return math.Pow(fa, fb), nil
	}

	return nt, invalidArithmetic("pow", TypeOf(a), TypeOf(b))
}

func aIdiv(a, b types.Val) (types.Val, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return math.Floor(fa / fb), nil
	}

	va, ok3 := a.(types.Vector)
	vb, ok4 := b.(types.Vector)
	switch {
	case ok3 && ok4:
		return types.Vector{
			f32Floor(va[0] / vb[0]),
			f32Floor(va[1] / vb[1]),
			f32Floor(va[2] / vb[2]),
			f32Floor(va[3] / vb[3]),
		}, nil
	case ok1 && ok4:
		f := float32(fa)
		return types.Vector{
			f32Floor(f / vb[0]),
			f32Floor(f / vb[1]),
			f32Floor(f / vb[2]),
			f32Floor(f / vb[3]),
		}, nil
	case ok3 && ok2:
		f := float32(fb)
		return types.Vector{
			f32Floor(va[0] / f),
			f32Floor(va[1] / f),
			f32Floor(va[2] / f),
			f32Floor(va[3] / f),
		}, nil
	}

	return nil, invalidArithmetic("idiv", TypeOf(a), TypeOf(b))
}

// vectors dont have these comparisons
func jumpLe(a, b types.Val) (bool, error) {
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

func jumpLt(a, b types.Val) (bool, error) {
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

func jumpGt(a, b types.Val) (bool, error) {
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

func jumpGe(a, b types.Val) (bool, error) {
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

func gettable(index, v types.Val) (types.Val, error) {
	switch t := v.(type) {
	case *Table:
		return t.Get(index), nil
	case types.Vector: // direction,,, and mmmagnitude!! oh yeah!!11!!
		switch index {
		case "x", "X": // yes, we'll allow the capitalised versions anyway
			return float64(t[0]), nil // alright, who's the wise guy whose idea it was to put a float32 on the stack
		case "y", "Y":
			return float64(t[1]), nil
		case "z", "Z":
			return float64(t[2]), nil
			// case "w", "W":
			// 	return float64(t[3]), nil
		}
	}
	return nil, invalidIndex(TypeOf(v), index)
}

type toWrap struct {
	proto     *types.Proto
	protoList []*types.Proto
	upvals    []*upval
	alive     *bool
	env       types.Env
	// Store the last return, as it's the only one that's relevant
	requireCache map[string]types.Val
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
		c.resume <- &[]types.Val{i, v}
		// fmt.Println("-2- yielded!")
	}

	c.resume <- nil
}

func moveStack(stack *[]types.Val, src []types.Val, b, t int32) {
	for t+b >= int32(len(*stack)) { // graah stack expansion
		*stack = append(*stack, nil)
	}

	for i := range b {
		if i >= int32(len(src)) {
			(*stack)[t+i] = nil
			continue
		}
		(*stack)[t+i] = src[i]
	}
}

func getImport(i types.Inst, towrap toWrap, stack *[]types.Val) error {
	k0 := i.K0
	imp := exts[k0]
	if imp == nil {
		imp = towrap.env[k0]
	}

	if count := i.KC; count >= 2 {
		t1, ok := imp.(*Table)
		if !ok {
			return invalidIndex("nil", i.K1)
		}

		imp = t1.Get(i.K1)
		// fmt.Println("GETIMPORT2", i.A, (*stack)[i.A])

		if count == 3 {
			t2, ok := imp.(*Table)
			if !ok {
				return invalidIndex(TypeOf(imp), i.K2)
			}

			imp = t2.Get(i.K2)
			// fmt.Println("GETIMPORT3", i.A, (*stack)[i.A])
		}
	}

	(*stack)[i.A] = imp
	return nil
}

func newClosure(pc *int32, i types.Inst, towrap toWrap, p *types.Proto, stack *[]types.Val, openUpvals *[]*upval, upvals []*upval) {
	newProto := towrap.protoList[p.Protos[i.D]-1]

	nups := newProto.Nups
	towrap.upvals = make([]*upval, nups)

	// wrap is reused for closures
	towrap.proto = newProto

	(*stack)[i.A] = wrapclosure(towrap)
	// fmt.Println("WRAPPING WITH", uvs)

	// fmt.Println("nups", nups)
	for n := range nups {
		switch pseudo := p.Code[*pc]; pseudo.A {
		case 0: // -- value
			towrap.upvals[n] = &upval{
				value:   (*stack)[pseudo.B],
				selfRef: true,
			}
		case 1: // -- reference
			index := pseudo.B
			// fmt.Println("index", index, len(openUpvals))
			// for si, sv := range *stack {
			// 	fmt.Printf("  [%d] = %v\n", si, sv)
			// }

			var prev *upval
			if index < uint8(len(*openUpvals)) {
				prev = (*openUpvals)[index]
			}

			if prev == nil {
				prev = &upval{
					store: *stack,
					index: index,
				}

				for index >= uint8(len(*openUpvals)) {
					*openUpvals = append(*openUpvals, nil)
				}
				(*openUpvals)[index] = prev
			}

			towrap.upvals[n] = prev
			// fmt.Println("set upvalue", i, "to", prev)
		case 2: // -- upvalue
			// fmt.Println("moving", i, pseudo.B)
			towrap.upvals[n] = upvals[pseudo.B]
		}
		*pc++
	}
}

func namecall(pc, top *int32, i *types.Inst, p *types.Proto, stack *[]types.Val, co *types.Coroutine, op *uint8) (err error) {
	A, B := i.A, i.B
	kv := i.K.(string)
	// fmt.Println("kv", kv)

	(*stack)[A+1] = (*stack)[B]

	// -- Special handling for native namecall behaviour
	callInst := p.Code[*pc]
	callOp := callInst.Opcode

	// -- Copied from the CALL handler
	callA, callB, callC := int32(callInst.A), int32(callInst.B), callInst.C

	var params int32
	if callB == 0 {
		params = *top - callA
	} else {
		params = callB - 1
	}

	ok, retList, err := namecallHandler(co, kv, stack, callA+1, callA+params)
	if err != nil {
		return
	}
	if !ok {
		// fmt.Println("namecall", kv, "not found")
		switch t := (*stack)[B].(type) {
		case *Table:
			var call types.Val
			if t.Hash == nil {
				call = nil
			} else {
				call = t.GetHash(kv)
			}

			if call == nil {
				return missingMethod(TypeOf(t), kv)
			}

			(*stack)[A] = call
			return
		case string:
			return missingMethod(TypeOf(t), kv)
		default:
			return invalidIndex(TypeOf(t), kv)
		}
	}

	*i = *callInst
	*op = callOp

	co.Dbg.Line = p.InstLineInfo[*pc+1]

	retCount := int32(len(retList))

	if callC == 0 {
		*top = callA + retCount - 1
	} else {
		retCount = int32(callC - 1)
	}

	moveStack(stack, retList, retCount, callA)
	*pc += 2 // -- adjust for aux, Skip next CALL instruction
	return
}

func handleRequire(towrap toWrap, lc compiled, co *types.Coroutine) ([]types.Val, error) {
	if c, ok := towrap.requireCache[lc.Filepath]; ok {
		return []types.Val{c}, nil
	}

	// since environments only store global libraries etc, using the same env here should be fine??
	c2, _ := loadmodule(lc, co.Env, towrap.requireCache, co.ProgramArgs)
	reqrets, err := c2.Resume()
	if err != nil {
		return nil, err
	}
	if len(reqrets) == 0 { // i have no reqrets
		return nil, errors.New("module must return a value")
	}

	// only the last return value (weird luau behaviour...)
	ret := reqrets[len(reqrets)-1]
	switch ret.(type) {
	case *Table, types.Function:
	default:
		return nil, errors.New("module must return a table or function")
	}

	towrap.requireCache[lc.Filepath] = ret
	return []types.Val{ret}, nil
}

func call(top *int32, i types.Inst, towrap toWrap, stack *[]types.Val, co *types.Coroutine) (err error) {
	A, B, C := int32(i.A), int32(i.B), i.C

	var params int32
	if B == 0 {
		params = *top - A
	} else {
		params = B - 1
	}

	// fmt.Println(A, B, C, (*stack)[A], params)

	f := (*stack)[A]
	fn, ok := f.(types.Function)
	// fmt.Println("calling with", (*stack)[A+1:][:params])
	if !ok {
		return uncallableType(TypeOf(f))
	}

	// fmt.Println("upvals1", len(upvals))
	retList, err := (*fn.Run)(co, (*stack)[A+1:][:params]...) // not inclusive
	// fmt.Println("upvals2", len(upvals))
	if err != nil {
		return
	}
	// fmt.Println("resultt", retList)
	retCount := int32(len(retList))

	// fmt.Println("COUNT", retCount)
	if retCount == 1 { // requires should return only 1 value anyway
		if lc, ok := retList[0].(compiled); ok {
			// it's a require
			// fmt.Println("REQUIRE", lc.filepath)

			if retList, err = handleRequire(towrap, lc, co); err != nil {
				return
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
		*top = A + retCount - 1
	} else {
		retCount = int32(C - 1)
	}

	moveStack(stack, retList, retCount, A)
	return
}

// for gloop lel
func forgloop(pc, top *int32, i types.Inst, stack *[]types.Val, co *types.Coroutine, generalisedIterators *map[types.Inst]*iterator) (err error) {
	A := int32(i.A)
	res := int32(i.K.(uint32))

	*top = A + 6

	switch it := (*stack)[A].(type) {
	case types.Function:
		// fmt.Println("IT func", fn, (*stack)[A+1], (*stack)[A+2])
		vals, err := (*it.Run)(co, (*stack)[A+1], (*stack)[A+2])
		if err != nil {
			return err
		}

		moveStack(stack, vals, res, A+3)
		// fmt.Println(A+3, (*stack)[A+3])

		if (*stack)[A+3] == nil {
			*pc += 2
			return nil
		}

		(*stack)[A+2] = (*stack)[A+3]
	case *Table:
		// fmt.Println("GETTING GENITER", typeOf(it))
		iter := *(*generalisedIterators)[i]

		if !iter.running {
			// fmt.Println("-1- sending thru the wire")
			// fmt.Println((*stack)[A+1], (*stack)[A+2]) // <nil> <nil>
			iter.args <- it
			// fmt.Println("-1- sent")
		}

		vals := <-iter.resume
		// fmt.Println("-1- received!", vals)

		if vals == nil {
			delete((*generalisedIterators), i)
			*pc += 2
			return
		}

		moveStack(stack, *vals, res, A+3)

		(*stack)[A+2] = (*stack)[A+3]
	default:
		return fmt.Errorf("attempt to iterate over a %s value", TypeOf(it))
	}
	*pc += i.D + 1
	return
}

func dupClosure(pc *int32, i types.Inst, towrap toWrap, p *types.Proto, stack *[]types.Val, upvals []*upval) {
	newProto := towrap.protoList[i.K.(uint32)]

	nups := newProto.Nups
	towrap.upvals = make([]*upval, nups)

	// reusing wrapping again bcause we're eco friendly
	towrap.proto = newProto

	(*stack)[i.A] = wrapclosure(towrap)

	for n := range nups {
		switch pseudo := p.Code[*pc]; pseudo.A {
		case 0: // value
			towrap.upvals[n] = &upval{
				value:   (*stack)[pseudo.B],
				selfRef: true,
			}

		// -- references dont get handled by DUPCLOSURE
		case 2: // upvalue
			towrap.upvals[n] = upvals[pseudo.B]
		}

		*pc++
	}
}

func execute(towrap toWrap, stack *[]types.Val, co *types.Coroutine, vargsList []types.Val, vargsLen uint8) (r []types.Val, err error) {
	p, upvals := towrap.proto, towrap.upvals
	// int32 > uint32 lel
	pc, top, openUpvals, generalisedIterators := int32(1), int32(-1), []*upval{}, map[types.Inst]*iterator{}

	var handlingBreak bool
	var i types.Inst
	var op uint8

	// a a a a
	// stayin' alive
	// fmt.Println("starting with upvals", upvals)
	for *towrap.alive {
		if !handlingBreak {
			i = *p.Code[pc-1]
			op = i.Opcode
		}
		handlingBreak = false

		co.Dbg.Line = p.InstLineInfo[pc-1]
		co.Dbg.Name = p.Dbgname

		// fmt.Println(top)

		// if len(upvals) > 0 {
		// 	fmt.Println("upval", upvals[0])
		// }
		// fmt.Printf("OP %-2d PC %-3d UV %d\n", op, pc+1, len(upvals))

		switch op {
		case 0: // NOP
			// -- Do nothing
			pc++
		case 1: // BREAK
			op = p.Dbgcode[pc]
			handlingBreak = true
		case 2: // LOADNIL
			(*stack)[i.A] = nil
			pc++
		case 3: // LOADB
			(*stack)[i.A] = i.B == 1
			pc += int32(i.C) + 1
		case 4: // LOADN
			(*stack)[i.A] = float64(i.D) // never put an int on the stack
			pc++
		case 5: // LOADK
			// fmt.Println("LOADK", i.A, i.K)
			(*stack)[i.A] = i.K
			pc++
		case 6: // MOVE
			// we should (ALMOST) never have to change the size of the stack (p.maxstacksize)
			(*stack)[i.A] = (*stack)[i.B]
			pc++
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
			if uv := upvals[i.B]; uv.selfRef {
				(*stack)[i.A] = uv.value
			} else {
				// fmt.Println("GETTING UPVAL", uv)
				// fmt.Println("Setting stacka to", uv.store[uv.index])

				(*stack)[i.A] = uv.store[uv.index]
			}
			pc++
		case 10: // SETUPVAL
			if uv := upvals[i.B]; !uv.selfRef {
				uv.store[uv.index] = (*stack)[i.A]
			} else {
				uv.value = (*stack)[i.A]
			}
			pc++
		case 11: // CLOSEUPVALS
			A := i.A

			for n, uv := range openUpvals {
				if uv == nil || uv.selfRef || uv.index < A {
					continue
				}
				// fmt.Println("closing upvalue", uv)
				uv.value = uv.store[uv.index]
				uv.store = nil
				uv.selfRef = true
				openUpvals[n] = nil
				// fmt.Println("closed", uv)
			}
			pc++
		case 12: // GETIMPORT
			if err := getImport(i, towrap, stack); err != nil {
				return nil, err
			}
			pc += 2 // -- adjust for aux
		case 13: // GETTABLE
			if (*stack)[i.A], err = gettable((*stack)[i.C], (*stack)[i.B]); err != nil {
				return nil, err
			}
			pc++
		case 14: // SETTABLE
			index := (*stack)[i.C]
			t, ok := (*stack)[i.B].(*Table) // SETTABLE or SETTABLEKS on a types.Vector actually does return "attempt to index vector with 'whatever'"
			if !ok {
				return nil, invalidIndex(TypeOf((*stack)[i.B]), index)
			}
			if t.Readonly {
				return nil, errReadonly
			}

			// fmt.Println("SETTABLE", index, (*stack)[i.A])
			t.Set(index, (*stack)[i.A])
			pc++
		case 15: // GETTABLEKS
			if (*stack)[i.A], err = gettable(i.K, (*stack)[i.B]); err != nil {
				return nil, err
			}
			pc += 2 // -- adjust for aux
		case 16: // SETTABLEKS
			index := i.K
			t, ok := (*stack)[i.B].(*Table)
			if !ok {
				// fmt.Println("indexing", typeOf((*stack)[i.B]), "with", index)
				return nil, invalidIndex(TypeOf((*stack)[i.B]), index)
			}
			if t.Readonly {
				return nil, errReadonly
			}

			t.Set(index, (*stack)[i.A])
			pc += 2 // -- adjust for aux
		case 17: // GETTABLEN
			t := (*stack)[i.B].(*Table)
			idx := i.C + 1

			(*stack)[i.A] = t.Get(float64(idx))
			pc++
		case 18: // SETTABLEN
			t := (*stack)[i.B].(*Table)
			if t.Readonly {
				return nil, errReadonly
			}

			// fmt.Println("SETTABLEN", i.C+1, (*stack)[i.A])
			t.SetInt(int(i.C)+1, (*stack)[i.A])
			pc++
		case 19: // NEWCLOSURE
			newClosure(&pc, i, towrap, p, stack, &openUpvals, upvals)
			pc++
		case 20: // NAMECALL
			pc++
			if err = namecall(&pc, &top, &i, p, stack, co, &op); err != nil {
				return
			}
		case 21: // CALL
			if err = call(&top, i, towrap, stack, co); err != nil {
				return
			}
			pc++
		case 22: // RETURN
			pc++
			A := int32(i.A)
			b := int32(i.B) - 1

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
			if (*stack)[i.A] == (*stack)[i.Aux] {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 28:
			if j, err := jumpLe((*stack)[i.A], (*stack)[i.Aux]); err != nil {
				return nil, err
			} else if j {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 29:
			if j, err := jumpLt((*stack)[i.A], (*stack)[i.Aux]); err != nil {
				return nil, err
			} else if j {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 30:
			if (*stack)[i.A] != (*stack)[i.Aux] {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 31:
			if j, err := jumpGt((*stack)[i.A], (*stack)[i.Aux]); err != nil {
				return nil, err
			} else if j {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 32:
			if j, err := jumpGe((*stack)[i.A], (*stack)[i.Aux]); err != nil {
				return nil, err
			} else if j {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 33: // arithmetic
			if (*stack)[i.A], err = aAdd((*stack)[i.B], (*stack)[i.C]); err != nil {
				return
			}
			pc++
		case 34:
			if (*stack)[i.A], err = aSub((*stack)[i.B], (*stack)[i.C]); err != nil {
				return
			}
			pc++
		case 35:
			if (*stack)[i.A], err = aMul((*stack)[i.B], (*stack)[i.C]); err != nil {
				return
			}
			pc++
		case 36:
			if (*stack)[i.A], err = aDiv((*stack)[i.B], (*stack)[i.C]); err != nil {
				return
			}
			pc++
		case 37:
			if (*stack)[i.A], err = aMod((*stack)[i.B], (*stack)[i.C]); err != nil {
				return
			}
			pc++
		case 38:
			if (*stack)[i.A], err = aPow((*stack)[i.B], (*stack)[i.C]); err != nil {
				return
			}
			pc++
		case 81:
			if (*stack)[i.A], err = aIdiv((*stack)[i.B], (*stack)[i.C]); err != nil {
				return
			}
			pc++
		case 39: // arithmetik
			if (*stack)[i.A], err = aAdd((*stack)[i.B], i.K); err != nil {
				return
			}
			pc++
		case 40:
			if (*stack)[i.A], err = aSub((*stack)[i.B], i.K); err != nil {
				return
			}
			pc++
		case 41:
			if (*stack)[i.A], err = aMul((*stack)[i.B], i.K); err != nil {
				return
			}
			pc++
		case 42:
			if (*stack)[i.A], err = aDiv((*stack)[i.B], i.K); err != nil {
				return
			}
			pc++
		case 43:
			if (*stack)[i.A], err = aMod((*stack)[i.B], i.K); err != nil {
				return
			}
			pc++
		case 44:
			if (*stack)[i.A], err = aPow((*stack)[i.B], i.K); err != nil {
				return
			}
			pc++
		case 82:
			if (*stack)[i.A], err = aIdiv((*stack)[i.B], i.K); err != nil {
				return
			}
			pc++
		case 45: // logic AND
			if a, b := (*stack)[i.B], (*stack)[i.C]; truthy(a) {
				(*stack)[i.A] = b
			} else {
				(*stack)[i.A] = a
			}
			pc++
		case 46: // logic OR
			if a, b := (*stack)[i.B], (*stack)[i.C]; truthy(a) {
				(*stack)[i.A] = a
			} else if truthy(b) {
				(*stack)[i.A] = b
			} else {
				(*stack)[i.A] = false
			}
			pc++
		case 47: // logik AND
			// fmt.Println("LOGIK")
			if a, b := (*stack)[i.B], i.K; truthy(a) {
				(*stack)[i.A] = b
			} else {
				(*stack)[i.A] = a
			}
			pc++
		case 48: // logik OR
			// fmt.Println("LOGIK")
			if a, b := (*stack)[i.B], i.K; truthy(a) {
				(*stack)[i.A] = a
			} else if truthy(b) {
				(*stack)[i.A] = b
			} else {
				(*stack)[i.A] = false
			}
			pc++
		case 49: // CONCAT
			s := strings.Builder{}

			var first uint8
			for n := i.B; n <= i.C; n++ {
				toWrite, ok := (*stack)[n].(string)
				if !ok {
					// ensure correct order of operands in error message
					return nil, invalidConcat(TypeOf((*stack)[n-first]), TypeOf((*stack)[n+1-first]))
				}
				s.WriteString(toWrite)
				first = 1
			}
			(*stack)[i.A] = s.String()
			pc++
		case 50: // NOT
			(*stack)[i.A] = !truthy((*stack)[i.B])
			pc++
		case 51: // MINUS
			a, ok := (*stack)[i.B].(float64)
			if !ok {
				return nil, invalidUnm(TypeOf((*stack)[i.B]))
			}

			(*stack)[i.A] = -a
			pc++
		case 52: // LENGTH
			switch t := (*stack)[i.B].(type) {
			case *Table:
				(*stack)[i.A] = float64(t.Len())
			case string:
				(*stack)[i.A] = float64(len(t))
			default:
				return nil, invalidLength(TypeOf(t))
			}
			pc++
		case 53: // NEWTABLE
			(*stack)[i.A] = &Table{}

			pc += 2 // -- adjust for aux
		case 54: // DUPTABLE
			serialised := &Table{}
			// fmt.Println("TEMPLATING")
			for _, id := range i.K.([]uint32) { // template
				serialised.Set(p.K[id], nil) // constants
			}
			(*stack)[i.A] = serialised
			pc++
		case 55: // SETLIST
			A, B := i.A, int32(i.B)
			c := int32(i.C) - 1

			if c == luau_multret {
				c = top - B + 1
			}

			s := (*stack)[A].(*Table)
			if s.Readonly {
				return nil, errReadonly
			}

			// one-indexed lol
			for n, v := range (*stack)[B:min(B+c, int32(len(*stack)))] {
				s.SetInt(n+int(i.Aux), v)
			}
			// (*stack)[A] = s // in-place

			pc += 2 // -- adjust for aux
		case 56: // FORNPREP
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
			} else if index < limit {
				pc += i.D
			}
			pc++
		case 57: // FORNLOOP
			A := i.A
			limit := (*stack)[A].(float64)
			step := (*stack)[A+1].(float64)
			init := (*stack)[A+2].(float64) + step

			(*stack)[A+2] = init

			if step > 0 {
				if limit >= init {
					pc += i.D
				}
			} else if limit <= init {
				pc += i.D
			}
			pc++
		case 58: // FORGLOOP
			if err := forgloop(&pc, &top, i, stack, co, &generalisedIterators); err != nil {
				return nil, err
			}
		case 59, 61: // FORGPREP_INEXT, FORGPREP_NEXT
			if _, ok := (*stack)[i.A].(types.Function); !ok {
				return nil, fmt.Errorf("attempt to iterate over a %s value", TypeOf((*stack)[i.A])) // -- encountered non-function value
			}
			pc += i.D + 1
		case 60: // FASTCALL3
			// Skipped
			pc += 2 // adjust for aux
		case 63: // GETVARARGS
			A := int32(i.A)
			b := int32(i.B) - 1

			// fmt.Println("MULTRET", b, vargsLen)
			if b == luau_multret {
				b = int32(vargsLen)
				top = A + b - 1
			}

			// stack may get expanded here
			// (MAX STACK SIZE IS A LIE!!!!!!!!!!!!!!!!!!!!!!!)
			moveStack(stack, vargsList, b, A)
			pc++
		case 64: // DUPCLOSURE
			dupClosure(&pc, i, towrap, p, stack, upvals)
			pc++
		case 65: // PREPVARARGS
			// Handled by wrapper
			pc++
		case 66: // LOADKX
			(*stack)[i.A] = float64(i.K.(uint32)) // kv (graah)

			pc += 2 // -- adjust for aux
		case 67: // JUMPX
			pc += i.E + 1
		case 68: // FASTCALL
			pc++
			// Skipped
		case 69: // COVERAGE
			i.E++
			pc++
		case 70: // CAPTURE
			// Handled by CLOSURE
			panic("encountered unhandled CAPTURE")
		case 71: // SUBRK
			if (*stack)[i.A], err = aSub(i.K, (*stack)[i.C]); err != nil {
				return nil, err
			}
			pc++
		case 72: // DIVRK
			if (*stack)[i.A], err = aDiv(i.K, (*stack)[i.C]); err != nil {
				return nil, err
			}
			pc++
		case 73: // FASTCALL1
			// Skipped
			pc++
		case 74, 75: // FASTCALL2, FASTCALL2K
			// Skipped
			pc += 2 // adjust for aux
		case 76: // FORGPREP
			pc += i.D + 1
			if _, ok := (*stack)[i.A].(types.Function); ok {
				break
			}

			loopInst := *p.Code[pc-1]
			if generalisedIterators[loopInst] != nil {
				break
			}

			c := &iterator{
				args:   make(chan *Table),
				resume: make(chan *[]types.Val),
			}
			go iterate(c)
			// fmt.Println("SETTING GENITER", loopInst)
			generalisedIterators[loopInst] = c
		case 77: // JUMPXEQKNIL
			if ra := (*stack)[i.A]; ra == nil != i.KN {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 78: //  JUMPXEQKB
			if kv, ra := i.K.(bool), (*stack)[i.A].(bool); ra == kv != i.KN {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 79: // JUMPXEQKN
			if kv, ra := i.K.(float64), (*stack)[i.A].(float64); ra == kv != i.KN {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 80: // JUMPXEQKS
			if kv, ra := i.K.(string), (*stack)[i.A].(string); ra == kv != i.KN {
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

func wrapclosure(towrap toWrap) types.Function {
	proto := towrap.proto

	return fn("", func(co *types.Coroutine, args ...types.Val) (r []types.Val, err error) {
		maxs, np := proto.MaxStackSize, proto.NumParams // maxs 2 lel

		la := uint8(len(args)) // we can't have more than 255 args anyway right?
		// fmt.Println("MAX STACK SIZE", maxs)
		stack := make([]types.Val, maxs)
		copy(stack, args[:min(np, la)])

		var list []types.Val
		if np < la {
			list = args[np:]
		}

		originalDebug := co.Dbg

		dbg := &types.Debugging{ /* enabled: proto.lineinfoenabled, opcode: 255 */ }
		// fmt.Println("started on", co.Dbg.line, dbg.line)
		co.Dbg = dbg

		r, err = execute(towrap, &stack, co, list, max(la-np, 0))
		// fmt.Println("ended on", co.Dbg.line, dbg.line)
		if !*towrap.alive {
			return
		}
		if err != nil {
			return nil, &types.CoError{
				Line:    dbg.Line,
				Dbgname: dbg.Name,
				Path:    co.Dbgpath,
				Sub:     err,
			}
		}

		// prevent line mismatches (error/loc.luau)
		co.Dbg = originalDebug
		return
	})
}

func loadmodule(m compiled, env types.Env, requireCache map[string]types.Val, args types.ProgramArgs) (co types.Coroutine, cancel func()) {
	alive := true

	towrap := toWrap{
		proto:        m.MainProto,
		protoList:    m.ProtoList,
		alive:        &alive,
		env:          env,
		requireCache: requireCache,
	}

	return types.Coroutine{
		Body:           wrapclosure(towrap),
		Env:            env,
		Filepath:       m.Filepath,
		Dbgpath:        m.Dbgpath,
		RequireHistory: m.RequireHistory,
		YieldChan:      make(chan types.Yield, 1),
		ResumeChan:     make(chan []types.Val, 1),
		Compiler:       m.Compiler,
		ProgramArgs:    args,
	}, func() { alive = false }
}
