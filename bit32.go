package main

import (
	"fmt"
)

var (
	NBITS   = 32
	ALLONES = ^uint32(0)
)

// trim extra bits
func trim(x uint32) uint32 {
	return x & ALLONES
}

// builds a number with 'n' ones (1 <= n <= NBITS)
func mask(n int) uint32 {
	return ^((ALLONES << 1) << (n - 1))
}

func andaux(args Args) uint32 {
	r := ALLONES
	for range args.args {
		r &= uint32(args.GetNumber())
	}
	return trim(r)
}

func b_shift(r uint32, i int) uint32 {
	if i < 0 { // shift right?
		i = -i
		if i >= NBITS {
			return 0
		}
		return trim(r) >> i
	}

	// shift left
	if i >= NBITS {
		return 0
	}
	return trim(r << i)
}

func bit32_arshift(args Args) Ret {
	r := uint32(args.GetNumber())
	i := int(args.GetNumber())

	if i < 0 || (r&(1<<(NBITS-1)) == 0) {
		return float64(b_shift(r, -i))
	}

	fmt.Println(r, r&(1<<(NBITS-1)))
	// arithmetic shift for 'negative' number
	if i >= NBITS {
		return float64(ALLONES)
	}
	return float64(trim((r >> i) | ^(ALLONES >> i)))
}

func bit32_band(args Args) Ret {
	return float64(andaux(args))
}

func bit32_bnot(args Args) Ret {
	r := ^uint32(args.GetNumber())

	return float64(trim(r))
}

func bit32_bor(args Args) Ret {
	r := uint32(0)
	for range args.args {
		r |= uint32(args.GetNumber())
	}
	return float64(trim(r))
}

func bit32_btest(args Args) Ret {
	return andaux(args) != 0
}

func bit32_bxor(args Args) Ret {
	r := uint32(0)
	for range args.args {
		r ^= uint32(args.GetNumber())
	}
	return float64(trim(r))
}

func bit32_byteswap(args Args) Ret {
	n := uint32(args.GetNumber())

	return float64((n << 24) | ((n << 8) & 0xff0000) | ((n >> 8) & 0xff00) | (n >> 24))
}

func bit32_countlz(args Args) Ret {
	v := uint32(args.GetNumber())

	for i := range NBITS {
		if v&(1<<(NBITS-1-i)) != 0 {
			return float64(i)
		}
	}
	return float64(NBITS)
}

func bit32_countrz(args Args) Ret {
	v := uint32(args.GetNumber())

	for i := range NBITS {
		if v&(1<<i) != 0 {
			return float64(i)
		}
	}
	return float64(NBITS)
}

/*
** get field and width arguments for field-manipulation functions,
** checking whether they are valid.
 */
func fieldargs(args Args) (int, int) {
	f := int(args.GetNumber())
	w := int(args.GetNumber(1))

	if f < 0 {
		panic("field cannot be negative")
	}
	if w < 1 {
		panic("width must be positive")
	}
	if f+w > NBITS {
		panic("trying to access non-existent bits")
	}
	return f, w
}

func bit32_extract(args Args) Ret {
	r := uint32(args.GetNumber())

	f, w := fieldargs(args)
	return float64((r >> f) & mask(w))
}

func bit32_replace(args Args) Ret {
	r := uint32(args.GetNumber())
	v := uint32(args.GetNumber())

	f, w := fieldargs(args)
	m := mask(w)
	v &= m // erase bits outside given width
	return float64((r & ^(m << f)) | (v << f))
}

func b_rot(r uint32, i int) uint32 {
	if i &= (NBITS - 1); i != 0 { // i %= NBITS
		// avoid undefined shift of NBITS when i == 0
		r = trim(r)
		r = (r << i) | (r >> (NBITS - i))
	}
	return trim(r)
}

func bit32_lrotate(args Args) Ret {
	r := uint32(args.GetNumber())
	i := int(args.GetNumber())

	return float64(b_rot(r, i))
}

func bit32_lshift(args Args) Ret {
	x := uint32(args.GetNumber())
	disp := int(args.GetNumber())

	return float64(b_shift(x, disp))
}

func bit32_rrotate(args Args) Ret {
	r := uint32(args.GetNumber())
	i := int(args.GetNumber())

	return float64(b_rot(r, -i))
}

func bit32_rshift(args Args) Ret {
	x := uint32(args.GetNumber())
	disp := int(args.GetNumber())

	return float64(b_shift(x, -disp))
}

var libbit32 = NewTable([][2]any{
	MakeFn1("arshift", bit32_arshift),
	MakeFn1("band", bit32_band),
	MakeFn1("bnot", bit32_bnot),
	MakeFn1("bor", bit32_bor),
	MakeFn1("btest", bit32_btest),
	MakeFn1("bxor", bit32_bxor),
	MakeFn1("byteswap", bit32_byteswap),
	MakeFn1("countlz", bit32_countlz),
	MakeFn1("countrz", bit32_countrz),
	MakeFn1("extract", bit32_extract),
	MakeFn1("replace", bit32_replace),
	MakeFn1("lrotate", bit32_lrotate),
	MakeFn1("lshift", bit32_lshift),
	MakeFn1("replace", bit32_replace),
	MakeFn1("rrotate", bit32_rrotate),
	MakeFn1("rshift", bit32_rshift),
})
