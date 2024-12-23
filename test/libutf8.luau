local s = "hello 世界"

print(utf8.char())
print(utf8.char(104, 101, 108, 108, 111, 32, 228, 184, 150, 231, 149, 140))
print(utf8.char(0x4e16, 0x754c))

for i, v in utf8.codes(s) do
	print(i, v)
end

print(utf8.codepoint(s))
print(utf8.codepoint(s, 3))
print(utf8.codepoint(s, 5, 7))
print(utf8.codepoint(s, 3, 6))
print(utf8.codepoint(s, 3, -6))
print(utf8.codepoint(s, 4, -6))
print(utf8.codepoint(s, 5, -6))
print(utf8.codepoint(s, 6, -6))
print(utf8.codepoint(s, 7, -6))

print(utf8.len(s))

print(utf8.offset(s, 1, 1))
print(utf8.offset(s, 1, -11))
print(utf8.offset(s, 1, 5))
print(utf8.offset(s, 6, 1))
