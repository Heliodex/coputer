local function f()
	print "f"
	print(coroutine.isyieldable())
	print(coroutine.status(coroutine.running()))
end

local fc = coroutine.create(f)
print("resuming...", coroutine.status(fc))
print(coroutine.resume(fc))
print("finished!", coroutine.status(fc))
print(coroutine.resume(fc))
print("finished!", coroutine.status(fc))

print(coroutine.isyieldable())

local function r()
	coroutine.yield(1)
	coroutine.yield(2)
	coroutine.yield(3)
end

local rc = coroutine.create(r)

for i = 1, 5 do
	print(coroutine.resume(rc))
end
print(coroutine.status(rc))

print "hi"

local function s(a: number)
	a = coroutine.yield(a * 2)
	a = coroutine.yield(a * 2)
	a = coroutine.yield(a * 2)
end

local sc = coroutine.create(s)

for i = 1, 5 do
	print(coroutine.resume(sc, i))
end
print(coroutine.status(sc))

local w = coroutine.wrap(s)

for i = 1, 4 do
	print(w(i))
end

-- https://www.lua.org/bugs.html#5.2.2-8
print(coroutine.resume(coroutine.running()))

print(coroutine.resume(coroutine.create(function()
	coroutine.resume(coroutine.running())
	coroutine.yield()
end)))

-- wow
print(
	coroutine.resume(
		coroutine.create(
			function() coroutine.yield(coroutine.resume(coroutine.running())) end
		)
	)
)

local cc
cc = coroutine.create(function() coroutine.yield(coroutine.resume(cc)) end)
print(coroutine.resume(cc))
