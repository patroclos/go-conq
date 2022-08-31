package conq_test

import "github.com/patroclos/go-conq"

func ExampleHelpSelector() {
	// Output:
	var _ conq.HelpSelector = func(c *conq.Cmd, hs conq.HelpSubject, h conq.Helper, s string) (accept bool, recurse bool) {
		return true, false
	}
}
