// package vm implements Litecode's deterministic Luau virtual machine and standard library toolkit.
package vm

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"unsafe"

	"github.com/Heliodex/coputer/litecode/internal"
	"github.com/Heliodex/coputer/litecode/types"
)

const safe = false

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
		YieldChan:  make(chan internal.Yield, 1),
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

func checkkmode(i *internal.Inst, k []types.Val) {
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
func readInst(codeList *[]*internal.Inst, s *stream) bool {
	value := s.rWord()

	opcode := uint8(value)
	opinfo := opList[opcode]

	i := internal.Inst{
		OpInfo: opinfo,
		Opcode: opcode,
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

		*codeList = append(*codeList, &internal.Inst{})
	}

	return opinfo.HasAux
}

func readProto(stringList []string, s *stream) (p *internal.Proto, err error) {
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
		// yeah
		case 1: // Bool
			p.K[i] = s.rBool()
		case 2: // Number
			p.K[i] = s.rFloat64()
		case 3: // String
			p.K[i] = stringList[s.rVarInt()-1]
		case 4: // Import
			// see resolveImport"Safe" in ref impl
			p.K[i] = s.rWord() // ⚠️ strange, TODO need something to test this ⚠️
			// fmt.Println("case 4", p.K[i])
		case 5: // Table
			// moot, whatever
			for range s.rVarInt() {
				s.skipVarInt()
			}

			// fmt.Println("case 5", p.K[i])
		case 6: // Closure
			// pain in the cranium
			p.K[i] = s.rVarInt() // ⚠️ not a val ⚠️
			// fmt.Println("case 6", p.K[i])
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

func deserialise(b []byte) (des internal.Deserialised, err error) {
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
		if protoList[i], err = readProto(stringList, s); err != nil {
			return
		}
	}

	mainProto := protoList[s.rVarInt()]
	mainProto.Dbgname = "(main)"

	return internal.Deserialised{
		MainProto: mainProto,
		ProtoList: protoList,
	}, s.CheckEnd()
}

type iterator struct {
	args    chan *types.Table
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
func TypeOf(v types.Val) string {
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
	case *types.Table:
		return "table"
	case types.Function:
		return "function"
	case *types.Coroutine:
		return "thread"
	case *types.Buffer:
		return "buffer"
	case types.Vector:
		return "vector"
	}
	return "userdata"
}

func aAdd(a, b types.Val) (types.Val, error) {
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

	return nil, invalidArithmetic("add", TypeOf(a), TypeOf(b))
}

func aSub(a, b types.Val) (types.Val, error) {
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

	return nil, invalidArithmetic("sub", TypeOf(a), TypeOf(b))
}

func aMul(a, b types.Val) (types.Val, error) {
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

	return nil, invalidArithmetic("mul", TypeOf(a), TypeOf(b))
}

func aDiv(a, b types.Val) (types.Val, error) {
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

	return nil, invalidArithmetic("div", TypeOf(a), TypeOf(b))
}

func aMod(a, b types.Val) (types.Val, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa - fb*math.Floor(fa/fb), nil
	}

	return nil, invalidArithmetic("mod", TypeOf(a), TypeOf(b))
}

func aPow(a, b types.Val) (types.Val, error) {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return math.Pow(fa, fb), nil
	}

	return nil, invalidArithmetic("pow", TypeOf(a), TypeOf(b))
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

