package conq

import (
	"io/fs"
	"os"
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
		In:   os.Stdin,
		Out:  os.Stdout,
		Err:  os.Stderr,
		Args: os.Args[1:],
	}
}
