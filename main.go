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
	env           map[any]any
	filepath      string // lel nowhere else to put this
	yield, resume chan Rets
	status        Status
	o             uint8
	started       bool
}

func createCoroutine(body *Function) *Coroutine {
	// first time i actually ran into the channel axiom issues
	return &Coroutine{
		body:   body,
		yield:  make(chan Rets, 1),
		resume: make(chan Rets, 1),
	}
}

func (co *Coroutine) Resume(args ...any) Rets {
	if !co.started {
		co.started = true
		co.status = Running

		go func() {
			co.yield <- (*co.body)(co, args...)
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
	if b <= a {
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

type Vector [4]float32

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
	var subbed int

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

var VectorCtor = func(x, y, z, w float32) Vector {
	return Vector{x, y, z, w}
}

var NamecallHandler = func(co *Coroutine, kv string, stack *[]any, c1, c2 int) (ok bool, retList []any) {
	switch kv {
	case "format":
		str := (*stack)[c1].(string)
		args := (*stack)[c1+1 : c2+1]

		return true, []any{fmtstring(str, Args{args, "format", co, 0})}
	}
	return
}

var Extensions = map[any]any{
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
	"type": MakeFn1("type", global_type)[1],
	// "typeof":   MakeFn1("typeof", global_type)[1], // same because no metatables
	"ipairs":   MakeFn("ipairs", global_ipairs)[1],
	"pairs":    MakeFn("pairs", global_pairs)[1],
	"next":     MakeFn("next", global_next)[1],
	"tonumber": MakeFn1("tonumber", global_tonumber)[1],
	"tostring": MakeFn1("tostring", global_tostring)[1],
	"_VERSION": "Luau", // todo: custom

	"require": MakeFn1("require", global_require)[1],
}

// var VectorSize = 4
// var AllowProxyErrors = false

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
		inst.K0 = k[id0] // maybe can .(string) this
		// fmt.Println("AUX", inst.K0)

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
		inst.K = k[extract(inst.aux, 0, 24)]
		inst.KN = extract(inst.aux, 31, 1) == 1
	case 7: // B
		inst.K = k[inst.B]
	case 8: // AUX number low 16 bits
		inst.K = inst.aux & 0xF
	}
}

type Stream struct {
	data []byte
	pos  uint32
}

func (s *Stream) rByte() byte {
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
		str[i] = s.data[s.pos+i]
	}
	s.pos += size

	return string(str)
}

func (s *Stream) CheckEnd() {
	if s.pos != uint32(len(s.data)) {
		panic("deserialiser position mismatch")
	}
}

func readInstruction(codeList *[]*Inst, s *Stream) (usesAux bool) {
	value := s.rWord()
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
		aux := s.rWord()
		inst.aux = int(aux)

		*codeList = append(*codeList, &Inst{
			opname: "auxvalue",
			value:  aux,
		})
	}
	return
}

