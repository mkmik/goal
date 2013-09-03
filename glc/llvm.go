package main

import (
	"github.com/axw/gollvm/llvm"
	"go/ast"
)

func (s *Scope) ParseLlvmType(typeName ast.Expr) llvm.Type {
	return s.ParseType(typeName).LlvmType()
}

func SymbolsToLlvmTypes(ss []Symbol) (res []llvm.Type) {
	for _, s := range ss {
		res = append(res, s.LlvmType())
	}
	return
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
