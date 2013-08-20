package main

import (
	"github.com/axw/gollvm/llvm"
	"go/ast"
	"fmt"
)

func LlvmType(t Type) (llvm.Type, error) {
	switch t := t.(type) {
	case PrimitiveType:
		switch t {
		case Int:
			return llvm.Int32Type(), nil
		case Int8:
			return llvm.Int8Type(), nil
		case Int16:
			return llvm.Int16Type(), nil
		case Int32:
			return llvm.Int32Type(), nil
		case Int64:
			return llvm.Int64Type(), nil
		default:
			return llvm.Type{}, fmt.Errorf("Cannot translate primitive type %#v to llvm type", t)
		}
	default:
		return llvm.Type{}, fmt.Errorf("Cannot translate type %#v to llvm type", t)
	}
}

func (s *Scope) ParseLlvmType(typeName ast.Expr) (llvm.Type, error) {
	t, err := s.ParseType(typeName)
	if err != nil {
		return llvm.Type{}, err
	}
	return LlvmType(t)
}


func (s *Scope) ParseLlvmTypes(fl *ast.FieldList) (res []llvm.Type, err error) {
	if fl == nil {
		return nil, nil
	}
	for _, f := range fl.List {
		t, err := s.ParseLlvmType(f.Type)
		if err != nil {
			return nil, err
		}
		if f.Names == nil {
			res = append(res, t)
		} else {
			args := make([]llvm.Type, len(f.Names))
			for i := range f.Names {
				args[i] = t
			}
			res = append(res, args...)
		}
	}
	return
}