func readProto(bytecodeid uint32, stringList []string, s *Stream) Proto {
	maxstacksize := s.rByte()
	numparams := s.rByte()
	nups := s.rByte()
	isvararg := s.rByte() != 0

	s.rByte()            // -- flags
	s.pos += s.rVarInt() // typesize

	sizecode := s.rVarInt()
	codelist := new([]*Inst)

	var skipnext bool
	for range sizecode {
		if skipnext {
			skipnext = false
			continue
		}
		skipnext = readInstruction(codelist, s)
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
			klist[i] = VectorCtor(s.rFloat32(), s.rFloat32(), s.rFloat32(), s.rFloat32())
		default:
			panic(fmt.Sprintf("Unknown ktype %d", kt))
		}
	}

	// -- 2nd pass to replace constant references in the instruction
	for i := range sizecode {
		checkkmode((*codelist)[i], klist)
	}

	sizep := s.rVarInt()
	protos := make([]uint32, sizep)
	for i := range sizep {
		protos[i] = s.rVarInt() + 1
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
	var instructionlineinfo []uint32

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
		var lastline uint32
		for i := range intervals {
			lastline += s.rWord()
			abslineinfo[i] = uint32(uint64(lastline) % (uint64(math.Pow(2, 32)))) // TODO: 1-based indexing
		}

		instructionlineinfo = make([]uint32, sizecode)
		for i := range sizecode {
			// -- p->abslineinfo[pc >> p->linegaplog2] + p->lineinfo[pc];
			instructionlineinfo = append(instructionlineinfo, abslineinfo[i>>linegaplog2]+lineinfo[i])
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

func Deserialise(data []byte) Deserialised {
	s := &Stream{data: data}

	if luauVersion := s.rByte(); luauVersion == 0 {
		panic("the provided bytecode is an error message")
	} else if luauVersion != 6 {
		panic("the version of the provided bytecode is unsupported")
	} else if s.rByte() != 3 { // types version
		panic("the types version of the provided bytecode is unsupported")
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
	protoList := make([]Proto, protoCount)
	for i := range protoCount {
		protoList[i] = readProto(i-1, stringList, s)
	}

	mainProto := protoList[s.rVarInt()]
	mainProto.debugname = "(main)"
	s.CheckEnd()

	return Deserialised{mainProto, protoList}
}

type Iterator struct {
	args, resume chan *[]any
	running      bool
}

type Upval struct {
	value   any
	store   *[]any
	index   int
	selfRef bool
}

func truthy(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return v != nil
}

var luautype = map[string]string{
	"nil":             "nil",
	"float64":         "number",
	"string":          "string",
	"bool":            "boolean",
	"*main.Table":     "table",
	"*main.Function":  "function",
	"*main.Coroutine": "thread",
	"*main.Buffer":    "buffer",
	"main.Vector":     "vector",
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

func invalidConcat(t1, t2 string) string {
	return fmt.Sprintf("attempt to concatenate %s with %s", luautype[t1], luautype[t2])
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

func aAdd(a, b any) any {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa + fb
	}

	va, ok3 := a.(Vector)
	vb, ok4 := b.(Vector)
	if ok3 && ok4 {
		return Vector{va[0] + vb[0], va[1] + vb[1], va[2] + vb[2], va[3] + vb[3]}
	}

	panic(invalidArithmetic("add", typeOf(a), typeOf(b)))
}

func aSub(a, b any) any {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa - fb
	}

	va, ok3 := a.(Vector)
	vb, ok4 := b.(Vector)
	if ok3 && ok4 {
		return Vector{va[0] - vb[0], va[1] - vb[1], va[2] - vb[2], va[3] - vb[3]}
	}

	panic(invalidArithmetic("sub", typeOf(a), typeOf(b)))
}

func aMul(a, b any) any {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa * fb
	}

	va, ok3 := a.(Vector)
	vb, ok4 := b.(Vector)
	if ok3 && ok4 {
		return Vector{va[0] * vb[0], va[1] * vb[1], va[2] * vb[2], va[3] * vb[3]}
	} else if ok1 && ok4 {
		f := float32(fa)
		return Vector{f * vb[0], f * vb[1], f * vb[2], f * vb[3]}
	} else if ok3 && ok2 {
		f := float32(fb)
		return Vector{va[0] * f, va[1] * f, va[2] * f, va[3] * f}
	}

	panic(invalidArithmetic("mul", typeOf(a), typeOf(b)))
}

func aDiv(a, b any) any {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa / fb
	}

	va, ok3 := a.(Vector)
	vb, ok4 := b.(Vector)
	if ok3 && ok4 {
		return Vector{va[0] / vb[0], va[1] / vb[1], va[2] / vb[2], va[3] / vb[3]}
	} else if ok1 && ok4 {
		f := float32(fa)
		return Vector{f / vb[0], f / vb[1], f / vb[2], f / vb[3]}
	} else if ok3 && ok2 {
		f := float32(fb)
		return Vector{va[0] / f, va[1] / f, va[2] / f, va[3] / f}
	}

	panic(invalidArithmetic("div", typeOf(a), typeOf(b)))
}

func aMod(a, b any) float64 {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa - fb*math.Floor(fa/fb)
	}

	panic(invalidArithmetic("mod", typeOf(a), typeOf(b)))
}

