package commander

import (
	"fmt"
	"strings"

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

// TODO: put completion into a subcommand, so its entirely optional and can be custom mounted so to speak
func (c Commander) doCompletion(cmd *conq.Cmd, line string, point int, ctype comptype) error {
	if point >= 0 && point < len(line) {
		line = line[:point]
	}

	a := complArgs(line)
	cmd, path := c.ResolveCmd(cmd, a.Completed)
	a = sliceArgs(a, len(path))

	// subcommand completion
	var options []string = c.O.CompleteOptions(a, cmd.Opts...)
	for _, sub := range cmd.Commands {
		options = append(options, sub.Name)
	}

	for _, opt := range options {
		if !strings.HasPrefix(opt, a.Last) {
			continue
		}
		fmt.Println(opt)
	}
	return nil
}

func (c Commander) Execute(root *conq.Cmd, ctx conq.Ctx) error {
	line, point, ctype, ok := completionContext()
	if ok {
		return c.doCompletion(root, line, point, ctype)
	}

	ctx.OptValues = map[string]any{}
	cmd, path := c.ResolveCmd(root, ctx.Args)
	ctx.Args = ctx.Args[len(path):]
	ctx.Path = make([]*conq.Cmd, 1, len(path)+1)
	ctx.Path[0] = root
	ctx.Path = append(ctx.Path, path...)
	ctx, err := c.O.ExtractOptions(ctx, cmd.Opts...)
	if err != nil {
		return fmt.Errorf("failed extracting options: %w", err)
	}

	for _, opt := range cmd.Opts {
		o := opt.Opt()
		if !o.Require {
			continue
		}
		if _, ok := ctx.OptValues[o.Name]; !ok {
			return fmt.Errorf("missing required option %q", o.Name)
		}
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
		ctx.OptValues[o.Name] = val
		ctx.Args = ctx.Args[1:]
	}

	return cmd.Run(ctx)
}

func (c Commander) ResolveCmd(root *conq.Cmd, args []string) (cmd *conq.Cmd, path []*conq.Cmd) {
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
