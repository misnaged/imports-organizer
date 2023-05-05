package import_gen

import "go/ast"

func Imports(vals []string) []*ast.ImportSpec {
	var specs []*ast.ImportSpec
	for i := range vals {
		specs = append(specs, NewSpec(vals[i]))
	}
	return specs
}
func NewSpec(val string) *ast.ImportSpec {
	return &ast.ImportSpec{
		Path: NewBasicLit(val),
	}
}
func NewBasicLit(val string) *ast.BasicLit {
	return &ast.BasicLit{
		Value: val,
	}
}
func ReplaceImports(file *ast.File, vals []string) {
	file.Imports = Imports(vals)
}
