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
print(tonumber(nil))
print(tonumber(true))
print(tonumber(false))
print(tonumber "1.5")
print(tonumber "15")
print(tonumber "00000000000000000000000000000015")
print(tonumber "-5000")
print(tonumber "1e5")
print(tonumber "1E5")
print(tonumber "1E5.123")
print(tonumber "1.123E5")
print(tonumber "-1E5")
print(tonumber("d", 16))
-- print(tonumber("-1", 16))
-- print(tonumber("-0x1", 16))
print(tonumber "0xdeadbeef")
print(tonumber("0xdeadbeef", 10)) -- surprisingly does not return nil
print(tonumber("0xdeadbeef", 16))
print(tonumber("0deadbeef", 16))
print(tonumber("FAFF", 16))
print(tonumber("0b1111", 2))
print(tonumber("1111", 2))
-- print(tonumber("-1111", 2))
print(tonumber("1211", 3))
print(tonumber("000121", 3))
print(tonumber("whatsup", 36))

-- negative bases
print(tonumber "1234")
print(tonumber("1234", 4))
print(tonumber("-1234", 4))
print(tonumber("1234", 5))
print(tonumber("-1234", 5))
print(tonumber "0x1234")
print(tonumber "-0x1234")
print(tonumber("0x1234", 16))
print(tonumber("-0x1234", 16))
print(tonumber("1234", 15))
print(tonumber("-1234", 15))

print(tostring(nil))
print(tostring(1))
print(tostring(1 / 3))
print(tostring(1e10))
print(tostring(1e10 / 3))
print(tostring "hello")
-- other number-to-string tests are in nprint.luau, print() uses the same algorithm as tostring()

print(_VERSION)
