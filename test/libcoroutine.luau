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

-- local function r()
-- 	coroutine.yield(1)
-- 	coroutine.yield(2)
-- 	coroutine.yield(3)
-- end

-- local fc = coroutine.create(r)

-- print(coroutine.resume(fc))
-- print(coroutine.resume(fc))
-- print(coroutine.resume(fc))
-- print(coroutine.status(fc))
