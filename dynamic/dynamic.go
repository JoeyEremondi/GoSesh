/**
* This module defines the Checker type, which provides a wrapper around GoVector
capabilities. On top of GoVector's functions, it provides dynamic checking against a session type
(usually defined by a mockup).
*/
package dynamic

import (
	"fmt"
	"net"
	"reflect"

	"github.com/JoeyEremondi/GoSesh/multiparty"
	"github.com/arcaneiceman/GoVector/capture"
	"github.com/arcaneiceman/GoVector/govec"
)

//Stores the "current" session type in the computation.
//When network calls are made through the checker, they are checked
//against its type. It will panic if messages are of the wrong type,
//if sends and receives are mixed up or to the wrong party,
//or if sent labels are incorrect.
type Checker struct {
	gv               *govec.GoLog
	currentType      multiparty.LocalType
	expectedSortType multiparty.Sort
	currentLabel     *string
	//TODO other stuff handy to have here?
}

//Create a checker with the given id (participant name)
//and (local) session type.
//GoVector logs are stored in ID_LogFile.txt, where ID is the value of id
func CreateChecker(id string, t multiparty.LocalType) Checker {
	ret := Checker{govec.Initialize(id, id+"_LogFile.txt"), t, multiparty.Sort("ERROR INITIAL SORT"), nil}
	//make sure we start with a type we can deal with
	ret.unfoldIfRecursive()
	return ret
}

//Unfold any top-level recursive types, if they're the current type
//Otherwise, do nothing
func (checker *Checker) unfoldIfRecursive() {
	for {
		switch t := checker.currentType.(type) {
		case multiparty.LocalRecursiveType:
			checker.currentType = t.UnfoldOneLevel()
			//Check if there's nested recursion by looping again
			continue
		default:
			//When we're done, set the sort we're expecting in the next message
			//if it's a send or receive
			checker.setExpectedSort()
			return
		}
	}
}

//Look at the current type, and if it's a send or receive
//Store sort (message type) in a checker variable
func (checker *Checker) setExpectedSort() {
	switch t := checker.currentType.(type) {
	//Send and receive: just progress to the "next" type
	case multiparty.LocalSendType:
		checker.expectedSortType = t.Value
	case multiparty.LocalReceiveType:
		checker.expectedSortType = t.Value
	}
}

//After doing something on the network, we advance to the "next" type of our session type
func (checker *Checker) advanceType() error {

	//Then, advance the type, if we can
	switch t := checker.currentType.(type) {

	//Send and receive: just progress to the "next" type
	case multiparty.LocalSendType:
		checker.currentType = t.Next

	case multiparty.LocalReceiveType:
		checker.currentType = t.Next

	//Branch and select: what type we progress to depends on the label that was
	//sent or received, so we use that to choose the next type
	case multiparty.LocalBranchingType:
		if checker.currentLabel != nil {
			checker.currentType = t.Branches[*checker.currentLabel]
			checker.currentLabel = nil
		} else {

		}

	case multiparty.LocalSelectionType:
		if checker.currentLabel != nil {
			checker.currentType = t.Branches[*checker.currentLabel]
			checker.currentLabel = nil
		} else {

		}

	case multiparty.LocalEndType:
		return fmt.Errorf("Tried to keep going after hitting the End type")

	default:
		panic("Missing a case for session types! Means recursion probably was improperly removed")
	}
	//Finally, unroll any recursion types that we have at the top level
	checker.unfoldIfRecursive()
	return nil
}

