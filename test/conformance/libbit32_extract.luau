local function p(s) print(s) end

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
