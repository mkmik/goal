package lovm

import (
	"fmt"
	"io"
	"strings"
)

type Type interface {
	Name() string
	Dereference() Type
	EmitDecl(w io.Writer, name string)
	EmitDef(w io.Writer, name string, body func())
}

type BasicType struct {
	name string
	Base Type
}

func (b BasicType) Name() string {
	return b.name
}

func (b BasicType) Dereference() Type {
	if b.Base == nil {
		panic(fmt.Errorf("dereferencing a non reference type: %v", b.Name()))
	}
	return b.Base
}

func (b BasicType) EmitDecl(w io.Writer, name string) {
	fmt.Fprintf(w, "%s = external global %s\n", name, b.Name())
}

func (b BasicType) EmitDef(w io.Writer, name string, body func()) {
	fmt.Fprintf(w, "%s = global %s ", name, b.Name())
	body()
	fmt.Fprintf(w, "\n")
}

func IntType(size int) Type {
	return BasicType{fmt.Sprintf("i%d", size), nil}
}

type FuncType struct {
	ReturnType Type
	ParamTypes []Type
	Variadic   bool
}

func (f FuncType) Name() string {
	return f.funcDecl("")
}

func (b FuncType) Dereference() Type {
	panic("dereferencing a non reference type")
}

func (f FuncType) funcDecl(name string) string {
	args := make([]string, len(f.ParamTypes))
	for i, p := range f.ParamTypes {
		args[i] = p.Name()
	}
	if f.Variadic {
		args = append(args, "...")
	}

	return fmt.Sprintf("%s %s(%s)", f.ReturnType.Name(), name, strings.Join(args, ", "))
}

func (f FuncType) EmitDecl(w io.Writer, name string) {
	fmt.Fprintf(w, "declare %s\n", f.funcDecl(name))
}

func (f FuncType) EmitDef(w io.Writer, name string, body func()) {
	fmt.Fprintf(w, "define %s {\n", f.funcDecl(name))
	body()
	fmt.Fprintf(w, "}\n")
}

func FunctionType(ret Type, variadic bool, params ...Type) FuncType {
	return FuncType{
		ReturnType: ret,
		ParamTypes: params,
		Variadic:   variadic,
	}
}

func PointerType(typ Type) Type {
	return BasicType{fmt.Sprintf("%s *", typ.Name()), typ}
}

func ArrayType(typ Type, size int) Type {
	return BasicType{fmt.Sprintf("[%d x %s]", size, typ.Name()), PointerType(typ)}
}

func VoidType() Type {
	return BasicType{"void", nil}
}

func DereferenceTypes(base Type, indices ...int) Type {
	if len(indices) > 0 {
		return DereferenceTypes(base.Dereference(), indices[1:]...)
	} else {
		return base
	}
}
