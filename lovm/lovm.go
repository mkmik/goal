package lovm

import (
	"fmt"
	"log"
	"strings"
)

type Block struct {
	Labelable
	Phis     []Value
	Values   []Value
	Preds    []*Block
	Vars     map[Register]Value
	Function *Function
}

type Emitter interface {
	Prepare(*Function, *Block)
	Emit(*Function)
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

func (v *Valuable) Prepare(fun *Function, b *Block) {
	v.Res = fun.Tmps.Next()
}

func (v *Labelable) Prepare(fun *Function) {
	v.Res = fun.Labels.Next()
}

type Binop struct {
	Valuable
	Instr string
	Typ   Type
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
	Typ    Type
	Result Value
}

type Const struct {
	Typ Type
	Val string
}

type RefOp struct {
	Valuable
	Typ Type
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

func (b Const) Emit(*Function) {
	// no instructions emitted for consts
}

func (b Const) Prepare(*Function, *Block) {
	// no instructions emitted for consts
}

func (b RefOp) Emit(fun *Function) {
	if false {
		log.Fatalf("RefOps have to be replaced during prepare")
	}
}

func (r *RefOp) Prepare(fun *Function, b *Block) {
	r.Valuable.Prepare(fun, b)

	phis := []PhiParam{}
	for _, p := range b.Preds {
		if v, ok := p.ResolveVar(r.Sym); ok {
			phis = append(phis, PhiParam{v.Name(), p.Name()})
		}
	}
	b.Phis = append(b.Phis, &PhiOp{*r, phis})
}

func (b PhiOp) Emit(fun *Function) {
	comps := []string{}
	for _, phi := range b.Phis {
		comps = append(comps, fmt.Sprintf("[ %s, %s ]", phi.Value, phi.Label))
	}
	fun.Emitf("%s = phi %s %s", b.Name(), b.Typ.Name, strings.Join(comps, ", "))
}

func ConstInt(typ Type, value int) Const {
	return Const{typ, fmt.Sprintf("%d", value)}
}

func (b *Binop) Emit(fun *Function) {
	fun.Emitf("%s = %s %s %s, %s", b.Name(), b.Instr, b.Typ.Name, b.Op1.Name(), b.Op2.Name())
}

func (b *BranchOp) Name() string {
	log.Fatalf("Branch ops should never be named")
	return ""
}

func (b *BranchOp) Prepare(*Function, *Block) {
}

func (b *BranchOp) Emit(fun *Function) {
	fun.Emitf("br label %s", b.Labels[0].Name())
}

func (b *BranchIfOp) Emit(fun *Function) {
	fun.Emitf("br i1 %s, label %s, label %s", b.Cond.Name(), b.Labels[0].Name(), b.Labels[1].Name())
}

func (b *ReturnOp) Emit(fun *Function) {
	fun.Emitf("ret %s %s", b.Typ.Name, b.Result.Name())
}

func (b *Block) Add(value Value) Value {
	if !b.Function.Values[value] {
		b.Values = append(b.Values, value)
		b.Function.Values[value] = true
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

func (b *Block) Return(typ Type, value Value) {
	b.Add(value)
	b.Add(&ReturnOp{Valuable{}, typ, value})
}

func (b *Block) Name() string {
	return fmt.Sprintf("%%label%d", b.Res)
}

func (b *Block) Prepare(fun *Function) {
	b.Labelable.Prepare(fun)
	for _, v := range b.Values {
		v.Prepare(fun, b)
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

func (b *Block) Emit(fun *Function) {
	preds := b.PrettyPreds()
	fun.Emitf("label%d:\t\t\t\t\t\t; preds = %s", b.Res, preds)
	fun.Indent = "  "
	defer func() {
		fun.Indent = ""
	}()

	for _, v := range b.Values {
		v.Emit(fun)
	}
}

func NewBlock(fun *Function) *Block {
	return &Block{Function: fun, Vars: map[Register]Value{}}
}
