local ws = coroutine.wrap(function() end)
ws()
coroutine.yield(ws())
