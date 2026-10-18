-- string library: sub/byte/char/format return fresh values and format goes through the std arg machinery
local s = string.rep("hello world, ", 500)
local checksum = 0

for i = 1, 3000 do
	local p = i % #s + 1
	checksum += #string.sub(s, p, p + 15)
	checksum += string.byte(s, p)
	checksum += #string.char(i % 256)
	checksum += #string.format("%d:%s", i, string.char(i % 256))
end

local parts = {}
for i = 1, 20000 do
	parts[i] = string.format("%d", i)
end

local joined = table.concat(parts, ",")
print(checksum, #joined, #string.gsub(joined, "1", "x"))
