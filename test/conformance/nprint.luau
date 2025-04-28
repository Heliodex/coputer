print(5)
print(0)
print(-0)
print(6)
print(10)
print(100)
print(1e6)
print(1e9)
print(1e12)
print(1e15)
print(1e16)
print(12456789012345678)
print(124567890123456789)
print(1 / 2)
print(1 / 3)
print(0.1 + 0.2)
print(1e19)
print(1e20)
print(1e21)
print(1e36)
print(1e150)
print(1.5e-150)
print(-2e150)
print(-12.5e-150)
print(1 / 0)
print(0 / 0)
print(33 * 53 * 1.4)
print(33 * 53 ^ 1.4)
print(-1e20)
print(-1e-20)
print(math.pi)
print(math.huge)
print(0.111111111111111111111111111)
print(0.222222222222222222222222222)
print(0.000001)
print(0.000002)
print(5.5)

for i = 0, 10 do
	print(6.6 ^ i)
end
for i = 0, 10 do
	print(-6.6 ^ i)
end
-- i assume there's differences between math.Pow and math.h  pow
print(1035102341.4011246)
print(1035102341.4011247)
for i = 12, 15 do
	print(6.6 ^ i)
end

-- zero
print(0)
print(0 / -1)

-- specials
print(1 / 0)
print(-1 / 0)
print(0 / 0)

-- integers
print(1)
print(42)
print(-4294967296)
print(9007199254740991)

-- decimals
print(0.5)
print(0.1)
print(-0.17)
print(math.pi)

-- scientific
print(1e+30)
print(-1e+24)

-- fuzzing corpus
-- Note: If the assert below fires it may indicate floating point denormalized values are not handled as expected.
print(5.4536123983019448e-311)
print(5.4834368411298348e-311)
print(4.4154895841930002e-305)
print(1125968630513728)
print(3.3951932655938423e-313)
print(1.625)
print(4.9406564584124654e-324)
print(2.0049288280105384)
print(3.0517578125e-05)
print(1.383544921875)
print(3.0053350932691001)
print(0.0001373291015625)
print(-1.9490628022799998e+289)
print(-0.00610404721867928)
print(0.00014495849609375)
print(0.453125)
print(-4.2375343999999997e+73)
print(1.3202313930270133e-192)
print(3.6984408976312836e+19)
print(2.0563000527063302)
print(4.8970527433648997e-260)
print(1.62890625)
print(1.1295093211933533e+65)
