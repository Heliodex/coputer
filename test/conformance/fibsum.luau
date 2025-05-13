-- from 2.luau

local max = 10
local fib = { 1, 1 }
local total = 0

while fib[1] + fib[2] < max do
	table.insert(fib, 1, fib[1] + fib[2])
end

for _, v in fib do
	total += v
end

print(total)
