package vm

import (
	"errors"
	"fmt"
	"iter"
	"math"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// p sure that 'globals' is a misnomer here but whatever

/* -- fantastic globals and whether to implement them --

loadstring: security and better api should be used
newproxy: not much use without metatables
rawequal: not much use without metatables
rawget: not much use without metatables
rawlen: not much use without metatables
rawset: not much use without metatables
select: this function's kinda stupid
typeof: not much use without metatables
*/

func ipairs_iter(args Args) (r Rets, err error) {
	a := args.GetTable()
	i := args.GetNumber() + 1

	if a.Array == nil || int(i) > len(a.Array) {
		return
	} else if v := a.Array[int(i)-1]; v != nil {
		return Rets{i, v}, nil
	}
	return // would prefer nil, nil but whateverrrrr
}

var ipairs = MakeFn("ipairs", ipairs_iter)

func global_ipairs(args Args) (r Rets, err error) {
	a := args.GetTable()

	return Rets{ipairs, a, float64(0)}, nil
}

// The call next(t, k), where k is a key of the table t, returns a next key in the table, in an arbitrary order. (It returns also the value associated with that key, as a second return value.) The call next(t, nil) returns a first pair. When there are no more pairs, next returns nil.
func global_next(args Args) (r Rets, err error) {
	t := args.GetTable()
	fk := args.GetAny(nil)

	if fk == nil {
		next, stop := iter.Pull2(t.Iter())
		defer stop()

		k, v, ok := next()
		if ok {
			return Rets{k, v}, nil
		}
	}

	next, stop := iter.Pull2(t.Iter())
	defer stop()

	for {
		k, _, ok := next()
		if !ok {
			break
		} else if k != fk {
			continue
		}

		k, v, ok := next()
		if !ok {
			break
		}
		return Rets{k, v}, nil
	}

	return Rets{nil}, nil // one nil?
}

func global_pairs(args Args) (r Rets, err error) {
	t := args.GetTable()

	return Rets{MakeFn("next", global_next), t}, nil
}

const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func global_tonumber(args Args) (r Rets, err error) {
	value := args.GetAny()
	radix := uint64(args.GetNumber(10))

	str, ok := value.(string)
	if !ok || radix < 2 || radix > 36 {
		return Rets{nil}, nil
		// panic("base out of range") // invalid argument #2
	}

	if radix == 10 {
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return Rets{f}, nil
		}
	}

	var negative bool
	if strings.HasPrefix(str, "-") {
		negative = true
	}

	switch radix {
	case 10, 16:
		if strings.HasPrefix(str, "0x") {
			str = str[2:]
			radix = 16
		}
		// case 2:
		// 	if strings.HasPrefix(str, "0b") {
		// 		str = str[2:]
		// 	}
	}

	str = strings.ToUpper(str)

	radixChars := chars[:radix]

	var n uint64
	for _, c := range str {
		n *= radix
		index := strings.IndexRune(radixChars, c)
		if index == -1 {
			return Rets{nil}, nil
		}
		n += uint64(index)
	}

	if negative {
		return Rets{float64(-n)}, nil
	}
	return Rets{float64(n)}, nil
}

const (
	kPow10TableMin = -292
	kPow10TableMax = 324
)

var kPow5Table = [16]uint64{
	0x8000000000000000, 0xa000000000000000, 0xc800000000000000, 0xfa00000000000000, 0x9c40000000000000, 0xc350000000000000,
	0xf424000000000000, 0x9896800000000000, 0xbebc200000000000, 0xee6b280000000000, 0x9502f90000000000, 0xba43b74000000000,
	0xe8d4a51000000000, 0x9184e72a00000000, 0xb5e620f480000000, 0xe35fa931a0000000,
}

