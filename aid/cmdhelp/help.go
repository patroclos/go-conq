package cmdhelp

import (
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/patroclos/go-conq"
	"github.com/patroclos/go-conq/aid"
)

func New(helpdir *embed.FS) *conq.Cmd {
	return &conq.Cmd{
		Name: "help",
		Run: func(c conq.Ctx) error {
			subj := aid.HelpSubject{Cmd: c.Path[0]}
			if helpdir != nil {
				if err := printSection(*helpdir, c); err == nil {
					return nil
				}
			}
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

			if hl, ok := c.Com.(interface{ Helper() aid.Helper }); ok {
				fmt.Fprintf(c.Out, "%s\n", hl.Helper().Help(subj))
				return nil
			}
			return fmt.Errorf("no helper configured con commander")
		},
	}
}

func printSection(fs embed.FS, c conq.Ctx) error {
	tmpl, err := template.ParseFS(fs, "help/*.tmpl")
	if err != nil {
		fmt.Fprintf(c.Err, "%v\n", err)
		return fmt.Errorf("failed parsing help-templates: %w", err)
	}

	path := fmt.Sprintf("%s.tmpl", c.Path[0].Name)
	if len(c.Args) > 0 {
		var pth strings.Builder
		pth.WriteString(c.Args[0])
		for _, cmd := range c.Args[1:] {
			fmt.Fprintf(&pth, "/%s", cmd)
		}
		pth.WriteString(".tmpl")
		path = pth.String()
	}
	err = tmpl.ExecuteTemplate(c.Out, path, HelpContext(c))
	return err
}

type HelpContext conq.Ctx

func (c HelpContext) Root() *conq.Cmd {
	return c.Path[0]
}

func (c HelpContext) Cmd() *conq.Cmd {
	return c.Path[len(c.Path)-1]
}
