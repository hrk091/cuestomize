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

func Extract(path string) {
	_, filename := filepath.Split(path)
	buf, err := os.ReadFile(path)
	mustNil(err)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, buf, 0)
	mustNil(err)

	mustNil(dump(f))
}

func dump(f *ast.File) error {
	buf := &bytes.Buffer{}
	fset := token.NewFileSet()
	err := format.Node(buf, fset, f)
	if err != nil {
		return err
	}
	fmt.Printf("%+v", buf.String())
	return nil
}

func mustNil(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
