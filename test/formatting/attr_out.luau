@native
@deprecated
local function leftpad(str: string, len: number, char: string): string
	local pad = char:rep(len - #str)

	return pad .. str
end
