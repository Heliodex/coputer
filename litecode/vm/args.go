package vm

import (
	"errors"

	. "github.com/Heliodex/coputer/litecode/types"
)

func args_web(args Args) (r []Val, err error) {
	pargs, ok := args.Co.ProgramArgs.(WebArgs)
	if !ok {
		return nil, errors.New("web args only available in web mode")
	}

	query := make(map[Val]Val, len(pargs.Url.Query))
	for k, vs := range pargs.Url.Query {
		params := make([]Val, len(vs))
		for i, v := range vs {
			params[i] = v
		}

		query[k] = &Table{
			List:     params,
			Readonly: true,
		}
	}

	headers := make(map[Val]Val, len(pargs.Headers))
	for k, v := range pargs.Headers {
		headers[k] = v
	}

	bb := Buffer(pargs.Body)

	webargs := &Table{
		Hash: map[Val]Val{
			"url": &Table{
				Hash: map[Val]Val{
					"rawpath":  pargs.Url.Rawpath,
					"path":     pargs.Url.Path,
					"rawquery": pargs.Url.Rawquery,
					"query": &Table{
						Hash:     query,
						Readonly: true,
					},
				},
				Readonly: true,
			},
			"method": pargs.Method,
			"headers": &Table{
				Hash:     headers,
				Readonly: true,
			},
			"body": &bb,
		},
		Readonly: true,
	}

	return []Val{webargs}, nil
}

var libargs = NewLib([]Function{
	MakeFn("web", args_web),
})
