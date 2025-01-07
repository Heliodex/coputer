local divisors = 300

local function countDivisors(n: number)
	local count = 0
	for i = 1, n ^ 0.5 do
		if n % i == 0 then
			count += if i == n / i then 1 else 2
		end
	end
	return count
end

for i = 1, math.huge do
	local num = 1
	for j = 2, i do
		num += j
	end

	if countDivisors(num) > divisors then
		print(num)
		return
	end
end
