local max = 18-1
local start = max // 2 + 1

for i = max, math.huge, max do
	local found = true
	for j = start, max do
		if i % j ~= 0 then
			found = false
			break
		end
	end

	if found then
		print(i)
		return
	end
end
