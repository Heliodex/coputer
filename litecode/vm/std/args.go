package std

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
			params[i] = Str(v)
		}

		query[Str(k)] = TableVal(&Table{
			List:     params,
			Readonly: true,
		})
	}

	headers := make(map[Val]Val, len(pargs.Headers))
	for k, v := range pargs.Headers {
		headers[Str(k)] = Str(v)
	}

	bb := Buffer(pargs.Body)

	webargs := &Table{
		Hash: map[Val]Val{
			Str("url"): TableVal(&Table{
				Hash: map[Val]Val{
					Str("rawpath"):  Str(pargs.Url.Rawpath),
					Str("path"):     Str(pargs.Url.Path),
					Str("rawquery"): Str(pargs.Url.Rawquery),
					Str("query"): TableVal(&Table{
						Hash:     query,
						Readonly: true,
					}),
				},
				Readonly: true,
			}),
			Str("method"): Str(pargs.Method),
			Str("headers"): TableVal(&Table{
				Hash:     headers,
				Readonly: true,
			}),
			Str("body"): BufVal(&bb),
		},
		Readonly: true,
	}

	return []Val{TableVal(webargs)}, nil
}

var Libargs = NewLib([]Function{
	MakeFn("web", args_web),
})
