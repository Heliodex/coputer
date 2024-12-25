local b = buffer.create(12)
print(buffer.tostring(b), buffer.len(b))

local s = buffer.fromstring "hello"
print(buffer.tostring(s), buffer.len(s))

buffer.writei8(b, 0, 119)
buffer.writei8(b, 1, 111)
buffer.writei8(b, 2, 114)
buffer.writei8(b, 3, 108)
buffer.writei8(b, 4, 100)
buffer.writei8(b, 4, 356)

print(buffer.tostring(b))

print(buffer.readi8(b, 0))
print(buffer.readi8(b, 1))
print(buffer.readi8(b, 2))
print(buffer.readu8(b, 4))

-- 8
buffer.writeu8(b, 4, 2 ^ 8 - 16)
print(buffer.readi8(b, 4))
print(buffer.readu8(b, 4))

-- 16
buffer.writeu16(b, 4, 2 ^ 16 - 16)
for i = 4, 5 do
	print(buffer.readi8(b, i))
	print(buffer.readu8(b, i))
end
print(buffer.readi16(b, 4))
print(buffer.readu16(b, 4))

-- 32
buffer.writeu32(b, 4, 2 ^ 32 - 16)
for i = 4, 7 do
	print(buffer.readi8(b, i))
	print(buffer.readu8(b, i))
end
for i = 4, 6 do
	print(buffer.readi16(b, i))
	print(buffer.readu16(b, i))
end
print(buffer.readi32(b, 4))
print(buffer.readu32(b, 4))

-- f32
buffer.writef32(b, 4, 2 ^ 32 - 16)
for i = 4, 7 do
	print(buffer.readi8(b, i))
	print(buffer.readu8(b, i))
end
for i = 4, 6 do
	print(buffer.readi16(b, i))
	print(buffer.readu16(b, i))
end
print(buffer.readi32(b, 4))
print(buffer.readu32(b, 4))
