local number = 1000
local sum = 0

for i = 1, number - 1 do
	if i % 3 == 0 or i % 5 == 0 then
		sum += i
	end
end

print(sum)
