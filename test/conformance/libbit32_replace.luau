local function p(s) print(s) end

p(bit32.replace(0x12345678, 5, 28, 4))
p(bit32.replace(0x12345678, 0x87654321, 0, 32))
p(bit32.replace(0, 1, 2))
p(bit32.replace(0, -1, 4))
p(bit32.replace(-1, 0, 31))
p(bit32.replace(-1, 0, 1, 2))