var kPow10Table = [39][3]uint64{
	{0xff77b1fcbebcdc4f, 0x25e8e89c13bb0f7b, 0x333443443333443b},
	{0x8dd01fad907ffc3b, 0xae3da7d97f6792e4, 0xbbb3ab3cb3ba3cbc},
	{0x9d71ac8fada6c9b5, 0x6f773fc3603db4aa, 0x4ba4bc4bb4bb4bcc},
	{0xaecc49914078536d, 0x58fae9f773886e19, 0x3ba3bc33b43b43bb},
	{0xc21094364dfb5636, 0x985915fc12f542e5, 0x33b43b43a33b33cb},
	{0xd77485cb25823ac7, 0x7d633293366b828c, 0x34b44c444343443c},
	{0xef340a98172aace4, 0x86fb897116c87c35, 0x333343333343334b},
	{0x84c8d4dfd2c63f3b, 0x29ecd9f40041e074, 0xccaccbbcbcbb4bbc},
	{0x936b9fcebb25c995, 0xcab10dd900beec35, 0x3ab3ab3ab3bb3bbb},
	{0xa3ab66580d5fdaf5, 0xc13e60d0d2e0ebbb, 0x4cc3dc4db4db4dbb},
	{0xb5b5ada8aaff80b8, 0x0d819992132456bb, 0x33b33a34c33b34ab},
	{0xc9bcff6034c13052, 0xfc89b393dd02f0b6, 0x33c33b44b43c34bc},
	{0xdff9772470297ebd, 0x59787e2b93bc56f8, 0x43b444444443434c},
	{0xf8a95fcf88747d94, 0x75a44c6397ce912b, 0x443334343443343b},
	{0x8a08f0f8bf0f156b, 0x1b8e9ecb641b5900, 0xbbabab3aa3ab4ccc},
	{0x993fe2c6d07b7fab, 0xe546a8038efe402a, 0x4cb4bc4db4db4bcc},
	{0xaa242499697392d2, 0xdde50bd1d5d0b9ea, 0x3ba3ba3bb33b33bc},
	{0xbce5086492111aea, 0x88f4bb1ca6bcf585, 0x44b44c44c44c43cb},
	{0xd1b71758e219652b, 0xd3c36113404ea4a9, 0x44c44c44c444443b},
	{0xe8d4a51000000000, 0x0000000000000000, 0x444444444444444c},
	{0x813f3978f8940984, 0x4000000000000000, 0xcccccccccccccccc},
	{0x8f7e32ce7bea5c6f, 0xe4820023a2000000, 0xbba3bc4cc4cc4ccc},
	{0x9f4f2726179a2245, 0x01d762422c946591, 0x4aa3bb3aa3ba3bab},
	{0xb0de65388cc8ada8, 0x3b25a55f43294bcc, 0x3ca33b33b44b43bc},
	{0xc45d1df942711d9a, 0x3ba5d0bd324f8395, 0x44c44c34c44b44cb},
	{0xda01ee641a708de9, 0xe80e6f4820cc9496, 0x33b33b343333333c},
	{0xf209787bb47d6b84, 0xc0678c5dbd23a49b, 0x443444444443443b},
	{0x865b86925b9bc5c2, 0x0b8a2392ba45a9b3, 0xdbccbcccb4cb3bbb},
	{0x952ab45cfa97a0b2, 0xdd945a747bf26184, 0x3bc4bb4ab3ca3cbc},
	{0xa59bc234db398c25, 0x43fab9837e699096, 0x3bb3ac3ab3bb33ac},
	{0xb7dcbf5354e9bece, 0x0c11ed6d538aeb30, 0x33b43b43b34c34dc},
	{0xcc20ce9bd35c78a5, 0x31ec038df7b441f5, 0x34c44c43c44b44cb},
	{0xe2a0b5dc971f303a, 0x2e44ae64840fd61e, 0x333333333333333c},
	{0xfb9b7cd9a4a7443c, 0x169840ef017da3b2, 0x433344443333344c},
	{0x8bab8eefb6409c1a, 0x1ad089b6c2f7548f, 0xdcbdcc3cc4cc4bcb},
	{0x9b10a4e5e9913128, 0xca7cf2b4191c8327, 0x3ab3cb3bc3bb4bbb},
	{0xac2820d9623bf429, 0x546345fa9fbdcd45, 0x3bb3cc43c43c43cb},
	{0xbf21e44003acdd2c, 0xe0470a63e6bd56c4, 0x44b34a43b44c44bc},
	{0xd433179d9c8cb841, 0x5fa60692a46151ec, 0x43a33a33a333333c},
}

