local sum = 1000

for a = 1, sum do
	for b = a + 1, (sum - a) / 2 do
		local c = sum - a - b
		if a ^ 2 + b ^ 2 == c ^ 2 then
			print(a * b * c)
			return
		end
	end
end
