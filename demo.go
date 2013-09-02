package main

import (
	"fmt"
	"goal/lovm"
	"os"
)

func main() {
	ctx := lovm.NewContext(os.Stdout)
	mod := ctx.NewModule()
	fun := mod.NewFunction()
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
	builder.Return(typ, builder.Ref(typ, varA))

	fmt.Printf("define i32 @main() {\n")
	ctx.Emit()
	fmt.Printf("}\n")
}
