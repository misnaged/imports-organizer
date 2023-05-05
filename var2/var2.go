package main

import (
	"bytes"
	"fmt"
	"github.com/dave/dst/decorator"
	"github.com/misnaged/annales/logger"
	"github.com/misnaged/import_organizer/package_collector"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"os"
	"strconv"
	"strings"
	_ "unsafe"

	_ "go/format"
)

//go:linkname parse go/format.parse
func parse(fset *token.FileSet, filename string, src []byte, fragmentOk bool) (
	file *ast.File,
	sourceAdj func(src []byte, indent int) []byte,
	indentAdj int,
	err error,
)

//go:linkname format go/format.format
func format(
	fset *token.FileSet,
	file *ast.File,
	sourceAdj func(src []byte, indent int) []byte,
	indentAdj int,
	src []byte,
	cfg printer.Config,
) ([]byte, error)

func (abs *AbstractSxTree) GetKeys() ([]string, []string) {
	var keys, values []string
	for k, v := range abs.advanced {
		k, _ = strconv.Unquote(k)
		keys = append(keys, k)
		values = append(values, v)
	}

	return keys, values
}
func (abs *AbstractSxTree) CombineIntoPath(imprs []*ast.ImportSpec) {
	for i := range imprs {
		if imprs[i].Doc != nil && imprs[i].Name == nil {
			commentedPath := fmt.Sprintf("//%s \n %s", imprs[i].Doc.Text(), imprs[i].Path.Value)
			abs.advanced[imprs[i].Path.Value] = commentedPath
		}
		if imprs[i].Doc == nil && imprs[i].Name != nil {
			namedPath := fmt.Sprintf("%s %s", imprs[i].Name.String(), imprs[i].Path.Value)
			abs.advanced[imprs[i].Path.Value] = namedPath
		}
		if imprs[i].Doc != nil && imprs[i].Name != nil {
			commentedNamedPath := fmt.Sprintf("//%s\n %s %s", imprs[i].Doc.Text(), imprs[i].Name.String(), imprs[i].Path.Value)
			abs.advanced[imprs[i].Path.Value] = commentedNamedPath
		}
		if imprs[i].Doc == nil && imprs[i].Name == nil {
			abs.advanced[imprs[i].Path.Value] = imprs[i].Path.Value
		}
	}
}

//func NewAdvancedImports(comment, name, path string) *AdvancedImports {
//	return &AdvancedImports{
//		comment: comment,
//		name:    name,
//		path:    path,
//	}
//}

type AbstractSxTree struct {
	advanced        map[string]string
	fset            *token.FileSet
	astFile         *ast.File
	decl            []ast.Decl
	formattedOutput []byte
	sourceAdj       func(src []byte, indent int) []byte
	indentAdj       int
}

func (abs *AbstractSxTree) Example(imprs []*ast.ImportSpec) {
	for i := range imprs {
		if imprs[i].Doc != nil {
			fmt.Println(imprs[i].Doc.Text(), imprs[i].Name.String(), imprs[i].Path.Value)
		}
	}
}

func (abs *AbstractSxTree) ImportsFromFiles(b []byte, file string) []*ast.ImportSpec {
	err := abs.NewTree(b, file)
	if err != nil {
		logger.Log().Error(err.Error())
		return nil
	}
	return abs.astFile.Imports
}
func (abs *AbstractSxTree) VanillaCleaner(b []byte, file string) error {
	err := abs.NewTree(b, file)
	if err != nil {
		return fmt.Errorf("NewTree in ClearImports %w", err)
	}
	abs.astFile, err = parser.ParseFile(abs.fset, file, b, parser.ParseComments)
	if err != nil {
		log.Fatalln(err)
	}
	astutil.DeleteImport(abs.fset, abs.astFile, file)

	for i := range abs.astFile.Imports {
		abs.astFile.Imports[i].Path.Value = ""
		abs.astFile.Imports[i].Path.ValuePos = 0
		abs.astFile.Imports[i].Comment = nil
		abs.astFile.Imports[i].Doc = nil
		abs.astFile.Imports[i].Name = nil

		//for idx := 0; idx < len(abs.astFile.Decls); idx++ {
		//	d := abs.astFile.Decls[idx]
		//	switch d.(type) {
		//	case *ast.FuncDecl:
		//	case *ast.GenDecl:
		//		dd := d.(*ast.GenDecl)
		//		if dd.Tok == token.IMPORT {
		//			abs.astFile.Imports = append(abs.astFile.Imports, dd.Specs[i].(*ast.ImportSpec))
		//		}
		//	}
		//}
	}

	out, err := format(abs.fset, abs.astFile, abs.sourceAdj, abs.indentAdj, b, cfg)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	err = os.WriteFile(file, out, os.ModeAppend)
	if err != nil {
		return fmt.Errorf(" WriteFile %w", err)
	}
	return nil
}

