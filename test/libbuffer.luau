local b = buffer.create(10)
print(buffer.tostring(b), buffer.len(b))

local s = buffer.fromstring("hello")
print(buffer.tostring(s), buffer.len(s))
