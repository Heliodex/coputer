local s = "hello 世界"

for i, v in utf8.codes(s) do
	print(i, v)
end

print(utf8.len(s))

print(utf8.charpattern)
