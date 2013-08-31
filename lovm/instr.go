package lovm

func IAdd(typ string, op1, op2 Value) Value {
	return &Binop{"add", "", typ, op1, op2}
}
