package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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

type Symbol interface {
}

// types

type Type interface {
}

type PrimitiveType uint

type MapType struct {
	Key   Type
	Value Type
}

type SliceType struct {
	Value Type
}

//
type SymbolMap map[string]Symbol

// variables

type Function struct {
}

type Variable struct {
	typ Type
}

// visitors

type Scope struct {
	Symbols SymbolMap
	Parent  *Scope
}

func NewScope(parent *Scope) *Scope {
	return &Scope{map[string]Symbol{}, parent}
}

type ModuleVisitor struct {
	*Scope
}

// contains common state shared accross the function
type FunctionVisitor struct {
}

// contains scope local to a block
type BlockVisitor struct {
	*Scope
	*FunctionVisitor
}

func (v *ModuleVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		switch n := node.(type) {
		case *ast.FuncDecl:
			fmt.Printf("FUNC DECL %s: %#v\n", n.Name, n)
			if n.Body != nil {
				fv := &FunctionVisitor{}
				ast.Walk(&BlockVisitor{NewScope(v.Scope), fv}, n.Body)
			}
			return nil
		case *ast.DeclStmt:
			fmt.Printf("DECL STMT %#v\n", n)
		default:
			fmt.Printf("----- Module visitor: UNKNOWN %#v\n", node)
			return v
		}
	} else {
		//		fmt.Printf("popping\n")
	}
	return nil
}

func (s *Scope) AddDecl(d ast.Decl) error {
	gen := d.(*ast.GenDecl)

	for _, sp := range gen.Specs {
		vs := sp.(*ast.ValueSpec)
		for idx, n := range vs.Names {
			fmt.Printf("----------------------- var %s %#v = %#v \n", n, vs.Type, vs.Values[idx])
			typ, err := s.ParseType(vs.Type)
			if err != nil {
				return err
			}
			s.AddVar(n.Name, Variable{typ})
		}
	}
	return nil
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

func (s *Scope) AddVar(name string, variable Variable) error {
	if _, ok := s.Symbols[name]; ok {
		return fmt.Errorf("Multiple declarations of %s", name)
	}
	s.Symbols[name] = variable
	return nil
}

func (v *BlockVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		switch n := node.(type) {
		case *ast.ExprStmt:
			fmt.Printf("EXPR STATEMENT %#v\n", n)
		case *ast.DeclStmt:
			err := v.AddDecl(n.Decl)
			if err != nil {
				log.Fatal("syntax error:", err)
			}
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE {
				fmt.Printf("DEFINE ASSIGN STMT %#v\n", n)
			} else {
				fmt.Printf("PLAIN ASSIGN STMT %#v ... %#v\n", n, n.Lhs[0])
			}
		default:
			fmt.Printf("----- Function visitor: UNKNOWN %#v\n", node)
			return v
		}
	} else {
		//		fmt.Printf("popping\n")
		v.DumpScope()
	}
	return nil

}

func CompileFile(tree *ast.File) error {
	DumpToFile(tree, "/tmp/ast")

	fmt.Printf("compiling %#v\n", tree)
	v := &ModuleVisitor{NewScope(nil)}
	ast.Walk(v, tree)
	return nil
}

func OpenAndCompileFile(name string) error {
	var fset token.FileSet
	ast, err := parser.ParseFile(&fset, name, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	err = CompileFile(ast)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	files := flag.Args()
	fmt.Println("test", files)

	for _, name := range files {
		err := OpenAndCompileFile(name)
		if err != nil {
			log.Fatal("Error compiling", err)
		}
	}
}
