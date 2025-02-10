do
	-- Example 3: Using plain parameter
	local s = "Hello, world! (Hello)"
	local pattern = "(Hello)"
	local start_index, end_index = string.find(s, pattern, 1, true)
	print(start_index, end_index) -- Output: 14 20 (wrong)
end

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

local function stringfind(s, p, i): ...any
	return if i then string.find(s, p, i) else string.find(s, p)
end
local function stringmatch(s, p, i): ...any
	return if i then string.match(s, p, i) else string.match(s, p)
end

for _, func in { stringfind, stringmatch } do
	do
		local s = "hello 世界"
		print(func(s, "hello"))
		print(func(s, "hello", -12))
		print(func(s, "世界", 7))
		print(func(s, "世界", -6))

		print(func(s, "hello", 3))
		print(func(s, "hello", -11))
		print(func(s, "世界", 8))
		print(func(s, "世界", -5))
		print(func(s, "sup"))
	end

	do
		-- Example 1: Basic usage
		local s = "Hello, world!"
		local pattern = "world"
		local start_index, end_index = func(s, pattern)
		print(start_index, end_index) -- Output: 8 12
	end

	do
		-- Example 2: Using init parameter
		local s = "Hello, world! Hello, Roblox!"
		local pattern = "Hello"
		local start_index, end_index = func(s, pattern, 10)
		print(start_index, end_index) -- Output: 15 19
	end

	do
		-- Example 4: No Pattern found
		local s = "Hello, world!"
		local pattern = "Roblox"
		local start_index, end_index = func(s, pattern)
		print(start_index, end_index) -- Output: nil
	end

	print() -- it's time

	local ss = "mulllllltiplllle repeating characters"
	print(func(ss, "l+"))
	print(func(ss, "l+", 3))
	print(func(ss, "l+", 6))
	print(func(ss, "l+", 8))
	print(func(ss, "l+", 9))
	print(func(ss, "l+", 15))
	print(func(ss, "l+", 16))
	print(func(ss, "c+", 16))
	print(func(ss, "c*", 16))
	print(func(ss, "e*"))
	print(func(ss, "e+"))
	print(func(ss, "^"))

	print(func("m", "m?m"))
	print(func("mu", "c?u"))
	print(func(ss, "c?u", 2))
	print(func(ss, "c?u", 3))

	ss = "sup"
	print(func(ss, "^"))
	print(func(ss, "$"))
	print(func(ss, "$", 4))
	print(func(ss, "$", 5))

	print(func(ss, "p$"))
	print(func(ss, "^sup$"))
	print(func(ss, "^sip$"))
	print(func(ss, "^s.+"))
	print(func(ss, "^u.+"))
	print(func(ss, "u.+"))

	print(func("Price: $123.45", "%$%d+%.%d+"))
	print(func("Price: $123.45", "$%d+%.%d+"))
	print(func("Price: $123.45", "%$(%d+%.%d+)"))
	print(func("yea$123yea", "$%d+"))
	print(func("yea$123yea", "$(%d+)"))
	print(func("yea$123yea", "($%d+)"))

	ss = "Deadline is 30/05/1999, firm"
	print(func(ss, "%d%d/%d%d/%d%d%d%d"))
	print(func(ss, "(%d%d/%d%d/%d%d%d%d)"))
	print(func(ss, "(%d%d)/(%d%d)/(%d%d%d%d)"))
	print(func(ss, "(%d+)/(%d+)/(%d+)"))
	print(func("hello, up-down!", "%a+"))
	print(func("hello, up-down!", "%A+"))
	print(func("the \0 null charact\0\0er", "%z"))
	print(func("the \0 null charact\0\0er", "%z+", 10))
	print(func("Full stop.", "."))
	print(func("Full stop.", "%."))
	print(func("10% luck 20% skill 15% contentrated power of will", "%%"))

	print(func("Hello!", "^Hell"))
	print(func("World, Hello!", "^Hell"))

	ss = "HELLO! worldz"
	print(func(ss, "[a-z]"))
	print(func(ss, "[a-y]"))
	print(func(ss, "[A-Z]"))
	print(func(ss, "[A-M]"))
	print(func(ss, "[a-z]+"))
	print(func(ss, "[a-y]+"))
	print(func(ss, "[A-Z]+"))
	print(func(ss, "[A-M]+"))

	print(func("1234567890abcdefghijklmnop", "%x+"))
	print(func("abcbabcbdefedefebac", "[abc]+"))

	print(func("-20", "[+-]?%d+"))
	print(func("It's -20K", "([+-]?%d+)")) -- lolwut
	print(func("It's 20K", "([+-]?%d+)"))
	print(func("It's +K", "([+-]?%d+)"))

	print(func("duplicateduplicate", "(%a+)%1"))
	print(func("duplicateduplicate after", "(%a+)%1 after"))
	print(func("dupe dupe", "(%a+) %1"))
	print(func("dupe dupe after", "(%a+) %1 after"))

	print(func("one |two| three |four| five", "|[^|]*|"))
	print(func("foo: 123 bar: 456", "(%a+):%s*(%d+)%s+(%a+):%s*(%d+)"))
	print(func("%j", "%j"))
	print(func("%j", "%%j"))
	print(func("12345712301-helloworld8yeah", "([0-7%l%-]+)"))

	print(func("no meaning no", "[%a-z]+"))
	print(func("no meaning no", "[^%a-z]+"))
	print(func("no meaning no", "[a-%%]+"))
	print(func("no meaning no", "[^a-%%]+"))

	print(func("these (are (balanced)) parentheses", "%b()"))
	print(func("these (are (balanced)) parentheses", "ese (%b()) paren"))
	print(func("(these (aren't (balanced)) parentheses", "%b()"))
	print(func("these (aren't (balanced)) parentheses)", "%b()"))

	-- the final frointer
	print(func("abcabcdefdef", "%f[def]"))
	print(func("abcabcdefdef", "(%f[def])"))

	print(func("a", "%f[a]"))
	print(func("a", "%f[^%z]"))
	print(func("a", "%f[^%l]"))
	print(func("aba", "%f[a%z]"))
	print(func("aba", "%f[%z]"))
	print(func("aba", "%f[%l%z]"))
	print(func("aba", "%f[^%l%z]"))

	print(func(" alo aalo allo", "%f[%S].-%f[%s].-%f[%S]"))

	-- \0 in patterns
	print(func("b$a", "$\0?"))
	print(func("abc\0efg", "%\0"))

	-- magic char after \0
	print(func("abc\0\0", "\0."))
	print(func("abcx\0\0abc\0abc", "x\0\0abc\0a."))

	print(func("This is a string", "(.()(%a+))"))
	print(func("\a\b\f\n\r\t\v", "^%c%c%c%c%c%c%c$"))

	print(func("", "")) -- empty patterns are tricky
	print(func("alo", ""))
	print(func("a\0o a\0o a\0o", "a", 1)) -- first position
	print(func("a\0o a\0o a\0o", "a\0o", 2)) -- starts in the middle
	print(func("a\0o a\0o a\0o", "a\0o", 9)) -- starts in the middle
	print(func("a\0a\0a\0a\0\0ab", "\0ab", 2)) -- finds at the end
	print(func("a\0a\0a\0a\0\0ab", "b")) -- last position
	print(func("a\0a\0a\0a\0\0ab", "b\0")) -- check ending
	print(func("", "\0"))
	print(func("alo123alo", "12"))
	print(func("alo123alo", "^12"))

	print(func("(�lo)", "%(�"))
end

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

-- big strings
local a = string.rep("a", 300000)
print(string.find(a, "^a*.?$"))
print(string.find(a, "^a*.?b$"))
print(string.find(a, "^a-.?$"))
