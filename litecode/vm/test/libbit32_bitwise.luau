-- lua

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
