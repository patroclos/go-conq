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

	if len(sub.Cmd.Commands) > 0 {
		fmt.Fprintf(&b, "Commands: %s", sub.Cmd.Commands[0].Name)
		for _, c := range sub.Cmd.Commands[1:] {
			fmt.Fprintf(&b, ", %s", c.Name)
		}
		b.WriteString("\n")
	}

	if len(sub.Cmd.Opts) == 0 {
		return
	}

	var sorted []conq.Opter
	var required []conq.Opter
	for _, opt := range sub.Cmd.Opts {
		if opt.Opt().Require {
			required = append(required, opt)
			continue
		}
		sorted = append(sorted, opt)
	}
	sorted = append(required, sorted...)
	// required options sorted to top
	b.WriteString("Options:\n")
	var longest int
	for _, opt := range sorted {
		o := opt.Opt()
		if l := len(o.Type.Name()); longest < l {
			longest = l
		}
	}
	format := fmt.Sprintf("%%%dv  ", -longest)
	for _, opt := range sorted {
		o := opt.Opt()
		if o.Type != nil {
			fmt.Fprintf(&b, format, o.Type.Name())
		}
		switch o.Require {
		case true:
			fmt.Fprintf(&b, "%s (required)\n", o.Name)
		case false:
			fmt.Fprintf(&b, "%s\n", o.Name)
		}
	}

	return
}
