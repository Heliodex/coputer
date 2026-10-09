if x then
	print(1)
elseif y then
	print(2)
else
	print(3)
end

while x do
	break
end

repeat
	continue
until x

for i = 1, 10, 2 do
	print(i)
end

for k, v in pairs(t) do
	print(k, v)
end

do
	local x = 1
end
