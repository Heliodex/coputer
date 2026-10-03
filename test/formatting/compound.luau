local x = 1
local t = {}
local i = "k"
x = x - 1
x = x + 2
x = x * 3
x = x / 4
x = x // 5
x = x % 6
x = x ^ 7
x = x .. "a"
x = (x) + 1
x = x + x
g = g + 1
t.k = t.k * 2
t[i] = t[i] .. "b"
t[1] = t[1] + 1
t["k"] = t["k"] + 1
x = x + (if c then a else b)
local function upvalue()
	x = x + 1
end
x = 1 - x
x = y - 1
t.a.k = t.a.k + 1
t[f()] = t[f()] + 1
g.k = g.k + 1
a, b = a + 1, b + 1
x = x == 1
x = x and 1
