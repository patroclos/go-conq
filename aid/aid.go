package aid

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/patroclos/go-conq"
)

var DefaultHelp conq.Helper = basicHelper{}

type basicHelper struct{}

func (basicHelper) Help(sub conq.HelpSubject) (help string) {
	var b strings.Builder
	defer func() {
		help = b.String()
	}()

	fmt.Fprintf(&b, "usage: %s", sub.Cmd.Name)
	if len(sub.Cmd.Opts) > 0 {
		fmt.Fprint(&b, " [options]")
	}
	for _, arg := range sub.Cmd.Args {
		o := arg.Opt()
		switch o.Require {
		case true:
			fmt.Fprintf(&b, " %s", o.Name)
		case false:
			fmt.Fprintf(&b, " [%s]", o.Name)
		}
	}
	b.WriteString("\n\n")

	headlineStyle := color.New(color.Bold, color.Underline)

	if len(sub.Cmd.Opts) > 0 {
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
		headlineStyle.Fprint(&b, "Options:\n")
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
	}

	if len(sub.Cmd.Args) > 0 {

		headlineStyle.Fprint(&b, "\nArguments:\n")
		var longest int
		for _, arg := range sub.Cmd.Args {
			if l := len(arg.Opt().Name); l > longest {
				longest = l
			}
		}
		format := fmt.Sprintf("%%%dv  ", -longest)
		for _, arg := range sub.Cmd.Args {
			o := arg.Opt()
			if o.Type != nil {
				fmt.Fprintf(&b, format, o.Type.Name())
			}
			switch o.Require {
			case true:
				fmt.Fprintf(&b, "%s\n", o.Name)
			case false:
				fmt.Fprintf(&b, "%s (optional)\n", o.Name)
			}
		}
	}

	if len(sub.Cmd.Commands) > 0 {
		headlineStyle.Fprint(&b, "\nCommands")
		fmt.Fprintf(&b, ": %s", sub.Cmd.Commands[0].Name)
		for _, c := range sub.Cmd.Commands[1:] {
			fmt.Fprintf(&b, ", %s", c.Name)
		}
		b.WriteString("\n")
	}

	if len(sub.Cmd.Env) > 0 {
		fmt.Fprintf(&b, "\nEnvironment Variables:\n")
		var longest int
		for _, arg := range sub.Cmd.Env {
			if l := len(arg.Opt().Name); l > longest {
				longest = l
			}
		}
		format := fmt.Sprintf("%%%dv  ", -longest)
		for _, arg := range sub.Cmd.Env {
			o := arg.Opt()
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
	}

	return
}
