package vm

import (
	"errors"

	"github.com/Heliodex/coputer/litecode/types"
)

func args_web(args Args) (r []types.Val, err error) {
	pargs, ok := args.Co.programArgs.(WebArgs)
	if !ok {
		return nil, errors.New("web args only available in web mode")
	}

	headers := make(map[types.Val]types.Val, len(pargs.Headers))
	for k, v := range pargs.Headers {
		headers[k] = v
	}

	webargs := &Table{
		Hash: map[types.Val]types.Val{
			"url": &Table{
				Hash: map[types.Val]types.Val{
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
			"body": types.Buffer(pargs.Body),
		},
		readonly: true,
	}

	return []types.Val{webargs}, nil
}

var libargs = NewLib([]types.Function[*Coroutine]{
	MakeFn("web", args_web),
})
