local c = coroutine.create(function() print "yeah" end)

print(coroutine.resume(c))
coroutine.close(c)
print(coroutine.resume(c))
