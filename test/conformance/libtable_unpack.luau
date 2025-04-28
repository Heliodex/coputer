local u = { 1, 2, 3, 4, 5, 6, 7, [9] = 6, bruh = 6 }
print(table.unpack(u))
print(table.unpack(u, 1, 5))
print(table.unpack(u, 6, 10))
