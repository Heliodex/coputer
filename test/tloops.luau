local j = {}
local k = function() end

local pointerKeys: { [any]: any } = {
	[j] = "tkey",
	[k] = "fkey",
}

local table = {
	"this",
	"is",
	"a",
	"test",
	"of",
	"a",
	"table",
	[8] = "gov",
	-- string keys
	key = "test3",
}

for i, v in table do
	print(i, v)
end

print(#pointerKeys)
print(#table)
print(pointerKeys[j], pointerKeys[k])
