package lovm

import (
	"fmt"
	"log"
	"strings"
)

type Block struct {
	Labelable
	Phis    []Value
	Values  []Value
	Preds   []*Block
	Vars    map[Register]Value
	Context *Context
}

type Emitter interface {
	Prepare(*Context, *Block)
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

type Register interface {
}

type Valuable struct {
	Res Sequence
}

type Labelable struct {
	Res Sequence
}

func (b Valuable) Name() string {
	return fmt.Sprintf("%%%d", b.Res)
}

func (v *Valuable) Prepare(ctx *Context, b *Block) {
	v.Res = ctx.Tmps.Next()
}

func (v *Labelable) Prepare(ctx *Context) {
	v.Res = ctx.Labels.Next()
}

type Binop struct {
	Valuable
	Instr string
	Typ   string
	Op1   Value
	Op2   Value
}

type BranchOp struct {
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

type RefOp struct {
	Valuable
	Typ string
	Sym Register
}

type PhiParam struct {
	Value string
	Label string
}

type PhiOp struct {
	RefOp
	Phis []PhiParam
}

func (c Const) Name() string {
	return c.Val
}

func (b Const) Emit(*Context) {
	// no instructions emitted for consts
}

func (b Const) Prepare(*Context, *Block) {
	// no instructions emitted for consts
}

func (b RefOp) Emit(ctx *Context) {
	if false {
		log.Fatalf("RefOps have to be replaced during prepare")
	}
}

func (r *RefOp) Prepare(ctx *Context, b *Block) {
	r.Valuable.Prepare(ctx, b)

	phis := []PhiParam{}
	for _, p := range b.Preds {
		if v, ok := p.ResolveVar(r.Sym); ok {
			phis = append(phis, PhiParam{v.Name(), p.Name()})
		}
	}
	b.Phis = append(b.Phis, &PhiOp{*r, phis})
}

func (b PhiOp) Emit(ctx *Context) {
	comps := []string{}
	for _, phi := range b.Phis {
		comps = append(comps, fmt.Sprintf("[ %s, %s ]", phi.Value, phi.Label))
	}
	ctx.Emitf("%s = phi %s %s", b.Name(), b.Typ, strings.Join(comps, ", "))
}

func ConstInt(typ string, value int) Const {
	return Const{typ, fmt.Sprintf("%d", value)}
}

func (b *Binop) Emit(ctx *Context) {
	ctx.Emitf("%s = %s %s %s, %s", b.Name(), b.Instr, b.Typ, b.Op1.Name(), b.Op2.Name())
}

func (b *BranchOp) Name() string {
	log.Fatalf("Branch ops should never be named")
	return ""
}

func (b *BranchOp) Prepare(*Context, *Block) {
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

func (b *Block) Add(value Value) Value {
	if !b.Context.Values[value] {
		b.Values = append(b.Values, value)
		b.Context.Values[value] = true
	}
	return value
}

func (b *Block) ResolveVar(symbol Register) (Value, bool) {
	if v, ok := b.Vars[symbol]; ok {
		return v, ok
	}
	if len(b.Preds) == 1 {
		return b.Preds[0].ResolveVar(symbol)
	} else if len(b.Preds) > 1 {
		log.Printf("Assert: block %#v has multiple predecessors but no PHI for var %#v", b, symbol)
	}
	return nil, false
}

func (b *Block) Assign(symbol Register, value Value) Value {
	res := b.Add(value)
	b.Vars[symbol] = value
	return res
}

func (b *Block) AddPred(source *Block) {
	for _, p := range b.Preds {
		if p == source {
			return
		}
	}
	b.Preds = append(b.Preds, source)
}

func (b *Block) Branch(target *Block) {
	target.AddPred(b)
	b.Add(&BranchOp{[]*Block{target}})
}

func (b *Block) BranchIf(value Value, ifTrue, ifFalse *Block) {
	b.Add(value)
	ifTrue.AddPred(b)
	ifFalse.AddPred(b)
	b.Add(&BranchIfOp{BranchOp{[]*Block{ifTrue, ifFalse}}, value})
}

func (b *Block) Return(typ string, value Value) {
	b.Add(value)
	b.Add(&ReturnOp{Valuable{}, typ, value})
}

func (b *Block) Name() string {
	return fmt.Sprintf("%%label%d", b.Res)
}

func (b *Block) Prepare(ctx *Context) {
	b.Labelable.Prepare(ctx)
	for _, v := range b.Values {
		v.Prepare(ctx, b)
	}
	// insert phi nodes
	for _, p := range b.Phis {
		b.Vars[p.(*PhiOp).Sym] = p
	}
	b.Values = append(b.Phis, b.Values...)
}

func (b *Block) PrettyPreds() string {
	res := make([]string, len(b.Preds))
	for i, p := range b.Preds {
		res[i] = p.Name()
	}
	return strings.Join(res, ", ")
}

func (b *Block) Emit(ctx *Context) {
	preds := b.PrettyPreds()
	ctx.Emitf("label%d:\t\t\t\t\t\t; preds = %s", b.Res, preds)
	ctx.Indent = "  "
	defer func() {
		ctx.Indent = ""
	}()

	for _, v := range b.Values {
		v.Emit(ctx)
	}
}

func NewBlock(ctx *Context) *Block {
	return &Block{Context: ctx, Vars: map[Register]Value{}}
}
