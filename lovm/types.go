package lovm

import (
	"fmt"
	"strings"
)

type Type interface {
	Name() string
}

type BasicType struct {
	name string
}

func (b BasicType) Name() string {
	return b.name
}

func IntType(size int) Type {
	return BasicType{fmt.Sprintf("i%d", size)}
}

type FuncType struct {
	ReturnType   string
	ParamTypes   []string
}

func (f FuncType) Name() string {
	return fmt.Sprintf("%s (%s)", f.ReturnType, strings.Join(f.ParamTypes, ", "))
}

func (f FuncType) Decl(name string) string {
	return fmt.Sprintf("%s %s(%s)", f.ReturnType, name, strings.Join(f.ParamTypes, ", "))
}

func FunctionType(ret Type, params []Type) FuncType {
	paramTypes := make([]string, len(params))
	for i, p := range params {
		paramTypes[i] = p.Name()
	}
	return FuncType{
		ReturnType:   ret.Name(),
		ParamTypes:   paramTypes,
	}
}

func PointerType(typ Type) Type {
	return BasicType{fmt.Sprintf("%s *", typ.Name())}
}
