local layout = {}

function layout.box(options: string)
	return function() return parseOptions(options) end
end

return layout
