package main

import (
	"fmt"
	"io"
	"os"
)

type Sequence int

type Context struct {
	Tmps   *Sequence
	Scopes *Sequence
}

type Block struct {
	Values []Value
	Vars   map[Symbol]Value
}

type Value interface {
	Name() string
	Emit(w io.Writer)
}

type Symbol struct {
	Name  string
	Scope Sequence
}

type Binop struct {
	Instr string
	Res   string
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

func (b Const) Emit(w io.Writer) {
	// no instructions emitted for consts
}

func (b Binop) Name() string {
	return b.Res
}

func (b Binop) Emit(w io.Writer) {
	fmt.Fprintf(w, "%s = %s %s %s, %s\n", b.Res, b.Instr, b.Typ, b.Op1.Name(), b.Op2.Name())
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

func (b *Block) Emit(w io.Writer, ctx Context) {
	for _, v := range b.Values {
		v.Emit(w)
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

func (d DebugInstr) Emit(w io.Writer) {
	fmt.Fprintf(w, "%s\n", d.Source)
}

func NewBlock() Block {
	return Block{Vars: map[Symbol]Value{}}
}

func NewContext() Context {
	t := Sequence(0)
	s := Sequence(0)
	return Context{&t, &s}
}

func (s *Sequence) Next() Sequence {
	res := *s
	(*s)++
	return res
}

func main() {
	ctx := NewContext()
	entry := NewBlock()
	entry.Assign(Symbol{"a", Sequence(0)}, Const{"i64", "0"})
	entry.Assign(Symbol{"a", Sequence(0)}, Binop{"add", "%0", "i64", Const{"i64", "1"}, Const{"i64", "1"}})
	entry.Emit(os.Stdout, ctx)
}
