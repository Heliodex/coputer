// package vm implements Litecode's deterministic Luau virtual machine and standard library toolkit.
package vm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"

	// unsafe code paths removed... for now

	"github.com/Heliodex/coputer/litecode/internal"
	. "github.com/Heliodex/coputer/litecode/types"
)

var (
	errReadonly = errors.New("attempt to modify a readonly table")
	errNilIndex = errors.New("table index is nil")
)

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

var opList = [83]internal.OpInfo{
	{Mode: 0, KMode: 0, HasAux: false}, // NOP
	{Mode: 0, KMode: 0, HasAux: false}, // BREAK
	{Mode: 1, KMode: 0, HasAux: false}, // LOADNIL
	{Mode: 3, KMode: 0, HasAux: false}, // LOADB
	{Mode: 4, KMode: 0, HasAux: false}, // LOADN
	{Mode: 4, KMode: 3, HasAux: false}, // LOADK
	{Mode: 2, KMode: 0, HasAux: false}, // MOVE
	{Mode: 1, KMode: 1, HasAux: true},  // GETGLOBAL
	{Mode: 1, KMode: 1, HasAux: true},  // SETGLOBAL
	{Mode: 2, KMode: 0, HasAux: false}, // GETUPVAL
	{Mode: 2, KMode: 0, HasAux: false}, // SETUPVAL
	{Mode: 1, KMode: 0, HasAux: false}, // CLOSEUPVALS
	{Mode: 4, KMode: 4, HasAux: true},  // GETIMPORT
	{Mode: 3, KMode: 0, HasAux: false}, // GETTABLE
	{Mode: 3, KMode: 0, HasAux: false}, // SETTABLE
	{Mode: 3, KMode: 1, HasAux: true},  // GETTABLEKS
	{Mode: 3, KMode: 1, HasAux: true},  // SETTABLEKS
	{Mode: 3, KMode: 0, HasAux: false}, // GETTABLEN
	{Mode: 3, KMode: 0, HasAux: false}, // SETTABLEN
	{Mode: 4, KMode: 0, HasAux: false}, // NEWCLOSURE
	{Mode: 3, KMode: 1, HasAux: true},  // NAMECALL
	{Mode: 3, KMode: 0, HasAux: false}, // CALL
	{Mode: 2, KMode: 0, HasAux: false}, // RETURN
	{Mode: 4, KMode: 0, HasAux: false}, // JUMP
	{Mode: 4, KMode: 0, HasAux: false}, // JUMPBACK
	{Mode: 4, KMode: 0, HasAux: false}, // JUMPIF
	{Mode: 4, KMode: 0, HasAux: false}, // JUMPIFNOT
	{Mode: 4, KMode: 0, HasAux: true},  // JUMPIFEQ
	{Mode: 4, KMode: 0, HasAux: true},  // JUMPIFLE
	{Mode: 4, KMode: 0, HasAux: true},  // JUMPIFLT
	{Mode: 4, KMode: 0, HasAux: true},  // JUMPIFNOTEQ
	{Mode: 4, KMode: 0, HasAux: true},  // JUMPIFNOTLE
	{Mode: 4, KMode: 0, HasAux: true},  // JUMPIFNOTLT
	{Mode: 3, KMode: 0, HasAux: false}, // ADD
	{Mode: 3, KMode: 0, HasAux: false}, // SUB
	{Mode: 3, KMode: 0, HasAux: false}, // MUL
	{Mode: 3, KMode: 0, HasAux: false}, // DIV
	{Mode: 3, KMode: 0, HasAux: false}, // MOD
	{Mode: 3, KMode: 0, HasAux: false}, // POW
	{Mode: 3, KMode: 2, HasAux: false}, // ADDK
	{Mode: 3, KMode: 2, HasAux: false}, // SUBK
	{Mode: 3, KMode: 2, HasAux: false}, // MULK
	{Mode: 3, KMode: 2, HasAux: false}, // DIVK
	{Mode: 3, KMode: 2, HasAux: false}, // MODK
	{Mode: 3, KMode: 2, HasAux: false}, // POWK
	{Mode: 3, KMode: 0, HasAux: false}, // AND
	{Mode: 3, KMode: 0, HasAux: false}, // OR
	{Mode: 3, KMode: 2, HasAux: false}, // ANDK
	{Mode: 3, KMode: 2, HasAux: false}, // ORK
	{Mode: 3, KMode: 0, HasAux: false}, // CONCAT
	{Mode: 2, KMode: 0, HasAux: false}, // NOT
	{Mode: 2, KMode: 0, HasAux: false}, // MINUS
	{Mode: 2, KMode: 0, HasAux: false}, // LENGTH
	{Mode: 2, KMode: 0, HasAux: true},  // NEWTABLE
	{Mode: 4, KMode: 3, HasAux: false}, // DUPTABLE
	{Mode: 3, KMode: 0, HasAux: true},  // SETLIST
	{Mode: 4, KMode: 0, HasAux: false}, // FORNPREP
	{Mode: 4, KMode: 0, HasAux: false}, // FORNLOOP
	{Mode: 4, KMode: 8, HasAux: true},  // FORGLOOP
	{Mode: 4, KMode: 0, HasAux: false}, // FORGPREP_INEXT
	{Mode: 3, KMode: 1, HasAux: true},  // FASTCALL3
	{Mode: 4, KMode: 0, HasAux: false}, // FORGPREP_NEXT
	{Mode: 0, KMode: 0, HasAux: false}, // FORGLOOP_NEXT (deprecated)
	{Mode: 2, KMode: 0, HasAux: false}, // GETVARARGS
	{Mode: 4, KMode: 3, HasAux: false}, // DUPCLOSURE
	{Mode: 1, KMode: 0, HasAux: false}, // PREPVARARGS
	{Mode: 1, KMode: 1, HasAux: true},  // LOADKX
	{Mode: 5, KMode: 0, HasAux: false}, // JUMPX
	{Mode: 3, KMode: 0, HasAux: false}, // FASTCALL
	{Mode: 5, KMode: 0, HasAux: false}, // COVERAGE
	{Mode: 2, KMode: 0, HasAux: false}, // CAPTURE
	{Mode: 3, KMode: 7, HasAux: false}, // SUBRK
	{Mode: 3, KMode: 7, HasAux: false}, // DIVRK
	{Mode: 3, KMode: 0, HasAux: false}, // FASTCALL1
	{Mode: 3, KMode: 0, HasAux: true},  // FASTCALL2
	{Mode: 3, KMode: 1, HasAux: true},  // FASTCALL2K
	{Mode: 4, KMode: 0, HasAux: false}, // FORGPREP
	{Mode: 4, KMode: 5, HasAux: true},  // JUMPXEQKNIL
	{Mode: 4, KMode: 5, HasAux: true},  // JUMPXEQKB
	{Mode: 4, KMode: 6, HasAux: true},  // JUMPXEQKN
	{Mode: 4, KMode: 6, HasAux: true},  // JUMPXEQKS
	{Mode: 3, KMode: 0, HasAux: false}, // IDIV
	{Mode: 3, KMode: 2, HasAux: false}, // IDIVK
}

