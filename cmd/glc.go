package main

import (
	"flag"
	"fmt"
	"github.com/axw/gollvm/llvm"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"runtime/debug"
	"strings"
)

var (
	optimize = flag.Bool("optimize", false, "true to enable llvm optimization passes")
)

type Symbol struct {
	Name  string
	Type  Type
	Value *llvm.Value
	// debug
	inherited bool
}

func (s Symbol) LlvmType() llvm.Type {
	return s.Type.LlvmType()
}

//
type SymbolMap map[string]Symbol

// visitors
type Scope struct {
	*token.FileSet
	Symbols       SymbolMap
	ParentSymbols SymbolMap
}

func (s Scope) GetScope() Scope {
	return s
}

func NewScope(parent *Scope) Scope {
	return NewFileSetScope(parent.FileSet, parent)
}

func NewFileSetScope(fset *token.FileSet, parent *Scope) Scope {
	parentSymbols := SymbolMap{}
	if parent != nil {
		parentSymbols = MergeSymbolMaps(parent.ParentSymbols, parent.Symbols)
	}
	return Scope{fset, SymbolMap{}, parentSymbols}
}

func MergeSymbolMaps(maps ...SymbolMap) SymbolMap {
	//fmt.Printf("MERGING SYMBOL MAPS %#v\n", maps)
	res := SymbolMap{}
	for _, m := range maps {
		for n, s := range m {
			v := *s.Value
			s.Value = &v
			res[n] = s
		}
	}
	return res
}

type Visitor interface {
	ast.Visitor
	GetScope() Scope
}

type SkipRoot struct {
	Visitor
}

func (s SkipRoot) Visit(ast.Node) ast.Visitor {
	return s.Visitor
}

func Walk(visitor Visitor, node ast.Node) {
	defer func() {
		if err := recover(); err != nil {
			switch e := err.(type) {
			case error:
				if strings.HasPrefix(e.Error(), "runtime error:") {
					fmt.Fprintf(os.Stderr, "%s\n", e)
					debug.PrintStack()
					os.Exit(1)
				}
			}
			fmt.Fprintf(os.Stderr, "%s: %s\n", visitor.GetScope().Position(node.Pos()), err)
			os.Exit(1)
		}
	}()

	ast.Walk(visitor, node)
}

func Perrorf(format string, args ...interface{}) {
	panic(fmt.Errorf(format, args...))
}

type ModuleVisitor struct {
	Scope
	Module      llvm.Module
	PackageName string
}

// contains common state shared accross the function
type FunctionVisitor struct {
	*ModuleVisitor
	*FunctionVisitor
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
			functionType := v.ParseFuncType(n.Type)
			llvmFunction := llvm.AddFunction(v.Module, n.Name.Name, functionType.LlvmType())
			if err := v.AddVar(Symbol{Name: n.Name.Name, Type: functionType, Value: &llvmFunction}); err != nil {
				Perrorf("cannot add symbol %#v: %s", n.Name.Name, err)
			}

			newScope := NewScope(&v.Scope)
			for i, p := range functionType.Params {
				if p.Name != "" {
					value := llvmFunction.Param(i)
					p.Value = &value
					if err := newScope.AddVar(p); err != nil {
						Perrorf("cannot add symbol %#v: %s", p, err)
					}
				}
			}

			if n.Body != nil {
				builder := llvm.NewBuilder()
				defer builder.Dispose()

				entry := llvm.AddBasicBlock(llvmFunction, "entry")
				builder.SetInsertPointAtEnd(entry)

				fv := &FunctionVisitor{v, nil, functionType, llvmFunction, builder}
				bv := &BlockVisitor{newScope, fv, entry}
				Walk(SkipRoot{bv}, n.Body)
			}
			return nil
		case *ast.DeclStmt:
			Perrorf("Unimplemented decl stmt")
		case *ast.File:
			// root element, nothing to do
			return v
		case *ast.Ident:
			if v.PackageName != "" {
				Perrorf("Cannot have more than one top level ast.Ident")
			}
			v.PackageName = n.Name
			return nil
		case *ast.GenDecl:
			switch n.Tok {
			case token.IMPORT:
				// ignore imports for now
			default:
				Perrorf("UNIMPLEMENTED UNKNOWN GENDECL: %#v", node)
			}
		default:
			Perrorf("-----: Module visitor: UNKNOWN %#v\n", node)
			return v
		}
	} else {
		//		fmt.Printf("popping\n")
	}
	return nil
}

