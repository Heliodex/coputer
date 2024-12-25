package main

import (
	"encoding/binary"
	"math"
)

type Buffer []byte

func buffer_create(args Args) Ret {
	size := int(args.GetNumber())

	b := make(Buffer, size)
	return &b
}

func buffer_fromstring(args Args) Ret {
	str := args.GetString()

	b := Buffer(str)
	return &b
}

func buffer_tostring(args Args) Ret {
	b := *args.GetBuffer()

	return string(b)
}

func buffer_len(args Args) Ret {
	b := *args.GetBuffer()

	return float64(len(b))
}

func readValues(args *Args) (b Buffer, offset int) {
	b = *args.GetBuffer()
	offset = int(args.GetNumber())
	return
}

func buffer_readi8(args Args) Ret {
	b, offset := readValues(&args)

	return float64(int8(b[offset]))
}

func buffer_readu8(args Args) Ret {
	b, offset := readValues(&args)

	return float64(uint8(b[offset]))
}

func buffer_readi16(args Args) Ret {
	b, offset := readValues(&args)

	b2 := b[offset : offset+2]
	return float64(int16(binary.LittleEndian.Uint16(b2)))
}

func buffer_readu16(args Args) Ret {
	b, offset := readValues(&args)

	b2 := b[offset : offset+2]
	return float64(binary.LittleEndian.Uint16(b2))
}

func buffer_readi32(args Args) Ret {
	b, offset := readValues(&args)

	b4 := b[offset : offset+4] // we are inb4
	return float64(int32(binary.LittleEndian.Uint32(b4)))
}

func buffer_readu32(args Args) Ret {
	b, offset := readValues(&args)

	b4 := b[offset : offset+4]
	return float64(binary.LittleEndian.Uint32(b4))
}

func buffer_readf32(args Args) Ret {
	b, offset := readValues(&args)

	b4 := b[offset : offset+4]
	return float64(math.Float32frombits(binary.LittleEndian.Uint32(b4)))
}

func buffer_readf64(args Args) Ret {
	b, offset := readValues(&args)

	b8 := b[offset : offset+8]
	return float64(math.Float64frombits(binary.LittleEndian.Uint64(b8)))
}

type num interface {
	int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

func writeValues[V num](args *Args) (b Buffer, offset int, value V) {
	b, offset = readValues(args)
	value = V(args.GetNumber())
	return
}

func checkValues(l, offset, size int) {
	if offset + size > l || offset < 0 {
		panic("buffer access out of bounds")
	}
}

func buffer_writei8(args Args) {
	b, offset, value := writeValues[int8](&args)
	checkValues(len(b), offset, 1)

	b[offset] = byte(value)
}

func buffer_writeu8(args Args) {
	b, offset, value := writeValues[uint8](&args)
	checkValues(len(b), offset, 1)

	b[offset] = byte(value)
}

func buffer_writei16(args Args) {
	b, offset, value := writeValues[int16](&args)
	checkValues(len(b), offset, 2)

	b2 := b[offset : offset+2]
	binary.LittleEndian.PutUint16(b2, uint16(value))
}

func buffer_writeu16(args Args) {
	b, offset, value := writeValues[uint16](&args)
	checkValues(len(b), offset, 2)

	b2 := b[offset : offset+2]
	binary.LittleEndian.PutUint16(b2, value)
}

func buffer_writei32(args Args) {
	b, offset, value := writeValues[int32](&args)
	checkValues(len(b), offset, 4)

	b4 := b[offset : offset+4]
	binary.LittleEndian.PutUint32(b4, uint32(value))
}

func buffer_writeu32(args Args) {
	b, offset, value := writeValues[uint32](&args)
	checkValues(len(b), offset, 4)

	b4 := b[offset : offset+4]
	binary.LittleEndian.PutUint32(b4, value)
}

func buffer_writef32(args Args) {
	b, offset, value := writeValues[float32](&args)
	checkValues(len(b), offset, 4)

	b4 := b[offset : offset+4]
	binary.LittleEndian.PutUint32(b4, math.Float32bits(value))
}

func buffer_writef64(args Args) {
	b, offset, value := writeValues[float64](&args)
	checkValues(len(b), offset, 8)

	b8 := b[offset : offset+8]
	binary.LittleEndian.PutUint64(b8, math.Float64bits(value))
}

func buffer_readstring(args Args) Ret {
	b, offset := readValues(&args)
	count := int(args.GetNumber())
	checkValues(len(b), offset, count)

	bl := b[offset : offset+count]
	return string(bl)
}

func buffer_writestring(args Args) {
	b, offset := readValues(&args)
	value := args.GetString()
	count := int(args.GetNumber(float64(len(value))))
	checkValues(len(b), offset, count)

	copy(b[offset:offset+count], value)
}

func buffer_copy(args Args) {
	target, targetOffset := readValues(&args)
	source := *args.GetBuffer()
	sourceOffset := int(args.GetNumber(0))
	count := int(args.GetNumber(float64(len(source))))
	checkValues(len(source), sourceOffset, count)
	checkValues(len(target), targetOffset, count)

	copy(target[targetOffset:targetOffset+count], source[sourceOffset:sourceOffset+count])
}

func buffer_fill(args Args) {
	b, offset, value := writeValues[byte](&args)
	count := int(args.GetNumber(float64(len(b))))
	checkValues(len(b), offset, count)

	for i := range count {
		b[offset+i] = value
	}
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
