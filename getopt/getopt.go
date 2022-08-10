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
			names = append(names, o.Predict.Predict(a)...)
		}
	}
	return names
}

func (*getopt) ExtractOptions(ctx conq.Ctx, opts ...conq.Opter) (conq.Ctx, error) {
	ctx.Values = make(map[string]any, len(opts))
	ctx.Strings = make(map[string]string, len(opts))

	var targetOpt *conq.O
a:
	for _, arg := range ctx.Args {
		if targetOpt != nil {
			switch targetOpt.Parse {
			case nil:
				ctx.Values[targetOpt.Name] = arg
				ctx.Strings[targetOpt.Name] = arg
			default:
				val, err := targetOpt.Parse(arg)
				if err != nil {
					return ctx, fmt.Errorf("parsing option %q failed: %w", targetOpt.Name, err)
				}
				ctx.Values[targetOpt.Name] = val
				ctx.Strings[targetOpt.Name] = arg
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
				switch o.Parse {
				case nil:
					ctx.Values[n] = v
					ctx.Strings[n] = v
				default:
					v2, err := o.Parse(v)
					if err != nil {
						return ctx, fmt.Errorf("parsing option %q failed: %w", n, err)
					}
					ctx.Values[n] = v2
					ctx.Strings[n] = v
				}
				continue a
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
