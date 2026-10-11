-- Explicit type instantiation syntax, `expr<<T...>>`
local function identity<T>(x: T): T
	return x
end

local a = identity<<number>>(1)
print(a)

local function typePacks<T...>(...: T...): T...
	return ...
end

local b, c = typePacks<<(string, number)>>(1, "a")
print(b, c)

local t = {}
function t:method<T>(x: T): T
	return x
end

local d = t:method<<number>>(1)
print(d)
