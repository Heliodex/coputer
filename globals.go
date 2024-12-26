package main

// p sure that 'globals' is a misnomer here but whatever

func global_type(args Args) Ret {
	obj := args.GetAny()

	t, ok := luautype[typeOf(obj)]
	if !ok {
		return "userdata"
	}
	return t
}
