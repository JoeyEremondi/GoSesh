package mockup

import "go/ast"

func removeMain(n ast.Node) {
	ast.Inspect(n, func(argNode ast.Node) bool {
		switch innerN := argNode.(type) {
		case *ast.FuncDecl:
			if innerN.Name.String() == "main" {
				innerN.Name.Name = "oldm"
				return false
			} else {
				return true
			}
		default:
			return true
		}
	})
}
