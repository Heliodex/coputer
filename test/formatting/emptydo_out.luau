local x = 1

-- hoisted comment
local function f()
	return x
end

do
	print(x)
end

print(x)
