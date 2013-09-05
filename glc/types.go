package main

import (
	"fmt"
	"go/ast"
	"goal/lovm"
	"goal/util"
)

var (
	PointerSize = 8
)

var (
	Any    = AnyType{}
	Int    = PrimitiveType{"int", true, lovm.IntType(32)}
	Int8   = PrimitiveType{"int8", true, lovm.IntType(8)}
	Int16  = PrimitiveType{"int16", true, lovm.IntType(16)}
	Int32  = PrimitiveType{"int32", true, lovm.IntType(32)}
	Int64  = PrimitiveType{"int64", true, lovm.IntType(64)}
	Uint   = PrimitiveType{"int", false, lovm.IntType(32)}
	Uint8  = PrimitiveType{"int8", false, lovm.IntType(8)}
	Uint16 = PrimitiveType{"int16", false, lovm.IntType(16)}
	Uint32 = PrimitiveType{"int32", false, lovm.IntType(32)}
	Uint64 = PrimitiveType{"int64", false, lovm.IntType(64)}
	Bool   = PrimitiveType{"bool", false, lovm.IntType(1)}
	String = PrimitiveType{"string", false, lovm.PointerType(lovm.IntType(8))}
	// TODO(mkm): should be an interface type
	Error = PrimitiveType{"error", false, lovm.PointerType(lovm.IntType(8))}
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
	LlvmType() lovm.Type
	//	String() string
}

type IntegerType interface {
	Size() int
}

type PrimitiveType struct {
	Name     string
	Signed   bool
	llvmType lovm.Type
}

type AnyType struct{}

func (b AnyType) LlvmType() lovm.Type {
	return lovm.VoidType()
}

func (b AnyType) String() string {
	return "Type(Any)"
}

func (b PrimitiveType) LlvmType() lovm.Type {
	return b.llvmType
}

func (b PrimitiveType) String() string {
	return fmt.Sprintf("Type(%s)", b.Name)
}

type MapType struct {
	Key   Type
	Value Type
}

func (b MapType) LlvmType() lovm.Type {
	return lovm.PointerType(lovm.IntType(8))
}

type SliceType struct {
	Value Type
}

func (b SliceType) LlvmType() lovm.Type {
	// TODO(mkm) use struct type
	return lovm.PointerType(lovm.IntType(8))
}

type FunctionType struct {
	Params  []Symbol
	Results []Symbol
	// TODO(mkm) receivers
}

func (s *Scope) ParseType(typeName ast.Expr) Type {
	res, err := s.ResolveType(typeName)
	if err != nil {
		util.Perrorf("%s", err)
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

func (t FunctionType) LlvmType() lovm.Type {
	func_arg_types := SymbolsToLlvmTypes(t.Params)
	func_ret_types := SymbolsToLlvmTypes(t.Results)

	var func_ret_type lovm.Type
	switch len(func_ret_types) {
	case 0:
		func_ret_type = lovm.VoidType()
	case 1:
		func_ret_type = func_ret_types[0]
	default:
		util.Perrorf("not migrated yet to lovm")
		//func_ret_type = lovm.StructType(func_ret_types, false)
	}
	return lovm.FunctionType(func_ret_type, false, func_arg_types...)
}
