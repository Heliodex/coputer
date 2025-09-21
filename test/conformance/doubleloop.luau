local x = 0
local a = { 1, 2, 3 }
for _, v in a do
	print(v)
	x += 1
	break
end
for _, v in a do
	print(v)
	x += 2
	break
end
print(x)
