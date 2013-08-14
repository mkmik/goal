package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

type visitor struct {
	depth int
}

func (v *visitor) Indent() {
	for i := 0; i < v.depth; i++ {
		fmt.Printf(" ")
	}
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	next := &visitor{v.depth + 1}

	if node != nil {
		v.Indent()

		switch n := node.(type) {
		case *ast.FuncDecl:
			fmt.Printf("FUNC DECL %s: %#v\n", n.Name, n)
			if n.Body != nil {
				ast.Walk(next, n.Body)
			}
			return nil
		case *ast.ExprStmt:
			v.Indent()
			fmt.Printf("EXPR STATEMENT %#v\n", n)
		case *ast.DeclStmt:
			v.Indent()
			fmt.Printf("DECL STMT %#v\n", n)
		case *ast.AssignStmt:
			v.Indent()
			if n.Tok == token.DEFINE {
				fmt.Printf("DEFINE ASSIGN STMT %#v\n", n)
			} else {
				fmt.Printf("PLAIN ASSIGN STMT %#v ... %#v\n", n, n.Lhs[0])
			}
		default:
			fmt.Printf("-------- %#v\n", node)
			return next
		}
	} else {
		//		fmt.Printf("popping\n")
	}
	return nil
}

func CompileFile(tree *ast.File) error {
	fmt.Printf("compiled %#v\n", tree)
	v := &visitor{}
	ast.Walk(v, tree)

	DumpToFile(tree, "/tmp/ast")
	return nil
}

func OpenAndCompileFile(name string) error {
	var fset token.FileSet
	ast, err := parser.ParseFile(&fset, name, nil, 0)
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