func (abs *AbstractSxTree) DaveCleaner(b []byte, file string) error {
	err := abs.NewTree(b, file)
	if err != nil {
		return fmt.Errorf("NewTree in ClearImports %w", err)
	}
	f, err := decorator.Parse(b)
	if err != nil {
		return fmt.Errorf("decorator.Parse failed: %w", err)
	}

	//.Body.List[0].(*dst.ExprStmt).X.(*dst.CallExpr)
	for i := range f.Imports {
		f.Imports[i].Decorations().Start.Replace("")
		f.Imports[i].Decorations().End.Replace("")
		f.Imports[i].Decs.Start.Replace("")
		f.Imports[i].Decs.End.Replace("")

		f.Imports[i].Decs.NodeDecs.Start.Replace("")
		f.Imports[i].Decs.NodeDecs.End.Replace("")
		f.Imports[i].Decorations().Start.Replace("")
	}

	err = os.WriteFile(file, []byte(""), 0666)
	if err != nil {
		return fmt.Errorf(" WriteFile %w", err)
	}
	fileToOpen, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return fmt.Errorf(" failed to open file: %w", err)
	}
	defer fileToOpen.Close()
	err = decorator.Fprint(fileToOpen, f)
	if err != nil {
		return fmt.Errorf(" printing decoration to the file failed: %w", err)
	}
	return nil
}