func (s *BlockVisitor) AddDecl(d ast.Decl) error {
	gen := d.(*ast.GenDecl)

	for _, sp := range gen.Specs {
		vs := sp.(*ast.ValueSpec)
		for idx, n := range vs.Names {
			if vs.Type == nil {
				Perrorf("cannot declare a var without a type")
			}
			typ := s.ParseType(vs.Type)
			var value llvm.Value
			if vs.Values != nil {
				ev := &ExpressionVisitor{s, llvm.Value{}, typ}
				Walk(ev, vs.Values[idx])
				value = ev.Value
			}
			if err := s.AddVar(Symbol{Name: n.Name, Type: typ, Value: &value}); err != nil {
				Perrorf("cannot add var %s: %s", n.Name, err)
			}
		}
	}
	return nil
}

func (s *Scope) ResolveSymbol(name string) Symbol {
	if res, ok := s.Symbols[name]; ok {
		return res
	}
	if res, ok := s.ParentSymbols[name]; ok {
		return res
	}

	Perrorf("cannot resolve symbol: %s", name)
	return Symbol{}
}

func (s *Scope) AddVar(variable Symbol) error {
	name := variable.Name
	if _, ok := s.Symbols[name]; ok {
		return fmt.Errorf("Multiple declarations of %s", name)
	}
	if variable.Value == nil {
		value := llvm.ConstInt(variable.Type.LlvmType(), 0, false)
		variable.Value = &value
	}
	s.Symbols[name] = variable
	return nil
}

func (v *ExpressionVisitor) IsConst(node ast.Node) bool {
	switch e := node.(type) {
	case *ast.BasicLit:
		return true
	case *ast.ParenExpr:
		return v.IsConst(e.X)
	default:
		return false
	}
}

func (v *ExpressionVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		switch n := node.(type) {
		case *ast.ParenExpr:
			return v
		case *ast.BinaryExpr:
			if v.IsConst(n.X) && v.IsConst(n.Y) {
				Perrorf("expression with only constants not implemented yet...")
			}

			var xev, yev *ExpressionVisitor
			if v.IsConst(n.X) {
				yev = v.Evaluate(n.Y)
				v.Type = yev.Type
				xev = v.Evaluate(n.X)
			} else {
				xev = v.Evaluate(n.X)
				v.Type = xev.Type
				yev = v.Evaluate(n.Y)
			}

			if xev.Type == Any {
				xev.Type = yev.Type
			}
			if yev.Type == Any {
				yev.Type = xev.Type
			}

			if xev.Type != yev.Type {
				Perrorf("Types %#v and %#v are not compatible (A)", xev.Type, yev.Type)
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
			case token.LSS:
				v.Value = v.Builder.CreateICmp(llvm.IntSLT, xev.Value, yev.Value, "")
				v.Type = Bool
			case token.GTR:
				v.Value = v.Builder.CreateICmp(llvm.IntSGT, xev.Value, yev.Value, "")
				v.Type = Bool
			default:
				Perrorf("inimplemented binary operator %v", n.Op)
			}

			return nil
		case *ast.BasicLit:
			if v.Type == nil {
				// panic(fmt.Errorf("runtime error: constant without type info")
				v.Type = Int
				fmt.Printf("XXXXXXXXXXXXXXXXXXXXXXXX %#v", v.Type)
			}

			if v.Type == Bool {
				Perrorf("Boolean arithmetic is not allowed: %#v", n)
			}
			v.Value = llvm.ConstIntFromString(v.Type.LlvmType(), n.Value, 10)
		case *ast.Ident:
			if n.Name == "true" {
				v.Type = Bool
				v.Value = llvm.ConstInt(v.Type.LlvmType(), 1, false)
			} else if n.Name == "false" {
				v.Type = Bool
				v.Value = llvm.ConstInt(v.Type.LlvmType(), 0, false)
			} else {
				symbol := v.ResolveSymbol(n.Name)
				v.Type = symbol.Type
				v.Value = *symbol.Value
			}
			return nil
		case *ast.CallExpr:
			if id, ok := n.Fun.(*ast.Ident); ok {
				if typ, err := v.ResolveType(id); err == nil {
					if len(n.Args) != 1 {
						Perrorf("type conversion can have only one argument")
					}
					ev := v.Evaluate(n.Args[0])
					v.Value = v.Builder.CreateIntCast(ev.Value, typ.LlvmType(), "")
				} else {
					Perrorf("UNIMPLEMENTED FUNCTION CALL %#v (err was: %v)\n", id, err)
				}
				return nil
			}
			Perrorf("Unimplemented call %#v", node)
		default:
			Perrorf("----- Expression visitor: UNKNOWN %#v\n", node)
			return v
		}
	}
	return nil
}

