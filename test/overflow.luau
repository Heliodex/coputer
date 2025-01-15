-- https://www.lua.org/bugs.html#5.1.4-2

local function lunpack(i: number, ...: number)
	if i == 0 then return ... end
	return lunpack(i - 1, i, ...)
end

local function lunpack2(i: number, ...: number)
	return if i == 0 then ... else lunpack(i - 1, i, ...)
end

print(lunpack(5))
print(lunpack2(5))
