package litecode

import (
	"encoding/binary"
	"math"
)

type Buffer []byte

func buffer_create(args Args) (r Rets, err error) {
	size := int(args.GetNumber())

	b := make(Buffer, size)
	return Rets{&b}, nil
}

func buffer_fromstring(args Args) (r Rets, err error) {
	str := args.GetString()

	b := Buffer(str)
	return Rets{&b}, nil
}

func buffer_tostring(args Args) (r Rets, err error) {
	b := *args.GetBuffer()

	return Rets{string(b)}, nil
}

func buffer_len(args Args) (r Rets, err error) {
	b := *args.GetBuffer()

	return Rets{float64(len(b))}, nil
}

func readValues(args *Args) (b Buffer, offset int) {
	b = *args.GetBuffer()
	offset = int(args.GetNumber())
	return
}

func buffer_readi8(args Args) (r Rets, err error) {
	b, offset := readValues(&args)

	return Rets{float64(int8(b[offset]))}, nil
}

func buffer_readu8(args Args) (r Rets, err error) {
	b, offset := readValues(&args)

	return Rets{float64(uint8(b[offset]))}, nil
}

func buffer_readi16(args Args) (r Rets, err error) {
	b, offset := readValues(&args)

	b2 := b[offset : offset+2]
	return Rets{float64(int16(binary.LittleEndian.Uint16(b2)))}, nil
}

func buffer_readu16(args Args) (r Rets, err error) {
	b, offset := readValues(&args)

	b2 := b[offset : offset+2]
	return Rets{float64(binary.LittleEndian.Uint16(b2))}, nil
}

func buffer_readi32(args Args) (r Rets, err error) {
	b, offset := readValues(&args)

	b4 := b[offset : offset+4] // we are inb4
	return Rets{float64(int32(binary.LittleEndian.Uint32(b4)))}, nil
}

func buffer_readu32(args Args) (r Rets, err error) {
	b, offset := readValues(&args)

	b4 := b[offset : offset+4]
	return Rets{float64(binary.LittleEndian.Uint32(b4))}, nil
}

func buffer_readf32(args Args) (r Rets, err error) {
	b, offset := readValues(&args)

	b4 := b[offset : offset+4]
	return Rets{float64(math.Float32frombits(binary.LittleEndian.Uint32(b4)))}, nil
}

func buffer_readf64(args Args) (r Rets, err error) {
	b, offset := readValues(&args)

	b8 := b[offset : offset+8]
	return Rets{float64(math.Float64frombits(binary.LittleEndian.Uint64(b8)))}, nil
}

