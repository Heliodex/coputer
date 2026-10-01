local single = { 1, 2, 3 }
local record = { a = 1, b = 2 }
local general = { ["key"] = 1, [2] = "two" }
local keys = { ["ok"] = 1, ["_also_ok"] = 2, ["not an id"] = 3, ["end"] = 4, ["1abc"] = 5, ["with-hyphen"] = 6 }
local mixed = {
	1,
	a = 2,
	[3] = 4,
}
local mixed2 = { 1, a = 2, [3] = 4 }
local empty = {}
