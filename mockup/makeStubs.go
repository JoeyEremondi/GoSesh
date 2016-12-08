package mockup

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"

	"github.com/JoeyEremondi/GoSesh/multiparty"
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

//Used for both selection and branching
func defaultLabelAndCases(branches map[string]multiparty.LocalType) (string, string) {
	//Get a default label
	//And make a case for each possible branch
	ourLabel := ""
	caseStrings := ""
	for label, branchType := range branches {
		if ourLabel == "" {
			ourLabel = label
		}
		caseStrings += fmt.Sprintf(`
	case "%s":
		%s

			`, label, stub(branchType))
	}
	return ourLabel, caseStrings
}

func addQuotes(s multiparty.Participant) string {
	return fmt.Sprintf(`"%s"`, s)
}

//Stubs for each syntactic variant
func stub(tGeneric multiparty.LocalType) string {
	switch t := tGeneric.(type) {

	//////////////////////////////
	case multiparty.LocalSendType:
		//Generate a variable for each argument, assigning it the default value
		//Along with an array that contains them all serialized as strings
		//assignmentStrings += fmt.Sprintf("sendArgs[%d] = serialize_%s(sendArg_%d)\n", i, sort, i)

		//Serialize each argument, then do the send, and whatever comes after
		return fmt.Sprintf(`
	var sendArg %s //TODO put a value here
	sendBuf := checker.PrepareSend("TODO govec send message", sendArg)
	checker.WriteToUDP(%s, writeFun, sendBuf, addrMaker)
	%s
		`, t.Value, addQuotes(t.Participant), stub(t.Next))

	//////////////////////////////
	case multiparty.LocalReceiveType:
		//Generate a variable for each argument, assigning it the default value
		//Along with an array that contains them all serialized as strings
		assignmentString := ""
		assignmentString += fmt.Sprintf("var receivedValue %s\n", t.Value)
		assignmentString += "checker.UnpackReceive(\"TODO unpack message\", recvBuf, &receivedValue)"
		//Serialize each argument, then do the send, and whatever comes after
		return fmt.Sprintf(`
	recvBuf := make([]byte, 1024)
	checker.ReadFromUDP(%s, readFun, recvBuf)
	%s
	%s
		`, addQuotes(t.Participant), assignmentString, stub(t.Next))

	//////////////////////////////
	case multiparty.LocalBranchingType:
		if len(t.Branches) == 0 {
			panic("Cannot have a Branching with 0 branches")
		}

		_, caseStrings := defaultLabelAndCases(t.Branches)

		//In our code, set the label value to default, then branch based on the label value
		return fmt.Sprintf(`
	ourBuf := make([]byte, 1024)
	checker.ReadFromUDP(%s, readFun, ourBuf)
	var receivedLabel string
	checker.UnpackReceive("TODO Unpack Message", ourBuf, &receivedLabel)
	switch receivedLabel{
		%s
	default:
		panic("Invalid label sent at selection choice")
	}
			`, addQuotes(t.Participant), caseStrings)

	//////////////////////////////
	case multiparty.LocalSelectionType:
		if len(t.Branches) == 0 {
			panic("Cannot have a Selection with 0 branches")
		}

		ourLabel, caseStrings := defaultLabelAndCases(t.Branches)

		//In our code, set the label value to default, then branch based on the label value
		return fmt.Sprintf(`
	var labelToSend = "%s" //TODO which label to send
	buf := checker.PrepareSend("TODO Select message", labelToSend)
	checker.WriteToUDP(%s, writeFun, buf, addrMaker)
	switch labelToSend{
		%s
	default:
		panic("Invalid label sent at selection choice")
	}
			`, ourLabel, addQuotes(t.Participant), caseStrings)

	//////////////////////////////
	case multiparty.LocalNameType:
		//When we see a reference to a type, it was bound by a recursive definition
		//So we jump back to whatever code does the thing recursively
		return fmt.Sprintf("continue %s", t)

	//////////////////////////////
	case multiparty.LocalRecursiveType:
		//Create a labeled infinite loop
		//Any type we refer to ourselves in this type,
		//we jump back to the top of the loop
		return fmt.Sprintf(`
	%s:
	for {
		%s
	}
			`, t.Bind, stub(t.Body))

	case multiparty.LocalEndType:
		return "return"
	case multiparty.ProjectionType:
		return stub(t.T)
	}
	panic(fmt.Sprintf("Invalid local type! %T\n", tGeneric))
}

