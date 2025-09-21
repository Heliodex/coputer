local function x(a, ...)
	print("first argument:", a)
	print("variadic arguments:", ...)
	return ...
end

print(x(1, 2, 3, 4))
