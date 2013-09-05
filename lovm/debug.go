package lovm

import (
	"fmt"
	"strings"
)

func assertNotNil(args ...interface{}) {
	var failedArgs []string
	for i, a := range args {
		if a == nil {
			failedArgs = append(failedArgs, fmt.Sprintf("%d", i))
		}
	}
	if failedArgs != nil {
		panic(fmt.Errorf("runtime error: Nil assertion failed. Failed args: %s",
			strings.Join(failedArgs, ", ")))
	}
}

type DebugInstr struct {
	Source string
}

func DebugInstrf(format string, args ...interface{}) DebugInstr {
	return DebugInstr{fmt.Sprintf(format, args...)}
}

func (d DebugInstr) Name() string {
	return "%debuginstr"
}

func (d DebugInstr) Emit(ctx *Function) {
	ctx.Emitf("%s\n", d.Source)
}
