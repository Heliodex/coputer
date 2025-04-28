local unpack = table.unpack

local res, res2 = 27
print(res, res2)

local a: any, b: any = 1, 2 + 3
print(a == 1 and b == 5)
a = {}
local f: any
function f() return 10, 11, 12 end
a.x, b, a[1] = 1, 2, f()
print(a.x == 1 and b == 2 and a[1] == 10)
a[f()], b, a[f() + 3] = f(), a, "x"
print(a[10] == 10 and b == a and a[13] == "x")

do
	local f = function(n)
		local x = {}
		for i = 1, n do
			x[i] = i
		end
		return unpack(x)
	end
	local a, b, c
	a, b = 0, f(1)
	print(a == 0 and b == 1)
	local A
	A, b = 0, f(1)
	print(A == 0 and b == 1)
	a, b, c = 0, 5, f(4)
	print(a == 0 and b == 5 and c == 1)
	a, b, c = 0, 5, f(0)
	print(a == 0 and b == 5 and c == nil)
end

local c, d
a, b, c, d = 1 and nil, 1 or nil, (1 and (nil or 1)), 6
print(not a and b and c and d == 6)

d = 20
a, b, c, d = f()
print(a == 10 and b == 11 and c == 12 and d == nil)
a, b = f(), 1, 2, 3, f()
print(a == 10 and b == 1)

print((a < b) == false and (a > b) == true)
print((10 and 2))
print((10 or 2))
print((10 or print(nil)))
print(not (nil and print(nil)))
print((nil or "alo"))
print((nil and 10))
print((false and 10))
print((true or 10))
print((false or 10))
print(false ~= nil)
print(nil ~= false)
print(not nil)
print(not not nil)
print(not not 1)
print(not not a)
print(not not (6 or nil))
print(not not (nil and 56))
print(not not (nil and true))
print "+"

a = {}
a[true] = 20
a[false] = 10
print(a[1 < 2] == 20 and a[1 > 2] == 10)

function f(a) return a end

local a: any = {}
for i = 3000, -3000, -1 do
	a[i] = i
end
a[10e30] = "alo"
a[true] = 10
a[false] = 20
print(a[10e30] == "alo" and a[not 1] == 20 and a[10 < 20] == 10)
for i = 3000, -3000, -1 do
	if a[i] ~= i then print(i, a[i]) end
end
a[print] = print
a[f] = print
a[a] = a
print(a[a][a][a][a][print] == print)
a[print](a[a[f]] == a[print])
a = nil

a = { 10, 9, 8, 7, 6, 5, 4, 3, 2, [-3] = "a", [f] = print, a = "a", b = "ab" }
a, a.x, a.y = a, a[-3]
print(a[1] == 10 and a[-3] == a.a and a[f] == print and a.x == "a" and not a.y)
a[1], f(a)[2], b, c = { ["alo"] = print }, 10, a[1], a[f], 6, 10, 23, f(a), 2
a[1].alo(a[2] == 10 and b == 10 and c == print)

a[2 ^ 31] = 10
a[2 ^ 31 + 1] = 11
a[-2 ^ 31] = 12
a[2 ^ 32] = 13
a[-2 ^ 32] = 14
a[2 ^ 32 + 1] = 15
a[10 ^ 33] = 16

print(
	a[2 ^ 31],
	a[2 ^ 31 + 1],
	a[-2 ^ 31],
	a[2 ^ 32],
	a[-2 ^ 32],
	a[2 ^ 32 + 1],
	a[10 ^ 33]
)

a = nil

do
	local a, i, j, b
	a = { "a", "b" }
	i = 1
	j = 2
	b = a
	i, a[i], a, j, a[j], a[i + j] = j, i, i, b, j, i
	print(i, b[1], a, j == b, b[2], b[3])
end
