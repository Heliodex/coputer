local webargs = args.web()

local body = `hello {webargs.method} {webargs.url.path} world! {webargs.url.rawpath}`

return {
	headers = {
		["content-type"] = "text/plain; charset=utf-8",
	},
	body = buffer.fromstring(body),
} :: WebRes
