local voidTags = {
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

type VoidTagName = keyof<typeof(voidTags)>

local contentTags = {
	a = true,
	abbr = true,
	address = true,
	article = true,
	aside = true,
	audio = true,
	b = true,
	bdi = true,
	bdo = true,
	blockquote = true,
	body = true,
	button = true,
	canvas = true,
	caption = true,
	cite = true,
	code = true,
	colgroup = true,
	data = true,
	datalist = true,
	dd = true,
	del = true,
	details = true,
	dfn = true,
	dialog = true,
	div = true,
	dl = true,
	dt = true,
	em = true,
	fencedframe = true,
	fieldset = true,
	figcaption = true,
	figure = true,
	footer = true,
	form = true,
	h1 = true,
	h2 = true,
	h3 = true,
	h4 = true,
	h5 = true,
	h6 = true,
	head = true,
	header = true,
	hgroup = true,
	html = true,
	i = true,
	iframe = true,
	ins = true,
	kbd = true,
	label = true,
	legend = true,
	li = true,
	main = true,
	map = true,
	mark = true,
	menu = true,
	meter = true,
	nav = true,
	noscript = true,
	object = true,
	ol = true,
	optgroup = true,
	option = true,
	output = true,
	p = true,
	picture = true,
	pre = true,
	progress = true,
	q = true,
	rp = true,
	rt = true,
	ruby = true,
	s = true,
	samp = true,
	script = true,
	search = true,
	section = true,
	select = true,
	selectedcontent = true,
	slot = true,
	small = true,
	span = true,
	strong = true,
	style = true,
	sub = true,
	summary = true,
	sup = true,
	table = true,
	tbody = true,
	td = true,
	template = true,
	textarea = true,
	tfoot = true,
	th = true,
	thead = true,
	time = true,
	title = true,
	tr = true,
	u = true,
	ul = true,
	var = true,
	video = true,
}

type ContentTagName = keyof<typeof(contentTags)>

export type VoidTag = {
	name: VoidTagName,
	attributes: { [string]: boolean | string },
}
export type ContentTag = {
	name: ContentTagName,
	attributes: { [string]: boolean | string },
	children: Tags,
}
export type Tag = string | VoidTag | ContentTag
export type Tags = { Tag }
export type Attrs = { [string]: string | boolean }

-- fusion vibez
local function tagVoid(name: VoidTagName)
	return function(attributes: Attrs)
		return {
			name = name,
			attributes = attributes,
		} :: VoidTag
	end
end

local function tagContent(name: ContentTagName)
	return function(attributes: Attrs)
		return function(children: Tags)
			return {
				name = name,
				attributes = attributes,
				children = if type(children) == "string"
					then { children }
					else children,
			} :: ContentTag
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

local function renderVoidTag(t: VoidTag): string
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

	return table.concat(rendered, "")
end

local function renderContentTag(t: ContentTag): string
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

	-- children
	for _, child in t.children do
		table.insert(
			rendered,
			if type(child) == "string"
				then if rawText then child else sanitise(child)
				else if voidTags[child.name]
					then renderVoidTag(child :: VoidTag)
					else renderContentTag(child :: ContentTag)
		)
	end

	-- end tag
	table.insert(rendered, `</{t.name}>`)

	return table.concat(rendered, "")
end

return {
	renderVoidTag = renderVoidTag,
	renderContentTag = renderContentTag,
	tagVoid = tagVoid,
	tag = tagContent,
}
