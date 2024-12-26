print(type(nil))
print(type(1))
print(type "hello")
print(type(true))
print(type {})
print(type(function() end))
print(type(coroutine.create(function() end)))
print(type(buffer.create(0)))

local t = { "one", "two", "three", four = 4 }
for i, v in ipairs(t) do
	print(i, v)
end
for i, v in next, t do
	print(i, v)
end
for i, v in pairs(t) do
	print(i, v)
end
