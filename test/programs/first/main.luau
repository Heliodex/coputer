local webargs = args.web()

local html = require "./html"
local render, tag = html.render, html.tag

local css = [[
* {
	margin: 0;
	box-sizing: border-box;
	font-family: sans-serif;
}

body {
	padding: 1rem;
}

h1, h2, h3, h4, h5, h6 {
	padding: 1rem 0 0.5rem 0;
}
]]

local head: html.Tags = {
	tag "style" {}(css),
}

local function rets(body: html.Tags): WebRes
	local root = tag "html" { lang = "en" } {
		tag "head" {}(head),
		tag "body" {}(body),
	}

	return {
		headers = {
			["content-type"] = "text/html; charset=utf-8",
		},
		body = buffer.fromstring("<!doctype html>" .. render(root)),
	} :: WebRes
end

local headertext: html.Tags = {}
for k, v in webargs.headers do
	table.insert(headertext, tag "p" {} { k, " = ", v })
end

local querytext: html.Tags = {}
for k, vs in webargs.url.query do
	for _, v in vs do
		table.insert(querytext, tag "p" {} { k, " = ", v })
	end
end

return rets {
	tag "h1" {} "Hello, world!",
	tag "p" {} "Welcome to the first ever page hosted on Coputer!",
	tag "details" {} {
		tag "summary" {} "Here's what the page knows about the request:",
		tag "ul" {} {
			tag "div" {} {
				tag "b" {} "Body:",
				tag "p" {} {
					buffer.tostring(webargs.body),
					`({buffer.len(webargs.body)} bytes)`,
				},
			},
			tag "div" {} {
				tag "b" {} "Headers:",
				tag "p" {}(headertext),
			},
			tag "div" {} {
				tag "b" {} "Method:",
				tag "p" {}(webargs.method),
			},
			tag "div" {} {
				tag "b" {} "URL:",
				tag "ul" {} {
					tag "div" {} {
						tag "b" {} "Path:",
						tag "p" {}(webargs.url.path),
					},
					tag "div" {} {
						tag "b" {} "Query:",
						tag "p" {}(querytext),
					},
					tag "div" {} {
						tag "b" {} "Raw path:",
						tag "p" {}(webargs.url.rawpath),
					},
					tag "div" {} {
						tag "b" {} "Raw query:",
						tag "p" {}(webargs.url.rawquery),
					},
				},
			},
		},
		tag "p" {} "...and nothing more.",
	},

	tag "h2" {} "What is a Coputer?",
	tag "figure" { style = "padding-bottom: 1rem" } {
		tag "img" {
			src = "https://raw.githubusercontent.com/Heliodex/coputer/638eb77c0e7867eaaf53b078a2cb61b03181d37b/coflower.svg",
		} {},
		tag "figcaption" {} {
			tag "small" {} "This munchkin is the icon/logo for Coputer.",
		},
	},
	tag "p" {} "Coputer is a network for running programs like the one used to serve you this page.",
	tag "p" {} {
		"These 'Coputer programs' are written in ",
		tag "a" {
			href = "https://luau.org",
			target = "_blank",
		} "Luau",
		" and are entirely deterministic.",
	},
	tag "p" {} "The network is decentralised and collaborative &ndash; this page could have been delivered to you from anywhere across the network, even if some of its nodes are unresponsive, have blocked you, or refuse to serve this page.",
	tag "p" {} "Anybody should be able to upload a Coputer program for free and have it be accessible 24/7. (This probably sounds less magical now than it would have several years ago.)",
}
