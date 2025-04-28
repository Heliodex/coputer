local s = "hello 世界"

print(utf8.codepoint(s))
print(utf8.codepoint(s, 3))
print(utf8.codepoint(s, 5, 7))
print(utf8.codepoint(s, 3, 6))
print(utf8.codepoint(s, 3, -6))
print(utf8.codepoint(s, 4, -6))
print(utf8.codepoint(s, 5, -6))
print(utf8.codepoint(s, 6, -6))
print(utf8.codepoint(s, 7, -6))
