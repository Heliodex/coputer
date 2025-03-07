-- test *everything*

local x = "test"
local y = "right"

print(x > y)
print(x < y)
print(x >= y)
print(x <= y)
print(x == y)
print(x ~= y)

local a = 6.7
local b = 1.2

print(a < b)
print(a > b)
print(a <= b)
print(a >= b)
print(a == b)
print(a ~= b)

print("alo" < "alo1")
print("" < "a")
print("alo\0alo" < "alo\0b")
print("alo\0alo\0\0" > "alo\0alo\0")
print("alo" < "alo\0")
print("alo\0" > "alo")
print("\0" < "\1")
print("\0\0" < "\0\1")
print("\1\0a\0a" <= "\1\0a\0a")
print("\1\0a\0b" <= "\1\0a\0a")
print("\0\0\0" < "\0\0\0\0")
print("\0\0\0\0" < "\0\0\0")
print("\0\0\0" <= "\0\0\0\0")
print("\0\0\0\0" <= "\0\0\0")
print("\0\0\0" <= "\0\0\0")
print("\0\0\0" >= "\0\0\0")
print("\0\0b" < "\0\0a\0")
