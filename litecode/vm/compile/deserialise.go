package compile

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/Heliodex/coputer/litecode/internal"
	. "github.com/Heliodex/coputer/litecode/types"
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

func checkkmode(i *internal.Inst, k []Val) {
	// fmt.Println("AUX:", i.Aux, "KMODE:", i.KMode)
	switch aux := i.Aux; i.KMode {
	case 1: // AUX
		if aux < uint32(len(k)) { // sometimes huge for some reason
			i.K = k[aux]
		}
	case 2: // C
		i.K = k[i.C]
		// fmt.Println("SET K TO", i.K, "FROM", i.C)
	case 3: // D
		i.K = k[i.D]
	case 4: // AUX import
		count := uint8(aux >> 30)
		i.KC = count

		id0 := aux >> 20 & 0x3ff
		i.K0 = k[id0].(string) // lmk if this panics lol
		// fmt.Println("AUX", i.K0)

		if count < 2 {
			break
		}
		id1 := aux >> 10 & 0x3ff
		i.K1 = k[id1].(string)

		if count < 3 { // should never be >3
			break
		}
		id2 := aux & 0x3ff
		i.K2 = k[id2].(string)
	case 5: // AUX boolean low 1 bit
		i.K = aux&1 == 1
		i.KN = aux>>31 == 1
	case 6: // AUX number low 24 bits
		i.K = k[aux&(1<<24-1)]
		i.KN = aux>>31 == 1
	case 7: // B
		i.K = k[i.B]
	case 8: // AUX number low 16 bits ig
		i.K = uint8(aux & 0xf) // forgloop
	}

	// fmt.Println("K:", i.K)
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
	// fmt.Println("String size:", size)
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
		if i.A = int32(value >> 8); i.A >= 0x800_000 { // why no arbitrary width integers, go
			i.A -= 0x1000_000
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
	// fmt.Println("Opcode:", opcode, "A:", i.A, "B:", i.B, "C:", i.C, "D:", i.D)

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
		// fmt.Println("i:", i, "v:", v)
		// fmt.Println("linegaplog2:", linegaplog2, "abslineinfo:", abslineinfo)
		instLineInfo[i] = abslineinfo[i>>linegaplog2] + uint32(v)
	}
	return
}

func (s *stream) skipDebugInfo() {
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
	// fmt.Println("Reading proto...", s.data[s.pos:])
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
	// fmt.Println("Sizecode:", sizecode)
	for i := uint32(0); i < sizecode; i++ {
		if s.readInst(&p.Code) {
			// fmt.Println("Insts (aux)  ", *p.Code[i-1], *p.Code[i])
			i++
		}
	}

	sizek := s.rVarInt()
	K := make([]Val, sizek) // krazy

	for i := range sizek {
		// fmt.Println("Ktype:", kt)
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

	// fmt.Println("K:", K)

	// -- 2nd pass to replace constant references in the instruction
	for i := range sizecode {
		// fmt.Println("Checking inst - aux:", p.Code[i].Aux, "kmode:", p.Code[i].KMode)
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
		s.skipDebugInfo()
	}
	return
}

func Deserialise(b []byte) (d internal.Deserialised, err error) {
	s := &stream{data: b}

	if luauVersion := s.rByte(); luauVersion == 0 {
		return internal.Deserialised{}, errors.New("the provided bytecode is an error message")
	} else if luauVersion != 6 {
		return internal.Deserialised{}, errors.New("the version of the provided bytecode is unsupported")
	}
	if s.rByte() != 3 { // types version
		return internal.Deserialised{}, errors.New("the types version of the provided bytecode is unsupported")
	}

	// fmt.Println("Rest:", s.data[s.pos:])

	stringCount := s.rVarInt()
	// fmt.Println("String count:", stringCount)
	stringList := make([]string, stringCount)
	for i := range stringCount {
		stringList[i] = s.rString()
		// fmt.Println("String:", stringList[i])
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

	// fmt.Println("Rest:", s.data[s.pos:])

	protoCount := s.rVarInt()
	// fmt.Println("Rest:", s.data[s.pos:])

	protoList := make([]*internal.Proto, protoCount)
	for i := range protoCount {
		if protoList[i], err = s.readProto(stringList); err != nil {
			return
		}
		// fmt.Println("Proto:", protoList[i])
	}

	// fmt.Println("Rest:", s.data[s.pos:])

	mainProto := protoList[s.rVarInt()]
	mainProto.Dbgname = "(main)"

	return internal.Deserialised{
		MainProto: mainProto,
		ProtoList: protoList,
	}, s.checkEnd()
}
