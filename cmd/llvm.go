package main

import (
	"github.com/axw/gollvm/llvm"
	"go/ast"
	"log"
)

func LlvmType(t Type) llvm.Type {
	switch t := t.(type) {
	case PrimitiveType:
		switch t {
		case Int:
			return llvm.Int32Type()
		case Int8:
			return llvm.Int8Type()
		case Int16:
			return llvm.Int16Type()
		case Int32:
			return llvm.Int32Type()
		case Int64:
			return llvm.Int64Type()
		default:
			log.Fatalf("Cannot translate primitive type %#v to llvm type", t)
		}
	default:
		log.Fatalf("Cannot translate type %#v to llvm type", t)
	}
	// unreachable
	return llvm.Type{}
}

func (s *Scope) ParseLlvmType(typeName ast.Expr) llvm.Type {
	return LlvmType(s.ParseType(typeName))
}


func (s *Scope) ParseLlvmTypes(fl *ast.FieldList) (res []llvm.Type) {
	if fl == nil {
		return nil
	}
	for _, f := range fl.List {
		t := s.ParseLlvmType(f.Type)
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
