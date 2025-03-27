package vm

import "errors"

func args_web(args Args) (r Rets, err error) {
	pargs, ok := args.Co.programArgs.(WebArgs)
	if !ok {
		return nil, errors.New("web args only available in web mode")
	}

	headers := make(map[any]any, len(pargs.Headers))
	for k, v := range pargs.Headers {
		headers[k] = v
	}

	webargs := &Table{
		Hash: map[any]any{
			"url": &Table{
				Hash: map[any]any{
					"rawpath":  pargs.Url.Rawpath,
					"path":     pargs.Url.Path,
					"rawquery": pargs.Url.Rawquery,
					"query":    pargs.Url.Query,
				},
				readonly: true,
			},
			"method": pargs.Method,
			"headers": &Table{
				Hash:     headers,
				readonly: true,
			},
			"body": Buffer(pargs.Body),
		},
		readonly: true,
	}

	return Rets{webargs}, nil
}

var libargs = NewLib([]Function{
	MakeFn("web", args_web),
})
