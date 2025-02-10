package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// OSExitFromMainAnalyzer - analyzer which check if os.Exit() call ever appeared in func main() in package main
var NoExitInMainAnalyzer = &analysis.Analyzer{
	Name:     "noexitinmain",
	Doc:      "check if os.Exit() call ever appeared in func main() in package main",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		packageIsMain, funcIsMain := false, false
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.File:
				packageIsMain = x.Name.Name == "main"
			case *ast.FuncDecl:
				funcIsMain = x.Name.Name == "main"
			case *ast.SelectorExpr:
				ident, ok := x.X.(*ast.Ident)
				if packageIsMain && funcIsMain && ok && ident.Name == "os" && x.Sel.Name == "Exit" {
					pass.Reportf(ident.NamePos, "os.Exit called in main func in main package")
				}
			}
			return true
		})
	}
	return nil, nil
}
