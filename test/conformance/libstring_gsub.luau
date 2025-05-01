print(string.gsub("alo alo", "a", "x"))

-- Simple replacement
print(string.gsub("I love tacos!", "tacos", "Roblox")) --> I love Roblox! 1
-- Using a pattern (hint: %w+ matches whole words)
print(string.gsub("I like red!", "%w+", "word")) --> word word word 3
-- Replacement table
print(string.gsub("I play Roblox.", "%w+", { I = "Je", play = "joue à" })) --> Je joue à Roblox. 3
-- Replacement function
print(string.gsub("I have 2 cats.", "%d+", function(n)
	local nn = tonumber(n)
	if nn then return tostring(nn * 12) end
	return ""
end)) --> I have 24 cats. 1
-- Replace only twice
print(string.gsub("aaa", "a", "b", 2)) --> "bba", 2

print(string.gsub("12345", "%d+", { [12345] = "67890" }))
print(string.gsub("12345", "%d+", { ["12345"] = "67890" }))

local f = [[
return function ( a , b , c , d , e )
  local x = a >= b or c or ( d and e ) or nil
  return x
end , { a = 1 , b = 2 >= 1 , } or { 1 };
]]
print(string.gsub(f, "%s+", "\n"))

-- print "long strings"
local x =
	"01234567890123456789012345678901234567890123456789012345678901234567890123456789"
print(string.len(x))
local s = ""
local n = 0

while n < 3 do
	s ..= x
	n += 1
end
print(string.len(s), #s)

local function strsin(s: string)
	local n = tonumber(s)

	-- precision kix my ass
	return if n then string.sub(tostring(math.sin(n)), 1, 10) else "0"
end

print(s)
local ss, i = string.gsub(s, "(%d%d%d%d)", strsin)
print(ss)
print(i, 20000 / 4)

print(string.gsub("�lo �lo", "�", "x"))
print(string.gsub("alo �lo  ", " +$", "")) -- trim
print(string.gsub("  alo alo  ", "^%s*(.-)%s*$", "%1")) -- double trim
print(string.gsub("alo  alo  \n 123\n ", "%s+", " "))
local t = "ab� d"
local a, b = string.gsub(t, "(.)", "%1@")
print(a, b)
print(string.gsub(t, "", "@"))
print("@" .. a, b)
a, b = string.gsub("ab�d", "(.)", "%0@", 2)
print(a == "a@b@�d" and b == 2)
print(string.gsub("alo alo", "()[al]", "%1") == "12o 56o")
print(string.gsub("abc=xyz", "(%w*)(%p)(%w+)", "%3%2%1-%0"))
print(string.gsub("abc", "%w", "%1%0"))
print(string.gsub("abc", "%w+", "%0%1"))
print(string.gsub("���", "$", "\0��"))
print(string.gsub("", "^", "r"))
print(string.gsub("", "$", "r"))
print "+"

print(string.gsub("um (dois) tres (quatro)", "(%(%w+%))", string.upper))
