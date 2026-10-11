-- Export value syntax (`export local`, `export function`, `export const`)
export local a = 1
export const b = 2
export function f(x)
	return x
end
export local c, d = 3, 4

print(a)
print(b)
print(f(5))
print(c, d)
