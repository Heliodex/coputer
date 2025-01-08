print "firstfile"

local r = require "./f3"
print(r.requiredFrom)
r.requiredFrom = "f1"
print(r.requiredFrom)

return {}
