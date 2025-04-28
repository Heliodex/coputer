local a = "Hello "
local b = a .. "World"
local c: boolean? = false

if #b == 11 then
	c = true
elseif c == nil then
	print "nil"
end

print(b, c)
