package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"golang.org/x/mod/modfile"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	flag.Parse()
	args := flag.Args()
	path := validateArgs(args)
	validateIsModuleRoot()
	convert(path)
	showCueGetScript(path)
}

func convert(path string) {
	data, err := os.ReadFile(path)
	mustNil(err)
	in := bytes.NewReader(data)

	outDir, outFile := getOutFilePath(path)
	mustNil(os.MkdirAll(outDir, 0750))
	out, err := os.OpenFile(filepath.Join(outDir, outFile), os.O_CREATE|os.O_RDWR, 0666)
	mustNil(err)
	defer out.Close()

	mustNil(convertMapKeyToString(path, in, out))
}

func showCueGetScript(path string) {
	outDir, _ := getOutFilePath(path)
	modPath, err := resolveModuleName()
	mustNil(err)

	fmt.Println(cueGetScript(modPath, outDir))
}

func cueGetScript(modPath, outDir string) string {
	return fmt.Sprintf(`
# Init cue.mod if not setup yet
if [[ ! -d cue.mod ]]; then
    cue mod init %s 
fi
# Generate cue type defs
cue get go %s
`, modPath, filepath.Join(modPath, outDir))
}

func validateArgs(args []string) string {
	if len(args) == 0 {
		log.Fatal("target file does not specified")
	}
	path := args[0]
	if filepath.Ext(path) != ".go" {
		log.Fatal("given file is not CUE file")
	}
	return path
}

func validateIsModuleRoot() {
	if _, err := os.Stat("go.mod"); err != nil {
		log.Fatal("Must run at go module root.")
	}
}

func resolveModuleName() (string, error) {
	buf, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}
	modPath := modfile.ModulePath(buf)
	if modPath == "" {
		return "", fmt.Errorf("resolve module name from go.mod")
	}
	return modPath, nil
}

func getOutFilePath(path string) (string, string) {
	dir, filename := filepath.Split(path)
	trimmed := strings.TrimSuffix(dir, "/")
	dir = filepath.Join("types", trimmed)
	return dir, filename
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
	fmt.Fprintf(out, "package %s", f.Name)
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
	fmt.Fprintf(buf, "\n\n")
	fset := token.NewFileSet()
	if err := format.Node(buf, fset, f); err != nil {
		return err
	}
	return nil
}

func mustNil(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
