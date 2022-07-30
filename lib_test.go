package conq_test

import (
	"fmt"

	"github.com/patroclos/go-conq"
	"github.com/patroclos/go-conq/getopt"
	"github.com/posener/complete"
)

func ExampleCmd() {
	var OptDepth = conq.ReqOpt[int](conq.O{Name: "depth"})
	var OptPath = conq.Opt[string](conq.O{Name: "path", Predict: complete.PredictAnything})

	cmd := &conq.Cmd{
		Opts: []conq.Opter{OptDepth, OptPath},
		Run: func(c conq.Ctx) error {
			depth := OptDepth.Get(c)
			path, err := OptPath.Get(c)
			if err != nil {
				path = "default"
			}

			fmt.Fprintf(c.Out, "Doing something to depth:%d in path:%q", depth, path)
			return nil
		},
	}

	ctx := conq.OSContext()
	ctx.Args = []string{"--depth", "500"}
	err := conq.New(getopt.New(), nil).Execute(cmd, ctx)
	if err != nil {
		panic(err)
	}
	// Output: Doing something to depth:500 in path:"default"
}
