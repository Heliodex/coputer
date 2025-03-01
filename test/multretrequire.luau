require "./dir/multret" -- print "ret"
print "ran"
local found = require "./dir/multret"

print(found.x, found.y) -- does it only return the final one???

local found2 = require "./dir/multret2"
print "ran2"

print(found2())
