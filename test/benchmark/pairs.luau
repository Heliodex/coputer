-- pairs/next iteration: each next call walks the table's deterministic order
local t = {}
for i = 1, 128 do
	t[i] = i * 3
end

local sum = 0
for _ = 1, 1000 do
	for _, v in pairs(t) do
		sum += v
	end
end

print(sum)
