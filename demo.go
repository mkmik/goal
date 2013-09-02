package main

import (
	"fmt"
	"goal/lovm"
	"os"
)

func main() {
	ctx := lovm.NewContext(os.Stdout)
	entry := ctx.NewBlock()
	builder := ctx.NewBuilder()
	builder.SetInsertionPoint(entry)

	varA := lovm.Symbol{"a", lovm.Sequence(0)}

	builder.Assign(varA, lovm.Const{"i32", "0"})

	op1 := builder.IAdd("i32", lovm.ConstInt("i32", 1), lovm.ConstInt("i32", 2))
	op2 := builder.IAdd("i32", op1, lovm.ConstInt("i32", 3))

	builder.Assign(varA, op1)
	builder.Assign(varA, op2)

	ifTrue := ctx.NewBlock()
	ifFalse := ctx.NewBlock()
	endIf := ctx.NewBlock()

	cnd := builder.ICmp("i32", "sgt", op2, lovm.ConstInt("i32", 4))
	builder.BranchIf(cnd, ifTrue, ifFalse)
	builder.SetInsertionPoint(ifTrue)
	builder.Assign(varA, builder.IAdd("i32", op1, lovm.ConstInt("i32", 4)))
	builder.Branch(endIf)

	builder.SetInsertionPoint(ifFalse)
	builder.Branch(endIf)

	builder.SetInsertionPoint(endIf)
	builder.Return("i32", builder.Ref("i32", varA))

	fmt.Printf("define i32 @main() {\n")
	ctx.Emit()
	fmt.Printf("}\n")
}
