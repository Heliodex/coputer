package main

import (
	"fmt"
	"iter"
	"math"
	"reflect"
	"slices"
	"strings"
)

func p2gte(n uint) uint {
	n--
	for i := range 5 { // lol copilot is magic
		n |= n >> (1 << i)
	}
	return n + 1
}

func arrayKey(k any) (uint, bool) {
	fk, ok := k.(float64)
	return uint(fk), ok && fk == math.Floor(fk) && fk > 0
}

type Table struct {
	array    *[]any
	hash     *map[any]any
	asize    uint
	readonly bool
}

type (
	Function func(co *Coroutine, args ...any) []any
	Status   uint8
)

const (
	Suspended Status = iota
	Running
	Normal
	Dead
)

type Coroutine struct {
	body          *Function
	status        Status
	started       bool
	yield, resume chan Rets
}

func createCoroutine(body *Function) *Coroutine {
	// first time i actually ran into the channel axiom issues
	return &Coroutine{
		body:   body,
		yield:  make(chan Rets, 1),
		resume: make(chan Rets, 1),
	}
}

func (co *Coroutine) Resume(args ...any) (y Rets) {
	if !co.started {
		co.started = true
		co.status = Running

		go func() {
			(*co.body)(co, args...)
			co.status = Dead
			if len(co.yield) == 0 {
				// finish up
				co.yield <- nil
			}
		}()
	} else {
		co.status = Running
		co.resume <- args
	}
	return <-co.yield
}

func NewTable(toHash [][2]any) *Table {
	hash := map[any]any{}
	for _, v := range toHash {
		hash[v[0]] = v[1]
	}
	return &Table{
		hash:     &hash,
		asize:    0,
		readonly: true,
	}
}

// O(n) length? *scoffs*
func (t *Table) Len() (len float64) {
	if t.array == nil {
		return
	}
	// return len(*t.array)
	for _, v := range *t.array {
		if v == nil {
			break
		}
		len++
	}
	return
}

// "The first step in the rehash is to decide the sizes of the new
// array part and the new hash part. So, Lua traverses all entries, counting and
// classifying them, and then chooses as the size of the array part the largest power
// of 2 such that more than half the elements of the array part are filled."
// - Lua performance tips, Roberto Ierusalimschy
func (t *Table) Rehash(nk any, nv any) {
	if t.readonly {
		panic("attempt to modify a readonly table")
	}

	var lenArray, lenHash uint
	arrayExists, hashExists := t.array != nil, t.hash != nil

	if arrayExists {
		lenArray = uint(len(*t.array))
	}
	if hashExists {
		lenHash = uint(len(*t.hash))
	}

	entries := make(map[any]any, lenArray+lenHash)
	// arrayEntries := make([]any, 0)
	// having this be an actual array gives terrible performance for large keys
	// "gee Curtis, O(n) rehashing?"
	arrayEntries := make(map[uint]any, 0)

	// array kvs
	if arrayExists {
		for i, v := range *t.array {
			if v == nil {
				continue
			}
			entries[float64(i+1)] = v
			arrayEntries[uint(i)] = v
		}
	}

	// hash kvs
	if hashExists {
		for k, v := range *t.hash {
			if v == nil {
				continue
			} else if ak, ok := arrayKey(k); ok {
				entries[float64(ak)] = v
				arrayEntries[ak] = v
			} else {
				entries[k] = v
			}
		}
	}

	// new kv
	// fmt.Println("new nknv", nk, nv)
	if ank, ok := arrayKey(nk); ok {
		entries[float64(ank)] = nv
		// fmt.Println("setting array", ank, ank-1)
		arrayEntries[ank-1] = nv
	} else {
		entries[nk] = nv
	}

	// fmt.Println("ALL ENTRIES", entries)

	// count the number of elements in the table
	totalSize := uint(len(entries))
	if totalSize == 0 {
		t.array = nil
		t.hash = nil
		t.asize = 0
		return
	}

	maxP2 := p2gte(totalSize)
	lenArrayEntries := uint(len(arrayEntries))

	var maxToFill uint
	var arrayEntries2 []any

	// halve the size of the array until more than half of the spaces are filled
	for {
		// fmt.Println("halving", maxP2, "until", maxToFill, lenArrayEntries)
		var intsFilled uint
		maxToFill = min(maxP2, totalSize, lenArrayEntries)

		// fmt.Println(arrayEntries)

		arrayEntries2 = make([]any, maxToFill)
		for i, v := range arrayEntries {
			if i >= maxToFill || v == nil {
				continue
			}

			intsFilled++
			arrayEntries2[i] = v
		}

		// fmt.Println("INTSFILLED", intsFilled)

		if intsFilled > maxP2/2 || maxP2 == 0 {
			break
		}
		maxP2 >>= 1
	}

	if maxP2 > 0 {
		t.asize = maxP2

		// fill the new array
		newArray := make([]any, maxP2)
		copy(newArray, arrayEntries2) // birh
		t.array = &newArray

		for i, v := range newArray {
			// remove from entries
			if v != nil {
				// fmt.Println("DELETING", float64(i+1), entries[float64(i+1)])
				delete(entries, float64(i+1))
			}
		}
		// fmt.Println("Remaining", entries)
	}

	t.hash = &entries

	// fmt.Println()
	// fmt.Println("REHASHED")
	// fmt.Println("ARRAY", t.array)
	// fmt.Println("HASH", entries)
	// fmt.Println("ASIZE", t.asize)
	// fmt.Println()
}

