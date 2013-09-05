package lovm

import (
	"fmt"
)

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
