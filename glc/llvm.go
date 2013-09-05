package main

import (
	"go/ast"
	"goal/lovm"
)

func (s *Scope) ParseLlvmType(typeName ast.Expr) lovm.Type {
	return s.ParseType(typeName).LlvmType()
}

func SymbolsToLlvmTypes(ss []Symbol) (res []lovm.Type) {
	for _, s := range ss {
		res = append(res, s.LlvmType())
	}
	return
}

func (s *Scope) ParseLlvmTypes(fl *ast.FieldList) (res []lovm.Type) {
	if fl == nil {
		return nil
	}
	for _, f := range fl.List {
		t := s.ParseLlvmType(f.Type)
		if f.Names == nil {
			res = append(res, t)
		} else {
			args := make([]lovm.Type, len(f.Names))
			for i := range f.Names {
				args[i] = t
			}
			res = append(res, args...)
		}
	}
	return
}
