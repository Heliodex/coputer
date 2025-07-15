export type Tag = {
	name: string,
	attributes: { [string]: boolean | string },
	children: Tags,
}
export type Tags = { string | Tag }

type TagMap = { [string]: true? }

local voidTags: TagMap = {
	area = true,
	base = true,
	br = true,
	col = true,
	embed = true,
	hr = true,
	img = true,
	input = true,
	link = true,
	meta = true,
	source = true,
	track = true,
	wbr = true,
}

-- fusion vibez
local function tag(name: string)
	return function(attributes: { [string]: string | boolean })
		return function(children: string | Tags)
			return {
				name = name,
				attributes = attributes,
				children = if type(children) == "string"
					then { children }
					else children,
			} :: Tag
		end
	end
end

local rawTextTags: { [string]: boolean? } = {
	script = true,
	style = true,
	textarea = true,
	title = true,
}

local function sanitise(str: string): string
	str = string.gsub(str, "<", "&lt;")
	str = string.gsub(str, ">", "&gt;")
	return str
end

local function render(t: Tag): string
	local void = voidTags[t.name] or false
	local rawText = rawTextTags[t.name] or false

	-- start
	local rendered = { `<{t.name}` }

	-- attributes
	for attr, value in t.attributes do
		table.insert(rendered, ` {attr}`)
		if type(value) == "string" then
			table.insert(rendered, `="{string.gsub(value, '"', "&quot;")}"`)
		end
	end
	table.insert(rendered, ">")

	if void then return table.concat(rendered, "") end

	-- children
	for _, child in t.children do
		table.insert(
			rendered,
			if type(child) == "string"
				then if rawText then child else sanitise(child)
				else render(child)
		)
	end

	-- end tag
	table.insert(rendered, `</{t.name}>`)

	return table.concat(rendered, "")
end

return {
	render = render,
	tag = tag,
}
