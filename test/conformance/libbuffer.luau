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
buffer.writei16(b, 2, 22222)
buffer.writei32(b, 0, 99887766)

local function p(s)
	print(s) -- prevents return of more than one value
end

p(buffer.tostring(b))

p(buffer.readi8(b, 0))
p(buffer.readi8(b, 1))
p(buffer.readi8(b, 2))
p(buffer.readu8(b, 4))

-- 8
buffer.writeu8(b, 4, 2 ^ 8 - 16)
p(buffer.readi8(b, 4))
p(buffer.readu8(b, 4))

-- 16
buffer.writeu16(b, 4, 2 ^ 16 - 16)
for i = 4, 5 do
	p(buffer.readi8(b, i))
	p(buffer.readu8(b, i))
end
p(buffer.readi16(b, 4))
p(buffer.readu16(b, 4))

-- 32
buffer.writeu32(b, 4, 2 ^ 32 - 16)
for i = 4, 7 do
	p(buffer.readi8(b, i))
	p(buffer.readu8(b, i))
end
for i = 4, 6 do
	p(buffer.readi16(b, i))
	p(buffer.readu16(b, i))
end
p(buffer.readi32(b, 4))
p(buffer.readu32(b, 4))
p(buffer.readf32(b, 4))

-- f32
buffer.writef32(b, 4, 2 ^ 32 - 16)
for i = 4, 7 do
	p(buffer.readi8(b, i))
	p(buffer.readu8(b, i))
end
for i = 4, 6 do
	p(buffer.readi16(b, i))
	p(buffer.readu16(b, i))
end
p(buffer.readi32(b, 4))
p(buffer.readu32(b, 4))
p(buffer.readf32(b, 4))

-- f64
buffer.writef64(b, 4, 2 ^ 64 - 16)
for i = 4, 11 do
	p(buffer.readi8(b, i))
	p(buffer.readu8(b, i))
end
for i = 4, 10 do
	p(buffer.readi16(b, i))
	p(buffer.readu16(b, i))
end
for i = 4, 8 do
	p(buffer.readi32(b, i))
	p(buffer.readu32(b, i))
end
-- print(buffer.readf64(b, 4)) -- bruh

print()
for i = 0, 32 do
	p(buffer.readbits(b, 4, i))
end
for i = 0, 32 do
	buffer.writebits(b, 4, i, 0b10101010)
end
for i = 0, 32 do
	p(buffer.readbits(b, 4, i))
end

p(buffer.readstring(b, 4, 8))
buffer.writestring(b, 4, "hellooooooo", 5)
buffer.writestring(b, 8, "epic")
p(buffer.tostring(b)) -- pai\n

local b2 = buffer.create(12)
buffer.fill(b2, 4, 97, 5)
p(buffer.tostring(b2))
buffer.fill(b2, 0, 65)
p(buffer.tostring(b2))

buffer.copy(b, 1, b2, 4, 5)
p(buffer.tostring(b))
