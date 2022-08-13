package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/patroclos/go-conq"
	"github.com/patroclos/go-conq/aid"
	"github.com/patroclos/go-conq/commander"
	"github.com/patroclos/go-conq/getopt"
)

func TestUnansiCommand(t *testing.T) {
	com := commander.New(getopt.New(), aid.DefaultHelp)
	cmd := New()
	ctx := conq.OSContext()
	ctx.Args = []string{"unansi"}
	ctx.In = strings.NewReader("\033[31mcolorful")
	buf := bytes.NewBuffer(make([]byte, 0, 16))
	ctx.Out = buf

	err := com.Execute(cmd, ctx)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if output != "colorful" {
		t.Errorf("expected colorful, got %q", output)
	}
}

func ExampleShortFlag() {
	com := commander.New(getopt.New(), aid.DefaultHelp)
	cmd := New()
	ctx := conq.OSContext()
	ctx.Args = []string{"help", "-v"}

	if err := com.Execute(cmd, ctx); err != nil {
		panic(err)
	}
}
