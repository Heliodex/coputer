local m = { 1, 2, 3, 4, 5, 6, 7 }
local d = { 11, 12, 13, 14, 15, 16, 17 }

table.move(m, 1, 3, 2, d)
table.move(m, 1, 3, 17, d)
-- table.move(m, 1, 3, -6, d) -- works! but untestable

for i, v in d do
	print(i, v)
end
