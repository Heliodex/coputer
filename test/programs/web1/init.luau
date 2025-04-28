local webargs = args.web()

local body = `hello {webargs.method} {webargs.url.path} world! {webargs.url.rawpath}`

return {
	headers = {
		["content-type"] = "text/html",
	},
	body = buffer.fromstring(body),
} :: WebRes
