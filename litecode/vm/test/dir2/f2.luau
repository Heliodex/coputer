print "secondfile"

local r = require "./f3"
print(r.requiredFrom)
r.requiredFrom = "f2"
print(r.requiredFrom)

return {}
