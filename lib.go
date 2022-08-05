package conq

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/alexflint/go-scalar"
	"github.com/posener/complete"
)

type Cmd struct {
	Name     string
	Commands []*Cmd
	Run      func(Ctx) error
	Help     func(Helper)
	Opts     []Opter
	Args     []Opter
}

type Ctx struct {
	In        io.Reader
	Out, Err  io.Writer
	Args      []string
	OptValues map[string]any
}

type O struct {
	Name    string
	Predict complete.Predictor
	Require bool
	Parse   func(string) (interface{}, error)
}

type Opter interface {
	Opt() O
}

// Optioner provides extraction and completion of CLI options.
type Optioner interface {
	ExtractOptions(Ctx, ...Opter) (Ctx, error)
	CompleteOptions(complete.Args, ...Opter) []string
}

type Opt[T any] O

func (o Opt[T]) Get(c Ctx) (val T, err error) {
	x, ok := c.OptValues[o.Name]
	if !ok {
		return val, fmt.Errorf("option %q not present", o.Name)
	}
	val, ok = x.(T)
	if !ok {
		return val, fmt.Errorf("value for option %q is of type %T, expected %T", o.Name, x, val)
	}
	return
}

func (o Opt[T]) Opt() O {
	if o.Parse == nil {
		o.Parse = func(s string) (interface{}, error) {
			if !scalar.CanParse(reflect.TypeOf((*T)(nil)).Elem()) {
				return nil, fmt.Errorf("cannot automatically parse non-scalar value into %q option", o.Name)
			}
			var x T
			err := scalar.Parse(&x, s)
			return x, err
		}
	}
	return O(o)
}

type ReqOpt[T any] O

func (o ReqOpt[T]) Get(c Ctx) (val T) {
	val, err := Opt[T](o.Opt()).Get(c)
	if err != nil {
		panic(fmt.Sprintf("unexpected error accessing required option, this is likely a bug in the commander: %v", err))
	}
	return val
}

func (o ReqOpt[T]) Opt() O {
	x := Opt[T](o).Opt()
	x.Require = true
	return x
}

func New(o Optioner, h Helper) Commander {
	return Commander{o, h}
}

type Commander struct {
	o Optioner
	h Helper
}

func (c Commander) doCompletion(cmd *Cmd, line string, point int) error {
	if point >= 0 && point < len(line) {
		line = line[:point]
	}

	a := complArgs(line)
	cmd, path := c.resolveCmd(cmd, a.Completed)
	a = sliceArgs(a, len(path))

	// subcommand completion
	var options []string = c.o.CompleteOptions(a, cmd.Opts...)
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

func (c Commander) Execute(cmd *Cmd, ctx Ctx) error {
	line, point, ok := completionContext()
	if ok {
		return c.doCompletion(cmd, line, point)
	}

	ctx.OptValues = map[string]any{}
	cmd, path := c.resolveCmd(cmd, ctx.Args)
	ctx.Args = ctx.Args[len(path):]
	ctx, err := c.o.ExtractOptions(ctx, cmd.Opts...)
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

func (c Commander) resolveCmd(root *Cmd, args []string) (cmd *Cmd, path []*Cmd) {
	cmd = root
a:
	if len(args) == 0 {
		return
	}
	for _, x := range cmd.Commands {
		if args[len(path)] != x.Name {
			continue
		}
		path = append(path, cmd)
		cmd = x
		args = args[1:]
		goto a
	}
	return cmd, path
}
