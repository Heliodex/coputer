package main

import "fmt"

func string_byte(args Args) (bytes Rets) {
	s := args.GetString()
	i := args.GetNumber(1)
	j := args.GetNumber(i)

	ls := len(s)

	ui, uj := int(i), min(int(j), ls)
	if ui < 0 {
		ui += ls
	} else {
		ui -= 1
	}
	if uj < 0 {
		uj += ls
	} else {
		uj -= 1
	}

	cap := uj - ui + 1
	if cap < 0 {
		return
	}

	bytes = make(Rets, cap)
	chars := []byte(s)
	for i := range bytes {
		bytes[i] = chars[i+ui]
	}
	return
}

func string_char(args Args) Ret {
	l := len(args.args)
	if l == 0 {
		return ""
	}

	bytes := make([]byte, l)
	for i := range bytes {
		a := args.GetNumber()
		if a < 0 || a >= 256 {
			panic(fmt.Sprintf("invalid argument #%d to 'char' (invalid value)", i+1))
		}

		bytes[i] = byte(a)
	}

	return string(bytes)
}

var libstring = NewTable([][2]any{
	MakeFn("byte", string_byte),
	MakeFn1("char", string_char),
})
