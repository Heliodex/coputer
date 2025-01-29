do
	local s = "hello 世界"

	print(string.find(s, "hello"))
	print(string.find(s, "hello", -12))
	print(string.find(s, "世界", 7))
	print(string.find(s, "世界", -6))

	print(string.find(s, "hello", 3))
	print(string.find(s, "hello", -11))
	print(string.find(s, "世界", 8))
	print(string.find(s, "世界", -5))
	print(string.find(s, "sup"))
end

do
	-- Example 1: Basic usage
	local s = "Hello, world!"
	local pattern = "world"
	local start_index, end_index = string.find(s, pattern)
	print(start_index, end_index) -- Output: 8 12
end

do
	-- Example 2: Using init parameter
	local s = "Hello, world! Hello, Roblox!"
	local pattern = "Hello"
	local start_index, end_index = string.find(s, pattern, 10)
	print(start_index, end_index) -- Output: 15 19
end

do
	-- Example 3: Using plain parameter
	local s = "Hello, world! (Hello)"
	local pattern = "(Hello)"
	local start_index, end_index = string.find(s, pattern, 1, true)
	print(start_index, end_index) -- Output: 14 20 (wrong)
end

do
	-- Example 4: No Pattern found
	local s = "Hello, world!"
	local pattern = "Roblox"
	local start_index, end_index = string.find(s, pattern)
	print(start_index, end_index) -- Output: nil
end

print()

do
	local s = "hello world"
	local i, j = string.find(s, "hello")
	print(i, j) --> 1    5
	print(string.sub(s, i or 0, j or 0)) --> hello
	print(string.find(s, "world")) --> 7    11
	i, j = string.find(s, "l")
	print(i, j) --> 3    3
	print(string.find(s, "lll")) --> nil
end

print() -- it's time

local ss = "mulllllltiplllle repeating characters"
print(string.find(ss, "l+"))
print(string.find(ss, "l+", 3))
print(string.find(ss, "l+", 6))
print(string.find(ss, "l+", 8))
print(string.find(ss, "l+", 9))
print(string.find(ss, "l+", 15))
print(string.find(ss, "l+", 16))
print(string.find(ss, "c+", 16))
print(string.find(ss, "c*", 16))
print(string.find(ss, "e*"))
print(string.find(ss, "e+"))
print(string.find(ss, "^"))

print(string.find("m", "m?m"))
print(string.find("mu", "c?u"))
print(string.find(ss, "c?u", 2))
print(string.find(ss, "c?u", 3))

print(string.find("sup", "^"))
print(string.find("sup", "$"))
print(string.find("sup", "$", 4))
print(string.find("sup", "$", 5))

print(string.find("sup", "p$"))
print(string.find("sup", "^sup$"))
print(string.find("sup", "^sip$"))
print(string.find("sup", "^s.+"))
print(string.find("sup", "^u.+"))
print(string.find("sup", "u.+"))

print(string.find("Price: $123.45", "%$%d+%.%d+"))
print(string.find("Price: $123.45", "$%d+%.%d+"))
print(string.find("Price: $123.45", "%$(%d+%.%d+)"))