var kDigitTable = [200]byte{
	'0', '0', '0', '1', '0', '2', '0', '3', '0', '4', '0', '5', '0', '6', '0', '7', '0', '8', '0', '9',
	'1', '0', '1', '1', '1', '2', '1', '3', '1', '4', '1', '5', '1', '6', '1', '7', '1', '8', '1', '9',
	'2', '0', '2', '1', '2', '2', '2', '3', '2', '4', '2', '5', '2', '6', '2', '7', '2', '8', '2', '9',
	'3', '0', '3', '1', '3', '2', '3', '3', '3', '4', '3', '5', '3', '6', '3', '7', '3', '8', '3', '9',
	'4', '0', '4', '1', '4', '2', '4', '3', '4', '4', '4', '5', '4', '6', '4', '7', '4', '8', '4', '9',
	'5', '0', '5', '1', '5', '2', '5', '3', '5', '4', '5', '5', '5', '6', '5', '7', '5', '8', '5', '9',
	'6', '0', '6', '1', '6', '2', '6', '3', '6', '4', '6', '5', '6', '6', '6', '7', '6', '8', '6', '9',
	'7', '0', '7', '1', '7', '2', '7', '3', '7', '4', '7', '5', '7', '6', '7', '7', '7', '8', '7', '9',
	'8', '0', '8', '1', '8', '2', '8', '3', '8', '4', '8', '5', '8', '6', '8', '7', '8', '8', '8', '9',
	'9', '0', '9', '1', '9', '2', '9', '3', '9', '4', '9', '5', '9', '6', '9', '7', '9', '8', '9', '9',
}

// x*y => 128-bit product (lo+hi)
func mul128(x, y uint64) (uint64, uint64) {
	x0, x1 := uint64(uint32(x)), uint64(uint32(x>>32))
	y0, y1 := uint64(uint32(y)), uint64(uint32(y>>32))
	p11, p01, p10, p00 := x1*y1, x0*y1, x1*y0, x0*y0
	mid := p10 + p00>>32 + uint64(uint32(p01))
	r0 := mid<<32 | uint64(uint32(p00))
	r1 := p11 + mid>>32 + p01>>32
	return r1, r0
}

func b2i(b bool) uint64 {
	if b { // damn integer booleans
		return 1
	}
	return 0 // nock moment #1
}

// (x*y)>>64 => 128-bit product (lo+hi)
func mul192hi(xhi, xlo, y uint64) (uint64, uint64) {
	z2, z1 := mul128(xhi, y)
	z1c, _ := mul128(xlo, y)

	z1 += z1c
	z2 += b2i(z1 < z1c)

	return z2, z1
}

// 9.3. Rounding to odd (+ figure 8 + result 23)
func roundodd(ghi, glo, cp uint64) uint64 {
	xhi, _ := mul128(glo, cp)
	yhi, ylo := mul128(ghi, cp)

	z := ylo + xhi
	return yhi + b2i(z < xhi) | b2i(z > 1)
}