func (abs *AbstractSxTree) TrimSpace(b []byte, file string) error {
	err := abs.NewTree(b, file)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	for i := 0; i < len(abs.astFile.Decls); i++ {
		d := abs.astFile.Decls[i]
		switch d.(type) {
		case *ast.FuncDecl:
		case *ast.GenDecl:
			dd := d.(*ast.GenDecl)
			if dd.Tok == token.IMPORT {
				iSpec := &ast.ImportSpec{Path: &ast.BasicLit{Value: strings.TrimSpace("")}}
				dd.Specs = append(dd.Specs, iSpec)
			}
		}
	}
	out, err := format(abs.fset, abs.astFile, abs.sourceAdj, abs.indentAdj, b, cfg)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	err = os.WriteFile(file, out, 0666)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (abs *AbstractSxTree) ConditionToSort(arr []string) bool {
	for i := range arr {
		if arr[i] != "\n" && arr[i] == abs.advanced[arr[i]] {
			return true
		}
	}
	return false
}
func (abs *AbstractSxTree) Sort(str []string) []string {
	k, v := abs.GetKeys()
	for arrIdx := range str {
		for i := range k {
			if str[arrIdx] == k[i] {
				str[arrIdx] = v[i]
			}
		}
		//kekich := fmt.Sprintf("str:%s", str[arrIdx])
		//lol2.Printf(kekich)

	}
	//for i := range k {
	//keka := fmt.Sprintf("k:%s v:%s", k[i], v[i])
	//lol.Printf(keka)
	//}
	return str
}

func (abs *AbstractSxTree) WriteImports(b []byte, str []string, file string) error {
	err := abs.NewTree(b, file)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	for i := 0; i < len(abs.astFile.Decls); i++ {
		abs.astFile, err = parser.ParseFile(abs.fset, file, b, parser.ImportsOnly)
		if err != nil {
			log.Fatalln(err)
		}
		d := abs.astFile.Decls[i]
		switch d.(type) {
		case *ast.FuncDecl:
		case *ast.GenDecl:

			dd := d.(*ast.GenDecl)
			if dd.Tok == token.IMPORT {
				for ii := range str {
					if str[ii] == "\n" {
						iSpec := &ast.ImportSpec{Path: &ast.BasicLit{Value: str[ii]}}
						dd.Specs = append(dd.Specs, iSpec)
					} else {
						iSpec := &ast.ImportSpec{Path: &ast.BasicLit{Value: str[ii]}}
						dd.Specs = append(dd.Specs, iSpec)
					}
				}
			}
		}

	}

	var buf bytes.Buffer
	out, err := format(abs.fset, abs.astFile, abs.sourceAdj, abs.indentAdj, b, cfg)
	buf.Write(out)
	lol.Println(string(buf.Bytes()))
	f, err := decorator.Parse(b)
	if err != nil {
		return fmt.Errorf("decorator.Parse failed: %w", err)
	}
	for i := range f.Imports {
		f.Imports[i].Path.Value = ""
	}
	var bek bytes.Buffer
	decorator.Fprint(&bek, f)
	lol2.Println(string(bek.Bytes()))
	return nil
}
func (abs *AbstractSxTree) NewTree(b []byte, filename string) error {
	abs.fset = token.NewFileSet()
	file, sourceAdj, ident, err := parse(abs.fset, filename, b, true)
	if err != nil {
		return fmt.Errorf("NewTree %w", err)
	}
	abs.astFile, abs.sourceAdj, abs.indentAdj = file, sourceAdj, ident
	return nil
}

var (
	kek, _  = os.Create("kek.txt")
	lol     = log.New(kek, "", 0)
	kek2, _ = os.Create("kek2.txt")
	lol2    = log.New(kek2, "", 0)
)

func main() {
	package_collector.Populate()
	package_collector.Std()
	package_collector.ExternalPackages()
	//lol3.Println(package_collector.Packages)
	files := package_collector.GoFiles()
	for i := range files {
		if ImportCheck(files[i]) {
			str := MergeImportsSlices(files[i])
			abs := &AbstractSxTree{
				advanced: make(map[string]string),
			}
			r, err := os.ReadFile(files[i])
			if err != nil {
				logger.Log().Errorf("error ReadFile %v", err)
				os.Exit(1)
			}
			imports := abs.ImportsFromFiles(r, files[i])

			abs.CombineIntoPath(imports)
			//k, _ := abs.GetKeys()

			err = abs.VanillaCleaner(r, files[i])
			if err != nil {
				logger.Log().Errorf("vanilla clear import %v", err)
				os.Exit(1)
			}
			//err := abs.DaveCleaner(r, files[i])
			//if err != nil {
			//	logger.Log().Errorf("dave clear import %v", err)
			//	os.Exit(1)
			//}
			re, err := os.ReadFile(files[i])
			if err != nil {
				logger.Log().Errorf("error ReadFile %v", err)
				os.Exit(1)
			}

			if err = abs.TrimSpace(re, files[i]); err != nil {
				logger.Log().Errorf("TrimSpace %v", err)
				os.Exit(1)
			}
			str = abs.Sort(str)

			if err = abs.WriteImports(re, str, files[i]); err != nil {
				logger.Log().Errorf("WriteImports %v", err)
				os.Exit(1)
			}
		}
	}
}

var cfg = printer.Config{Mode: printer.RawFormat, Tabwidth: 0}

func ImportCheck(file string) bool {
	fset := token.NewFileSet() // positions are relative to fset

	src, _ := os.ReadFile(file)

	f, _ := parser.ParseFile(fset, "", src, parser.ImportsOnly)

	if len(f.Imports) > 1 {
		return true
	}

	return false
}

func getImports(file string) ([]string, error) {
	fset := token.NewFileSet() // positions are relative to fset

	src, err := os.ReadFile(file)

	f, err := parser.ParseFile(fset, "", src, parser.ImportsOnly)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	imports := make([]string, len(f.Imports))
	for i, s := range f.Imports {
		imports[i] = s.Path.Value
	}
	for i := range imports {
		imports[i], err = strconv.Unquote(imports[i])
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
	}

	return imports, nil
}

func MergeImportsSlices(file string) []string {

	imports, _ := getImports(file)

	std := SortImportsByRank(imports, 0)
	std = append(std, "\n")
	ext := SortImportsByRank(imports, 1)
	ext = append(ext, "\n")
	std = append(std, ext...)

	out := SortImportsByRank(imports, 2)
	out = append(out, "\n")

	proj := SortImportsByRank(imports, 3)
	proj = append(proj, "\n")
	out = append(out, proj...)
	std = append(std, out...)
	return std
}

func SortImportsByRank(importsFromFile []string, rank int) []string {
	var rankedList []string
	for pkg, ii := range package_collector.Packages {
		for i := range importsFromFile {
			//rank = package_collector.Packages[pkg]
			if pkg == importsFromFile[i] && ii == rank {
				rankedList = append(rankedList, pkg)
			}
		}
	}

	return rankedList
}
