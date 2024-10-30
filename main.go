package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"reflect"
	"strings"
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
		(*dst)[float64(i+t)] = v
	}
}

func ttisnumber(v any) bool {
	return reflect.TypeOf(v).Kind() == reflect.Float64
}

func ttisfunction(v any) bool {
	return reflect.TypeOf(v).Kind() == reflect.Func
}

// bit32 extraction
func extract(n, field, width int) uint32 {
	return uint32(n>>field) & uint32(math.Pow(2, float64(width))-1)
}

func isalpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// Copied directly from BuiltinDefinitions.cpp
func sub(s string, args []string) string {
	size := len(s)
	subbed := 0

	res := strings.Builder{}

	for i := 0; i < size; i++ {
		if s[i] != '%' {
			res.WriteByte(s[i])
			continue
		}

		i++

		if i < size && s[i] == '%' {
			res.WriteByte('%')
			continue
		}

		// -- we just ignore all characters (including flags/precision) up until first alphabetic character
		for i < size && !(s[i] > 0 && (isalpha(s[i]) || s[i] == '*')) {
			i++
		}

		if i == size {
			break
			// } else if s[i] != '*' {
			// 	panic("unknown format") // none other than * at the moment
		} else if subbed >= len(args) {
			res.WriteString("nil") // TODO: idk
			continue
		}

		res.WriteString(args[subbed])
		subbed++
	}

	return res.String()
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
	NamecallHandler  func(kv string, stack *[]any, c1, c2 int) (ok bool, ret []any)
	Extensions       map[any]any
	AllowProxyErrors bool
	DecodeOp         func(op uint32) uint32
}

var luau_print = func(args ...any) (ret []any) {
	args = append([]any{"printed:"}, args...)
	fmt.Println(args...)
	return
}

var luau_settings = LuauSettings{
	VectorCtor: func(...float32) any {
		panic("vectorCtor was not provided")
	},
	VectorSize: 4,
	NamecallHandler: func(kv string, stack *[]any, c1, c2 int) (ok bool, ret []any) {
		switch kv {
		case "format":
			fmt.Println("kv", kv)
			for i, v := range *stack {
				fmt.Printf("    [%d] = %v\n", i, v)
			}
			fmt.Println("c1", c1)
			fmt.Println("c2", c2)

			str := (*stack)[c1].(string)
			args := (*stack)[c1+1 : c2+1]

			// convert all args to strings
			strArgs := make([]string, len(args))
			for i, v := range args { // not inclusive again
				strArgs[i] = fmt.Sprintf("%v", v)
			}

			return true, []any{sub(str, strArgs)}
		}
		panic(fmt.Sprintf("unknown __namecall: %s", kv))
	},
	Extensions: map[any]any{
		"print": &luau_print,
	},
	AllowProxyErrors: false,
	DecodeOp: func(op uint32) uint32 {
		// println("decoding op", op)
		return op
	},
}

