local webargs = args.web()

local layout = require "./layout"
local e, c = layout.elements, layout.config

local function page(code: number, content: layout.Tags)
	local body: layout.Tags = {
		e.nav {
			e.box {
				padding = c.paddingAll(4),
				list = c.listHorizontal(2),
			} {
				e.link "/" { "Home" },
				e.link "/system" { "System" },
				e.link "/programs" { "Programs" },
			},
		},

		e.main { padding = c.paddingAll(4) }(content),
	}

	return e.rets(code, {}, body)
end

if webargs.url.path == "/" then
	local home = require "./home"
	return page(200, home(webargs))
elseif webargs.url.path == "/system" then
	local system = require "./system"
	return page(200, system)
elseif webargs.url.path == "/programs" then
	local programs = require "./programs"
	return page(200, programs)
end

return page(404, {
	e.h(1, { "404" }),
	e.box { padding = c.paddingXY(0, 4) } {
		e.p { "Page not found." },
	},
})
