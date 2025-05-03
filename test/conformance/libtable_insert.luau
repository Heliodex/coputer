local g = { "a", "b", "c", "d", "e", [7] = "g" }

local f = {}
table.insert(f, 1, 1)
print(f[1])

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

local fi = {}
table.insert(fi, 1, "whats up")
table.insert(fi, 6.5, "whats up")

print(fi[1], fi[6.5])
