-- check self capture: does self go into any upvalues?
print((function()
	local t = { f = 5 }
	function t:get() return self.f end
	return t:get()
end)())
