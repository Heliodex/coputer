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

print(string.find(s, "hello"))
print(string.find(s, "hello", -12))
print(string.find(s, "世界", 7))
print(string.find(s, "世界", -6))

print(string.find(s, "hello", 3))
print(string.find(s, "hello", -11))
print(string.find(s, "世界", 8))
print(string.find(s, "世界", -5))
print(string.find(s, "sup"))

print(string.format("an asterisk: %c", 42))
print(string.format("%d %i", 42, 42))
print(string.format("%d %i", -42, -42))
print(string.format("the answer is %d!", 42))
-- print(string.format("%e", 42))
-- print(string.format("%E", 42))
print(string.format("%f", 42.5))
print(string.format("%f", -42.5))
print(string.format("%o", 25))
print(string.format("%x %X", 255, 255))
print(string.format("%o", -25))
print(string.format("%x %X", -255, -255)) -- LOL
print(string.format("%s", "whatsup"))
print(string.format("%q", "whatsup"))
print(string.format("%q", 'print "hello"'))
print(string.format("%q", [[
	print "hello"
]]))
print(string.format("%u", 42))
print(string.format("%u", -42))
print(string.format("%*", "anythingyyy"))
print(string.format("%*", {}))
