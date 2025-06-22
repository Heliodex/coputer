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

	headers := make(map[Val]Val, len(pargs.Headers))
	for k, v := range pargs.Headers {
		headers[k] = v
	}

	webargs := &Table{
		Hash: map[Val]Val{
			"url": &Table{
				Hash: map[Val]Val{
					"rawpath":  pargs.Url.Rawpath,
					"path":     pargs.Url.Path,
					"rawquery": pargs.Url.Rawquery,
					"query":    pargs.Url.Query,
				},
				Readonly: true,
			},
			"method": pargs.Method,
			"headers": &Table{
				Hash:     headers,
				Readonly: true,
			},
			"body": Buffer(pargs.Body),
		},
		Readonly: true,
	}

	return []Val{webargs}, nil
}

var libargs = NewLib([]Function{
	MakeFn("web", args_web),
})
