package main

import (
	"goal/lovm"
	"os"
)

func main() {
	ctx := lovm.NewContext(os.Stdout)
	mod := ctx.NewModule()
	fun := mod.NewFunction("@main", lovm.FunctionType(lovm.IntType(32), false))
	entry := fun.NewBlock()
	builder := fun.NewBuilder()
	builder.SetInsertionPoint(entry)

	varA := lovm.Symbol{"a", lovm.Sequence(0)}
	typ := lovm.IntType(32)

	builder.Assign(varA, lovm.Const{typ, "0"})

	op1 := builder.IAdd(typ, lovm.ConstInt(typ, 1), lovm.ConstInt(typ, 2))
	op2 := builder.IAdd(typ, op1, lovm.ConstInt(typ, 3))

	builder.Assign(varA, op1)
	builder.Assign(varA, op2)

	ifTrue := fun.NewBlock()
	ifFalse := fun.NewBlock()
	endIf := fun.NewBlock()

	cnd := builder.ICmp(typ, "sgt", op2, lovm.ConstInt(typ, 4))
	builder.BranchIf(cnd, ifTrue, ifFalse)
	builder.SetInsertionPoint(ifTrue)
	builder.Assign(varA, builder.IAdd(typ, op1, lovm.ConstInt(typ, 4)))
	builder.Branch(endIf)

	builder.SetInsertionPoint(ifFalse)
	builder.Branch(endIf)

	builder.SetInsertionPoint(endIf)

	printfType := lovm.FunctionType(lovm.IntType(32), true, lovm.PointerType(lovm.IntType(8)))
	printfSym := mod.DeclareExternal("@printf", printfType)
	a := builder.Ref(typ, varA)
	str := mod.ConstString("hello world\n")
	builder.Call(printfSym.Type(), printfSym.Name(), builder.GEP(str.Type(), str, 0, 0), a)
	builder.Return(typ, a)

	ctx.Emit()
}
