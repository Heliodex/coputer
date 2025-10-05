local a = {}

function a:method()
	print "method called"
	print(not self)
end

a:method()
a.method(a)
