package main

import (
	"fmt"
	"go/ast"
	"io"
	"log"
	"os"
)

type dumper struct {
	out   io.Writer
	depth int
}

func (v *dumper) Indent() {
	for i := 0; i < v.depth; i++ {
		fmt.Fprintf(v.out, " ")
	}
}

func (v *dumper) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		v.Indent()
		fmt.Fprintf(v.out, "%#v\n", node)
		return &dumper{v.out, v.depth + 1}
	}
	return nil
}

func DumpToFile(tree ast.Node, fileName string) {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("cannot open debug file", err)
	}
	defer f.Close()
	ast.Walk(&dumper{out: f}, tree)
}

func Dump(tree ast.Node, fileName string) {
	ast.Walk(&dumper{out: os.Stdout}, tree)
}

func (s *Scope) DumpScope() {
	fmt.Printf("Scope:\n")
	for k, v := range s.Symbols {
		fmt.Printf("%s : %#v\n", k, v)
	}
	fmt.Printf("end scope\n")
}
