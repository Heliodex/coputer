-- string-keyed table churn: exercises the hash part, its map hashing and the key sort on lookup
local n = 5000
local keys = table.create(n)
for i = 1, n do
	keys[i] = tostring(i)
end

local t = {}
local total = 0

for _ = 1, 5 do
	for i = 1, n do
		t[keys[i]] = i
	end

	local sum = 0
	for i = 1, n do
		sum += t[keys[i]]
	end
	total += sum

	for i = 1, n do
		t[keys[i]] = nil
	end
end

print(total)
