package vm

import "errors"

func args_web(args Args) (r []Val, err error) {
	pargs, ok := args.Co.programArgs.(WebArgs)
	if !ok {
		return nil, errors.New("web args only available in web mode")
	}

	headers := make(map[Val]Val, len(pargs.Headers))
	for k, v := range pargs.Headers {
		headers[k] = v
	}

	webargs := &Table[Val, Val]{
		Hash: map[Val]Val{
			"url": &Table[Val, Val]{
				Hash: map[Val]Val{
					"rawpath":  pargs.Url.Rawpath,
					"path":     pargs.Url.Path,
					"rawquery": pargs.Url.Rawquery,
					"query":    pargs.Url.Query,
				},
				readonly: true,
			},
			"method": pargs.Method,
			"headers": &Table[Val, Val]{
				Hash:     headers,
				readonly: true,
			},
			"body": Buffer(pargs.Body),
		},
		readonly: true,
	}

	return []Val{webargs}, nil
}

var libargs = NewLib([]Function{
	MakeFn("web", args_web),
})
