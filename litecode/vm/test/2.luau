local max = 4e100
local fib = { 1, 2 }
local total = 0

while fib[#fib] + fib[#fib - 1] < max do
	table.insert(fib, fib[#fib] + fib[#fib - 1])
end

for _, v in fib do
	if v % 2 == 0 then
		total += v
	end
end

print(total)
