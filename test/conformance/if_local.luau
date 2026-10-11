-- `if local` / `if const` statements and expressions
local function classify(x)
	return if local y = x then y else 0
end

if local z = 42 then
	print(z)
end

if const w = 7 then
	print(w)
end

local tookElse = false
if local q = nil then
	error("unreachable")
else
	tookElse = true
end
print(tookElse)

print(classify(5))
print(classify(nil))
