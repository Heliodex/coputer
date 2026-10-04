local function add(a, b)
	return a + b
end

local function generic<T, U...>(x: T, ...: number): T
	return x
end

local anon = function(x)
	return x
end

function obj.method(a, b)
end

function obj:method2(a)
end

@native
function attributed()
end