func gettable(k, v types.Val) (types.Val, error) {
	switch t := v.(type) {
	case *types.Table:
		return t.Get(k), nil
	case types.Vector: // direction,,, and mmmagnitude!! oh yeah!!11!!
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

func getImport(i internal.Inst, towrap toWrap, stack *[]types.Val) (err error) {
	k0 := i.K0
	imp := exts[k0]
	if imp == nil {
		imp = towrap.env[k0]
	}

	count := i.KC

	if count < 2 {
		(*stack)[i.A] = imp
		return
	}

	t1, ok := imp.(*types.Table)
	if !ok {
		return invalidIndex("nil", i.K1)
	}

	imp = t1.GetHash(i.K1)
	// fmt.Println("GETIMPORT2", i.A, (*stack)[i.A])

	if count < 3 {
		(*stack)[i.A] = imp
		return
	}

	t2, ok := imp.(*types.Table)
	if !ok {
		return invalidIndex(TypeOf(imp), i.K2)
	}

	(*stack)[i.A] = t2.GetHash(i.K2)
	// fmt.Println("GETIMPORT3", i.A, (*stack)[i.A])

	return
}

func newClosure(pc *int32, i internal.Inst, towrap toWrap, p *internal.Proto, stack *[]types.Val, openUpvals *[]*upval, upvals []*upval) {
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
					store: *stack,
					index: idx,
				}

				for idx >= uint8(len(*openUpvals)) {
					*openUpvals = append(*openUpvals, nil)
				}
				(*openUpvals)[idx] = prev
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

func namecall(pc, top *int32, i *internal.Inst, p *internal.Proto, stack *[]types.Val, co *types.Coroutine, op *uint8) (err error) {
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
		case *types.Table:
			call := t.GetHash(kv)
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
	case *types.Table, types.Function:
	default:
		return nil, errors.New("module must return a table or function")
	}

	towrap.requireCache[lc.Filepath] = ret
	return []types.Val{ret}, nil
}

func call(top *int32, i internal.Inst, towrap toWrap, stack *[]types.Val, co *types.Coroutine) (err error) {
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
func forgloop(pc, top *int32, i internal.Inst, stack *[]types.Val, co *types.Coroutine, generalisedIterators *map[internal.Inst]*iterator) (err error) {
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
	case *types.Table:
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

func dupClosure(pc *int32, i internal.Inst, towrap toWrap, p *internal.Proto, stack *[]types.Val, upvals []*upval) {
	newProto := towrap.protoList[i.K.(uint32)] // 6 closure

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
	pc, top, openUpvals, generalisedIterators := int32(1), int32(-1), []*upval{}, map[internal.Inst]*iterator{}

	var handlingBreak bool
	var i internal.Inst
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
			idx := (*stack)[i.C]
			t, ok := (*stack)[i.B].(*types.Table) // SETTABLE or SETTABLEKS on a types.Vector actually does return "attempt to index vector with 'whatever'"
			if !ok {
				return nil, invalidIndex(TypeOf((*stack)[i.B]), idx)
			}
			if t.Readonly {
				return nil, errReadonly
			}

			// fmt.Println("SETTABLE", idx, (*stack)[i.A])
			t.Set(idx, (*stack)[i.A])
			pc++
		case 15: // GETTABLEKS
			if (*stack)[i.A], err = gettable(i.K, (*stack)[i.B]); err != nil {
				return nil, err
			}
			pc += 2 // -- adjust for aux
		case 16: // SETTABLEKS
			idx := i.K
			t, ok := (*stack)[i.B].(*types.Table)
			if !ok {
				// fmt.Println("indexing", typeOf((*stack)[i.B]), "with", idx)
				return nil, invalidIndex(TypeOf((*stack)[i.B]), idx)
			}
			if t.Readonly {
				return nil, errReadonly
			}

			t.Set(idx, (*stack)[i.A])
			pc += 2 // -- adjust for aux
		case 17: // GETTABLEN
			idx := float64(i.C + 1)
			t, ok := (*stack)[i.B].(*types.Table)
			if !ok {
				// fmt.Println("gettableninvalidindex")
				return nil, invalidIndex(TypeOf((*stack)[i.B]), idx)
			}

			(*stack)[i.A] = t.Get(idx)
			pc++
		case 18: // SETTABLEN
			idx := int(i.C) + 1
			t, ok := (*stack)[i.B].(*types.Table)
			if !ok {
				// fmt.Println("gettableninvalidindex")
				return nil, invalidIndex(TypeOf((*stack)[i.B]), float64(idx))
			}
			if t.Readonly {
				return nil, errReadonly
			}

			// fmt.Println("SETTABLEN", i.C+1, (*stack)[i.A])
			t.SetInt(idx, (*stack)[i.A])
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
			var b strings.Builder
			for first, n := uint8(0), i.B; n <= i.C; n++ {
				toWrite, ok := (*stack)[n].(string)
				if !ok {
					// ensure correct order of operands in error message
					return nil, invalidConcat(TypeOf((*stack)[n-first]), TypeOf((*stack)[n+1-first]))
				}
				b.WriteString(toWrite)
				first = 1
			}
			(*stack)[i.A] = b.String()
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
			case *types.Table:
				(*stack)[i.A] = float64(t.Len())
			case string:
				(*stack)[i.A] = float64(len(t))
			default:
				return nil, invalidLength(TypeOf(t))
			}
			pc++
		case 53: // NEWTABLE
			(*stack)[i.A] = &types.Table{}
			pc += 2 // -- adjust for aux
		case 54: // DUPTABLE
			(*stack)[i.A] = &types.Table{} // doesn't really apply here...
			pc++
		case 55: // SETLIST
			A, B := i.A, int32(i.B)
			c := int32(i.C) - 1

			if c == luau_multret {
				c = top - B + 1
			}

			s := (*stack)[A].(*types.Table)
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

			idx, ok := (*stack)[A+2].(float64)
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
				if idx > limit {
					pc += i.D
				}
			} else if idx < limit {
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
			// THIS OPCODE NEVER EVEN FUCKING RUNS
			(*stack)[i.A] = i.K

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
				args:   make(chan *types.Table),
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

		var list []types.Val
		if np < la {
			list = args[np:]
		}

		// fmt.Println("MAX STACK SIZE", maxs)
		stack := make([]types.Val, max(maxs, la-np)) // at least not have to resize *as* much when getting vargs
		copy(stack, args[:min(np, la)])

		// prevent line mismatches (error/loc.luau)
		initDbg := co.Dbg
		defer func() {
			co.Dbg = initDbg
		}()

		// fmt.Println("started on", co.Dbg.Line)
		r, err = execute(towrap, &stack, co, list, max(la-np, 0))
		// fmt.Println("ended on", co.Dbg.Line)
		if !*towrap.alive {
			return
		}
		if err != nil {
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
		YieldChan:      make(chan internal.Yield, 1),
		ResumeChan:     make(chan []types.Val, 1),
		Compiler:       m.Compiler,
		ProgramArgs:    args,
	}, func() { alive = false }
}
