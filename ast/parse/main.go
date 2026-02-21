package main

import "fmt"

func main() {
	ok, res := Parse(`local x = 500
local y = 0x1f4
local z = 0b1_1111_0100
local overflow = 1e999
local n = -42
local noverflow = -1e999
`, Options{})
	if !ok {
		fmt.Println("Parse failed with errors:")
		for _, err := range res.Errors {
			fmt.Printf("- %s at %s\n", err.Message, err.Location)
		}
		return
	}

	fmt.Println(res)
}
