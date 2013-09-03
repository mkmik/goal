package main

import (
	"fmt"
	"github.com/axw/gollvm/llvm"
	"go/ast"
)

var (
	PointerSize = 8
)

var (
	Any    = AnyType{}
	Int    = PrimitiveType{"int", true, llvm.Int32Type()}
	Int8   = PrimitiveType{"int8", true, llvm.Int8Type()}
	Int16  = PrimitiveType{"int16", true, llvm.Int16Type()}
	Int32  = PrimitiveType{"int32", true, llvm.Int32Type()}
	Int64  = PrimitiveType{"int64", true, llvm.Int64Type()}
	Uint   = PrimitiveType{"int", false, llvm.Int32Type()}
	Uint8  = PrimitiveType{"int8", false, llvm.Int8Type()}
	Uint16 = PrimitiveType{"int16", false, llvm.Int16Type()}
	Uint32 = PrimitiveType{"int32", false, llvm.Int32Type()}
	Uint64 = PrimitiveType{"int64", false, llvm.Int64Type()}
	Bool   = PrimitiveType{"bool", false, llvm.Int1Type()}
	String = PrimitiveType{"string", false, llvm.PointerType(llvm.Int8Type(), 0)}
	// TODO(mkm): should be an interface type
	Error = PrimitiveType{"error", false, llvm.PointerType(llvm.Int8Type(), 0)}
)

var (
	primitiveTypes = []PrimitiveType{
		Int,
		Int8,
		Int16,
		Int32,
		Int64,
		Uint,
		Uint8,
		Uint16,
		Uint32,
		Uint64,
		Bool,
		String,
		Error,
	}
	primitiveTypeByName = make(map[string]Type)
)

func init() {
	for _, t := range primitiveTypes {
		primitiveTypeByName[t.Name] = t
	}
}

type Type interface {
	LlvmType() llvm.Type
	//	String() string
}

type IntegerType interface {
	Size() int
}

type PrimitiveType struct {
	Name     string
	Signed   bool
	llvmType llvm.Type
}

type AnyType struct{}

func (b AnyType) LlvmType() llvm.Type {
	return llvm.VoidType()
}

func (b AnyType) String() string {
	return "Type(Any)"
}

func (b PrimitiveType) LlvmType() llvm.Type {
	return b.llvmType
}

func (b PrimitiveType) String() string {
	return fmt.Sprintf("Type(%s)", b.Name)
}

type MapType struct {
	Key   Type
	Value Type
}

func (b MapType) LlvmType() llvm.Type {
	return llvm.PointerType(llvm.Int8Type(), 0)
}

type SliceType struct {
	Value Type
}

func (b SliceType) LlvmType() llvm.Type {
	// TODO(mkm) use struct type
	return llvm.PointerType(llvm.Int8Type(), 0)
}

type FunctionType struct {
	Params  []Symbol
	Results []Symbol
	// TODO(mkm) receivers
}

func (s *Scope) ParseType(typeName ast.Expr) Type {
	res, err := s.ResolveType(typeName)
	if err != nil {
		Perrorf("%s", err)
	}
	return res
}

func (s *Scope) ResolveType(typeName ast.Expr) (Type, error) {
	switch t := typeName.(type) {
	case *ast.Ident:
		if primitive, ok := primitiveTypeByName[t.Name]; ok {
			return primitive, nil
		} else {
			return nil, fmt.Errorf("unknown type: %s", t)
		}
	case *ast.SelectorExpr:
		return nil, fmt.Errorf("NOT IMPLEMENTED YET: qualified type names")
	case *ast.MapType:
		return nil, fmt.Errorf("NOT IMPLEMENTED YET: map type")
	case *ast.ArrayType:
		return nil, fmt.Errorf("NOT IMPLEMENTED YET: array type")
	case *ast.ChanType:
		return nil, fmt.Errorf("NOT IMPLEMENTED YET: chan type")
	default:
		return nil, fmt.Errorf("runtime error: unknown type class: %#v", typeName)
	}
	// unreachable
	return nil, nil
}

func (s *Scope) ParseSymbols(fl *ast.FieldList) (res []Symbol) {
	if fl == nil {
		return nil
	}
	for _, f := range fl.List {
		t := s.ParseType(f.Type)
		if f.Names == nil {
			res = append(res, Symbol{Name: "", Type: t})
		} else {
			args := make([]Symbol, len(f.Names))
			for i, n := range f.Names {
				args[i] = Symbol{Name: n.Name, Type: t}
			}
			res = append(res, args...)
		}
	}
	return
}

func (s *Scope) ParseTypes(fl *ast.FieldList) (types []Type) {
	for _, s := range s.ParseSymbols(fl) {
		types = append(types, s.Type)
	}
	return
}

func (s *Scope) ParseFuncType(ft *ast.FuncType) FunctionType {
	return FunctionType{
		Params:  s.ParseSymbols(ft.Params),
		Results: s.ParseSymbols(ft.Results),
	}
}

func (t FunctionType) LlvmType() llvm.Type {
	func_arg_types := SymbolsToLlvmTypes(t.Params)
	func_ret_types := SymbolsToLlvmTypes(t.Results)

	var func_ret_type llvm.Type
	switch len(func_ret_types) {
	case 0:
		func_ret_type = llvm.VoidType()
	case 1:
		func_ret_type = func_ret_types[0]
	default:
		func_ret_type = llvm.StructType(func_ret_types, false)
	}
	return llvm.FunctionType(func_ret_type, func_arg_types, false)
}