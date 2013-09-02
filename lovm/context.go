package lovm

import (
	"io"
)

type Context struct {
	Writer io.Writer
}

func NewContext(w io.Writer) Context {
	return Context{
		Writer: w,
	}
}

type Module struct {
	*Context
	functions []Function
}