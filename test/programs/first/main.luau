local webargs = args.web()

local layout = require "./layout"
local e, c = layout.elements, layout.config

local function page(content: layout.Tags)
	local body: layout.Tags = {
		e.nav {
			e.box {
				padding = c.paddingAll(4),
				list = c.listHorizontal(2),
			} {
				e.link "/" { "Home" },
				e.link "/system" { "System" },
			},
		},

		e.main { padding = c.paddingAll(4) }(content),
	}

	return e.rets({}, body)
end

if webargs.url.path == "/" then
	local home = require "./home"
	return page(home(webargs))
elseif webargs.url.path == "/system" then
	local system = require "./system"
	return page(system(webargs))
end

return page {
	e.h(1, { "404" }),
	e.box { padding = c.paddingXY(0, 4) } {
		e.p { "Page not found." },
	},
}
