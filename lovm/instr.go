package lovm

import (
	"fmt"
)

func (b *Builder) IAdd(typ Type, op1, op2 Value) Value {
	return b.Add(&Binop{Valuable{Typ: typ}, "add", op1, op2})
}

func (b *Builder) ICmp(typ Type, op string, op1, op2 Value) Value {
	return b.Add(&Binop{Valuable{Typ: typ}, fmt.Sprintf("icmp %s", op), op1, op2})
}

func (b *Builder) Ref(typ Type, sym Symbol) Value {
	return b.Add(&RefOp{Valuable{Typ: typ}, sym})
}

func (b *Builder) Call(typ Type, fun string, args ...Value) Value {
	return b.Add(&CallOp{Valuable{Typ: typ}, fun, args})
}

func (b *Builder) GEP(typ Type, base Value, indices ...int) Value {
	return b.Add(&GEPOp{Valuable{Typ: DereferenceTypes(typ, indices...)}, base, indices})
}
