package conq

import (
	"io/fs"
	"log"
	"os"

	"github.com/Xuanwo/go-locale"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func IsTerm(f interface{}) bool {
	s, ok := f.(interface{ Stat() (fs.FileInfo, error) })
	if !ok {
		return false
	}
	st, err := s.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice == os.ModeCharDevice
}

func OSContext() Ctx {
	return Ctx{
		In:      os.Stdin,
		Out:     os.Stdout,
		Err:     os.Stderr,
		Args:    os.Args[1:],
		Printer: ctxPrinter(),
	}
}

func ctxPrinter() *message.Printer {
	tags, err := locale.DetectAll()
	if err != nil {
		log.Println("fallback english")
		return message.NewPrinter(language.English)
	}
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.String()
	}
	match := message.MatchLanguage(names...)
	return message.NewPrinter(match)
}
