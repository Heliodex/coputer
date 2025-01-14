-- upvalues, loops, continue
local uv
do
	local i = 0
	function uv()
		print(i)
		i += 1
		print(i)
	end
end

uv()

print((function()
	local res = {}

	for i = 1, 5 do
		i *= 2
		-- phantom bug disappeared when i was debugging it idk what i changed
		-- maybe it was the pointers? idk, pesky pointers
		res[#res + 1] = function() return i end
	end

	local sum = 0
	for _, f in pairs(res) do
		sum += f()
	end

	return sum
end)())

print((function()
	local res = {}

	for i in ipairs { 1, 2, 3, 4, 5, 6, 7, 8, 9, 10 } do
		res[#res + 1] = function() return i end
		if i == 5 then continue end
		i *= 2
	end

	local sum = 0
	for _, f in pairs(res) do
		sum += f()
	end

	return sum
end)())

print((function()
	local res = {}

	local i = 1
	while i <= 10 do
		local j = i
		res[#res + 1] = function() return j end
		if i == 5 then
			i += 1
			continue
		end
		i += 1
		j *= 2
	end

	local sum = 0
	for _, f in pairs(res) do
		sum += f()
	end

	return sum
end)())

print((function()
	local res = {}

	local i = 1
	repeat
		local j = i
		res[#res + 1] = function() return j end
		if i == 5 then
			i += 1
			continue
		end
		i += 1
		j *= 2
	until i > 10

	local sum = 0
	for _, f in pairs(res) do
		sum += f()
	end

	return sum
end)())
