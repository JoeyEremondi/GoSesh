package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/JoeyEremondi/GoSesh/dynamic"
	"github.com/JoeyEremondi/GoSesh/mockup"
	"github.com/JoeyEremondi/GoSesh/multiparty"
)

func makeGlobalType() {

	channelAToB := mockup.Channel{
		Name:        "127.0.0.1:24602",
		Source:      "A",
		Destination: "B"}

	channelBToA := mockup.Channel{
		Name:        "127.0.0.1:24601",
		Source:      "B",
		Destination: "A"}

	channelAToC := mockup.Channel{
		Name:        "127.0.0.1:24603",
		Source:      "A",
		Destination: "C"}

	channelCToA := mockup.Channel{
		Name:        "127.0.0.1:24601",
		Source:      "C",
		Destination: "A"}

	setGlobalType(
		mockup.Send(channelAToB, mockup.MessageType{Type: "int"}),
		mockup.Switch(channelBToA,
			mockup.Case("B-Fail",
				mockup.Send(channelAToC, mockup.MessageType{Type: "int"}),
				mockup.Switch(channelCToA,
					mockup.Case("C-Fail",
						mockup.Send(channelAToB, mockup.MessageType{Type: "bool"}),
						mockup.Send(channelAToC, mockup.MessageType{Type: "bool"}),
					),
					mockup.Case("C-Commit",
						mockup.Send(channelAToB, mockup.MessageType{Type: "bool"}),
						mockup.Send(channelAToC, mockup.MessageType{Type: "bool"}),
					),
				),
			),
			mockup.Case("B-Commit",
				mockup.Send(channelAToC, mockup.MessageType{Type: "int"}),
				mockup.Switch(channelCToA,
					mockup.Case("C-Fail",
						mockup.Send(channelAToB, mockup.MessageType{Type: "bool"}),
						mockup.Send(channelAToC, mockup.MessageType{Type: "bool"})),
					mockup.Case("C-Commit",
						mockup.Send(channelAToB, mockup.MessageType{Type: "bool"}),
						mockup.Send(channelAToC, mockup.MessageType{Type: "bool"}),
					),
				),
			),
		),
	)
}

var topGlobalType multiparty.GlobalType

func setGlobalType(events ...mockup.Event) {
	topGlobalType = mockup.Link(events...)
}

func handleError(e error) {
	if e != nil {
		panic(e)
	}
}

var PROTOCOL string = "udp"
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

func makeCheckerReaderWriter(part string) (dynamic.Checker,
	func(multiparty.Channel) *net.UDPAddr,
	func(multiparty.Channel, []byte) (int, *net.UDPAddr, error),
	func(multiparty.Channel, []byte, *net.UDPAddr) (int, error)) {

	localType, err := topGlobalType.Project(multiparty.Participant(part))
	if err != nil {
		panic(err)
	}
	allRecvChannels := make(map[multiparty.Channel]bool)
	mockup.FindReceivingChannels(localType, &allRecvChannels)

	connMap := make(map[multiparty.Channel]*net.UDPConn)

	var conn *net.UDPConn
	areFirst := true

	for ch, _ := range allRecvChannels {
		connMap[ch] = ConnectNode(string(ch))
		if areFirst {
			areFirst = false
			conn = connMap[ch]
		}

	}

	checker := dynamic.CreateChecker(part, localType)
	addrMap := make(map[multiparty.Channel]*net.UDPAddr)
	addrMaker := func(p multiparty.Channel) *net.UDPAddr {
		addr, ok := addrMap[p]
		if ok && addr != nil {
			return addr
		} else {
			addr, _ := net.ResolveUDPAddr("udp", string(p))
			addrMap[p] = addr
			return addr
		}
	}
	readFun := makeChannelReader(&connMap)
	writeFun := makeChannelWriter(conn, &addrMap)
	return checker, addrMaker, readFun, writeFun
}

//Higher order function: takes in a (possibly empty) map of addresses for channels
//Then returns the function which looks up the address for that channel (if it exists)
//And does the write
func makeChannelWriter(conn *net.UDPConn, addrMap *map[multiparty.Channel]*net.UDPAddr) func(multiparty.Channel, []byte, *net.UDPAddr) (int, error) {
	return func(p multiparty.Channel, b []byte, addr *net.UDPAddr) (int, error) {
		return conn.WriteToUDP(b, addr)
	}
}

func makeChannelReader(channelMap *map[multiparty.Channel]*net.UDPConn) func(multiparty.Channel, []byte) (int, *net.UDPAddr, error) {
	return func(ch multiparty.Channel, b []byte) (int, *net.UDPAddr, error) {
		return (*channelMap)[ch].ReadFromUDP(b)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	makeGlobalType()

	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) < 1 {
		panic("Need to give an argument for which node to run!")
	}

	if argsWithoutProg[0] == "--C" {
		C_main(argsWithoutProg[1:])
		return
	}

	if argsWithoutProg[0] == "--A" {
		A_main(argsWithoutProg[1:])
		return
	}

	if argsWithoutProg[0] == "--B" {
		B_main(argsWithoutProg[1:])
		return
	}

	panic(fmt.Sprintf("Invalid node argument %s provided", argsWithoutProg[0]))
}

