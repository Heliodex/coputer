-- this file is the pits

local f = nil
function f(x) return a:x(x) end
print(type(f))

local pack = table.pack
local unpack = table.unpack

-- testing local-function recursion
local fact = false
do
	local res = 1
	local function fact(n: number)
		if n == 0 then
			return res
		else
			return n * fact(n - 1)
		end
	end
	print(fact(5) == 120)
end
print(fact == false)

-- testing declarations
local a = { i = 10 }
local self = 20
function a:x(x: number) return x + self.i end
function a.y(x: number) return x + self end

print(a:x(1) + 10 == a.y(1))

a.t = { i = -100 }
a["t"].x = function(self, a, b) return self.i + a + b end

print(a.t:x(2, 3) == -95)

do
	local a = { x = 0 }
	function a:add(x)
		self.x, a.y = self.x + x, 20
		return self
	end
	print(a:add(10):add(20):add(30).x == 60 and a.y == 20)
end

local a = { b = { c = {} } }

function a.b.c.f1(x: number) return x + 1 end
function a.b.c:f2(x, y) self[x] = y end
print(a.b.c.f1(4) == 5)
a.b.c:f2("k", 12)
print(a.b.c.k == 12)

print "+"

local t = nil -- 'declare' t
function f(a, b, c)
	local d = "a"
	t = { a, b, c, d }
end

f( -- this line change must be valid
	1,
	2
)
print(t[1], t[2], t[3], t[4])
f(
	1,
	2, -- this one too
	3,
	4
)
print(t[1], t[2], t[3], t[4])

local dummy
do
	function dummy(n)
		if n > 0 then dummy(n - 1) end
	end
end

dummy(10)

local function deep(n)
	if n > 0 then deep(n - 1) end
end
deep(10)
deep(200)

-- testing tail call
function deep(n)
	if n > 0 then
		return deep(n - 1)
	else
		return 101
	end
end
print(deep(10000))
a = {}
function a:deep(n)
	if n > 0 then
		return self:deep(n - 1)
	else
		return 101
	end
end
print(a:deep(10000))

print "+"

a = nil
(function(x) a = x end)(23)
print(a, (function(x) return x * 2 end)(20))

local x: any, y, z, a: any
a = {}
local lim = 2000
for i = 1, lim do
	a[i] = i
end
-- print(select(lim, unpack(a)) == lim and select("#", unpack(a)) == lim)
x = unpack(a)
print(x == 1)
x = { unpack(a) }
print(#x, x[1], x[lim])
x = { unpack(a, lim - 2) }
print(#x, x[1], x[3])
x = { unpack(a, 10, 6) }
print(next(x)) -- no elements
x = { unpack(a, 11, 10) }
print(next(x)) -- no elements
x, y = unpack(a, 10, 10)
print(x, y)
x, y, z = unpack(a, 10, 11)
print(x, y, z)
a, x = unpack { 1 }
print(a, x)
a, x = unpack({ 1, 2 }, 1, 1)
print(a, x)

-- testing closures

-- fixed-point operator
local Y = function(le)
	local function a(f)
		return le(function(x) return f(f)(x) end)
	end
	return a(a)
end

-- non-recursive factorial

local F = function(f)
	return function(n)
		if n == 0 then
			return 1
		else
			return n * f(n - 1)
		end
	end
end

local fat = Y(F)

print(fat(0), fat(4), Y(F)(5), 5 * Y(F)(4))

local function g(z)
	local function f(a, b, c, d)
		return function(x, y) return a + b + c + d + a + x + y + z end
	end
	return f(z, z + 1, z + 2, z + 3)
end

f = g(10)
print(f(9, 16))

Y, F, f = nil
print "+"

-- testing multiple returns

local function unlpack(t, i: any?): (any, any)
	i = i or 1
	if i <= #t then return t[i], unlpack(t, i + 1) end
end

local function equaltab(t1, t2)
	print(#t1 == #t2)
	for i, v1 in ipairs(t1) do
		print(v1 == t2[i])
	end
end

function f() return 1, 2, 30, 4 end
local function ret2(a, b) return a, b end

local a, b, c, d = unlpack { 1, 2, 3 }
print(a, b, c, d)
a = { 1, 2, 3, 4, false, 10, "alo", false }
equaltab(pack(unlpack(a)), a)
equaltab(pack(unlpack(a), -1), { 1, -1 })
a, b, c, d = ret2(f()), ret2(f())
print(a, b, c, d)
a, b, c, d = unlpack(pack(ret2(f()), ret2(f())))
print(a, b, c, d)
a, b, c, d = unlpack(pack(ret2(f()), (ret2(f()))))
print(a, b, c, d)

a = ret2 { unlpack { 1, 2, 3 }, unlpack { 3, 2, 1 }, unlpack { "a", "b" } }
print(a[1], a[2], a[3], a[4])

-- testing calls with 'incorrect' arguments
print(math.sin(1, 2) == math.sin(1))
table.sort(
	{ 10, 9, 8, 4, 19, 23, 0, 0 },
	function(a, b) return a < b end,
	"extra arg"
)

-- test for bug in parameter adjustment
print((function() return nil end)(4))
print((function()
	local a
	return a
end)(4))
print((function(a) return a end)())
