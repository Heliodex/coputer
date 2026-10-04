-- leading comment
local x = 1 -- trailing comment
--[[ block comment ]]
--[[
	tabbed
		block
			comment
]]
--[=[
nested
--[[ block ]]
comment
]=]
--
local function f()
	-- inner comment
	return x -- return comment
end
do
	--[[ multi
	line ]]
end
-- final comment
--
