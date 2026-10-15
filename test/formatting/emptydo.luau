local x = 1
do
end
do end
do
	do end
end
do
	-- hoisted comment
end
local function f()
	do end
	return x
end
do
	print(x)
end
print(x)
