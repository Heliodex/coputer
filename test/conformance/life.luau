local ALIVE, DEAD = "#", " "

type shape = { { number } }
type cells = {
	[number]: { number },
	h: number,
	w: number,
}

local function ARRAY2D(w, h)
	local t = { w = w, h = h } :: cells
	for y = 1, h do
		t[y] = {}
		for x = 1, w do
			t[y][x] = 0
		end
	end
	return t
end

-- give birth to a "shape" within the cell array
local function spawn(c: cells, shape: shape, left: number, top: number)
	for y = 0, #shape - 1 do
		for x = 0, #shape[1] - 1 do
			c[top + y][left + x] = shape[y + 1][x + 1]
		end
	end
end

-- run the CA and produce the next generation
local function evolve(c: cells, next: cells)
	local ym1, y, yp1, yi = c.h - 1, c.h, 1, c.h
	while yi > 0 do
		local xm1, x, xp1, xi = c.w - 1, c.w, 1, c.w
		while xi > 0 do
			local sum = c[ym1][xm1]
				+ c[ym1][x]
				+ c[ym1][xp1]
				+ c[y][xm1]
				+ c[y][xp1]
				+ c[yp1][xm1]
				+ c[yp1][x]
				+ c[yp1][xp1]

			next[y][x] = if sum == 2 then c[y][x] elseif sum == 3 then 1 else 0
			xm1, x, xp1, xi = x, xp1, xp1 + 1, xi - 1
		end
		ym1, y, yp1, yi = y, yp1, yp1 + 1, yi - 1
	end
end

-- output the array to screen
local function draw(c: cells)
	for y = 1, c.h do
		local out = "" -- accumulate to reduce flicker
		for x = 1, c.w do
			out ..= if c[y][x] == 1 then ALIVE else DEAD
		end
		print(out)
	end
end

--
-- shapes suitable for use with spawn() above
--
local GLIDER = { { 0, 0, 1 }, { 1, 0, 1 }, { 0, 1, 1 } } :: shape
local EXPLODE = {
	{ 0, 1, 0 },
	{ 1, 1, 1 },
	{ 1, 0, 1 },
	{ 0, 1, 0 },
} :: shape
local FISH = {
	{ 0, 1, 1, 1, 1 },
	{ 1, 0, 0, 0, 1 },
	{ 0, 0, 0, 0, 1 },
	{ 1, 0, 0, 1, 0 },
} :: shape

-- the main routine
local function LIFE(w, h)
	-- create two arrays
	local thisgen = ARRAY2D(w, h)
	local nextgen = ARRAY2D(w, h)

	-- create some life
	-- about 1000 generations of fun, then a glider steady-state
	spawn(thisgen, GLIDER, 5, 4)
	spawn(thisgen, EXPLODE, 25, 10)
	spawn(thisgen, FISH, 4, 12)

	for gen = 1, 100 do
		evolve(thisgen, nextgen)
		thisgen, nextgen = nextgen, thisgen
	end
	draw(thisgen)
end

LIFE(40, 20)
