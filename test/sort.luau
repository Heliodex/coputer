-- force quicksort to degrade to heap sort
-- discover quick sort killer (this triggers heap sort which is what we want more or less; note that the "internal" heap sort iterations will result in a different order that wouldn't fully defeat a vanilla quicksort)
-- see https://igoro.com/archive/quicksort-killer/
local keys = {}
local candidate = 0
local nx = 0

local t = table.create(100, 0)
for k in t do
	t[k] = k
end

table.sort(t, function(x, y)
	if keys[x] == nil and keys[y] == nil then
		if x == candidate then
			keys[x] = nx
		else
			keys[y] = nx
		end
		nx += 1
	end

	if keys[x] == nil then
		candidate = x
		return true
	elseif keys[y] == nil then
		candidate = y
		return false
	end

	return keys[x] < keys[y]
end)

-- repeat the sort for the generated sequence; it should produce an integer sequence and trigger heap sort, although we can't confirm the latter
local arr = table.create(#t)
for k, v in t do
	arr[v] = k
end

table.sort(arr)
for k, v in arr do
	if k ~= v then print(k, v) end
end
