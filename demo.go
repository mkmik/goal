package main

import (
	"fmt"
	"goal/lovm"
	"os"
)

func main() {
	ctx := lovm.NewContext(os.Stdout)
	entry := ctx.NewBlock()
	entry.Assign(lovm.Symbol{"a", lovm.Sequence(0)}, lovm.Const{"i32", "0"})

	op1 := lovm.IAdd("i32", lovm.ConstInt("i32", 1), lovm.ConstInt("i32", 2))
	op2 := lovm.IAdd("i32", op1, lovm.ConstInt("i32", 3))
	entry.Assign(lovm.Symbol{"a", lovm.Sequence(0)}, op1)
	entry.Assign(lovm.Symbol{"a", lovm.Sequence(0)}, op2)

	ifTrue := ctx.NewBlock()
	ifFalse := ctx.NewBlock()
	endIf := ctx.NewBlock()

	cnd := lovm.ICmp("i32", "sgt", op2, lovm.ConstInt("i32", 4))
	entry.BranchIf(cnd, ifTrue, ifFalse)
	ifTrue.Branch(endIf)
	ifFalse.Branch(endIf)

	fmt.Printf("define i32 @main(i32) {\n")
	endIf.Return("i32", op2)
	ctx.Emit()
	fmt.Printf("}\n")
}
