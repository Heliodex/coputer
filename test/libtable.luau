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

local created = table.create(5)
for i = 1, 5 do
	created[i] = i
end
for i, v in created do
	print(i == v)
end

local created2 = table.create(5, "hllo")
print(table.concat(created2))

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
print(table.isfrozen(table.freeze({})))

local x = {
	[1] = 1,
	[2] = 2,
	[3] = 3,
	[7] = 7,
	[29] = 29,
	[29.7] = 29.7,
	[-math.huge] = 29,
	[123456789] = 29,
}

print(x[1])
print(x[2])
print(x[3])
print(x[7])
print(x[29])
print(x[29.7])
print(table.maxn(x))
print(#x)
