local function generateClearedTable(arraySize, hashSize)
	local tb: { [string | number]: number | boolean } = {}

	for i = 1, arraySize do
		tb[i] = i
	end

	for i = 1, hashSize do
		tb[tostring(i)] = i
	end

	table.clear(tb)
	return tb
end

print(#generateClearedTable(0, 0))
print(#generateClearedTable(10, 0))
print(#generateClearedTable(0, 10))
print(#generateClearedTable(10, 10))

print(next(generateClearedTable(0, 0)))
print(next(generateClearedTable(10, 0)))
print(next(generateClearedTable(0, 10)))
print(next(generateClearedTable(10, 10)))

for i = 1, 16 do
	local t1 = generateClearedTable(16, 0)
	local t2 = table.create(16)
	t1[i] = true
	t2[i] = true

	print(#t1, #t2)
end

do
	local things =
		{ "foo", "bar", "baz", "foobar", "a", "b", "c", "d", "e", "f", "g" }
	local tb = generateClearedTable(0, 20)
	local containsAll = {}
	for _, v in ipairs(things) do
		tb[v] = true
	end
	for k, _ in pairs(tb) do
		containsAll[k] = true
	end
	for _, v in ipairs(things) do
		print(tb[v])
		print(containsAll[v])
	end
end

--[[

do -- Check expanding the array part after clear
	local tb = generateClearedTable(10, 0)
	for i = 1, 40 do
		tb[i] = i
	end
	a(#tb == 40, "wrong array part size after expand")
end

do -- Check expanding the hash part after clear
	local tb = generateClearedTable(0, 10)
	for i = 1, 40 do
		tb[tostring(i)] = i
	end
	local count = 0
	for _ in pairs(tb) do
		count = count + 1
	end
	a(count == 40, "wrong hash part size after expand")
end
]]
