package main

// opList contains information about the instruction, each instruction is defined in this format:
// { OP_NAME, OP_MODE, K_MODE, HAS_AUX }
// OP_MODE specifies what type of registers the instruction uses if any
//		0 = NONE
//		1 = A
//		2 = AB
//		3 = ABC
//		4 = AD
//		5 = AE
// K_MODE specifies if the instruction has a register that holds a constant table index, which will be directly converted to the constant in the 2nd pass
//		0 = NONE
//		1 = AUX
//		2 = C
//		3 = D
//		4 = AUX import
//		5 = AUX boolean low 1 bit
//		6 = AUX number low 24 bits
// HAS_AUX boolean specifies whether the instruction is followed up with an AUX word, which may be used to execute the instruction.

type Operator struct {
	Name   string
	Mode   uint8
	KMode  uint8
	HasAux bool
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

const (
	LUA_MULTRET                = -1
	LUA_GENERALISED_TERMINATOR = -2
)

type LuauSettings struct {
	VectorCtor       func()
	VectorSize       uint8
	Extensions       any
	AllowProxyErrors bool
	DecodeOp         func(op uint8) uint8
}

var luau_settings = LuauSettings{
	VectorCtor: func() {
		panic("vectorCtor was not provided")
	},
	VectorSize:       4,
	Extensions:       nil,
	AllowProxyErrors: false,
	DecodeOp: func(op uint8) uint8 {
		// println("decoding op", op)
		return op
	},
}

func main() {
	println("Hello, World!")
}
