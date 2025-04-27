local bit_or = bit32.bor
local bit_and = bit32.band
local bit_xor = bit32.bxor
local bit_lshift = bit32.lshift
local bit_rshift = bit32.rshift

-- convert little-endian 32-bit int to a 4-char string
local function lei2str(i)
	local f = function(s) return string.char(bit_and(bit_rshift(i, s), 255)) end
	local l2 = f(0) .. f(8) .. f(16) .. f(24)
	return l2
end

-- convert raw string to big-endian int
local function str2bei(s): number
	local v: number = 0
	for i = 1, #s do
		local b = string.byte(s, i)
		v *= 256
		v += b
	end
	return v
end

-- convert raw string to little-endian int
local function str2lei(s)
	local v = 0
	for i = #s, 1, -1 do
		v *= 256
		v += string.byte(s, i)
	end
	return v
end

-- cut up a string in little-endian ints of given size
local function cut_le_str(s, ...): { number }
	local o, r = 1, {}
	local args = { ... }
	for i = 1, #args do
		table.insert(r, str2lei(string.sub(s, o, o + args[i] - 1)))
		o += args[i]
	end
	return r
end

type MD5 = {
	a: number,
	b: number,
	c: number,
	d: number,
	pos: number,
	buf: string,
	new: () -> MD5,
	tohex: (s: string) -> string,
	sum: (s: string) -> string,
	sumhexa: (s: string) -> string,
	update: (self: MD5, s: string) -> MD5,
	finish: (self: MD5) -> string,
}

local md5 = {} :: MD5

-- An MD5 mplementation in Lua, requires bitlib (hacked to use LuaBit from above, ugh)
-- 10/02/2001 jcw@equi4.com

local CONSTS = {
	0xd76aa478,
	0xe8c7b756,
	0x242070db,
	0xc1bdceee,
	0xf57c0faf,
	0x4787c62a,
	0xa8304613,
	0xfd469501,
	0x698098d8,
	0x8b44f7af,
	0xffff5bb1,
	0x895cd7be,
	0x6b901122,
	0xfd987193,
	0xa679438e,
	0x49b40821,
	0xf61e2562,
	0xc040b340,
	0x265e5a51,
	0xe9b6c7aa,
	0xd62f105d,
	0x02441453,
	0xd8a1e681,
	0xe7d3fbc8,
	0x21e1cde6,
	0xc33707d6,
	0xf4d50d87,
	0x455a14ed,
	0xa9e3e905,
	0xfcefa3f8,
	0x676f02d9,
	0x8d2a4c8a,
	0xfffa3942,
	0x8771f681,
	0x6d9d6122,
	0xfde5380c,
	0xa4beea44,
	0x4bdecfa9,
	0xf6bb4b60,
	0xbebfbc70,
	0x289b7ec6,
	0xeaa127fa,
	0xd4ef3085,
	0x04881d05,
	0xd9d4d039,
	0xe6db99e5,
	0x1fa27cf8,
	0xc4ac5665,
	0xf4292244,
	0x432aff97,
	0xab9423a7,
	0xfc93a039,
	0x655b59c3,
	0x8f0ccc92,
	0xffeff47d,
	0x85845dd1,
	0x6fa87e4f,
	0xfe2ce6e0,
	0xa3014314,
	0x4e0811a1,
	0xf7537e82,
	0xbd3af235,
	0x2ad7d2bb,
	0xeb86d391,
	0x67452301,
	0xefcdab89,
	0x98badcfe,
	0x10325476,
}

local function f(x: number, y: number, z: number): number
	return bit_or(bit_and(x, y), bit_and(-x - 1, z))
end
local function g(x: number, y: number, z: number): number
	return bit_or(bit_and(x, z), bit_and(y, -z - 1))
end
local function h(x: number, y: number, z: number): number
	return bit_xor(x, bit_xor(y, z))
