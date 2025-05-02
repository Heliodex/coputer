local memo: { { number } } = {}

local function findRoutes(x: number, y: number)
	if x == 0 and y == 0 then return 1 end
	if x < 0 or y < 0 then return 0 end
	if memo[x] and memo[x][y] then return memo[x][y] end

	local routes = findRoutes(x - 1, y) + findRoutes(x, y - 1)

	memo[x] = memo[x] or {}
	memo[x][y] = routes

	return routes
end

print(findRoutes(20, 20))
