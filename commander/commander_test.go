package commander

import (
	"testing"

	"github.com/patroclos/go-conq"
)

func TestResolveNestedSubcommand(t *testing.T) {
	cmdr := New(nil, nil)

	cmd := &conq.Cmd{
		Commands: []*conq.Cmd{
			{Name: "foo"},
			{Name: "flooz", Commands: []*conq.Cmd{{Name: "blarg"}}},
			{Name: "fez"},
		},
	}

	ctx := conq.OSContext("flooz blarg")
	ctx.Args = []string{"flooz", "blarg"}

	ctx = cmdr.ResolveCmd(cmd, ctx)

	if len(ctx.Path) != 3 {
		t.Fatal("expected 3 cmds in stack, got", len(ctx.Path))
	}

	if ctx.Path[len(ctx.Path)-1].Name != "blarg" {
		t.Error("wrong command resolved")
	}
}
