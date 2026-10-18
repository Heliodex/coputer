package std

import (
	"fmt"
	"math"

	. "github.com/Heliodex/coputer/litecode/types"
)

func invalidArithmetic(op, ta, tb string) error {
	return fmt.Errorf("attempt to perform arithmetic (%s) on %s and %s", op, ta, tb)
}

func invalidCompare(op, ta, tb string) error {
	return fmt.Errorf("attempt to compare %s %s %s", ta, op, tb)
}

func invalidUnm(t string) error {
	return fmt.Errorf("attempt to perform arithmetic (unm) on %s", t)
}

func Add(a, b Val) (Val, error) {
	fa, ok1 := AsNum(a)
	fb, ok2 := AsNum(b)
	if ok1 && ok2 {
		return Num(fa + fb), nil
	}

	va, ok3 := AsVector(a)
	vb, ok4 := AsVector(b)
	if ok3 && ok4 {
		return VecVal(Vector{va[0] + vb[0], va[1] + vb[1], va[2] + vb[2], va[3] + vb[3]}), nil
	}

	return NilVal, invalidArithmetic("add", TypeOf(a), TypeOf(b))
}

func Sub(a, b Val) (Val, error) {
	fa, ok1 := AsNum(a)
	fb, ok2 := AsNum(b)
	if ok1 && ok2 {
		return Num(fa - fb), nil
	}

	va, ok3 := AsVector(a)
	vb, ok4 := AsVector(b)
	if ok3 && ok4 {
		return VecVal(Vector{va[0] - vb[0], va[1] - vb[1], va[2] - vb[2], va[3] - vb[3]}), nil
	}

	return NilVal, invalidArithmetic("sub", TypeOf(a), TypeOf(b))
}

func Mul(a, b Val) (Val, error) {
	fa, ok1 := AsNum(a)
	fb, ok2 := AsNum(b)
	if ok1 && ok2 {
		return Num(fa * fb), nil
	}

	va, ok3 := AsVector(a)
	vb, ok4 := AsVector(b)
	switch {
	case ok3 && ok4:
		return VecVal(Vector{va[0] * vb[0], va[1] * vb[1], va[2] * vb[2], va[3] * vb[3]}), nil
	case ok1 && ok4:
		f := float32(fa)
		return VecVal(Vector{f * vb[0], f * vb[1], f * vb[2], f * vb[3]}), nil
	case ok3 && ok2:
		f := float32(fb)
		return VecVal(Vector{va[0] * f, va[1] * f, va[2] * f, va[3] * f}), nil
	}

	return NilVal, invalidArithmetic("mul", TypeOf(a), TypeOf(b))
}

func Div(a, b Val) (Val, error) {
	fa, ok1 := AsNum(a)
	fb, ok2 := AsNum(b)
	if ok1 && ok2 {
		return Num(fa / fb), nil
	}

	va, ok3 := AsVector(a)
	vb, ok4 := AsVector(b)
	switch {
	case ok3 && ok4:
		return VecVal(Vector{va[0] / vb[0], va[1] / vb[1], va[2] / vb[2], va[3] / vb[3]}), nil
	case ok1 && ok4:
		f := float32(fa)
		return VecVal(Vector{f / vb[0], f / vb[1], f / vb[2], f / vb[3]}), nil
	case ok3 && ok2:
		f := float32(fb)
		return VecVal(Vector{va[0] / f, va[1] / f, va[2] / f, va[3] / f}), nil
	}

	return NilVal, invalidArithmetic("div", TypeOf(a), TypeOf(b))
}

func Mod(a, b Val) (Val, error) {
	fa, ok1 := AsNum(a)
	fb, ok2 := AsNum(b)
	if ok1 && ok2 {
		return Num(fa - fb*math.Floor(fa/fb)), nil
	}

	return NilVal, invalidArithmetic("mod", TypeOf(a), TypeOf(b))
}

func Pow(a, b Val) (Val, error) {
	fa, ok1 := AsNum(a)
	fb, ok2 := AsNum(b)
	if ok1 && ok2 {
		return Num(math.Pow(fa, fb)), nil
	}

	return NilVal, invalidArithmetic("pow", TypeOf(a), TypeOf(b))
}

func Idiv(a, b Val) (Val, error) {
	fa, ok1 := AsNum(a)
	fb, ok2 := AsNum(b)
	if ok1 && ok2 {
		return Num(math.Floor(fa / fb)), nil
	}

	va, ok3 := AsVector(a)
	vb, ok4 := AsVector(b)
	switch {
	case ok3 && ok4:
		return VecVal(Vector{
			f32Floor(va[0] / vb[0]),
			f32Floor(va[1] / vb[1]),
			f32Floor(va[2] / vb[2]),
			f32Floor(va[3] / vb[3]),
		}), nil
	case ok1 && ok4:
		f := float32(fa)
		return VecVal(Vector{
			f32Floor(f / vb[0]),
			f32Floor(f / vb[1]),
			f32Floor(f / vb[2]),
			f32Floor(f / vb[3]),
		}), nil
	case ok3 && ok2:
		f := float32(fb)
		return VecVal(Vector{
			f32Floor(va[0] / f),
			f32Floor(va[1] / f),
			f32Floor(va[2] / f),
			f32Floor(va[3] / f),
		}), nil
	}

	return NilVal, invalidArithmetic("idiv", TypeOf(a), TypeOf(b))
}

func Unm(a Val) (Val, error) {
	fa, ok1 := AsNum(a)
	if ok1 {
		return Num(-fa), nil
	}

	va, ok2 := AsVector(a)
	if ok2 {
		return VecVal(Vector{-va[0], -va[1], -va[2], -va[3]}), nil
	}

	return NilVal, invalidUnm(TypeOf(a))
}

// vectors dont have these comparisons
func Le(a, b Val) (bool, error) {
	fa, ok1 := AsNum(a)
	fb, ok2 := AsNum(b)
	if ok1 && ok2 {
		return fa <= fb, nil
	}

	sa, ok1 := AsString(a)
	sb, ok2 := AsString(b)
	if ok1 && ok2 {
		return sa <= sb, nil
	}

	return false, invalidCompare("<=", TypeOf(a), TypeOf(b))
}

func Lt(a, b Val) (bool, error) {
	fa, ok1 := AsNum(a)
	fb, ok2 := AsNum(b)
	if ok1 && ok2 {
		return fa < fb, nil
	}

	sa, ok1 := AsString(a)
	sb, ok2 := AsString(b)
	if ok1 && ok2 {
		return sa < sb, nil
	}

	return false, invalidCompare("<", TypeOf(a), TypeOf(b))
}

func Gt(a, b Val) (bool, error) {
	fa, ok1 := AsNum(a)
	fb, ok2 := AsNum(b)
	if ok1 && ok2 {
		return fa > fb, nil
	}

	sa, ok1 := AsString(a)
	sb, ok2 := AsString(b)
	if ok1 && ok2 {
		return sa > sb, nil
	}

	return false, invalidCompare(">", TypeOf(a), TypeOf(b))
}

func Ge(a, b Val) (bool, error) {
	fa, ok1 := AsNum(a)
	fb, ok2 := AsNum(b)
	if ok1 && ok2 {
		return fa >= fb, nil
	}

	sa, ok1 := AsString(a)
	sb, ok2 := AsString(b)
	if ok1 && ok2 {
		return sa >= sb, nil
	}

	return false, invalidCompare(">=", TypeOf(a), TypeOf(b))
}
