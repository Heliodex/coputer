-- the file that uncovered the bug that doubleloop.luau was needed to solve
-- https://gist.github.com/Heliodex/799bb679b7e825b09c74ed47606ba717

type M = "M"
local m: M = "M"

type I = "I"
local i: I = "I"

type U = "U"
local u: U = "U"

type MIU = M | I | U
type String = { MIU }

local collection: { String } = {
	{ m, i },
} :: { String }

local function equals(a: String, b: String): boolean
	if #a ~= #b then return false end

	for j = 1, #a do
		if a[j] ~= b[j] then return false end
	end

	return true
end

local function has(c: { String }, s: String)
	for _, v in collection do
		if equals(v, s) then return true end
	end
	return false
end

local function add(s: String)
	if not has(collection, s) then table.insert(collection, s) end
end

local function str(s: { String })
	local ss = {}
	for _, v in s do
		table.insert(ss, table.concat(v))
	end

	print(table.concat(ss, " "))
end

-- Rule I: If you possess a string whose last letter is I, you can add on a U at the end.
local function rule1(s: String): { String }
	if s[#s] ~= i then return {} end

	local copy = table.clone(s)
	table.insert(copy, u)
	return { copy }
end

-- Rule II: Suppose you have Mx. Then you may add Mxx to your collection.
local function rule2(s: String): { String }
	if s[1] ~= m then return {} end

	local copy = table.clone(s)
	for j = 2, #s do
		table.insert(copy, s[j])
	end

	return { copy }
end

-- Rule III: If III occurs in one of the strings in your collection, you may make a new string with U in place of III.
local function rule3(s: String): { String }
	if #s < 3 then return {} end
	local results = {}

	for j = 1, #s - 2 do
		if s[j] ~= i or s[j + 1] ~= i or s[j + 2] ~= i then continue end
		local copy = table.clone(s)
		copy[j] = u
		table.remove(copy, j + 1)
		table.remove(copy, j + 1)
		table.insert(results, copy)
	end

	return results
end

-- Rule IV: If UU occurs inside one of your strings, you can drop it.
local function rule4(s: String): { String }
	if #s < 2 then return {} end
	local results = {}

	for j = 1, #s - 1 do
		if s[j] ~= u or s[j + 1] ~= u then continue end
		local copy = table.clone(s)
		table.remove(copy, j)
		table.remove(copy, j)
		table.insert(results, copy)
	end

	return results
end

local fns: { (String) -> { String } } = { rule1, rule2, rule3, rule4 }

-- There is a bug due to the order of the functions being applied I think (eg. it doesn't produce MIIIUUIUU at the right time) but I cba figuring it out atm
for idx = 1, 6 do
	print(`\nCollection of {#collection}`)
	str(collection)

	print(`--- Iteration {idx} ---`)
	local tcc = table.clone(collection)
	for fni, fn in fns do
		-- print(`\nRule {fni}`)
		for _, v in tcc do
			local strs = fn(v)
			-- str(strs)

			for _, s in strs do
				add(s)
			end
		end
	end
end

print(`\nCollection of {#collection}`)
str(collection)
