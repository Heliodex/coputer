package vm

import "errors"

// ngl this might be the easiest library yet (or at least because most of its functionality is coroutines in the main file instead of here)

func coroutine_close(args Args) (r []Val, err error) {
	co := args.GetCoroutine()

	co.status = CoDead
	return
}

func coroutine_create(args Args) (r []Val, err error) {
	f := args.GetFunction()

	return []Val{createCoroutine(f, args.Co)}, nil
}

func coroutine_isyieldable(args Args) (r []Val, err error) {
	return []Val{true}, nil // phuck yo metamethod/C-call boundary
}

func coroutine_resume(args Args) (r []Val, err error) {
	co := args.GetCoroutine()
	a := args.List[1:]

	if co.status == CoDead {
		return []Val{false, "cannot resume dead coroutine"}, nil
	} else if co.status == CoRunning {
		return []Val{false, "cannot resume running coroutine"}, nil
	}

	// fmt.Println("C.R resuming")
	res, err := co.Resume(a...)
	// fmt.Println("C.R resumed", r)
	if err != nil {
		return
	}
	return append([]Val{true}, res...), nil
}

func coroutine_running(args Args) (r []Val, err error) {
	// fmt.Println("RUNNING")
	return []Val{args.Co}, nil
}

func coroutine_status(args Args) (r []Val, err error) {
	co := args.GetCoroutine()

	switch co.status {
	case CoSuspended:
		return []Val{"suspended"}, nil
	case CoRunning:
		return []Val{"running"}, nil
	case CoNormal:
		return []Val{"normal"}, nil
	}
	return []Val{"dead"}, nil
}

func coroutine_wrap(args Args) (r []Val, err error) {
	f := args.GetFunction()

	co := createCoroutine(f, args.Co)

	return []Val{fn("wrap", func(_ *Coroutine, args ...Val) (r []Val, err error) {
		if co.status == CoDead {
			return nil, errors.New("cannot resume dead coroutine") // ought to be better (return false, error message) if we can figure out how
		} else if co.status == CoRunning {
			return nil, errors.New("cannot resume running coroutine")
		}
		return co.Resume(args...)
	})}, nil
}

func coroutine_yield(args Args) (r []Val, err error) {
	co := args.Co

	if co.status == CoRunning {
		// fmt.Println("C.Y suspending coroutine")
		co.status = CoSuspended
	}

	// fmt.Println("C.Y yielding")
	co.yield <- yield{rets: args.List}
	// fmt.Println("C.Y yielded", "waiting for resume")
	return <-co.resume, nil
	// fmt.Println("C.Y resumed")
}

var libcoroutine = NewLib([]Function{
	MakeFn("close", coroutine_close),
	MakeFn("create", coroutine_create),
	MakeFn("isyieldable", coroutine_isyieldable),
	MakeFn("resume", coroutine_resume),
	MakeFn("running", coroutine_running),
	MakeFn("status", coroutine_status),
	MakeFn("wrap", coroutine_wrap),
	MakeFn("yield", coroutine_yield),
})
