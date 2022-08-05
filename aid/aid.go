package aid

import (
	"fmt"
	"strings"

	"github.com/patroclos/go-conq"
)

var DefaultHelp Helper = basicHelper{}

type HelpSubject struct {
	Cmd *conq.Cmd
	Opt *conq.O
	Ctx *conq.Ctx
}

type Helper interface {
	Help(HelpSubject) string
}

type basicHelper struct{}

func (basicHelper) Help(sub HelpSubject) (help string) {
	var b strings.Builder
	defer func() {
		help = b.String()
	}()

	fmt.Fprintf(&b, "usage: %s [options]", sub.Cmd.Name)
	for _, arg := range sub.Cmd.Args {
		o := arg.Opt()
		switch o.Require {
		case true:
			fmt.Fprintf(&b, " %s", o.Name)
		case false:
			fmt.Fprintf(&b, " [%s]", o.Name)
		}
	}
	b.WriteString("\n")

	if len(sub.Cmd.Opts) == 0 {
		return
	}

	b.WriteString("Options:\n")
	for _, opt := range sub.Cmd.Opts {
		o := opt.Opt()
		switch o.Require {
		case true:
			fmt.Fprintf(&b, "%s (required)\n", o.Name)
		case false:
			fmt.Fprintf(&b, "%s\n", o.Name)
		}
	}

	return
}

func NewHelp() *conq.Cmd {
	return &conq.Cmd{
		Name: "help",
		Run: func(c conq.Ctx) error {
			subj := HelpSubject{Cmd: c.Path[0]}
		a:
			for len(c.Args) > 0 {
				for _, cmd := range subj.Cmd.Commands {
					if cmd.Name != c.Args[0] {
						continue
					}
					subj.Cmd = cmd
					c.Args = c.Args[1:]
					continue a
				}

				return fmt.Errorf("attempted to resolve unknown command %q on %s", c.Args[0], subj.Cmd.Name)
			}

			h := DefaultHelp.Help(subj)
			fmt.Fprintf(c.Out, "%s\n", h)
			return nil
		},
	}
}
