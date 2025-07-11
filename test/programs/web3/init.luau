local webargs = args.web()

local path = webargs.url.path

if path == "/" then
	local body = ""
	for k, v in webargs.url.query do
		body ..= `{k} =\n`
		for _, val in v do
			body ..= `- {val}\n`
		end
	end

	body ..= `Raw query: {webargs.url.rawquery}`

	return {
		body = buffer.fromstring(body),
	} :: WebRes
end

return { statuscode = 404 } :: WebRes
