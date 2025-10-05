require "./dir/multret" -- print "ret"
print "ran"
local found = require "./dir/multret"

print(found.x) -- only returns one now

local found2 = require "./dir/multret2"
print "ran2"

print(found2()) -- found 2 fall in love