end
local function i(x: number, y: number, z: number): number
	return bit_xor(y, bit_or(x, -z - 1))
end
local function z(
	ff: (number, number, number) -> number,
	a: number,
	b,
	c,
	d,
	x: number,
	s,
	ac: number
): number
	local a2 = bit_and(a + ff(b, c, d) + x + ac, 0xFFFFFFFF)
	-- be *very* careful that left shift does not cause rounding!
	return bit_or(
		bit_lshift(bit_and(a2, bit_rshift(0xFFFFFFFF, s)), s),
		bit_rshift(a2, 32 - s)
	) + b
end

local function transform(
	A: number,
	B: number,
	C: number,
	D: number,
	X: { number }
)
	local a, b, c, d = A, B, C, D
	local t = CONSTS

	a = z(f, a, b, c, d, X[0], 7, t[1])
	d = z(f, d, a, b, c, X[1], 12, t[2])
	c = z(f, c, d, a, b, X[2], 17, t[3])
	b = z(f, b, c, d, a, X[3], 22, t[4])
	a = z(f, a, b, c, d, X[4], 7, t[5])
	d = z(f, d, a, b, c, X[5], 12, t[6])
	c = z(f, c, d, a, b, X[6], 17, t[7])
	b = z(f, b, c, d, a, X[7], 22, t[8])
	a = z(f, a, b, c, d, X[8], 7, t[9])
	d = z(f, d, a, b, c, X[9], 12, t[10])
	c = z(f, c, d, a, b, X[10], 17, t[11])
	b = z(f, b, c, d, a, X[11], 22, t[12])
	a = z(f, a, b, c, d, X[12], 7, t[13])
	d = z(f, d, a, b, c, X[13], 12, t[14])
	c = z(f, c, d, a, b, X[14], 17, t[15])
	b = z(f, b, c, d, a, X[15], 22, t[16])

	a = z(g, a, b, c, d, X[1], 5, t[17])
	d = z(g, d, a, b, c, X[6], 9, t[18])
	c = z(g, c, d, a, b, X[11], 14, t[19])
	b = z(g, b, c, d, a, X[0], 20, t[20])
	a = z(g, a, b, c, d, X[5], 5, t[21])
	d = z(g, d, a, b, c, X[10], 9, t[22])
	c = z(g, c, d, a, b, X[15], 14, t[23])
	b = z(g, b, c, d, a, X[4], 20, t[24])
	a = z(g, a, b, c, d, X[9], 5, t[25])
	d = z(g, d, a, b, c, X[14], 9, t[26])
	c = z(g, c, d, a, b, X[3], 14, t[27])
	b = z(g, b, c, d, a, X[8], 20, t[28])
	a = z(g, a, b, c, d, X[13], 5, t[29])
	d = z(g, d, a, b, c, X[2], 9, t[30])
	c = z(g, c, d, a, b, X[7], 14, t[31])
	b = z(g, b, c, d, a, X[12], 20, t[32])

	a = z(h, a, b, c, d, X[5], 4, t[33])
	d = z(h, d, a, b, c, X[8], 11, t[34])
	c = z(h, c, d, a, b, X[11], 16, t[35])
	b = z(h, b, c, d, a, X[14], 23, t[36])
	a = z(h, a, b, c, d, X[1], 4, t[37])
	d = z(h, d, a, b, c, X[4], 11, t[38])
	c = z(h, c, d, a, b, X[7], 16, t[39])
	b = z(h, b, c, d, a, X[10], 23, t[40])
	a = z(h, a, b, c, d, X[13], 4, t[41])
	d = z(h, d, a, b, c, X[0], 11, t[42])
	c = z(h, c, d, a, b, X[3], 16, t[43])
	b = z(h, b, c, d, a, X[6], 23, t[44])
	a = z(h, a, b, c, d, X[9], 4, t[45])
	d = z(h, d, a, b, c, X[12], 11, t[46])
	c = z(h, c, d, a, b, X[15], 16, t[47])
	b = z(h, b, c, d, a, X[2], 23, t[48])

	a = z(i, a, b, c, d, X[0], 6, t[49])
	d = z(i, d, a, b, c, X[7], 10, t[50])
	c = z(i, c, d, a, b, X[14], 15, t[51])
	b = z(i, b, c, d, a, X[5], 21, t[52])
	a = z(i, a, b, c, d, X[12], 6, t[53])
	d = z(i, d, a, b, c, X[3], 10, t[54])
	c = z(i, c, d, a, b, X[10], 15, t[55])
	b = z(i, b, c, d, a, X[1], 21, t[56])
	a = z(i, a, b, c, d, X[8], 6, t[57])
	d = z(i, d, a, b, c, X[15], 10, t[58])
	c = z(i, c, d, a, b, X[6], 15, t[59])
	b = z(i, b, c, d, a, X[13], 21, t[60])
	a = z(i, a, b, c, d, X[4], 6, t[61])
	d = z(i, d, a, b, c, X[11], 10, t[62])
	c = z(i, c, d, a, b, X[2], 15, t[63])
	b = z(i, b, c, d, a, X[9], 21, t[64])

	return bit_and(A + a, 0xFFFFFFFF),
		bit_and(B + b, 0xFFFFFFFF),
		bit_and(C + c, 0xFFFFFFFF),
		bit_and(D + d, 0xFFFFFFFF)
