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

ss = "sup"
print(string.find(ss, "^"))
print(string.find(ss, "$"))
print(string.find(ss, "$", 4))
print(string.find(ss, "$", 5))

print(string.find(ss, "p$"))
print(string.find(ss, "^sup$"))
print(string.find(ss, "^sip$"))
print(string.find(ss, "^s.+"))
print(string.find(ss, "^u.+"))
print(string.find(ss, "u.+"))

print(string.find("Price: $123.45", "%$%d+%.%d+"))
print(string.find("Price: $123.45", "$%d+%.%d+"))
print(string.find("Price: $123.45", "%$(%d+%.%d+)"))
print(string.find("yea$123yea", "$%d+"))
print(string.find("yea$123yea", "$(%d+)"))
print(string.find("yea$123yea", "($%d+)"))

ss = "Deadline is 30/05/1999, firm"
print(string.find(ss, "%d%d/%d%d/%d%d%d%d"))
print(string.find(ss, "(%d%d/%d%d/%d%d%d%d)"))
print(string.find(ss, "(%d%d)/(%d%d)/(%d%d%d%d)"))
print(string.find(ss, "(%d+)/(%d+)/(%d+)"))
print(string.find("hello, up-down!", "%a+"))
print(string.find("hello, up-down!", "%A+"))
print(string.find("the \0 null charact\0\0er", "%z"))
print(string.find("the \0 null charact\0\0er", "%z+", 10))
print(string.find("Full stop.", "."))
print(string.find("Full stop.", "%."))
print(string.find("10% luck 20% skill 15% contentrated power of will", "%%"))

print(string.find("Hello!", "^Hell"))
print(string.find("World, Hello!", "^Hell"))

ss = "HELLO! worldz"
print(string.find(ss, "[a-z]"))
print(string.find(ss, "[a-y]"))
print(string.find(ss, "[A-Z]"))
print(string.find(ss, "[A-M]"))
print(string.find(ss, "[a-z]+"))
print(string.find(ss, "[a-y]+"))
print(string.find(ss, "[A-Z]+"))
print(string.find(ss, "[A-M]+"))

print(string.find("1234567890abcdefghijklmnop", "%x+"))
print(string.find("abcbabcbdefedefebac", "[abc]+"))

print(string.find("-20", "[+-]?%d+"))
print(string.find("It's -20K", "([+-]?%d+)")) -- lolwut
print(string.find("It's 20K", "([+-]?%d+)"))
print(string.find("It's +K", "([+-]?%d+)"))

print(string.find("duplicateduplicate", "(%a+)%1"))
print(string.find("duplicateduplicate after", "(%a+)%1 after"))
print(string.find("dupe dupe", "(%a+) %1"))
print(string.find("dupe dupe after", "(%a+) %1 after"))

print(string.find("one |two| three |four| five", "|[^|]*|"))
print(string.find("foo: 123 bar: 456", "(%a+):%s*(%d+)%s+(%a+):%s*(%d+)"))
print(string.find("%j", "%j"))
print(string.find("%j", "%%j"))
print(string.find("12345712301-helloworld8yeah", "([0-7%l%-]+)"))

print(string.find("no meaning no", "[%a-z]+"))
print(string.find("no meaning no", "[^%a-z]+"))
print(string.find("no meaning no", "[a-%%]+"))
print(string.find("no meaning no", "[^a-%%]+"))

print(string.find("these (are (balanced)) parentheses", "%b()"))
print(string.find("these (are (balanced)) parentheses", "ese (%b()) paren"))
print(string.find("(these (aren't (balanced)) parentheses", "%b()"))
print(string.find("these (aren't (balanced)) parentheses)", "%b()"))

-- the final frointer
print(string.find("abcabcdefdef", "%f[def]"))
print(string.find("abcabcdefdef", "(%f[def])"))

print(string.find("a", "%f[a]"))
print(string.find("a", "%f[^%z]"))
print(string.find("a", "%f[^%l]"))
print(string.find("aba", "%f[a%z]"))
print(string.find("aba", "%f[%z]"))
print(string.find("aba", "%f[%l%z]"))
print(string.find("aba", "%f[^%l%z]"))

print(string.find("This is a string", "(.()(%a+))"))
print(string.find("\a\b\f\n\r\t\v", "^%c%c%c%c%c%c%c$"))

print(string.find("", "")) -- empty patterns are tricky
print(string.find("alo", ""))
print(string.find("a\0o a\0o a\0o", "a", 1)) -- first position
print(string.find("a\0o a\0o a\0o", "a\0o", 2)) -- starts in the middle
print(string.find("a\0o a\0o a\0o", "a\0o", 9)) -- starts in the middle
print(string.find("a\0a\0a\0a\0\0ab", "\0ab", 2)) -- finds at the end
print(string.find("a\0a\0a\0a\0\0ab", "b")) -- last position
print(string.find("a\0a\0a\0a\0\0ab", "b\0")) -- check ending
print(string.find("", "\0"))
print(string.find("alo123alo", "12"))
print(string.find("alo123alo", "^12"))

local function f(s, p)
	local i, e = string.find(s, p)
	if i then
		print(i, e, string.sub(s, i, e))
	else
		print(nil)
	end
end

f("aloALO", "%l*")
f("aLo_ALO", "%a*")

f("  \n\r*&\n\r   xuxu  \n\n", "%g%g%g+")

f("aaab", "a*")
f("aaa", "^.*$")
f("aaa", "b*")
f("aaa", "ab*a")
f("aba", "ab*a")
f("aaab", "a+")
f("aaa", "^.+$")
f("aaa", "b+")
f("aaa", "ab+a")
f("aba", "ab+a")
f("a$a", ".$")
f("a$a", ".%$")
f("a$a", ".$.")
f("a$a", "$$")
f("a$b", "a$")
f("a$a", "$")
f("", "b*")
f("aaa", "bb*")
f("aaab", "a-")
f("aaa", "^.-$")

f("bbb", "b.*b")
f("baaabaaab", "b.*b")
f("aabaaabaaabaaaba", "b.*b")
f("aabaaabaaabaaaba", "b.-b")

f("alo xo", ".o$")
f(" \n isto � assim", "%S%S*")
f(" \n isto � assim", "%S*$")
f(" \n isto � assim", "[a-z]*$")
f("um caracter ? extra", "[^%sa-z]")
f("", "a?")
f("�", "�?")
f("�bl", "�?b?l?")
f("  �bl", "�?b?l?")
f("aa", "^aa?a?a")
f("]]]�b", "[^]]")
f("0alo alo", "%x*")
f("alo alo", "%C+")

print(string.find("(�lo)", "%(�"))

-- big strings
local a = string.rep("a", 300000)
print(string.find(a, "^a*.?$"))
print(string.find(a, "^a*.?b$"))
print(string.find(a, "^a-.?$"))
