package commander

import (
	"fmt"
	"os"
	"strings"

	"github.com/patroclos/go-conq"
	"github.com/patroclos/go-conq/aid"
	"golang.org/x/text/message"
)

type Commander struct {
	O conq.Optioner
	H aid.Helper
	P *message.Printer
}

func New(o conq.Optioner, h aid.Helper) Commander {
	return Commander{
		O: o,
		H: h,
	}
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
	ctx.Com = c
	ctx = c.ResolveCmd(root, ctx)

	cmd := ctx.Path[len(ctx.Path)-1]
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
		envTxt, ok := os.LookupEnv(o.Name)
		if !ok {
			if o.Require {
				return fmt.Errorf("missing required environment-variable: %q", o.Name)
			}
			continue
		}
		if o.Parse == nil {
			ctx.Strings[o.Name] = envTxt
			continue
		}
		val, err := o.Parse(envTxt)
		if err != nil {
			return fmt.Errorf("failed parsing environment variable %s: %w", o.Name, err)
		}

		ctx.Strings[o.Name] = envTxt
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

	if cmd.Run == nil {
		var pth strings.Builder
		pth.WriteString(ctx.Path[0].Name)
		for _, x := range ctx.Path[1:] {
			fmt.Fprintf(&pth, " %s", x.Name)
		}
		return fmt.Errorf("would've run %q, but no Run function defined", pth.String())
	}

	return cmd.Run(ctx)
}

// Path should always include the root command and the leaf-command that's being executed
func (c Commander) ResolveCmd(root *conq.Cmd, ctx conq.Ctx) (oc conq.Ctx) {
	oc = ctx
	cmd := root
	oc.Path = conq.Pth{cmd}

a:
	if len(oc.Args) == 0 {
		return
	}
	for _, x := range cmd.Commands {
		if oc.Args[0] != x.Name {
			continue
		}
		oc.Path = append(oc.Path, x)
		cmd = x
		oc.Args = oc.Args[1:]
		goto a
	}
	return oc
}
