package aid

import (
	"fmt"
	"strings"

	"github.com/patroclos/go-conq"
)

func SubjectIdentifier(path conq.Pth, opt conq.Opter) string {
	var b strings.Builder
	for i, c := range path {
		switch i {
		case 0:
			b.WriteString(c.Name)
		default:
			fmt.Fprintf(&b, ".%s", c.Name)
		}
	}
	if opt != nil {
		fmt.Fprintf(&b, "[%s]", opt.Opt().Name)
	}

	return b.String()
}
