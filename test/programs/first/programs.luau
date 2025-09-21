local layout = require "./layout"
local e = layout.elements

return {
	e.h(1, { "Programs" }),
	e.box { padding = { top = 4, bottom = 6 } } {
		e.p {
			"This page provides nothing at the moment lol",
		},
	},
} :: layout.Tags
