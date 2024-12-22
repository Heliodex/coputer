local s = "hello 世界"

print(utf8.char())
print(utf8.char(104, 101, 108, 108, 111, 32, 228, 184, 150, 231, 149, 140))
print(utf8.char(0x4e16, 0x754c))

for i, v in utf8.codes(s) do
	print(i, v)
end
