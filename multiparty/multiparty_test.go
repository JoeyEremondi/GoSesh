package multiparty

import (
	"go/format"
	"os"
	"testing"

	"github.com/JoeyEremondi/GoSesh/multiparty"
	//"github.com/JoeyEremondi/GoSesh/multiparty"
)

//
// var BoolValue = func() []Sort {
// 	ans := make([]Sort, 1, 1)
// 	ans[0] = Sort("bool")
// 	return ans
// }()
// var NatValue = func() []Sort {
// 	ans := make([]Sort, 1, 1)
// 	ans[0] = Sort("nat")
// 	return ans
// }()
//
// func TestUnfolding(test *testing.T) {
// 	//example from page 5
// 	example1 := RecursiveType{bind: NameType("X"),
// 		body: ParallelType{a: ValueType{value: nil, prefix: Prefix{P1: "A", P2: "B", channel: "s"}, next: EndType{}}, b: ValueType{value: nil, prefix: Prefix{P1: "B", P2: "A", channel: "t"}, next: NameType("X")}}}
//
// 	unfold1 := RecursiveType{bind: NameType("X"),
// 		body: ParallelType{a: ValueType{value: nil, prefix: Prefix{P1: "A", P2: "B", channel: "s"}, next: EndType{}}, b: ValueType{value: nil, prefix: Prefix{P1: "B", P2: "A", channel: "t"}, next: RecursiveType{bind: NameType("X"), body: ParallelType{a: ValueType{value: nil, prefix: Prefix{P1: "A", P2: "B", channel: "s"}, next: EndType{}}, b: ValueType{value: nil, prefix: Prefix{P1: "B", P2: "A", channel: "t"}, next: NameType("X")}}}}}}
//
// 	if !unfold(example1, make(map[NameType]GlobalType)).equals(unfold1) {
// 		test.FailNow()
// 	}
// }
//
// func TestLinearityAndCoherence(test *testing.T) {
// 	//test linearity for same example of unfolding
// 	example1 := RecursiveType{bind: NameType("X"),
// 		body: ParallelType{a: ValueType{value: BoolValue, prefix: Prefix{P1: "A", P2: "B", channel: "s"}, next: EndType{}}, b: ValueType{value: BoolValue, prefix: Prefix{P1: "B", P2: "A", channel: "t"}, next: NameType("X")}}}
//
// 	if linear(example1) {
// 		test.Errorf("ERROR: Example1 should not be linear\n")
// 	}
// 	//Examples in section 3.2 of Honda et al. (2008)
// 	simplestreaming := RecursiveType{bind: "t", body: ValueType{value: BoolValue, prefix: Prefix{P1: "DP", P2: "K", channel: "d"}, next: ValueType{
// 		value:  BoolValue,
// 		prefix: Prefix{P1: "KP", P2: "K", channel: "k"},
// 		next: ValueType{
// 			value:  BoolValue,
// 			prefix: Prefix{P1: "K", P2: "C", channel: "c"},
// 			next: ValueType{
// 				value:  BoolValue,
// 				prefix: Prefix{P1: "DP", P2: "K", channel: "d"},
// 				next: ValueType{
// 					value:  BoolValue,
// 					prefix: Prefix{P1: "KP", P2: "K", channel: "k"},
// 					next: ValueType{
// 						value:  BoolValue,
// 						prefix: Prefix{P1: "K", P2: "C", channel: "c"},
// 						next:   NameType("t")}}}}}}}
//
// 	if !linear(simplestreaming) {
// 		test.Errorf("ERROR: Simple_Streaming should be linear\n")
// 	}
//
// 	if !coherent(simplestreaming) {
// 		test.Errorf("ERROR: Simple_Streaming should be coherent\n")
// 	}
//
// 	twobuyerprotocol := ValueType{value: BoolValue, prefix: Prefix{P1: "B1", P2: "S", channel: "s"},
// 		next: ValueType{value: NatValue, prefix: Prefix{P1: "S", P2: "B1", channel: "b1"},
// 			next: ValueType{value: NatValue, prefix: Prefix{P1: "S", P2: "B2", channel: "b2"},
// 				next: ValueType{value: NatValue, prefix: Prefix{P1: "B1", P2: "B2", channel: "b'2"},
// 					next: BranchingType{prefix: Prefix{P1: "B2", P2: "S", channel: "s"},
// 						branches: map[string]GlobalType{
// 							"ok": ValueType{value: BoolValue, prefix: Prefix{P1: "B2", P2: "S", channel: "s"},
// 								next: ValueType{value: BoolValue, prefix: Prefix{P1: "S", P2: "B2", channel: "b2"},
// 									next: EndType{}}},
// 							"quit": EndType{}}}}}}}
//
// 	if !linear(twobuyerprotocol) {
// 		test.Errorf("ERROR: Two Buyer Protocol should be linear.\n")
// 	}
//
// 	if !coherent(twobuyerprotocol) {
// 		test.Errorf("ERROR: Two Buyer Protocol should be coherent.\n")
// 	}
//
// 	//Example of section 4.2, Honda et al. (2008)
// 	linearincoherent := BranchingType{
// 		prefix: Prefix{P1: "A", P2: "B", channel: "k"},
// 		branches: map[string]GlobalType{
// 			"ok": ValueType{value: BoolValue, prefix: Prefix{P1: "C", P2: "D", channel: "k'"},
// 				next: EndType{}},
// 			"quit": ValueType{value: NatValue, prefix: Prefix{P1: "C", P2: "D", channel: "k'"},
// 				next: EndType{}}}}
//
// 	if !linear(linearincoherent) {
// 		test.Errorf("ERROR: Incoherent but linear example is not linear.\n")
// 	}
//
// 	if coherent(linearincoherent) {
// 		test.Errorf("ERROR: Incoherent but linear example is coherent! should not be.")
// 	}
// }
//
// func TestASTParse(test *testing.T) {
// 	t := `
// 	package p
// 	const c = 1.0
// 	var X = f(3.14)*2 + c
// 	`
// 	println("*******************\n\n\n ")
// 	println(t)
//
// 	// Create the AST by parsing src.
// 	fset := token.NewFileSet() // positions are relative to fset
// 	f, err := parser.ParseFile(fset, "src.go", t, 0)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// Inspect the AST and print all identifiers and literals.
// 	ast.Inspect(f, func(n ast.Node) bool {
// 		var s string
// 		switch x := n.(type) {
// 		case *ast.BasicLit:
// 			s = x.Value
// 		case *ast.Ident:
// 			s = x.Name
// 		}
// 		if s != "" {
// 			fmt.Printf("%s:\t%s\n", fset.Position(n.Pos()), s)
// 		}
// 		return true
// 	})
// }

func TestBasicStub(test *testing.T) {

	ourMap := map[string]multiparty.GlobalType{
		"isGood": multiparty.EndType{},
		"isBad":  multiparty.NameType("T"),
	}

	t := multiparty.RecursiveType{
		Bind: "T",
		Body: multiparty.ValueType{
			ValuePrefix: multiparty.Prefix{P1: "foo", P2: "bar", PChannel: "channel"},
			Value:       "int",
			ValueNext:   multiparty.BranchingType{BranchPrefix: multiparty.Prefix{P1: "bar", P2: "foo", PChannel: "channel2"}, Branches: ourMap},
		}}
	println("*******************\n\n\n ")
	outFile, err := os.Create("test.go.out")
	defer outFile.Close()

	goodFile, err := os.Create("../testFile/main.go")
	defer goodFile.Close()

	outFile.WriteString(GenerateProgram(t))

	stub := []byte(program(t))
	formatted, err := format.Source(stub)
	if err != nil {
		panic(err)
	}
	//println(string(formatted))
	goodFile.WriteString(string(formatted))
}
