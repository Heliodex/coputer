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

	fmt.Println("Root:")
	fmt.Println(indentStart(res.Root.String(), 2))
	// fmt.Println("  NodeLoc", res.Root.NodeLoc)
	// fmt.Println("  Body")
	// for _, stat := range res.Root.Body {
	// 	fmt.Printf("    %T at %s\n", stat, stat.GetLocation())
	// }
	// fmt.Println("  HasEnd", res.Root.HasEnd)
	// fmt.Println("  HasSemicolon", res.Root.HasSemicolon)
	fmt.Println("CommentLocations", res.CommentLocations)
	fmt.Println("HotComments", res.HotComments)
	fmt.Println("CstNodeMap", res.CstNodeMap)

}
