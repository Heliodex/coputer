package main

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Heliodex/coputer/bundle"
	. "github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/litecode/vm"
)

func startWeb(v any) (rets WebRets, err error) {
	t, ok := v.(*Table)
	if !ok {
		fmt.Println("no table", v, vm.TypeOf(v))
		return WebRets{}, errors.New("web program did not return a table")
	}

	// ahh, error checking beauty

	var tstatuscode float64
	if statuscode := t.GetHash("statuscode"); statuscode == nil {
		tstatuscode = 200
	} else if tstatuscode, ok = statuscode.(float64); !ok {
		return WebRets{}, errors.New("return statuscode, if provided, must be a number")
	}
	rets.StatusCode = int(tstatuscode)
	if rets.StatusCode < 100 || rets.StatusCode > 599 {
		return WebRets{}, errors.New("return statuscode, if provided, must be between 100 and 599")
	}

	statusmesage := t.GetHash("statusmessage")
	if statusmesage == nil {
		// default status message for the given status code
		rets.StatusMessage = http.StatusText(rets.StatusCode)
		if rets.StatusMessage == "" {
			rets.StatusMessage = fmt.Sprintf("Error %d", rets.StatusCode)
		}
	} else if rets.StatusMessage, ok = statusmesage.(string); !ok {
		return WebRets{}, errors.New("return statusmessage, if provided, must be a string")
	}

	if headers := t.GetHash("headers"); headers != nil {
		theaders, ok := headers.(*Table)
		if !ok {
			return WebRets{}, errors.New("return headers, if provided, must be a table")
		}
		rets.Headers = make(map[string]string, len(theaders.Hash))

		// we don't have to care about the list content, but we will here
		for k, v := range theaders.Iter() {
			if tk, ok := k.(string); !ok {
				return WebRets{}, errors.New("header keys must be strings")
			} else if tv, ok := v.(string); !ok {
				return WebRets{}, errors.New("header values must be strings")
			} else {
				rets.Headers[tk] = tv
			}
		}
	}

	if body := t.GetHash("body"); body != nil {
		tbody, ok := body.(*Buffer)
		if !ok {
			return WebRets{}, errors.New("return body, if provided, must be a buffer")
		}
		rets.Body = *tbody
	} else { // default to status message if no body
		if rets.Headers == nil {
			rets.Headers = make(map[string]string, 1)
		}
		rets.Headers["content-type"] = "text/plain; charset=utf-8"
		rets.Body = []byte(rets.StatusMessage)
	}
	return
}

func Start(c Compiler, hash string, args ProgramArgs) (output ProgramRets, err error) {
	p, err := vm.Compile(c, filepath.Join(bundle.ProgramsDir, hash, bundle.Entrypoint))
	if err != nil {
		return
	}

	luau_print := vm.MakeFn("print", func(args vm.Args) (r []Val, err error) {
		for _, arg := range args.List {
			fmt.Print("\t")
			fmt.Print(vm.ToString(arg))
		}
		fmt.Println() // yeah
		return
	})

	var env Env
	env.AddFn(luau_print)

	co, cancel := p.Load(env, args)

	go func() {
		time.Sleep(5 * time.Second)
		cancel()
	}()

	r, err := co.Resume()
	if err != nil {
		return
	} else if len(r) != 1 {
		return nil, errors.New("program did not return a single value")
	}

	switch ret := r[0]; args.Type() {
	case TestProgramType:
		return nil, errors.New("test program type not supported in this context")
	case WebProgramType:
		return startWeb(ret)
	default:
		return nil, errors.New("unknown program type")
	}
}
