-- make sure custom iterators work! example is from PIL 7.1
local function list_iter(a)
	local i = 0
	local n = #a
	return function()
		i += 1
		return if i <= n then a[i] else nil
	end
end

local a = ""
for e in list_iter { 4, 2, 1 } do
	print(e)
	a ..= tostring(e)
end
print(a)
