-- `if local` / `if const` statements and expressions
local function classify(x)
	return if local y = x then y else 0
end

if local z = 42 then
	assert(z == 42)
end

if const w = 7 then
	assert(w == 7)
end

local tookElse = false
if local q = nil then
	error("unreachable")
else
	tookElse = true
end
assert(tookElse)

assert(classify(5) == 5)
assert(classify(nil) == 0)
