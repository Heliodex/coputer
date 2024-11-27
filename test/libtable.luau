local t = {
	"one",
	a = 1,
}

local function tt()
	print(t[1])
	print(t.a)
	print(#t)
end

tt()
table.clear(t)
tt()

print()

t[2] = "two"
local t2 = t
local t3 = table.clone(t)
t2[2] = "three"
t3[2] = "four"

print(t[2])
print(t2[2])
print(t3[2])

print(table.concat(t))
print(table.concat(t, ", "))

local c = { "hello", "to", "the", "world" }

print(table.concat(c))
print(table.concat(c, ", "))

local f = {}
table.insert(f, 1, 1)
print(f[1])

local g = { "a", "b", "c", "d", "e", [7] = "g" }
table.insert(g, 3, "z")

for i, v in g do
	print(i, v)
end

print(table.isfrozen(math))
print(table.isfrozen({}))
