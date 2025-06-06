local g = { "a", "b", "c", "d", "e", [7] = "g" }

print(table.find(g, "c"))
print(table.find(g, "h"))
print(table.find(g, "c", 3))
print(table.find(g, "c", 4))
print(table.find(g, "e"))

-- LLaMe
local h = {test = 5, new = 6}

print(table.find(h, 5))
print(table.find(h, 6))

-- nocte lel
local find1 = { "content", "created_at", "id", "kind", "pubkey", "sig", "tags" }
local find2 = { "id", "pubkey", "created_at", "kind", "tags", "content", "sig" }

for _, v in find1 do
	local found = table.find(find2, v)
	print(found, v)
end
