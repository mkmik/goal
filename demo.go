package main

import (
	"goal/lovm"
	"os"
)

func main() {
	ctx := lovm.NewContext(os.Stdout)
	entry := lovm.NewBlock()
	entry.Assign(lovm.Symbol{"a", lovm.Sequence(0)}, lovm.Const{"i64", "0"})

	op1 := lovm.IAdd("i64", lovm.ConstInt("i64", 1), lovm.ConstInt("i64", 2))
	op2 := lovm.IAdd("i64", op1, lovm.ConstInt("i64", 3))
	entry.Assign(lovm.Symbol{"a", lovm.Sequence(0)}, op1)
	entry.Assign(lovm.Symbol{"a", lovm.Sequence(0)}, op2)
	entry.Emit(&ctx)
}
