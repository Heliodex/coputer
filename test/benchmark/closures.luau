-- closure creation and upvalue capture, then calling each closure
local function make(n)
	return function() return n end
end

local sum = 0
for i = 1, 50000 do
	local f = make(i)
	sum += f()
end

print(sum)
