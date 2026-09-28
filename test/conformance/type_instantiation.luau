-- Explicit type instantiation syntax, `expr<<T...>>`
local function identity<T>(x: T): T
	return x
end

local a = identity<<number>>(1)
assert(a == 1)

local function typePacks<T...>(...: T...): T...
	return ...
end

local b, c = typePacks<<(string, number)>>(1, "a")
assert(b == 1 and c == "a")

local t = {}
function t:method<T>(x: T): T
	return x
end

local d = t:method<<number>>(1)
assert(d == 1)
