-- Export value syntax (`export local`, `export function`, `export const`)
export local a = 1
export const b = 2
export function f(x)
	return x
end
export local c, d = 3, 4

assert(a == 1)
assert(b == 2)
assert(f(5) == 5)
assert(c == 3 and d == 4)
