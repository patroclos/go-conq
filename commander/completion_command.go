package commander

import (
	"fmt"
	"strings"

	"github.com/patroclos/go-conq"
)

var CmdCompletion *conq.Cmd = &conq.Cmd{
	Name: "completion",
	Run: func(c conq.Ctx) error {
		line, point, ctype, ok := completionContext()
		// not in completion mode, show install instructions
		if !ok {
			var pth []string
			for _, x := range c.Path {
				pth = append(pth, x.Name)
			}

			fmt.Fprintf(c.Out, "complete -C %+v\n", c.Path)
			return nil
		}
		return doCompletion(c.Com, c.Path[0], line, point, ctype)
	},
}

// TODO: put completion into a subcommand, so its entirely optional and can be custom mounted so to speak
// TODO: look at cobras custom ctype handline, do we need it aswell? do we want our own customizations?
func doCompletion(com conq.Commander, cmd *conq.Cmd, line string, point int, ctype comptype) error {
	if point >= 0 && point < len(line) {
		line = line[:point]
	}

	a := complArgs(line)
	cmd, path := com.ResolveCmd(cmd, a.Completed)
	a = sliceArgs(a, len(path))

	// subcommand completion
	var options []string = com.Optioner().CompleteOptions(a, cmd.Opts...)
	if len(options) == 0 {
		for _, sub := range cmd.Commands {
			options = append(options, sub.Name)
		}
	}

	for _, opt := range options {
		if !strings.HasPrefix(opt, a.Last) {
			continue
		}
		fmt.Println(opt)
	}
	return nil
}
