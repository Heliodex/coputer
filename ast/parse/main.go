package main

import "fmt"

func main() {
	ok, res := Parse(`if true then
elseif false then
end

if true then
else
	if false then
	end
end
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
