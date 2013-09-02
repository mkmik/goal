package lovm

import (
	"fmt"
	"strings"
	"io"
)

type Type interface {
	Name() string
	EmitDecl(w io.Writer, name string)
	EmitDef(w io.Writer, name string, body func())
}

type BasicType struct {
	name string
}

func (b BasicType) Name() string {
	return b.name
}

func (b BasicType) EmitDecl(w io.Writer, name string) {
	fmt.Fprintf(w, "%s = extern global %s\n", name, b.Name())
}

func (b BasicType) EmitDef(w io.Writer, name string, body func()) {
	fmt.Fprintf(w, "%s = global %s ")
	body()
	fmt.Fprintf(w, "\n")
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

func (f FuncType) funcDecl(name string) string {
	return fmt.Sprintf("%s %s(%s)", f.ReturnType, name, strings.Join(f.ParamTypes, ", "))
}

func (f FuncType) EmitDecl(w io.Writer, name string) {
	fmt.Fprintf(w, "declare %s\n", f.funcDecl(name))
}

func (f FuncType) EmitDef(w io.Writer, name string, body func()) {
	fmt.Fprintf(w, "define %s {\n", f.funcDecl(name))
	body()
	fmt.Fprintf(w, "}\n")
}

func FunctionType(ret Type, params ...Type) FuncType {
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
