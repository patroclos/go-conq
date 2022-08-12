package unansi

import (
	"fmt"
	"io"
	"regexp"

	"github.com/patroclos/go-conq"
)

func New() *conq.Cmd {
	return &conq.Cmd{
		Name: "unansi",
		Run:  run,
	}
}

func run(c conq.Ctx) error {
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

	re := regexp.MustCompile(ansi)
	txt, err := io.ReadAll(c.In)
	if err != nil {
		return fmt.Errorf("failed reading stdin: %w", err)
	}
	txt = re.ReplaceAll(txt, nil)
	fmt.Fprintf(c.Out, "%s", txt)
	return nil
}
