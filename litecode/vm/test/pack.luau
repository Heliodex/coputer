-- manual pack

local function pack(first, ...)
	if not first then return {} end
	local t = pack(...)
	table.insert(t, 1, first)
	return t
end
local function foo(...) return pack(...) end
print(#foo(0, 1, 2))
