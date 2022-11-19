package conq_test

import (
	"fmt"

	"github.com/patroclos/go-conq"
	"github.com/patroclos/go-conq/aid"
	"github.com/patroclos/go-conq/aid/cmdhelp"
	"github.com/patroclos/go-conq/commander"
	"github.com/patroclos/go-conq/getopt"
	"github.com/posener/complete"
)

func ExampleCmd_Opts() {
	cmd := makeCmd()
	ctx := conq.OSContext()
	ctx.Args = []string{"--depth", "500", "testerino"}
	err := commander.New(getopt.New(), nil).Execute(cmd, ctx)
	if err != nil {
		panic(err)
	}
	// Output: Doing something to depth:500 in path:"default"
	// Query: "testerino"
}

func ExampleDefaultHelp() {
	cmd := makeCmd()
	ctx := conq.OSContext()
	ctx.Args = []string{"help"}
	err := commander.New(getopt.New(), aid.DefaultHelp).Execute(cmd, ctx)
	if err != nil {
		panic(err)
	}
	// Output: usage: app [options] query
	//
	// Options:
	// int     depth (required)
	// string  path
	//
	// Arguments:
	// string  query
	//
	// Commands: help
}

func makeCmd() *conq.Cmd {
	var OptDepth = conq.ReqOpt[int]{Name: "depth"}
	var OptPath = conq.Opt[string]{Name: "path", Predict: complete.PredictAnything}
	var ArgQuery = conq.ReqOpt[string]{Name: "query"}

	return &conq.Cmd{
		Name:     "app",
		Opts:     []conq.Opter{OptDepth, OptPath},
		Args:     []conq.Opter{ArgQuery},
		Commands: []*conq.Cmd{cmdhelp.New(nil)},
		Run: func(c conq.Ctx) error {
			depth := OptDepth.Get(c)
			path, err := OptPath.Get(c)
			if err != nil {
				path = "default"
			}

			fmt.Fprintf(c.Out, "Doing something to depth:%d in path:%q\n", depth, path)
			fmt.Fprintf(c.Out, "Query: %q\n", ArgQuery.Get(c))
			return nil
		},
	}
}
