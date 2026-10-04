const table = {
	1,
	-- intentional
	23,
	4,
	5,
	6,
	7,
	8,
	9,
	0,
	-- wrap around again
	1,
	2,
	3,
	4,
	5,
	--[[ oh yeah half way ]]
	6,
	-- halfway through the 2nd half I mean
	7,
	8,
	9,
	0,
}

local nested = {
	-- outer comment
	inner = {
		-- inner comment
		1,
		--[[ inner block ]]
		2,
	},
	-- after inner
	3,
}

local blocks = {
	-- line comment
	--[[ block comment ]]
	--[==[ long block
	with equals ]==]
	--[[ multi
	line ]]
	--[[
	block on its own lines
	]]
	--
	1,
	2, -- trailing
	-- after last
}

local empty = {
	-- only a comment
}

local general = {
	-- before key
	[2] = "two",
	--[[ between ]]
	["not an id"] = 1,
}

local deep = {
	a = {
		b = {
			-- deep comment
			c = 1,
		},
		-- middle
		d = 2,
	},
	-- top
	e = 3,
}
