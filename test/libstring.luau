local s = "hello 世界"

print(string.byte(s, 5, 7))
print(string.byte(s, 9, 99))
print(string.byte(s, -5, 5))
print(string.byte(s, -6, 6))
print(string.byte(s, -7, 7))

print(string.byte(s, 3, -6))
print(string.byte(s, 4, -6))
print(string.byte(s, 5, -6))
print(string.byte(s, 6, -6))
print(string.byte(s, 7, -7))

print(string.char())
print(string.char(104, 101, 108, 108, 111, 32, 228, 184, 150, 231, 149, 140))
