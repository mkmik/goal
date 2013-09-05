package util

import (
	"fmt"
	"strings"
)

type Sequence int

func (s *Sequence) Next() Sequence {
	res := *s
	(*s)++
	return res
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
