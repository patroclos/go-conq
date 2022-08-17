package getopt_test

import (
	"net"
	"testing"

	"github.com/patroclos/go-conq"
	"github.com/patroclos/go-conq/aid"
	"github.com/patroclos/go-conq/commander"
	"github.com/patroclos/go-conq/getopt"
)

// TODO: test cases (flag without value, short option, generic modifiers, aliases, assignment-style)

func TestLongSeparatedSingle(t *testing.T) {
	cmdr := commander.New(getopt.New(), aid.DefaultHelp)

	optFoo := &conq.ReqOpt[string]{Name: "foo"}
	cmd := &conq.Cmd{
		Name: "test-command",
		Opts: conq.Opts{
			optFoo,
		},
		Run: func(c conq.Ctx) error {
			if optFoo.Get(c) != "bar" {
				t.Errorf("expected %q got %q", "bar", optFoo.Get(c))
			}
			return nil
		},
	}

	ctx := conq.OSContext()
	ctx.Args = []string{"--foo", "bar"}
	err := cmdr.Execute(cmd, ctx)

	if err != nil {
		t.Error(err)
	}
}

func TestShortFlag(t *testing.T) {
	optFoo := &conq.ReqOpt[bool]{Name: "foo,f"}
	cmd := &conq.Cmd{
		Name: "test-command",
		Opts: conq.Opts{optFoo},
		Run: func(c conq.Ctx) error {
			if !optFoo.Get(c) {
				t.Error("flag not true")
			}
			return nil
		},
	}

	ctx := conq.OSContext()
	ctx.Args = []string{"-f"}

	err := commander.New(getopt.New(), aid.DefaultHelp).Execute(cmd, ctx)
	if err != nil {
		t.Error(err)
	}
}

func TestShortAssign(t *testing.T) {
	optFoo := &conq.ReqOpt[net.IP]{Name: "ipaddr,a"}

	ctx := conq.OSContext("-a=192.168.2.100")
	cmd := &conq.Cmd{
		Opts: conq.Opts{optFoo},
		Run: func(c conq.Ctx) error {
			if optFoo.Get(c).String() != "192.168.2.100" {
				t.Error("expected ipaddr,a to be 192.168.2.1, got", optFoo.Get(c))
			}
			return nil
		},
	}
	err := commander.New(getopt.New(), aid.DefaultHelp).Execute(cmd, ctx)
	if err != nil {
		t.Error(err)
	}
}

func TestLongAssign(t *testing.T) {
	optFoo := &conq.ReqOpt[int]{Name: "number,n"}
	cmd := &conq.Cmd{
		Name: "test-command",
		Opts: conq.Opts{optFoo},
		Run: func(c conq.Ctx) error {
			if optFoo.Get(c) != 69420 {
				t.Errorf("expected %s=%d, but got %q", optFoo.Name, 69420, optFoo.Get(c))
			}
			return nil
		},
	}

	ctx := conq.OSContext()
	ctx.Args = []string{"--number=69420"}

	err := commander.New(getopt.New(), aid.DefaultHelp).Execute(cmd, ctx)
	if err != nil {
		t.Error(err)
	}
}

func TestLastAssignmentWins(t *testing.T) {
	optFoo := &conq.ReqOpt[int]{Name: "number,n"}
	cmd := &conq.Cmd{
		Name: "test-command",
		Opts: conq.Opts{optFoo},
		Run: func(c conq.Ctx) error {
			if optFoo.Get(c) != 9004 {
				t.Errorf("expected %s=%d, but got %d", optFoo.Name, 9004, optFoo.Get(c))
			}
			return nil
		},
	}

	ctx := conq.OSContext()
	ctx.Args = []string{"--number", "69420", "--number", "9004"}

	err := commander.New(getopt.New(), aid.DefaultHelp).Execute(cmd, ctx)
	if err != nil {
		t.Error(err)
	}
}

func TestCustomParser(t *testing.T) {
	optFoo := &conq.ReqOpt[int]{
		Name: "foo",
		Parse: func(s string) (any, error) {
			switch s {
			case "nice":
				return 69, nil
			default:
				t.Fatal("unexpected parse-input", s)
				return nil, nil
			}
		},
	}

	ctx := conq.OSContext("--foo", "nice")
	cmd := &conq.Cmd{
		Name: "test-command",
		Opts: conq.Opts{optFoo},
		Run: func(c conq.Ctx) error {
			if val := optFoo.Get(c); val != 69 {
				t.Errorf("expected 'nice'=>69, but got %d", val)
			}
			return nil
		},
	}

	err := commander.New(getopt.New(), aid.DefaultHelp).Execute(cmd, ctx)
	if err != nil {
		t.Error(err)
	}
}
