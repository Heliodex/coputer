-- method calls: each call is a NAMECALL lookup followed by a CALL
local obj = { n = 0 }

function obj:inc(k)
	self.n += k
	return self.n
end

local sum = 0
for i = 1, 100000 do
	sum += obj:inc(i)
end

print(sum)
