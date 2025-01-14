local function isPrime(n: number)
	if n % 2 == 0 then return false end
	for i = 3, math.sqrt(n), 2 do
		if n % i == 0 then return false end
	end
	return n ~= 1
end

local p = 1
for i = 1, math.huge, 2 do
	if isPrime(i) then
		p += 1
		if p == 10001 then
			print(i)
			break
		end
	end
end
