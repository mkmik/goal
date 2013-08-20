package main

import (
	"flag"
	"fmt"
	"github.com/axw/gollvm/llvm"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strconv"
)

type Symbol struct {
	Name string
	Type Type
	Value llvm.Value
}

//
type SymbolMap map[string]Symbol

// visitors

type Scope struct {
	Symbols SymbolMap
	Parent  *Scope
}

func NewScope(parent *Scope) Scope {
	return Scope{map[string]Symbol{}, parent}
}

type ModuleVisitor struct {
	Scope
	Module llvm.Module
}

// contains common state shared accross the function
type FunctionVisitor struct {
	Function llvm.Value
	Builder  llvm.Builder
}

// contains scope local to a block
type BlockVisitor struct {
	Scope
	*FunctionVisitor
	Block llvm.BasicBlock
}

type ExpressionVisitor struct {
	*BlockVisitor
	// result of expression
	Value llvm.Value
	Type  Type
}

func (v *ModuleVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		switch n := node.(type) {
		case *ast.FuncDecl:
			fmt.Printf("FUNC DECL %s: %#v\n", n.Name, n.Type)
			func_arg_types := v.ParseLlvmTypes(n.Type.Params)
			func_ret_types := v.ParseLlvmTypes(n.Type.Results)
			var func_ret_type llvm.Type
			switch len(func_ret_types) {
			case 0:
				func_ret_type = llvm.VoidType()
			case 1:
				func_ret_type = func_ret_types[0]
			default:
				func_ret_type = llvm.StructType(func_ret_types, false)
			}
			func_type := llvm.FunctionType(func_ret_type, func_arg_types, false)
			llvmFunction := llvm.AddFunction(v.Module, n.Name.Name, func_type)
			if n.Body != nil {
				builder := llvm.NewBuilder()
				defer builder.Dispose()

				entry := llvm.AddBasicBlock(llvmFunction, "")
				builder.SetInsertPointAtEnd(entry)

				fv := &FunctionVisitor{llvmFunction, builder}
				ast.Walk(&BlockVisitor{NewScope(&v.Scope), fv, entry}, n.Body)
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
			typ := s.ParseType(vs.Type)
			s.AddVar(n.Name, Symbol{Name: n.Name, Type: typ})
		}
	}
	return nil
}


func (s *Scope) AddVar(name string, variable Symbol) error {
	if _, ok := s.Symbols[name]; ok {
		return fmt.Errorf("Multiple declarations of %s", name)
	}
	s.Symbols[name] = variable
	return nil
}

func (v *ExpressionVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		switch n := node.(type) {
		case *ast.BinaryExpr:
			fmt.Printf("MY BINARY: %#v\n", n.Y)
			xev := *v
			yev := *v
			ast.Walk(&xev, n.X)
			ast.Walk(&yev, n.Y)

			// types must match, thus take either one
			v.Type = xev.Type
			switch n.Op {
			case token.ADD:
				fmt.Printf("MY ADDING!\n")
				v.Value = v.Builder.CreateAdd(xev.Value, yev.Value, "")
			default:
				log.Fatalf("inimplemented binary operator %v", n.Op)
			}
			
			return nil
		case *ast.BasicLit:
			fmt.Printf("MY LITERAL: %#v\n", n)
			llvmType := LlvmType(v.Type)
			val, err := strconv.ParseUint(n.Value, 10, 64)
			if err != nil {
				log.Fatal(err)
			}
			v.Value = llvm.ConstInt(llvmType, val, false)
		case *ast.Ident:
			fmt.Printf("MY EXPR IDENT: %#v\n", n)
			v.Value = v.Function.Param(2)
			return nil
		default:
			log.Fatalf("----- Function visitor: UNKNOWN %#v\n", node)
			return v
		}
	}
	return nil
}

func (v *BlockVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		switch n := node.(type) {
		case *ast.ReturnStmt:
			fmt.Printf("MY EXPR %#v\n", n.Results[0])
			values := make([]llvm.Value, len(n.Results))
			types := make([]llvm.Type, len(n.Results))
			// TODO(mkm) fetch them from function delcaration
			functionReturnTypes := []Type{Int64}
			for i, e := range n.Results {
				ev := &ExpressionVisitor{v, llvm.Value{}, functionReturnTypes[i]}
				ast.Walk(ev, e)
				values[i] = ev.Value
				types[i] = LlvmType(ev.Type)
			}
			res := values[0]
			v.Builder.CreateRet(res)
		case *ast.ExprStmt:
			log.Fatalf("NOT IMPLEMENTED YET: expression statements")
		case *ast.DeclStmt:
			err := v.AddDecl(n.Decl)
			if err != nil {
				log.Fatal("syntax error:", err)
			}
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE {
				log.Fatalf("NOT IMPLEMENTED YET: type inference in var decl")
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
	v := &ModuleVisitor{NewScope(nil), llvm.NewModule(tree.Name.Name)}
	ast.Walk(v, tree)

	fmt.Printf("LLVM: -----------\n")
	v.Module.Dump()
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
