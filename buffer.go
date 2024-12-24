package main

type Buffer []byte

func buffer_create(args Args) Ret {
	size := args.GetNumber()

	b := make(Buffer, int(size))
	return &b
}

func buffer_fromstring(args Args) Ret {
	str := args.GetString()

	b := Buffer(str)
	return &b
}

func buffer_tostring(args Args) Ret {
	b := args.GetBuffer()

	return string(*b)
}

func buffer_len(args Args) Ret {
	b := args.GetBuffer()

	return float64(len(*b))
}

func buffer_readi8(args Args) Ret {
	panic("not implemented")
}

func buffer_readu8(args Args) Ret {
	panic("not implemented")
}

func buffer_readi16(args Args) Ret {
	panic("not implemented")
}

func buffer_readu16(args Args) Ret {
	panic("not implemented")
}

func buffer_readi32(args Args) Ret {
	panic("not implemented")
}

func buffer_readu32(args Args) Ret {
	panic("not implemented")
}

func buffer_readf32(args Args) Ret {
	panic("not implemented")
}

func buffer_readf64(args Args) Ret {
	panic("not implemented")
}

func buffer_writei8(args Args) {
	panic("not implemented")
}

func buffer_writeu8(args Args) {
	panic("not implemented")
}

func buffer_writei16(args Args) {
	panic("not implemented")
}

func buffer_writeu16(args Args) {
	panic("not implemented")
}

func buffer_writei32(args Args) {
	panic("not implemented")
}

func buffer_writeu32(args Args) {
	panic("not implemented")
}

func buffer_writef32(args Args) {
	panic("not implemented")
}

func buffer_writef64(args Args) {
	panic("not implemented")
}

func buffer_readstring(args Args) Ret {
	panic("not implemented")
}

func buffer_writestring(args Args) {
	panic("not implemented")
}

func buffer_copy(args Args) {
	panic("not implemented")
}

func buffer_fill(args Args) {
	panic("not implemented")
}

var libbuffer = NewTable([][2]any{
	MakeFn1("create", buffer_create),
	MakeFn1("fromstring", buffer_fromstring),
	MakeFn1("tostring", buffer_tostring),
	MakeFn1("len", buffer_len),
	MakeFn1("readi8", buffer_readi8),
	MakeFn1("readu8", buffer_readu8),
	MakeFn1("readi16", buffer_readi16),
	MakeFn1("readu16", buffer_readu16),
	MakeFn1("readi32", buffer_readi32),
	MakeFn1("readu32", buffer_readu32),
	MakeFn1("readf32", buffer_readf32),
	MakeFn1("readf64", buffer_readf64),
	MakeFn0("writei8", buffer_writei8),
	MakeFn0("writeu8", buffer_writeu8),
	MakeFn0("writei16", buffer_writei16),
	MakeFn0("writeu16", buffer_writeu16),
	MakeFn0("writei32", buffer_writei32),
	MakeFn0("writeu32", buffer_writeu32),
	MakeFn0("writef32", buffer_writef32),
	MakeFn0("writef64", buffer_writef64),
	MakeFn1("readstring", buffer_readstring),
	MakeFn0("writestring", buffer_writestring),
	MakeFn0("copy", buffer_copy),
	MakeFn0("fill", buffer_fill),
})
