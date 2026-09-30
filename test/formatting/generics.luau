local function generic<T, U...>(x: T, ...: number): T
	return x
end
local function generic2<T, U>(x: T, ...: U): T
	return x
end

generic(x, 1, 2, 3)
generic<<string>>(x, y, z, a)
generic2(x, 1, 2, 3)
generic2<<string, number>>(x, y, z, a)