func (t *Table) SetArray(i uint, v any) {
	if i > t.asize {
		t.Rehash(float64(i), v)
		return
	}
	(*t.array)[i-1] = v
}

func (t *Table) SetHash(i any, v any) {
	// fmt.Println("setting hash", i, v)
	if t.hash == nil {
		t.hash = &map[any]any{i: v}
		return
	}
	(*t.hash)[i] = v
}

func (t *Table) ForceSet(i, v any) {
	if fi, ok := i.(float64); ok && fi == math.Floor(fi) && (1 <= fi || fi > float64(t.asize)) {
		t.SetArray(uint(fi), v)
		return
	}
	t.SetHash(i, v)
}

func (t *Table) Set(i, v any) {
	if t.readonly {
		panic("attempt to modify a readonly table")
	}
	t.ForceSet(i, v)
}

func (t *Table) GetArray(i uint) any {
	if 1 <= i && i <= t.asize {
		return (*t.array)[i-1]
	}
	return nil
}

func (t *Table) Get(i any) any {
	if fi, ok := i.(float64); ok && fi == math.Floor(fi) {
		if v := t.GetArray(uint(fi)); v != nil {
			return v
		}
	}

	if t.hash == nil {
		return nil
	}
	return (*t.hash)[i]
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

func iterHash(hash map[any]any, yield func(any, any) bool) {
	// order keys in map
	keys := make([]any, 0, len(hash))
	for k := range hash {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, mapKeySort)
	for _, k := range keys {
		if !yield(k, hash[k]) {
			return
		}
	}
}

// 1.23 goes hard
func (t *Table) Iter() iter.Seq2[any, any] {
	return func(yield func(any, any) bool) {
		if t.array != nil {
			iterArray(*t.array, yield)
		}
		if t.hash != nil {
			iterHash(*t.hash, yield)
		}
	}
}

func move(src []any, a, b, t int, dst *[]any) {
	if b < a {
		return
	}

	for t+b-a >= len(*dst) {
		*dst = append(*dst, nil)
	}

	for i := a; i <= b; i++ {
		if i >= len(src) {
			(*dst)[t+i-a] = nil
			continue
		}
		(*dst)[t+i-a] = src[i]
	}
}

// ???
func moveTable(src []any, a, b, t int, dst *Table) {
	if b < a {
		return
	}

	for i, v := range src[a:min(b, len(src))] {
		dst.Set(float64(i+t), v)
	}
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
	VectorCtor      func(...float32) any
	NamecallHandler func(kv string, stack *[]any, c1, c2 int) (ok bool, ret []any)
	// DecodeOp        func(op uint32) uint32
	Extensions map[any]any
	// VectorSize uint8
	// AllowProxyErrors bool
}

var luau_settings = LuauSettings{
	VectorCtor: func(...float32) any {
		panic("vectorCtor was not provided")
	},
	NamecallHandler: func(kv string, stack *[]any, c1, c2 int) (ok bool, ret []any) {
		switch kv {
		case "format":
			str := (*stack)[c1].(string)
			args := (*stack)[c1+1 : c2+1]

			// convert all args to strings
			strArgs := make([]string, len(args))
			for i, v := range args { // not inclusive again
				strArgs[i] = fmt.Sprintf("%v", v)
			}

			return true, []any{sub(str, strArgs)}
		}
		return
	},
	// DecodeOp: func(op uint32) uint32 {
	// 	// println("decoding op", op)
	// 	return op
	// },
	Extensions: map[any]any{
		"math":      libmath,
		"table":     libtable,
		"string":    libstring,
		"coroutine": libcoroutine,
		"bit32":     libbit32,
		"utf8":      libutf8,
		// fuck os
		// and debug
		"buffer": libbuffer,

		// globals
		"type":   MakeFn1("type", global_type)[1],
		// "typeof":   MakeFn1("typeof", global_type)[1], // same because no metatables
	},
	// VectorSize: 4,
	// AllowProxyErrors: false,
}

type Inst struct {
	K, K0, K1, K2          any
	opname                 string
	A, B, C, D, E, KC, aux int
	value                  uint32
	kmode, opcode, opmode  uint8
	KN, usesAux            bool
}

type Varargs struct {
	list []any
	len  int
}

type Proto struct {
	debugname                   string
	k                           []any
	code                        []*Inst
	instructionlineinfo, protos []uint32
	debugcode                   []uint8

	linedefined, sizecode, sizek, sizep, bytecodeid uint32
	maxstacksize, numparams, nups                   uint8
	isvararg, lineinfoenabled                       bool
}

type Deserialised struct {
	mainProto Proto
	protoList []Proto
}

func checkkmode(inst *Inst, k []any) {
	switch inst.kmode {
	case 1: // AUX
		if inst.aux < len(k) { // sometimes huge for some reason
			inst.K = k[inst.aux]
		}
	case 2: // C
		inst.K = k[inst.C]
		// fmt.Println("SET K TO", inst.K, "FROM", inst.C)
	case 3: // D
		inst.K = k[inst.D]
	case 4: // AUX import
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
	case 5: // AUX boolean low 1 bit
		inst.K = extract(inst.aux, 0, 1) == 1
		inst.KN = extract(inst.aux, 31, 1) == 1
	case 6: // AUX number low 24 bits
		inst.K = k[int(extract(inst.aux, 0, 24))] // TODO: 1-based indexing
		inst.KN = extract(inst.aux, 31, 1) == 1
	case 7: // B
		inst.K = k[inst.B] // TODO: 1-based indexing
	case 8: // AUX number low 16 bits
		inst.K = inst.aux & 0xF
	}
}

type Stream struct {
	data []byte
	pos  uint32
}

func (s *Stream) rByte() uint8 {
	b := s.data[s.pos]
	s.pos += 1
	return b
}

func (s *Stream) rWord() uint32 {
	w := uint32(s.data[s.pos]) |
		uint32(s.data[s.pos+1])<<8 |
		uint32(s.data[s.pos+2])<<16 |
		uint32(s.data[s.pos+3])<<24
	s.pos += 4
	return w
}

func (s *Stream) rFloat32() float32 {
	return math.Float32frombits(s.rWord())
}

func (s *Stream) rFloat64() float64 {
	return math.Float64frombits(uint64(s.rWord()) | uint64(s.rWord())<<32)
}

func (s *Stream) rVarInt() (result uint32) {
	for i := range 4 {
		value := uint32(s.rByte())
		result |= ((value & 0x7F) << (i * 7))
		if value&0x80 == 0 {
			break
		}
	}
	return
}

func (s *Stream) rString() string {
	size := s.rVarInt()
	if size == 0 {
		return ""
	}

	str := make([]byte, size)
	for i := range size {
		str[i] = s.data[s.pos+uint32(i)]
	}
	s.pos += size

	return string(str)
}

func (s *Stream) CheckEnd() {
	if s.pos != uint32(len(s.data)) {
		panic("deserialiser position mismatch")
	}
}

func luau_deserialise(data []byte) Deserialised {
	stream := Stream{data: data}

	rByte, rWord, rFloat32, rFloat64, rVarInt, rString := stream.rByte, stream.rWord, stream.rFloat32, stream.rFloat64, stream.rVarInt, stream.rString

	if luauVersion := rByte(); luauVersion == 0 {
		panic("the provided bytecode is an error message")
	} else if luauVersion != 6 {
		panic("the version of the provided bytecode is unsupported")
	} else if rByte() != 3 { // types version
		panic("the types version of the provided bytecode is unsupported")
	}

	stringCount := rVarInt()
	stringList := make([]string, stringCount)
	for i := range stringCount {
		stringList[i] = rString()
	}

	readInstruction := func(codeList *[]*Inst) (usesAux bool) {
		// value := luau_settings.DecodeOp(readWord())
		value := rWord()
		opcode := uint8(value & 0xFF)

		opinfo := opList[opcode]
		opmode := opinfo.Mode
		usesAux = opinfo.HasAux

		inst := &Inst{
			opname:  opinfo.Name,
			kmode:   opinfo.KMode,
			opcode:  opcode,
			opmode:  opmode,
			usesAux: usesAux,
		}

		*codeList = append(*codeList, inst)

		switch opmode {
		case 1: // A
			inst.A = int(value>>8) & 0xFF
		case 2: // AB
			inst.A = int(value>>8) & 0xFF
			inst.B = int(value>>16) & 0xFF
		case 3: // ABC
			inst.A = int(value>>8) & 0xFF
			inst.B = int(value>>16) & 0xFF
			inst.C = int(value>>24) & 0xFF
		case 4: // AD
			inst.A = int(value>>8) & 0xFF
			inst.D = int(value>>16) & 0xFFFF
			if inst.D >= 0x8000 {
				inst.D -= 0x10000
			}
		case 5: // AE
			inst.E = int(value>>8) & 0xFFFFFF
			if inst.E >= 0x800000 {
				inst.E -= 0x1000000
			}
		}

		if usesAux {
			aux := rWord()
			inst.aux = int(aux)

			*codeList = append(*codeList, &Inst{
				opname: "auxvalue",
				value:  aux,
			})
		}
		return
	}

	readProto := func(bytecodeid uint32) Proto {
		maxstacksize := rByte()
		numparams := rByte()
		nups := rByte()
		isvararg := rByte() != 0

		rByte()                 // -- flags
		stream.pos += rVarInt() // typesize

		sizecode := rVarInt()
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

		sizek := rVarInt()
		klist := make([]any, sizek)

		for i := range sizek {
			var k any
			switch kt := rByte(); kt {
			case 0: // Nil
				k = nil
			case 1: // Bool
				k = rByte() != 0
			case 2: // Number
				k = rFloat64()
			case 3: // String
				k = stringList[rVarInt()-1] // TODO: 1-based indexing
			case 4: // Function
				k = rWord()
			case 5: // Table
				dataLength := rVarInt()
				k = make([]uint32, dataLength)

				for j := range dataLength {
					k.([]uint32)[j] = rVarInt() // TODO: 1-based indexing
				}
			case 6: // Closure
				k = rVarInt()
			case 7: // Vector
				k = luau_settings.VectorCtor(rFloat32(), rFloat32(), rFloat32(), rFloat32())
			default:
				panic(fmt.Sprintf("Unknown ktype %d", kt))
			}

			klist[i] = k
		}

		// -- 2nd pass to replace constant references in the instruction
		for i := range sizecode {
			checkkmode((*codelist)[i], klist)
		}

		sizep := rVarInt()
		protos := make([]uint32, sizep)
		for i := range sizep {
			protos[i] = rVarInt() + 1 // TODO: 1-based indexing
		}

		linedefined := rVarInt()

		var debugname string
		if debugnameindex := rVarInt(); debugnameindex == 0 {
			debugname = "(??)"
		} else {
			debugname = stringList[debugnameindex-1] // TODO: 1-based indexing
		}

		// -- lineinfo
		lineinfoenabled := rByte() != 0
		var instructionlineinfo []uint32

		if lineinfoenabled {
			linegaplog2 := rByte()
			intervals := uint32((sizecode-1)>>linegaplog2) + 1

			lineinfo := make([]uint32, sizecode)
			var lastoffset uint32
			for i := range sizecode {
				lastoffset += uint32(rByte()) // TODO: type convs?
				lineinfo[i] = lastoffset
			}

			abslineinfo := make([]uint32, intervals)
			var lastline uint32
			for i := range intervals {
				lastline += rWord()
				abslineinfo[i] = uint32(uint64(lastline) % (uint64(math.Pow(2, 32)))) // TODO: 1-based indexing
			}

			instructionlineinfo = make([]uint32, sizecode)
			for i := range sizecode {
				// -- p->abslineinfo[pc >> p->linegaplog2] + p->lineinfo[pc];
				instructionlineinfo = append(instructionlineinfo, abslineinfo[i>>linegaplog2]+lineinfo[i]) // TODO: 1-based indexing
			}
		}

		// -- debuginfo
		if rByte() != 0 {
			fmt.Println("DEBUGINFO")
			for range rVarInt() { // sizel
				rVarInt()
				rVarInt()
				rVarInt()
				rByte()
			}
			for range rVarInt() { // sizeupvalues
				rVarInt()
			}
		}

		return Proto{
			debugname,
			klist,
			*codelist,
			instructionlineinfo,
			protos,
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
		}
	}

	// userdataRemapping (not used in VM, left unused)
	index := rByte()
	for index != 0 {
		rVarInt()
		index = rByte()
	}

	protoCount := rVarInt()
	protoList := make([]Proto, protoCount)
	for i := range protoCount {
		protoList[i] = readProto(i - 1)
	}

	mainProto := protoList[rVarInt()]
	mainProto.debugname = "(main)"
	stream.CheckEnd()

	return Deserialised{mainProto, protoList}
}

type Iterator struct {
	args, resume chan *[]any
	running      bool
}

type Upval struct {
	value, store any
	index        int
	selfRef      bool
}

func truthy(v any) bool {
	if b, ok := v.(bool); ok {
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
	51: "unm",
	71: "sub",
	72: "div",
	81: "idiv",
	82: "idiv",
}

var luautype = map[string]string{
	"nil":            "nil",
	"float64":        "number",
	"string":         "string",
	"bool":           "boolean",
	"*main.Table":    "table",
	"*main.Function": "function",
	"*main.Coroutine": "thread",
	"*main.Buffer":   "buffer",
}

func sfops[T string | float64](op uint8, a, b T) bool {
	switch jumpops[op] {
	case "<=":
		return a <= b
	case "<":
		return a < b
	case ">":
		return a > b
	case ">=":
		return a >= b
	}

	panic("unknown floatjump operation")
}

func aops(op uint8, a, b float64) float64 {
	switch arithops[op] {
	case "add":
		return a + b
	case "sub":
		return a - b
	case "mul":
		return a * b
	case "div":
		return a / b
	case "mod":
		return a - b*math.Floor(a/b)
	case "pow":
		return math.Pow(a, b)
	case "idiv":
		return math.Floor(a / b)
	}

	panic("unknown arithmetic operation")
}

func invalidCompare(op string, ta, tb string) string {
	return fmt.Sprintf("attempt to compare %s %s %s", luautype[ta], op, luautype[tb])
}

func incomparableType(t string, eq bool) string {
	return fmt.Sprintf("type %s cannot be compared; this comparison would always return %t", luautype[t], eq)
}

func uncallableType(v string) string {
	return fmt.Sprintf("attempt to call a %s value", luautype[v])
}

func invalidArithmetic(op string, ta, tb string) string {
	return fmt.Sprintf("attempt to perform arithmetic (%s) on %s and %s", op, luautype[ta], luautype[tb])
}

func invalidUnm(t string) string {
	return fmt.Sprintf("attempt to perform arithmetic (unm) on %s", luautype[t])
}

func invalidCond(t string) string {
	return fmt.Sprintf("attempt to compare non-boolean type %s in condition", luautype[t])
}

func invalidFor(pos, t string) string {
	return fmt.Sprintf("invalid 'for' %s (number expected, got %s)", pos, luautype[t])
}

func invalidLength(t string) string {
	return fmt.Sprintf("attempt to get length of a %s value", luautype[t])
}

func invalidIndex(ta string, val any) string {
	tb := luautype[typeOf(val)]
	if tb == "string" {
		tb = fmt.Sprintf("'%v'", val)
	}

	panic(fmt.Sprintf("attempt to index %v with %v", luautype[ta], tb))
}

func typeOf(v any) string {
	if v == nil { // prevent nil pointer dereference
		return "nil"
	}
	return reflect.TypeOf(v).String()
}

func arithmetic(op uint8, a, b any) float64 {
	ta, tb := typeOf(a), typeOf(b)
	if ta == "float64" && tb == "float64" {
		return aops(op, a.(float64), b.(float64))
	}

	panic(invalidArithmetic(arithops[op], ta, tb))
}

func logic(op uint8, a, b any) any {
	switch op {
	case 45, 47: // AND
		if !truthy(a) {
			return false
		}
	case 46, 48: // OR
		if truthy(a) {
			return a
		}
	default:
		panic("unknown logic operation")
	}

	if truthy(b) {
		return b
	}
	return false
}

func jump(op uint8, a, b any) bool {
	ta, tb := typeOf(a), typeOf(b)
	if op == 27 || op == 30 {
		tru := op == 27

		switch a.(type) {
		case float64, string, bool, nil:
		default:
			panic(incomparableType(ta, tru)) // Also deliberately restricting the ability to compare types that would always return false
		}

		// JUMPIFEQ, JUMPIFNOTEQ
		return (a == b) == tru
	} else if ta == "float64" && tb == "float64" {
		return sfops(op, a.(float64), b.(float64))
	} else if ta == "string" && tb == "string" {
		return sfops(op, a.(string), b.(string))
	}

	panic(invalidCompare(jumpops[op], ta, tb))
}

func luau_load(module Deserialised, env map[any]any) (Coroutine, func()) {
	protolist := module.protoList
	alive := true

	var luau_wrapclosure func(proto Proto, upvals []Upval) *Function

	luau_execute := func(
		proto Proto,
		upvals []Upval,
		stack *[]any,
		protos []uint32,
		code []*Inst,
		co *Coroutine,
		varargs Varargs,
	) []any {
		top, pc, open_upvalues, generalised_iterators := -1, 1, new([]*Upval), map[Inst]*Iterator{}
		extensions := luau_settings.Extensions

		var handlingBreak bool
		var inst Inst
		var op uint8

		// a a a a
		// stayin' alive
		for alive {
			if !handlingBreak {
				inst = *code[pc-1]
				op = inst.opcode
			}
			handlingBreak = false

			pc += 1

			// fmt.Println("OP", op, "PC", pc)

			switch op {
			case 0: // NOP
				// -- Do nothing
			case 1: // BREAK
				pc -= 1
				op = proto.debugcode[pc]
				handlingBreak = true
			case 2: // LOADNIL
				(*stack)[inst.A] = nil
			case 3: // LOADB
				(*stack)[inst.A] = inst.B == 1
				pc += inst.C
			case 4: // LOADN
				(*stack)[inst.A] = float64(inst.D) // never put an int on the stack
			case 5: // LOADK
				(*stack)[inst.A] = inst.K
			case 6: // MOVE
				// we should (ALMOST) never have to change the size of the stack (proto.maxstacksize)
				(*stack)[inst.A] = (*stack)[inst.B]
			case 7: // GETGLOBAL
				kv := inst.K

				if extensions[kv] != nil {
					(*stack)[inst.A] = extensions[kv]
				} else {
					(*stack)[inst.A] = env[kv]
				}

				pc += 1 // -- adjust for aux
			case 8: // SETGLOBAL
				// LOL
				kv := inst.K
				if _, ok := kv.(string); ok {
					if extensions[kv] != nil {
						panic(fmt.Sprintf("attempt to redefine global '%s'", kv))
					}
					panic(fmt.Sprintf("attempt to set global '%s'", kv))
				}

				// env[kv] = (*stack)[inst.A]

				// pc += 1 // -- adjust for aux
			case 9: // GETUPVAL
				if uv := upvals[inst.B]; uv.selfRef {
					(*stack)[inst.A] = uv.store.(Upval).value
				} else {
					// fmt.Println("GETTING UPVAL", uv.store)

					(*stack)[inst.A] = (*uv.store.(*[]any))[uv.index]
				}
			case 10: // SETUPVAL
				if uv := upvals[inst.B]; !uv.selfRef {
					(*uv.store.(*[]any))[uv.index] = (*stack)[inst.A]
				}
			case 11: // CLOSEUPVALS
				for i, uv := range *open_upvalues {
					if uv == nil || uv.selfRef || uv.index < inst.A {
						continue
					}
					uv.value = (*uv.store.(*[]any))[uv.index]
					uv.store = uv
					uv.selfRef = true
					(*open_upvalues)[i] = nil
				}
			case 12: // GETIMPORT
				k0 := inst.K0
				imp := extensions[k0]
				if imp == nil {
					imp = env[k0]
				}

				switch inst.KC { // count
				case 1:
					(*stack)[inst.A] = imp
				case 2:
					t := imp.(*Table)
					(*stack)[inst.A] = t.Get(inst.K1)
				case 3:
					t := imp.(*Table)
					(*stack)[inst.A] = t.Get(inst.K1).([]any)[inst.K2.(uint32)-1]
				}

				pc += 1 // -- adjust for aux
			case 13, 14: // GETTABLE, SETTABLE
				index := (*stack)[inst.C]
				t, ok := (*stack)[inst.B].(*Table)
				if !ok {
					panic(invalidIndex(typeOf((*stack)[inst.B]), index))
				}

				if op == 13 {
					(*stack)[inst.A] = t.Get(index)
				} else {
					// fmt.Println("SETTABLE", index, (*stack)[inst.A])
					t.Set(index, (*stack)[inst.A])
				}
			case 15, 16: // GETTABLEKS, SETTABLEKS
				index := inst.K
				t, ok := (*stack)[inst.B].(*Table)
				if !ok {
					// fmt.Println("indexing", typeOf((*stack)[inst.B]), "with", index)
					panic(invalidIndex(typeOf((*stack)[inst.B]), index))
				}

				if op == 15 {
					(*stack)[inst.A] = t.Get(index)
				} else {
					t.Set(index, (*stack)[inst.A])
				}

				pc += 1 // -- adjust for aux
			case 17: // GETTABLEN
				(*stack)[inst.A] = (*stack)[inst.B].(*Table).Get(float64(inst.C + 1))
			case 18: // SETTABLEN
				(*stack)[inst.B].(*Table).Set(float64(inst.C+1), (*stack)[inst.A])
			case 19: // NEWCLOSURE
				newPrototype := protolist[protos[inst.D]-1]

				nups := newPrototype.nups
				upvalues := make([]Upval, nups)
				(*stack)[inst.A] = luau_wrapclosure(newPrototype, upvalues)

				// fmt.Println("nups", nups)
				for i := range nups {
					switch pseudo := code[pc-1]; pseudo.A {
					case 0: // -- value
						upvalue := Upval{
							value:   (*stack)[pseudo.B],
							selfRef: true,
						}
						upvalue.store = upvalue

						upvalues[i] = upvalue
					case 1: // -- reference
						index := pseudo.B
						// fmt.Println("index", index, len(*open_upvalues))

						var prev *Upval
						if index < len(*open_upvalues) {
							prev = (*open_upvalues)[index]
						}

						if prev == nil {
							prev = &Upval{
								store: stack,
								index: index,
							}

							for len(*open_upvalues) <= index {
								*open_upvalues = append(*open_upvalues, nil)
							}
							(*open_upvalues)[index] = prev
						}

						upvalues[i] = *prev
					case 2: // -- upvalue
						upvalues[i] = upvals[pseudo.B]
					}
					pc += 1
				}
			case 20: // NAMECALL
				// fmt.Println("NAMECALL")

				A, B := inst.A, inst.B
				kv := inst.K.(string)
				// fmt.Println("kv", kv)

				(*stack)[A+1] = (*stack)[B]

				// -- Special handling for native namecall behaviour
				callInst := code[pc]
				callOp := callInst.opcode

				pc += 1 // -- adjust for aux

				// -- Copied from the CALL handler
				callA, callB, callC := callInst.A, callInst.B, callInst.C

				var params int
				if callB == 0 {
					params = top - callA
				} else {
					params = callB - 1
				}

				ok, ret_list := luau_settings.NamecallHandler(kv, stack, callA+1, callA+params)
				if !ok {
					(*stack)[A] = (*stack)[B].(*Table).Get(kv)
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
			case 21: // CALL
				A, B, C := inst.A, inst.B, inst.C

				var params int
				if B == 0 {
					params = top - A
				} else {
					params = B - 1
				}

				// fmt.Println(A, (*stack)[A])

				f := (*stack)[A]
				fn, ok := f.(*Function)
				if !ok {
					panic(uncallableType(typeOf(f)))
				}

				ret_list := (*fn)(co, (*stack)[A+1:A+params+1]...) // not inclusive
				ret_num := int(len(ret_list))

				// development checking lelell
				for _, v := range ret_list {
					switch v.(type) {
					case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
						panic(fmt.Sprintf("Hey idiot YOU RETURNED AN INTEGER INSTEAD OFA  FLOAT FROM YUR FUNCTION O MY GOD %v", v))
					case float32:
						panic(fmt.Sprintf("u  dun fukt up %v", v))
					}
				}

				if C == 0 {
					top = A + ret_num - 1
				} else {
					ret_num = C - 1
				}

				move(ret_list, 0, ret_num, A, stack)
			case 22: // RETURN
				A, B := inst.A, inst.B
				b := B - 1

				// nresults
				if b == LUAU_MULTRET {
					b = top - A + 1
				}

				return (*stack)[A:max(A+b, 0)]
			case 23, 24: // JUMP, JUMPBACK
				pc += inst.D
			case 25, 26: // JUMPIF, JUMPIFNOT
				if truthy((*stack)[inst.A]) == (op == 25) {
					pc += inst.D
				}
			case 27, 28, 29, 30, 31, 32: // jump
				if jump(op, (*stack)[inst.A], (*stack)[inst.aux]) {
					pc += inst.D
				} else {
					pc += 1
				}
			case 33, 34, 35, 36, 37, 38, 81: // arithmetic
				(*stack)[inst.A] = arithmetic(op, (*stack)[inst.B], (*stack)[inst.C])
			case 39, 40, 41, 42, 43, 44, 82: // arithmetik
				(*stack)[inst.A] = arithmetic(op, (*stack)[inst.B], inst.K)
			case 45, 46: // logic
				(*stack)[inst.A] = logic(op, (*stack)[inst.B], (*stack)[inst.C])
			case 47, 48: // logik
				fmt.Println("LOGIK")
				(*stack)[inst.A] = logic(op, (*stack)[inst.B], inst.K)
			case 49: // CONCAT
				s := strings.Builder{}
				for i := inst.B; i <= inst.C; i++ {
					s.WriteString((*stack)[i].(string))
				}
				(*stack)[inst.A] = s.String()
			case 50: // NOT
				cond, ok := (*stack)[inst.B].(bool)
				if !ok {
					panic(invalidCond(typeOf((*stack)[inst.B])))
				}

				(*stack)[inst.A] = !cond
			case 51: // MINUS
				a, ok := (*stack)[inst.B].(float64)
				if !ok {
					panic(invalidUnm(typeOf((*stack)[inst.B])))
				}

				(*stack)[inst.A] = -a
			case 52: // LENGTH
				switch t := (*stack)[inst.B].(type) {
				case *Table:
					(*stack)[inst.A] = t.Len()
				case string:
					(*stack)[inst.A] = float64(len(t))
				default:
					panic(invalidLength(typeOf(t)))
				}
			case 53: // NEWTABLE
				(*stack)[inst.A] = &Table{}

				pc += 1 // -- adjust for aux
			case 54: // DUPTABLE
				template := inst.K.([]uint32)
				serialised := &Table{}
				for _, id := range template {
					serialised.Set(proto.k[id], nil) // constants
				}
				(*stack)[inst.A] = serialised
			case 55: // SETLIST
				A, B := inst.A, inst.B
				c := inst.C - 1

				if c == LUAU_MULTRET {
					c = top - B + 1
				}

				s := (*stack)[A].(*Table)

				// one-indexed lol
				moveTable(*stack, B, B+c, inst.aux, s)
				(*stack)[A] = s

				pc += 1 // -- adjust for aux
			case 56: // FORNPREP
				A := inst.A

				index, ok := (*stack)[A+2].(float64)
				if !ok {
					panic(invalidFor("initial value", typeOf((*stack)[A+2])))
				}

				limit, ok := (*stack)[A].(float64)
				if !ok {
					panic(invalidFor("limit", typeOf((*stack)[A])))
				}

				step, ok := (*stack)[A+1].(float64)
				if !ok {
					panic(invalidFor("step", typeOf((*stack)[A+1])))
				}

				if step > 0 {
					if index > limit {
						pc += inst.D
					}
				} else if limit > index {
					pc += inst.D
				}
			case 57: // FORNLOOP
				A := inst.A
				limit := (*stack)[A].(float64)
				step := (*stack)[A+1].(float64)
				init := (*stack)[A+2].(float64) + step

				(*stack)[A+2] = init

				if step > 0 {
					if init <= limit {
						pc += inst.D
					}
				} else if limit <= init {
					pc += inst.D
				}
			case 58: // FORGLOOP
				A := inst.A
				res := inst.K.(int)

				top = int(A + 6)

				switch it := (*stack)[A].(type) {
				case *Function:
					// fmt.Println("IT func", it)

					vals := (*it)(co, []any{(*stack)[A+1], (*stack)[A+2]})

					move(vals, 0, res, A+3, stack)

					// fmt.Println(A+3, (*stack)[A+3])

					if (*stack)[A+3] != nil {
						(*stack)[A+2] = (*stack)[A+3]
						pc += inst.D
					} else {
						pc += 1
					}
				default:
					iter := *generalised_iterators[inst]

					if !iter.running {
						args := &[]any{it, (*stack)[A+1], (*stack)[A+2]}
						// fmt.Println("-1- sending thru the wire", args)
						iter.args <- args
						// fmt.Println("-1- sent")
					}
					vals := <-iter.resume
					// fmt.Println("-1- received!", vals)

					if vals == nil {
						delete(generalised_iterators, inst)
						pc += 1
					} else {
						move(*vals, 0, res, A+3, stack)

						(*stack)[A+2] = (*stack)[A+3]
						pc += inst.D
					}
				}
			case 59, 61: // FORGPREP_INEXT, FORGPREP_NEXT
				if _, ok := (*stack)[inst.A].(*Function); !ok {
					panic(fmt.Sprintf("attempt to iterate over a %s value", typeOf((*stack)[inst.A]))) // -- encountered non-function value
				}
				pc += inst.D
			case 60: // FASTCALL3
				// Skipped
				pc += 1 // adjust for aux
			case 63: // GETVARARGS
				A := inst.A
				b := inst.B - 1

				if b == LUAU_MULTRET {
					b = varargs.len
					top = A + b - 1
				}

				// MAX STACK SIZE IS A LIE!!!!!!!!!!!!!!!!!!!!!!!
				// uh, expand the stack
				// fmt.Println("EXPANDING STACK", A+b)
				for len(*stack) < A+b {
					*stack = append(*stack, nil)
				}

				move(varargs.list, 0, b, A, stack)
			case 64: // DUPCLOSURE
				newPrototype := protolist[inst.K.(uint32)] // TODO: 1-based indexing

				nups := newPrototype.nups
				upvalues := make([]Upval, nups)
				(*stack)[inst.A] = luau_wrapclosure(newPrototype, upvalues)

				for i := range nups {
					pseudo := code[pc-1]
					pc += 1

					t := pseudo.A
					if t == 0 { // value
						upvalue := Upval{
							value:   (*stack)[pseudo.B],
							selfRef: true,
						}
						upvalue.store = upvalue
						upvalues[i] = upvalue

						// -- references dont get handled by DUPCLOSURE
					} else if t == 2 { // upvalue
						upvalues[i] = upvals[pseudo.B]
					}
				}
			case 65: // PREPVARARGS
				// Handled by wrapper
			case 66: // LOADKX
				(*stack)[inst.A] = inst.K.(uint32) // kv

				pc += 1 // -- adjust for aux
			case 67: // JUMPX
				pc += inst.E
			case 68: // FASTCALL
				// Skipped
			case 69: // COVERAGE
				inst.E += 1
			case 70: // CAPTURE
				// Handled by CLOSURE
				panic("encountered unhandled CAPTURE")
			case 71, 72: // SUBRK, DIVRK
				// fmt.Println("ARITHMETIRK")
				(*stack)[inst.A] = arithmetic(op, inst.K, (*stack)[inst.C])
			case 73: // FASTCALL1
				// Skipped
			case 74, 75: // FASTCALL2, FASTCALL2K
				// Skipped
				pc += 1 // adjust for aux
			case 76: // FORGPREP
				pc += inst.D
				if _, ok := (*stack)[inst.A].(*Function); ok {
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

				generalised_iterators[loopInstruction] = c
			case 77, 78, 79, 80: // JUMPXEQKNIL, JUMPXEQKB, JUMPXEQKN, JUMPXEQKS
				var jmp bool
				if op == 77 {
					jmp = ((*stack)[inst.A] == nil) != inst.KN
				} else if op == 78 {
					kv := inst.K.(bool)
					ra, ok := (*stack)[inst.A].(bool)

					jmp = ok && (ra == kv) != inst.KN
				} else if op == 79 || op == 80 {
					kv := inst.K.(float64)
					ra := (*stack)[inst.A].(float64)

					jmp = (ra == kv) != inst.KN
				}

				if jmp {
					pc += inst.D
				} else {
					pc += 1
				}
			default:
				panic(fmt.Sprintf("Unsupported Opcode: %s op: %d", inst.opname, op))
			}
		}

		for i, uv := range *open_upvalues {
			if uv.selfRef {
				continue
			}
			uv.value = (*uv.store.(*[]any))[uv.index]
			uv.store = uv
			uv.selfRef = true
			(*open_upvalues)[i] = nil
		}

		for i := range generalised_iterators {
			generalised_iterators[i].running = false
			delete(generalised_iterators, i)
		}
		return []any{}
	}

	luau_wrapclosure = func(proto Proto, upvals []Upval) *Function {
		wrapped := Function(func(co *Coroutine, passed ...any) []any {
			maxstacksize, numparams := proto.maxstacksize, proto.numparams

			stack := make([]any, maxstacksize)
			// fmt.Println("MAX STACK SIZE", maxstacksize)
			varargs := Varargs{list: []any{}}

			move(passed, 0, int(numparams), 0, &stack)

			n := uint8(len(passed))
			if numparams < n {
				start := int(numparams + 1)
				l := int(n) - int(numparams)
				varargs.len = l

				// expand varargs list
				varargs.list = make([]any, l)

				move(passed, start-1, start+l-1, 0, &varargs.list)
			}

			// TODO: dee bugg ingg
			return luau_execute(proto, upvals, &stack, proto.protos, proto.code, co, varargs)
		})

		return &wrapped
	}

	return Coroutine{
			body:   luau_wrapclosure(module.mainProto, []Upval{}),
			yield:  make(chan Rets, 1),
			resume: make(chan Rets, 1),
		}, func() {
			alive = false
		}
}
