local webargs = args.web()

local layout = require "./layout"
local e, c = layout.elements, layout.config

local headertext: layout.Tags = {}
for k, v in webargs.headers do
	table.insert(headertext, e.p { k, " = ", v })
end

local querytext: layout.Tags = {}
for k, vs in webargs.url.query do
	for _, v in vs do
		table.insert(querytext, e.p { k, " = ", v })
	end
end

local body: layout.Tags = {
	e.main { padding = c.paddingAll(4) } {
		e.h(1, { "Hello, world!" }),
		e.box { padding = c.paddingXY(0, 4) } {
			e.p { "Welcome to the first ever page hosted on Coputer!" },
			e.details { "Here's what the page knows about the request:" } {
				e.list {
					e.box {} {
						e.bold "Body:",
						e.p {
							buffer.tostring(webargs.body),
							`({buffer.len(webargs.body)} bytes)`,
						},
					},
					e.box {} {
						e.bold "Headers:",
						table.unpack(headertext),
					},
					e.box {} {
						e.bold "Method:",
						e.p { webargs.method },
					},
					e.box {} {
						e.bold "URL:",
						e.list {
							e.box {} {
								e.bold "Path:",
								e.p { webargs.url.path },
							},
							e.box {} {
								e.bold "Query:",
								table.unpack(querytext),
							},
							e.box {} {
								e.bold "Raw path:",
								e.p { webargs.url.rawpath },
							},
							e.box {} {
								e.bold "Raw query:",
								e.p { webargs.url.rawquery },
							},
						},
					},
				},
				e.p { "...and nothing more." },
			},
		},

		e.h(2, { "What is a Coputer?" }),
		e.box { padding = c.paddingXY(0, 4) } {
			e.box { padding = c.paddingAll(4) } {
				e.image(
					"https://raw.githubusercontent.com/Heliodex/coputer/638eb77c0e7867eaaf53b078a2cb61b03181d37b/coflower.svg",
					"Coputer flower logo",
					{ "This munchkin is the icon/logo for Coputer." }
				),
			},
			e.p {
				"Coputer is a network for running programs like the one used to serve you this page.",
			},
			e.p {
				"These 'Coputer programs' are written in ",
				e.outlink "https://luau.org" { "Luau" },
				" and are entirely deterministic.",
			},
			e.p {
				"The network is decentralised and collaborative &ndash; this page could have been delivered to you from anywhere across the network, even if some of its nodes are unresponsive, have blocked you, or refuse to serve this page.",
			},
			e.p {
				"Anybody should be able to upload a Coputer program for free and have it be accessible 24/7. (This probably sounds less magical now than it would have several years ago.)",
			},
		},

		e.box { padding = c.paddingXY(0, 4) } {
			e.p {
				"This page first went live on ",
				e.bold "17 July 2025",
				". Watch this space! It will be updated with more information soon.",
			},

			e.box {} {
				e.outlink "https://github.com/Heliodex/coputer/blob/main/test/programs/first/main.luau" {
					"Source code for this page",
				},
			},
			e.box {} {
				"Source code for Coputer: ",
				e.outlink "https://github.com/Heliodex/coputer" {
					"github.com/Heliodex/coputer",
				},
			},
		},
	},
}

return e.rets({}, body)
