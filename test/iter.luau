local function iter(a, i: number): (number?, any?)
	i += 1
	local v = a[i]
	if v then return i, v end
	return
end

local function p(a) return iter, a, 0 end

for i, v in p { 1, 2, 3 } do
	print(i, v)
end
for i, v in iter, { 4, 5, 6 }, 0 do
	print(i, v)
end
