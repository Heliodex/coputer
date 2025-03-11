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

local cond = true
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

-- testing byteswap
print(bit32.byteswap(0x10203040))
print(bit32.byteswap(0))
print(bit32.byteswap(-1))

-- bit32.bor(n, 0) must clear top bits
-- we check this obscuring the constant through a global to make sure this gets evaluated fully
local high32 = 0x42_1234_5678
print(bit32.bor(high32, 0))

p(bit32.replace(-1, 0, 1, 2))

