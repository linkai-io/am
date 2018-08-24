package convert

import (
	"testing"
)

func TestConvertStructs(t *testing.T) {
	/*
		src, err := os.Open("protoc.go")
		if err != nil {
			t.Fatalf("error opening file: %s\n", err)
		}

		fset := token.NewFileSet() // positions are relative to fset
		f, err := parser.ParseFile(fset, "", src, 0)
		if err != nil {
			panic(err)
		}

		// Print the AST.
		// ast.Print(fset, f)
		ast.Fprint(os.Stdout, fset, f, func(name string, value reflect.Value) bool {
			if ast.NotNilFilter(name, value) {
				return value.Type().String() != "*ast.Object"
			}
			return false
		})
	*/
}
