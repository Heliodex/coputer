package types

// This file is the single accessor layer for Val. All code that reads or writes a value's underlying Go representation goes through here, so the concrete representation of Val can change without touching call sites.

// Markers for value kinds that have no payload to box. The same pointer is shared by every value of that kind, so a pointer compare identifies it and struct equality still compares by value.
var (
	numKind   = new(byte)
	trueKind  = new(byte)
	falseKind = new(byte)
)

// NilVal is the Luau nil value.
var NilVal Val

// Num returns a number value.
func Num(f float64) Val { return Val{Num: f, Ref: numKind} }

// IntVal returns an integer value. Luau integers come from `i`-suffixed literals.
func IntVal(i int64) Val { return Val{Ref: i} }

// Str returns a string value.
func Str(s string) Val { return Val{Ref: s} }

// Bool returns a boolean value.
func Bool(b bool) Val {
	if b {
		return Val{Ref: trueKind}
	}
	return Val{Ref: falseKind}
}

// TableVal returns a table value.
func TableVal(t *Table) Val { return Val{Ref: t} }

// FnVal returns a function value.
func FnVal(f Function) Val { return Val{Ref: f} }

// CoVal returns a coroutine value.
func CoVal(c *Coroutine) Val { return Val{Ref: c} }

// BufVal returns a buffer value.
func BufVal(b *Buffer) Val { return Val{Ref: b} }

// VecVal returns a vector value.
func VecVal(v Vector) Val { return Val{Ref: v} }

// ObjVal returns an opaque object as a value. It is only for values passed between VM internals, never for Luau code.
func ObjVal(x any) Val { return Val{Ref: x} }

// IsNil reports whether v is nil.
func IsNil(v Val) bool { return v.Ref == nil }

// IsNum reports whether v is a number.
func IsNum(v Val) bool { return v.Ref == numKind }

// IsInt reports whether v is an integer.
func IsInt(v Val) bool { _, ok := v.Ref.(int64); return ok }

// IsString reports whether v is a string.
func IsString(v Val) bool { _, ok := v.Ref.(string); return ok }

// IsBool reports whether v is a boolean.
func IsBool(v Val) bool { return v.Ref == trueKind || v.Ref == falseKind }

// IsTable reports whether v is a table.
func IsTable(v Val) bool { _, ok := v.Ref.(*Table); return ok }

// IsFunction reports whether v is a function.
func IsFunction(v Val) bool { _, ok := v.Ref.(Function); return ok }

// IsCoroutine reports whether v is a coroutine.
func IsCoroutine(v Val) bool { _, ok := v.Ref.(*Coroutine); return ok }

// IsBuffer reports whether v is a buffer.
func IsBuffer(v Val) bool { _, ok := v.Ref.(*Buffer); return ok }

// IsVector reports whether v is a vector.
func IsVector(v Val) bool { _, ok := v.Ref.(Vector); return ok }

// AsNum returns v as a number, and whether it was one.
func AsNum(v Val) (float64, bool) {
	if v.Ref == numKind {
		return v.Num, true
	}
	return 0, false
}

// AsInt returns v as an integer, and whether it was one.
func AsInt(v Val) (int64, bool) { i, ok := v.Ref.(int64); return i, ok }

// AsString returns v as a string, and whether it was one.
func AsString(v Val) (string, bool) { s, ok := v.Ref.(string); return s, ok }

// AsBool returns v as a boolean, and whether it was one.
func AsBool(v Val) (bool, bool) {
	switch v.Ref {
	case trueKind:
		return true, true
	case falseKind:
		return false, true
	}
	return false, false
}

// AsTable returns v as a table, and whether it was one.
func AsTable(v Val) (*Table, bool) { t, ok := v.Ref.(*Table); return t, ok }

// AsFunction returns v as a function, and whether it was one.
func AsFunction(v Val) (Function, bool) { f, ok := v.Ref.(Function); return f, ok }

// AsCoroutine returns v as a coroutine, and whether it was one.
func AsCoroutine(v Val) (*Coroutine, bool) { c, ok := v.Ref.(*Coroutine); return c, ok }

// AsBuffer returns v as a buffer, and whether it was one.
func AsBuffer(v Val) (*Buffer, bool) { b, ok := v.Ref.(*Buffer); return b, ok }

// AsVector returns v as a vector, and whether it was one.
func AsVector(v Val) (Vector, bool) { x, ok := v.Ref.(Vector); return x, ok }

// AsObject returns the boxed payload of a value built with ObjVal. It is only for VM-internal objects, never for Luau values.
func AsObject(v Val) any { return v.Ref }

// MustNum returns v as a number, panicking if it is not one. Use it only where the type is already guaranteed.
func MustNum(v Val) float64 { return v.Num }

// MustInt returns v as an integer, panicking if it is not one. Use it only where the type is already guaranteed.
func MustInt(v Val) int64 { return v.Ref.(int64) }

// MustString returns v as a string, panicking if it is not one. Use it only where the type is already guaranteed.
func MustString(v Val) string { return v.Ref.(string) }

// MustBool returns v as a boolean, panicking if it is not one. Use it only where the type is already guaranteed.
func MustBool(v Val) bool {
	b, _ := AsBool(v)
	return b
}

// MustTable returns v as a table, panicking if it is not one. Use it only where the type is already guaranteed.
func MustTable(v Val) *Table { return v.Ref.(*Table) }

// MustFunction returns v as a function, panicking if it is not one. Use it only where the type is already guaranteed.
func MustFunction(v Val) Function { return v.Ref.(Function) }

// MustCoroutine returns v as a coroutine, panicking if it is not one. Use it only where the type is already guaranteed.
func MustCoroutine(v Val) *Coroutine { return v.Ref.(*Coroutine) }

// MustBuffer returns v as a buffer, panicking if it is not one. Use it only where the type is already guaranteed.
func MustBuffer(v Val) *Buffer { return v.Ref.(*Buffer) }

// MustVector returns v as a vector, panicking if it is not one. Use it only where the type is already guaranteed.
func MustVector(v Val) Vector { return v.Ref.(Vector) }
