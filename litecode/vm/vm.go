// package vm implements Litecode's deterministic Luau virtual machine and standard library toolkit.
package vm

import (
	"errors"
	"fmt"
	"strings"

	// unsafe code paths removed... for now

	"github.com/Heliodex/coputer/litecode/internal"
	. "github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/litecode/vm/compile"
	"github.com/Heliodex/coputer/litecode/vm/std"
)

var (
	errReadonly = errors.New("attempt to modify a readonly table")
	errNilIndex = errors.New("table index is nil")
)

// Functions and Tables are used as pointers normally, as they need to be hashed

func fn(name string, co *Coroutine, f func(co *Coroutine, args ...Val) (r []Val, err error)) Function {
	return Function{
		Run:  &f,
		Name: name,
		Co:   co,
	}
}

const luau_multret = -1

func namecallHandler(co *Coroutine, kv string, stack *[]Val, c1, c2 int32) (ok bool, retList []Val, err error) {
	switch kv {
	case "format":
		str := (*stack)[c1].(string)
		args := (*stack)[1:][c1:c2]

		f, err := std.Fmtstring(str, std.Args{Co: co, List: args, Name: "format"})
		if err != nil {
			return false, nil, err
		}
		return true, []Val{f}, nil
	}
	return
}

// The only global environment, never mutated (yes, it only stores string keys)
var exts = map[string]Val{
	"args":      std.Libargs, // custom
	"bit32":     std.Libbit32,
	"buffer":    std.Libbuffer,
	"coroutine": std.Libcoroutine,
	"math":      std.Libmath,
	"string":    std.Libstring,
	"table":     std.Libtable,
	"utf8":      std.Libutf8,
	"vector":    std.Libvector,
	// fuck os
	// and debug

	"_VERSION": "Luau", // todo: custom
}

func init() {
	for _, g := range std.Globals {
		exts[g.Name] = g
	}
}

// var VectorSize = 4
// var AllowProxyErrors = false

type upval struct {
	Val
	store *Val
}

func falsy(v Val) bool {
	return v == nil || v == false
}

func uncallableType(v string) error {
	return fmt.Errorf("attempt to call a %s value", v)
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
	tb := std.TypeOf(v)
	if tb == "string" {
		tb = fmt.Sprintf("'%v'", v)
	}

	return fmt.Errorf("attempt to index %v with %v", ta, tb)
}

func invalidIter(t string) error {
	return fmt.Errorf("attempt to iterate over a %s value", t)
}

func missingMethod(ta string, v Val) error {
	tb := std.TypeOf(v)
	if tb == "string" {
		tb = fmt.Sprintf("'%v'", v)
	}

	return fmt.Errorf("attempt to call missing method %v of %v", tb, ta)
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

	return nil, invalidIndex(std.TypeOf(v), k)
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
		return invalidIndex(std.TypeOf(imp), K2)
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
				return missingMethod(std.TypeOf(t), kv)
			}

			(*stack)[i.A] = call
		case string:
			return missingMethod(std.TypeOf(t), kv)
		default:
			return invalidIndex(std.TypeOf(t), kv)
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

func handleRequire(towrap toWrap, p compile.Program, co *Coroutine) (rets []Val, err error) {
	if c, ok := towrap.requireCache[p.Filepath]; ok {
		return []Val{c}, nil
	}

	// since environments only store global libraries etc, using the same env here should be fine??
	c2, _ := loadmodule(p, co.Env, towrap.requireCache, co.ProgramArgs)
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

	towrap.requireCache[p.Filepath] = ret
	return []Val{ret}, nil
}

