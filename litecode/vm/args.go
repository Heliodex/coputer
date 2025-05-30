package vm

import (
	"errors"

	"github.com/Heliodex/coputer/litecode/types"
)

func args_web(args Args) (r []types.Val, err error) {
	pargs, ok := args.Co.ProgramArgs.(types.WebArgs)
	if !ok {
		return nil, errors.New("web args only available in web mode")
	}

	headers := make(map[types.Val]types.Val, len(pargs.Headers))
	for k, v := range pargs.Headers {
		headers[k] = v
	}

	webargs := &types.Table{
		Hash: map[types.Val]types.Val{
			"url": &types.Table{
				Hash: map[types.Val]types.Val{
					"rawpath":  pargs.Url.Rawpath,
					"path":     pargs.Url.Path,
					"rawquery": pargs.Url.Rawquery,
					"query":    pargs.Url.Query,
				},
				Readonly: true,
			},
			"method": pargs.Method,
			"headers": &types.Table{
				Hash:     headers,
				Readonly: true,
			},
			"body": types.Buffer(pargs.Body),
		},
		Readonly: true,
	}

	return []types.Val{webargs}, nil
}

var libargs = NewLib([]types.Function{
	MakeFn("web", args_web),
})
