package getopt

import (
	"fmt"
	"strings"

	"github.com/patroclos/go-conq"
	"github.com/posener/complete"
)

func New() conq.Optioner {
	return &getopt{}
}

type getopt struct{}

func (*getopt) CompleteOptions(a complete.Args, opts ...conq.Opter) []string {
	names := make([]string, 0, len(opts))
	for _, opt := range opts {
		o := opt.Opt()
		flag := fmt.Sprintf("--%s", o.Name)
		names = append(names, flag)

		if o.Predict != nil && strings.HasPrefix(a.LastCompleted, "--") && !strings.Contains(a.LastCompleted, "=") {
			// complete using o.Predict?
			for _, pred := range o.Predict.Predict(a) {
				names = append(names, pred)
			}
		}

		if a.LastCompleted != flag {
		}
	}
	return names
}

func (*getopt) ExtractOptions(ctx conq.Ctx, opts ...conq.Opter) (conq.Ctx, error) {
	ctx.OptValues = make(map[string]any, len(opts))

	var targetOpt *conq.O
a:
	for _, arg := range ctx.Args {
		if targetOpt != nil {
			if targetOpt.Parse != nil {
				val, err := targetOpt.Parse(arg)
				if err != nil {
					return ctx, fmt.Errorf("parsing option %q failed: %w", targetOpt.Name, err)
				}
				ctx.OptValues[targetOpt.Name] = val
			} else {
				ctx.OptValues[targetOpt.Name] = arg
			}
			ctx.Args = ctx.Args[1:]
			targetOpt = nil
			continue
		}
		if !strings.HasPrefix(arg, "--") {
			break
		}

		name := arg[2:]
		if len(name) == 0 {
			ctx.Args = ctx.Args[1:]
			break
		}

		if idx := strings.Index(name, "="); idx != -1 {
			n, v := name[:idx], name[idx+1:]
			ctx.Args = ctx.Args[1:]

			for _, opt := range opts {
				o := opt.Opt()
				if o.Name != n {
					continue
				}
				if o.Parse != nil {
					v2, err := o.Parse(v)
					if err != nil {
						return ctx, fmt.Errorf("parsing option %q failed: %w", n, err)
					}
					ctx.OptValues[n] = v2
					continue a
				} else {
					ctx.OptValues[n] = v
					continue a
				}
			}

			return ctx, fmt.Errorf("unrecognized option %q", n)
		}

		for _, opt := range opts {
			o := opt.Opt()
			if name != o.Name {
				continue
			}
			targetOpt = &o
			ctx.Args = ctx.Args[1:]
			continue a
		}

		return ctx, fmt.Errorf("unrecognized option %q", name)
	}

	if targetOpt != nil {
		return ctx, fmt.Errorf("expecting value for option %s", targetOpt.Name)
	}

	return ctx, nil
}
