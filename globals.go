package main

import "iter"

// p sure that 'globals' is a misnomer here but whatever

/* -- fantastic globals and whether to implement them --

- ipairs: yes
loadstring: meh, security and better api should be used
- next: yes
newproxy: meh, not much use without metatables
 pairs: meh, just use generalised iteration
- print: yes, though where it logs to might be different
rawequal: meh, not much use without metatables
 rawget: meh, not much use without metatables
 rawlen: meh, not much use without metatables
 rawset: meh, not much use without metatables
- require: yes, but may be difficult
select: meh, this function's kinda stupid
- tonumber: yes, though would be nice to have these in specific types
- tostring: yes, though would be nice to have these in specific types
- type: yes
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

	iter := Function(func(co *Coroutine, vargs ...any) []any {
		return ipairs_iter(Args{vargs, "ipairs", co, 0})
	})
	return Rets{&iter, a, float64(0)}
}

// The call next(t, k), where k is a key of the table t, returns a next key in the table, in an arbitrary order. (It returns also the value associated with that key, as a second return value.) The call next(t, nil) returns a first pair. When there are no more pairs, next returns nil.
func global_next(args Args) (pair Rets) {
	t := args.GetTable()
	fk := args.GetAny()

	if fk == nil {
		// for k, v := range t.Iter() {
		// 	return Rets{k, v}
		// }
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

	// for k, v := range t.Iter() {
	// 	if next {
	// 		return Rets{k, v}
	// 	} else if k == fk {
	// 		next = true
	// 	}
	// }
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

	return Rets{global_next, t}
}

func global_tonumber(args Args) Ret {
	panic("not implemented")
}

func global_tostring(args Args) Ret {
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
