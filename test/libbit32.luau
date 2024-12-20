print(bit32.arshift(100, 1))
print(bit32.arshift(101, -1))
for i = 4, -4, -1 do
	print(bit32.arshift(0x12345678, i))
	print(bit32.arshift(0xFFFFFFFF, i))
	print(bit32.arshift(0x7FFFFFFF, i))
end
print(bit32.arshift(100.5, 6))
print(bit32.arshift(100.5, -21))
print(bit32.arshift(100.5, 99))
print(bit32.arshift(100.5, -99))
print(bit32.arshift(-100.5, 6))
print(bit32.arshift(-100.5, -21))
print(bit32.arshift(-100.5, 99))
print(bit32.arshift(-100.5, -99))

local funcs: { (number, ...number) -> any } = {
	bit32.band,
	bit32.bor,
	bit32.btest,
	bit32.bxor,
}
local tests = {
	{ 0x12345678, 0x87654321 },
	{ 2 ^ 4 - 1, 2 ^ 5 - 1, 2 ^ 6 - 1, 2 ^ 7 - 1 },
	{ 2 ^ 4, 2 ^ 5, 2 ^ 6, 2 ^ 7 },
	{ 1 - 2 ^ 4, 1 - 2 ^ 5, 1 - 2 ^ 6, 1 - 2 ^ 7 },
	{ 0, 0x87654321 },
}

for _, func in funcs do
	for _, test in tests do
		print(func(table.unpack(test)))
	end
end

print(bit32.bnot(0x12345678))
print(bit32.bnot(-0x12345678))

local funcs2: { (number) -> any } = {
	bit32.byteswap,
	bit32.countlz,
	bit32.countrz,
}
local tests2 = {
	0x12345678,
	0xffff0000,
	0x87654321,
	5,
	-60,
}

for _, func in funcs2 do
	for _, test in tests2 do
		print(func(test))
	end
end
