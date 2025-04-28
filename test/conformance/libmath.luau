local x = 5.5

local nan = 0/0
local inf = 1/0

print(nan)
print(inf)

print(math.abs(-x))

print(math.pi, math.pi / math.pi)

print(math.sin(math.pi / 2))
print(math.cos(math.pi))
-- print(math.tan(1)) -- uhh
print(math.asin(1))
print(math.acos(1))
print(math.atan(1))

print(math.atan2(1, -2))

print(math.deg(math.pi))
print(math.deg(math.pi * x))


print(math.modf(5))
print(math.modf(x))

print(math.pow(5, 2))
print(math.pow(5, 2.5))
print(math.round(x))

print(math.sqrt(25))
print(math.sqrt(24))

print(math.huge)

print(math.map(0.5, 0, 1, 6, 99))
