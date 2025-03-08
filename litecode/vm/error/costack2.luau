local ff = coroutine.create(function()
	local function f1()
		local f = "hi"
		f()
	end

	local function f2() f1() end
	local function f3() f2() end
	local function f4() f3() end

	f4()
end)

coroutine.resume(ff)
