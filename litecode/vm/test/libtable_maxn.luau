local x = {
	[1] = 1,
	[2] = 2,
	[3] = 3,
	[7] = 7,
	[29] = 29,
	[29.7] = 29.7,
	[-math.huge] = 29,
	[123456789] = 29,
}

print(x[1])
print(x[2])
print(x[3])
print(x[7])
print(x[29])
print(x[29.7])
print(table.maxn(x))
print(#x)
