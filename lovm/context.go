package lovm

import (
	"goal/util"
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

type Global struct {
	Name  string
	Type  Type
	Attrs []string
	Init  Constant
}

type Constant interface {
	Emit(w io.Writer)
}

type Module struct {
	*Context
	Name      string
	Functions []*Function
	Externals []External
	Globals   []Global
	Interned  util.Sequence
}

func (ctx *Context) NewModule(name string) *Module {
	mod := &Module{
		Context: ctx,
		Name:    name,
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

func (mod *Module) DeclareExternal(name string, signature Type) SymRef {
	mod.Externals = append(mod.Externals, External{name, signature})
	return SymRef{name, PointerType(signature)}
}

func (mod *Module) AddFunction(f *Function) {
	mod.Functions = append(mod.Functions, f)
}

func (mod *Module) Emit() {
	for _, e := range mod.Externals {
		e.Type.EmitDecl(mod.Writer, e.Name)
	}
	for _, g := range mod.Globals {
		g.Type.EmitDef(mod.Writer, g.Name, func() {
			g.Init.Emit(mod.Writer)
		})
	}
	for _, f := range mod.Functions {
		f.Emit()
	}
}
