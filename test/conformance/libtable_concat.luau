local t = {
	"one",
	"two",
	"three",
	"four",
	a = 1,
}

print(table.concat(t))
print(table.concat(t, ", "))

local c = { "hello", "to", "the", "world" }

print(table.concat(c))
print(table.concat(c, ", "))

local created2 = table.create(5, "hllo")
print(table.concat(created2))

print(table.concat(created2, ", ", 5, 4)) -- ""

-- TODO: test fractional i arg

