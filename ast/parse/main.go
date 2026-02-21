package main

import "fmt"

const src = `local x = 'this is a "test"'
local y = [[this is a "test"]]
local z = ` + "`this 'is' a \"test\"`" + `

do
	local x = "this is\n]]a test"
	local y = [[this is
a test]]
	local z = ` + "`this is\\na test`" + `
end

local esc = "\a\b\f\n\r\t\v"
`

func main() {
	ok, res := Parse(src, Options{})
	if !ok {
		fmt.Println("Parse failed with errors:")
		for _, err := range res.Errors {
			fmt.Printf("- %s at %s\n", err.Message, err.Location)
		}
		return
	}

	fmt.Println(res)
}
