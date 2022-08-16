package getopt

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/patroclos/go-conq"
	"github.com/patroclos/go-conq/completion"
)

func New() conq.Optioner {
	return &getopt{}
}

type getopt struct{}

func (*getopt) CompleteOptions(ctx completion.Context, opts ...conq.Opter) []string {
	a := ctx.Args
	names := make([]string, 0, len(opts))
	for _, opt := range opts {
		o := opt.Opt()

		for _, name := range strings.Split(o.Name, ",") {
			if len(name) == 1 {
				names = append(names, fmt.Sprintf("-%s", name))
				continue
			}
			names = append(names, fmt.Sprintf("--%s", name))

			lastCompIsFlag := strings.HasPrefix(a.LastCompleted, "--")
			wasAssign := strings.Contains(a.LastCompleted, "=")
			if o.Predict != nil && lastCompIsFlag && !wasAssign {
				// complete using o.Predict?
				names = append(names, o.Predict.Predict(a)...)
			}
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

		if len(arg) == 0 {
			ctx.Args = ctx.Args[1:]
			continue
		}

		for _, opt := range opts {
			o := opt.Opt()
			for _, name := range strings.Split(o.Name, ",") {
				if len(name) == 1 {
					if arg != fmt.Sprintf("-%s", name) {
						continue
					}
					if o.Type.Kind() == reflect.Bool {
						// nice to have: we still might want to swallow true/false/1/0 args following this
						ctx.Args = ctx.Args[1:]
						ctx.Strings[o.Name] = ""
						ctx.Values[o.Name] = true
						continue a
					}

					if len(ctx.Args) == 1 {
						return ctx, fmt.Errorf("missing value for option %q", o.Name)
					}

					targetOpt = &o
					ctx.Args = ctx.Args[1:]
					continue a
				}
			}
		}

		// TODO: iterate over options earlier and check for shorthands
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
				var match string
				for _, name := range strings.Split(o.Name, ",") {
					if name == n {
						match = name
						break
					}
				}
				if match == "" {
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
			var match string
			for _, oname := range strings.Split(o.Name, ",") {
				if name == oname {
					match = oname
					break
				}
			}
			if match == "" {
				continue
			}

			switch o.Type.Kind() {
			case reflect.Bool:
				targetOpt = nil
				ctx.Args = ctx.Args[1:]
				ctx.Strings[o.Name] = ""
				ctx.Values[o.Name] = true
				continue a
			default:
				targetOpt = &o
				ctx.Args = ctx.Args[1:]
				continue a
			}
		}

		return ctx, fmt.Errorf("unrecognized option %q", name)
	}

	if targetOpt != nil {
		return ctx, fmt.Errorf("expecting value for option %s", targetOpt.Name)
	}

	return ctx, nil
}
