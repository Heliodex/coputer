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

print(tonumber {})
print(tonumber { 1 })
print(tonumber(false))
print(tonumber(true))
print(tonumber "1.5")
print(tonumber "15")
print(tonumber "00000000000000000000000000000015")
print(tonumber "-5000")
print(tonumber "1e5")
print(tonumber "1E5")
print(tonumber "-1E5")
print(tonumber("d", 16))
-- print(tonumber("-1", 16))
-- print(tonumber("-0x1", 16))
print(tonumber("0deadbeef", 16))
print(tonumber("FAFF", 16))
print(tonumber("0b1111", 2))
print(tonumber("1111", 2))
print(tonumber("1211", 3))
-- print(tonumber("-1111", 2))
