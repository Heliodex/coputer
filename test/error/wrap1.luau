local ws
local function cs() coroutine.yield(ws()) end

ws = coroutine.wrap(cs)
print(ws())
