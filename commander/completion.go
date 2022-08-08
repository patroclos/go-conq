package commander

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/posener/complete"
)

type completionCtx struct {
	line  string
	point int
	t     comptype
}

type comptype int

func (t comptype) String() string {
	return fmt.Sprintf("COMP_TYPE='%c'", t)
}

const (
	ctNormal = comptype('\t')
	ctSucc   = comptype('?')
	ctAlt    = comptype('!')
	ctMod    = comptype('@')
	ctMenu   = comptype('%')
)

func completionContext() (line string, point int, ctype comptype, ok bool) {
	line = os.Getenv("COMP_LINE")
	if line == "" {
		return
	}

	point, err := strconv.Atoi(os.Getenv("COMP_POINT"))
	if err != nil {
		point = len(line)
	}

	tint, err := strconv.Atoi(os.Getenv("COMP_TYPE"))
	if err != nil {
		return
	}
	ctype = comptype(tint)
	ok = true
	return
}

func complArgs(line string) complete.Args {
	var (
		all       []string
		completed []string
		last      string
		lastComp  string
	)
	parts := strings.Fields(line)

	if len(line) > 0 && unicode.IsSpace(rune(line[len(line)-1])) {
		parts = append(parts, "")
	}

	if len(parts) > 0 {
		all = parts[1:]
		if len(all) > 0 {
			completed = all[:len(all)-1]
		}
	}

	if len(parts) > 0 {
		last = parts[len(parts)-1]
	}

	if len(completed) > 0 {
		lastComp = completed[len(completed)-1]
	}

	return complete.Args{
		All:           all,
		Completed:     completed,
		Last:          last,
		LastCompleted: lastComp,
	}
}

func sliceArgs(a complete.Args, start int) complete.Args {
	a.All = a.All[start:]
	if start > len(a.Completed) {
		start = len(a.Completed)
	}
	a.Completed = a.Completed[start:]

	a.Last, a.LastCompleted = "", ""
	if len(a.All) > 0 {
		a.Last = a.All[len(a.All)-1]
	}
	if len(a.Completed) > 0 {
		a.LastCompleted = a.Completed[len(a.Completed)-1]
	}
	return a
}
