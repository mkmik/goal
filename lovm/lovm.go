package lovm

import (
	"fmt"
)

type Block struct {
	Valuable
	Values  []Value
	Vars    map[Symbol]Value
	Context *Context
}

type Emitter interface {
	Prepare(*Context)
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

func (b *Valuable) Prepare(ctx *Context) {
	b.Res = ctx.Tmps.Next()
}

type Binop struct {
	Valuable
	Instr string
	Typ   string
	Op1   Value
	Op2   Value
}

type BranchOp struct {
	Valuable
	Labels []*Block
}

type BranchIfOp struct {
	BranchOp
	Cond Value
}

type ReturnOp struct {
	Valuable
	Typ    string
	Result Value
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

func (b Const) Prepare(Ectx *Context) {
	// no instructions emitted for consts
}

func ConstInt(typ string, value int) Const {
	return Const{typ, fmt.Sprintf("%d", value)}
}

func (b *Binop) Emit(ctx *Context) {
	ctx.Emitf("%s = %s %s %s, %s", b.Name(), b.Instr, b.Typ, b.Op1.Name(), b.Op2.Name())
}

func (b *BranchOp) Emit(ctx *Context) {
	ctx.Emitf("br label %s", b.Labels[0].Name())
}

func (b *BranchIfOp) Emit(ctx *Context) {
	ctx.Emitf("br i1 %s, label %s, label %s", b.Cond.Name(), b.Labels[0].Name(), b.Labels[1].Name())
}

func (b *ReturnOp) Emit(ctx *Context) {
	ctx.Emitf("ret %s %s", b.Typ, b.Result.Name())
}

func (b *Block) Add(value Value) {
	if !b.Context.Values[value] {
		b.Values = append(b.Values, value)
		b.Context.Values[value] = true
	}
}

func (b *Block) Assign(symbol Symbol, value Value) {
	b.Add(value)
	b.Vars[symbol] = value
}

func (b *Block) Branch(target *Block) {
	b.Add(&BranchOp{Valuable{}, []*Block{target}})
}

func (b *Block) BranchIf(value Value, ifTrue, ifFalse *Block) {
	b.Add(value)
	b.Add(&BranchIfOp{BranchOp{Valuable{}, []*Block{ifTrue, ifFalse}}, value})
}

func (b *Block) Return(typ string, value Value) {
	b.Add(value)
	b.Add(&ReturnOp{Valuable{}, typ, value})
}

func (b *Block) Name() string {
	return fmt.Sprintf("%%label%d", b.Res)
}

func (b *Block) Prepare(ctx *Context) {
	b.Valuable.Prepare(ctx)
	for _, v := range b.Values {
		v.Prepare(ctx)
	}
}

func (b *Block) Emit(ctx *Context) {
	ctx.Emitf("label%d:", b.Res)
	ctx.Indent = "  "
	defer func() {
		ctx.Indent = ""
	}()

	for _, v := range b.Values {
		v.Emit(ctx)
	}
}

func NewBlock(ctx *Context) *Block {
	return &Block{Context: ctx, Vars: map[Symbol]Value{}}
}
