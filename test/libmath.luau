local x = 5.5

local nan = 0/0
local inf = 1/0

print(nan)
print(inf)

print(math.floor(x))
print(math.floor(-x))
print(math.abs(-x))

print(math.pi, math.pi / math.pi)

print(math.sin(math.pi / 2))
print(math.cos(math.pi))
-- print(math.tan(1)) -- uhh
print(math.asin(1))
print(math.acos(1))
print(math.atan(1))

print(math.atan2(1, -2))

print(math.clamp(0, 1, 3))
print(math.clamp(2, 1, 3))
print(math.clamp(4, 1, 3))

print(math.max(1, 99, 56, 22))
print(math.max(50))
print(math.min(1, 99, 56, 22))
print(math.min(50))

print(math.max(1, 99, 56, 22, inf))
print(math.max(1, 99, 56, 22, -inf))
print(math.max(1, 99, 56, 22, nan))
print(math.min(1, 99, 56, 22, inf))
print(math.min(1, 99, 56, 22, -inf))
print(math.min(1, 99, 56, 22, nan))

print(math.deg(math.pi))
print(math.deg(math.pi * x))
print(math.fmod(5, 2))
print(math.fmod(5, -2))
print(math.fmod(-5, 2))
print(math.fmod(-5, -2))

print(math.frexp(5))
print(math.frexp(-5))
print(math.frexp(0))
print(math.frexp(math.pi))

print(math.ldexp(5, 2))
print(math.ldexp(-5, 2))
print(math.ldexp(-5, 2.1))
print(math.ldexp(-5, 2.9))
print(math.ldexp(-5, 3))

print(math.log(5))
print(math.log(5, 2))
-- print(math.log10(50_000))
-- print(math.log10(100_000))
print(math.log(50_000, 10))
print(math.log(100_000, 10))
print(math.modf(5))
print(math.modf(x))

print(math.noise(1))
print(math.noise(1, 2))
print(math.noise(1, 2, 3))
print(math.noise(1.158))
print(math.noise(-1.158))
print(math.noise(1.158, 5.723))
print(math.noise(1.158, 6))
print(math.noise(291.158, 296))
print(math.noise(-291.158, -296))
print(math.noise(5e5 / 7, 5e6 / 9))

print(math.pow(5, 2))
print(math.pow(5, 2.5))
print(math.round(x))
print(math.sign(0))
print(math.sign(-0))
print(math.sign(1))
print(math.sign(-1))

print(math.sqrt(25))
print(math.sqrt(24))

print(math.huge)

print(math.map(0.5, 0, 1, 6, 99))

