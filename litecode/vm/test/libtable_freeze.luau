print(table.isfrozen(math))
print(table.isfrozen {})
print(table.isfrozen(table.freeze {}))

local t = { x = 1 }

print(table.isfrozen(t))
table.freeze(t)
print(table.isfrozen(t))
