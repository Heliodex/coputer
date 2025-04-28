local created = table.create(5)
for i = 1, 5 do
	created[i] = i
end
for i, v in created do
	print(i, v)
end

print(#created)
