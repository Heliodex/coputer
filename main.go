package main

import (
	"fmt"
	"math"
	"reflect"
)

func move(src []any, a, b, t int, dst *[]any) {
	if b < a {
		return
	}

	ret := make([]any, len(*dst))

	// add up until t
	s1 := (*dst)[:min(t, len(*dst))]
	for i, v := range s1 {
		ret[i] = v
	}

	s2 := src[a:min(b, len(src))]
	for i, v := range s2 {
		ret[i+t] = v
	}

	tl := len(s1) + len(s2)
	if tl < len(*dst) {
		for i, v := range (*dst)[t+b:] {
			ret[i+tl] = v
		}
	}

	*dst = ret
}

// ???
func moveMap(src []any, a, b, t int, dst *map[any]any) {
	if b < a {
		return
	}

	for i, v := range src[a:min(b, len(src))] {
		(*dst)[i+t] = v
	}
}

func ttisnumber(v any) bool {
	return reflect.TypeOf(v).Kind() == reflect.Float64
}

func ttisboolean(v any) bool {
	return reflect.TypeOf(v).Kind() == reflect.Bool
}

func ttisfunction(v any) bool {
	return reflect.TypeOf(v).Kind() == reflect.Func
}

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

type Operator struct {
	Name        string
	Mode, KMode uint8
	HasAux      bool
}

