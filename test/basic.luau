-- bruh

local function concat(head, ...)
    -- if select('#', ...) == 0 then
    --     return tostring(head)
    -- else
    --     return tostring(head) .. "," .. concat(...)
    -- end
	local s = tostring(head)
	for i, v in {...} do
		s ..= "," .. tostring(v)
	end
	return s
end

local function printr(b)
	print(b)
end

-- constants
printr(tostring(1) == "1")
printr(tostring(-1) == "-1")
printr(tostring(1.125) == "1.125")
printr(tostring(true) == "true")
printr(tostring(nil) == "nil")

-- locals
printr((function() local a = 1 return a end)() == 1)
printr((function() local a, b, c = 1, 2, 3 return c end)() == 3)
printr((function() local a, b, c = 1, 2 return c end)() == nil)
printr((function() local a = 1, 2 return a end)() == 1)

-- function calls
local function foo(a, b) return b end
printr(foo(1) == nil)
printr(foo(1, 2) == 2)
printr(foo(1, 2, 3) == 2)

-- assignments
printr((function() local a = 1 a = 2 return a end)() == 2)
printr((function() local b local a = 1 a, b = 1 return a end)() == 1)
printr((function() local b local a = 1 a, b = 1 return b end)() == nil)
printr((function() local b local a = 1 b = 2 a, b = b, a return a end)() == 2)
printr((function() local b local a = 1 b = 2 a, b = b, a return b end)() == 1)
printr((function() local a = 1 (function () a = 2 end)() return a end)() == 2)

-- assignments with local conflicts
printr((function() local a, b = 1, {} a, b[a] = 43, -1 return a + b[1] end)() == 42)
printr((function() local a = {} local b = a a[1], a = 43, -1 return a + b[1] end)() == 42)
printr((function() local a, b = 1, {} a, b[a] = (function() return 43, -1 end)() return a + b[1] end)() == 42)
printr((function() local a = {} local b = a a[1], a = (function() return 43, -1 end)() return a + b[1] end)() == 42)

-- upvalues
printr((function() local a = 1 function foo() return a end return foo() end)() == 1)

-- check upvalue propagation - foo must have numupvalues=1
printr((function() local a = 1 function foo() return function() return a end end return foo()() end)() == 1)

-- check that function args are properly closed over
printr((function() function foo(a) return function () return a end end return foo(1)() end)() == 1)

-- this checks local aliasing - b & a should share the same local slot, but the capture must return 1 instead of 2
printr((function() function foo() local f do local a = 1 f = function () return a end end local b = 2 return f end return foo()() end)() == 1)

-- this checks local mutability - we capture a ref to 1 but must return 2
printr((function() function foo() local a = 1 local function f() return a end a = 2 return f end return foo()() end)() == 2)

-- this checks upval mutability - we change the value from a context where it's upval
printr((function() function foo() local a = 1 (function () a = 2 end)() return a end return foo() end)() == 2)

-- check self capture: does self go into any upvalues?
printr((function() local t = {f=5} function t:get() return (function() return self.f end)() end return t:get() end)() == 5)

-- check self capture & close: is self copied to upval?
printr((function() function foo() local t = {f=5} function t:get() return function() return self.f end end return t:get() end return foo()() end)() == 5)

-- if
printr((function() local a = 1 if a then a = 2 end return a end)() == 2)
printr((function() local a if a then a = 2 end return a end)() == nil)

printr((function() local a = 0 if a then a = 1 else a = 2 end return a end)() == 1)
printr((function() local a if a then a = 1 else a = 2 end return a end)() == 2)

-- binary ops
printr((function() local a = 1 a = a + 2 return a end)() == 3)
printr((function() local a = 1 a = a - 2 return a end)() == -1)
printr((function() local a = 1 a = a * 2 return a end)() == 2)
printr((function() local a = 1 a = a / 2 return a end)() == 0.5)

