local function fact(n: number)
	if n == 0 then
		return 1
	end
	return n * fact(n - 1)
end

print(fact(10))
