declare extern type Foo with
	prop: number
	function method(self, foo: number): string
end

declare extern type Bar extends Foo with
	prop2: string
end
