local t: { number } = {}

for i = 10, 1, -1 do -- murders the table implementation
	t[i] = i
end

for i, v in t do
	print(i, v)
end
