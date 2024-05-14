// Package customlint defines an custom Analyzer that disables for os.Exit calls.
package customlint

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// The primary type in the API is Analyzer.
// An Analyzer statically describes an analysis function:
// its name, documentation, flags, relationship to other analyzers, and of course, its logic.
var Analyzer = &analysis.Analyzer{
	Name: "custom",
	Doc:  "check for os.Exit calls",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			ast.Inspect(file, func(node ast.Node) bool {
				for _, decls := range file.Decls {
					if funcDecl, ok := decls.(*ast.FuncDecl); ok && funcDecl.Name.Name == "main" {
						body := *funcDecl.Body
						for _, stmt := range body.List {
							if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
								if call, ok := exprStmt.X.(*ast.CallExpr); ok {
									if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
										if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == "os" {
											if fun.Sel.Name == "Exit" {
												pass.Report(analysis.Diagnostic{
													Pos:     call.Pos(),
													Message: "direct calls of os.Exit are not allowed",
												})
											}
										}
									}
								}
							}
						}
					}
				}
				return true
			})
		}
	}
	return nil, nil
}
