package lovm

import (
	"fmt"
	"io"
)

type Sequence int

type Context struct {
	Writer io.Writer
	Indent string
	Tmps   Sequence
	Scopes Sequence
	Labels Sequence

	Blocks []*Block
	Values map[Value]bool
}

func NewContext(w io.Writer) (res Context) {
	return Context{
		Writer: w,
		Values: map[Value]bool{},
	}
}

func (s *Sequence) Next() Sequence {
	res := *s
	(*s)++
	return res
}

func (ctx *Context) NewBlock() *Block {
	res := NewBlock(ctx)
	ctx.Blocks = append(ctx.Blocks, res)
	return res
}

func (ctx *Context) Emitf(format string, args ...interface{}) {
	io.WriteString(ctx.Writer, ctx.Indent)
	fmt.Fprintf(ctx.Writer, format, args...)
	io.WriteString(ctx.Writer, "\n")
}

func (ctx *Context) Emit() {
	for _, b := range ctx.Blocks {
		b.Prepare(ctx)
	}

	for _, b := range ctx.Blocks {
		b.Emit(ctx)
	}
}