func schubfach(exponent int, fraction uint64) (uint64, int) {
	// Extract c & q such that c*2^q == |v|
	c := fraction
	q := exponent - 1023 - 51

	if exponent != 0 { // normal numbers have implicit leading 1
		c |= 1 << 52
		q--
	}

	// 8.3. Fast path for integers
	if uint(-q) < 53 && c&(1<<-q-1) == 0 {
		return c >> -q, 0
	}

	// 5. Rounding interval
	irr := c == 1<<52 && q != -1074 // Qmin
	out := uint64(int(c & 1))

	// 9.8.1. Boundaries for c
	cbl := 4*c - 2 + b2i(irr)
	cb := 4 * c
	cbr := 4*c + 2

	// 9.1. Computing k and h
	const Q = 20
	const C = 315652   // floor(2^Q * log10(2))
	const A = -131008  // floor(2^Q * log10(3/4))
	const C2 = 3483294 // floor(2^Q * log2(10))
	var k int
	if irr {
		k = (q*C + A) >> Q
	} else {
		k = q * C >> Q
	}
	h := q + -k*C2>>Q + 1 // see (9) in 9.9

	// 9.8.2. Overestimates of powers of 10
	// Recover 10^-k fraction using compact tables generated by tools/numutils.py
	// The 128-bit fraction is encoded as 128-bit baseline * power-of-5 * scale + offset
	gtoff := -k - kPow10TableMin
	gt := kPow10Table[gtoff>>4]

	g15 := gtoff & 15
	ghi, glo := mul192hi(gt[0], gt[1], kPow5Table[g15])

	// Apply 1-bit scale + 3-bit offset; note, offset is intentionally applied without carry, numutils.py validates that this is sufficient
	gterr := gt[2] >> (g15 * 4) & 15
	gtscale := gterr >> 3

	ghi <<= gtscale
	ghi += glo >> 63 & gtscale
	glo <<= gtscale
	glo -= gterr&7 - 4

	// 9.9. Boundaries for v
	vbl := roundodd(ghi, glo, cbl<<h)
	vb := roundodd(ghi, glo, cb<<h)
	vbr := roundodd(ghi, glo, cbr<<h)

	// Main algorithm; see figure 7 + figure 9
	s := vb / 4
	if s >= 10 {
		sp := s / 10

		upin := vbl+out <= 40*sp
		wpin := vbr >= 40*sp+40+out

		if upin != wpin {
			return sp + b2i(wpin), k + 1
		}
	}

	// Figure 7 contains the algorithm to select between u (s) and w (s+1)
	// rup computes the last 4 conditions in that algorithm
	// rup is only used when uin == win, but since these branches predict poorly we use branchless selects
	if uin, win := vbl+out <= 4*s, 4*s+4+out <= vbr; uin != win {
		return s + b2i(win), k
	}

	rup := vb >= 4*s+2+1-s&1
	return s + b2i(rup), k
}

type bufPos struct {
	buf [48]byte
	pos int
}

func (b *bufPos) String() string {
	return string(b.buf[:b.pos])
}

func (b *bufPos) Get(i int) byte {
	return b.buf[b.pos+i]
}

func (b *bufPos) Set(i int, v byte) {
	b.buf[b.pos+i] = v
}

func printunsignedrev(end *bufPos, num uint64) *bufPos {
	for num >= 10000 {
		tail := num % 10000

		dt1, dt2 := tail/100*2, tail%100*2
		end.Set(-4, kDigitTable[dt1])
		end.Set(-3, kDigitTable[dt1+1])
		end.Set(-2, kDigitTable[dt2])
		end.Set(-1, kDigitTable[dt2+1])

		num /= 10000
		end.pos -= 4
	}

	for num >= 10 {
		dt := num % 100 * 2
		end.Set(-2, kDigitTable[dt])
		end.Set(-1, kDigitTable[dt+1])

		num /= 100
		end.pos -= 2
	}

	if num != 0 {
		end.pos--
		end.Set(0, '0'+byte(num))
	}

	return end
}

func printexp(buf *bufPos, v int) string {
	p, b := buf.pos, buf.buf

	b[p] = 'e'
	if v < 0 {
		b[p+1] = '-'
		v = -v
	} else {
		b[p+1] = '+'
	}

	if v >= 100 {
		b[p+2] = '0' + byte(v/100)
		p++
		v %= 100
	}

	dt := v * 2
	b[p+2], b[p+3] = kDigitTable[dt], kDigitTable[dt+1]

	return string(b[:p+4])
}

func trimzero(end *bufPos) *bufPos {
	for end.Get(-1) == '0' {
		end.pos--
	}

	return end
}

