package lovm

import (
	"fmt"
	"io"
)

type Sequence int

type Function struct {
	*Context
	Indent string
	Tmps   Sequence
	Labels Sequence

	Blocks []*Block
	Values map[Value]bool
}

func NewFunction(ctx *Context) Function {
	return Function{
		Context: ctx,
		Values:  map[Value]bool{},
	}
}

func (s *Sequence) Next() Sequence {
	res := *s
	(*s)++
	return res
}

func (ctx *Function) NewBlock() *Block {
	res := NewBlock(ctx)
	ctx.Blocks = append(ctx.Blocks, res)
	return res
}

func (ctx *Function) Emitf(format string, args ...interface{}) {
	io.WriteString(ctx.Writer, ctx.Indent)
	fmt.Fprintf(ctx.Writer, format, args...)
	io.WriteString(ctx.Writer, "\n")
}

func (ctx *Function) Emit() {
	for _, b := range ctx.Blocks {
		b.Prepare(ctx)
	}

	for _, b := range ctx.Blocks {
		b.Emit(ctx)
	}
}