type Inst struct {
	A, B, C, D, E, KC     int
	K, K0, K1, K2         any
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
	cursor := uint32(0)

	readByte := func() uint8 {
		b := stream[cursor]
		cursor += 1
		return b
	}

	word := func() uint32 {
		return uint32(stream[cursor]) | uint32(stream[cursor+1])<<8 | uint32(stream[cursor+2])<<16 | uint32(stream[cursor+3])<<24
	}

	readWord := func() uint32 {
		w := word()
		cursor += 4
		return w
	}

	readFloat := func() float32 {
		f := math.Float32frombits(word())
		cursor += 4
		return f
	}

	readDouble := func() float64 {
		word1 := word()
		cursor += 4
		word2 := word()
		cursor += 4

		d := math.Float64frombits(uint64(word1) | uint64(word2)<<32)
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
			inst.D = int(value>>16) & 0xFFFF
			if inst.D >= 0x8000 {
				inst.D -= 0x10000
			}
		case 5: /* AE */
			inst.E = int(value>>8) & 0xFFFFFF
			if inst.E >= 0x800000 {
				inst.E -= 0x1000000
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
		for i := range sizecode {
			checkkmode((*codelist)[i], klist)
		}

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

func truthy(v any) bool {
	b, ok := v.(bool)
	if ok {
		return b
	}
	return v != nil
}

var jumpops = map[uint8]string{
	27: "==",
	28: "<=",
	29: "<",
	30: "~=",
	31: ">",
	32: ">=",
}

var arithops = map[uint8]string{
	33: "add",
	34: "sub",
	35: "mul",
	36: "div",
	37: "mod",
	38: "pow",
	39: "add",
	40: "sub",
	41: "mul",
	42: "div",
	43: "mod",
	44: "pow",
	81: "idiv",
	82: "idiv",
}

var goluautype = map[string]string{
	"float64":                              "number",
	"string":                               "string",
	"bool":                                 "boolean",
	"nil":                                  "nil",
	"map[interface {}]interface {}":        "table",
	"func(...interface {}) []interface {}": "function",
}

func sfops[T string | float64](op uint8, op1, op2 T) bool {
	switch jumpops[op] {
	case "<=":
		return op1 <= op2
	case "<":
		return op1 < op2
	case ">":
		return op1 > op2
	case ">=":
		return op1 >= op2
	}

	panic("unknown floatjump operation")
}

func aops(op uint8, op1, op2 float64) float64 {
	switch arithops[op] {
	case "add":
		return op1 + op2
	case "sub":
		return op1 - op2
	case "mul":
		return op1 * op2
	case "div":
		return op1 / op2
	case "mod":
		return math.Mod(op1, op2)
	case "pow":
		return math.Pow(op1, op2)
	case "idiv":
		return math.Floor(op1 / op2)
	}

	panic("unknown arithmetic operation")
}

func invalidCompare(op uint8, t1, t2 string) string {
	return fmt.Sprintf("attempt to compare %s %s %s", goluautype[t1], jumpops[op], goluautype[t2])
}

func incomparableType(t string, eq bool) string {
	return fmt.Sprintf("type %s cannot be compared; this comparison would always return %t", goluautype[t], eq)
}

func uncallableType(v string) string {
	return fmt.Sprintf("attempt to call a %s value", goluautype[v])
}

func invalidArithmetic(op uint8, t1, t2 string) string {
	return fmt.Sprintf("attempt to perform arithmetic (%s) on %s and %s", arithops[op], goluautype[t1], goluautype[t2])
}

func invalidCond(t string) string {
	return fmt.Sprintf("attempt to compare non-boolean type %s in condition", goluautype[t])
}

func typeOf(v any) string {
	if v == nil { // prevent nil pointer dereference
		return "nil"
	}
	return reflect.TypeOf(v).String()
}

func arithmetic(op uint8, op1, op2 any) float64 {
	t1, t2 := typeOf(op1), typeOf(op2)
	if t1 == "float64" && t2 == "float64" {
		return aops(op, op1.(float64), op2.(float64))
	}

	panic(invalidArithmetic(op, t1, t2))
}

func logic(op uint8, value, op1 any) any {
	switch op {
	case 45, 47: /* AND */
		if !truthy(value) {
			return false
		}
	case 46, 48: /* OR */
		if truthy(value) {
			return value
		}
	default:
		panic("unknown logic operation")
	}

	if truthy(op1) {
		return op1
	}
	return false
}

func jump(op uint8, op1, op2 any) bool {
	t1, t2 := typeOf(op1), typeOf(op2)
	if op == 27 || op == 30 {
		tru := op == 27

		switch op1.(type) {
		case float64, string, bool, nil:
		default:
			panic(incomparableType(t1, tru)) // Also deliberately restricting the ability to compare types that would always return false
		}

		/* JUMPIFEQ, JUMPIFNOTEQ */
		return (op1 == op2) == tru
	} else if t1 == "float64" && t2 == "float64" {
		return sfops(op, op1.(float64), op2.(float64))
	} else if t1 == "string" && t2 == "string" {
		return sfops(op, op1.(string), op2.(string))
	}

	panic(invalidCompare(op, t1, t2))
}

func luau_load(module Deserialise, env map[any]any) (func(...any) []any, func()) {
	protolist := module.protoList
	mainProto := module.mainProto

	alive := true
	luau_close := func() {
		alive = false
	}

	type UpvalIndex struct {
		selfRef bool
		i       int
	}

	type Upval struct {
		value any
		index UpvalIndex
		store any
	}

	var luau_wrapclosure func(module Deserialise, proto Proto, upvals []Upval) *func(...any) []any
	luau_wrapclosure = func(module Deserialise, proto Proto, upvals []Upval) *func(...any) []any {
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
					(*stack)[inst.A] = float64(inst.D) // never put an int on the stack
				case 5: /* LOADK */
					(*stack)[inst.A] = inst.K
				case 6: /* MOVE */
					// we should (ALMOST) never have to change the size of the stack (proto.maxstacksize)
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

					if uv.index.selfRef {
						(*stack)[inst.A] = uv.store.(Upval).value
					} else {
						(*stack)[inst.A] = (*uv.store.(*[]any))[uv.index.i]
					}
				case 10: /* SETUPVAL */
					uv := upvals[inst.B]

					if !uv.index.selfRef {
						(*uv.store.(*[]any))[uv.index.i] = (*stack)[inst.A]
					}
				case 11: /* CLOSEUPVALS */
					for i, uv := range *open_upvalues {
						if uv.index.selfRef || uv.index.i < inst.A {
							continue
						}
						uv.value = (*uv.store.(*[]any))[uv.index.i]
						uv.store = uv
						uv.index.selfRef = true
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

					pc += 1 // -- adjust for aux
				case 13: /* GETTABLE */
					(*stack)[inst.A] = (*stack)[inst.B].(map[any]any)[(*stack)[inst.C]]
				case 14: /* SETTABLE */
					(*stack)[inst.B].(map[any]any)[(*stack)[inst.C]] = (*stack)[inst.A]
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
								index: UpvalIndex{selfRef: true},
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
									index: UpvalIndex{i: index},
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
					fmt.Println("NAMECALL")

					A, B := inst.A, inst.B
					kv := inst.K.(string)
					fmt.Println("kv", kv)

					sb := (*stack)[B]
					fmt.Println("sb", sb)
					(*stack)[A+1] = sb // TODO: 1-based indexing

					pc += 1 // -- adjust for aux

					// -- Special handling for native namecall behaviour
					nativeNamecall := luau_settings.NamecallHandler

					callInst := code[pc-1]
					callOp := callInst.opcode

					// -- Copied from the CALL handler
					callA, callB, callC := callInst.A, callInst.B, callInst.C

					var params int
					if callB == 0 {
						params = top - callA
					} else {
						params = callB - 1
					}

					ok, ret_list := nativeNamecall(kv, stack, callA+1, callA+params)
					if !ok {
						break
					}

					pc += 1 // -- Skip next CALL instruction

					inst = *callInst
					op = callOp

					ret_num := len(ret_list)

					if callC == 0 {
						top = callA + ret_num - 1
					} else {
						ret_num = callC - 1
					}

					move(ret_list, 0, ret_num, callA, stack)
				case 21: /* CALL */
					A, B, C := inst.A, inst.B, inst.C

					var params int
					if B == 0 {
						params = top - A
					} else {
						params = B - 1
					}

					fmt.Println(A, (*stack)[A])

					f := (*stack)[A]
					fn, ok := f.(*func(...any) []any)

					if !ok {
						panic(uncallableType(typeOf(f)))
					}

					fmt.Println("*calling with", (*stack)[A+1:A+params+1])

					ret_list := (*fn)((*stack)[A+1 : A+params+1]...) // not inclusive
					ret_num := int(len(ret_list))

					if C == 0 {
						top = A + ret_num - 1
					} else {
						ret_num = C - 1
					}

					move(ret_list, 0, ret_num, A, stack)
				case 22: /* RETURN */
					A, B := inst.A, inst.B
					b := B - 1

					// nresults
					if b == LUAU_MULTRET {
						b = top - A + 1
					}

					return (*stack)[A:max(A+b, 0)]
				case 23, 24: /* JUMP, JUMPBACK */
					pc += inst.D
				case 25, 26: /* JUMPIF, JUMPIFNOT */
					if truthy((*stack)[inst.A]) == (op == 25) {
						pc += inst.D
					}
				case 27, 28, 29, 30, 31, 32: /* jump */
					if jump(op, (*stack)[inst.A], (*stack)[inst.aux]) {
						pc += inst.D
					} else {
						pc += 1
					}
				case 33, 34, 35, 36, 37, 38, 81: /* arithmetic */
					(*stack)[inst.A] = arithmetic(op, (*stack)[inst.B], (*stack)[inst.C])
				case 39, 40, 41, 42, 43, 44, 82: /* arithmetik */
					fmt.Println("ARITHMETIK", typeOf((*stack)[inst.B]), typeOf(inst.K))
					(*stack)[inst.A] = arithmetic(op, (*stack)[inst.B], inst.K)
				case 45, 46: /* logic */
					(*stack)[inst.A] = logic(op, (*stack)[inst.B], (*stack)[inst.C])
				case 47, 48: /* logik */
					fmt.Println("LOGIK")
					(*stack)[inst.A] = logic(op, (*stack)[inst.B], inst.K)
				case 49: /* CONCAT */
					s := strings.Builder{}
					for i := inst.B; i <= inst.C; i++ {
						s.WriteString((*stack)[i].(string))
					}
					(*stack)[inst.A] = s.String()
				case 50: /* NOT */
					cond := (*stack)[inst.B]
					if cond != true && cond != false {
						panic(invalidCond(typeOf(cond)))
					}

					(*stack)[inst.A] = !(*stack)[inst.B].(bool)
				case 51: /* MINUS */
					(*stack)[inst.A] = -(*stack)[inst.B].(float64)
				case 52: /* LENGTH */
					t := (*stack)[inst.B].(map[any]any)

					length := float64(0)
					for {
						if _, ok := t[length+1]; !ok {
							break
						}
						length++
					}
					(*stack)[inst.A] = length
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

					limit, ok := (*stack)[A].(float64)
					if !ok {
						panic("invalid 'for' limit (number expected)")
					}

					step, ok := (*stack)[A+1].(float64)
					if !ok {
						panic("invalid 'for' step (number expected)")
					}

					index, ok := (*stack)[A+2].(float64)
					if !ok {
						panic("invalid 'for' index (number expected)")
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
					limit := (*stack)[A].(float64)
					step := (*stack)[A+1].(float64)
					index := (*stack)[A+2].(float64) + step

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
					res := int(inst.K.(float64))

					top = int(A + 6)

					switch it := (*stack)[A].(type) {
					case *func(...any) []any:
						fmt.Println("IT func", it)

						vals := (*it)((*stack)[A+1], (*stack)[A+2])

						move(vals, 0, res, A+3, &(*stack))

						fmt.Println(A+3, (*stack)[A+3])

						if (*stack)[A+3] != nil {
							(*stack)[A+2] = (*stack)[A+3]
							pc += inst.D
						} else {
							pc += 1
						}
					default:
						fmt.Println("IT gen", it)

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
				case 59, 61: /* FORGPREP_INEXT, FORGPREP_NEXT */
					if _, ok := (*stack)[inst.A].(*func(...any) []any); !ok {
						panic(fmt.Sprintf("attempt to iterate over a %s value", typeOf((*stack)[inst.A]))) // --  encountered non-function value
					}
					pc += inst.D
				case 60: /* FASTCALL3 */
					/* Skipped */
					pc += 1 // adjust for aux
				case 63: /* GETVARARGS */
					A := inst.A
					b := inst.B - 1

					if b == LUAU_MULTRET {
						b = int(varargs.len)
						top = A + b - 1
					}

					// MAX STACK SIZE IS A LIE!!!!!!!!!!!!!!!!!!!!!!!
					// uh, expand the stack
					for len(*stack) < A+b {
						*stack = append(*stack, nil)
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
								index: UpvalIndex{selfRef: true},
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
				case 71, 72: /* SUBRK, DIVRK */
					fmt.Println("ARITHMETIRK")
					(*stack)[inst.A] = arithmetic(op, inst.K, (*stack)[inst.C])
				case 73: /* FASTCALL1 */
					/* Skipped */
				case 74, 75: /* FASTCALL2, FASTCALL2K */
					/* Skipped */
					pc += 1 // adjust for aux
				case 76: /* FORGPREP */
					pc += inst.D
					if _, ok := (*stack)[inst.A].(*func(...any) []any); ok {
						break
					}

					loopInstruction := *code[pc-1]
					if generalised_iterators[loopInstruction] != nil {
						break
					}

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
					ra, ok := (*stack)[inst.A].(bool)

					if ok && (ra == kv) != kn {
						pc += inst.D
					} else {
						pc += 1
					}
				case 79, 80: /* JUMPXEQKN, JUMPXEQKS */
					kv := inst.K.(float64)
					kn := inst.KN
					ra := (*stack)[inst.A].(float64)

					if (ra == kv) != kn {
						pc += inst.D
					} else {
						pc += 1
					}
				default:
					panic(fmt.Sprintf("Unsupported Opcode: %s op: %d", inst.opname, op))
				}
			}

			for i, uv := range *open_upvalues {
				if uv.index.selfRef {
					continue
				}
				uv.value = (*uv.store.(*[]any))[uv.index.i]
				uv.store = uv
				uv.index.selfRef = true
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

				move(passed, int(start)-1, int(start+l)-1, 0, &varargs.list)
			}

			// TODO: dee bugg ingg
			result := luau_execute(&stack, proto.protos, proto.code, varargs)
			fmt.Println("RESULT", result)

			return result
		}

		return &wrapped
	}

	return *luau_wrapclosure(module, mainProto, []Upval{}), luau_close
}

func main() {
	fmt.Println("Compiling")

	// execute luau-compile
	cmd := exec.Command("luau-compile", "--binary", "-O0", "main.luau")
	// get the output
	bytecode, err := cmd.Output()
	if err != nil {
		fmt.Println("error running luau-compile:", err)
		os.Exit(1)
	}

	deserialised := luau_deserialise(bytecode)

	exec, _ := luau_load(deserialised, map[any]any{})

	exec()
}
