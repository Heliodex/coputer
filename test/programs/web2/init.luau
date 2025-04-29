local webargs = args.web()

-- web2 lel

local method = webargs.method
if method ~= "GET" then return { statuscode = 405 } :: WebRes end

local path = webargs.url.path
local headers = { ["content-type"] = "text/html; charset=utf-8" }

if path == "/" then
	return {
		headers = headers,
		body = buffer.fromstring "<h1>WELCOME TO MY WEBSITE</h1>",
	} :: WebRes
end

if path == "/hello" then
	return {
		headers = headers,
		body = buffer.fromstring "<p>hello page</p>",
	} :: WebRes
end

if path == "/error" then return { statuscode = 454 } :: WebRes end

return { statuscode = 404 } :: WebRes
