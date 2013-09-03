package lovm

import (
	"fmt"
	"io"
)

type Sequence int

type Function struct {
	*Module
	Indent string
	Tmps   Sequence
	Labels Sequence

	Blocks []*Block
	Values map[Value]bool
	Params []*Param
	Type   FuncType
	Name   string
}

func (mod *Module) NewFunction(name string, typ Type) *Function {
	signature := typ.(FuncType)
	fun := &Function{
		Module: mod,
		Values: map[Value]bool{},
		Type:   signature,
		Name:   name,
	}

	mod.AddFunction(fun)
	paramBlock := NewBlock(fun)
	for _, paramType := range signature.ParamTypes {
		param := &Param{Valuable{}, paramType}
		fun.Params = append(fun.Params, param)
		paramBlock.Add(param)
	}
	paramBlock.Prepare(fun)
	return fun
}

func (s *Sequence) Next() Sequence {
	res := *s
	(*s)++
	return res
}

func (fun *Function) NewBlock() *Block {
	res := NewBlock(fun)
	fun.Blocks = append(fun.Blocks, res)
	return res
}

func (fun *Function) Param(idx int) Value {
	return fun.Params[idx]
}

func (fun *Function) Emitf(format string, args ...interface{}) {
	io.WriteString(fun.Writer, fun.Indent)
	fmt.Fprintf(fun.Writer, format, args...)
	io.WriteString(fun.Writer, "\n")
}

func (fun *Function) Emit() {
	for _, b := range fun.Blocks {
		b.Prepare(fun)
	}

	fun.Type.EmitDef(fun.Writer, fun.Name, func() {
		for _, b := range fun.Blocks {
			b.Emit(fun)
		}
	})
}
