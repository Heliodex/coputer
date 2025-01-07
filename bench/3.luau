local number = 600851475143

local function isPrime(n: number)
	if n % 2 == 0 then return false end
	for i = 3, math.sqrt(n), 2 do
		if n % i == 0 then return false end
	end
	return n ~= 1
end

local largestPrimeFactor = 0

for i = 1, math.sqrt(number) + 1, 2 do
	if number % i == 0 and isPrime(i) then largestPrimeFactor = i end
end

print(largestPrimeFactor)
