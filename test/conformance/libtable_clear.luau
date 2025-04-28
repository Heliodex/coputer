local t = {
	"one",
	a = 1,
}

local function tt()
	print(t[1])
	print(t.a)
	print(#t)
end

tt()
table.clear(t)
tt()