func aPow(a, b any) float64 {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return math.Pow(fa, fb)
	}

	panic(invalidArithmetic("pow", typeOf(a), typeOf(b)))
}

func aIdiv(a, b any) any {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return math.Floor(fa / fb)
	}

	va, ok3 := a.(Vector)
	vb, ok4 := b.(Vector)
	if ok3 && ok4 {
		return Vector{
			fFloor(va[0] / vb[0]),
			fFloor(va[1] / vb[1]),
			fFloor(va[2] / vb[2]),
			fFloor(va[3] / vb[3]),
		}
	} else if ok1 && ok4 {
		f := float32(fa)
		return Vector{
			fFloor(f / vb[0]),
			fFloor(f / vb[1]),
			fFloor(f / vb[2]),
			fFloor(f / vb[3]),
		}
	} else if ok3 && ok2 {
		f := float32(fb)
		return Vector{
			fFloor(va[0] / f),
			fFloor(va[1] / f),
			fFloor(va[2] / f),
			fFloor(va[3] / f),
		}
	}

	panic(invalidArithmetic("idiv", typeOf(a), typeOf(b)))
}

func jumpLe(a, b any) bool {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa <= fb
	}

	sa, ok1 := a.(string)
	sb, ok2 := b.(string)
	if ok1 && ok2 {
		return sa <= sb
	}

	panic(invalidCompare("<=", typeOf(a), typeOf(b)))
}

func jumpLt(a, b any) bool {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa < fb
	}

	sa, ok1 := a.(string)
	sb, ok2 := b.(string)
	if ok1 && ok2 {
		return sa < sb
	}

	panic(invalidCompare("<", typeOf(a), typeOf(b)))
}

func jumpGt(a, b any) bool {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa > fb
	}

	sa, ok1 := a.(string)
	sb, ok2 := b.(string)
	if ok1 && ok2 {
		return sa > sb
	}

	panic(invalidCompare(">", typeOf(a), typeOf(b)))
}

func jumpGe(a, b any) bool {
	fa, ok1 := a.(float64)
	fb, ok2 := b.(float64)
	if ok1 && ok2 {
		return fa >= fb
	}

	sa, ok1 := a.(string)
	sb, ok2 := b.(string)
	if ok1 && ok2 {
		return sa >= sb
	}

	panic(invalidCompare(">=", typeOf(a), typeOf(b)))
}

func jumpEq(a, b any) bool {
	switch a.(type) {
	case float64, string, bool, nil:
		return a == b // JUMPIFEQ
	}

	panic(incomparableType(typeOf(a), true)) // Also deliberately restricting the ability to compare types that would always return false
}

