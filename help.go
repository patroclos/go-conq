package conq

import "io/fs"

// HelpSubject represents a command&option-tree node that is subject to a help query.
type HelpSubject struct {
	Cmd *Cmd
	Opt *O
	Ctx *Ctx
}

// Helper provides a helptexts for nodes in the command&options tree.
type Helper interface {
	Help(HelpSubject) string
}

// HelpSelector is a func that is used when walking the command-tree to assemble
// only the subjects the selector accepts into the helptext.
type HelpSelector func(*Cmd, HelpSubject, Helper, string) (accept, recurse bool)

// GO AWAY!

// CmdHelp details how this command is to be treated by help/assistance commands
type CmdHelp struct {
	// Filter for the subjects that will considered in help-generation.
	// An interesting usecase could be to skip non-runnable commands or select for
	// a certain depth in the tree.
	Select HelpSelector
	// Articles contains templates for generating helptexts using cmdhelp.HelpContext.
	Articles fs.FS
}
