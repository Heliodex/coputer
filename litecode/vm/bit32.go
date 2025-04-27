package vm

import "math/bits"

const (
	nbits   = 32
	allones = ^uint32(0)
)

// trim extra bits
func trim(x uint32) uint32 {
	return x & allones
}

// builds a number with 'n' ones (1 <= n <= nbits)
func bitmask(n int) uint32 {
	return ^((allones - 1) << (n - 1))
}

func andaux(args Args) uint32 {
	x := allones
	for range args.List {
		x &= uint32(args.GetNumber())
	}
	return trim(x)
}

func b_shift(r uint32, i int) uint32 {
	if i < 0 { // shift right?
		i = -i
		// if i >= nbits {
		// 	return 0
		// }
		return trim(r) >> i
	}

	// shift left
	// if i >= nbits {
	// 	return 0
	// }
	return trim(r << i)
}

func bit32_arshift(args Args) (r []Val, err error) {
	x := uint32(args.GetNumber())
	i := int(args.GetNumber())

	if i < 0 || x&(1<<(nbits-1)) == 0 {
		return []Val{float64(b_shift(x, -i))}, nil
	} else if i >= nbits {
		// arithmetic shift for 'negative' number
		return []Val{float64(allones)}, nil
	}
	return []Val{float64(trim(x>>i | ^(allones >> i)))}, nil
}

func bit32_band(args Args) (r []Val, err error) {
	return []Val{float64(andaux(args))}, nil
}

func bit32_bnot(args Args) (r []Val, err error) {
	x := ^uint32(args.GetNumber())

	return []Val{float64(trim(x))}, nil
}

func bit32_bor(args Args) (r []Val, err error) {
	var x uint32
	for range args.List {
		x |= uint32(args.GetNumber())
	}
	return []Val{float64(trim(x))}, nil
}

func bit32_btest(args Args) (r []Val, err error) {
	return []Val{andaux(args) != 0}, nil
}

func bit32_bxor(args Args) (r []Val, err error) {
	var x uint32
	for range args.List {
		x ^= uint32(args.GetNumber())
	}
	return []Val{float64(trim(x))}, nil
}

func bit32_byteswap(args Args) (r []Val, err error) {
	n := uint32(args.GetNumber())

	return []Val{float64(bits.ReverseBytes32(n))}, nil
}

func bit32_countlz(args Args) (r []Val, err error) {
	v := uint32(args.GetNumber())

	return []Val{float64(bits.LeadingZeros32(v))}, nil
}

func bit32_countrz(args Args) (r []Val, err error) {
	v := uint32(args.GetNumber())

	return []Val{float64(bits.TrailingZeros32(v))}, nil
}

/*
** get field and width arguments for field-manipulation functions,
** checking whether they are valid.
 */
func fieldargs(args Args) (f, w int, msg string, ok bool) {
	f = int(args.GetNumber())
	w = int(args.GetNumber(1))

	if f < 0 {
		return 0, 0, "field cannot be negative", false
	} else if w < 1 {
		return 0, 0, "width must be positive", false
	} else if f+w > nbits {
		return 0, 0, "trying to access non-existent bits", false
	}
	return f, w, "", true
}

func bit32_extract(args Args) (r []Val, err error) {
	x := uint32(args.GetNumber())

	f, w, msg, ok := fieldargs(args)
	if !ok {
		return []Val{msg, false}, nil
	}
	return []Val{float64(x >> f & bitmask(w)), true}, nil
}

func bit32_replace(args Args) (r []Val, err error) {
	x := uint32(args.GetNumber())
	v := uint32(args.GetNumber())

	f, w, msg, ok := fieldargs(args)
	if !ok {
		return []Val{msg, false}, nil
	}
	m := bitmask(w)
	v &= m // erase bits outside given width
	return []Val{float64(x & ^(m<<f) | v<<f), true}, nil
}

func bit32_lrotate(args Args) (r []Val, err error) {
	x := uint32(args.GetNumber())
	i := int(args.GetNumber())

	return []Val{float64(bits.RotateLeft32(x, i))}, nil
}

func bit32_lshift(args Args) (r []Val, err error) {
	x := uint32(args.GetNumber())
	disp := int(args.GetNumber())

	return []Val{float64(b_shift(x, disp))}, nil
}

func bit32_rrotate(args Args) (r []Val, err error) {
	x := uint32(args.GetNumber())
	i := int(args.GetNumber())

	return []Val{float64(bits.RotateLeft32(x, -i))}, nil
}

func bit32_rshift(args Args) (r []Val, err error) {
	x := uint32(args.GetNumber())
	disp := int(args.GetNumber())

	return []Val{float64(b_shift(x, -disp))}, nil
}

var libbit32 = NewLib([]Function{
	MakeFn("arshift", bit32_arshift),
	MakeFn("band", bit32_band),
	MakeFn("bnot", bit32_bnot),
	MakeFn("bor", bit32_bor),
	MakeFn("btest", bit32_btest),
	MakeFn("bxor", bit32_bxor),
	MakeFn("byteswap", bit32_byteswap),
	MakeFn("countlz", bit32_countlz),
	MakeFn("countrz", bit32_countrz),
	MakeFn("extract", bit32_extract),
	MakeFn("lrotate", bit32_lrotate),
	MakeFn("lshift", bit32_lshift),
	MakeFn("replace", bit32_replace),
	MakeFn("rrotate", bit32_rrotate),
	MakeFn("rshift", bit32_rshift),
})
