-- local j = {}
-- local k = function() end

local table = {
	"this",
	"is",
	"a",
	"test",
	"of",
	"a",
	"table",
	-- string keys
	key = "test",
	key1 = "test2",
	-- hash keys
	[-1] = "test3",
	-- [0] = "test4",
	-- [9] = "test5",
	-- [10] = "test6",
	-- [3.14159] = "pi",
	-- [j] = "tkey",
	-- [k] = "fkey",
}

for i, v in table do
	print(i, v)
end

print(#table)
-- print(table[j], table[k])
