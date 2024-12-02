local function fact(n: number) return if n == 0 then 1 else n * fact(n - 1) end

print(fact(10))
