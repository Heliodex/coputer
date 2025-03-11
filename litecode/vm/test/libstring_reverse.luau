local s = "hello 世界"

print(string.reverse "hello")
print(string.reverse "世界") -- doesn't work, but conforms with reference implementation
print(string.reverse "")
print(string.reverse(s))
