local r = { 1, 2, 3, 4, 5 }

print(table.remove(r, 1))
table.remove(r, 5)
table.remove(r, 6)

for i, v in r do
	print(i, v)
end
