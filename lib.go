package conq

import (
	"fmt"
	"io"
	"reflect"

	"github.com/alexflint/go-scalar"
	"github.com/patroclos/go-conq/completion"
	"github.com/posener/complete"
	"golang.org/x/text/message"
)

// Cmd is a node in a rooted node tree describing a command-hierarchy.
type Cmd struct {
	Name     string
	Commands []*Cmd
	Run      func(Ctx) error
	Opts     Opts
	Args     Opts
	Env      Opts
	Version  string
}

type Pth []*Cmd

// Ctx is the context in which a command (Cmd) runs.  It contains the std-streams,
// arguments (options extracted before calling Cmd.Run), option-values, the path
// within the command-tree this invocation is located in and a locale-aware
// message-printer.
type Ctx struct {
	In       io.Reader
	Out, Err io.Writer
	Args     []string
	Values   map[string]any
	Strings  map[string]string
	Printer  *message.Printer
	Path     Pth
	Com      Commander
}

type Commander interface {
	ResolveCmd(root *Cmd, ctx Ctx) Ctx
	Execute(root *Cmd, ctx Ctx) error
	Optioner() Optioner
}

// Optioner provides extraction and completion of CLI options.
type Optioner interface {
	ExtractOptions(Ctx, ...Opter) (Ctx, error)
	CompleteOptions(completion.Context, ...Opter) []string
}

// This interface exists to facilitate the Opt[T] and ReqOpt[T] types with filter effects
type Opter interface {
	Opt() O
}

// O is a descriptor for command-parameters (options, positional arguments or environment variables)
type O struct {
	// a comma-separated list of at least one name followed by aliases
	Name string
	// should invoking a command fail, if this option isn't set?
	Require bool
	// a parser for the string extracted from the shell arguments
	Parse func(string) (interface{}, error)
	// describes the type of results returned by Parse
	Type reflect.Type
	// shell-completion
	Predict complete.Predictor
}

func (o O) WithName(name string) O {
	o.Name = name
	return o
}

// Opt[T any] wraps a base-option (usually only containing a name) in an Opter
// interface, which will apply defaults to O.Parse and O.Type values.
// The default O.Parse implementation will use the github.com/alexflint/go-scalar
// package to parse a (value T) from a string.  The default implementations
// supports the encoding.TextUnmarshaler interface.
// Opt[T] is meant both as the definition for the option and as the access-hatch
// for it's values, so it provides `Get(Ctx)(T,error)` and `Getp(Ctx)(*T, error)`
// to access the options value or pointer to it from the Ctx.Values map.
type Opt[T any] O

type Opts []Opter

func (o Opt[T]) Get(c Ctx) (val T, err error) {
	x, ok := c.Values[o.Name]
	if !ok {
		return val, fmt.Errorf("option %q not present", o.Name)
	}
	val, ok = x.(T)
	if !ok {
		return val, fmt.Errorf("value for option %q is of type %T, expected %T", o.Name, x, val)
	}
	return
}

func (o Opt[T]) Getp(c Ctx) (val *T, err error) {
	x, ok := c.Values[o.Name]
	if !ok {
		return val, fmt.Errorf("option %q not present", o.Name)
	}
	v, ok := x.(T)
	if !ok {
		return val, fmt.Errorf("value for option %q is of type %T, expected %T", o.Name, x, val)
	}
	return &v, nil
}

func (o Opt[T]) Opt() O {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	if o.Parse == nil {
		o.Parse = func(s string) (interface{}, error) {
			if !scalar.CanParse(typ) {
				return nil, fmt.Errorf("cannot automatically parse non-scalar value into %q option", o.Name)
			}
			var val T
			err := scalar.Parse(&val, s)
			return val, err
		}
	}
	o.Type = typ
	return O(o)
}

// ReqOpt[T any] is a simple wrapper for Opt[T].  It's Opter implementation sets
// the O.Require to true.  Assuming the Cmd has been setup correctly we can now
// know for sure that the Ctx is going to have a value for this option.  That's
// why the `Get(Ctx)T` and `Getp(Ctx)*T` have been simplified from their counterparts
// in Opt[T]
type ReqOpt[T any] O

func (o ReqOpt[T]) Get(c Ctx) (val T) {
	val, err := Opt[T](o.Opt()).Get(c)
	if err != nil {
		panic(fmt.Sprintf("unexpected error accessing required option, this is likely a bug in the commander: %v", err))
	}
	return val
}

func (o ReqOpt[T]) Getp(c Ctx) (val *T) {
	val, err := Opt[T](o.Opt()).Getp(c)
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