//UnpackReceive : Wrapper around GoVector's pack and unpack functions
//Checks that the current session type is expecting a recieve,
//and that the message is unpacked into the correct type
func (checker *Checker) UnpackReceive(mesg string, buf []byte, unpack interface{}) {

	//Do the GoVector unpack
	checker.gv.UnpackReceive(mesg, buf, unpack)

	//Make sure we're in a receive or a branch
	switch t := checker.currentType.(type) {
	case multiparty.LocalReceiveType:
		// Check that the interface type is the correct Sort for the send/receive pair
		interfaceType := reflect.TypeOf(unpack).Elem().String()
		sortType := reflect.ValueOf(checker.expectedSortType).String()

		if sortType != interfaceType {
			panic(fmt.Sprintf("Wrong type for message data in UnpackReceive, given %s expected %s",
				interfaceType, sortType))
		}
	case multiparty.LocalBranchingType:
		//Make sure that what was sent was a label (string)
		//And that it is one of the labels of our current type
		switch unpackString := unpack.(type) {
		case *string:
			_, ok := t.Branches[*unpackString]
			checker.currentLabel = unpackString
			if !ok {
				allBranches := ""
				for label, _ := range t.Branches {
					allBranches += label + ", "
				}
				panic(fmt.Sprintf(
					"Received invalid label %s at branching point, should be one of %s",
					*unpackString, allBranches))
			}

		case string:
			_, ok := t.Branches[unpackString]
			checker.currentLabel = &unpackString
			if !ok {
				allBranches := ""
				for label, _ := range t.Branches {
					allBranches += label + ", "
				}
				panic(fmt.Sprintf(
					"Received invalid label %s at branching point, should be one of %s",
					unpackString, allBranches))
			}

		default:
			panic("Unpacking data of the wrong type at a Branching point. Should be a string")
		}

	case multiparty.LocalSendType:
		panic("Tried to do receive on send type")

	case multiparty.LocalSelectionType:
		panic("Tried to do receive on selection type")

	case multiparty.LocalEndType:
		panic("Tried to do a receive when we should be done communications.")

	default:
		panic(fmt.Sprintf("Unknown type %T in UnpackReceive", t))
	}

	//Now that we're done, advance our type to whatever we do next
	err := checker.advanceType()
	if err != nil {
		panic(err)
	}
}

// PrepareSend : Prepare a send with GoVector
// Check that the current session type is expecting a send,
// and that the given value has the correct type
func (checker *Checker) PrepareSend(msg string, buf interface{}) []byte {
	// Fill the buffer with contents of message
	gvBuffer := checker.gv.PrepareSend(msg, buf)
	// Make sure we're in a send or a branch
	switch t := checker.currentType.(type) {
	// Check that the interface passed in the correct Sort for the send/receive pair
	case multiparty.LocalSendType:
		interfaceType := reflect.TypeOf(buf).String()
		sortType := reflect.ValueOf(checker.expectedSortType).String()

		if sortType != interfaceType {
			panic(fmt.Sprintf("Wrong message type in PrepareSend, given %s expected %s", interfaceType, sortType))
		}

	case multiparty.LocalSelectionType:
		// Make sure that what was sent was a label (string)
		// And that it is one of the labels of our current type
		switch unpackString := buf.(type) {
		case *string:
			_, ok := t.Branches[*unpackString]
			checker.currentLabel = unpackString
			if !ok {
				allBranches := ""
				for label, _ := range t.Branches {
					allBranches += label + ", "
				}
				panic(fmt.Sprintf("Sent invalid label %s at branching point, should be one of %s", *unpackString, allBranches))
			}

		case string:
			_, ok := t.Branches[unpackString]
			checker.currentLabel = &unpackString
			if !ok {
				allBranches := ""
				for label, _ := range t.Branches {
					allBranches += label + ", "
				}
				panic(fmt.Sprintf("Sent invalid label %s at branching point, should be one of %s", unpackString, allBranches))
			}

		default:
			panic("Unpacking data of the wrong type at a Selection point. Should be a string")
		}
	case multiparty.LocalReceiveType:
		panic("Tried to do receive on send type")

	default:
		panic(fmt.Sprintf("Unknown type in PrepareSend %T", t))
	}

	return gvBuffer

}

//Make sure the given channel matches the channel of the current type
func (checker *Checker) checkRecvChannel(c multiparty.Channel) {
	switch t := checker.currentType.(type) {
	case multiparty.LocalReceiveType:
		if t.Channel != c {
			panic(fmt.Sprintf("Expected to receive on channel %s, but was given %s", t.Channel, c))
		}
	case multiparty.LocalBranchingType:
		if t.Channel != c {
			panic(fmt.Sprintf("Expected to receive on channel %s, but was given %s", t.Channel, c))
		}
	default:
		//TODO say what was expected
		panic(fmt.Sprintf("Cannot do a receive on non-receive localType %T", t))
	}
}

