for i = 0, 5 do
	print(i)
end

for i = 1, 10, 2 do
	print(i)
end

for i = 1, 4, 0.3 do
	print(i)
end

for i = 1, 1, 0 do
	print(i)
	break
end

print()

local nan = 0 / 0
local inf = 1 / 0
for i = 1, 5, nan do
	print(i)
end
for i = 5, 1, nan do
	print(i)
end
for i = 1, 0 do
	print(i)
end

for i = inf, 1, -1 do
	print(i)
	break
end
for i = 1, inf do
	if i < 5000 then continue end
	print(i)
	break
end
