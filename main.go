package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
)

func main() {
	flag.Parse()
	args := flag.Args()

	Extract(args[0])
}

type VisitFunc func(n ast.Node) ast.Visitor

func (v VisitFunc) Visit(n ast.Node) ast.Visitor {
	return v(n)
}

func Extract(path string) {
	_, filename := filepath.Split(path)
	buf, err := os.ReadFile(path)
	mustNil(err)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, buf, 0)
	mustNil(err)

	var v ast.Visitor
	v = VisitFunc(func(n ast.Node) ast.Visitor {
		if n == nil {
			return v
		}
		if _, ok := n.(*ast.MapType); !ok {
			// continue to next node using the same visitor
			return v
		}

		called := false
		w := VisitFunc(func(n ast.Node) ast.Visitor {
			if called {
				return nil
			}
			called = true
			if ident, ok := n.(*ast.Ident); ok {
				if ident.Name != "string" {
					ident.Name = "string"
				}
			}
			return nil
		})
		return w
	})

	fmt.Printf("package %s_decls\n\n", f.Name)
	for _, d := range f.Decls {
		if _, ok := d.(*ast.GenDecl); ok {
			ast.Walk(v, d)
			mustNil(dump(d))
		}
	}
}

func dump(f any) error {
	buf := &bytes.Buffer{}
	fset := token.NewFileSet()
	err := format.Node(buf, fset, f)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n\n", buf.String())
	return nil
}

func mustNil(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
