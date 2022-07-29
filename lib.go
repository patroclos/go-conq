package conq

import (
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/alexflint/go-scalar"
	"github.com/posener/complete"
)

type Cmd struct {
	Name     string
	Commands []*Cmd
	Run      func(Ctx) error
	Help     func(Helper)
	Opts     []Opter
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

// OptionExtractor provides extraction and completion for CLI options (eg. getopt style flags)
// During completion, the OptionExtractor must determin which, if any, option is currently being edited
// so that the commander can grab and use it's Predictor
type OptionExtractor interface {
	ExtractOptions(Ctx, ...Opter) (Ctx, error)
}

func New(o OptionExtractor, h Helper) Commander {
	return Commander{o, h}
}

type Commander struct {
	o OptionExtractor
	h Helper
}

// TODO completion
func (c Commander) Execute(cmd *Cmd, args []string) error {
	ctx := Ctx{
		In:   os.Stdin,
		Out:  os.Stdout,
		Err:  os.Stderr,
		Args: os.Args,
	}
	ctx.OptValues = map[string]any{}
	cmd, ctx.Args = c.resolveCmd(cmd, args)
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

	return cmd.Run(ctx)
}

// TODO
func (c Commander) resolveCmd(cmd *Cmd, args []string) (*Cmd, []string) {
	return cmd, args
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
	return Opt[T](o).Opt()
}

// Helper is the io.Writer of structured help generation
type Helper interface{}
