package cmdhelp

import (
	"fmt"
	"io/fs"

	"github.com/patroclos/go-conq"
	"github.com/patroclos/go-conq/aid"
)

func New(helpdir fs.FS) *conq.Cmd {
	return &conq.Cmd{
		Name: "help",
		Run: func(c conq.Ctx) error {
			subj := aid.HelpSubject{Cmd: c.Path[0]}
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
