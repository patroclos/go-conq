# conq - go1.18+ generic commander

text/template based help-generation:
```
$ example help example
`example` is a CLI app built with patroclos/go-conq.

example [commands] [options] arguments
example help [commands|subjects]

Commands:
- help
- completion
- foo
- bar
```

conq contains Option instances (conq.O struct), these stand for flag-parameters
and positional-arguments and are usually used with the generic Opt[T] and
ReqOpt[T] types, which implement parsing for scalars and types implementing
encoding.TextUnmarshaler.  See example to see interesting standard-library
types that can just be used as-is, things like IP, Mac-Addresses, etc.

The biggest focus of conq is composability.  Eventually there should even be
standardized tree-operations to run on a Cmd-tree, like wrapping help-sections,
marking deprecations, etc.

The standard `help` command is just another command-package in `aid/cmdhelp`,
and can be replaced completely, changing how help is resolved from multiple sources
and rendered.

Parameters are something special in conq.  They have a name (ie a 'depth' parameter).
There is a builtin getopt `conq.Optioner` implementation.  It enables extraction and
completion of getopt-style flags, and is also completely modular.  You can make your
own parameter-style extractor.
