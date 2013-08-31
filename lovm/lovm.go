package lovm

import (
	"fmt"
)

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

func NewBlock() *Block {
	return &Block{Vars: map[Symbol]Value{}}
}
