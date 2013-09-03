package lovm

import (
	"fmt"
)

const (
	IntSLT = "slt"
	IntSGT = "sgt"
)

func (b *Builder) IAdd(op1, op2 Value) Value {
	return b.Add(&Binop{Valuable{Typ: op1.Type()}, "add", op1, op2})
}

func (b *Builder) ISub(op1, op2 Value) Value {
	return b.Add(&Binop{Valuable{Typ: op1.Type()}, "sub", op1, op2})
}

func (b *Builder) IMul(op1, op2 Value) Value {
	return b.Add(&Binop{Valuable{Typ: op1.Type()}, "mul", op1, op2})
}

func (b *Builder) ISDiv(op1, op2 Value) Value {
	return b.Add(&Binop{Valuable{Typ: op1.Type()}, "sdiv", op1, op2})
}

func (b *Builder) ISRem(op1, op2 Value) Value {
	return b.Add(&Binop{Valuable{Typ: op1.Type()}, "srem", op1, op2})
}

func (b *Builder) IICmp(op string, op1, op2 Value) Value {
	return b.Add(&Binop{Valuable{Typ: op1.Type()}, fmt.Sprintf("icmp %s", op), op1, op2})
}

func (b *Builder) Ref(typ Type, sym Register) Value {
	return b.Add(&RefOp{Valuable{Typ: typ}, sym})
}

func (b *Builder) Call(typ Type, fun string, args ...Value) Value {
	return b.Add(&CallOp{Valuable{Typ: typ}, fun, args})
}

func (b *Builder) GEP(base Value, indices ...int) Value {
	return b.Add(&GEPOp{Valuable{Typ: DereferenceTypes(base.Type(), indices...)}, base, indices})
}
