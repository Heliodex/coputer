local t = {
	"one",
	a = 1,
}

t[2] = "two"
local t2 = t
local t3 = table.clone(t)
t2[2] = "three"
t3[2] = "four"

print(t[2])
print(t2[2])
print(t3[2])

local t4 = table.clone { 1, 2, 3, 4, 5, 6, 7, 8, 9, 10 }
print(t4[1], t4[10], #t4)

local t5 = table.clone { [88] = 5, hello = "world" }
print(t5[88], t5.hello, #t5)
