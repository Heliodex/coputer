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

for _, test in tests3 do
	for _, func in funcs3 do
		print(func(test))
	end
	print(bit32.extract(test, 4))
	print(bit32.extract(test, 4, 4))
	print(bit32.extract(test, 4, 8))
	print(bit32.extract(test, 0, 8))
end

for _, t1 in tests3 do
	for _, t2 in tests3 do
		bit32.replace(t1, t2, 4)
		bit32.replace(t1, t2, 4, 4)
		bit32.replace(t1, t2, 4, 8)
		bit32.replace(t1, t2, 0, 8)
	end
end
