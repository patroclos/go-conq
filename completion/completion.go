package completion

import "github.com/posener/complete"

type Context struct {
	Args complete.Args

	// is any of this used?
	Values  map[string]any
	Strings map[string]string
	Options []string
	Closed  bool
}
