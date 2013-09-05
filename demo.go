package main

import (
	"goal/lovm"
	"goal/util"
	"log"
	"os"
)

type Symbol struct {
	Name  string
	Scope util.Sequential
}

func main() {
	ctx := lovm.NewContext(os.Stdout)
	mod := ctx.NewModule("main")
	fun := mod.NewFunction("main", lovm.FunctionType(lovm.IntType(32), false, lovm.IntType(32), lovm.IntType(32)))
	entry := fun.NewBlock()
	builder := fun.NewBuilder()
	builder.SetInsertionPoint(entry)

	seq := util.Sequence(0)

	varA := Symbol{"a", seq.Next()}
	varP := Symbol{"p", seq.Next()}
	varQ := Symbol{"q", seq.Next()}
	typ := lovm.IntType(32)

	builder.Assign(varA, lovm.Const{typ, "0"})

	param := fun.Param(0)
	builder.Assign(varP, param)
	param2 := fun.Param(1)
	builder.Assign(varQ, param2)

	builder.Assign(varP, builder.IAdd(param, param2))

	op1 := builder.IAdd(lovm.ConstInt(typ, 1), builder.Ref(typ, varP))
	builder.Return(op1)

	ctx.Emit()
}

func mainx() {
	ctx := lovm.NewContext(os.Stdout)
	mod := ctx.NewModule("main")
	//	fun := mod.NewFunction("main", lovm.FunctionType(lovm.IntType(32), false, lovm.IntType(32), lovm.PointerType(lovm.PointerType(lovm.IntType(8)))))
	fun := mod.NewFunction("main", lovm.FunctionType(lovm.IntType(32), false, lovm.IntType(32), lovm.IntType(32)))
	entry := fun.NewBlock()
	builder := fun.NewBuilder()
	builder.SetInsertionPoint(entry)

	seq := util.Sequence(0)

	varA := Symbol{"a", seq.Next()}
	varP := Symbol{"p", seq.Next()}
	varQ := Symbol{"q", seq.Next()}
	typ := lovm.IntType(32)

	builder.Assign(varA, lovm.Const{typ, "0"})

	param := fun.Param(0)
	builder.Assign(varP, param)
	param2 := fun.Param(1)
	builder.Assign(varQ, param2)

	builder.Assign(varP, builder.IAdd(param, param2))

	//	op1 := builder.IAdd(lovm.ConstInt(typ, 1), param)
	op1 := builder.IAdd(lovm.ConstInt(typ, 1), builder.Ref(typ, varP))
	op2 := builder.IAdd(op1, lovm.ConstInt(typ, 3))

	builder.Assign(varA, op1)
	builder.Assign(varA, op2)

	ifTrue := fun.NewBlock()
	ifFalse := fun.NewBlock()
	endIf := fun.NewBlock()

	cnd := builder.IICmp(lovm.IntSGT, op2, lovm.ConstIntFromString(typ, "B", 16))
	builder.BranchIf(cnd, ifTrue, ifFalse)
	builder.SetInsertionPoint(ifTrue)
	builder.Assign(varA, builder.IAdd(op1, lovm.ConstInt(typ, 4)))
	builder.Branch(endIf)

	builder.SetInsertionPoint(ifFalse)
	dummy := builder.GetInsertBlock()
	if ifFalse != dummy {
		log.Fatalf("assertion error, %v != %v", dummy, ifFalse)
	}
	builder.Branch(endIf)

	builder.SetInsertionPoint(endIf)

	printfType := lovm.FunctionType(lovm.IntType(32), true, lovm.PointerType(lovm.IntType(8)))
	printfSym := mod.DeclareExternal("printf", printfType)
	a := builder.Ref(typ, varA)
	str := mod.ConstString("hello world\n")
	builder.Call(printfSym.Type(), printfSym.Name(), builder.GEP(str, 0, 0), a)
	builder.Return(a)

	ctx.Emit()
}
