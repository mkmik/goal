package main

import (
	"go/ast"
	"log"
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
	switch t := typeName.(type) {
	case *ast.Ident:
		switch t.Name {
		case "int":
			return Int
		case "int8":
			return Int8
		case "int16":
			return Int16
		case "int32":
			return Int32
		case "int64":
			return Int64
		case "uint8":
			return Int8
		case "uint16":
			return Int16
		case "uint32":
			return Int32
		case "uint64":
			return Int64
		case "string":
			return String
		case "error":
			return Error
		default:
			log.Fatalf("unknown type: %s", t)
		}
	case *ast.SelectorExpr:
		log.Fatalf("NOT IMPLEMENTED YET: qualified type names")
	case *ast.MapType:
		log.Fatalf("NOT IMPLEMENTED YET: map type")
	case *ast.ArrayType:
		log.Fatalf("NOT IMPLEMENTED YET: array type")
	case *ast.ChanType:
		log.Fatalf("NOT IMPLEMENTED YET: chan type")
	default:
		log.Fatalf("unknown type class: %#v", typeName)
	}
	// unreachable
	return nil
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
