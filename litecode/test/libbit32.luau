local funcs1: { (number, number) -> any } = {
	bit32.arshift,
	bit32.lrotate,
	bit32.lshift,
	bit32.rrotate,
	bit32.rshift,
}

for _, func in funcs1 do
	print(func(100, 1))
	print(func(101, -1))
	for i = 4, -4, -1 do
		print(func(0x12345678, i))
		print(func(0xFFFFFFFF, i))
		print(func(0x7FFFFFFF, i))
	end
	print(func(100.5, 6))
	print(func(100.5, -21))
	print(func(100.5, 99))
	print(func(100.5, -99))
	print(func(-100.5, 6))
	print(func(-100.5, -21))
	print(func(-100.5, 99))
	print(func(-100.5, -99))
end

local funcs2: { (...number) -> any } = {
	bit32.band,
	bit32.bor,
	bit32.btest,
	bit32.bxor,
}
local tests2 = {
	{ 0x12345678, 0x87654321 },
	{ 2 ^ 4 - 1, 2 ^ 5 - 1, 2 ^ 6 - 1, 2 ^ 7 - 1 },
	{ 2 ^ 4, 2 ^ 5, 2 ^ 6, 2 ^ 7 },
	{ 1 - 2 ^ 4, 1 - 2 ^ 5, 1 - 2 ^ 6, 1 - 2 ^ 7 },
	{ 0, -0x87654321 },
}

for _, test in tests2 do
	for _, func in funcs2 do
		print(func(table.unpack(test)))
	end
end

print(bit32.bnot(0x12345678))
print(bit32.bnot(-0x12345678))

local funcs3: { (number) -> any } = {
	bit32.byteswap,
	bit32.countlz,
	bit32.countrz,
}
local tests3 = {
	0x12345678,
	0xffff0000,
	-0x87654321,
	5,
	-60.5,
}

local function p(s)
	print(s) -- prevents return of more than one value
end

for _, test in tests3 do
	for _, func in funcs3 do
		p(func(test))
	end
	p(bit32.extract(test, 4))
	p(bit32.extract(test, 4, 4))
	p(bit32.extract(test, 4, 8))
	p(bit32.extract(test, 0, 8))
	for _, test2 in tests3 do
		bit32.replace(test, test2, 4)
		bit32.replace(test, test2, 4, 4)
		bit32.replace(test, test2, 4, 8)
		bit32.replace(test, test2, 0, 8)
	end
end

-- lua

-- This file is part of the Luau programming language and is licensed under MIT License; see LICENSE.txt for details
-- This file is based on Lua 5.x tests -- https://github.com/lua/lua/tree/master/testes
print "testing bitwise operations"

print(bit32.band() == bit32.bnot(0))
print(bit32.btest())
print(bit32.bor())
print(bit32.bxor())

print(bit32.band() == bit32.band(0xffffffff))
print(bit32.band(1, 2))

-- out-of-range numbers
print(bit32.band(-1))
print(bit32.band(2 ^ 33 - 1))
print(bit32.band(-2 ^ 33 - 1))
print(bit32.band(2 ^ 33 + 1))
print(bit32.band(-2 ^ 33 + 1))
print(bit32.band(-2 ^ 40))
print(bit32.band(2 ^ 40))
print(bit32.band(-2 ^ 40 - 2))
print(bit32.band(2 ^ 40 - 4))

print(bit32.lrotate(0, -1))
print(bit32.lrotate(0, 7))
print(bit32.lrotate(0x12345678, 4))
print(bit32.rrotate(0x12345678, -4))
print(bit32.lrotate(0x12345678, -8))
print(bit32.rrotate(0x12345678, 8))
print(bit32.lrotate(0xaaaaaaaa, 2))
print(bit32.lrotate(0xaaaaaaaa, -2))

local cond = true -- lol just to prevent reams of outputs over long loops
for i = -50, 50 do
	cond = cond
		and (bit32.lrotate(0x89abcdef, i) == bit32.lrotate(0x89abcdef, i % 32))
end
if not cond then print "ERROR" end

print(bit32.lshift(0x12345678, 4))
print(bit32.lshift(0x12345678, 8))
print(bit32.lshift(0x12345678, -4))
print(bit32.lshift(0x12345678, -8))
print(bit32.lshift(0x12345678, 32))
print(bit32.lshift(0x12345678, -32))
print(bit32.rshift(0x12345678, 4))
print(bit32.rshift(0x12345678, 8))
print(bit32.rshift(0x12345678, 32))
print(bit32.rshift(0x12345678, -32))
print(bit32.arshift(0x12345678, 0))
print(bit32.arshift(0x12345678, 1))
print(bit32.arshift(0x12345678, -1))
print(bit32.arshift(-1, 1))
print(bit32.arshift(-1, 24))
print(bit32.arshift(-1, 32))
print(bit32.arshift(-1, -1) == (-1 * 2) % 2 ^ 32)

