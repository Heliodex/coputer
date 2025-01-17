package litecode

import "errors"

// ngl this might be the easiest library yet (or at least because most of its functionality is coroutines in the main file instead of here)

func coroutine_close(args Args) (r Rets, err error) {
	co := args.GetCoroutine()

	co.status = Dead
	return
}

func coroutine_create(args Args) (r Rets, err error) {
	f := args.GetFunction()

	return Rets{createCoroutine(f)}, nil
}

func coroutine_isyieldable(args Args) (r Rets, err error) {
	return Rets{true}, nil // phuck yo metamethod/C-call boundary
}

func coroutine_resume(args Args) (r Rets, err error) {
	co := args.GetCoroutine()
	a := args.Args[1:]

	if co.status == Dead {
		return Rets{false, "cannot resume dead coroutine"}, nil
	} else if co.status == Running {
		return Rets{false, "cannot resume running coroutine"}, nil
	}

	// fmt.Println("C.R resuming")
	res, err := co.Resume(a...)
	// fmt.Println("C.R resumed", r)
	if err != nil {
		return
	}
	return append([]any{true}, res...), nil
}

func coroutine_running(args Args) (r Rets, err error) {
	// fmt.Println("RUNNING")
	return Rets{args.Co}, nil
}

func coroutine_status(args Args) (r Rets, err error) {
	co := args.GetCoroutine()

	switch co.status {
	case Suspended:
		return Rets{"suspended"}, nil
	case Running:
		return Rets{"running"}, nil
	case Normal:
		return Rets{"normal"}, nil
	}
	return Rets{"dead"}, nil
}

func coroutine_wrap(args Args) (r Rets, err error) {
	f := args.GetFunction()

	co := createCoroutine(f)
	return Rets{Fn(func(_ *Coroutine, args ...any) (r Rets, err error) {
		if co.status == Dead {
			return nil, errors.New("cannot resume dead coroutine") // ought to be better (return false, error message) if we can figure out how
		} else if co.status == Running {
			return nil, errors.New("cannot resume running coroutine")
		}
		return co.Resume(args...)
	})}, nil
}

func coroutine_yield(args Args) (r Rets, err error) {
	a, co := args.Args, args.Co

	if co.status == Running {
		// fmt.Println("C.Y suspending coroutine")
		co.status = Suspended
	}

	// fmt.Println("C.Y yielding", a)
	co.yield <- Yield{rets: a}
	// fmt.Println("C.Y yielded", a, "waiting for resume")
	return <-co.resume, nil
	// fmt.Println("C.Y resumed", r)
}

var libcoroutine = NewTable([][2]any{
	MakeFn("close", coroutine_close),
	MakeFn("create", coroutine_create),
	MakeFn("isyieldable", coroutine_isyieldable),
	MakeFn("resume", coroutine_resume),
	MakeFn("running", coroutine_running),
	MakeFn("status", coroutine_status),
	MakeFn("wrap", coroutine_wrap),
	MakeFn("yield", coroutine_yield),
})