func (v *ExpressionVisitor) Evaluate(exp ast.Expr) *ExpressionVisitor {
	ev := *v
	Walk(&ev, exp)
	return &ev
}

func (v *BlockVisitor) Evaluate(typ Type, exp ast.Expr) *ExpressionVisitor {
	ev := &ExpressionVisitor{v, llvm.Value{}, typ}
	Walk(ev, exp)
	return ev
}

func (v *BlockVisitor) EvaluateBlock(exp ast.Stmt) *BlockVisitor {
	newScope := NewScope(&v.Scope)
	bv := &BlockVisitor{newScope, v.FunctionVisitor, llvm.BasicBlock{}}
	Walk(SkipRoot{bv}, exp)
	return bv
}

func (v *BlockVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		switch n := node.(type) {
		case *ast.ReturnStmt:
			functionReturnSymbols := v.FunctionType.Results
			if len(functionReturnSymbols) != len(n.Results) {
				Perrorf("too many/too few arguments to return")
			}

			values := make([]llvm.Value, len(n.Results))
			types := make([]llvm.Type, len(n.Results))

			for i, e := range n.Results {
				ev := &ExpressionVisitor{v, llvm.Value{}, functionReturnSymbols[i].Type}
				Walk(ev, e)
				values[i] = ev.Value
				types[i] = ev.Type.LlvmType()
			}

			var res llvm.Value
			switch len(values) {
			case 1:
				res = values[0]
			default:
				Perrorf("unimplemented multiple return values")
			}
			v.Builder.CreateRet(res)
		case *ast.ExprStmt:
			ev := &ExpressionVisitor{v, llvm.Value{}, Any}
			Walk(ev, n.X)

			//Perrorf("NOT IMPLEMENTED YET: expression statements")
		case *ast.DeclStmt:
			err := v.AddDecl(n.Decl)
			if err != nil {
				log.Fatal("syntax error:", err)
			}
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE {
				Perrorf("NOT IMPLEMENTED YET: type inference in var decl")
			} else {
				if len(n.Lhs) != len(n.Rhs) {
					Perrorf("too many/too few expressions in assignment")
				}

				symbols := make([]Symbol, len(n.Lhs))
				for i, e := range n.Lhs {
					symbols[i] = v.ResolveSymbol(e.(*ast.Ident).Name)
				}

				values := make([]llvm.Value, len(n.Lhs))
				for i, e := range n.Rhs {
					ev := &ExpressionVisitor{v, llvm.Value{}, symbols[i].Type}
					Walk(ev, e)
					values[i] = ev.Value
				}
				for i, sym := range symbols {
					fmt.Printf("UPDATING SCOPE, var:%s, value was: %v, now: %v\n", sym.Name, *sym.Value, values[i])
					*sym.Value = values[i]
				}
			}
		case *ast.IfStmt:
			cond := v.Evaluate(Bool, n.Cond)
			entry := v.Builder.GetInsertBlock()
			iftrue := llvm.AddBasicBlock(v.Function, "iftrue")
			endif := llvm.AddBasicBlock(v.Function, "endif")

			target := endif
			if n.Else != nil {
				iffalse := llvm.AddBasicBlock(v.Function, "iffalse")
				target = iffalse
			}
			v.Builder.CreateCondBr(cond.Value, iftrue, target)

			v.Builder.SetInsertPointAtEnd(iftrue)
			end := v.Builder.CreateBr(endif)
			v.Builder.SetInsertPointBefore(end)
			ifTrueVisitor := v.EvaluateBlock(n.Body)

			var ifFalseVisitor *BlockVisitor
			if n.Else != nil {
				v.Builder.SetInsertPointAtEnd(target)
				end = v.Builder.CreateBr(endif)
				v.Builder.SetInsertPointBefore(end)

				ifFalseVisitor = v.EvaluateBlock(n.Else)
			}

			v.Builder.SetInsertPointAtEnd(endif)
			fmt.Printf("inserting PHI, \nbase: %v\n", v.Symbols)
			fmt.Printf("iftrue: %v\n", ifTrueVisitor.Symbols)
			if ifFalseVisitor != nil {
				fmt.Printf("iffalse: %v\n", ifFalseVisitor.Symbols)
			}

			fmt.Println("UPDATED VARIABLES IN IF TRUE CHILD BLOCKS:")
			ForUpdatedVars(v.Scope, ifTrueVisitor.Scope, func(a, b Symbol) {
				fmt.Println("SYM!!!!!!!!!!!!", a.Name, a.Value, b.Value)

				phi := v.Builder.CreatePHI(a.Type.LlvmType(), "")
				phiVals := []llvm.Value{*a.Value, *b.Value}
				phiBlocks := []llvm.BasicBlock{entry, iftrue}
				phi.AddIncoming(phiVals, phiBlocks)

				*a.Value = phi
			})
			if ifFalseVisitor != nil {
				fmt.Println("UPDATED VARIABLES IN IF FALSE BLOCKS:")
				ForUpdatedVars(v.Scope, ifFalseVisitor.Scope, func(a, b Symbol) {
					fmt.Println("SYM!!!!!!!!!!!!", a.Name, a.Value, b.Value)

					phi := v.Builder.CreatePHI(a.Type.LlvmType(), "")
					phiVals := []llvm.Value{*a.Value, *b.Value}
					phiBlocks := []llvm.BasicBlock{entry, target}
					phi.AddIncoming(phiVals, phiBlocks)

					*a.Value = phi
				})
			}

			//Perrorf("Unimplemented if statement %#v\n", node)
		default:
			Perrorf("----- Block visitor: UNKNOWN %#v\n", node)
			return v
		}
	} else {
		//		fmt.Printf("popping\n")
		v.DumpScope()
	}
	return nil
}

