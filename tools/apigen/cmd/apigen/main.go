package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"apigen/internal/apigen"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("apigen: ")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [opts] {src_dir | <src_file>...}\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	var (
		pkgName string
		outFile string
	)
	flag.StringVar(&pkgName, "p", "", "package name")
	flag.StringVar(&outFile, "o", "", "output file name, by default output to <pkg_name>_apigen.go, if '-' output to stdout")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	dir, files, err := parseArgs(args)
	if err != nil {
		log.Fatal(err)
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, newFileFilter(files), parser.ParseComments)
	if err != nil {
		log.Fatalf("can't parser.ParseDir: %v", err)
	}

	if len(pkgs) == 0 {
		log.Fatalf("not any package was detected")
	}

	pkgNames := make([]string, 0, len(pkgs))
	for k := range pkgs {
		pkgNames = append(pkgNames, k)
	}

	var pkg *ast.Package
	if pkgName != "" {
		v, ok := pkgs[pkgName]
		if !ok {
			log.Fatalf("%v package not found. available: %v", pkgName, strings.Join(pkgNames, ", "))
		}
		pkg = v
	} else {
		if len(pkgNames) > 1 {
			log.Fatalf("detected more one packages: %v. please select one of them by -p flag", strings.Join(pkgNames, ", "))
		}
		pkg = pkgs[pkgNames[0]]
	}

	genCfg, err := apigen.ParseFiles(pkg.Files)
	if err != nil {
		if e, ok := err.(*apigen.ParseError); ok {
			log.Fatalf("%v: %v", fset.Position(e.Pos), e.Error())
		}
		log.Fatal(err)
	}

	var buf bytes.Buffer
	if err := apigen.GenCode(&buf, genCfg); err != nil {
		if e, ok := err.(*apigen.ParseError); ok {
			log.Fatalf("%v: %v", fset.Position(e.Pos), e.Error())
		}
		log.Fatal(err)
	}

	if outFile == "" {
		outFile = dir + "/" + pkg.Name + "_apigen.go"
	}

	if outFile == "-" {
		if _, err := buf.WriteTo(os.Stdout); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := os.WriteFile(outFile, buf.Bytes(), 0666); err != nil {
			log.Fatal(err)
		}
	}
}

func parseArgs(args []string) (dir string, files []string, _ error) {
	const op = "parseArgs"

	for _, fp := range args {
		fileInfo, err := os.Stat(fp)
		if err != nil {
			if os.IsNotExist(err) {
				return "", nil, fmt.Errorf("%s: %s file not found", op, fp)
			}
			return "", nil, fmt.Errorf("%s: %v", op, err)
		}
		if fileInfo.IsDir() {
			if dir != "" {
				return "", nil, fmt.Errorf("%s: directory must be one", op)
			}
			dir = fp
			continue
		}
		fileDir := filepath.Dir(fp)
		if dir == "" {
			dir = fileDir
		} else if fileDir != dir {
			return "", nil, fmt.Errorf("%s: all files must be in the same directory", op)
		}
		files = append(files, filepath.Base(fp))
	}
	return dir, files, nil
}

func newFileFilter(files []string) func(fs.FileInfo) bool {
	return func(fileInfo fs.FileInfo) bool {
		if strings.HasSuffix(fileInfo.Name(), "_apigen.go") {
			return false
		}
		if len(files) == 0 {
			return true
		}
		for _, fn := range files {
			if fn == fileInfo.Name() {
				return true
			}
		}
		return false
	}
}
