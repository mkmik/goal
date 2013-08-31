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

	Blocks []*Block
}

func (ctx *Context) NewBlock() *Block {
	res := NewBlock()
	ctx.Blocks = append(ctx.Blocks, res)
	return res
}

func (ctx *Context) Emit() {
	for _, b := range ctx.Blocks {
		b.Emit(ctx)
	}
}

func (ctx *Context) Emitf(format string, args ...interface{}) {
	io.WriteString(ctx.Writer, ctx.Indent)
	fmt.Fprintf(ctx.Writer, format, args...)
	io.WriteString(ctx.Writer, "\n")
}

type Block struct {
	Valuable
	Values []Value
	Vars   map[Symbol]Value
}

type Emitter interface {
	Emit(*Context)
}

type Value interface {
	Emitter
	Name() string
}

type Symbol struct {
	Name  string
	Scope Sequence
}

type Valuable struct {
	Res Sequence
}

func (b Valuable) Name() string {
	return fmt.Sprintf("%%%d", b.Res)
}

func (b *Valuable) Emit(ctx *Context) {
	b.Res = ctx.Tmps.Next()
}

type Binop struct {
	Valuable
	Instr string
	Typ   string
	Op1   Value
	Op2   Value
}

type Const struct {
	Typ string
	Val string
}

func (c Const) Name() string {
	return c.Val
}

func (b Const) Emit(Ectx *Context) {
	// no instructions emitted for consts
}

func ConstInt(typ string, value int) Const {
	return Const{typ, fmt.Sprintf("%d", value)}
}

func (b *Binop) Emit(ctx *Context) {
	b.Valuable.Emit(ctx)
	ctx.Emitf("%s = %s %s %s, %s", b.Name(), b.Instr, b.Typ, b.Op1.Name(), b.Op2.Name())
}

func (b *Block) Assign(symbol Symbol, value Value) {
	b.Values = append(b.Values, value)
	b.Vars[symbol] = value
}

func (b *Block) Branch(target *Block) {
	b.Values = append(b.Values, DebugInstrf("br ...."))
}

func (b *Block) BranchIf(value Value, ifTrue, ifFalse *Block) {
	b.Values = append(b.Values, value)
	b.Values = append(b.Values, DebugInstrf("brif ...."))
}

func (b *Block) Name() string {
	return fmt.Sprintf(".label%d", b.Res)
}

func (b *Block) Emit(ctx *Context) {
	b.Valuable.Emit(ctx)
	ctx.Emitf("%s", b.Name())
	ctx.Indent = "  "
	defer func() {
		ctx.Indent = ""
	}()

	for _, v := range b.Values {
		v.Emit(ctx)
	}
}

type DebugInstr struct {
	Source string
}

func DebugInstrf(format string, args ...interface{}) DebugInstr {
	return DebugInstr{fmt.Sprintf(format, args...)}
}

func (d DebugInstr) Name() string {
	return "%debuginstr"
}

func (d DebugInstr) Emit(ctx *Context) {
	ctx.Emitf("%s\n", d.Source)
}

func NewBlock() *Block {
	return &Block{Vars: map[Symbol]Value{}}
}

func NewContext(w io.Writer) Context {
	return Context{Writer: w}
}

func (s *Sequence) Next() Sequence {
	res := *s
	(*s)++
	return res
}
