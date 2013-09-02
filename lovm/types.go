package lovm

import (
	"fmt"
	"strings"
)

type Type struct {
	Name string
}

func IntType(size int) Type {
	return Type{fmt.Sprintf("i%d", size)}
}

func FunctionType(ret Type, params []Type) Type {
	paramNames := make([]string, len(params))
	for i, p := range params {
		paramNames[i] = p.Name
	}
	return Type{fmt.Sprintf("%s (%s)", ret.Name, strings.Join(paramNames, ", "))}
}

func PointerType(typ Type) Type {
	return Type{fmt.Sprintf("%s *", typ.Name)}
}
