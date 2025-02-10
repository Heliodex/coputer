for w in string.gmatch("first second word", "%w+") do
	print(w)
end

local a = 0
for i in string.gmatch("abcde", "()") do
	print(i)
	a = i
end
print(a, 6)

local t1 = { n = 0 }
for w in string.gmatch("first second word", "%w+") do
	t1.n += 1
	t1[t1.n] = w
end
print(t1[1], t1[2], t1[3])

local t2 = { 3, 6, 9 }
for i in string.gmatch("xuxx uu ppar r", "()(.)%2") do
	print(i, table.remove(t2, 1))
end
print(#t2)

for i, v in string.gmatch("13 14 10 = 11, 15= 16, 22=23", "(%d+)%s*=%s*(%d+)") do
	print(i, v)
end
