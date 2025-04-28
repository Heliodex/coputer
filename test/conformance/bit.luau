local bit_or, bit_and, bit_not, bit_rshift

local function tbl2number(tbl: { number })
	local result = 0
	local power = 1
	for i = 1, #tbl do
		result += tbl[i] * power
		power = power * 2
	end
	return result
end

local function expand(t1: { any }, t2: { any })
	local big, small = t1, t2
	if #big < #small then
		big, small = small, big
	end
	-- expand small
	for i = #small + 1, #big do
		small[i] = 0
	end
end

local to_bits -- needs to be declared before bit_not

function bit_not(n: number): number
	local tbl = to_bits(n)
	local size = math.max(#tbl, 32)
	for i = 1, size do
		if tbl[i] == 1 then
			tbl[i] = 0
		else
			tbl[i] = 1
		end
	end
	return tbl2number(tbl)
end

-- defined as local above
function to_bits(n)
	if n < 0 then
		-- negative
		return to_bits(bit_not(math.abs(n)) + 1)
	end
	-- to bits table
	local tbl = {}
	local cnt = 1
	local last = 0
	while n > 0 do
		last = n % 2
		tbl[cnt] = last
		n = (n - last) / 2
		cnt = cnt + 1
	end

	return tbl
end

function bit_or(m, n)
	local tbl_m = to_bits(m)
	local tbl_n = to_bits(n)
	expand(tbl_m, tbl_n)

	local tbl = {}
	for i = 1, #tbl_m do
		if tbl_m[i] == 0 and tbl_n[i] == 0 then
			tbl[i] = 0
		else
			tbl[i] = 1
		end
	end

	return tbl2number(tbl)
end

function bit_and(m, n)
	local tbl_m = to_bits(m)
	local tbl_n = to_bits(n)
	expand(tbl_m, tbl_n)

	local tbl = {}
	for i = 1, #tbl_m do
		if tbl_m[i] == 0 or tbl_n[i] == 0 then
			tbl[i] = 0
		else
			tbl[i] = 1
		end
	end

	return tbl2number(tbl)
end

function bit_rshift(n, bits)
	local high_bit = 0
	if n < 0 then
		-- negative
		n = bit_not(math.abs(n)) + 1
		high_bit = 0x80000000
	end

	local floor = math.floor
	for _ = 1, bits do
		n /= 2
		n = bit_or(floor(n), high_bit)
	end
	return floor(n)
end

local function lei2str(i)
	local function f(s) return string.char(bit_and(bit_rshift(i, s), 255)) end
	local l2 = f(0) .. f(8) .. f(16) .. f(24)
	return l2
end

local P = lei2str(bit_and(8 * 10, 0xFFFFFFFF))
print(P)
print(string.sub(P, 1, 1))