-- binary ops with fp specials, neg zero, large constants
-- argument is passed into anonymous function to prevent constant folding
printr((function(a) return tostring(a + 0) end)(-0) == "0")
printr((function(a) return tostring(a - 0) end)(-0) == "-0")
printr((function(a) return tostring(0 - a) end)(0) == "0")
printr((function(a) return tostring(a - a) end)(1 / 0) == "nan")
printr((function(a) return tostring(a * 0) end)(0 / 0) == "nan")
printr((function(a) return tostring(a / (2^1000)) end)(2^1000) == "1")
printr((function(a) return tostring(a / (2^-1000)) end)(2^-1000) == "1")

-- floor division should always round towards -Infinity
printr((function() local a = 1 a = a // 2 return a end)() == 0)
printr((function() local a = 3 a = a // 2 return a end)() == 1)
printr((function() local a = 3.5 a = a // 2 return a end)() == 1)
printr((function() local a = -1 a = a // 2 return a end)() == -1)
printr((function() local a = -3 a = a // 2 return a end)() == -2)
printr((function() local a = -3.5 a = a // 2 return a end)() == -2)

printr((function() local a = 5 a = a % 2 return a end)() == 1)
printr((function() local a = 3 a = a ^ 2 return a end)() == 9)
printr((function() local a = 3 a = a ^ 3 return a end)() == 27)
printr((function() local a = 9 a = a ^ 0.5 return a end)() == 3)
printr((function() local a = -2 a = a ^ 2 return a end)() == 4)
printr((function() local a = -2 a = a ^ 0.5 return tostring(a) end)() == "nan")

printr((function() local a = '1' a = a .. '2' return a end)() == "12")
printr((function() local a = '1' a = a .. '2' .. '3' return a end)() == "123")

printr((function() local a = 1 a = a == 2 return a end)() == false)
printr((function() local a = 1 a = a ~= 2 return a end)() == true)
printr((function() local a = 1 a = a < 2 return a end)() == true)
printr((function() local a = 1 a = a <= 2 return a end)() == true)
printr((function() local a = 1 a = a > 2 return a end)() == false)
printr((function() local a = 1 a = a >= 2 return a end)() == false)

printr((function() local a = 1 a = a and 2 return a end)() == 2)
printr((function() local a = nil a = a and 2 return a end)() == nil)
printr((function() local a = 1 a = a or 2 return a end)() == 1)
printr((function() local a = nil a = a or 2 return a end)() == 2)

printr((function() local a a = 1 local b = 2 b = a and b return b end)() == 2)
printr((function() local a a = nil local b = 2 b = a and b return b end)() == nil)
printr((function() local a a = 1 local b = 2 b = a or b return b end)() == 1)
printr((function() local a a = nil local b = 2 b = a or b return b end)() == 2)

printr((function(a) return 12 % a end)(5) == 2)

-- unary ops
printr((function() local a = true a = not a return a end)() == false)
printr((function() local a = false a = not a return a end)() == true)

printr((function() return #{1,2} end)() == 2)
printr((function() return #'g' end)() == 1)

printr((function() local a = 1 a = -a return a end)() == -1)

-- while/repeat
printr((function() local a = 10 local b = 1 while a > 1 do b = b * 2 a = a - 1 end return b end)() == 512)
printr((function() local a = 10 local b = 1 repeat b = b * 2 a = a - 1 until a == 1 return b end)() == 512)

printr((function() local a = 10 local b = 1 while true do b = b * 2 a = a - 1 if a == 1 then break end end return b end)() == 512)
printr((function() local a = 10 local b = 1 while true do b = b * 2 a = a - 1 if a == 1 then break else end end return b end)() == 512)
printr((function() local a = 10 local b = 1 repeat b = b * 2 a = a - 1 if a == 1 then break end until false return b end)() == 512)
printr((function() local a = 10 local b = 1 repeat b = b * 2 a = a - 1 if a == 1 then break else end until false return b end)() == 512)

-- this makes sure a - 4 doesn't clobber a (which would happen if the lifetime of locals inside the repeat..until block is contained within
-- the block and ends before the condition is evaluated
printr((function() repeat local a = 5 until a - 4 < 0 or a - 4 >= 0 end)() == nil)

-- numeric for
-- basic tests with positive/negative step sizes
printr((function() local a = 1 for b=1,9 do a = a * 2 end return a end)() == 512)
printr((function() local a = 1 for b=1,9,2 do a = a * 2 end return a end)() == 32)
printr((function() local a = 1 for b=1,9,-2 do a = a * 2 end return a end)() == 1)
printr((function() local a = 1 for b=9,1,-2 do a = a * 2 end return a end)() == 32)

-- make sure break works
printr((function() local a = 1 for b=1,9 do a = a * 2 if a == 128 then break end end return a end)() == 128)
printr((function() local a = 1 for b=1,9 do a = a * 2 if a == 128 then break else end end return a end)() == 128)

-- make sure internal index is protected against modification
printr((function() local a = 1 for b=9,1,-2 do a = a * 2 b = nil end return a end)() == 32)

-- make sure that when step is 0, we treat it as backward iteration (and as such, iterate zero times or indefinitely)
-- this is consistent with Lua 5.1; future Lua versions emit an error when step is 0; LuaJIT instead treats 0 as forward iteration
-- we repeat tests twice, with and without constant folding
local zero = tonumber("0")
printr((function() local c = 0 for i=1,10,0 do c += 1 if c > 10 then break end end return c end)() == 0)
printr((function() local c = 0 for i=10,1,0 do c += 1 if c > 10 then break end end return c end)() == 11)
printr((function() local c = 0 for i=1,10,zero do c += 1 if c > 10 then break end end return c end)() == 0)
printr((function() local c = 0 for i=10,1,zero do c += 1 if c > 10 then break end end return c end)() == 11)

--[[
	-- make sure that when limit is nan, we iterate zero times (this is consistent with Lua 5.1; future Lua versions break this)
	-- we repeat tests twice, with and without constant folding
	local nan = tonumber("nan")
	printr((function() local c = 0 for i=1,0/0 do c += 1 end return c end)() == 0)
	printr((function() local c = 0 for i=1,0/0,-1 do c += 1 end return c end)() == 0)
	printr((function() local c = 0 for i=1,nan do c += 1 end return c end)() == 0)
	printr((function() local c = 0 for i=1,nan,-1 do c += 1 end return c end)() == 0)

	-- make sure that when step is nan, we treat it as backward iteration and as such iterate once iff start<=limit
	printr((function() local c = 0 for i=1,10,0/0 do c += 1 end return c end)() == 0)
	printr((function() local c = 0 for i=10,1,0/0 do c += 1 end return c end)() == 1)
	printr((function() local c = 0 for i=1,10,nan do c += 1 end return c end)() == 0)
	printr((function() local c = 0 for i=10,1,nan do c += 1 end return c end)() == 1)
]]


-- make sure that when index becomes nan mid-iteration, we correctly exit the loop (this is broken in Lua 5.1; future Lua versions fix this)
printr((function() local c = 0 for i=-math.huge,0,math.huge do c += 1 end return c end)() == 1)
printr((function() local c = 0 for i=math.huge,math.huge,-math.huge do c += 1 end return c end)() == 1)

-- generic for
-- ipairs
printr((function() local a = '' for k in ipairs({"5", "6", "7"}) do a = a .. tostring(k) end return a end)() == "123")
printr((function() local a = '' for k,v in ipairs({"5", "6", "7"}) do a = a .. tostring(k) end return a end)() == "123")
printr((function() local a = '' for k,v in ipairs({"5", "6", "7"}) do a = a .. v end return a end)() == "567")

-- ipairs with gaps
printr((function() local a = '' for k in ipairs({"5", "6", "7", nil, "8"}) do a = a .. tostring(k) end return a end)() == "123")
printr((function() local a = '' for k,v in ipairs({"5", "6", "7", nil, "8"}) do a = a .. tostring(k) end return a end)() == "123")
printr((function() local a = '' for k,v in ipairs({"5", "6", "7", nil, "8"}) do a = a .. v end return a end)() == "567")

-- manual ipairs/inext
local inext = ipairs({"5","6","7"})
printr(concat(inext({"5","6","7"}, 2)) == "3,7")

-- pairs on array
printr((function() local a = '' for k in pairs({"5", "6", "7"}) do a = a .. tostring(k) end return a end)() == "123")
printr((function() local a = '' for k,v in pairs({"5", "6", "7"}) do a = a .. tostring(k) end return a end)() == "123")
printr((function() local a = '' for k,v in pairs({"5", "6", "7"}) do a = a .. v end return a end)() == "567")

-- pairs on array with gaps
printr((function() local a = '' for k in pairs({"5", "6", "7", nil, "8"}) do a = a .. tostring(k) end return a end)() == "1235")
printr((function() local a = '' for k,v in pairs({"5", "6", "7", nil, "8"}) do a = a .. tostring(k) end return a end)() == "1235")
printr((function() local a = '' for k,v in pairs({"5", "6", "7", nil, "8"}) do a = a .. v end return a end)() == "5678")

-- pairs on table
printr((function() local a = {} for k in pairs({a=1, b=2, c=3}) do a[k] = 1 end return a.a + a.b + a.c end)() == 3)
printr((function() local a = {} for k,v in pairs({a=1, b=2, c=3}) do a[k] = 1 end return a.a + a.b + a.c end)() == 3)
printr((function() local a = {} for k,v in pairs({a=1, b=2, c=3}) do a[k] = v end return a.a + a.b + a.c end)() == 6)

-- pairs manually
printr((function() local a = '' for k in next,{"5", "6", "7"} do a = a .. tostring(k) end return a end)() == "123")
printr((function() local a = '' for k,v in next,{"5", "6", "7"} do a = a .. tostring(k) end return a end)() == "123")
printr((function() local a = '' for k,v in next,{"5", "6", "7"} do a = a .. v end return a end)() == "567")
printr((function() local a = {} for k in next,{a=1, b=2, c=3} do a[k] = 1 end return a.a + a.b + a.c end)() == 3)
printr((function() local a = {} for k,v in next,{a=1, b=2, c=3} do a[k] = 1 end return a.a + a.b + a.c end)() == 3)
printr((function() local a = {} for k,v in next,{a=1, b=2, c=3} do a[k] = v end return a.a + a.b + a.c end)() == 6)

-- too many vars
printr((function() local a = '' for k,v,p in pairs({a=1, b=2, c=3}) do a = a .. tostring(p) end return a end)() == "nilnilnil")

-- make sure break works
printr((function() local a = 1 for _ in pairs({1,2,3}) do a = a * 2 if a == 4 then break end end return a end)() == 4)
printr((function() local a = 1 for _ in pairs({1,2,3}) do a = a * 2 if a == 4 then break else end end return a end)() == 4)

-- make sure internal index is protected against modification
printr((function() local a = 1 for b in pairs({1,2,3}) do a = a * 2 b = nil end return a end)() == 8)

-- make sure custom iterators work! example is from PIL 7.1
local function list_iter(t)
    local i = 0
    local n = #(t)
    return function()
        i = i + 1
        if i <= n then return t[i] end
    end
end

printr((function() local a = '' for e in list_iter({4,2,1}) do a = a .. tostring(e) end return a end)() == "421")

-- make sure multret works in context of pairs() - this is a very painful to handle combination due to complex internal details
printr((function() local function f() return {"5","6","7"},8,9,0 end local a = '' for k,v in ipairs(f()) do a = a .. v end return a end)() == "567")

-- table literals
-- basic tests
printr((function() local t = {} return #t end)() == 0)

printr((function() local t = {1, 2} return #t end)() == 2)
printr((function() local t = {1, 2} return t[1] + t[2] end)() == 3)

printr((function() local t = {data = 4} return t.data end)() == 4)
printr((function() local t = {[1+2] = 4} return t[3] end)() == 4)

printr((function() local t = {data = 4, [1+2] = 5} return t.data + t[3] end)() == 9)

printr((function() local t = {[1] = 1, [2] = 2} return t[1] + t[2] end)() == 3)

-- table literals may contain duplicate fields; the language doesn't specify assignment order but we currently assign left to right
printr((function() local t = {data = 4, data = nil, data = 42} return t.data end)() == 42)
printr((function() local t = {data = 4, data = nil, data = 42, data = nil} return t.data end)() == nil)

-- multiple returns
-- local=
printr((function() function foo() return "2", "3", "4" end local a, b, c = foo() return ''..a..b..c end)() == "234")
printr((function() function foo() return "2", "3", "4" end local a, b, c = "1", foo() return ''..a..b..c end)() == "123")
printr((function() function foo() return "2" end local a, b, c = "1", foo() return ''..a..b..tostring(c) end)() == "12nil")

-- assignments
printr((function() function foo() return "2", "3" end local a, b, c, d = "1", foo() return ''..a..b..c..tostring(d) end)() == "123nil")
printr((function() function foo() return "2", "3" end local a, b, c, d a, b, c, d = "1", foo() return ''..a..b..c..tostring(d) end)() == "123nil")

-- varargs
-- local=
printr((function() function foo(...) local a, b, c = ... return a + b + c end return foo(1, 2, 3) end)() == 6)
printr((function() function foo(x, ...) local a, b, c = ... return a + b + c end return foo(1, 2, 3, 4) end)() == 9)

-- assignments
printr((function() function foo(...) local a, b, c = ... return a + b + c end return foo(1, 2, 3) end)() == 6)
printr((function() function foo(x, ...) local a, b, c = ... return a + b + c end return foo(1, 2, 3, 4) end)() == 9)

-- extra nils
printr((function() function foo(...) local a, b, c = ... return tostring(a) .. tostring(b) .. tostring(c) end return foo(1, 2) end)() == "12nil")

-- varargs + multiple returns
-- return
printr((function() function foo(...) return ... end return concat(foo(1, 2, 3)) end)() == "1,2,3")
printr((function() function foo(...) return ... end return foo() end)() == nil)
printr((function() function foo(a, ...) return a + 10, ... end return concat(foo(1, 2, 3)) end)() == "11,2,3")

-- call
printr((function() function foo(...) return ... end local function bar(...) return foo(...) end return concat(bar(1, 2, 3)) end)() == "1,2,3")
printr((function() function foo(...) return ... end local function bar(...) return foo(...) end return bar() end)() == nil)
printr((function() function foo(a, ...) return a + 10, ... end local function bar(a, ...) return foo(a * 2, ...) end return concat(bar(1, 2, 3)) end)() == "12,2,3")

-- multret + table literals
-- basic tests
printr((function() function foo(...) return { ... } end return #(foo()) end)() == 0)
printr((function() function foo(...) return { ... } end return #(foo(1, 2, 3)) end)() == 3)
printr((function() function foo() return 1, 2, 3 end return #({foo()}) end)() == 3)

-- table access
printr((function() local t = {6, 9, 7} return t[2] end)() == 9)
printr((function() local t = {6, 9, 7} return t[0] end)() == nil)
printr((function() local t = {6, 9, 7} return t[4] end)() == nil)
printr((function() local t = {6, 9, 7} return t[4.5] end)() == nil)
printr((function() local t = {6, 9, 7, [4.5]=11} return t[4.5] end)() == 11)
printr((function() local t = {6, 9, 7, a=11} return t['a'] end)() == 11)

printr((function() local t = {6, 9, 7} t[2] = 10 return t[2] end)() == 10)
printr((function() local t = {6, 9, 7} t[0] = 5 return t[0] end)() == 5)
printr((function() local t = {6, 9, 7} t[4] = 10 return t[4] end)() == 10)
printr((function() local t = {6, 9, 7} t[4.5] = 10 return t[4.5] end)() == 10)
printr((function() local t = {6, 9, 7} t['a'] = 11 return t['a'] end)() == 11)

-- and/or
-- rhs is a constant
printr((function() local a = 1 a = a and 2 return a end)() == 2)
printr((function() local a = nil a = a and 2 return a end)() == nil)
printr((function() local a = 1 a = a or 2 return a end)() == 1)
printr((function() local a = nil a = a or 2 return a end)() == 2)

-- rhs is a local
printr((function() local a = 1 local b = 2 a = a and b return a end)() == 2)
printr((function() local a = nil local b = 2 a = a and b return a end)() == nil)
printr((function() local a = 1 local b = 2 a = a or b return a end)() == 1)
printr((function() local a = nil local b = 2 a = a or b return a end)() == 2)

-- table access:  fast-path tests for array lookup
-- in-bounds array lookup shouldn't call into Lua, but if the element isn't there we'll still call the metatable
printr((function() local a = {9, [1.5] = 7} return a[1], a[2], a[1.5] end)() == 9,nil,7)

-- comparison
-- basic types
printr((function()  local a = nil return concat(a == nil, a ~= nil) end)() == "true,false")
printr((function()  local a = nil return concat(a == 1, a ~= 1) end)() == "false,true")
printr((function()  local a = 1 return concat(a == 1, a ~= 1) end)() == "true,false")
printr((function()  local a = 1 return concat(a == 2, a ~= 2) end)() == "false,true")
printr((function()  local a = true return concat(a == true, a ~= true) end)() == "true,false")
printr((function()  local a = true return concat(a == false, a ~= false) end)() == "false,true")
printr((function()  local a = 'a' return concat(a == 'a', a ~= 'a') end)() == "true,false")
printr((function()  local a = 'a' return concat(a == 'b', a ~= 'b') end)() == "false,true")

-- tables, reference equality (no mt)
-- printr((function() local a = {} return concat(a == a, a ~= a) end)() == "true,false")
-- printr((function() local a = {} b = {} return concat(a == b, a ~= b) end)() == "false,true")

-- string comparison
printr((function() local function cmp(a,b) return a<b,a<=b,a>b,a>=b end return concat(cmp('a', 'b')) end)() == "true,true,false,false")
printr((function() local function cmp(a,b) return a<b,a<=b,a>b,a>=b end return concat(cmp('a', 'a')) end)() == "false,true,false,true")
printr((function() local function cmp(a,b) return a<b,a<=b,a>b,a>=b end return concat(cmp('a', '')) end)() == "false,false,true,true")
printr((function() local function cmp(a,b) return a<b,a<=b,a>b,a>=b end return concat(cmp('', '\\0')) end)() == "true,true,false,false")
printr((function() local function cmp(a,b) return a<b,a<=b,a>b,a>=b end return concat(cmp('abc', 'abd')) end)() == "true,true,false,false")
printr((function() local function cmp(a,b) return a<b,a<=b,a>b,a>=b end return concat(cmp('ab\\0c', 'ab\\0d')) end)() == "true,true,false,false")
printr((function() local function cmp(a,b) return a<b,a<=b,a>b,a>=b end return concat(cmp('ab\\0c', 'ab\\0')) end)() == "false,false,true,true")
printr((function() local function cmp(a,b) return a<b,a<=b,a>b,a>=b end return concat(cmp('\\0a', '\\0b')) end)() == "true,true,false,false")
printr((function() local function cmp(a,b) return a<b,a<=b,a>b,a>=b end return concat(cmp('a', 'a\\0')) end)() == "true,true,false,false")
printr((function() local function cmp(a,b) return a<b,a<=b,a>b,a>=b end return concat(cmp('a', '\200')) end)() == "true,true,false,false")

-- array access
printr((function() local a = {4,5,6} return a[3] end)() == 6)
printr((function() local a = {4,5,nil,6} return a[3] end)() == nil)
printr((function() local a = {4,5,6} a[3] = 8 return a[3] end)() == 8)
printr((function() local a = {4,5,nil,6} a[3] = 8 return a[3] end)() == 8)

-- array index for literal
printr((function() local a = {4, 5, nil, 6} return concat(a[1], a[3], a[4], a[100]) end)())
printr((function() local a = {4, 5, nil, 6} a[1] = 42 a[3] = 0 a[100] = 75 return concat(a[1], a[3], a[75], a[100]) end)())


-- many arguments & locals
local function f(p1, p2, p3, p4, p5, p6, p7, p8, p9, p10,
    p11, p12, p13, p14, p15, p16, p17, p18, p19, p20,
    p21, p22, p23, p24, p25, p26, p27, p28, p29, p30,
    p31, p32, p33, p34, p35, p36, p37, p38, p39, p40,
    p41, p42, p43, p44, p45, p46, p48, p49, p50, ...)
    local a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14
end

printr(f() == nil)

-- upvalues & loops (validates timely closing)
printr((function()
    local res = {}

    for i=1,5 do
        res[#res+1] = (function() return i end)
    end

    local sum = 0
    for i,f in pairs(res) do sum = sum + f() end

    return sum
end)() == 15)

printr((function()
    local res = {}

    for i in ipairs({1,2,3,4,5}) do
        res[#res+1] =(function() return i end)
    end

    local sum = 0
    for i,f in pairs(res) do sum = sum + f() end

    return sum
end)() == 15)

printr((function()
    local res = {}

    local i = 0
    while i <= 5 do
        local j = i
        res[#res+1] = (function() return j end)
        i = i + 1
    end

    local sum = 0
    for i,f in pairs(res) do sum = sum + f() end

    return sum
end)() == 15)

printr((function()
    local res = {}

    local i = 0
    repeat
        local j = i
        res[#res+1] = (function() return j end)
        i = i + 1
    until i > 5

    local sum = 0
    for i,f in pairs(res) do sum = sum + f() end

    return sum
end)() == 15)

-- upvalues & loops & break!
printr((function()
    local res = {}

    for i=1,10 do
        res[#res+1] = (function() return i end)
        if i == 5 then
            break
        end
    end

    local sum = 0
    for i,f in pairs(res) do sum = sum + f() end

    return sum
end)() == 15)

printr((function()
    local res = {}

    for i in ipairs({1,2,3,4,5,6,7,8,9,10}) do
        res[#res+1] =(function() return i end)
        if i == 5 then
            break
        end
    end

    local sum = 0
    for i,f in pairs(res) do sum = sum + f() end

    return sum
end)() == 15)

printr((function()
    local res = {}

    local i = 0
    while i < 10 do
        local j = i
        res[#res+1] = (function() return j end)
        if i == 5 then
            break
        end
        i = i + 1
    end

    local sum = 0
    for i,f in pairs(res) do sum = sum + f() end

    return sum
end)() == 15)

printr((function()
    local res = {}

    local i = 0
    repeat
        local j = i
        res[#res+1] = (function() return j end)
        if i == 5 then
            break
        end
        i = i + 1
    until i >= 10

    local sum = 0
    for i,f in pairs(res) do sum = sum + f() end

    return sum
end)() == 15)

-- ipairs will not iterate through hash part
printr((function()
    local arr = { [1] = 1, [42] = 42, x = 10 }
    local sum = 0
    for i,v in ipairs(arr) do
    sum = sum + v
    end
    return sum
end)() == 1)

-- the reason why this test is interesting is it ensures we do correct mutability analysis for locals
local function chainTest(n)
     local first = nil
     local last = nil

     -- Build chain of n equality constraints
     for i = 0, n do
            local name = "v" .. tostring(i);
            if i == 0 then first = name end
            if i == n then last = name end
     end

     return concat(first, last)
end

printr(chainTest(100) == "v0,v100")

-- this validates import fallbacks
printr(idontexist == nil)
printr(math.idontexist == nil)

-- make sure that NaN is preserved by the bytecode compiler
local realnan = tostring(math.abs(0)/math.abs(0))
printr(tostring(0/0*0) == realnan)
printr(tostring((-1)^(1/2)) == realnan)

-- make sure that negative zero is preserved by bytecode compiler
printr(tostring(0) == "0")
printr(tostring(-0) == "-0")

--[=[
	-- test newline handling in long strings
	printr((function()
		local s1 = [[
	]]
		local s2 = [[
	
	]]
		 local s3 = [[
	foo
	bar]]
		 local s4 = [[
	foo
	bar
	]]
		return concat(s1,s2,s3,s4)
	end)() == ",\n,foo\nbar,foo\nbar\n")
]=]

-- fastcall
-- positive tests for all simple examples; note that in this case the call is a multret call (nresults=LUA_MULTRET)
printr((function() return math.abs(-5) end)() == 5)
printr((function() local abs = math.abs return abs(-5) end)() == 5)
printr((function() local abs = math.abs function foo() return abs(-5) end return foo() end)() == 5)

-- vararg testing - in this case nparams = LUA_MULTRET and it gets adjusted before execution
printr((function() function foo(...) return math.abs(...) end return foo(-5) end)() == 5)
printr((function() function foo(...) local abs = math.abs return abs(...) end return foo(-5) end)() == 5)
printr((function() local abs = math.abs function foo(...) return abs(...) end return foo(-5) end)() == 5)

-- very large unpack
print( table.unpack({1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1}))

-- basic continue in for/while/repeat loops
printr((function() local a = 1 for i=1,8 do a = a + 1 if a < 5 then continue end a = a * 2 end return a end)() == 190)
printr((function() local a = 1 while a < 100 do a = a + 1 if a < 5 then continue end a = a * 2 end return a end)() == 190)
printr((function() local a = 1 repeat a = a + 1 if a < 5 then continue end a = a * 2 until a > 100 return a end)() == 190)

-- shrinking array part
printr((function()
    local t = table.create(100, 42)
    for i=1,90 do t[i] = nil end
    t[101] = 42
    local sum = 0
    for _,v in ipairs(t) do sum += v end
    for _,v in pairs(t) do sum += v end
    return sum
end)() == 462)

-- upvalues: recursive capture
printr((function() local function fact(n) return n < 1 and 1 or n * fact(n-1) end return fact(5) end)() == 120)

-- basic compound assignment
printr((function()
    local a = 1
    local b = 2
    local c = { value = 3 }
    local d = { 4 }
    local e = 3
    local f = 2

    a += 5
    b -= a
    c.value *= 3
    d[1] /= b
    e %= 2
    f ^= 4

    return a,b,c.value,d[1],e,f
end)())

-- compound concat
printr((function()
    local a = 'a'

    a ..= 'b'
    a ..= 'c' .. 'd'
    a ..= 'e' .. 'f' .. a

    return a
end)() == "abcdefabcd")

-- compound assignment with side effects validates lhs is evaluated once
printr((function()
    local res = { 1, 2, 3 }
    local count = 0

    res[(function() count += 1 return count end)()] += 5
    res[(function() count += 1 return count end)()] += 6
    res[(function() count += 1 return count end)()] += 7

    return res[1], res[2], res[3]
end)())

-- checking for a CFG issue that was missed in IR
printr((function(b)
    local res = 0

    if b then
        for i = 1, 100 do
            res += i
        end
    else
        res += 100000
    end

    return res
end)(true) == 5050)
