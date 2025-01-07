-- IT'S THE RETURN OF THE-
-- oh wait, no wait, you're kidding
-- he didn't just say what I think he did, did he?

local longest = 0
local longestl = 0

for i = 1, 3e4 do
	local c, l = i, 0

	while c ~= 1 do
		l += 1
		c = if c % 2 == 0 then c / 2 else 3 * c + 1
	end
	if l > longestl then
		longest, longestl = i, l
	end
end

print(longest)
