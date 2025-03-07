local f1 = coroutine.create(function()
	local f = "hi"
	f()
end)

local f2 = coroutine.create(function() coroutine.resume(f1) end)
local f3 = coroutine.create(function() coroutine.resume(f2) end)
local f4 = coroutine.create(function() coroutine.resume(f3) end)

coroutine.resume(f4)
