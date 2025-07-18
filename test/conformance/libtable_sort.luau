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

local t = { 9, 8, 7, 6, 7, 8, 9 }

table.sort(t, function(a, b) return end)
for i, v in t do
	print(i, v)
end

table.sort(t, function(a, b) return nil end)
for i, v in t do
	print(i, v)
end
