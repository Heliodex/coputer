-- big strings
local a = string.rep("a", 1e7)
print(string.find(a, "^a*.?$"))
print(string.find(a, "^a*.?b$"))
print(string.find(a, "^a-.?$"))
