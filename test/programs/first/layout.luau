local elements = {}
local config = {}

local html = require "./html"

export type Tags = html.Tags

function elements.h(level: number, children: html.Tags)
	if level == 1 then return html.tag "h1" {}(children) end
	if level == 2 then return html.tag "h2" {}(children) end
	if level == 3 then return html.tag "h3" {}(children) end
	if level == 4 then return html.tag "h4" {}(children) end
	if level == 5 then return html.tag "h5" {}(children) end
	if level == 6 then return html.tag "h6" {}(children) end
	return html.tag "h1" {}(children) -- default to h1 if level is invalid
end

function elements.bold(text: string) return html.tag "b" {} { text } end

function elements.p(children: html.Tags) return html.tag "p" {}(children) end

function elements.link(url: string)
	return function(children: html.Tags)
		return html.tag "a" { href = url }(children)
	end
end

function elements.outlink(url: string)
	return function(children: html.Tags)
		return html.tag "a" { href = url, target = "_blank" }(children)
	end
end

function elements.list(items: html.Tags) return html.tag "ul" {}(items) end

function elements.details(summary: html.Tags)
	return function(children: html.Tags)
		return html.tag "details" {} {
			html.tag "summary" {}(summary),
			table.unpack(children),
		}
	end
end

local unit = 0.25
local unitunit = "rem"

type Options = {
	padding: Padding?,
}

type Padding = {
	top: number?,
	right: number?,
	bottom: number?,
	left: number?,
}

function config.paddingAll(padding: number): Padding
	return {
		top = padding,
		right = padding,
		bottom = padding,
		left = padding,
	}
end

function config.paddingXY(px: number, py: number): Padding
	return {
		top = py,
		right = px,
		bottom = py,
		left = px,
	}
end

local function parseOptions(options: Options)
	local attrs: html.Attrs = {}

	if options.padding then
		local style = ""
		if attrs.style and attrs.style ~= true then style = attrs.style end

		if options.padding.top then
			style ..= `padding-top: {options.padding.top * unit}{unitunit};`
		end
		if options.padding.right then
			style ..= `padding-right: {options.padding.right * unit}{unitunit};`
		end
		if options.padding.bottom then
			style ..= `padding-bottom: {options.padding.bottom * unit}{unitunit};`
		end
		if options.padding.left then
			style ..= `padding-left: {options.padding.left * unit}{unitunit};`
		end

		attrs.style = style
	end

	return attrs
end

function elements.box(options: Options)
	return function(children: html.Tags)
		return html.tag "div"(parseOptions(options))(children)
	end
end

function elements.main(options: Options)
	return function(children: html.Tags)
		return html.tag "main"(parseOptions(options))(children)
	end
end

function elements.image(
	src: string,
	alt: string,
	caption: html.Tags
): html.ContentTag
	return html.tag "figure" {} {
		html.tagVoid "img" { src = src, alt = alt },
		html.tag "figcaption" {}(caption),
	}
end

local defaultCss = [[
* {
	margin: 0;
	box-sizing: border-box;
	font-family: sans-serif;
}
]]

local defaultStyle = html.tag "style" {} { defaultCss }

function elements.rets(head: html.Tags, body: html.Tags): WebRes
	local root = html.tag "html" { lang = "en" } {
		html.tag "head" {} { table.unpack(head), defaultStyle },
		html.tag "body" {}(body),
	}

	return {
		headers = {
			["content-type"] = "text/html; charset=utf-8",
		},
		body = buffer.fromstring(
			"<!doctype html>" .. html.renderContentTag(root)
		),
	} :: WebRes
end

return {
	elements = elements,
	config = config,
}
