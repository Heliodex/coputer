local s = "hello 世界"

print(string.find(s, "hello"))
print(string.find(s, "hello", -12))
print(string.find(s, "世界", 7))
print(string.find(s, "世界", -6))

print(string.find(s, "hello", 3))
print(string.find(s, "hello", -11))
print(string.find(s, "世界", 8))
print(string.find(s, "世界", -5))
print(string.find(s, "sup"))
