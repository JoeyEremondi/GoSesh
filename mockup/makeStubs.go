package mockup

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
)

func modifyMain(n ast.Node) {
	ast.Inspect(n, func(argNode ast.Node) bool {
		switch innerN := argNode.(type) {
		//Rename our main, so we can add a different main
		case *ast.FuncDecl:
			if innerN.Name.String() == "main" {
				innerN.Name.Name = "makeGlobal"
			}

		case *ast.SelectorExpr:
			switch pkgIdent := innerN.X.(type) {
			case *ast.Ident:
				if pkgIdent.Name == "mockup" && innerN.Sel.Name == "CreateStubProgram" {
					innerN.Sel.Name = "setGlobalType"
					pkgIdent.Name = "main"
				}
			default:
				return true
			}

		default:
			return true
		}
		return true
	})
}

func main() {
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "../example/loopUntilGood/loopUntilGood.go", nil, 0)
	if err != nil {
		panic(err)
	}
	modifyMain(f)
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, f)
	println(buf.String())
}
