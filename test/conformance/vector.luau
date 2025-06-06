local v = vector.create(1, 2, 3)

print(v)
print(v * 2)
print(2 * v)
print(2 / v)
print(v / 2)
print(2 // v)
print(v // 2)
print(v + v)
print(v - v)
print(v * v)
print(v / v)
print(v // v)

print(v.x, v.y, v.z)
print(v["x"], v["y"], v["z"])

local x, y, z = "x", "y", "z"
print(v[x], v[y], v[z]) -- yeet

print(v.X, v.Y, v.Z) -- graah

local u = vector.create(1.23, -4.56, 7.89)
local t = vector.create(0.12, 1.45, 2.78)

-- print(u.x, u.y, u.z) --  i remember when
-- print(t.x, t.y, t.z) -- i remember i remember when i lost my mind

print(vector.abs(u))
-- print(vector.angle(v, u))
-- print(vector.angle(v, u, t))
print(vector.ceil(u))
print(vector.floor(u))
print(vector.clamp(u, t, v))
print(vector.cross(v, u), vector.cross(v, t), vector.cross(u, t))
print(vector.dot(v, u), vector.dot(v, t), vector.dot(u, t))
print(vector.magnitude(u))
print(vector.max(v, u, t))
print(vector.min(v, u, t))
print(vector.normalize(v), vector.normalize(u), vector.normalize(t))
print(vector.sign(v), vector.sign(u), vector.sign(t))
print(vector.one)
print(vector.zero)
print(vector.angle(u, t, u))
print(vector.angle(u, t, t))

local function fma(a: vector, b: vector, c: vector) return a * b + c end

local function approx(a: vector): vector
	local r = vector.create(1, 1, 1)
	local aa = a
	for _ = 1, 6 do
		r += aa * 0.123
		aa *= a
	end
	return r
end

local A = vector.create(1, 2, 3)
local B = vector.create(4, 5, 6)
local C = vector.create(7, 8, 9)

print(approx(A), fma(A, B, C))
