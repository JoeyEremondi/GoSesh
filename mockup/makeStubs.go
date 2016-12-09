package mockup

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"

	"github.com/JoeyEremondi/GoSesh/multiparty"
	"golang.org/x/tools/go/ast/astutil"
)

func modifiedFileString(infile string, imports ...string) string {
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, infile, nil, 0)
	if err != nil {
		panic(err)
	}
	modifyMain(f)

	for _, i := range imports {
		astutil.AddImport(fset, f, i)
	}

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, f)
	return buf.String()
}

func modifyMain(n ast.Node) {
	ast.Inspect(n, func(argNode ast.Node) bool {
		switch innerN := argNode.(type) {
		//Rename our main, so we can add a different main
		case *ast.FuncDecl:
			if innerN.Name.String() == "main" {
				innerN.Name.Name = "makeGlobal"
			}

		case *ast.CallExpr:
			switch selExpr := innerN.Fun.(type) {
			case *ast.SelectorExpr:
				switch pkgIdent := selExpr.X.(type) {
				case *ast.Ident:
					if pkgIdent.Name == "mockup" && selExpr.Sel.Name == "CreateStubProgram" {
						innerN.Fun = &ast.Ident{NamePos: pkgIdent.Pos(), Name: "setGlobalType", Obj: nil}
						innerN.Args = innerN.Args[2:]

					}
				default:
					return true
				}
			}

		default:
			return true
		}
		return true
	})
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
	checker.WriteToUDP("%s", writeFun, sendBuf, addrMaker)
	%s
		`, t.Value, t.Channel, stub(t.Next))

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
	checker.ReadFromUDP("%s", readFun, recvBuf)
	%s
	%s
		`, t.Channel, assignmentString, stub(t.Next))

	//////////////////////////////
	case multiparty.LocalBranchingType:
		if len(t.Branches) == 0 {
			panic("Cannot have a Branching with 0 branches")
		}

		_, caseStrings := defaultLabelAndCases(t.Branches)

		//In our code, set the label value to default, then branch based on the label value
		return fmt.Sprintf(`
	ourBuf := make([]byte, 1024)
	checker.ReadFromUDP("%s", readFun, ourBuf)
	var receivedLabel string
	checker.UnpackReceive("TODO Unpack Message", ourBuf, &receivedLabel)
	switch receivedLabel{
		%s
	default:
		panic("Invalid label sent at selection choice")
	}
			`, t.Channel, caseStrings)

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
	checker.WriteToUDP("%s", writeFun, buf, addrMaker)
	switch labelToSend{
		%s
	default:
		panic("Invalid label sent at selection choice")
	}
			`, ourLabel, t.Channel, caseStrings)

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
		fmt.Printf("Participant %s\n", part)
		fmt.Printf("Adding Participant %s\n", part)
		seenParticipants[part] = true
		participantCases += fmt.Sprintf(`
if argsWithoutProg[0] == "%s"{
	%s_main(argsWithoutProg[1:])
}
			`, part, part)

		ourProjection, err := t.Project(part)
		if err != nil {
			panic(err)
		}
		participantFunctions += fmt.Sprintf(`
func %s_main(args []string){
	localType, err := topGlobalType.Project("%s")
	if err != nil {
		panic(err)
	}
	allRecvChannels := mockup.FindReceivingChannels(localType)
	if len(allRecvChannels) == 0{
		//TODO is this bad?
		panic("This party never does a receive! We have no IP address.")
	}
	conn := ConnectNode(string(allRecvChannels[0]))

	connMap := make(map[multiparty.Channel]*net.UDPConn)
	for _,ch := range allRecvChannels{
		connMap[ch] = ConnectNode(string(ch))
	}

	checker := dynamic.CreateChecker("%s", localType)
	addrMap := make(map[multiparty.Channel]*net.UDPAddr)
	addrMaker := func(p multiparty.Channel)*net.UDPAddr{
		addr, ok := addrMap[p]
		if ok && addr != nil {
			return addr
		} else {
			addr, _ := net.ResolveUDPAddr("udp", string(p))
			//TODO check err
			addrMap[p] = addr
			return addr
		}
	}
	readFun := makeChannelReader(&connMap)
	writeFun := makeChannelWriter(conn, &addrMap)
	%s
}
			`, part, part, part, stub(ourProjection))
	}
	return fmt.Sprintf(`
var topGlobalType multiparty.GlobalType

func setGlobalType(events ...mockup.Event){
	topGlobalType = mockup.Link(events...)
}



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
func makeChannelWriter(conn *net.UDPConn, addrMap *map[multiparty.Channel]*net.UDPAddr)(func(multiparty.Channel, []byte, *net.UDPAddr) (int, error)){
	return func(p multiparty.Channel, b []byte, addr *net.UDPAddr) (int, error){
		//TODO get addr from map!
		return conn.WriteToUDP(b, addr)
	}
}

func makeChannelReader(channelMap *map[multiparty.Channel]*net.UDPConn)(func(multiparty.Channel, []byte) (int, *net.UDPAddr, error)){
	return func(ch multiparty.Channel, b []byte) (int, *net.UDPAddr, error){
		return (*channelMap)[ch].ReadFromUDP(b)
	}
}

func main(){
	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) < 1 {
		panic("Need to give an argument for which node to run!")
	}

	%s
}
%s
	`, participantCases, participantFunctions)
}

func FindReceivingChannels(tGeneric multiparty.LocalType) []multiparty.Channel {
	switch t := tGeneric.(type) {

	case multiparty.LocalSendType:
		return FindReceivingChannels(t.Next)
	case multiparty.LocalReceiveType:
		return append(FindReceivingChannels(t.Next), t.Channel)
	case multiparty.LocalBranchingType:
		return []multiparty.Channel{t.Channel}

	case multiparty.LocalSelectionType:
		ret := []multiparty.Channel{t.Channel}
		for _, next := range t.Branches {
			ret = append(FindReceivingChannels(next), ret...)
		}
		return ret

	case multiparty.LocalNameType:
		return []multiparty.Channel{}

	case multiparty.LocalRecursiveType:
		return FindReceivingChannels(t.Body)

	case multiparty.LocalEndType:
		return []multiparty.Channel{}

	case multiparty.ProjectionType:
		return FindReceivingChannels(t.T)

	}
	panic(fmt.Sprintf("Invalid local type! %T\n", tGeneric))
}
