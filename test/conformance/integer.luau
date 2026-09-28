-- Integer literals with the `i` suffix (a distinct 64-bit integer type)
local a = 1i
local b = 0xFFi
local c = 0b1010i
local d = 1_000_000i
local e = -5i
local f = 0XAi

assert(a == 1i)
assert(b == 255i)
assert(c == 10i)
assert(d == 1000000i)
assert(e == -5i)
assert(f == 10i)
