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
print(string.format("%q", ""))
print(string.format("%q", "n\0il b\0ytez"))
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

print(string.format("The magic word is %s", "Roblox"))
print(string.format "Roblox")
print(string.format("The magic word is %q", "Roblox"))
print(string.format("%q", "Skip to \na new line and \nanother new line!"))

print(string.format("%e", 42))
print(string.format("%E", 42))

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

-- uncomment when https://github.com/luau-lang/luau/issues/1650
-- print(string.format("%c", 0) == "\0") -- huehuehuehuehuehehehe
-- print(string.format("%c", 0) == "")