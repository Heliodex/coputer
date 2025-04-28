local t = { x = 5 }

function t.fn(self: typeof(t))
	return self.x * 2
end

print(t:fn())
