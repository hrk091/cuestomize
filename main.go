package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"golang.org/x/tools/go/packages"
	"log"
)

func main() {
	flag.Parse()
	args := flag.Args()

	Extract(args)
}

func Extract(args []string) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedTypes | packages.NeedTypesSizes |
			packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedDeps |
			packages.NeedModule,
	}
	pkgs, err := packages.Load(cfg, args...)
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range pkgs {
		fmt.Printf("%s, %s\n", p.Name, p.ID)
		for _, f := range p.Syntax {
			fmt.Printf("%s\n", f.Name)
			must(dump(f))
		}
	}
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

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
