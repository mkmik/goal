package main

import (
	"go/ast"
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

func (_ *implementsType) isType() {}
func (_ PrimitiveType) isType()   {}

type MapType struct {
	implementsType
	Key   Type
	Value Type
}

type SliceType struct {
	implementsType
	Value Type
}


func (s *Scope) ParseType(typeName ast.Expr) (Type, error) {
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
}
