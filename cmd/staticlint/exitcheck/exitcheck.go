package exitcheck

// Package exitcheck содержит анализатор, запрещающий прямой вызов os.Exit в функции main пакета main.
//
// Назначение:
//   - Предотвращает преждевременное завершение программы через os.Exit в функции main.
//   - Рекомендует использовать возврат из main или обработку ошибок для корректного завершения.
//
// Пример нарушения:
//   func main() {
//       os.Exit(1) // будет ошибка
//   }
//
// Пример корректного кода:
//   func main() {
//       if err := run(); err != nil {
//           log.Fatal(err)
//       }
//   }
//
// Анализатор срабатывает только для функции main в пакете main.

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer - анализатор, который запрещает использование os.Exit в функции main пакета main
var Analyzer = &analysis.Analyzer{
	Name:     "exitcheck",
	Doc:      "Запрещает использование os.Exit в функции main пакета main",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)

		if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := fun.X.(*ast.Ident); ok && ident.Name == "os" {
				if fun.Sel.Name == "Exit" {
					if isInMainFunction(pass, call) && isInMainPackage(pass) {
						pass.Reportf(call.Pos(), "прямой вызов os.Exit в функции main пакета main запрещен")
					}
				}
			}
		}
	})

	return nil, nil
}

func isInMainFunction(pass *analysis.Pass, node ast.Node) bool {
	var inMain bool

	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if n == nil {
			return false
		}

		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == "main" {
				if funcDecl.Pos() <= node.Pos() && node.End() <= funcDecl.End() {
					inMain = true
					return false
				}
			}
		}

		return true
	})

	return inMain
}

func isInMainPackage(pass *analysis.Pass) bool {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			return true
		}
	}
	return false
}
