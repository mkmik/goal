package main

import (
	"flag"
	"fmt"
	"github.com/axw/gollvm/llvm"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

type Symbol struct {
	Name  string
	Type  Type
	Value *llvm.Value
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
	FunctionType FunctionType
	Function     llvm.Value
	Builder      llvm.Builder
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
			llvm_func_type := llvm.FunctionType(func_ret_type, func_arg_types, false)
			llvmFunction := llvm.AddFunction(v.Module, n.Name.Name, llvm_func_type)

			functionType := v.ParseFuncType(n.Type)
			err := v.AddVar(Symbol{Name: n.Name.Name, Type: functionType, Value: &llvmFunction})
			if err != nil {
				log.Fatalf("cannot add symbol %#v: %s", n.Name.Name, err)
			}

			newScope := NewScope(&v.Scope)
			for i, p := range functionType.Params {
				if p.Name != "" {
					value := llvmFunction.Param(i)
					p.Value = &value
					err := newScope.AddVar(p)
					if err != nil {
						log.Fatalf("cannot add symbol %#v: %s", p, err)
					}
				}
			}

			if n.Body != nil {
				builder := llvm.NewBuilder()
				defer builder.Dispose()

				entry := llvm.AddBasicBlock(llvmFunction, "")
				builder.SetInsertPointAtEnd(entry)

				fv := &FunctionVisitor{functionType, llvmFunction, builder}
				ast.Walk(&BlockVisitor{newScope, fv, entry}, n.Body)
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

func (s *BlockVisitor) AddDecl(d ast.Decl) error {
	fmt.Printf("MY adding decl %#v\n", d)
	gen := d.(*ast.GenDecl)

	for _, sp := range gen.Specs {
		vs := sp.(*ast.ValueSpec)
		for idx, n := range vs.Names {
			typ := s.ParseType(vs.Type)
			var value llvm.Value
			if vs.Values != nil {
				ev := &ExpressionVisitor{s, llvm.Value{}, typ}
				ast.Walk(ev, vs.Values[idx])
				value = ev.Value
			}
			err := s.AddVar(Symbol{Name: n.Name, Type: typ, Value: &value})
			if err != nil {
				log.Fatalf("cannot add var %s: %s", n.Name, err)
			}
		}
	}
	return nil
}

func (s *Scope) ResolveSymbol(name string) Symbol {
	res, ok := s.Symbols[name]
	if !ok {
		log.Fatalf("cannot resolve symbol: %s", name)
	}
	return res
}

func (s *Scope) AddVar(variable Symbol) error {
	name := variable.Name
	if _, ok := s.Symbols[name]; ok {
		return fmt.Errorf("Multiple declarations of %s", name)
	}
	if variable.Value == nil {
		value := llvm.ConstInt(LlvmType(variable.Type), 0, false)
		variable.Value = &value
	}
	s.Symbols[name] = variable
	return nil
}

func (v *ExpressionVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		switch n := node.(type) {
		case *ast.ParenExpr:
			return v
		case *ast.BinaryExpr:
			xev := v.Evaluate(n.X)
			yev := v.Evaluate(n.Y)

			if xev.Type != yev.Type {
				log.Fatalf("Types %#v and %#v are not compatible (A)", xev.Type, yev.Type)
			} else if v.Type != xev.Type {
				log.Fatalf("Types %#v and %#v are not compatible (B)", v.Type, xev.Type)
			} else {
				fmt.Printf("MY BINARY TYPES: %#v, %#v, %#v\n", v.Type, xev.Type, yev.Type)
			}
			// types must match, thus take either one
			v.Type = xev.Type
			switch n.Op {
			case token.ADD:
				v.Value = v.Builder.CreateAdd(xev.Value, yev.Value, "")
			case token.SUB:
				v.Value = v.Builder.CreateSub(xev.Value, yev.Value, "")
			case token.MUL:
				v.Value = v.Builder.CreateMul(xev.Value, yev.Value, "")
			case token.QUO:
				v.Value = v.Builder.CreateSDiv(xev.Value, yev.Value, "")
			case token.REM:
				v.Value = v.Builder.CreateSRem(xev.Value, yev.Value, "")
			default:
				log.Fatalf("inimplemented binary operator %v", n.Op)
			}

			return nil
		case *ast.BasicLit:
			v.Value = llvm.ConstIntFromString(LlvmType(v.Type), n.Value, 10)
		case *ast.Ident:
			symbol := v.ResolveSymbol(n.Name)
			v.Type = symbol.Type
			v.Value = *symbol.Value
			return nil
		case *ast.CallExpr:
			fmt.Printf("MY EXPR CALL: %#v\n", n)
			if id, ok := n.Fun.(*ast.Ident); ok {
				if typ, err := v.ResolveType(id); err == nil {
					fmt.Printf("MY EXPR TYPE CONVERSION: %#v, %#v\n", n, typ)
					if len(n.Args) != 1 {
						log.Fatalf("type conversion can have only one argument")
					}
					ev := v.Evaluate(n.Args[0])
					// TODO(mkm) choose whether bitcast, trunc or sext
					//v.Value = v.Builder.CreateBitCast(ev.Value, LlvmType(typ), "")
					v.Value = v.Builder.CreateIntCast(ev.Value, LlvmType(typ), "")
				} else {
					fmt.Printf("MY NORMAL CALL %#v (err was: %v)\n", id, err)
				}
				return nil
			}
			log.Fatalf("Unimplemented call %#v", node)
		default:
			log.Fatalf("----- Expression visitor: UNKNOWN %#v\n", node)
			return v
		}
	}
	return nil
}

func (v *ExpressionVisitor) Evaluate(exp ast.Expr) *ExpressionVisitor {
	ev := *v
	ast.Walk(&ev, exp)
	return &ev
}

func (v *BlockVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		switch n := node.(type) {
		case *ast.ReturnStmt:
			functionReturnSymbols := v.FunctionType.Results
			if len(functionReturnSymbols) != len(n.Results) {
				log.Fatalf("too many/too few arguments to return")
			}

			values := make([]llvm.Value, len(n.Results))
			types := make([]llvm.Type, len(n.Results))

			for i, e := range n.Results {
				ev := &ExpressionVisitor{v, llvm.Value{}, functionReturnSymbols[i].Type}
				ast.Walk(ev, e)
				values[i] = ev.Value
				types[i] = LlvmType(ev.Type)
			}

			var res llvm.Value
			switch len(values) {
			case 1:
				res = values[0]
			default:
				log.Fatalf("unimplemented multiple return values")
			}
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
				if len(n.Lhs) != len(n.Rhs) {
					log.Fatalf("too many/too few expressions in assignment")
				}

				symbols := make([]Symbol, len(n.Lhs))
				for i, e := range n.Lhs {
					symbols[i] = v.ResolveSymbol(e.(*ast.Ident).Name)
				}

				values := make([]llvm.Value, len(n.Lhs))
				for i, e := range n.Rhs {
					ev := &ExpressionVisitor{v, llvm.Value{}, symbols[i].Type}
					ast.Walk(ev, e)
					values[i] = ev.Value
				}
				for i, sym := range symbols {
					*sym.Value = values[i]
				}
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
