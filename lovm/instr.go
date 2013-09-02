package lovm

import (
	"fmt"
)

func (b *Builder) IAdd(typ Type, op1, op2 Value) Value {
	return b.Add(&Binop{Valuable{}, "add", typ, op1, op2})
}

func (b *Builder) ICmp(typ Type, op string, op1, op2 Value) Value {
	return b.Add(&Binop{Valuable{}, fmt.Sprintf("icmp %s", op), typ, op1, op2})
}

func (b *Builder) Ref(typ Type, sym Symbol) Value {
	return b.Add(&RefOp{Valuable{}, typ, sym})
}