print "+"
-- some special cases
local c = {
	0,
	1,
	2,
	3,
	10,
	0x80000000,
	0xaaaaaaaa,
	0x55555555,
	0xffffffff,
	0x7fffffff,
}

cond = true
for i, b in c do
	cond = cond and (bit32.band(b) == b)
	cond = cond and (bit32.band(b, b) == b)
	cond = cond and (bit32.btest(b, b) == (b ~= 0))
	cond = cond and (bit32.band(b, b, b) == b)
	cond = cond and (bit32.btest(b, b, b) == (b ~= 0))
	cond = cond and (bit32.band(b, bit32.bnot(b)) == 0)
	cond = cond and (bit32.bor(b, bit32.bnot(b)) == bit32.bnot(0))
	cond = cond and (bit32.bor(b) == b)
	cond = cond and (bit32.bor(b, b) == b)
	cond = cond and (bit32.bor(b, b, b) == b)
	cond = cond and (bit32.bxor(b) == b)
	cond = cond and (bit32.bxor(b, b) == 0)
	cond = cond and (bit32.bxor(b, 0) == b)
	cond = cond and (bit32.bxor(b, b, b) == b)
	cond = cond and (bit32.bxor(b, b, b, b) == 0)
	cond = cond and (bit32.bnot(b) ~= b)
	cond = cond and (bit32.bnot(bit32.bnot(b)) == b)
	cond = cond and (bit32.bnot(b) == 2 ^ 32 - 1 - b)
	cond = cond and (bit32.lrotate(b, 32) == b)
	cond = cond and (bit32.rrotate(b, 32) == b)
	cond = cond
		and (
			bit32.lshift(bit32.lshift(b, -4), 4)
			== bit32.band(b, bit32.bnot(0xf))
		)
	cond = cond
		and (
			bit32.rshift(bit32.rshift(b, 4), -4)
			== bit32.band(b, bit32.bnot(0xf))
		)
	for j = -40, 40 do
		cond = cond and (bit32.lshift(b, j) == math.floor((b * 2 ^ j) % 2 ^ 32))
	end

	if not cond then print("ERROR", i, b) end
end

print "+"

-- testing extract/replace

p(bit32.extract(0x12345678, 0, 4))
p(bit32.extract(0x12345678, 4, 4))
p(bit32.extract(0xa0001111, 28, 4))
p(bit32.extract(0xa0001111, 31, 1))
p(bit32.extract(0x50000111, 31, 1))
p(bit32.extract(0xf2345679, 0, 32))
p(bit32.extract(0xa0001111, 0))
p(bit32.extract(0xa0001111, 16))
p(bit32.extract(0xa0001111, 31))
p(bit32.extract(42, 1, 3))

local pos
pos = 1
p(bit32.extract(42, pos, 3)) -- test bit32.extract builtin instead of bit32.extractk

p(bit32.replace(0x12345678, 5, 28, 4))
p(bit32.replace(0x12345678, 0x87654321, 0, 32))
p(bit32.replace(0, 1, 2))
p(bit32.replace(0, -1, 4))
p(bit32.replace(-1, 0, 31))
p(bit32.replace(-1, 0, 1, 2))

-- testing countlz/countrc
print(bit32.countlz(0))
print(bit32.countlz(42))
print(bit32.countlz(0xffffffff))
print(bit32.countlz(0x80000000))
print(bit32.countlz(0x7fffffff))

print(bit32.countrz(0))
print(bit32.countrz(1))
print(bit32.countrz(42))
print(bit32.countrz(0x80000000))
print(bit32.countrz(0x40000000))
print(bit32.countrz(0x7fffffff))

-- testing byteswap
print(bit32.byteswap(0x10203040))
print(bit32.byteswap(0))
print(bit32.byteswap(-1))

-- bit32.bor(n, 0) must clear top bits
-- we check this obscuring the constant through a global to make sure this gets evaluated fully
local high32 = 0x42_1234_5678
print(bit32.bor(high32, 0))

p(bit32.replace(-1, 0, 1, 2))

-- many of the tests above go through fastcall path
-- to make sure the basic implementations are also correct we test some functions with string->number coercions
print(bit32.lrotate(0x12345678, 4))
print(bit32.rrotate(0x12345678, -4))
print(bit32.arshift(0x12345678, 1))
print(bit32.arshift(-1, 32))
print(bit32.arshift(-1, 1))
print(bit32.bnot(1))
print(bit32.band(1, 3))
print(bit32.band(1, 3))
print(bit32.band(1, 3, 5))
print(bit32.bor(1, 2))
print(bit32.bor(1, 2))
print(bit32.bor(1, 3, 5))
print(bit32.bxor(1, 3))
print(bit32.bxor(1, 3))
print(bit32.bxor(1, 3, 5))
print(bit32.btest(1, 3))
print(bit32.btest(1, 3))
print(bit32.countlz(42))
print(bit32.countrz(42))
p(bit32.extract(42, 1, 3))
print(bit32.byteswap(0xa1b2c3d4))
