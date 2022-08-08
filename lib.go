package conq

import (
	"fmt"
	"io"
	"reflect"

	"github.com/alexflint/go-scalar"
	"github.com/posener/complete"
)

type Cmd struct {
	Name     string
	Commands []*Cmd
	Run      func(Ctx) error
	Opts     []Opter
	Args     []Opter
}

type Ctx struct {
	In        io.Reader
	Out, Err  io.Writer
	Args      []string
	OptValues map[string]any
	Path      []*Cmd
}

type O struct {
	Name    string
	Predict complete.Predictor
	Require bool
	Parse   func(string) (interface{}, error)
	Type    reflect.Type
}

// This interface exists to facilitate the Opt[T] and ReqOpt[T] types with filter effects
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
	o.Type = reflect.TypeOf((*T)(nil)).Elem()
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
