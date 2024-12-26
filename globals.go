package main

import (
	"iter"
	"strconv"
	"strings"
)

// p sure that 'globals' is a misnomer here but whatever

/* -- fantastic globals and whether to implement them --

loadstring: meh, security and better api should be used
newproxy: meh, not much use without metatables
 pairs: meh, just use generalised iteration
rawequal: meh, not much use without metatables
 rawget: meh, not much use without metatables
 rawlen: meh, not much use without metatables
 rawset: meh, not much use without metatables
- require: yes, but may be difficult
select: meh, this function's kinda stupid
- tonumber: yes, though would be nice to have these in specific types
- tostring: yes, though would be nice to have these in specific types
typeof: meh, not much use without metatables
- _VERSION: yes, probably custom

*/

func ipairs_iter(args Args) Rets {
	a := args.GetTable()
	i := args.GetNumber()

	i += 1
	if v := a.GetArray(uint(i)); v != nil {
		return Rets{i, v}
	}
	return Rets{} // would prefer nil, nil but whateverrrrr
}

func global_ipairs(args Args) Rets {
	a := args.GetTable()

	return Rets{MakeFn("ipairs", ipairs_iter)[1], a, float64(0)}
}

// The call next(t, k), where k is a key of the table t, returns a next key in the table, in an arbitrary order. (It returns also the value associated with that key, as a second return value.) The call next(t, nil) returns a first pair. When there are no more pairs, next returns nil.
func global_next(args Args) (pair Rets) {
	t := args.GetTable()
	fk := args.GetAny()

	if fk == nil {
		next, stop := iter.Pull2(t.Iter())
		defer stop()

		for {
			k, v, ok := next()
			if !ok {
				break
			}
			return Rets{k, v}
		}
	}

	next, stop := iter.Pull2(t.Iter())
	defer stop()

	for {
		if k, _, ok := next(); !ok {
			break
		} else if k == fk {
			k, v, ok := next()
			if !ok {
				break
			}
			return Rets{k, v}
		}
	}

	return Rets{}
}

func global_pairs(args Args) Rets {
	t := args.GetTable()

	return Rets{MakeFn("next", global_next)[1], t}
}

const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func global_tonumber(args Args) Ret {
	value := args.GetAny()
	radix := uint64(args.GetNumber(10))

	str, ok := value.(string)
	if !ok {
		return nil
	} else if radix < 2 || radix > 36 {
		panic("invalid radix")
	}

	if radix == 10 {
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f
		}
	}

	negative := false
	if strings.HasPrefix(str, "-") {
		negative = true
	}

	switch radix {
	case 16:
		if strings.HasPrefix(str, "0x") {
			str = str[2:]
		}
		// case 2:
		// 	if strings.HasPrefix(str, "0b") {
		// 		str = str[2:]
		// 	}
	}

	str = strings.ToUpper(str)

	radixChars := chars[:radix]

	var n uint64
	for _, c := range str {
		n *= radix
		index := strings.IndexRune(radixChars, c)
		if index == -1 {
			return nil
		}
		n += uint64(index)
	}

	if negative {
		return float64(-n)
	}
	return float64(n)
}

func global_tostring(args Args) Ret {
	// value := args.GetAny()

	panic("not implemented")
}

func global_type(args Args) Ret {
	obj := args.GetAny()

	t, ok := luautype[typeOf(obj)]
	if !ok {
		return "userdata"
	}
	return t
}
