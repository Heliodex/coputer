local t = table.create(2e7, 0)

for i = 1, 2e7 do
	t[i] = i
end

print(t[5242880])
