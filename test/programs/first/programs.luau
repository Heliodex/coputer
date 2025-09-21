local layout = require "./layout"
local e = layout.elements

return {
	e.h(1, { "Programs" }),
	e.box { padding = { top = 4, bottom = 6 } } {
		e.p {
			"This page shows how Coputer programs are written, stored, and executed, and some of the rationale behind the design decisions.",
		},
	},

	e.h(2, { "Language" }, "language"),
	e.box { padding = { top = 4, bottom = 6 } } {
		e.p {
			"All Coputer programs are written in ",
			e.outlink "https://luau.org" { "Luau" },
			", a gradually typed scripting language derived from Lua.",
		},
		e.box {
			padding = { top = 2, bottom = 4, left = 4 },
			list = {
				direction = "horizontal",
				gap = 4,
			},
		} {
			e.image(
				"https://raw.githubusercontent.com/luau-lang/site/2d33fe820e91481dee7e128f42b385489e193905/logo.svg",
				"Luau logo",
				{ "The Luau logo" },
				35
			),
			e.image(
				"https://rfcs.luau.org/mascot.png",
				"Luau seal mascot",
				{ "Hina the Seal, the mascot for Luau" },
				50
			),
		},
	},
} :: layout.Tags
