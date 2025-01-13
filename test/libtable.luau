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

print(table.find(g, "c"))
print(table.find(g, "h"))
print(table.find(g, "c", 3))
print(table.find(g, "c", 4))

for i = 1, 10 do
	table.insert(g, 3, "z")
end

table.insert(g, 212, "z")
table.insert(g, 212, "A")
table.insert(g, "h")

for i, v in g do
	print(i, v)
end

local h = {}
table.insert(h, "H") -- h
table.insert(h, -5, "G")
print(h[1])
print(h[-5])

local n = { 3 }

print(table.unpack(n))
table.insert(n, 1, 2)
print(table.unpack(n))
table.insert(n, 1, 1)
print(table.unpack(n))
table.insert(n, 1, 0)
print(table.unpack(n))

print(table.isfrozen(math))
print(table.isfrozen {})
print(table.isfrozen(table.freeze {}))

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

local m = { 1, 2, 3, 4, 5, 6, 7 }
local d = { 11, 12, 13, 14, 15, 16, 17 }

table.move(m, 1, 3, 2, d)
table.move(m, 1, 3, 17, d)
-- table.move(m, 1, 3, -6, d) -- works! but untestable

for i, v in d do
	print(i, v)
end

local p = table.pack(1, 2, 3, 4, 5, 6, 7)
print(p[1])
print(p.n)

local r = { 1, 2, 3, 4, 5 }

print(table.remove(r, 1))
table.remove(r, 5)
table.remove(r, 6)

for i, v in r do
	print(i, v)
end

local s1 = { 1, 5, 4, 3, 2, 7, 6, bruh = 6 }

table.sort(s1)
for i, v in s1 do
	print(i, v)
end

local s2 = { "j", "i", "h", "g", "f", "e", "d", "c", "b", "a", bruh = 6 }
table.sort(s2)
print(table.concat(s2))

local s3 = { 1, 2, 3, 4, 5, 6, 7, bruh = 6 }
table.sort(s3, function(a, b) return a > b end)
for i, v in s3 do
	print(i, v)
end

local u = { 1, 2, 3, 4, 5, 6, 7, [9] = 6, bruh = 6 }
print(table.unpack(u))
print(table.unpack(u, 1, 5))
print(table.unpack(u, 6, 10))