// Functions and Tables are used as pointers normally, as they need to be hashed

func fn(name string, co *Coroutine, f func(co *Coroutine, args ...Val) (r []Val, err error)) Function {
	return Function{
		Run:  &f,
		Name: name,
		Co:   co,
	}
}

func createCoroutine(body Function, currentCo *Coroutine) *Coroutine {
	// first time i actually ran into the channel axiom issues
	return &Coroutine{
		Body:       body,
		Filepath:   currentCo.Filepath,
		Dbgpath:    currentCo.Dbgpath,
		YieldChan:  make(chan internal.Yield, 1),
		ResumeChan: make(chan []Val, 1),
	}
}

const luau_multret = -1

func namecallHandler(co *Coroutine, kv string, stack *[]Val, c1, c2 int32) (ok bool, retList []Val, err error) {
	switch kv {
	case "format":
		str := (*stack)[c1].(string)
		args := (*stack)[1:][c1:c2]

		f, err := fmtstring(str, Args{Co: co, List: args, name: "format"})
		if err != nil {
			return false, nil, err
		}
		return true, []Val{f}, nil
	}
	return
}

// The only global environment, never mutated (yes, it only stores string keys)
var exts = map[string]Val{
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
	// "typeof":   MakeFn("typeof", global_type), // same because no metatables
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

func checkkmode(i *internal.Inst, k []Val) {
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
		i.K0 = k[id0].(string) // lmk if this panics lol
		// fmt.Println("AUX", i.K0)

		if count < 2 {
			break
		}
		id1 := extend >> 10 & 0x3ff
		i.K1 = k[id1].(string)

		if count < 3 { // should never be >3
			break
		}
		id2 := extend & 0x3ff
		i.K2 = k[id2].(string)
	case 5: // AUX boolean low 1 bit
		i.K = i.Aux&1 == 1
		i.KN = i.Aux>>31 == 1
	case 6: // AUX number low 24 bits
		i.K = k[i.Aux&(1<<24-1)]
		i.KN = i.Aux>>31 == 1
	case 7: // B
		i.K = k[i.B]
	case 8: // AUX number low 16 bits ig
		i.K = uint8(i.Aux & 0xf) // forgloop
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

func (s *stream) rUint32() (w uint32) {
	w = binary.LittleEndian.Uint32(s.data[s.pos:])
	s.pos += 4
	return
}

func (s *stream) skipUint32() {
	s.pos += 4
}

// this is the only thing float32s are ever used for anyway
func (s *stream) rVector() (r Vector) {
	for i := range 4 {
		r[i] = math.Float32frombits(s.rUint32())
	}
	return
}

func (s *stream) rFloat64() (r float64) {
	r = math.Float64frombits(binary.LittleEndian.Uint64(s.data[s.pos:]))

	s.pos += 8
	return
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

func (s *stream) skipVarInt() {
	for range 4 {
		if s.rByte()&0b1000_0000 == 0 {
			return
		}
	}
}

func (s *stream) rString() (str string) {
	size := s.rVarInt()
	str = string(s.data[s.pos:][:size])

	s.pos += size
	return
}

func (s *stream) checkEnd() error {
	if s.pos != uint32(len(s.data)) {
		return errors.New("deserialiser position mismatch")
	}
	return nil
}

// reads either 1 or 2 words
func (s *stream) readInst(code *[]*internal.Inst) bool {
	value := s.rUint32()

	opcode := uint8(value)
	opinfo := opList[opcode]

	i := internal.Inst{
		Opcode: opcode,
		KMode:  opinfo.KMode,
	}

	// value >>= 8 // uint24 I guess
	switch opinfo.Mode {
	case 5: // other A lol
		if i.A = int32(value >> 8); i.A >= 0x800000 { // why no arbitrary width integers, go
			i.A -= 0x1000000
		}
	case 4: // AD
		i.A = int32(uint8(value >> 8))
		i.D = int32(int16(value >> 16))
	case 3: // ABC
		i.C = uint8(value >> 24)
		fallthrough
	case 2: // AB
		i.B = uint8(value >> 16)
		fallthrough
	case 1: // A
		i.A = int32(uint8(value >> 8)) // 8 bit
	}

	*code = append(*code, &i)
	if opinfo.HasAux {
		i.Aux = s.rUint32()

		*code = append(*code, &internal.Inst{})
		return true
	}

	return false
}

func (s *stream) readLineInfo(sizecode uint32) (instLineInfo []uint32) {
	linegaplog2 := s.rByte()

	lineinfo := make([]uint8, sizecode)
	var lastoffset uint8
	for i := range sizecode {
		lastoffset += s.rByte() // yeah this is how the reference implementation does it, whatever hpppens happens
		lineinfo[i] = lastoffset
	}

	intervals := (sizecode-1)>>linegaplog2 + 1
	abslineinfo := make([]uint32, intervals)
	var lastline uint32
	for i := range intervals {
		lastline += s.rUint32()
		// fmt.Println("lastline", lastline)
		abslineinfo[i] = lastline // overflow babyy (faster than % (1 << 32))
	}

	instLineInfo = make([]uint32, sizecode)
	for i, v := range lineinfo {
		// -- p->abslineinfo[pc >> p->linegaplog2] + p->lineinfo[pc];
		instLineInfo[i] = abslineinfo[i>>linegaplog2] + uint32(v)
	}
	return
}

func (s *stream) readDebugInfo() {
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

func (s *stream) readProto(stringList []string) (p *internal.Proto, err error) {
	p = &internal.Proto{
		MaxStackSize: s.rByte(),
		NumParams:    s.rByte(),
		Nups:         s.rByte(),
	}

	// s.rBool()            // isvararg
	// native := s.rByte() // -- flags
	// if (native&1) != 0 || (native&4) != 0 {
	// 	return nil, fmt.Errorf("native function not supported")
	// }

	s.pos += 2
	s.pos += s.rVarInt() // typesize

	sizecode := s.rVarInt()
	for i := uint32(0); i < sizecode; i++ {
		if s.readInst(&p.Code) {
			i++
		}
	}

	sizek := s.rVarInt()
	K := make([]Val, sizek) // krazy

	for i := range sizek {
		switch kt := s.rByte(); kt {
		case 0: // Nil
			// yeah
		case 1: // Bool
			K[i] = s.rBool()
		case 2: // Number
			K[i] = s.rFloat64()
		case 3: // String
			K[i] = stringList[s.rVarInt()-1]
		case 4: // Import
			// only used with useImportConstants
			s.skipUint32()
		case 5: // Table
			// moot, whatever
			for range s.rVarInt() {
				s.skipVarInt()
			}
		case 6: // Closure
			// pain in the cranium
			K[i] = s.rVarInt() // ⚠️ not a val ⚠️
			// fmt.Println("case 6", p.K[i])
		case 7: // Vector
			K[i] = s.rVector()
		default:
			return nil, fmt.Errorf("unknown ktype %d", kt)
		}
	}

	// -- 2nd pass to replace constant references in the instruction
	for i := range sizecode {
		checkkmode(p.Code[i], K)
	}

	sizep := s.rVarInt()
	p.Protos = make([]uint32, sizep)
	for i := range sizep {
		p.Protos[i] = s.rVarInt()
	}

	// p.LineDefined = s.rVarInt()
	s.skipVarInt()

	if dbgnamei := s.rVarInt(); dbgnamei == 0 {
		p.Dbgname = "(??)"
	} else {
		p.Dbgname = stringList[dbgnamei-1]
	}

	// LineInfoEnabled
	if s.rBool() {
		p.InstLineInfo = s.readLineInfo(sizecode)
	}

	if s.rBool() {
		s.readDebugInfo()
	}

	return
}

func deserialise(b []byte) (d internal.Deserialised, err error) {
	s := &stream{data: b}

	if luauVersion := s.rByte(); luauVersion == 0 {
		return internal.Deserialised{}, errors.New("the provided bytecode is an error message")
	} else if luauVersion != 6 {
		return internal.Deserialised{}, errors.New("the version of the provided bytecode is unsupported")
	}
	if s.rByte() != 3 { // types version
		return internal.Deserialised{}, errors.New("the types version of the provided bytecode is unsupported")
	}

	stringCount := s.rVarInt()
	stringList := make([]string, stringCount)
	for i := range stringCount {
		stringList[i] = s.rString()
	}

	// const userdataTypeLimit = 32

	// userdata type remapping table
	// for unknown userdata types, the entry will remap to common 'userdata' type
	// for idx := s.rByte(); idx != 0; idx = s.rByte() {
	// 	name := stringList[s.rVarInt()]

	// 	if idx-1 < userdataTypeLimit {}
	// }

	// (not used in VM, left unused)
	for s.rBool() {
		s.skipVarInt()
	}

	protoCount := s.rVarInt()
	protoList := make([]*internal.Proto, protoCount)
	for i := range protoCount {
		if protoList[i], err = s.readProto(stringList); err != nil {
			return
		}
	}

	mainProto := protoList[s.rVarInt()]
	mainProto.Dbgname = "(main)"

	return internal.Deserialised{
		MainProto: mainProto,
		ProtoList: protoList,
	}, s.checkEnd()
}

type upval struct {
	Val
	store *Val
}

func falsy(v Val) bool {
	return v == nil || v == false
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

func invalidIndex(ta string, v Val) error {
	tb := TypeOf(v)
	if tb == "string" {
		tb = fmt.Sprintf("'%v'", v)
	}

	return fmt.Errorf("attempt to index %v with %v", ta, tb)
}

func invalidIter(t string) error {
	return fmt.Errorf("attempt to iterate over a %s value", t)
}

func missingMethod(ta string, v Val) error {
	tb := TypeOf(v)
	if tb == "string" {
		tb = fmt.Sprintf("'%v'", v)
	}

	return fmt.Errorf("attempt to call missing method %v of %v", tb, ta)
}

// TypeOf returns the underlying VM datatype of a value as a string.
// This does not return the Luau type, as type() does.
func TypeOf(v Val) string {
	if v == nil {
		return "nil"
	}

	switch v.(type) {
	case float64:
		return "number"
	case string:
		return "string"
	case bool:
		return "boolean"
	case *Table:
		return "table"
	case Function:
		return "function"
	case *Coroutine:
		return "thread"
	case *Buffer:
		return "buffer"
	case Vector:
		return "vector"
	}
	return "userdata"
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
	switch {
	case ok3 && ok4:
		return Vector{va[0] * vb[0], va[1] * vb[1], va[2] * vb[2], va[3] * vb[3]}, nil
	case ok1 && ok4:
		f := float32(fa)
		return Vector{f * vb[0], f * vb[1], f * vb[2], f * vb[3]}, nil
	case ok3 && ok2:
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
	switch {
	case ok3 && ok4:
		return Vector{va[0] / vb[0], va[1] / vb[1], va[2] / vb[2], va[3] / vb[3]}, nil
	case ok1 && ok4:
		f := float32(fa)
		return Vector{f / vb[0], f / vb[1], f / vb[2], f / vb[3]}, nil
	case ok3 && ok2:
		f := float32(fb)
		return Vector{va[0] / f, va[1] / f, va[2] / f, va[3] / f}, nil
	}

	return nil, invalidArithmetic("div", TypeOf(a), TypeOf(b))
}

func aMod(a, b Val) (Val, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa - fb*math.Floor(fa/fb), nil
	}

	return nil, invalidArithmetic("mod", TypeOf(a), TypeOf(b))
}

func aPow(a, b Val) (Val, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return math.Pow(fa, fb), nil
	}

	return nil, invalidArithmetic("pow", TypeOf(a), TypeOf(b))
}

func aIdiv(a, b Val) (Val, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return math.Floor(fa / fb), nil
	}

	va, ok3 := a.(Vector)
	vb, ok4 := b.(Vector)
	switch {
	case ok3 && ok4:
		return Vector{
			f32Floor(va[0] / vb[0]),
			f32Floor(va[1] / vb[1]),
			f32Floor(va[2] / vb[2]),
			f32Floor(va[3] / vb[3]),
		}, nil
	case ok1 && ok4:
		f := float32(fa)
		return Vector{
			f32Floor(f / vb[0]),
			f32Floor(f / vb[1]),
			f32Floor(f / vb[2]),
			f32Floor(f / vb[3]),
		}, nil
	case ok3 && ok2:
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

func aUnm(a Val) (Val, error) {
	fa, ok1 := a.(float64)
	if ok1 {
		return -fa, nil
	}

	va, ok2 := a.(Vector)
	if ok2 {
		return Vector{-va[0], -va[1], -va[2], -va[3]}, nil
	}

	return nil, invalidUnm(TypeOf(a))
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

func gettable(k, v Val) (Val, error) {
	switch t := v.(type) {
	case *Table:
		return t.Get(k), nil
	case Vector: // direction,,, and mmmagnitude!! oh yeah!!11!!
		switch k {
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

	return nil, invalidIndex(TypeOf(v), k)
}

type toWrap struct {
	proto     *internal.Proto
	protoList []*internal.Proto
	upvals    []*upval
	alive     *bool
	env       Env
	// Store the last return, as it's the only one that's relevant
	requireCache map[string]Val
}

func moveStack(stack *[]Val, src []Val, b, t int32) {
	l := t + b - int32(len(*stack)) + 1 // graah stack expansion
	*stack = append(*stack, make([]Val, max(l, 0))...)

	for i := range b {
		if i >= int32(len(src)) {
			(*stack)[t+i] = nil
			continue
		}
		(*stack)[t+i] = src[i]
	}
}

func getImport(A, count uint8, K0, K1, K2 string, towrap toWrap, stack *[]Val) (err error) {
	imp := exts[K0]
	if imp == nil {
		imp = towrap.env[K0]
	}

	if count < 2 {
		(*stack)[A] = imp
		return
	}

	t1, ok := imp.(*Table)
	if !ok {
		return invalidIndex("nil", K1)
	}

	imp = t1.GetHash(K1)
	// fmt.Println("GETIMPORT2", i.A, (*stack)[i.A])

	if count < 3 {
		(*stack)[A] = imp
		return
	}

	t2, ok := imp.(*Table)
	if !ok {
		return invalidIndex(TypeOf(imp), K2)
	}

	(*stack)[A] = t2.GetHash(K2)
	// fmt.Println("GETIMPORT3", i.A, (*stack)[i.A])

	return
}

func newClosure(pc *int32, A uint8, towrap toWrap, code []*internal.Inst, stack *[]Val, co *Coroutine, openUpvals *[]*upval, upvals []*upval) {
	nups := towrap.proto.Nups
	towrap.upvals = make([]*upval, nups)

	(*stack)[A] = wrapclosure(towrap, co)
	// fmt.Println("WRAPPING WITH", uvs)

	// fmt.Println("nups", nups)
	for n := range nups {
		switch pseudo := code[*pc+1]; pseudo.A {
		case 0: // -- value
			towrap.upvals[n] = &upval{
				Val: (*stack)[pseudo.B],
			}
		case 1: // -- reference
			// -- references dont get handled by DUPCLOSURE
			idx := pseudo.B
			// fmt.Println("index", idx, len(openUpvals))
			// for si, sv := range *stack {
			// 	fmt.Printf("  [%d] = %v\n", si, sv)
			// }

			var prev *upval
			if idx < uint8(len(*openUpvals)) {
				prev = (*openUpvals)[idx]
			}

			if prev == nil {
				prev = &upval{
					store: &(*stack)[idx],
				}

				l := int(idx) - len(*openUpvals) + 1
				*openUpvals = append(*openUpvals, make([]*upval, max(l, 0))...)
				(*openUpvals)[idx] = prev
			}

			towrap.upvals[n] = prev
			// fmt.Println("set upvalue", i, "to", prev)
		case 2: // -- upvalue
			towrap.upvals[n] = upvals[pseudo.B]
		}
		*pc++
	}
}

func namecall(pc, top *int32, i *internal.Inst, code []*internal.Inst, lineInfo []uint32, stack *[]Val, co *Coroutine, op *uint8) (err error) {
	kv := i.K.(string)
	// fmt.Println("kv", kv)

	(*stack)[i.A+1] = (*stack)[i.B]

	// -- Special handling for native namecall behaviour
	callInst := code[*pc+1]

	// -- Copied from the CALL handler
	callA, callB, callC := callInst.A, int32(callInst.B), callInst.C

	var params int32
	if callB == 0 {
		params = *top - callA
	} else {
		params = callB
	}

	ok, retList, err := namecallHandler(co, kv, stack, callA+1, callA+params-1)
	if err != nil {
		return
	}
	if !ok {
		// fmt.Println("namecall", kv, "not found")
		switch t := (*stack)[i.B].(type) {
		case *Table:
			call := t.GetHash(kv)
			if call == nil {
				return missingMethod(TypeOf(t), kv)
			}

			(*stack)[i.A] = call
		case string:
			return missingMethod(TypeOf(t), kv)
		default:
			return invalidIndex(TypeOf(t), kv)
		}
		return
	}

	*i = *callInst
	*op = i.Opcode

	var retCount int32
	if callC == 0 { // todo: never runs
		retCount = int32(len(retList))
		*top = callA + retCount
	} else {
		retCount = int32(callC - 1)
	}

	moveStack(stack, retList, retCount, callA)
	*pc += 2 // -- adjust for aux, Skip next CALL instruction
	co.Dbg.Line = lineInfo[*pc]
	return
}

func handleRequire(towrap toWrap, lc compiled, co *Coroutine) (rets []Val, err error) {
	if c, ok := towrap.requireCache[lc.Filepath]; ok {
		return []Val{c}, nil
	}

	// since environments only store global libraries etc, using the same env here should be fine??
	c2, _ := loadmodule(lc, co.Env, towrap.requireCache, co.ProgramArgs)
	reqrets, err := c2.Resume()
	if err != nil {
		return
	}
	if len(reqrets) == 0 { // i have no reqrets
		return nil, errors.New("module must return a value")
	}

	// only the last return value (weird luau behaviour...)
	ret := reqrets[len(reqrets)-1]
	switch ret.(type) {
	case *Table, Function:
	default:
		return nil, errors.New("module must return a table or function")
	}

	towrap.requireCache[lc.Filepath] = ret
	return []Val{ret}, nil
}

func call(top *int32, A int32, B, C uint8, towrap toWrap, stack *[]Val, co *Coroutine) (err error) {
	// fmt.Println(A, B, C, (*stack)[A], params)

	f := (*stack)[A]
	fn, ok := f.(Function)
	// fmt.Println("calling with", (*stack)[A+1:][:params])
	if !ok {
		return uncallableType(TypeOf(f))
	}

	var params int32
	if B == 0 {
		params = *top - A
	} else {
		params = int32(B)
	}

	// fmt.Println("upvals1", len(upvals))
	rco := fn.Co
	if rco == nil { // make sure any function is called in the coroutine of its own file (mainly for correct error messages)
		rco = co
	}
	retList, err := (*fn.Run)(rco, (*stack)[A+1:][:params-1]...) // not inclusive
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

	if C == 0 {
		*top = A + retCount
	} else {
		retCount = int32(C - 1)
	}

	moveStack(stack, retList, retCount, A)
	return
}

// for gloop lel
func forgloop(pc, top *int32, i internal.Inst, stack *[]Val, co *Coroutine, genIters map[internal.Inst][]Val) (err error) {
	res := int32(i.K.(uint8)) // aux number low 16 bits

	switch s := (*stack)[i.A].(type) {
	case Function:
		// fmt.Println("IT func", fn, (*stack)[A+1], (*stack)[A+2])
		vals, err := (*s.Run)(co, (*stack)[i.A+1], (*stack)[i.A+2])
		if err != nil {
			return err
		}

		moveStack(stack, vals, res, i.A+3)
		// fmt.Println(A+3, (*stack)[A+3])

		if (*stack)[i.A+3] == nil {
			*pc += 2
			return nil
		}
	case *Table:
		// fmt.Println("GETTING GENITER", typeOf(it))
		it := genIters[i]

		if it == nil {
			// fmt.Println((*stack)[A+1], (*stack)[A+2]) // <nil> <nil>
			for i, v := range s.Iter() {
				// fmt.Println("yielding", i, v)
				it = append(it, i, v) // always in pairs
			}
		}

		if len(it) == 0 {
			delete(genIters, i) // don't touch my geniters
			*pc += 2
			return
		}

		moveStack(stack, it[:2], res, i.A+3)
		genIters[i] = it[2:] // q
	default:
		return invalidIter(TypeOf(s))
	}

	*top = i.A + 7
	(*stack)[i.A+2] = (*stack)[i.A+3]
	*pc += i.D + 1
	return
}

func execute(towrap toWrap, stack, vargsList []Val, co *Coroutine) (r []Val, err error) {
	p, upvals := towrap.proto, towrap.upvals
	// int32 > uint32 lel
	var pc, top int32
	var openUpvals []*upval
	var genIters map[internal.Inst][]Val

	// a a a a
	// stayin' alive
	// fmt.Println("starting with upvals", upvals)
	code, lineInfo, protos := p.Code, p.InstLineInfo, p.Protos
	co.Dbg.Name = p.Dbgname
	
	for ; *towrap.alive; co.Dbg.Line = lineInfo[pc] {
		// fmt.Println(top)

		// if len(upvals) > 0 {
		// 	fmt.Println("upval", upvals[0])
		// }

		i := *code[pc]
		switch op := i.Opcode; op {
		case 0: // NOP
			// -- Do nothing
			pc++
		case 2: // LOADNIL
			stack[i.A] = nil
			pc++
		case 3: // LOADB
			stack[i.A] = i.B == 1
			pc += int32(i.C + 1)
		case 4: // LOADN
			stack[i.A] = float64(i.D) // never put an int on the stack
			pc++
		case 5: // LOADK
			// fmt.Println("LOADK", i.A, i.K)
			stack[i.A] = i.K
			pc++
		case 6: // MOVE
			// we should (ALMOST) never have to change the size of the stack (p.maxstacksize)
			stack[i.A] = stack[i.B]
			pc++
		case 7: // GETGLOBAL
			kv := i.K.(string)

			if e, ok := exts[kv]; ok {
				stack[i.A] = e
			} else {
				stack[i.A] = towrap.env[kv]
			}
			pc += 2 // -- adjust for aux
		case 8: // SETGLOBAL
			// LOL
			kv := i.K.(string)

			if _, ok := exts[kv]; ok {
				return nil, fmt.Errorf("attempt to redefine global '%s'", kv)
			}
			return nil, fmt.Errorf("attempt to set global '%s'", kv)
		case 9: // GETUPVAL
			if uv := upvals[i.B]; uv.store == nil {
				stack[i.A] = uv.Val
			} else {
				// fmt.Println("GETTING UPVAL", uv)
				// fmt.Println("Setting stacka to", uv.store[uv.index])

				stack[i.A] = *uv.store
			}
			pc++
		case 10: // SETUPVAL
			if uv := upvals[i.B]; uv.store == nil {
				uv.Val = stack[i.A]
			} else {
				*uv.store = stack[i.A]
			}
			pc++
		case 11: // CLOSEUPVALS
			for n, uv := range openUpvals {
				if uv == nil {
					continue
				}
				// fmt.Println("closing upvalue", uv)
				uv.Val = *uv.store
				uv.store = nil
				openUpvals[n] = nil
				// fmt.Println("closed", uv)
			}
			pc++
		case 12: // GETIMPORT
			if err = getImport(uint8(i.A), i.KC, i.K0, i.K1, i.K2, towrap, &stack); err != nil {
				return
			}
			pc += 2 // -- adjust for aux
		case 13: // GETTABLE
			if stack[i.A], err = gettable(stack[i.C], stack[i.B]); err != nil {
				return
			}
			pc++
		case 14: // SETTABLE
			idx := stack[i.C]
			t, ok := stack[i.B].(*Table) // SETTABLE or SETTABLEKS on a Vector actually does return "attempt to index vector with 'whatever'"
			if !ok {
				return nil, invalidIndex(TypeOf(stack[i.B]), idx)
			}
			if t.Readonly {
				return nil, errReadonly
			}
			if idx == nil {
				return nil, errNilIndex
			}

			// fmt.Println("SETTABLE", idx, stack[i.A])
			t.Set(idx, stack[i.A])
			pc++
		case 15: // GETTABLEKS
			if stack[i.A], err = gettable(i.K, stack[i.B]); err != nil {
				return
			}
			pc += 2 // -- adjust for aux
		case 16: // SETTABLEKS
			idx := i.K
			t, ok := stack[i.B].(*Table)
			if !ok {
				// fmt.Println("indexing", typeOf(stack[i.B]), "with", idx)
				return nil, invalidIndex(TypeOf(stack[i.B]), idx)
			}
			if t.Readonly {
				return nil, errReadonly
			}
			if idx == nil {
				return nil, errNilIndex
			}

			t.Set(idx, stack[i.A])
			pc += 2 // -- adjust for aux
		case 17: // GETTABLEN
			idx := i.C + 1
			t, ok := stack[i.B].(*Table)
			if !ok {
				// fmt.Println("gettableninvalidindex")
				return nil, invalidIndex(TypeOf(stack[i.B]), float64(idx))
			}

			stack[i.A] = t.GetInt(int(idx))
			pc++
		case 18: // SETTABLEN
			idx := i.C + 1
			t, ok := stack[i.B].(*Table)
			if !ok {
				// fmt.Println("gettableninvalidindex")
				return nil, invalidIndex(TypeOf(stack[i.B]), float64(idx))
			}
			if t.Readonly {
				return nil, errReadonly
			}

			// fmt.Println("SETTABLEN", i.C+1, stack[i.A])
			t.SetInt(int(idx), stack[i.A])
			pc++
		case 19: // NEWCLOSURE
			towrap.proto = towrap.protoList[protos[i.D]]
			newClosure(&pc, uint8(i.A), towrap, code, &stack, co, &openUpvals, upvals)
			pc++
		case 20: // NAMECALL
			pc++
			if err = namecall(&pc, &top, &i, code, lineInfo, &stack, co, &op); err != nil {
				return
			}
		case 21: // CALL
			if err = call(&top, i.A, i.B, i.C, towrap, &stack, co); err != nil {
				return
			}
			pc++
		case 22: // RETURN
			b := int32(i.B) - 1

			// nresults
			if b == luau_multret {
				b = top - i.A
			}

			// execute() should pretty much always exit through here
			return stack[i.A:max(i.A+b, 0)], nil
		case 23, 24: // JUMP, JUMPBACK
			pc += i.D + 1
		case 25: // JUMPIF
			if falsy(stack[i.A]) {
				pc++
			} else {
				pc += i.D + 1
			}
		case 26: // JUMPIFNOT
			if falsy(stack[i.A]) {
				pc += i.D + 1
			} else {
				pc++
			}
		case 27: // jump
			if stack[i.A] == stack[i.Aux] {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 28:
			if j, err := jumpLe(stack[i.A], stack[i.Aux]); err != nil {
				return nil, err
			} else if j {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 29:
			if j, err := jumpLt(stack[i.A], stack[i.Aux]); err != nil {
				return nil, err
			} else if j {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 30:
			if stack[i.A] != stack[i.Aux] {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 31:
			if j, err := jumpGt(stack[i.A], stack[i.Aux]); err != nil {
				return nil, err
			} else if j {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 32:
			if j, err := jumpGe(stack[i.A], stack[i.Aux]); err != nil {
				return nil, err
			} else if j {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 33: // arithmetic
			if stack[i.A], err = aAdd(stack[i.B], stack[i.C]); err != nil {
				return
			}
			pc++
		case 34:
			if stack[i.A], err = aSub(stack[i.B], stack[i.C]); err != nil {
				return
			}
			pc++
		case 35:
			if stack[i.A], err = aMul(stack[i.B], stack[i.C]); err != nil {
				return
			}
			pc++
		case 36:
			if stack[i.A], err = aDiv(stack[i.B], stack[i.C]); err != nil {
				return
			}
			pc++
		case 37:
			if stack[i.A], err = aMod(stack[i.B], stack[i.C]); err != nil {
				return
			}
			pc++
		case 38:
			if stack[i.A], err = aPow(stack[i.B], stack[i.C]); err != nil {
				return
			}
			pc++
		case 81:
			if stack[i.A], err = aIdiv(stack[i.B], stack[i.C]); err != nil {
				return
			}
			pc++
		case 39: // arithmetik
			if stack[i.A], err = aAdd(stack[i.B], i.K); err != nil {
				return
			}
			pc++
		case 40:
			if stack[i.A], err = aSub(stack[i.B], i.K); err != nil {
				return
			}
			pc++
		case 41:
			if stack[i.A], err = aMul(stack[i.B], i.K); err != nil {
				return
			}
			pc++
		case 42:
			if stack[i.A], err = aDiv(stack[i.B], i.K); err != nil {
				return
			}
			pc++
		case 43:
			if stack[i.A], err = aMod(stack[i.B], i.K); err != nil {
				return
			}
			pc++
		case 44:
			if stack[i.A], err = aPow(stack[i.B], i.K); err != nil {
				return
			}
			pc++
		case 82:
			if stack[i.A], err = aIdiv(stack[i.B], i.K); err != nil {
				return
			}
			pc++
		case 71: // SUBRK
			if stack[i.A], err = aSub(i.K, stack[i.C]); err != nil {
				return
			}
			pc++
		case 72: // DIVRK
			if stack[i.A], err = aDiv(i.K, stack[i.C]); err != nil {
				return
			}
			pc++
		case 45: // logic AND
			if a, b := stack[i.B], stack[i.C]; falsy(a) {
				stack[i.A] = a
			} else {
				stack[i.A] = b
			}
			pc++
		case 46: // logic OR
			if a, b := stack[i.B], stack[i.C]; falsy(a) {
				stack[i.A] = b
			} else {
				stack[i.A] = a
			}
			pc++
		case 47: // logik AND
			// fmt.Println("LOGIK")
			if a, b := stack[i.B], i.K; falsy(a) {
				stack[i.A] = a
			} else {
				stack[i.A] = b
			}
			pc++
		case 48: // logik OR
			// fmt.Println("LOGIK")
			if a, b := stack[i.B], i.K; falsy(a) {
				stack[i.A] = b
			} else {
				stack[i.A] = a
			}
			pc++
		case 49: // CONCAT
			var b strings.Builder
			for first, n := uint8(0), i.B; n <= i.C; n, first = n+1, 1 {
				toWrite, ok := stack[n].(string)
				if !ok {
					// ensure correct order of operands in error message
					return nil, invalidConcat(TypeOf(stack[n-first]), TypeOf(stack[n+1-first]))
				}
				b.WriteString(toWrite)
			}
			stack[i.A] = b.String()
			pc++
		case 50: // NOT
			stack[i.A] = falsy(stack[i.B])
			pc++
		case 51: // MINUS
			if stack[i.A], err = aUnm(stack[i.B]); err != nil {
				return
			}
			pc++
		case 52: // LENGTH
			switch t := stack[i.B].(type) {
			case *Table:
				stack[i.A] = float64(t.Len())
			case string:
				stack[i.A] = float64(len(t))
			default:
				return nil, invalidLength(TypeOf(t))
			}
			pc++
		case 53: // NEWTABLE
			stack[i.A] = &Table{}
			pc += 2 // -- adjust for aux
		case 54: // DUPTABLE
			stack[i.A] = &Table{} // doesn't really apply here...
			pc++
		case 55: // SETLIST
			B := int32(i.B)
			c := int32(i.C) - 1

			if c == luau_multret {
				c = top - B
			}

			s := stack[i.A].(*Table)
			if s.Readonly {
				return nil, errReadonly
			}

			// one-indexed lol
			for n, v := range stack[i.B:min(B+c, int32(len(stack)))] {
				s.SetInt(n+int(i.Aux), v)
			}
			// stack[A] = s // in-place

			pc += 2 // -- adjust for aux
		case 56: // FORNPREP
			init, ok := stack[i.A+2].(float64)
			if !ok {
				return nil, invalidFor("initial value", TypeOf(stack[i.A+2]))
			}

			limit, ok := stack[i.A].(float64)
			if !ok {
				return nil, invalidFor("limit", TypeOf(stack[i.A]))
			}

			step, ok := stack[i.A+1].(float64)
			if !ok {
				return nil, invalidFor("step", TypeOf(stack[i.A+1]))
			}

			if s := step > 0; s && init > limit || !s && init < limit {
				pc += i.D + 1
			} else {
				pc++
			}
		case 57: // FORNLOOP
			// all checked in FORNPREP
			init := stack[i.A+2].(float64)
			limit := stack[i.A].(float64)
			step := stack[i.A+1].(float64)

			init += step
			stack[i.A+2] = init

			if s := step > 0; s && init <= limit || !s && init >= limit {
				pc += i.D + 1
			} else {
				pc++
			}
		case 58: // FORGLOOP
			if genIters == nil {
				genIters = map[internal.Inst][]Val{}
			}
			if err = forgloop(&pc, &top, i, &stack, co, genIters); err != nil {
				return
			}
		case 59, 61: // FORGPREP_INEXT, FORGPREP_NEXT
			if _, ok := stack[i.A].(Function); !ok {
				return nil, invalidIter(TypeOf(stack[i.A])) // -- encountered non-function value
			}
			pc += i.D + 1
		case 60: // FASTCALL3
			// Skipped
			pc += 2 // adjust for aux
		case 63: // GETVARARGS
			b := int32(i.B) - 1

			// fmt.Println("MULTRET", b, vargsLen)
			if b == luau_multret {
				b = int32(len(vargsList))
				top = i.A + b
			}

			// stack may get expanded here
			// (MAX STACK SIZE IS A LIE!!!!!!!!!!!!!!!!!!!!!!!)
			moveStack(&stack, vargsList, b, i.A)
			pc++
		case 64: // DUPCLOSURE
			// wrap is reused for closures
			towrap.proto = towrap.protoList[i.K.(uint32)] // 6 closure
			newClosure(&pc, uint8(i.A), towrap, code, &stack, co, nil, upvals)
			pc++
		case 65: // PREPVARARGS
			// Handled by wrapper
			pc++
		case 66: // LOADKX
			stack[i.A] = i.K
			pc += 2 // -- adjust for aux
		case 67: // JUMPX
			pc += i.A + 1 // lmfao
		case 68, 73: // FASTCALL, FASTCALL1
			// Skipped
			pc++
		case 74, 75: // FASTCALL2, FASTCALL2K
			// Skipped
			pc += 2 // adjust for aux
		case 76: // FORGPREP
			pc += i.D + 1 // what are we even supposed to do here, there's nothing to prepare
		case 77: // JUMPXEQKNIL
			if ra := stack[i.A]; ra == nil != i.KN {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 78, 79, 80: // JUMPXEQKB, JUMPXEQKN, JUMPXEQKS
			// actually the same apart from types (which aren't even correct anyway)
			if kv, ra := i.K, stack[i.A]; ra == kv != i.KN {
				pc += i.D + 1
			} else {
				pc += 2
			}
		default:
			return nil, fmt.Errorf("unsupported opcode: %d", op)
		}
	}

	return nil, errors.New("program execution cancelled")
}

func wrapclosure(towrap toWrap, existingCo *Coroutine) Function {
	proto := towrap.proto
	maxs, np := proto.MaxStackSize, proto.NumParams // maxs 2 lel

	return fn("", existingCo, func(co *Coroutine, args ...Val) (r []Val, err error) {
		la := uint8(len(args)) // we can't have more than 255 args anyway right?

		var list []Val
		if np < la {
			list = args[np:]
		}

		// fmt.Println("MAX STACK SIZE", maxs)
		stack := make([]Val, max(maxs, la-np)) // at least not have to resize *as* much when getting vargs
		copy(stack, args[:min(np, la)])

		// prevent line mismatches (error/loc.luau)
		initDbg := co.Dbg
		defer func() {
			co.Dbg = initDbg
		}()

		// fmt.Println("starting in coroutine", co.Dbgpath, "with args", args)
		// fmt.Println("started on", co.Dbg.Line)
		r, err = execute(towrap, stack, list, co)
		// fmt.Println("ended on", co.Dbg.Line)
		if !*towrap.alive {
			return
		}
		if err != nil {
			// fmt.Println("BAM!  error coroutine", co.Dbgpath, "with args", args)
			return nil, &internal.CoError{
				Line:    co.Dbg.Line,
				Dbgname: co.Dbg.Name,
				Path:    co.Dbgpath,
				Sub:     err,
			}
		}

		return
	})
}

func loadmodule(m compiled, env Env, requireCache map[string]Val, args ProgramArgs) (co Coroutine, cancel func()) {
	alive := true

	towrap := toWrap{
		proto:        m.MainProto,
		protoList:    m.ProtoList,
		alive:        &alive,
		env:          env,
		requireCache: requireCache,
	}

	return Coroutine{
		Body:           wrapclosure(towrap, nil),
		Env:            env,
		Filepath:       m.Filepath,
		Dbgpath:        m.Dbgpath,
		RequireHistory: m.RequireHistory,
		YieldChan:      make(chan internal.Yield, 1),
		ResumeChan:     make(chan []Val, 1),
		Compiler:       m.Compiler,
		ProgramArgs:    args,
	}, func() { alive = false }
}
