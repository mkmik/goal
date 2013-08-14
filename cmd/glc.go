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
	Int8 PrimitiveType = iota
	Int16
	Int32
	Int64
	Uint8
	Uint16
	Uint32
	Uint64
	String
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

// values

type Function struct {
}

type Value struct {
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

func (v *BlockVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		switch n := node.(type) {
		case *ast.ExprStmt:
			fmt.Printf("EXPR STATEMENT %#v\n", n)
		case *ast.DeclStmt:
			fmt.Printf("DECL STMT %#v\n", n)
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
	}
	return nil

}

func CompileFile(tree *ast.File) error {
	fmt.Printf("compiled %#v\n", tree)
	v := &ModuleVisitor{NewScope(nil)}
	ast.Walk(v, tree)

	DumpToFile(tree, "/tmp/ast")
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