func A_main(args []string) {
	checker, addrMaker, readFun, writeFun := makeCheckerReaderWriter("A")

	//////////////////////////////////////
	// Test: should throw error when missing this send
	//bStartBuf := checker.PrepareSend("Sending B Commit Req", 0)
	//checker.WriteToUDP("127.0.0.1:24602", writeFun, bStartBuf, addrMaker)

	bResponseBuf := make([]byte, 1024)
	checker.ReadFromUDP("127.0.0.1:24601", readFun, bResponseBuf)
	var bResponse string
	checker.UnpackReceive("Unpacking B response", bResponseBuf, &bResponse)
	switch bResponse {

	case "B-Fail":

		//We still have to get our message from C, even though we will abort
		cStartBuf := checker.PrepareSend("Sending C Commit Req", 0)
		checker.WriteToUDP("127.0.0.1:24603", writeFun, cStartBuf, addrMaker)

		cResponseBuf := make([]byte, 1024)
		checker.ReadFromUDP("127.0.0.1:24601", readFun, cResponseBuf)
		var cResponse string
		checker.UnpackReceive("Unpacking C Response", cResponseBuf, &cResponse)

		bCommitBuf := checker.PrepareSend("Telling B to Abort", false)
		checker.WriteToUDP("127.0.0.1:24602", writeFun, bCommitBuf, addrMaker)

		cCommitBuf := checker.PrepareSend("Telling C to Abort", false)
		checker.WriteToUDP("127.0.0.1:24603", writeFun, cCommitBuf, addrMaker)

		println("A Aborted")
		return

	case "B-Commit":

		sendBuf := checker.PrepareSend("Telling C to Start Commit", 0)
		checker.WriteToUDP("127.0.0.1:24603", writeFun, sendBuf, addrMaker)

		cResponseBuf := make([]byte, 1024)
		checker.ReadFromUDP("127.0.0.1:24601", readFun, cResponseBuf)
		var cResponse string
		checker.UnpackReceive("Unpacking Response from C", cResponseBuf, &cResponse)
		switch cResponse {

		case "C-Commit":

			bCommitBuf := checker.PrepareSend("Telling B To Commit", true)
			checker.WriteToUDP("127.0.0.1:24602", writeFun, bCommitBuf, addrMaker)

			cCommitBuf := checker.PrepareSend("Telling C to Commit", true)
			checker.WriteToUDP("127.0.0.1:24603", writeFun, cCommitBuf, addrMaker)

			println("A Commited")
			return

		case "C-Fail":

			bCommitBuf := checker.PrepareSend("Telling B to Abort", false)
			checker.WriteToUDP("127.0.0.1:24602", writeFun, bCommitBuf, addrMaker)

			cCommitBuf := checker.PrepareSend("Telling C to Abort", false)
			checker.WriteToUDP("127.0.0.1:24603", writeFun, cCommitBuf, addrMaker)
			println("A Aborted")
			return

		default:
			panic("Invalid label sent at selection choice")
		}

	default:
		panic("Invalid label sent at selection choice")
	}

}

func B_main(args []string) {
	checker, addrMaker, readFun, writeFun := makeCheckerReaderWriter("B")

	//////////////////////////////////////
	// Test: should throw error when missing this receive
	//startCommitBuf := make([]byte, 1024)
	//checker.ReadFromUDP("127.0.0.1:24602", readFun, startCommitBuf)
	//var startCommitMsg int
	//checker.UnpackReceive("Got message from A requesting Commit Phase 1", startCommitBuf, &startCommitMsg)

	var labelToSend string
	//Randomly choose if we commit or not
	if (rand.Int() % 2) == 0 {
		labelToSend = "B-Fail"
	} else {
		labelToSend = "B-Commit"
	}

	labelBuf := checker.PrepareSend("Telling A if we Commit", labelToSend)
	checker.WriteToUDP("127.0.0.1:24601", writeFun, labelBuf, addrMaker)

	shouldCommitBuf := make([]byte, 1024)
	checker.ReadFromUDP("127.0.0.1:24602", readFun, shouldCommitBuf)
	var shouldCommit bool
	checker.UnpackReceive("Unpacked A's final answer", shouldCommitBuf, &shouldCommit)
	if shouldCommit {
		println("B Comitted")
	} else {
		println("B Aborted")
	}

}

func C_main(args []string) {
	checker, addrMaker, readFun, writeFun := makeCheckerReaderWriter("C")

	startCommitBuf := make([]byte, 1024)
	checker.ReadFromUDP("127.0.0.1:24603", readFun, startCommitBuf)
	var startCommitMsg int
	checker.UnpackReceive("Got message from A requesting Commit Phase 1", startCommitBuf, &startCommitMsg)

	var labelToSend string
	//Randomly choose if we commit or not
	if (rand.Int() % 2) == 0 {
		labelToSend = "C-Fail"
	} else {
		labelToSend = "C-Commit"
	}

	labelBuf := checker.PrepareSend("Telling A if we Commit", labelToSend)
	checker.WriteToUDP("127.0.0.1:24601", writeFun, labelBuf, addrMaker)

	shouldCommitBuf := make([]byte, 1024)
	checker.ReadFromUDP("127.0.0.1:24603", readFun, shouldCommitBuf)
	var shouldCommit bool
	checker.UnpackReceive("Unpacked A's final answer", shouldCommitBuf, &shouldCommit)
	if shouldCommit {
		println("C Comitted")
	} else {
		println("C Aborted")
	}

}
