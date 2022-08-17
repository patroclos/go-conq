package cmdhelp

import (
	"fmt"
	"io/fs"
	"strings"
	"text/template"

	"github.com/patroclos/go-conq"
	"github.com/patroclos/go-conq/aid"
)

var OptVerbose = conq.Opt[bool](conq.O{Name: "verbose,v"})

func New(helpdir fs.FS) *conq.Cmd {
	return &conq.Cmd{
		Name: "help",
		Opts: conq.Opts{OptVerbose},
		Run: func(c conq.Ctx) error {
			if _, err := OptVerbose.Get(c); err == nil {
				c.Printer.Fprintf(c.Err, "VERBOSE MODE Command:%q Values: %v\n", c.Path[len(c.Path)-1].Name, c.Values)
			}

			subj := aid.HelpSubject{Cmd: c.Path[0]}
			if helpdir != nil {
				if err := printSection(helpdir, c); err == nil {
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

func printSection(dir fs.FS, c conq.Ctx) error {
	var paths []string
	fs.WalkDir(dir, "help", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if !strings.HasPrefix(path, "help") {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".tmpl") {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if len(paths) == 0 {
		return fmt.Errorf("no sections found")
	}
	tmpl, err := template.ParseFS(dir, paths...)
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

// HelpContext is the input for help-templates
type HelpContext conq.Ctx

func (c HelpContext) Root() *conq.Cmd {
	return c.Path[0]
}

func (c HelpContext) Cmd() *conq.Cmd {
	return c.Path[len(c.Path)-1]
}
