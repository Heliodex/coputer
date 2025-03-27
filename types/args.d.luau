-- type WebConf = {}

type WebArgs = {
	url: {
		--[[
			Raw path of the web request, after the base URL has been removed. Includes starting '/'.
		]]
		rawpath: string,
		--[[
			Path of the web request made.
		]]
		path: string,
		--[[
			Raw query string, not including '?'.
		]]
		rawquery: string,
		--[[
			Query string, parsed into a table of key-value pairs.
	
			@example
			```luau
			{ a = "1", b = "2" } -- ?a=1&b=2
			```
		]]
		query: { [string]: string },
	},
	--[[
		HTTP method of the request.
	]]
	method: string,
	--[[
		Headers of the request. Headers are case-insensitive, and are converted to lowercase.

		@example
		```luau
		{ ["user-agent"] = "Roblox/WinInet" }
		```
	]]
	headers: { [string]: string },
	--[[
		Body of the request.
	]]
	body: buffer,
}

export type WebRes = {
	--[[
		HTTP status code of the response. Defaults to 200 OK.
	]]
	statuscode: number?,
	--[[
		HTTP status message of the response. Defaults to the canonical status code message.
	]]
	statusmessage: string?,
	--[[
		Headers of the response. Headers are case-insensitive, and will be converted to lowercase.

		@example
		```luau
		{ ["content-type"] = "text/html" }
		```
	]]
	headers: { [string]: string }?,
	--[[
		Body of the response.
	]]
	body: buffer?,
}

-- oh my god we could do insane shit with type functions
-- we'll focus on libraries l8r

-- type Arg<T> = {}
-- local n: any

-- function libraryFn6<A, B, C, D, E, F>(a: Arg<A>, b: Arg<B>, c: Arg<C>, d: Arg<D>, e: Arg<E>, f: Arg<F>): (A, B, C, D, E, F)
-- 	return n :: A, n :: B, n :: C, n :: D, n :: E, n :: F
-- end
-- function libraryFn5<A, B, C, D, E>(a: Arg<A>, b: Arg<B>, c: Arg<C>, d: Arg<D>, e: Arg<E>): (A, B, C, D, E)
-- 	return n :: A, n :: B, n :: C, n :: D, n :: E
-- end
-- function libraryFn4<A, B, C, D>(a: Arg<A>, b: Arg<B>, c: Arg<C>, d: Arg<D>): (A, B, C, D)
-- 	return n :: A, n :: B, n :: C, n :: D
-- end
-- function libraryFn3<A, B, C>(a: Arg<A>, b: Arg<B>, c: Arg<C>): (A, B, C)
-- 	return n :: A, n :: B, n :: C
-- end
-- function libraryFn2<A, B>(a: Arg<A>, b: Arg<B>): (A, B)
-- 	return n :: A, n :: B
-- end
-- function libraryFn1<A>(a: Arg<A>): A
-- 	return n :: A
-- end

declare args: {
	web: () -> WebArgs,
	-- library: {
	-- 	get:
	-- 		& typeof(libraryFn6)
	-- 		& typeof(libraryFn5)
	-- 		& typeof(libraryFn4)
	-- 		& typeof(libraryFn3)
	-- 		& typeof(libraryFn2)
	-- 		& typeof(libraryFn1),
	-- 	number: Arg<number>,
	-- 	string: Arg<string>,
	-- 	boolean: Arg<boolean>,
	-- 	table: Arg<{ [any]: any }>,
	-- 	func: Arg<(any) -> any>,
	-- 	thread: Arg<thread>,
	-- 	buffer: Arg<buffer>,
	-- 	vector: Arg<vector>,
	-- },
}
