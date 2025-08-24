local x = 'this is a "test"'
local y = [[this is a "test"]]
local z = `this 'is' a "test"`

do
	local x = "this is\n]]a test"
	local y = [[this is
a test]]
	local z = `this is\na test`
end