type num interface {
	int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

var OOB = Rets{"buffer access out of bounds", false}

func writeValues[V num](args *Args) (b Buffer, offset int, value V) {
	b, offset = readValues(args)
	value = V(args.GetNumber())
	return
}

func buffer_writei8(args Args) (r Rets, err error) {
	b, offset, value := writeValues[int8](&args)
	if offset+1 > len(b) || offset < 0 {
		return OOB, nil
	}

	b[offset] = byte(value)
	return Rets{nil, true}, nil
}

func buffer_writeu8(args Args) (r Rets, err error) {
	b, offset, value := writeValues[uint8](&args)
	if offset+1 > len(b) || offset < 0 {
		return OOB, nil
	}

	b[offset] = byte(value)
	return Rets{nil, true}, nil
}

func buffer_writei16(args Args) (r Rets, err error) {
	b, offset, value := writeValues[int16](&args)
	if offset+2 > len(b) || offset < 0 {
		return OOB, nil
	}

	b2 := b[offset : offset+2]
	binary.LittleEndian.PutUint16(b2, uint16(value))
	return Rets{nil, true}, nil
}

func buffer_writeu16(args Args) (r Rets, err error) {
	b, offset, value := writeValues[uint16](&args)
	if offset+2 > len(b) || offset < 0 {
		return OOB, nil
	}

	b2 := b[offset : offset+2]
	binary.LittleEndian.PutUint16(b2, value)
	return Rets{nil, true}, nil
}

func buffer_writei32(args Args) (r Rets, err error) {
	b, offset, value := writeValues[int32](&args)
	if offset+4 > len(b) || offset < 0 {
		return OOB, nil
	}

	b4 := b[offset : offset+4]
	binary.LittleEndian.PutUint32(b4, uint32(value))
	return Rets{nil, true}, nil
}

func buffer_writeu32(args Args) (r Rets, err error) {
	b, offset, value := writeValues[uint32](&args)
	if offset+4 > len(b) || offset < 0 {
		return OOB, nil
	}

	b4 := b[offset : offset+4]
	binary.LittleEndian.PutUint32(b4, value)
	return Rets{nil, true}, nil
}

func buffer_writef32(args Args) (r Rets, err error) {
	b, offset, value := writeValues[float32](&args)
	if offset+4 > len(b) || offset < 0 {
		return OOB, nil
	}

	b4 := b[offset : offset+4]
	binary.LittleEndian.PutUint32(b4, math.Float32bits(value))
	return Rets{nil, true}, nil
}

func buffer_writef64(args Args) (r Rets, err error) {
	b, offset, value := writeValues[float64](&args)
	if offset+8 > len(b) || offset < 0 {
		return OOB, nil
	}

	b8 := b[offset : offset+8]
	binary.LittleEndian.PutUint64(b8, math.Float64bits(value))
	return Rets{nil, true}, nil
}

func buffer_readbits(args Args) (r Rets, err error) {
	b, bitoffset := readValues(&args)
	bitcount := int(args.GetNumber())

	if bitoffset < 0 {
		return OOB, nil
	} else if uint32(bitcount) > 32 {
		return Rets{"bit count is out of range of [0; 32]", false}, nil
	} else if uint64(bitoffset+bitcount) > uint64(len(b)*8) {
		return OOB, nil
	}

	startbyte := uint32(bitoffset / 8)
	endbyte := uint32((bitoffset + bitcount + 7) / 8)

	var data uint64
	dataa := b[startbyte:endbyte]
	for i, v := range dataa {
		data |= uint64(v) << (i * 8)
	}

	subbyteoffset := uint64(bitoffset & 7)
	mask := uint64((1 << bitcount) - 1)

	return Rets{float64((data >> subbyteoffset) & mask), true}, nil
}

func buffer_writebits(args Args) (r Rets, err error) {
	b, bitoffset := readValues(&args)
	bitcount := int(args.GetNumber())
	value := uint64(args.GetNumber())

	if bitoffset < 0 {
		return OOB, nil
	} else if uint32(bitcount) > 32 {
		return Rets{"bit count is out of range of [0; 32]", false}, nil
	} else if uint64(bitoffset+bitcount) > uint64(len(b)*8) {
		return OOB, nil
	}

	startbyte := uint32(bitoffset / 8)
	endbyte := uint32((bitoffset + bitcount + 7) / 8)

	var data uint64
	dataa := b[startbyte:endbyte]
	for i, v := range dataa {
		data |= uint64(v) << (i * 8)
	}

	subbyteoffset := uint64(bitoffset & 7)
	mask := uint64((1<<bitcount)-1) << subbyteoffset

	data = (data & ^mask) | (value << subbyteoffset)

	bs := b[startbyte:endbyte]
	var dataa2 [8]byte
	binary.LittleEndian.PutUint64(dataa2[:], data)
	copy(bs, dataa2[:])

	return Rets{nil, true}, nil
}

func buffer_readstring(args Args) (r Rets, err error) {
	b, offset := readValues(&args)
	count := int(args.GetNumber())
	if offset+count > len(b) || offset < 0 {
		return OOB, nil
	}

	bl := b[offset : offset+count]
	return Rets{string(bl), true}, nil
}

func buffer_writestring(args Args) (r Rets, err error) {
	b, offset := readValues(&args)
	value := args.GetString()
	count := int(args.GetNumber(float64(len(value))))
	if offset+count > len(b) || offset < 0 {
		return OOB, nil
	}

	copy(b[offset:offset+count], value)
	return Rets{nil, true}, nil
}

func buffer_copy(args Args) (r Rets, err error) {
	target, targetOffset := readValues(&args)
	source := *args.GetBuffer()
	sourceOffset := int(args.GetNumber(0))
	count := int(args.GetNumber(float64(len(source))))
	if sourceOffset+count > len(source) || sourceOffset < 0 ||
		targetOffset+count > len(target) || targetOffset < 0 {
		return OOB, nil
	}

	copy(target[targetOffset:targetOffset+count], source[sourceOffset:sourceOffset+count])
	return Rets{nil, true}, nil
}

func buffer_fill(args Args) (r Rets, err error) {
	b, offset, value := writeValues[byte](&args)
	count := int(args.GetNumber(float64(len(b))))
	if offset+count > len(b) || offset < 0 {
		return OOB, nil
	}

	for i := range count {
		b[offset+i] = value
	}
	return Rets{nil, true}, nil
}

var libbuffer = NewTable([][2]any{
	MakeFn("create", buffer_create),
	MakeFn("fromstring", buffer_fromstring),
	MakeFn("tostring", buffer_tostring),
	MakeFn("len", buffer_len),
	MakeFn("readi8", buffer_readi8),
	MakeFn("readu8", buffer_readu8),
	MakeFn("readi16", buffer_readi16),
	MakeFn("readu16", buffer_readu16),
	MakeFn("readi32", buffer_readi32),
	MakeFn("readu32", buffer_readu32),
	MakeFn("readf32", buffer_readf32),
	MakeFn("readf64", buffer_readf64),
	MakeFn("writei8", buffer_writei8),
	MakeFn("writeu8", buffer_writeu8),
	MakeFn("writei16", buffer_writei16),
	MakeFn("writeu16", buffer_writeu16),
	MakeFn("writei32", buffer_writei32),
	MakeFn("writeu32", buffer_writeu32),
	MakeFn("writef32", buffer_writef32),
	MakeFn("writef64", buffer_writef64),
	MakeFn("readbits", buffer_readbits),
	MakeFn("writebits", buffer_writebits),
	MakeFn("readstring", buffer_readstring),
	MakeFn("writestring", buffer_writestring),
	MakeFn("copy", buffer_copy),
	MakeFn("fill", buffer_fill),
})