var opList = []Operator{
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

const LUAU_MULTRET = -1

type LuauSettings struct {
	VectorCtor       func(...float32) any
	VectorSize       uint8
	Extensions       map[any]any
	AllowProxyErrors bool
	DecodeOp         func(op uint32) uint32
}

var luau_settings = LuauSettings{
	VectorCtor: func(...float32) any {
		panic("vectorCtor was not provided")
	},
	VectorSize:       4,
	Extensions:       nil,
	AllowProxyErrors: false,
	DecodeOp: func(op uint32) uint32 {
		// println("decoding op", op)
		return op
	},
}

type Inst struct {
	A, B, C, D, E         int
	K, K0, K1, K2         any
	KC                    int
	KN                    bool
	aux                   int
	kmode, opcode, opmode uint8
	opname                string
	usesAux               bool
	value                 uint32
}

type Varargs struct {
	len  uint32
	list []any
}

type Proto struct {
	maxstacksize, numparams, nups uint8
	isvararg                      bool
	linedefined                   uint32
	debugname                     string

	sizecode  uint32
	code      []*Inst
	debugcode []uint8

	sizek uint32
	k     []any

	sizep  uint32
	protos []uint32

	lineinfoenabled     bool
	instructionlineinfo []uint32

	bytecodeid uint32
}

type Deserialise struct {
	stringList []string
	protoList  []Proto

	mainProto Proto

	typesVersion uint8
}

func luau_deserialise(stream []byte) Deserialise {
	// fmt.Println("deserialising")
	cursor := uint32(0)

	readByte := func() uint8 {
		b := stream[cursor]
		cursor += 1
		// fmt.Println("readByte", b, "at", cursor)
		return b
	}

	word := func() uint32 {
		return uint32(stream[cursor]) | uint32(stream[cursor+1])<<8 | uint32(stream[cursor+2])<<16 | uint32(stream[cursor+3])<<24
	}

	readWord := func() uint32 {
		w := word()
		cursor += 4
		// fmt.Println("readWord", w, "at", cursor)
		return w
	}

	readFloat := func() float32 {
		f := math.Float32frombits(word())
		cursor += 4
		// fmt.Println("readFloat", f, "at", cursor)
		return f
	}

	readDouble := func() float64 {
		word1 := word()
		cursor += 4
		word2 := word()
		cursor += 4

		d := math.Float64frombits(uint64(word1) | uint64(word2)<<32)
		// fmt.Println("readDouble", d, "at", cursor)
		return d
	}

	readVarInt := func() uint32 {
		result := uint32(0)

		for i := range 4 {
			value := readByte()
			result |= ((uint32(value) & 0x7F) << (i * 7))
			if value&0x80 == 0 {
				break
			}
		}

		return result
	}

	readString := func() string {
		size := readVarInt()

		if size == 0 {
			return ""
		}

		str := make([]byte, size)
		for i := range str {
			str[i] = stream[cursor+uint32(i)]
		}
		cursor += size

		return string(str)
	}

	luauVersion := readByte()
	typesVersion := uint8(0)
	if luauVersion == 0 {
		panic("the provided bytecode is an error message")
	} else if luauVersion < 3 || luauVersion > 6 {
		panic("the version of the provided bytecode is unsupported")
	} else if luauVersion >= 4 {
		typesVersion = readByte()
	}

	stringCount := readVarInt()
	stringList := make([]string, stringCount)

	for i := range stringList {
		stringList[i] = readString()
	}

	readInstruction := func(codeList *[]*Inst) bool {
		value := luau_settings.DecodeOp(readWord())
		opcode := uint8(value & 0xFF)

		opinfo := opList[opcode]
		opmode := opinfo.Mode
		usesAux := opinfo.HasAux

		inst := &Inst{
			opcode:  opcode,
			opname:  opinfo.Name,
			opmode:  opmode,
			kmode:   opinfo.KMode,
			usesAux: usesAux,
		}

		*codeList = append(*codeList, inst)

		switch opmode {
		case 1: /* A */
			inst.A = int(value>>8) & 0xFF
		case 2: /* AB */
			inst.A = int(value>>8) & 0xFF
			inst.B = int(value>>16) & 0xFF
		case 3: /* ABC */
			inst.A = int(value>>8) & 0xFF
			inst.B = int(value>>16) & 0xFF
			inst.C = int(value>>24) & 0xFF
		case 4: /* AD */
			inst.A = int(value>>8) & 0xFF
			temp := int(value>>16) & 0xFFFF

			// fmt.Println("Setting D to", temp)
			if temp < 0x8000 {
				inst.D = temp
			} else {
				inst.D = temp - 0x10000
			}
		case 5: /* AE */
			temp := int(value>>8) & 0xFFFFFF

			if temp < 0x800000 {
				inst.E = temp
			} else {
				inst.E = temp - 0x1000000
			}
		}

		if usesAux {
			aux := readWord()
			inst.aux = int(aux)

			*codeList = append(*codeList, &Inst{value: aux, opname: "auxvalue"})
		}

		return usesAux
	}

	checkkmode := func(inst *Inst, k []any) {
		switch inst.kmode {
		case 1: /* AUX */
			inst.K = k[inst.aux]
		case 2: /* C */
			inst.K = k[inst.C]
			fmt.Println("SET K TO", inst.K, "FROM", inst.C)
		case 3: /* D */
			inst.K = k[inst.D]
		case 4: /* AUX import */
			extend := inst.aux
			count := extend >> 30
			inst.KC = count

			id0 := (extend >> 20) & 0x3FF
			inst.K0 = k[id0]

			if count >= 2 {
				id1 := (extend >> 10) & 0x3FF
				inst.K1 = k[id1]
			}
			if count == 3 { // >=?
				id2 := extend & 0x3FF
				inst.K2 = k[id2]
			}
		case 5: /* AUX boolean low 1 bit */
			inst.K = extract(inst.aux, 0, 1) == 1
			inst.KN = extract(inst.aux, 31, 1) == 1
		case 6: /* AUX number low 24 bits */
			inst.K = k[int(extract(inst.aux, 0, 24))] // TODO: 1-based indexing
			inst.KN = extract(inst.aux, 31, 1) == 1
		case 7: /* B */
			inst.K = k[inst.B] // TODO: 1-based indexing
		case 8: /* AUX number low 16 bits */
			inst.K = inst.aux & 0xF
		}
	}

	readProto := func(bytecodeid uint32) Proto {
		maxstacksize := readByte()
		numparams := readByte()
		nups := readByte()
		isvararg := readByte() != 0

		if luauVersion >= 4 {
			readByte() //-- flags
			typesize := readVarInt()
			cursor += typesize
		}

		sizecode := readVarInt()
		codelist := new([]*Inst)

		skipnext := false
		for range sizecode {
			if skipnext {
				skipnext = false
				continue
			}

			skipnext = readInstruction(codelist)
		}

		debugcodelist := make([]uint8, sizecode)
		for i := range sizecode {
			debugcodelist[i] = (*codelist)[i].opcode
		}

		sizek := readVarInt()
		klist := make([]any, sizek)

		for i := range int(sizek) {
			kt := readByte()
			var k any

			// fmt.Println("ktype", kt)

			switch kt {
			case 0: /* Nil */
				k = nil
			case 1: /* Bool */
				k = readByte() != 0
			case 2: /* Number */
				k = readDouble()
			case 3: /* String */
				k = stringList[readVarInt()-1] // TODO: 1-based indexing
			case 4: /* Function */
				k = readWord()
			case 5: /* Table */
				dataLength := readVarInt()
				k = make([]uint32, dataLength)

				for i := range dataLength {
					k.([]any)[i] = readVarInt() // TODO: 1-based indexing
				}
			case 6: /* Closure */
				k = readVarInt()
			case 7: /* Vector */
				x, y, z, w := readFloat(), readFloat(), readFloat(), readFloat()

				if luau_settings.VectorSize == 4 {
					k = luau_settings.VectorCtor(x, y, z, w)
				} else {
					k = luau_settings.VectorCtor(x, y, z)
				}
			default:
				panic(fmt.Sprintf("Unknown ktype %d", kt))
			}

			klist[i] = k
		}

		// -- 2nd pass to replace constant references in the instruction
		// fmt.Println("klist", klist)
		for i := range sizecode {
			checkkmode((*codelist)[i], klist)
		}

		// fmt.Println("READING SIZEP")
		sizep := readVarInt()
		protos := make([]uint32, sizep)

		for i := range sizep {
			protos[i] = readVarInt() + 1 // TODO: 1-based indexing
		}

		linedefined := readVarInt()

		debugnameindex := readVarInt()
		var debugname string

		if debugnameindex == 0 {
			debugname = "(??)"
		} else {
			debugname = stringList[debugnameindex-1] // TODO: 1-based indexing
		}

		// -- lineinfo
		lineinfoenabled := readByte() != 0
		var instructionlineinfo []uint32

		if lineinfoenabled {
			linegaplog2 := readByte()

			intervals := uint32((sizecode-1)>>linegaplog2) + 1

			lineinfo := make([]uint32, sizecode)
			abslineinfo := make([]uint32, intervals)

			lastoffset := uint32(0)
			for i := range sizecode {
				lastoffset += uint32(readByte()) // TODO: type convs?
				lineinfo[i] = lastoffset
			}

			lastline := uint32(0)
			for i := range intervals {
				lastline += readWord()
				abslineinfo[i] = uint32(uint64(lastline) % (uint64(math.Pow(2, 32)))) // TODO: 1-based indexing
			}

			instructionlineinfo = make([]uint32, sizecode)

			for i := range sizecode {
				// -- p->abslineinfo[pc >> p->linegaplog2] + p->lineinfo[pc];
				instructionlineinfo = append(instructionlineinfo, abslineinfo[i>>linegaplog2]+lineinfo[i]) // TODO: 1-based indexing
			}
		}

		// -- debuginfo
		if readByte() != 0 {
			sizel := readVarInt()
			for range sizel {
				readVarInt()
				readVarInt()
				readVarInt()
				readByte()
			}
			sizeupvalues := readVarInt()
			for range sizeupvalues {
				readVarInt()
			}
		}

		return Proto{
			maxstacksize: maxstacksize,
			numparams:    numparams,
			nups:         nups,
			isvararg:     isvararg,
			linedefined:  linedefined,
			debugname:    debugname,

			sizecode:  sizecode,
			code:      *codelist,
			debugcode: debugcodelist,

			sizek: sizek,
			k:     klist,

			sizep:  sizep,
			protos: protos,

			lineinfoenabled:     lineinfoenabled,
			instructionlineinfo: instructionlineinfo,

			bytecodeid: bytecodeid,
		}
	}

	// userdataRemapping (not used in VM, left unused)
	if typesVersion == 3 {
		index := readByte()

		for index != 0 {
			readVarInt()

			index = readByte()
		}
	}

	protoCount := readVarInt()
	protoList := make([]Proto, protoCount)

	for i := range protoCount {
		protoList[i] = readProto(i - 1)
		// fmt.Println("read proto", protoList[i].nups)
	}

	mainProto := protoList[readVarInt()]

	if cursor != uint32(len(stream)) {
		panic("deserialiser cursor position mismatch")
	}

	mainProto.debugname = "(main)"

	return Deserialise{
		stringList: stringList,
		protoList:  protoList,

		mainProto: mainProto,

		typesVersion: typesVersion,
	}
}

type Iterator struct {
	running bool
	args    chan *[]any
	resume  chan *[]any
}

func luau_load(module Deserialise, env map[any]any) (func(...any) []any, func()) {
	protolist := module.protoList
	mainProto := module.mainProto

	alive := true
	luau_close := func() {
		alive = false
	}

	type Upval struct {
		value any
		index any
		store any
	}

	var luau_wrapclosure func(module Deserialise, proto Proto, upvals []Upval) func(...any) []any
	luau_wrapclosure = func(module Deserialise, proto Proto, upvals []Upval) func(...any) []any {
		luau_execute := func(
			stack *[]any,
			protos []uint32,
			code []*Inst,
			varargs Varargs,
		) []any {
			// if "pc" means "program counter" then this makes a lot more sense than I thought
			top, pc, open_upvalues, generalised_iterators := -1, 1, new([]*Upval), map[Inst]*Iterator{}
			constants := proto.k
			debugopcodes := proto.debugcode
			extensions := luau_settings.Extensions

			handlingBreak := false
			inst, op := Inst{}, uint8(0)

			// a a a a
			// stayin' alive
			for alive { // TODO: check go scope bruh
				if !handlingBreak {
					inst = *code[pc-1]
					op = inst.opcode
				}
				handlingBreak = false

				pc += 1

				fmt.Println("OP", op, "PC", pc)

				switch op {
				case 0: /* NOP */
					// -- Do nothing
				case 1: /* BREAK */
					pc -= 1
					op = debugopcodes[pc]
					handlingBreak = true
				case 2: /* LOADNIL */
					(*stack)[inst.A] = nil
				case 3: /* LOADB */
					(*stack)[inst.A] = inst.B == 1
					pc += inst.C
				case 4: /* LOADN */
					(*stack)[inst.A] = inst.D
				case 5: /* LOADK */
					// fmt.Println("LOADK", inst.K)
					(*stack)[inst.A] = inst.K
				case 6: /* MOVE */
					// we should never have to change the size of the stack (proto.maxstacksize)
					(*stack)[inst.A] = (*stack)[inst.B]
				case 7: /* GETGLOBAL */
					kv := inst.K

					(*stack)[inst.A] = extensions[kv]
					if (*stack)[inst.A] == nil {
						(*stack)[inst.A] = env[kv]
					}

					pc += 1 // -- adjust for aux
				case 8: /* SETGLOBAL */
					kv := inst.K
					env[kv] = (*stack)[inst.A]

					pc += 1 // -- adjust for aux
				case 9: /* GETUPVAL */
					uv := upvals[inst.B]
					(*stack)[inst.A] = (*uv.store.(*[]any))[uv.index.(int)]
				case 10: /* SETUPVAL */
					uv := upvals[inst.B]
					(*uv.store.(*[]any))[uv.index.(int)] = (*stack)[inst.A]
				case 11: /* CLOSEUPVALS */
					for i, uv := range *open_upvalues {
						if uv.index.(int) < inst.A {
							continue
						}
						uv.value = (*uv.store.(*[]any))[uv.index.(int)]
						uv.store = uv
						uv.index = "value" // -- self reference
						(*open_upvalues)[i] = nil
					}
				case 12: /* GETIMPORT */
					count := inst.KC
					k0 := inst.K0
					imp := extensions[k0]
					if imp == nil {
						imp = env[k0]
					}

					switch count {
					case 1:
						(*stack)[inst.A] = imp
					case 2:
						(*stack)[inst.A] = imp.([]any)[inst.K1.(uint32)-1]
					case 3:
						(*stack)[inst.A] = imp.([]any)[inst.K1.(uint32)-1].([]any)[inst.K2.(uint32)-1]
					}

					// for i, v := range *stack {
					// 	fmt.Printf("aa    [%d] = %v\n", i, v)
					// }

					pc += 1 // -- adjust for aux
				case 13: /* GETTABLE */
					(*stack)[inst.A] = (*stack)[inst.B].([]any)[(*stack)[inst.C].(uint32)]
				case 14: /* SETTABLE */
					(*stack)[inst.B].([]any)[(*stack)[inst.C].(uint32)] = (*stack)[inst.A]
				case 15: /* GETTABLEKS */
					index := inst.K
					(*stack)[inst.A] = (*stack)[inst.B].(map[any]any)[index]

					pc += 1 // -- adjust for aux
				case 16: /* SETTABLEKS */
					index := inst.K
					(*stack)[inst.B].(map[any]any)[index] = (*stack)[inst.A]

					pc += 1 // -- adjust for aux
				case 17: /* GETTABLEN */
					(*stack)[inst.A] = (*stack)[inst.B].(map[int]any)[inst.C+1]
				case 18: /* SETTABLEN */
					(*stack)[inst.B].(map[int]any)[inst.C] = (*stack)[inst.A]
				case 19: /* NEWCLOSURE */
					newPrototype := protolist[protos[inst.D]-1]

					nups := newPrototype.nups
					upvalues := make([]Upval, nups)
					(*stack)[inst.A] = luau_wrapclosure(module, newPrototype, upvalues)

					fmt.Println("nups", nups)
					for i := range nups {
						pseudo := code[pc-1]
						t := pseudo.A

						pc += 1

						if t == 0 { /* value */
							upvalue := Upval{
								value: (*stack)[pseudo.B],
								index: "value", // -- self reference
							}
							upvalue.store = upvalue

							upvalues[i] = upvalue
						} else if t == 1 { /* reference */
							index := pseudo.B
							fmt.Println("index", index, len(*open_upvalues))

							var prev *Upval
							if index < len(*open_upvalues) {
								prev = (*open_upvalues)[index]
							}

							if prev == nil {
								prev = &Upval{
									index: index,
									store: stack,
								}

								for len(*open_upvalues) <= index {
									*open_upvalues = append(*open_upvalues, nil)
								}
								(*open_upvalues)[index] = prev
							}

							upvalues[i] = *prev
						} else if t == 2 { /* upvalue */
							upvalues[i] = upvals[pseudo.B]
						}
					}
				case 20: /* NAMECALL */
					A, B := inst.A, inst.B
					kv := inst.K.(uint32)

					sb := (*stack)[B]
					(*stack)[A+1] = sb // TODO: 1-based indexing
					(*stack)[A] = sb.([]any)[kv]

					pc += 1 // -- adjust for aux
				case 21: /* CALL */
					A, B, C := inst.A, inst.B, inst.C

					var params int
					if B == 0 {
						params = top - A
					} else {
						params = B - 1
					}

					fmt.Println(A, (*stack)[A])
					if (*stack)[A] == nil {
						panic("attempt to call a nil value")
					}

					fn := (*stack)[A].(func(...any) []any)
					fmt.Println("calling with", (*stack)[A+1:A+params+1])

					ret_list := fn((*stack)[A+1 : A+params+1]...) // not inclusive
					ret_num := int(len(ret_list))

					if C == 0 {
						top = A + ret_num - 1
					} else {
						ret_num = C - 1
					}

					move(ret_list, 0, ret_num, A, stack)
				case 22: /* RETURN */
					A, B := inst.A, inst.B
					b := (B - 1)

					// nresults
					if b == LUAU_MULTRET {
						b = top - A + 1
					}

					return (*stack)[A:max(A+b, 0)]
				case 23: /* JUMP */
					pc += inst.D
				case 24: /* JUMPBACK */
					pc += inst.D
				case 25: /* JUMPIF */
					if (*stack)[inst.A] != nil {
						pc += inst.D
					}
				case 26: /* JUMPIFNOT */
					if (*stack)[inst.A] == nil {
						pc += inst.D
					}
				case 27: /* JUMPIFEQ */
					if (*stack)[inst.A] == (*stack)[inst.aux] {
						pc += inst.D
					} else {
						pc += 1
					}
				case 28: /* JUMPIFLE */
					if (*stack)[inst.A].(int) <= (*stack)[inst.aux].(int) {
						pc += inst.D
					} else {
						pc += 1
					}
				case 29: /* JUMPIFLT */
					if (*stack)[inst.A].(int) < (*stack)[inst.aux].(int) {
						pc += inst.D
					} else {
						pc += 1
					}
				case 30: /* JUMPIFNOTEQ */
					if (*stack)[inst.A] == (*stack)[inst.aux] {
						pc += 1
					} else {
						pc += inst.D
					}
				case 31: /* JUMPIFNOTLE */
					if (*stack)[inst.A].(int) <= (*stack)[inst.aux].(int) {
						pc += 1
					} else {
						pc += inst.D
					}
				case 32: /* JUMPIFNOTLT */
					if (*stack)[inst.A].(int) < (*stack)[inst.aux].(int) {
						pc += 1
					} else {
						pc += inst.D
					}
				case 33: /* ADD */
					fmt.Println("ADD", inst.A, inst.B, inst.C)
					for i, v := range *stack {
						fmt.Printf("    [%d] = %v\n", i, v)
					}

					(*stack)[inst.A] = (*stack)[inst.B].(int) + (*stack)[inst.C].(int)
				case 34: /* SUB */
					(*stack)[inst.A] = (*stack)[inst.B].(int) - (*stack)[inst.C].(int)
				case 35: /* MUL */
					(*stack)[inst.A] = (*stack)[inst.B].(int) * (*stack)[inst.C].(int)
				case 36: /* DIV */
					(*stack)[inst.A] = (*stack)[inst.B].(int) / (*stack)[inst.C].(int)
				case 37: /* MOD */
					(*stack)[inst.A] = math.Mod(float64((*stack)[inst.B].(int)), float64((*stack)[inst.C].(int)))
				case 38: /* POW */
					(*stack)[inst.A] = math.Pow(float64((*stack)[inst.B].(int)), float64((*stack)[inst.C].(int)))
				case 39: /* ADDK */
					(*stack)[inst.A] = (*stack)[inst.B].(int) + int(inst.K.(float64))
				case 40: /* SUBK */
					(*stack)[inst.A] = (*stack)[inst.B].(int) - int(inst.K.(float64))
				case 41: /* MULK */
					(*stack)[inst.A] = (*stack)[inst.B].(int) * int(inst.K.(float64))
					// fmt.Println("MULK", (*stack)[inst.B], inst.K)
				case 42: /* DIVK */
					(*stack)[inst.A] = (*stack)[inst.B].(int) / int(inst.K.(float64))
				case 43: /* MODK */
					(*stack)[inst.A] = math.Mod(float64((*stack)[inst.B].(int)), float64(inst.K.(int)))
				case 44: /* POWK */
					(*stack)[inst.A] = math.Pow(float64((*stack)[inst.B].(int)), float64(inst.K.(int)))
				case 45: /* AND */
					value := (*stack)[inst.B]
					if value != nil {
						(*stack)[inst.A] = (*stack)[inst.C]
						if (*stack)[inst.A] == nil {
							(*stack)[inst.A] = false
						}
					} else {
						(*stack)[inst.A] = value
					}
				case 46: /* OR */
					value := (*stack)[inst.B]
					if value != nil {
						(*stack)[inst.A] = value
					} else {
						(*stack)[inst.A] = (*stack)[inst.C]
						if (*stack)[inst.A] == nil {
							(*stack)[inst.A] = false
						}
					}
				case 47: /* ANDK */
					value := (*stack)[inst.B]
					if value != nil {
						(*stack)[inst.A] = inst.K
						if (*stack)[inst.A] == nil {
							(*stack)[inst.A] = false
						}
					} else {
						(*stack)[inst.A] = value
					}
				case 48: /* ORK */
					value := (*stack)[inst.B]
					if value != nil {
						(*stack)[inst.A] = value
					} else {
						(*stack)[inst.A] = inst.K
						if (*stack)[inst.A] == nil {
							(*stack)[inst.A] = false
						}
					}
				case 49: /* CONCAT */
					// TODO: optimise w/ stringbuilders
					s := ""
					for i := inst.B; i <= inst.C; i++ {
						s += (*stack)[i].(string)
					}
					(*stack)[inst.A] = s
				case 50: /* NOT */
					(*stack)[inst.A] = !(*stack)[inst.B].(bool)
				case 51: /* MINUS */
					(*stack)[inst.A] = -(*stack)[inst.B].(float64)
				case 52: /* LENGTH */
					(*stack)[inst.A] = len((*stack)[inst.B].([]any)) // TODO: 1-based indexing
				case 53: /* NEWTABLE */
					(*stack)[inst.A] = map[any]any{}

					pc += 1 // -- adjust for aux
				case 54: /* DUPTABLE */
					template := inst.K.([]int)
					serialised := make([]any, len(template))
					for _, id := range template {
						serialised[constants[id].(uint32)] = nil // TODO: 1-based indexing
					}
				case 55: /* SETLIST */
					A, B := inst.A, inst.B
					c := inst.C - 1

					if c == LUAU_MULTRET {
						c = top - B + 1
					}

					s := (*stack)[A].(map[any]any)

					// one-indexed lol
					moveMap(*stack, B, B+c, inst.aux, &s)
					(*stack)[A] = s

					pc += 1 // -- adjust for aux
				case 56: /* FORNPREP */
					A := inst.A

					limit := (*stack)[A].(int)
					if !ttisnumber(limit) {
						number := limit

						if number == 0 { // TODO: check nils
							panic("invalid 'for' limit (number expected)")
						}

						(*stack)[A] = number
						limit = number
					}

					step := (*stack)[A+1].(int)
					if !ttisnumber(step) {
						number := step

						if number == 0 { // TODO: check nils
							panic("invalid 'for' step (number expected)")
						}

						(*stack)[A+1] = number
						step = number
					}

					index := (*stack)[A+2].(int)
					if !ttisnumber(index) {
						number := index

						if number == 0 { // TODO: check nils
							panic("invalid 'for' index (number expected)")
						}

						(*stack)[A+2] = number
						index = number
					}

					if step > 0 {
						if index > limit {
							pc += inst.D
						}
					} else if limit > index {
						pc += inst.D
					}
				case 57: /* FORNLOOP */
					A := inst.A
					limit := (*stack)[A].(int)
					step := (*stack)[A+1].(int)
					index := (*stack)[A+2].(int) + step

					(*stack)[A+2] = index

					if step > 0 {
						if index <= limit {
							pc += inst.D
						}
					} else if limit <= index {
						pc += inst.D
					}
				case 58: /* FORGLOOP */
					A := inst.A
					res := inst.K.(int)

					top = int(A + 6)

					it := (*stack)[A]
					fmt.Println("IT", it, ttisfunction(it))

					if ttisfunction(it) {
						vals := it.(func(...any) []any)((*stack)[A+1], (*stack)[A+2])

						move(vals, 0, res, A+3, &(*stack))

						fmt.Println(A+3, (*stack)[A+3])

						if (*stack)[A+3] != nil {
							(*stack)[A+2] = (*stack)[A+3]
							pc += inst.D
						} else {
							pc += 1
						}
					} else {
						iter := *generalised_iterators[inst]

						if !iter.running {
							args := &[]any{it, (*stack)[A+1], (*stack)[A+2]}
							fmt.Println("-1- sending thru the wire", args)
							iter.args <- args
							fmt.Println("-1- sent")
						}
						vals := <-iter.resume
						fmt.Println("-1- received!", vals)

						if vals == nil {
							delete(generalised_iterators, inst)
							pc += 1
						} else {
							move(*vals, 0, res, A+3, stack)

							(*stack)[A+2] = (*stack)[A+3]
							pc += inst.D
						}
					}
				case 59: /* FORGPREP_INEXT */
					if !ttisfunction((*stack)[inst.A]) {
						// yaaaaaaaaaaay reflection (i'm dying inside)
						panic(fmt.Sprintf("attempt to iterate over a %s value", reflect.TypeOf((*stack)[inst.A]))) // -- FORGPREP_INEXT encountered non-function value
					}

					pc += inst.D
				case 60: /* FASTCALL3 */
					/* Skipped */
					pc += 1 // adjust for aux
				case 61: /* FORGPREP_NEXT */
					if !ttisfunction((*stack)[inst.A]) {
						panic(fmt.Sprintf("attempt to iterate over a %s value", reflect.TypeOf((*stack)[inst.A]))) // -- FORGPREP_NEXT encountered non-function value
					}

					pc += inst.D
				case 63: /* GETVARARGS */
					A := inst.A
					b := inst.B - 1

					if b == LUAU_MULTRET {
						fmt.Println("MULTRET4")
						b = int(varargs.len)
						top = A + b - 1
					}

					move(varargs.list, 0, b, A, stack)
				case 64: /* DUPCLOSURE */
					newPrototype := protolist[inst.K.(uint32)] // TODO: 1-based indexing

					nups := newPrototype.nups
					upvalues := make([]Upval, nups)
					(*stack)[inst.A] = luau_wrapclosure(module, newPrototype, upvalues)

					for i := range nups {
						pseudo := code[pc]
						pc += 1

						t := pseudo.A
						if t == 0 { /* value */
							upvalue := Upval{
								value: (*stack)[pseudo.B],
								index: "value", // -- self reference
							}
							upvalue.store = upvalue

							upvalues[i] = upvalue

							// -- references dont get handled by DUPCLOSURE
						} else if t == 2 { /* upvalue */
							upvalues[i] = upvals[pseudo.B]
						}
					}
				case 65: /* PREPVARARGS */
					/* Handled by wrapper */
				case 66: /* LOADKX */
					kv := inst.K.(uint32)
					(*stack)[inst.A] = kv

					pc += 1 // -- adjust for aux
				case 67: /* JUMPX */
					pc += inst.E
				case 68: /* FASTCALL */
					/* Skipped */
				case 69: /* COVERAGE */
					inst.E += 1
				case 70: /* CAPTURE */
					/* Handled by CLOSURE */
					panic("encountered unhandled CAPTURE")
				case 71: /* SUBRK */
					(*stack)[inst.A] = inst.K.(float64) - (*stack)[inst.C].(float64)
				case 72: /* DIVRK */
					(*stack)[inst.A] = inst.K.(float64) / (*stack)[inst.C].(float64)
				case 73: /* FASTCALL1 */
					/* Skipped */
				case 74: /* FASTCALL2 */
					/* Skipped */
					pc += 1 // adjust for aux
				case 75: /* FASTCALL2K */
					/* Skipped */
					pc += 1 // adjust for aux
				case 76: /* FORGPREP */
					// ohhh no
					iterator := (*stack)[inst.A]

					if !ttisfunction(iterator) {
						loopInstruction := *code[pc+inst.D-1]
						if generalised_iterators[loopInstruction] == nil {
							c := &Iterator{
								args:   make(chan *[]any),
								resume: make(chan *[]any),
							}

							go func() {
								args := *<-c.args
								c.args = nil // we're done here
								c.running = true
								fmt.Println("-2- generating iterator", args)

								for i, v := range args[0].(map[any]any) {
									if !c.running {
										return
									}
									fmt.Println("-2- yielding", i, v)
									c.resume <- &[]any{i, v}
									fmt.Println("-2- yielded!")
								}

								c.resume <- nil
							}()

							generalised_iterators[loopInstruction] = c
						}
					}

					pc += inst.D
				case 77: /* JUMPXEQKNIL */
					kn := inst.KN

					if ((*stack)[inst.A] == nil) != kn {
						pc += inst.D
					} else {
						pc += 1
					}
				case 78: /* JUMPXEQKB */
					kv := inst.K.(bool)
					kn := inst.KN
					ra := (*stack)[inst.A]

					if ttisboolean(ra) && (ra.(bool) == kv) != kn {
						pc += inst.D
					} else {
						pc += 1
					}
				case 79: /* JUMPXEQKN */
					kv := inst.K.(uint32)
					kn := inst.KN
					ra := (*stack)[inst.A].(uint32)

					if (ra == kv) != kn {
						pc += inst.D
					} else {
						pc += 1
					}
				case 80: /* JUMPXEQKS */
					kv := inst.K.(uint32)
					kn := inst.KN
					ra := (*stack)[inst.A].(uint32)

					if (ra == kv) != kn {
						pc += inst.D
					} else {
						pc += 1
					}
				case 81: /* IDIV */
					(*stack)[inst.A] = (*stack)[inst.B].(uint32) / (*stack)[inst.C].(uint32)
				case 82: /* IDIVK */
					(*stack)[inst.A] = (*stack)[inst.B].(uint32) / inst.K.(uint32)
				default:
					panic(fmt.Sprintf("Unsupported Opcode: %s op: %d", inst.opname, op))
				}
			}

			for i, uv := range *open_upvalues {
				uv.value = (*uv.store.(*[]any))[uv.index.(int)]
				uv.store = uv
				uv.index = "value" // -- self reference
				(*open_upvalues)[i] = nil
			}

			for i := range generalised_iterators {
				generalised_iterators[i].running = false
				delete(generalised_iterators, i)
			}
			return []any{}
		}

		wrapped := func(passed ...any) []any {
			stack := make([]any, proto.maxstacksize)
			fmt.Println("MAX STACK SIZE", proto.maxstacksize)
			varargs := Varargs{
				len:  0,
				list: []any{},
			}

			// TODO: test table.move impl
			move(passed, 0, int(proto.numparams), 0, &stack)

			n := uint8(len(passed))
			if proto.numparams < n {
				start := proto.numparams + 1
				l := n - proto.numparams
				varargs.len = uint32(l)

				// expand varargs list
				varargs.list = make([]any, l)

				move(passed, int(start)-1, int(start+l)-2, 0, &varargs.list)
			}

			// TODO: dee bugg ingg
			result := luau_execute(&stack, proto.protos, proto.code, varargs)
			fmt.Println("RESULT", result)

			return result
		}

		// fmt.Println("wrapping closure")
		return wrapped
	}

	return luau_wrapclosure(module, mainProto, []Upval{}), luau_close
}

var bytecode = []byte{6, 3, 10, 3, 97, 100, 100, 5, 112, 114, 105, 110, 116, 13, 72, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100, 33, 33, 5, 104, 101, 108, 108, 111, 5, 119, 111, 114, 108, 100, 1, 33, 1, 120, 1, 121, 5, 101, 114, 114, 111, 114, 16, 84, 104, 105, 115, 32, 105, 115, 32, 97, 110, 32, 101, 114, 114, 111, 114, 0, 2, 3, 2, 0, 0, 0, 0, 2, 33, 2, 0, 1, 22, 2, 2, 0, 0, 0, 21, 1, 1, 24, 0, 0, 22, 0, 0, 0, 0, 10, 0, 0, 1, 0, 0, 66, 65, 0, 0, 0, 12, 0, 1, 0, 0, 0, 0, 64, 5, 1, 2, 0, 21, 0, 2, 1, 4, 0, 1, 0, 4, 3, 1, 0, 4, 1, 10, 0, 4, 2, 1, 0, 56, 1, 6, 0, 12, 4, 1, 0, 0, 0, 0, 64, 6, 5, 3, 0, 21, 4, 2, 1, 35, 0, 0, 3, 57, 1, 250, 255, 12, 1, 1, 0, 0, 0, 0, 64, 6, 2, 0, 0, 21, 1, 2, 1, 53, 1, 0, 0, 3, 0, 0, 0, 5, 2, 3, 0, 5, 3, 4, 0, 5, 4, 5, 0, 55, 1, 2, 4, 1, 0, 0, 0, 6, 2, 1, 0, 2, 3, 0, 0, 2, 4, 0, 0, 76, 2, 5, 0, 12, 7, 1, 0, 0, 0, 0, 64, 6, 8, 5, 0, 6, 9, 6, 0, 21, 7, 3, 1, 58, 2, 250, 255, 2, 0, 0, 0, 64, 2, 6, 0, 12, 3, 1, 0, 0, 0, 0, 64, 6, 4, 2, 0, 4, 5, 1, 0, 4, 6, 2, 0, 21, 4, 3, 0, 21, 3, 0, 1, 53, 3, 2, 0, 0, 0, 0, 0, 4, 4, 1, 0, 16, 4, 3, 153, 7, 0, 0, 0, 4, 4, 2, 0, 16, 4, 3, 152, 8, 0, 0, 0, 12, 4, 1, 0, 0, 0, 0, 64, 15, 5, 3, 153, 7, 0, 0, 0, 15, 6, 3, 152, 8, 0, 0, 0, 21, 4, 3, 1, 12, 4, 10, 0, 0, 0, 144, 64, 5, 5, 11, 0, 21, 4, 2, 1, 22, 0, 1, 0, 12, 3, 2, 4, 0, 0, 0, 64, 3, 3, 3, 4, 3, 5, 3, 6, 6, 0, 3, 7, 3, 8, 3, 9, 4, 0, 0, 144, 64, 3, 10, 1, 0, 1, 0, 1, 24, 0, 1, 0, 0, 0, 3, 1, 0, 0, 0, 1, 0, 0, 0, 1, 254, 5, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 1, 0, 0, 0, 0, 255, 0, 5, 4, 0, 0, 0, 0, 0, 0, 3, 0, 2, 0, 0, 1, 0, 0, 2, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 1, 1, 0, 0, 0, 0, 1}

func main() {
	deserialised := luau_deserialise(bytecode)

	exec, _ := luau_load(deserialised, map[any]any{
		"print": func(args ...any) (ret []any) {
			args = append([]any{"printed:"}, args...)
			fmt.Println(args...)
			return
		},
		"error": func(args ...any) (ret []any) {
			panic(args[0])
		},
	})

	exec()
}