end

----------------------------------------------------------------

function md5.update(self: MD5, s: string)
	self.pos = self.pos + #s
	s = self.buf .. s
	for ii = 1, #s - 63, 64 do
		local X = cut_le_str(
			string.sub(s, ii, ii + 63),
			4,
			4,
			4,
			4,
			4,
			4,
			4,
			4,
			4,
			4,
			4,
			4,
			4,
			4,
			4,
			4
		)
		-- assert(#X == 16)
		X[0] = table.remove(X, 1) -- zero based!
		self.a, self.b, self.c, self.d =
			transform(self.a, self.b, self.c, self.d, X)
	end
	self.buf = string.sub(s, math.floor(#s / 64) * 64 + 1, #s)
	return self
end

function md5.finish(self: MD5)
	local msgLen = self.pos
	local padLen = 56 - msgLen % 64

	if msgLen % 64 > 56 then
		padLen += 64
	end

	if padLen == 0 then padLen = 64 end

	local s = string.char(128)
		.. string.rep(string.char(0), padLen - 1)
		.. lei2str(bit_and(8 * msgLen, 0xFFFFFFFF))
		.. lei2str(math.floor(msgLen / 0x20000000))

	self:update(s)

	-- assert(self.pos % 64 == 0)
	return lei2str(self.a)
		.. lei2str(self.b)
		.. lei2str(self.c)
		.. lei2str(self.d)
end

----------------------------------------------------------------

function md5.new()
	local self: MD5 = {
		a = CONSTS[65],
		b = CONSTS[66],
		c = CONSTS[67],
		d = CONSTS[68],
		pos = 0,
		buf = "",
		new = md5.new,
		update = md5.update,
		finish = md5.finish,
		tohex = md5.tohex,
		sum = md5.sum,
		sumhexa = md5.sumhexa,
	}

	return self :: MD5
end

function md5.tohex(s)
	return string.format(
		"%08x%08x%08x%08x",
		str2bei(string.sub(s, 1, 4)),
		str2bei(string.sub(s, 5, 8)),
		str2bei(string.sub(s, 9, 12)),
		str2bei(string.sub(s, 13, 16))
	)
end

function md5.sum(s) return md5.new():update(s):finish() end

function md5.sumhexa(s) return md5.tohex(md5.sum(s)) end

local m = md5.new()
m:update "some bytes"
m:update "some more bytes"
m:update "etc"
local hres = md5.tohex(m:finish())
print(hres)
-- def ly ;)
