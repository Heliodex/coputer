-- two implementations of a sort function
-- this is an example only. Lua has now a built-in function "sort"

type cp = string | number

-- extracted from Programming Pearls, page 110
local function qsort(x: { cp }, l: number, u: number, f: (cp, cp) -> boolean)
	-- print(l, u)
	if l < u then
		local m = (u - (l - 1)) // 2 + l - 1 -- choose a random pivot in range l..u
		x[l], x[m] = x[m], x[l] -- swap pivot to first position
		local t = x[l] -- pivot value
		m = l
		local i = l + 1
		-- print(i, m)
		while i <= u do
			-- invariant: x[l+1..m] < t <= x[m+1..i-1]
			if f(x[i], t) then
				m += 1
				x[m], x[i] = x[i], x[m] -- swap x[i] and x[m]
			end
			i += 1
		end
		x[l], x[m] = x[m], x[l] -- swap pivot to a valid place
		-- x[l+1..m-1] < x[m] <= x[m+1..u]
		qsort(x, l, m - 1, f)
		qsort(x, m + 1, u, f)
	end
end

local function selectionsort(x: { cp }, n: number, f: (cp, cp) -> boolean)
	local i = 1
	while i <= n do
		local m, j = i, i + 1
		while j <= n do
			if f(x[j], x[m]) then m = j end
			j += 1
		end
		x[i], x[m] = x[m], x[i] -- swap x[i] and x[m]
		i += 1
	end
end

local function show(m: string, x: { cp })
	local out = ""

	out ..= (m .. "\n\t")
	local i = 1
	while x[i] do
		out ..= x[i]
		i += 1
		if x[i] then
			out ..= ", "
		end
	end
	print(out)
end

local function testsorts(x)
	local n = 1
	while x[n] do
		n += 1
	end
	n -= 1 -- count elements
	show("original", x)
	qsort(x, 1, n, function(x, y) return x < y end)
	show("after quicksort", x)
	selectionsort(x, n, function(x, y) return x > y end)
	show("after reverse selection sort", x)
	qsort(x, 1, n, function(x, y) return x < y end)
	show("after quicksort again", x)
end

-- array to be sorted
local x = {
	"Jan",
	"Feb",
	"Mar",
	"Apr",
	"May",
	"Jun",
	"Jul",
	"Aug",
	"Sep",
	"Oct",
	"Nov",
	"Dec",
}

-- for loops = 1, 10000 do
testsorts(x)
-- end
