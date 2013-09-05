package lovm

type Adder interface {
	Add(Value) Value
	Assign(Register, Value) Value
	Branch(*Block)
	BranchIf(value Value, ifTrue, ifFalse *Block)
	Return(Value)
}

type Builder struct {
	Adder
}

func (ctx *Context) NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) SetInsertionPoint(block *Block) {
	b.Adder = block
}

func (b *Builder) GetInsertBlock() *Block {
	return b.Adder.(*Block)
}