func ForUpdatedVars(parent, child Scope, fn func(a, b Symbol)) {
	for n, s := range parent.Symbols {
		if cs, ok := child.ParentSymbols[n]; ok {
			if *s.Value != *cs.Value {
				fn(s, cs)
			}
		}
	}
}

func CompileFile(fset *token.FileSet, tree *ast.File) error {
	DumpToFile(tree, "/tmp/ast")

	v := &ModuleVisitor{NewFileSetScope(fset, nil), llvm.NewModule(tree.Name.Name), ""}
	Walk(v, tree)

	fmt.Printf("LLVM: -----------\n")

	if *optimize {
		Optimize(v.Module)
	}
	v.Module.Dump()
	return nil
}

func Optimize(mod llvm.Module) {
	pass := llvm.NewPassManager()
	defer pass.Dispose()
	pass.AddConstantPropagationPass()
	pass.AddInstructionCombiningPass()
	pass.AddPromoteMemoryToRegisterPass()
	pass.AddGVNPass()
	pass.AddCFGSimplificationPass()
	pass.Run(mod)
}

func OpenAndCompileFile(name string) error {
	var fset token.FileSet
	ast, err := parser.ParseFile(&fset, name, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	err = CompileFile(&fset, ast)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	files := flag.Args()

	for _, name := range files {
		if err := OpenAndCompileFile(name); err != nil {
			log.Fatal("Error compiling", err)
		}
	}
}
