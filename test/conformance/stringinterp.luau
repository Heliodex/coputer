local x = "world"

print(`Hello, {x}!`)

print(`hello {"world"}`)
print(`Welcome {"to"} {"Luau"}!`)

print(`2 + 2 = {2 + 2}`)

print(`{1} {2} {3} {4} {5} {6} {7}`)

local combo = { "5", "2", "8", "9" }
print(`The lock combinations are: {table.concat(combo, ", ")}`)

print(`true = {true}`)

local name = "Luau"
print(`Welcome to {name}!`)

local nameNotConstantEvaluated = (function() return "Luau" end)()
print(`Welcome to {nameNotConstantEvaluated}!`)

-- should error?
print(`This {localName} does not exist`)
print(`This {nil} does not exist`)

print(`Welcome to \
{name}!`)

print(`empty`)

print(`Escaped brace: \{}`)
print(`Escaped brace: \{} {1}`)
print(`Escaped brace \{} with {"expression"}`)
print(`Backslash \ that escapes the space is not a part of the string...`)
print(`Escaped backslash \\`)
print(`Escaped backslash and braces \\ \{} {1}`)
print(`Escaped backtick: \` \\\` {1}`)

print(`Hello {`from inside {"a nested string"}`}`)

print(`1 {`2 {`3 {4}`}`}`, "1 2 3 4")

local health = 50
print(`You have {health}% health`)

local function shadowsString(string) return `Value is {string}` end

print(shadowsString "hello")
print(shadowsString(1))

print(`\u{0041}\t`, "A\t")
print("A        ", "A   ")
