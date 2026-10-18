-- pairs/next iteration over both table parts: each next call resumes the table's deterministic order
local list = {}
for i = 1, 128 do
	list[i] = i * 3
end

local hash = {}
for i = 1, 128 do
	hash[tostring(i)] = i
end

local sum = 0
for _ = 1, 1000 do
	for _, v in pairs(list) do
		sum += v
	end
	for _, v in pairs(hash) do
		sum += v
	end
end

print(sum)
