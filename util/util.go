package util

import (
	"fmt"
	"strings"
)

type Sequence int
type Sequential int

func (s *Sequence) Next() Sequential {
	res := *s
	(*s)++
	return Sequential(res)
}

func Perrorf(format string, args ...interface{}) {
	panic(fmt.Errorf(format, args...))
}

func AssertNotNil(args ...interface{}) {
	var failedArgs []string
	for i, a := range args {
		if a == nil {
			failedArgs = append(failedArgs, fmt.Sprintf("%d", i))
		}
	}
	if failedArgs != nil {
		Perrorf("runtime error: Nil assertion failed. Failed args: %s",
			strings.Join(failedArgs, ", "))
	}
}
