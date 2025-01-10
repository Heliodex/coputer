package main

import "errors"

// ngl this might be the easiest library yet (or at least because most of its functionality is coroutines in the main file instead of here)

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

func coroutine_resume(args Args) (Rets, error) {
	co := args.GetCoroutine()
	a := args.args[1:]

	if co.status == Dead {
		return []any{false, "cannot resume dead coroutine"}, nil
	}

	r, err := co.Resume(a...)
	if err != nil {
		return nil, err
	}
	return append([]any{true}, r...), nil
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
	f := args.GetFunction()

	co := createCoroutine(f)
	return Fn(func(_ *Coroutine, args ...any) (Rets, error) {
		if co.status == Dead {
			return nil, errors.New("cannot resume dead coroutine") // ought to be better (return false, error message) if we can figure out how
		}
		return co.Resume(args...)
	})
}

func coroutine_yield(args Args) Rets {
	a, co := args.args, args.co

	if co.status == Running {
		co.status = Suspended
	}
	co.yield <- Yield{rets: a}
	return <-co.resume
}

var libcoroutine = NewTable([][2]any{
	MakeFn0("close", coroutine_close),
	MakeFn1("create", coroutine_create),
	MakeFn1("isyieldable", coroutine_isyieldable),
	MakeFnE("resume", coroutine_resume),
	MakeFn1("running", coroutine_running),
	MakeFn1("status", coroutine_status),
	MakeFn1("wrap", coroutine_wrap),
	MakeFn("yield", coroutine_yield),
})
