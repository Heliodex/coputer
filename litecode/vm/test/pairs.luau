local tb: { [any]: any } = {
	balls = 5,
	6,
	7,
	8,
	9,
	10,
}

for i, v in pairs(tb) do
	print(i, v)
end
for i, v in ipairs(tb) do
	print(i, v)
end

print()
print(next(tb))
print(next(tb, 3))
print(next(tb, "balls"))
