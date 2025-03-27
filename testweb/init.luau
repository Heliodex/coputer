local webargs = args.web()

local body = `hello {webargs.url.path} world!`

return {
	headers = {
		["content-type"] = "text/html",
	},
	body = buffer.fromstring(body),
} :: WebRes
