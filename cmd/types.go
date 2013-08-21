package main

import (
	"go/ast"
	"log"
	"fmt"
)

const (
	Int PrimitiveType = iota
	Int8
	Int16
	Int32
	Int64
	Uint8
	Uint16
	Uint32
	Uint64
	String
	Error // TODO(mkm) should be builtin interface type
)

type Type interface {
	isType()
}

type PrimitiveType uint

// All concrete types embed ImplementsType which
// ensures that all types implement the Type interface.
type implementsType struct{}

func (_ implementsType) isType() {}
func (_ PrimitiveType) isType()  {}

type MapType struct {
	implementsType
	Key   Type
	Value Type
}

type SliceType struct {
	implementsType
	Value Type
}

type FunctionType struct {
	implementsType
	Params  []Symbol
	Results []Symbol
	// TODO(mkm) receivers
}

func (s *Scope) ParseType(typeName ast.Expr) Type {
	res, err := s.ResolveType(typeName)
	if err!= nil {
		log.Fatalf("%s", err)
	}
	return res
}

func (s *Scope) ResolveType(typeName ast.Expr) (Type, error) {
	switch t := typeName.(type) {
	case *ast.Ident:
		switch t.Name {
		case "int":
			return Int, nil
		case "int8":
			return Int8, nil
		case "int16":
			return Int16, nil
		case "int32":
			return Int32, nil
		case "int64":
			return Int64, nil
		case "uint8":
			return Int8, nil
		case "uint16":
			return Int16, nil
		case "uint32":
			return Int32, nil
		case "uint64":
			return Int64, nil
		case "string":
			return String, nil
		case "error":
			return Error, nil
		default:
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
		return nil, fmt.Errorf("unknown type class: %#v", typeName)
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
