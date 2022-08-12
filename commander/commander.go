package commander

import (
	"fmt"
	"os"

	"github.com/patroclos/go-conq"
	"github.com/patroclos/go-conq/aid"
)

type Commander struct {
	O conq.Optioner
	H aid.Helper
}

func New(o conq.Optioner, h aid.Helper) Commander {
	return Commander{o, h}
}

func (c Commander) Optioner() conq.Optioner {
	return c.O
}

func (c Commander) Helper() aid.Helper {
	return c.H
}

func (c Commander) Execute(root *conq.Cmd, ctx conq.Ctx) error {
	ctx.Values = nil
	ctx.Strings = nil
	cmd, path := c.ResolveCmd(root, ctx.Args)
	ctx.Args = ctx.Args[len(path):]
	ctx.Path = path
	ctx.Com = c
	ctx, err := c.O.ExtractOptions(ctx, cmd.Opts...)
	if err != nil {
		return fmt.Errorf("failed extracting options: %w", err)
	}

	for _, opt := range cmd.Opts {
		o := opt.Opt()
		if !o.Require {
			continue
		}
		if _, ok := ctx.Values[o.Name]; !ok {
			return fmt.Errorf("missing required option %q", o.Name)
		}
	}

	for _, opt := range cmd.Env {
		o := opt.Opt()
		val, ok := os.LookupEnv(o.Name)
		if !ok {
			if o.Require {
				return fmt.Errorf("missing required environment-variable: %q", o.Name)
			}
			continue
		}
		ctx.Values[o.Name] = val
	}

	for i, arg := range cmd.Args {
		o := arg.Opt()
		if len(ctx.Args) == 0 {
			if o.Require {
				return fmt.Errorf("missing required positional argument at position %d %q", i+1, o.Name)
			}
			break
		}

		val, err := o.Parse(ctx.Args[0])
		if err != nil {
			return fmt.Errorf("failed parsing argument %d %q: %w", i+1, o.Name, err)
		}
		ctx.Values[o.Name] = val
		ctx.Args = ctx.Args[1:]
	}

	return cmd.Run(ctx)
}

func (c Commander) ResolveCmd(root *conq.Cmd, args []string) (cmd *conq.Cmd, path conq.Pth) {
	cmd = root
a:
	if len(args) == 0 {
		return
	}
	for _, x := range cmd.Commands {
		if args[0] != x.Name {
			continue
		}
		path = append(path, cmd)
		cmd = x
		args = args[1:]
		goto a
	}
	return cmd, path
}
