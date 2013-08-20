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

func (v *ModuleVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		switch n := node.(type) {
		case *ast.FuncDecl:
			fmt.Printf("FUNC DECL %s: %#v\n", n.Name, n.Type)
			func_arg_types, err := v.ParseLlvmTypes(n.Type.Params)
			if err != nil {
				log.Fatal("Cannot args convert to llvm types: ", err)
			}
			func_ret_types, err := v.ParseLlvmTypes(n.Type.Results)
			if err != nil {
				log.Fatal("Cannot return convert to llvm types: ", err)
			}
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
			typ, err := s.ParseType(vs.Type)
			if err != nil {
				return err
			}
			s.AddVar(n.Name, Variable{typ})
		}
	}
	return nil
}

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

func (v *ExpressionVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		switch n := node.(type) {
		case *ast.BasicLit:
			fmt.Printf("MY LITERAL: %#v\n", n)
			llvmType, err := LlvmType(v.Type)
			if err != nil {
				log.Fatal(err)
			}
			val, err := strconv.ParseUint(n.Value, 10, 64)
			if err != nil {
				log.Fatal(err)
			}
			v.Value = llvm.ConstInt(llvmType, val, false)
		default:
			fmt.Printf("----- Function visitor: UNKNOWN %#v\n", node)
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
			var err error
			values := make([]llvm.Value, len(n.Results))
			types := make([]llvm.Type, len(n.Results))
			// TODO(mkm) fetch them from function delcaration
			functionReturnTypes := []Type{Int64}
			for i, e := range n.Results {
				ev := &ExpressionVisitor{v, llvm.Value{}, functionReturnTypes[i]}
				ast.Walk(ev, e)
				values[i] = ev.Value
				types[i], err = LlvmType(ev.Type)
				if err != nil {
					log.Fatal("evaluating return statement", err)
				}
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