func call(top *int32, A int32, B, C uint8, towrap toWrap, stack *[]Val, co *Coroutine) (err error) {
	// fmt.Println(A, B, C, (*stack)[A], params)

	f := (*stack)[A]
	fn, ok := f.(Function)
	// fmt.Println("calling with", (*stack)[A+1:][:params])
	if !ok {
		return uncallableType(std.TypeOf(f))
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
		if p, ok := retList[0].(compile.Program); ok {
			// it's a require
			// fmt.Println("REQUIRE", lc.filepath)

			if retList, err = handleRequire(towrap, p, co); err != nil {
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
		// fmt.Println("GETTING GENITER", std.TypeOf(it))
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
		return invalidIter(std.TypeOf(s))
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
				return nil, invalidIndex(std.TypeOf(stack[i.B]), idx)
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
				// fmt.Println("indexing", std.TypeOf(stack[i.B]), "with", idx)
				return nil, invalidIndex(std.TypeOf(stack[i.B]), idx)
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
				return nil, invalidIndex(std.TypeOf(stack[i.B]), float64(idx))
			}

			stack[i.A] = t.GetInt(int(idx))
			pc++
		case 18: // SETTABLEN
			idx := i.C + 1
			t, ok := stack[i.B].(*Table)
			if !ok {
				// fmt.Println("gettableninvalidindex")
				return nil, invalidIndex(std.TypeOf(stack[i.B]), float64(idx))
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
			if j, err := std.Le(stack[i.A], stack[i.Aux]); err != nil {
				return nil, err
			} else if j {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 29:
			if j, err := std.Lt(stack[i.A], stack[i.Aux]); err != nil {
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
			if j, err := std.Gt(stack[i.A], stack[i.Aux]); err != nil {
				return nil, err
			} else if j {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 32:
			if j, err := std.Ge(stack[i.A], stack[i.Aux]); err != nil {
				return nil, err
			} else if j {
				pc += i.D + 1
			} else {
				pc += 2
			}
		case 33: // arithmetic
			if stack[i.A], err = std.Add(stack[i.B], stack[i.C]); err != nil {
				return
			}
			pc++
		case 34:
			if stack[i.A], err = std.Sub(stack[i.B], stack[i.C]); err != nil {
				return
			}
			pc++
		case 35:
			if stack[i.A], err = std.Mul(stack[i.B], stack[i.C]); err != nil {
				return
			}
			pc++
		case 36:
			if stack[i.A], err = std.Div(stack[i.B], stack[i.C]); err != nil {
				return
			}
			pc++
		case 37:
			if stack[i.A], err = std.Mod(stack[i.B], stack[i.C]); err != nil {
				return
			}
			pc++
		case 38:
			if stack[i.A], err = std.Pow(stack[i.B], stack[i.C]); err != nil {
				return
			}
			pc++
		case 81:
			if stack[i.A], err = std.Idiv(stack[i.B], stack[i.C]); err != nil {
				return
			}
			pc++
		case 39: // arithmetik
			if stack[i.A], err = std.Add(stack[i.B], i.K); err != nil {
				return
			}
			pc++
		case 40:
			if stack[i.A], err = std.Sub(stack[i.B], i.K); err != nil {
				return
			}
			pc++
		case 41:
			if stack[i.A], err = std.Mul(stack[i.B], i.K); err != nil {
				return
			}
			pc++
		case 42:
			if stack[i.A], err = std.Div(stack[i.B], i.K); err != nil {
				return
			}
			pc++
		case 43:
			if stack[i.A], err = std.Mod(stack[i.B], i.K); err != nil {
				return
			}
			pc++
		case 44:
			if stack[i.A], err = std.Pow(stack[i.B], i.K); err != nil {
				return
			}
			pc++
		case 82:
			if stack[i.A], err = std.Idiv(stack[i.B], i.K); err != nil {
				return
			}
			pc++
		case 71: // SUBRK
			if stack[i.A], err = std.Sub(i.K, stack[i.C]); err != nil {
				return
			}
			pc++
		case 72: // DIVRK
			if stack[i.A], err = std.Div(i.K, stack[i.C]); err != nil {
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
					return nil, invalidConcat(std.TypeOf(stack[n-first]), std.TypeOf(stack[n+1-first]))
				}
				b.WriteString(toWrite)
			}
			stack[i.A] = b.String()
			pc++
		case 50: // NOT
			stack[i.A] = falsy(stack[i.B])
			pc++
		case 51: // MINUS
			if stack[i.A], err = std.Unm(stack[i.B]); err != nil {
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
				return nil, invalidLength(std.TypeOf(t))
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
				return nil, invalidFor("initial value", std.TypeOf(stack[i.A+2]))
			}

			limit, ok := stack[i.A].(float64)
			if !ok {
				return nil, invalidFor("limit", std.TypeOf(stack[i.A]))
			}

			step, ok := stack[i.A+1].(float64)
			if !ok {
				return nil, invalidFor("step", std.TypeOf(stack[i.A+1]))
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
				return nil, invalidIter(std.TypeOf(stack[i.A])) // -- encountered non-function value
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

func loadmodule(p compile.Program, env Env, requireCache map[string]Val, args ProgramArgs) (co Coroutine, cancel func()) {
	alive := true

	towrap := toWrap{
		proto:        p.MainProto,
		protoList:    p.ProtoList,
		alive:        &alive,
		env:          env,
		requireCache: requireCache,
	}

	return Coroutine{
		Function:       wrapclosure(towrap, nil),
		Env:            env,
		Filepath:       p.Filepath,
		Dbgpath:        p.Dbgpath,
		RequireHistory: p.RequireHistory,
		YieldChan:      make(chan internal.Yield, 1),
		ResumeChan:     make(chan []Val, 1),
		Compiler:       p.Compiler,
		ProgramArgs:    args,
	}, func() { alive = false }
}

func Load(p compile.Program, env Env, args ProgramArgs) (co Coroutine, cancel func()) {
	return loadmodule(p, env, map[string]Val{}, args)
}
