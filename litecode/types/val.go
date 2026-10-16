package types

// This file is the single accessor layer for Val. All code that reads or writes a value's underlying Go representation goes through here, so the concrete representation of Val can change without touching call sites.

// NilVal is the Luau nil value.
var NilVal Val

// Num returns a number value.
func Num(f float64) Val { return f }

// IntVal returns an integer value. Luau integers come from `i`-suffixed literals.
func IntVal(i int64) Val { return i }

// Str returns a string value.
func Str(s string) Val { return s }

// Bool returns a boolean value.
func Bool(b bool) Val { return b }

// TableVal returns a table value.
func TableVal(t *Table) Val { return t }

// FnVal returns a function value.
func FnVal(f Function) Val { return f }

// CoVal returns a coroutine value.
func CoVal(c *Coroutine) Val { return c }

// BufVal returns a buffer value.
func BufVal(b *Buffer) Val { return b }

// VecVal returns a vector value.
func VecVal(v Vector) Val { return v }

// ObjVal returns an opaque object as a value. It is only for values passed between VM internals, never for Luau code.
func ObjVal(x any) Val { return x }

// IsNil reports whether v is nil.
func IsNil(v Val) bool { return v == nil }

// IsNum reports whether v is a number.
func IsNum(v Val) bool { _, ok := v.(float64); return ok }

// IsInt reports whether v is an integer.
func IsInt(v Val) bool { _, ok := v.(int64); return ok }

// IsString reports whether v is a string.
func IsString(v Val) bool { _, ok := v.(string); return ok }

// IsBool reports whether v is a boolean.
func IsBool(v Val) bool { _, ok := v.(bool); return ok }

// IsTable reports whether v is a table.
func IsTable(v Val) bool { _, ok := v.(*Table); return ok }

// IsFunction reports whether v is a function.
func IsFunction(v Val) bool { _, ok := v.(Function); return ok }

// IsCoroutine reports whether v is a coroutine.
func IsCoroutine(v Val) bool { _, ok := v.(*Coroutine); return ok }

// IsBuffer reports whether v is a buffer.
func IsBuffer(v Val) bool { _, ok := v.(*Buffer); return ok }

// IsVector reports whether v is a vector.
func IsVector(v Val) bool { _, ok := v.(Vector); return ok }

// AsNum returns v as a number, and whether it was one.
func AsNum(v Val) (float64, bool) { f, ok := v.(float64); return f, ok }

// AsInt returns v as an integer, and whether it was one.
func AsInt(v Val) (int64, bool) { i, ok := v.(int64); return i, ok }

// AsString returns v as a string, and whether it was one.
func AsString(v Val) (string, bool) { s, ok := v.(string); return s, ok }

// AsBool returns v as a boolean, and whether it was one.
func AsBool(v Val) (bool, bool) { b, ok := v.(bool); return b, ok }

// AsTable returns v as a table, and whether it was one.
func AsTable(v Val) (*Table, bool) { t, ok := v.(*Table); return t, ok }

// AsFunction returns v as a function, and whether it was one.
func AsFunction(v Val) (Function, bool) { f, ok := v.(Function); return f, ok }

// AsCoroutine returns v as a coroutine, and whether it was one.
func AsCoroutine(v Val) (*Coroutine, bool) { c, ok := v.(*Coroutine); return c, ok }

// AsBuffer returns v as a buffer, and whether it was one.
func AsBuffer(v Val) (*Buffer, bool) { b, ok := v.(*Buffer); return b, ok }

// AsVector returns v as a vector, and whether it was one.
func AsVector(v Val) (Vector, bool) { x, ok := v.(Vector); return x, ok }

// AsObject returns v as the opaque object it was built from with ObjVal, or itself if it was not one.
func AsObject(v Val) any { return v }

// MustNum returns v as a number, panicking if it is not one. Use it only where the type is already guaranteed.
func MustNum(v Val) float64 { return v.(float64) }

// MustInt returns v as an integer, panicking if it is not one. Use it only where the type is already guaranteed.
func MustInt(v Val) int64 { return v.(int64) }

// MustString returns v as a string, panicking if it is not one. Use it only where the type is already guaranteed.
func MustString(v Val) string { return v.(string) }

// MustBool returns v as a boolean, panicking if it is not one. Use it only where the type is already guaranteed.
func MustBool(v Val) bool { return v.(bool) }

// MustTable returns v as a table, panicking if it is not one. Use it only where the type is already guaranteed.
func MustTable(v Val) *Table { return v.(*Table) }

// MustFunction returns v as a function, panicking if it is not one. Use it only where the type is already guaranteed.
func MustFunction(v Val) Function { return v.(Function) }

// MustCoroutine returns v as a coroutine, panicking if it is not one. Use it only where the type is already guaranteed.
func MustCoroutine(v Val) *Coroutine { return v.(*Coroutine) }

// MustBuffer returns v as a buffer, panicking if it is not one. Use it only where the type is already guaranteed.
func MustBuffer(v Val) *Buffer { return v.(*Buffer) }

// MustVector returns v as a vector, panicking if it is not one. Use it only where the type is already guaranteed.
func MustVector(v Val) Vector { return v.(Vector) }