//Make sure the given channel matches the channel of the current type
func (checker *Checker) checkSendChannel(c multiparty.Channel) {
	switch t := checker.currentType.(type) {
	case multiparty.LocalSendType:
		if t.Channel != c {
			panic(fmt.Sprintf("Expected to send to channel %s, but was given %s", t.Channel, c))
		}
	case multiparty.LocalSelectionType:
		if t.Channel != c {
			panic(fmt.Sprintf("Expected to send to channel %s, but was given %s", t.Channel, c))
		}
	default:
		//TODO say what was expected
		panic("Cannot do a send on a non-send localType")
	}
}

//A wrapper around the GoVector function of the same name.
//The function takes the channel (ip:port) being read from, and a callback which performs the correct network operation
//from the given channel.
func (checker *Checker) Read(c multiparty.Channel, read func(multiparty.Channel, []byte) (int, error), b []byte) (int, error) {
	checker.checkRecvChannel(c)

	curriedRead := func([]byte) (int, error) { return read(c, b) }
	return capture.Read(curriedRead, b)
}

//A wrapper around the GoVector function of the same name.
//The function takes the channel (ip:port) being sent to, and a callback which performs the correct network operation
//from the given channel.
func (checker *Checker) Write(c multiparty.Channel, write func(c multiparty.Channel, b []byte) (int, error), b []byte) (int, error) {
	checker.checkSendChannel(c)
	// Now that we're done, advance our type to whatever we do next
	err := checker.advanceType()
	if err != nil {
		panic(err)
	}
	curriedWrite := func(b []byte) (int, error) { return write(c, b) }
	return capture.Write(curriedWrite, b)
}

//A wrapper around the GoVector function of the same name.
//The function takes the channel (ip:port) being read from, and a callback which performs the correct network operation
//from the given channel.
func (checker *Checker) ReadFrom(c multiparty.Channel, readFrom func(multiparty.Channel, []byte) (int, net.Addr, error), b []byte) (int, net.Addr, error) {
	checker.checkRecvChannel(c)

	curriedRead := func(b []byte) (int, net.Addr, error) { return readFrom(c, b) }
	return capture.ReadFrom(curriedRead, b)
}

//A wrapper around the GoVector function of the same name.
//The function takes the channel (ip:port) being sent to, and a callback which performs the correct network operation
//from the given channel.
func (checker *Checker) WriteTo(c multiparty.Channel, writeTo func(multiparty.Channel, []byte, net.Addr) (int, error), b []byte, addrMaker func(multiparty.Channel) net.Addr) (int, error) {
	checker.checkSendChannel(c)
	// Now that we're done, advance our type to whatever we do next
	err := checker.advanceType()
	if err != nil {
		panic(err)
	}
	curriedWrite := func(b []byte, a net.Addr) (int, error) { return writeTo(c, b, a) }
	return capture.WriteTo(curriedWrite, b, addrMaker(c))
}

//A wrapper around the GoVector function of the same name.
//The function takes the channel (ip:port) being read from, and a callback which performs the correct network operation
//from the given channel.
func (checker *Checker) ReadFromUDP(c multiparty.Channel, readFrom func(multiparty.Channel, []byte) (int, *net.UDPAddr, error), b []byte) (int, *net.UDPAddr, error) {
	checker.checkRecvChannel(c)

	curriedRead := func(b []byte) (int, *net.UDPAddr, error) { return readFrom(c, b) }
	return capture.ReadFromUDP(curriedRead, b)
}

//A wrapper around the GoVector function of the same name.
//The function takes the channel (ip:port) being sent to, and a callback which performs the correct network operation
//from the given channel.
func (checker *Checker) WriteToUDP(c multiparty.Channel, writeTo func(multiparty.Channel, []byte, *net.UDPAddr) (int, error), b []byte, addrMaker func(multiparty.Channel) *net.UDPAddr) (int, error) {
	checker.checkSendChannel(c)
	// Now that we're done, advance our type to whatever we do next
	err := checker.advanceType()
	if err != nil {
		panic(err)
	}
	curriedWrite := func(b []byte, a *net.UDPAddr) (int, error) { return writeTo(c, b, a) }
	return capture.WriteToUDP(curriedWrite, b, addrMaker(c))
}
