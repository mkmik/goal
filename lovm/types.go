package lovm

import (
	"fmt"
	"strings"
)

type Type struct {
	Name string
	Decl string
}

func IntType(size int) Type {
	return Type{Name: fmt.Sprintf("i%d", size)}
}

func FunctionType(name string, ret Type, params []Type) Type {
	paramNames := make([]string, len(params))
	for i, p := range params {
		paramNames[i] = p.Name
	}
	return Type{
		Name: fmt.Sprintf("%s (%s)", ret.Name, strings.Join(paramNames, ", ")),
		Decl: fmt.Sprintf("%s %s(%s)", ret.Name, name, strings.Join(paramNames, ", ")),
	}
}

func PointerType(typ Type) Type {
	return Type{Name: fmt.Sprintf("%s *", typ.Name)}
}
