local largest = 0

for i = 3000, 1000, -1 do
	for j = 3000, 1000, -1 do
		local n = i * j
		if n > largest and tostring(n) == string.reverse(tostring(n)) then
			largest = n
		end
	end
end

print(largest)
