package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"goal/lovm"
	"log"
	"os"
	"runtime/debug"
	"strings"
)

var (
	optimize = flag.Bool("optimize", false, "true to enable llvm optimization passes")
	cfg      = flag.Bool("cfg", false, "view cfg")
)

type Symbol struct {
	Name  string
	Type  Type
	Value lovm.Value
}

func (s Symbol) LlvmType() lovm.Type {
	return s.Type.LlvmType()
}

//
type SymbolMap map[string]Symbol

// visitors
type Scope struct {
	*token.FileSet
	Symbols SymbolMap
}

func (s Scope) GetScope() Scope {
	return s
}

func NewScope(parent *Scope) Scope {
	return NewFileSetScope(parent.FileSet, parent)
}

func NewFileSetScope(fset *token.FileSet, parent *Scope) Scope {
	mergedSymbols := SymbolMap{}
	if parent != nil {
		mergedSymbols = MergeSymbolMaps(parent.Symbols)
	}
	return Scope{fset, mergedSymbols}
}

func MergeSymbolMaps(maps ...SymbolMap) SymbolMap {
	//fmt.Printf("MERGING SYMBOL MAPS %#v\n", maps)
	res := SymbolMap{}
	for _, m := range maps {
		for n, s := range m {
			v := s.Value
			s.Value = v
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
	Module      *lovm.Module
	PackageName string
}

// contains common state shared accross the function
type FunctionVisitor struct {
	*ModuleVisitor
	*FunctionVisitor
	FunctionType FunctionType
	Function     *lovm.Function
	Builder      *lovm.Builder
}

// contains scope local to a block
type BlockVisitor struct {
	Scope
	*FunctionVisitor
	Block *lovm.Block
}

type ExpressionVisitor struct {
	*BlockVisitor
	// result of expression
	Value lovm.Value
	Type  Type
}

func (v *ModuleVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		switch n := node.(type) {
		case *ast.FuncDecl:
			functionType := v.ParseFuncType(n.Type)
			llvmFunction := v.Module.NewFunction(n.Name.Name, functionType.LlvmType())
			// TODO(mkm) put it back
			//if err := v.AddVar(Symbol{Name: n.Name.Name, Type: functionType, Value: llvmFunction}); err != nil {
			//	Perrorf("cannot add symbol %#v: %s", n.Name.Name, err)
			//}

			newScope := NewScope(&v.Scope)
			for i, p := range functionType.Params {
				if p.Name != "" {
					value := llvmFunction.Param(i)
					p.Value = value
					if err := newScope.AddVar(p); err != nil {
						Perrorf("cannot add symbol %#v: %s", p, err)
					}
				}
			}

			if n.Body != nil {
				builder := llvmFunction.NewBuilder()

				entry := llvmFunction.NewBlock()
				builder.SetInsertionPoint(entry)

				fv := &FunctionVisitor{v, nil, functionType, llvmFunction, builder}
				bv := &BlockVisitor{newScope, fv, entry}
				Walk(SkipRoot{bv}, n.Body)

				// debug
				// TODO(mkm): put it back somehow
				//if *cfg {
				//lovm.ViewFunctionCFG(llvmFunction)
				//}
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
			var value lovm.Value
			if vs.Values != nil {
				ev := &ExpressionVisitor{s, nil, typ}
				Walk(ev, vs.Values[idx])
				value = ev.Value
			} else {
				value = lovm.ConstInt(typ.LlvmType(), 0)
			}
			if err := s.AddVar(Symbol{Name: n.Name, Type: typ, Value: value}); err != nil {
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

	Perrorf("cannot resolve symbol: %s", name)
	return Symbol{}
}

func (s *Scope) AddVar(variable Symbol) error {
	name := variable.Name
	if _, ok := s.Symbols[name]; ok {
		return fmt.Errorf("Multiple declarations of %s", name)
	}
	if variable.Value == nil {
		value := lovm.ConstInt(variable.Type.LlvmType(), 0)
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
				v.Value = v.Builder.IAdd(xev.Value, yev.Value)
			case token.SUB:
				v.Value = v.Builder.ISub(xev.Value, yev.Value)
			case token.MUL:
				v.Value = v.Builder.IMul(xev.Value, yev.Value)
			case token.QUO:
				v.Value = v.Builder.ISDiv(xev.Value, yev.Value)
			case token.REM:
				v.Value = v.Builder.ISRem(xev.Value, yev.Value)
			case token.LSS:
				v.Value = v.Builder.IICmp(lovm.IntSLT, xev.Value, yev.Value)
				v.Type = Bool
			case token.GTR:
				v.Value = v.Builder.IICmp(lovm.IntSGT, xev.Value, yev.Value)
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
			switch n.Kind {
			case token.INT:
				v.Value = lovm.ConstIntFromString(v.Type.LlvmType(), n.Value, 10)
			case token.STRING:
				v.Value = v.Module.ConstString(n.Value)
			default:
				Perrorf("Unimplemented literal: %#v", n)
			}
		case *ast.Ident:
			if n.Name == "true" {
				v.Type = Bool
				v.Value = lovm.ConstInt(v.Type.LlvmType(), 1)
			} else if n.Name == "false" {
				v.Type = Bool
				v.Value = lovm.ConstInt(v.Type.LlvmType(), 0)
			} else {
				symbol := v.ResolveSymbol(n.Name)
				v.Type = symbol.Type
				v.Value = symbol.Value
			}
			return nil
		case *ast.CallExpr:
			if id, ok := n.Fun.(*ast.Ident); ok {
				if typ, err := v.ResolveType(id); err == nil {
					if len(n.Args) != 1 {
						Perrorf("type conversion can have only one argument")
					}
					ev := v.Evaluate(n.Args[0])
					Perrorf("not migrated to new api: %v, %v", typ, ev)
					//v.Value = v.Builder.CreateIntCast(ev.Value, typ.LlvmType(), "")
				} else {
					fs := v.ResolveSymbol(id.Name)
					if _, ok := fs.Type.(FunctionType); !ok {
						Perrorf("Calling a non function")
					}
					args := []lovm.Value{}
					for _, a := range n.Args {
						ex := v.Evaluate(a)
						// TODO(mkm) check types
						args = append(args, ex.Value)
					}
					Perrorf("not migrated to new api")
					//v.Value = v.Builder.Call(*fs.Value, args)
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
	ev := &ExpressionVisitor{v, nil, typ}
	Walk(ev, exp)
	return ev
}

func (v *BlockVisitor) EvaluateBlock(exp ast.Stmt) *BlockVisitor {
	newScope := NewScope(&v.Scope)
	bv := &BlockVisitor{newScope, v.FunctionVisitor, &lovm.Block{}}
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

			values := make([]lovm.Value, len(n.Results))
			types := make([]lovm.Type, len(n.Results))

			for i, e := range n.Results {
				ev := &ExpressionVisitor{v, nil, functionReturnSymbols[i].Type}
				Walk(ev, e)
				values[i] = ev.Value
				types[i] = ev.Type.LlvmType()
			}

			var res lovm.Value
			switch len(values) {
			case 1:
				res = values[0]
			default:
				Perrorf("unimplemented multiple return values")
			}
			v.Builder.Return(res)
		case *ast.ExprStmt:
			ev := &ExpressionVisitor{v, nil, Any}
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

				values := make([]lovm.Value, len(n.Lhs))
				for i, e := range n.Rhs {
					ev := &ExpressionVisitor{v, nil, symbols[i].Type}
					Walk(ev, e)
					values[i] = ev.Value
				}
				for i, sym := range symbols {
					sym.Value = values[i]
				}
			}
		case *ast.IfStmt:
			//var scope Scope = v.Scope
			if n.Init != nil {
				// TODO(mkm) fix scoping issues when detecting changed values
				//scope = NewScope(&v.Scope)
				Walk(v, n.Init)
			}
			cond := v.Evaluate(Bool, n.Cond)
			iftrue := v.Function.NewBlock()
			iffalse := v.Function.NewBlock()
			endif := v.Function.NewBlock()

			v.Builder.BranchIf(cond.Value, iftrue, iffalse)

			v.Builder.SetInsertionPoint(iftrue)
			//ifTrueVisitor := v.EvaluateBlock(n.Body)
			v.EvaluateBlock(n.Body)
			//ifTrueSource := v.Builder.GetInsertBlock()
			v.Builder.Branch(endif)

			v.Builder.SetInsertionPoint(iffalse)
			//ifFalseSource := iffalse
			//var ifFalseVisitor *BlockVisitor
			if n.Else != nil {
				//ifFalseVisitor = v.EvaluateBlock(n.Else)
				v.EvaluateBlock(n.Else)
				//ifFalseSource = v.Builder.GetInsertBlock()
			}
			v.Builder.Branch(endif)
			v.Builder.SetInsertionPoint(endif)

			// TODO(mkm) use phis created by lovm
			/*
				type Phi struct {
					Parent Symbol
					Left   Symbol
					Right  Symbol
				}
				phis := map[string]Phi{}

				ForUpdatedVars(scope, ifTrueVisitor.Scope, func(a, b Symbol) {
					phi := phis[a.Name]
					phi.Parent = a
					phi.Left = b
					phis[a.Name] = phi
				})
				if ifFalseVisitor != nil {
					ForUpdatedVars(scope, ifFalseVisitor.Scope, func(a, b Symbol) {
						phi := phis[a.Name]
						phi.Parent = a
						phi.Right = b
						phis[a.Name] = phi
					})
				}

				for _, p := range phis {
					if p.Left.Value == nil {
						p.Left = p.Parent
					}
					if p.Right.Value == nil {
						p.Right = p.Parent
					}

					phi := v.Builder.CreatePHI(p.Parent.Type.LlvmType(), "")
					phiVals := []lovm.Value{*p.Left.Value, *p.Right.Value}
					phiBlocks := []*lovm.Block{ifTrueSource, ifFalseSource}
					phi.AddIncoming(phiVals, phiBlocks)

					*p.Parent.Value = phi
				}
			*/
		default:
			Perrorf("----- Block visitor: UNKNOWN %#v\n", node)
			return v
		}
	} else {
		//		fmt.Printf("popping\n")
		//v.DumpScope()
	}
	return nil
}

/*
func ForUpdatedVars(parent, child Scope, fn func(a, b Symbol)) {
	for n, s := range parent.Symbols {
		if cs, ok := child.Symbols[n]; ok {
			if *s.Value != *cs.Value {
				fn(s, cs)
			}
		}
	}
}
*/

func CompileFile(fset *token.FileSet, tree *ast.File) error {
	DumpToFile(tree, "/tmp/ast")

	ctx := lovm.NewContext(os.Stderr)
	v := &ModuleVisitor{NewFileSetScope(fset, nil), ctx.NewModule(tree.Name.Name), ""}
	Walk(v, tree)

	fmt.Printf("LLVM: -----------\n")

	ctx.Emit()
	return nil
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
