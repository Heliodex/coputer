local s = "hello 世界"

print(string.byte(s, 5, 7))
print(string.byte(s, 9, 99))
print(string.byte(s, -5, 5))
print(string.byte(s, -6, 6))
print(string.byte(s, -7, 7))

print(string.byte(s, 3, -6))
print(string.byte(s, 4, -6))
print(string.byte(s, 5, -6))
print(string.byte(s, 6, -6))
print(string.byte(s, 7, -7))

print(string.char())
print(string.char(104, 101, 108, 108, 111, 32, 228, 184, 150, 231, 149, 140))

print(string.find(s, "hello"))
print(string.find(s, "hello", -12))
print(string.find(s, "世界", 7))
print(string.find(s, "世界", -6))

print(string.find(s, "hello", 3))
print(string.find(s, "hello", -11))
print(string.find(s, "世界", 8))
print(string.find(s, "世界", -5))
print(string.find(s, "sup"))

print(string.format("an asterisk: %c", 42))
print(string.format("%d", 42))
print(string.format("%d %i", 42, 42))
print(string.format("%d %i", -42, -42))
print(string.format("the answer is %d%%!", 42))
print(string.format("%f", 42.5))
print(string.format("%f", -42.5))
print(string.format("%o", 25))
print(string.format("%x %X", 255, 255))
print(string.format("%o", -25))
print(string.format("%x %X", -255, -255)) -- LOL
print(string.format("%s", "whatsup"))
print(string.format("%q", "whatsup"))
print(string.format("%q", 'print "hello"'))
print(string.format("%q", [[
	print "hello"
]]))
print(string.format("%u", 42))
print(string.format("%u", -42))
print(string.format("%*", "anythingyyy"))
print(string.format("%*", 500))
print(string.format("%*", 500.5))

-- print(string.format("%e", 42))
-- print(string.format("%E", 42))

-- docs things
local str
str = "%-10d"
print(string.format(str, 300) .. "]")
str = "%-10i"
print(string.format(str, 300) .. "]")
-- 300       ]
-- There are 7 spaces between '300' and ']'

str = "%+i versus %+i"
print(string.format(str, 300, -300)) -- +300 versus -300

str = "There is a% i%% chance of rain in Seattle today."
print(string.format(str, 100))
-- -- There is a 100% chance of rain in Seattle today.

str = "%012i"
print("Score: " .. string.format(str, 15000))
-- Output: Score: 000000015000
-- -- The output has 12 digits total, left-padded with zeros

-- -- Add decimal with precision of 2 for a currency output
str = "$%.2f"
print(string.format(str, 300)) -- Output: $300.00

-- -- Return first 6 letters of a string
str = "%.6s"
print(string.format(str, "Robloxian")) -- Output: Roblox

str = "Once upon a time, there was a dragon named %s and it had %.8f horns."
print(string.format(str, "Pi", math.pi))
-- -- Output: Once upon a time, there was a dragon named Pi and it had 3.14159265 horns.
