package lovm

import (
	"fmt"
)

type Type struct {
	Name string
}

func IntType(size int) Type {
	return Type{fmt.Sprintf("i%d", size)}
}