func num2str2(exponent int, fraction uint64, buf *bufPos) string {
	// convert binary to decimal using Schubfach
	s, k := schubfach(exponent, fraction)

	// print the decimal to a temporary buffer; we'll need to insert the decimal point and figure out the format
	const decend = 20 // significand needs at most 17 digits; the rest of the buffer may be copied using fixed length memcpy
	db2 := &bufPos{[48]byte{}, decend}
	dec := printunsignedrev(db2, s)

	declen := decend - dec.pos
	dot := declen + k

	// the limits are somewhat not arbitrary
	if dot < -5 || dot > 21 {
		// scientific format
		buf.Set(0, dec.Get(0))
		buf.Set(1, '.')
		for i := range declen - 1 {
			buf.Set(i+2, dec.Get(i+1))
		}

		buf.pos += declen + 1
		exp := trimzero(buf)

		if exp.Get(-1) == '.' {
			exp.pos--
		}

		return printexp(exp, dot-1)
	}

	// fixed point format
	if dot <= 0 {
		buf.Set(0, '0')
		buf.Set(1, '.')

		for i := range -dot {
			buf.Set(i+2, '0')
		}
		for i := range declen {
			buf.Set(i+2-dot, dec.Get(i))
		}

		buf.pos += 2 + -dot + declen
		return trimzero(buf).String()
	} else if dot == declen {
		// no dot
		for i := range dot {
			buf.Set(i, dec.Get(i))
		}

		// fmt.Println("no dot", dec, buf, dot, declen)

		buf.pos += dot
		return buf.String()
	} else if dot < declen {
		// dot in the middle
		for i := range dot {
			buf.Set(i, dec.Get(i))
		}

		buf.Set(dot, '.')

		dec.pos += dot
		for i := range declen - dot {
			buf.Set(i+dot+1, dec.Get(i))
		}

		buf.pos += declen + 1
		return trimzero(buf).String()
	}

	// fmt.Println("no dot, zero padding", declen, dot)
	for i := range declen {
		buf.Set(i, dec.Get(i))
	}
	for i := range dot - declen {
		buf.Set(i+declen, '0')
	}

	return string(buf.buf[:buf.pos+dot])
}

func num2str(n float64) string {
	// IEEE-754
	bits := math.Float64bits(n)
	sign := uint8(bits>>63) == 0 // nock moment #2
	exponent := int(bits >> 52 & 0x7ff)
	fraction := bits & (1<<52 - 1)

	if exponent == 0x7ff {
		// specials
		if fraction != 0 {
			return "nan"
		} else if sign {
			return "inf"
		}
		return "-inf"
	} else if exponent == 0 && fraction == 0 {
		// zero
		if sign {
			return "0"
		}
		return "-0"
	}

	s := num2str2(exponent, fraction, &bufPos{})

	// sign bit
	if sign {
		return s
	}
	return "-" + s
}

// ToString returns a string representation of any value.
func ToString(a any) string {
	switch v := a.(type) {
	case nil:
		return "nil"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return num2str(v)
	case Vector:
		// just 3-wide 4-now
		return fmt.Sprintf("%s, %s, %s", num2str(float64(v[0])), num2str(float64(v[1])), num2str(float64(v[2])))
	case string:
		return strings.ReplaceAll(v, "\n", "\r\n") // bruh
	}
	return fmt.Sprint(a)
}

func global_tostring(args Args) (r Rets, err error) {
	value := args.GetAny()

	return Rets{ToString(value)}, nil
}

func global_type(args Args) (r Rets, err error) {
	obj := args.GetAny()

	t, ok := luautype[typeOf(obj)]
	if !ok {
		return Rets{"userdata"}, nil
	}
	return Rets{t}, nil
}

func isAbsolutePath(p string) bool {
	return len(p) >= 3 && isalpha(p[0]) && p[1] == ':' && (p[2] == '/' || p[2] == '\\') ||
		len(p) >= 1 && (p[0] == '/' || p[0] == '\\')
}

func hasValidPrefix(path string) bool {
	return path[:2] == "./" || path[:3] == "../"
}

func global_require(args Args) (r Rets, err error) {
	name := args.GetString()
	if isAbsolutePath(name) {
		return nil, invalidArg(1, "require", "cannot require an absolute path")
	}

	name = strings.ReplaceAll(name, "\\", "/")
	if !hasValidPrefix(name) {
		return nil, errors.New("require path must start with a valid prefix: ./ or ../")
	}

	// combine filepath and name to get the new path
	fp := args.Co.filepath
	path := filepath.Join(filepath.Dir(fp), name)
	path = strings.ReplaceAll(path, "\\", "/")
	// fmt.Println("REQUIRING", path)

	if path == fp {
		return nil, errors.New("cyclic module dependency: file requires itself")
	}

	c := args.Co.compiler

	// compile bytecodeee
	p, err := c.Compile(path)
	if err != nil {
		return nil, fmt.Errorf("error requiring module: %w", err)
	}

	rh := args.Co.requireHistory
	if slices.Contains(rh, path) {
		return nil, errors.New("cyclic module dependency: file requires itself indirectly")
	}
	p.requireHistory = append(rh, fp)

	// this is where we take it to the top babbyyyyy (with the same as parent global env)
	return Rets{p}, nil
}
