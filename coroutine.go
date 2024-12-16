package main

func coroutine_close(args Args) {
	co := args.GetCoroutine()
	co.status = Dead
}

func coroutine_create(args Args) Ret {
	f := args.GetFunction()

	return createCoroutine(f)
}

func coroutine_isyieldable(args Args) Ret {
	return true // phuck yo metamethod/C-call boundary
}

func coroutine_resume(args Args) Rets {
	co := args.GetCoroutine()
	a := args.args[1:]

	if co.status == Dead {
		return []any{false, "cannot resume dead coroutine"}
	}
	return append([]any{true}, co.Resume(a...)...)
}

func coroutine_running(args Args) Ret {
	return args.co
}

func coroutine_status(args Args) Ret {
	co := args.GetCoroutine()

	switch co.status {
	case Suspended:
		return "suspended"
	case Running:
		return "running"
	case Normal:
		return "normal"
	}
	return "dead"
}

func coroutine_wrap(args Args) Ret {
	panic("not implemented")
}

func coroutine_yield(args Args) Rets {
	a, co := args.args, args.co

	if co.status == Running {
		co.status = Suspended
	}
	co.yield <- a
	return <-co.resume
}

var libcoroutine = NewTable([][2]any{
	MakeFn0("close", coroutine_close),
	MakeFn1("create", coroutine_create),
	MakeFn1("isyieldable", coroutine_isyieldable),
	MakeFn("resume", coroutine_resume),
	MakeFn1("running", coroutine_running),
	MakeFn1("status", coroutine_status),
	MakeFn1("wrap", coroutine_wrap),
	MakeFn("yield", coroutine_yield),
})
