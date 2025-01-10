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

func buffer_writei8(args Args) Rets {
	b, offset, value := writeValues[int8](&args)
	if offset+1 > len(b) || offset < 0 {
		return Rets{"buffer access out of bounds", false}
	}

	b[offset] = byte(value)
	return Rets{nil, true}
}

func buffer_writeu8(args Args) Rets {
	b, offset, value := writeValues[uint8](&args)
	if offset+1 > len(b) || offset < 0 {
		return Rets{"buffer access out of bounds", false}
	}

	b[offset] = byte(value)
	return Rets{nil, true}
}

func buffer_writei16(args Args) Rets {
	b, offset, value := writeValues[int16](&args)
	if offset+2 > len(b) || offset < 0 {
		return Rets{"buffer access out of bounds", false}
	}

	b2 := b[offset : offset+2]
	binary.LittleEndian.PutUint16(b2, uint16(value))
	return Rets{nil, true}
}

func buffer_writeu16(args Args) Rets {
	b, offset, value := writeValues[uint16](&args)
	if offset+2 > len(b) || offset < 0 {
		return Rets{"buffer access out of bounds", false}
	}

	b2 := b[offset : offset+2]
	binary.LittleEndian.PutUint16(b2, value)
	return Rets{nil, true}
}

func buffer_writei32(args Args) Rets {
	b, offset, value := writeValues[int32](&args)
	if offset+4 > len(b) || offset < 0 {
		return Rets{"buffer access out of bounds", false}
	}

	b4 := b[offset : offset+4]
	binary.LittleEndian.PutUint32(b4, uint32(value))
	return Rets{nil, true}
}

func buffer_writeu32(args Args) Rets {
	b, offset, value := writeValues[uint32](&args)
	if offset+4 > len(b) || offset < 0 {
		return Rets{"buffer access out of bounds", false}
	}

	b4 := b[offset : offset+4]
	binary.LittleEndian.PutUint32(b4, value)
	return Rets{nil, true}
}

func buffer_writef32(args Args) Rets {
	b, offset, value := writeValues[float32](&args)
	if offset+4 > len(b) || offset < 0 {
		return Rets{"buffer access out of bounds", false}
	}

	b4 := b[offset : offset+4]
	binary.LittleEndian.PutUint32(b4, math.Float32bits(value))
	return Rets{nil, true}
}

func buffer_writef64(args Args) Rets {
	b, offset, value := writeValues[float64](&args)
	if offset+8 > len(b) || offset < 0 {
		return Rets{"buffer access out of bounds", false}
	}

	b8 := b[offset : offset+8]
	binary.LittleEndian.PutUint64(b8, math.Float64bits(value))
	return Rets{nil, true}
}

func buffer_readstring(args Args) Rets {
	b, offset := readValues(&args)
	count := int(args.GetNumber())
	if offset+count > len(b) || offset < 0 {
		return Rets{"buffer access out of bounds", false}
	}

	bl := b[offset : offset+count]
	return Rets{string(bl), true}
}

func buffer_writestring(args Args) Rets {
	b, offset := readValues(&args)
	value := args.GetString()
	count := int(args.GetNumber(float64(len(value))))
	if offset+count > len(b) || offset < 0 {
		return Rets{"buffer access out of bounds", false}
	}

	copy(b[offset:offset+count], value)
	return Rets{nil, true}
}

func buffer_copy(args Args) Rets {
	target, targetOffset := readValues(&args)
	source := *args.GetBuffer()
	sourceOffset := int(args.GetNumber(0))
	count := int(args.GetNumber(float64(len(source))))
	if sourceOffset+count > len(source) || sourceOffset < 0 ||
		targetOffset+count > len(target) || targetOffset < 0 {
		return Rets{"buffer access out of bounds", false}
	}

	copy(target[targetOffset:targetOffset+count], source[sourceOffset:sourceOffset+count])
	return Rets{nil, true}
}

func buffer_fill(args Args) Rets {
	b, offset, value := writeValues[byte](&args)
	count := int(args.GetNumber(float64(len(b))))
	if offset+count > len(b) || offset < 0 {
		return Rets{"buffer access out of bounds", false}
	}

	for i := range count {
		b[offset+i] = value
	}
	return Rets{nil, true}
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
	MakeFn("writei8", buffer_writei8),
	MakeFn("writeu8", buffer_writeu8),
	MakeFn("writei16", buffer_writei16),
	MakeFn("writeu16", buffer_writeu16),
	MakeFn("writei32", buffer_writei32),
	MakeFn("writeu32", buffer_writeu32),
	MakeFn("writef32", buffer_writef32),
	MakeFn("writef64", buffer_writef64),
	MakeFn("readstring", buffer_readstring),
	MakeFn("writestring", buffer_writestring),
	MakeFn("copy", buffer_copy),
	MakeFn("fill", buffer_fill),
})
