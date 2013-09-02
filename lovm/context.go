package lovm

import (
	"io"
)

type Context struct {
	Writer  io.Writer
	Modules []*Module
}

func NewContext(w io.Writer) Context {
	return Context{
		Writer: w,
	}
}

type External struct {
	Name string
	Type Type
}

type Module struct {
	*Context
	Functions []*Function
	Externals []External
}

func (ctx *Context) NewModule() *Module {
	mod := &Module{
		Context: ctx,
	}

	ctx.AddModule(mod)
	return mod
}

func (ctx *Context) AddModule(m *Module) {
	ctx.Modules = append(ctx.Modules, m)
}

func (ctx *Context) Emit() {
	for _, m := range ctx.Modules {
		m.Emit()
	}
}

func (mod *Module) NewFunction(name string, signature FuncType) *Function {
	fun := &Function{
		Module: mod,
		Values: map[Value]bool{},
		Type:   signature,
		Name:   name,
	}

	mod.AddFunction(fun)
	return fun
}

func (mod *Module) DeclareExternal(name string, signature Type) {
	mod.Externals = append(mod.Externals, External{name, signature})
}

func (mod *Module) AddFunction(f *Function) {
	mod.Functions = append(mod.Functions, f)
}

func (mod *Module) Emit() {
	for _, e := range mod.Externals {
		e.Type.EmitDecl(mod.Writer, e.Name)
	}
	for _, f := range mod.Functions {
		f.Emit()
	}
}
