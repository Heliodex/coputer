local words = {
	"one",
	"two",
	"three",
	"four",
	"five",
	"six",
	"seven",
	"eight",
	"nine",
	"ten",
	"eleven",
	"twelve",
	"thirteen",
	"fourteen",
	"fifteen",
	"sixteen",
	"seventeen",
	"eighteen",
	"nineteen",
	"twenty",
	[30] = "thirty",
	[40] = "forty",
	[50] = "fifty",
	[60] = "sixty",
	[70] = "seventy",
	[80] = "eighty",
	[90] = "ninety",
}

local function getWord(n: number)
	if words[n] then return words[n] end

	local mod, div = n % 100, n // 100
	if mod == 0 then return getWord(div) .. "hundred" end
	if n > 100 then return getWord(div) .. "hundredand" .. getWord(mod) end

	return getWord(n // 10 * 10) .. words[n % 10]
end

local sum = #"onethousand"
for i = 1, 999 do
	sum += #getWord(i)
end

print(sum)
