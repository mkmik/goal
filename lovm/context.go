package lovm

import (
	"io"
)

type Context struct {
	Writer io.Writer
	Modules []Module
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

fun (ctx *Context) Emit() {
	for _, m := range ctx.Modules {
		m.Emit()
	}
}

fun (ctx *Module) Emit() {
	for _, f := range ctx.Functions {
		f.Emit()
	}
}