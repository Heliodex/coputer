print "epic main function main.luau"

local fn
fn = require "./file2"
fn = require "./file2"

print "world"

local x = 5
local res = fn(x)

print "got res"
print(res)

local dir = require "./dir2"

print(dir())

return function(input: string): string
	return `hello {input} world!`
end