func execute(
	proto Proto,
	upvals []Upval,
	stack *[]any,
	co *Coroutine,
	varargs Varargs,
	alive *bool,
	protolist []Proto,
	env map[any]any,
	requireCache map[string]Rets,
) []any {
	protos, code := proto.protos, proto.code
	top, pc, openUpvalues, generalisedIterators := -1, 1, new([]*Upval), map[Inst]*Iterator{}

	var handlingBreak bool
	var inst Inst
	var op uint8

	// a a a a
	// stayin' alive
	for *alive {
		if !handlingBreak {
			inst = *code[pc-1]
			op = inst.opcode
		}
		handlingBreak = false

		// fmt.Println("OP", op, "PC", pc+1)

		switch op {
		case 0: // NOP
			pc += 1
			// -- Do nothing
		case 1: // BREAK
			op = proto.debugcode[pc]
			handlingBreak = true
		case 2: // LOADNIL
			pc += 1
			(*stack)[inst.A] = nil
		case 3: // LOADB
			pc += 1
			(*stack)[inst.A] = inst.B == 1
			pc += inst.C
		case 4: // LOADN
			pc += 1
			(*stack)[inst.A] = float64(inst.D) // never put an int on the stack
		case 5: // LOADK
			pc += 1
			// fmt.Println("LOADK", inst.A, inst.K)
			(*stack)[inst.A] = inst.K
		case 6: // MOVE
			pc += 1
			// we should (ALMOST) never have to change the size of the stack (proto.maxstacksize)
			(*stack)[inst.A] = (*stack)[inst.B]
		case 7: // GETGLOBAL
			kv := inst.K

			if Extensions[kv] != nil {
				(*stack)[inst.A] = Extensions[kv]
			} else {
				(*stack)[inst.A] = env[kv]
			}

			pc += 2 // -- adjust for aux
		case 8: // SETGLOBAL
			// LOL
			kv := inst.K
			if _, ok := kv.(string); ok {
				if Extensions[kv] != nil {
					panic(fmt.Sprintf("attempt to redefine global '%s'", kv))
				}
				panic(fmt.Sprintf("attempt to set global '%s'", kv))
			}
		case 9: // GETUPVAL
			pc += 1
			if uv := upvals[inst.B]; uv.selfRef {
				(*stack)[inst.A] = uv.value
			} else {
				// fmt.Println("GETTING UPVAL", uv.store)

				(*stack)[inst.A] = (*uv.store)[uv.index]
			}
		case 10: // SETUPVAL
			pc += 1
			if uv := upvals[inst.B]; !uv.selfRef {
				(*uv.store)[uv.index] = (*stack)[inst.A]
			}
		case 11: // CLOSEUPVALS
			pc += 1
			for i, uv := range *openUpvalues {
				if uv == nil || uv.selfRef || uv.index < inst.A {
					continue
				}
				uv.value = (*uv.store)[uv.index]
				uv.store = nil
				uv.selfRef = true
				(*openUpvalues)[i] = nil
			}
		case 12: // GETIMPORT
			k0 := inst.K0
			imp := Extensions[k0]
			if imp == nil {
				imp = env[k0]
			}

			// fmt.Println("IMPORTING", k0)

			switch inst.KC { // count
			case 1:
				// fmt.Println("GETIMPORT1", inst.A, imp)
				(*stack)[inst.A] = imp
			case 2:
				t := imp.(*Table)
				// fmt.Println("GETIMPORT2", inst.A, t.Get(inst.K1))
				(*stack)[inst.A] = t.Get(inst.K1)
			case 3:
				t := imp.(*Table)
				// fmt.Println("GETIMPORT3", inst.A, t.Get(inst.K1).([]any)[inst.K2.(uint32)-1])
				(*stack)[inst.A] = t.Get(inst.K1).([]any)[inst.K2.(uint32)-1]
			}

			pc += 2 // -- adjust for aux
		case 13: // GETTABLE
			pc += 1
			index := (*stack)[inst.C]
			t, ok := (*stack)[inst.B].(*Table)
			if !ok {
				panic(invalidIndex(typeOf((*stack)[inst.B]), index))
			}

			(*stack)[inst.A] = t.Get(index)
		case 14: // SETTABLE
			pc += 1
			index := (*stack)[inst.C]
			t, ok := (*stack)[inst.B].(*Table)
			if !ok {
				panic(invalidIndex(typeOf((*stack)[inst.B]), index))
			}

			// fmt.Println("SETTABLE", index, (*stack)[inst.A])
			t.Set(index, (*stack)[inst.A])
		case 15: // GETTABLEKS
			index := inst.K
			t, ok := (*stack)[inst.B].(*Table)
			if !ok {
				fmt.Println("indexing", typeOf((*stack)[inst.B]), "with", index)
				panic(invalidIndex(typeOf((*stack)[inst.B]), index))
			}

			(*stack)[inst.A] = t.Get(index)

			pc += 2 // -- adjust for aux
		case 16: // SETTABLEKS
			index := inst.K
			t, ok := (*stack)[inst.B].(*Table)
			if !ok {
				fmt.Println("indexing", typeOf((*stack)[inst.B]), "with", index)
				panic(invalidIndex(typeOf((*stack)[inst.B]), index))
			}

			t.Set(index, (*stack)[inst.A])

			pc += 2 // -- adjust for aux
		case 17: // GETTABLEN
			t := (*stack)[inst.B].(*Table)
			i := uint(inst.C + 1)

			if v := t.GetArray(i); v != nil {
				(*stack)[inst.A] = v
			} else if t.hash == nil {
				(*stack)[inst.A] = nil
			} else {
				(*stack)[inst.A] = (*t.hash)[float64(i)]
			}

			pc += 1
		case 18: // SETTABLEN
			t := (*stack)[inst.B].(*Table)
			if t.readonly {
				panic("attempt to modify a readonly table")
			} else if i, v := inst.C+1, (*stack)[inst.A]; 1 <= i || i > int(t.asize) {
				t.SetArray(uint(i), v)
			} else {
				t.SetHash(float64(i), v)
			}

			pc += 1
		case 19: // NEWCLOSURE
			newPrototype := protolist[protos[inst.D]-1]

			nups := newPrototype.nups
			upvalues := make([]Upval, nups)
			(*stack)[inst.A] = wrapclosure(newPrototype, upvalues, alive, protolist, env, requireCache)

			// fmt.Println("nups", nups)
			for i := range nups {
				switch pseudo := code[pc]; pseudo.A {
				case 0: // -- value
					upvalue := Upval{
						value:   (*stack)[pseudo.B],
						selfRef: true,
					}
					upvalue.store = nil

					upvalues[i] = upvalue
				case 1: // -- reference
					index := pseudo.B
					// fmt.Println("index", index, len(*open_upvalues))

					var prev *Upval
					if index < len(*openUpvalues) {
						prev = (*openUpvalues)[index]
					}

					if prev == nil {
						prev = &Upval{
							store: stack,
							index: index,
						}

						for len(*openUpvalues) <= index {
							*openUpvalues = append(*openUpvalues, nil)
						}
						(*openUpvalues)[index] = prev
					}

					upvalues[i] = *prev
				case 2: // -- upvalue
					upvalues[i] = upvals[pseudo.B]
				}
				pc += 1
			}
			pc += 1
		case 20: // NAMECALL
			pc += 1
			// fmt.Println("NAMECALL")

			A, B := inst.A, inst.B
			kv := inst.K.(string)
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

			ok, retList := NamecallHandler(co, kv, stack, callA+1, callA+params)
			if !ok {
				t := (*stack)[B].(*Table)

				if t.hash == nil {
					(*stack)[A] = nil
				} else {
					(*stack)[A] = (*t.hash)[kv]
				}
				break
			}

			pc += 2 // -- adjust for aux, Skip next CALL instruction

			inst = *callInst
			op = callOp

			retCount := len(retList)

			if callC == 0 {
				top = callA + retCount - 1
			} else {
				retCount = callC - 1
			}

			move(retList, 0, retCount, callA, stack)
		case 21: // CALL
			pc += 1
			A, B, C := inst.A, inst.B, inst.C

			var params int
			if B == 0 {
				params = top - A
			} else {
				params = B - 1
			}

			// fmt.Println(A, B, C, (*stack)[A], params)

			f := (*stack)[A]
			fn, ok := f.(*Function)
			// fmt.Println("calling with", (*stack)[A+1:A+params+1])
			if !ok {
				panic(uncallableType(typeOf(f)))
			}

			retList := (*fn)(co, (*stack)[A+1:A+params+1]...) // not inclusive
			retCount := len(retList)

			// fmt.Println("COUNT", retCount)
			if retCount == 1 {
				if p, ok := retList[0].(LoadParams); ok {
					// it's a require
					if c, ok := requireCache[p.path]; ok {
						retList = c
					} else {
						c2, _ := Load(p.deserialised, p.path, p.o, p.env, requireCache)
						result := c2.Resume()

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

			move(retList, 0, retCount, A, stack)
		case 22: // RETURN
			pc += 1
			A, B := inst.A, inst.B
			b := B - 1

			// nresults
			if b == LUAU_MULTRET {
				b = top - A + 1
			}

			return (*stack)[A:max(A+b, 0)]
		case 23, 24: // JUMP, JUMPBACK
			pc += inst.D + 1
		case 25: // JUMPIF
			if truthy((*stack)[inst.A]) {
				pc += inst.D + 1
			} else {
				pc += 1
			}
		case 26: // JUMPIFNOT
			if !truthy((*stack)[inst.A]) {
				pc += inst.D + 1
			} else {
				pc += 1
			}
		case 27: // jump
			if jumpEq((*stack)[inst.A], (*stack)[inst.aux]) {
				pc += inst.D + 1
			} else {
				pc += 2
			}
		case 28:
			if jumpLe((*stack)[inst.A], (*stack)[inst.aux]) {
				pc += inst.D + 1
			} else {
				pc += 2
			}
		case 29:
			if jumpLt((*stack)[inst.A], (*stack)[inst.aux]) {
				pc += inst.D + 1
			} else {
				pc += 2
			}
		case 30:
			if jumpEq((*stack)[inst.A], (*stack)[inst.aux]) {
				pc += 2
			} else {
				pc += inst.D + 1
			}
		case 31:
			if jumpGt((*stack)[inst.A], (*stack)[inst.aux]) {
				pc += inst.D + 1
			} else {
				pc += 2
			}
		case 32:
			if jumpGe((*stack)[inst.A], (*stack)[inst.aux]) {
				pc += inst.D + 1
			} else {
				pc += 2
			}
		case 33: // arithmetic
			pc += 1
			(*stack)[inst.A] = aAdd((*stack)[inst.B], (*stack)[inst.C])
		case 34:
			pc += 1
			(*stack)[inst.A] = aSub((*stack)[inst.B], (*stack)[inst.C])
		case 35:
			pc += 1
			(*stack)[inst.A] = aMul((*stack)[inst.B], (*stack)[inst.C])
		case 36:
			pc += 1
			(*stack)[inst.A] = aDiv((*stack)[inst.B], (*stack)[inst.C])
		case 37:
			pc += 1
			(*stack)[inst.A] = aMod((*stack)[inst.B], (*stack)[inst.C])
		case 38:
			pc += 1
			(*stack)[inst.A] = aPow((*stack)[inst.B], (*stack)[inst.C])
		case 81:
			pc += 1
			(*stack)[inst.A] = aIdiv((*stack)[inst.B], (*stack)[inst.C])
		case 39: // arithmetik
			pc += 1
			(*stack)[inst.A] = aAdd((*stack)[inst.B], inst.K)
		case 40:
			pc += 1
			(*stack)[inst.A] = aSub((*stack)[inst.B], inst.K)
		case 41:
			pc += 1
			(*stack)[inst.A] = aMul((*stack)[inst.B], inst.K)
		case 42:
			pc += 1
			(*stack)[inst.A] = aDiv((*stack)[inst.B], inst.K)
		case 43:
			pc += 1
			(*stack)[inst.A] = aMod((*stack)[inst.B], inst.K)
		case 44:
			pc += 1
			(*stack)[inst.A] = aPow((*stack)[inst.B], inst.K)
		case 82:
			pc += 1
			(*stack)[inst.A] = aIdiv((*stack)[inst.B], inst.K)

		case 45: // logic AND
			pc += 1
			a := (*stack)[inst.B]
			b := (*stack)[inst.C]

			if !truthy(a) {
				(*stack)[inst.A] = false
			} else if truthy(b) {
				(*stack)[inst.A] = b
			} else {
				(*stack)[inst.A] = false
			}
		case 46: // logic OR
			pc += 1
			a := (*stack)[inst.B]
			b := (*stack)[inst.C]

			if truthy(a) {
				(*stack)[inst.A] = a
			} else if truthy(b) {
				(*stack)[inst.A] = b
			} else {
				(*stack)[inst.A] = false
			}
		case 47: // logik AND
			pc += 1
			// fmt.Println("LOGIK")
			a := (*stack)[inst.B]
			b := inst.K

			if !truthy(a) {
				(*stack)[inst.A] = false
			} else if truthy(b) {
				(*stack)[inst.A] = b
			} else {
				(*stack)[inst.A] = false
			}
		case 48: // logik OR
			pc += 1
			// fmt.Println("LOGIK")
			a := (*stack)[inst.B]
			b := inst.K

			if truthy(a) {
				(*stack)[inst.A] = a
			} else if truthy(b) {
				(*stack)[inst.A] = b
			} else {
				(*stack)[inst.A] = false
			}

		case 49: // CONCAT
			pc += 1
			s := strings.Builder{}

			var first int
			for i := inst.B; i <= inst.C; i++ {
				toWrite, ok := (*stack)[i].(string)
				if !ok {
					// ensure correct order of operands in error message
					panic(invalidConcat(typeOf((*stack)[i+first]), typeOf((*stack)[i+1+first])))
				}
				s.WriteString(toWrite)
				first = -1
			}
			(*stack)[inst.A] = s.String()
		case 50: // NOT
			pc += 1
			cond, ok := (*stack)[inst.B].(bool)
			if !ok {
				panic(invalidCond(typeOf((*stack)[inst.B])))
			}

			(*stack)[inst.A] = !cond
		case 51: // MINUS
			pc += 1
			a, ok := (*stack)[inst.B].(float64)
			if !ok {
				panic(invalidUnm(typeOf((*stack)[inst.B])))
			}

			(*stack)[inst.A] = -a
		case 52: // LENGTH
			pc += 1
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

			pc += 2 // -- adjust for aux
		case 54: // DUPTABLE
			pc += 1
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

			pc += 2 // -- adjust for aux
		case 56: // FORNPREP
			pc += 1
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
			pc += 1
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
				// fmt.Println("IT func", it, (*stack)[A+1], (*stack)[A+2])
				vals := (*it)(co, (*stack)[A+1], (*stack)[A+2])

				move(vals, 0, res, A+3, stack)

				// fmt.Println(A+3, (*stack)[A+3])

				if (*stack)[A+3] != nil {
					(*stack)[A+2] = (*stack)[A+3]
					pc += inst.D + 1
				} else {
					pc += 2
				}
			default:
				iter := *generalisedIterators[inst]

				if !iter.running {
					args := &[]any{it, (*stack)[A+1], (*stack)[A+2]}
					// fmt.Println("-1- sending thru the wire", args)
					iter.args <- args
					// fmt.Println("-1- sent")
				}
				vals := <-iter.resume
				// fmt.Println("-1- received!", vals)

				if vals == nil {
					delete(generalisedIterators, inst)
					pc += 2
				} else {
					move(*vals, 0, res, A+3, stack)

					(*stack)[A+2] = (*stack)[A+3]
					pc += inst.D + 1
				}
			}
		case 59, 61: // FORGPREP_INEXT, FORGPREP_NEXT
			if _, ok := (*stack)[inst.A].(*Function); !ok {
				panic(fmt.Sprintf("attempt to iterate over a %s value", typeOf((*stack)[inst.A]))) // -- encountered non-function value
			}
			pc += inst.D + 1
		case 60: // FASTCALL3
			// Skipped
			pc += 2 // adjust for aux
		case 63: // GETVARARGS
			pc += 1
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
			pc += 1
			newPrototype := protolist[inst.K.(uint32)]

			nups := newPrototype.nups
			upvalues := make([]Upval, nups)
			(*stack)[inst.A] = wrapclosure(newPrototype, upvalues, alive, protolist, env, requireCache)

			for i := range nups {
				pseudo := code[pc-1]
				pc += 1

				t := pseudo.A
				if t == 0 { // value
					upvalue := Upval{
						value:   (*stack)[pseudo.B],
						selfRef: true,
					}
					upvalue.store = nil
					upvalues[i] = upvalue

					// -- references dont get handled by DUPCLOSURE
				} else if t == 2 { // upvalue
					upvalues[i] = upvals[pseudo.B]
				}
			}
		case 65: // PREPVARARGS
			pc += 1
			// Handled by wrapper
		case 66: // LOADKX
			(*stack)[inst.A] = inst.K.(uint32) // kv

			pc += 2 // -- adjust for aux
		case 67: // JUMPX
			pc += inst.E + 1
		case 68: // FASTCALL
			pc += 1
			// Skipped
		case 69: // COVERAGE
			pc += 1
			inst.E += 1
		case 70: // CAPTURE
			// Handled by CLOSURE
			panic("encountered unhandled CAPTURE")
		case 71: // SUBRK
			pc += 1
			(*stack)[inst.A] = aSub(inst.K, (*stack)[inst.C])
		case 72: // DIVRK
			pc += 1
			(*stack)[inst.A] = aDiv(inst.K, (*stack)[inst.C])
		case 73: // FASTCALL1
			pc += 1
			// Skipped
		case 74, 75: // FASTCALL2, FASTCALL2K
			// Skipped
			pc += 2 // adjust for aux
		case 76: // FORGPREP
			pc += inst.D + 1
			if _, ok := (*stack)[inst.A].(*Function); ok {
				break
			}

			loopInstruction := *code[pc-1]
			if generalisedIterators[loopInstruction] != nil {
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

			generalisedIterators[loopInstruction] = c
		case 77: // JUMPXEQKNIL
			ra := (*stack)[inst.A]

			if ra == nil != inst.KN {
				pc += inst.D + 1
			} else {
				pc += 2
			}
		case 78: //  JUMPXEQKB
			kv := inst.K.(bool)
			ra := (*stack)[inst.A].(bool)

			if ra == kv != inst.KN {
				pc += inst.D + 1
			} else {
				pc += 2
			}
		case 79: // JUMPXEQKN
			kv := inst.K.(float64)
			ra := (*stack)[inst.A].(float64)

			if ra == kv != inst.KN {
				pc += inst.D + 1
			} else {
				pc += 2
			}
		case 80: // JUMPXEQKS
			kv := inst.K.(string)
			ra := (*stack)[inst.A].(string)

			if ra == kv != inst.KN {
				pc += inst.D + 1
			} else {
				pc += 2
			}
		default:
			panic(fmt.Sprintf("Unsupported Opcode: %s op: %d", inst.opname, op))
		}
	}

	for i, uv := range *openUpvalues {
		if uv.selfRef {
			continue
		}
		uv.value = (*uv.store)[uv.index]
		uv.store = nil
		uv.selfRef = true
		(*openUpvalues)[i] = nil
	}

	for i := range generalisedIterators {
		generalisedIterators[i].running = false
		delete(generalisedIterators, i)
	}
	return []any{}
}

func wrapclosure(
	proto Proto,
	upvals []Upval,
	alive *bool,
	protolist []Proto,
	env map[any]any,
	requireCache map[string]Rets,
) *Function {
	wrapped := Function(func(co *Coroutine, passed ...any) []any {
		maxstacksize, numparams := proto.maxstacksize, proto.numparams

		stack := make([]any, maxstacksize)
		// fmt.Println("MAX STACK SIZE", maxstacksize)
		varargs := Varargs{list: []any{}}

		move(passed, 0, int(numparams), 0, &stack)

		n := uint8(len(passed))
		if numparams < n {
			start := int(numparams + 1)
			l := int(n - numparams)
			varargs.len = l

			// expand varargs list
			varargs.list = make([]any, l)

			move(passed, start-1, start+l-1, 0, &varargs.list)
		}

		// TODO: dee bugg ingg
		return execute(proto, upvals, &stack, co, varargs, alive, protolist, env, requireCache)
	})

	return &wrapped
}

func Load(module Deserialised, filepath string, o uint8, env map[any]any, requireCache map[string]Rets) (Coroutine, func()) {
	protolist := module.protoList
	alive := true

	return Coroutine{
		body:     wrapclosure(module.mainProto, []Upval{}, &alive, protolist, env, requireCache),
		env:      env,
		filepath: filepath,
		yield:    make(chan Rets, 1),
		resume:   make(chan Rets, 1),
		o:        o,
	}, func() { alive = false }
}
