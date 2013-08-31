package lovm

import (
	"fmt"
)

func IAdd(typ string, op1, op2 Value) Value {
	return &Binop{Valuable{}, "add", typ, op1, op2}
}

func ICmp(typ string, op string, op1, op2 Value) Value {
	return &Binop{Valuable{}, fmt.Sprintf("icmp %s", op), typ, op1, op2}
}
