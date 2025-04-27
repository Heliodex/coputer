local function test(c, ...)
	local args = { ... }
	for _, v in args do
		print(v)
	end

	local b = args[1]
	print(b)
	local a = args[2]
	print(a)
	return a, b, c
end

local a, b, c = test(1, 2, 0)
print(a, b, c)
