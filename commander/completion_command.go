package commander

import (
	"fmt"
	"strings"

	"github.com/patroclos/go-conq"
	"github.com/patroclos/go-conq/completion"
)

var CmdCompletion *conq.Cmd = &conq.Cmd{
	Name: "completion",
	Run: func(c conq.Ctx) error {
		line, point, ctype, ok := completionContext()
		// in completion mode, show install instructions
		if ok {
			return doCompletion(c.Com, c.Path[0], line, point, ctype)
		}

		// show some installation instructions and exit
		var pth strings.Builder
		pth.WriteString(c.Path[0].Name)
		for _, x := range c.Path[1:] {
			fmt.Fprintf(&pth, " %s", x.Name)
		}
		pth.WriteString(" completion")

		fmt.Fprintf(c.Out, "complete -C %q %s\n", pth.String(), c.Path[0].Name)
		return nil
	},
}

// TODO: put completion into a subcommand, so its entirely optional and can be custom mounted so to speak
// TODO: look at cobras custom ctype handline, do we need it aswell? do we want our own customizations?
func doCompletion(com conq.Commander, cmd *conq.Cmd, line string, point int, ctype comptype) error {
	if point >= 0 && point < len(line) {
		line = line[:point]
	}

	a := complArgs(line)

	coco := conq.OSContext()
	coco.Args = a.Completed
	coco = com.ResolveCmd(cmd, coco)

	// subcommand completion
	a = sliceArgs(a, len(coco.Path)-1)
	cc := completion.Context{
		Args: a,
	}
	var options []string = com.Optioner().CompleteOptions(cc, cmd.Opts...)
	if len(options) == 0 {
		for _, sub := range cmd.Commands {
			options = append(options, sub.Name)
		}
	}

	for _, opt := range options {
		if !strings.HasPrefix(opt, a.Last) {
			continue
		}
		// TODO: return values and let CmdCompletion.Run use the context-utilities
		// to print and filter options.
		fmt.Println(opt)
	}
	return nil
}
