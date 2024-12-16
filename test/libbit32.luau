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

print(bit32.band(0x12345678, 0x87654321))
print(bit32.band(2^4-1, 2^5-1, 2^6-1, 2^7-1))
print(bit32.band(1-2^4, 1-2^5, 1-2^6, 1-2^7))

print(bit32.bnot(0x12345678))
print(bit32.bnot(-0x12345678))
