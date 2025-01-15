local ws
local cs = function() coroutine.yield(ws()) end

ws = coroutine.wrap(cs)
print(ws())
