local a, b = 1, 2
print(a == 1, a ~= 1)
print(b == 1, b ~= 1)

local c, d = "hello", "world"
print(c == "hello", c ~= "hello")
print(d == "hello", d ~= "hello")

print(a == a, b ~= b)
print(c == c, d ~= d)
print(a == c, b == d) -- across types

local e, f = true, false
print(e == true, e ~= true)
print(f == true, f ~= true)
print(e ~= f, e == not f, f == not e, (not e) ~= not f)

local g, h = {}, {}
print(g == g, h == h)
print(g == h, g ~= h)
print(g == {}, h == {})

g, h = table.create(a, a), table.create(b, b)
print(g == g, h == h)
print(g == h, g ~= h)
print(g == table.create(a, a), h == table.create(b, b))

local i, j = function() end, function() end
print(i == i, j == j)
print(i == j, i ~= j)
print(i == function() end, j == function() end)
print(i() == j())

local k, l = coroutine.create(i), coroutine.create(j)
print(k == k, l == l)
print(k == l, k ~= l)
print(k == coroutine.create(i), l == coroutine.create(j))
print(coroutine.resume(k) == coroutine.resume(l))

local m, n = buffer.create(a), buffer.create(b)
print(m == m, n == n)
print(m == n, m ~= n)
print(m == buffer.create(a), n == buffer.create(b))
print(buffer.readu8(m, 0), buffer.readu8(n, 0))

local o, p = vector.create(1, 2, 3), vector.create(4, 5, 6)
print(o == o, p == p)
print(o == p, o ~= p)
print(o == vector.create(1, 2, 3), p == vector.create(4, 5, 6))
print(o.x, o.y, o.z)
print(p.x, p.y, p.z)
