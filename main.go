package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
)

func main() {
	flag.Parse()
	args := flag.Args()

	convert(args[0])
}

func convert(path string) {
	data, err := os.ReadFile(path)
	mustNil(err)
	in := bytes.NewReader(data)

	out, err := os.OpenFile(getOutFilePath(path), os.O_CREATE|os.O_RDWR, 0666)
	mustNil(err)
	defer out.Close()

	mustNil(convertMapKeyToString(path, in, out))
}

func getOutFilePath(filename string) string {
	return filename[:len(filename)-len(".go")] + "_decls.go"
}

type VisitFunc func(n ast.Node) ast.Visitor

func (v VisitFunc) Visit(n ast.Node) ast.Visitor {
	return v(n)
}

func convertMapKeyToString(path string, in io.Reader, out io.Writer) error {
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

	_, filename := filepath.Split(path)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, in, 0)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "package %s_decls\n\n", f.Name)
	for _, d := range f.Decls {
		if _, ok := d.(*ast.GenDecl); ok {
			ast.Walk(v, d)
			if err := dump(d, out); err != nil {
				return err
			}
		}
	}
	return nil
}

func dump(f any, buf io.Writer) error {
	fset := token.NewFileSet()
	if err := format.Node(buf, fset, f); err != nil {
		return err
	}
	if _, err := buf.Write([]byte("\n\n")); err != nil {
		return err
	}
	return nil
}

func mustNil(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
