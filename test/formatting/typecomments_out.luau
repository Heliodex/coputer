type x = {
	a: number,
	-- e is a turing machine (with type (string) -> string)
	e: (string) -> string,
	b: string,
	-- A number
	c: A<number>,
	d: typeof(x),
}
type y = {
	f: {
		g: {
			h: number,
			-- deeper comment
			--- doc comment
			---- what comment
			i: string,
		},
	},
}
type z = {
	--[[ doc moment
	yeah ]]
	[string]: {
		-- comment2
		j: number,
	},
}
