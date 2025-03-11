local nan = 0 / 0
local inf = 1 / 0

print(math.min(1, 99, 56, 22))
print(math.min(50))
print(math.max(1, 99, 56, 22))
print(math.max(50))

print(math.max(1, 99, 56, 22, inf))
print(math.max(1, 99, 56, 22, -inf))
print(math.max(1, 99, 56, 22, nan))


print(math.min(1, 99, 56, 22, inf))
print(math.min(1, 99, 56, 22, -inf))
print(math.min(1, 99, 56, 22, nan))