//Generate the program with all the stubs for a global type
func GenerateProgram(t multiparty.GlobalType) string {

	participantCases := ""
	participantFunctions := ""

	//We need this to remove duplicate participants, bug in Felipe's code?
	seenParticipants := make(map[multiparty.Participant]bool)

	for _, part := range t.Participants() {
		seenParticipants[part] = true
	}

	for part, _ := range seenParticipants {
		nodeName := "node_" + strings.Replace(strings.Replace(string(part), ":", "__", 1), ".", "_", 3)
		fmt.Printf("Participant %s\n", part)
		fmt.Printf("Adding Participant %s\n", part)
		seenParticipants[part] = true
		participantCases += fmt.Sprintf(`
if argsWithoutProg[0] == "%s"{
	%s_main(argsWithoutProg[1:])
}
			`, part, nodeName)

		ourProjection, err := t.Project(part)
		if err != nil {
			panic(err)
		}
		participantFunctions += fmt.Sprintf(`
func %s_main(args []string){
	conn := ConnectNode(%s)
	checker := dynamic.CreateChecker(%s, %#v)
	addrMap := make(map[dynamic.Participant]*net.UDPAddr)
	addrMaker := func(p dynamic.Participant)*net.UDPAddr{
		addr, ok := addrMap[p]
		if ok && addr != nil {
			return addr
		} else {
			addr, _ := net.ResolveUDPAddr("udp", p.String())
			//TODO check err
			addrMap[p] = addr
			return addr
		}
	}
	readFun := makeChannelReader(conn, &addrMap)
	writeFun := makeChannelWriter(conn, &addrMap)
	%s
}
			`, nodeName, addQuotes(part), addQuotes(part), ourProjection, stub(ourProjection))
	}
	return fmt.Sprintf(`
package main

import (
	"net"
	"os"

	"github.com/JoeyEremondi/GoSesh/dynamic"
	"github.com/JoeyEremondi/GoSesh/multiparty"
)

func handleError(e error){
	if e != nil{
			panic(e)
	}
}

var PROTOCOL string =  "udp"
var BUFFERSIZE int = 1000000

// calls this function to set it up
// ConnectNode : Set up a connection for a node
func ConnectNode(laddress string) *net.UDPConn {
	laddressUDP, addrError := net.ResolveUDPAddr(PROTOCOL, laddress)
	handleError(addrError)

	conn, connError := net.ListenUDP(PROTOCOL, laddressUDP)
	handleError(connError)
	conn.SetReadBuffer(BUFFERSIZE)

	return conn
}



//Higher order function: takes in a (possibly empty) map of addresses for channels
//Then returns the function which looks up the address for that channel (if it exists)
//And does the write
func makeChannelWriter(conn *net.UDPConn, addrMap *map[dynamic.Participant]*net.UDPAddr)(func(dynamic.Participant, []byte, *net.UDPAddr) (int, error)){
	return func(p dynamic.Participant, b []byte, addr *net.UDPAddr) (int, error){
		//TODO get addr from map!
		return conn.WriteToUDP(b, addr)
	}
}

func makeChannelReader(conn *net.UDPConn, addrMap *map[dynamic.Participant]*net.UDPAddr)(func(dynamic.Participant, []byte) (int, *net.UDPAddr, error)){
	return func(p dynamic.Participant, b []byte) (int, *net.UDPAddr, error){
		return conn.ReadFromUDP(b)
	}
}

func main(){
	argsWithoutProg := os.Args[1:]
	%s
}
%s
	`, participantCases, participantFunctions)
}
