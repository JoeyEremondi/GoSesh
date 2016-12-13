package dynamic

import (
	"fmt"
	"net"
	"reflect"

	"github.com/JoeyEremondi/GoSesh/multiparty"
	"github.com/arcaneiceman/GoVector/capture"
	"github.com/arcaneiceman/GoVector/govec"
)

//Checker : Store the "current" type in the computation
//so that we can ensure each network operation preserves its
type Checker struct {
	gv               *govec.GoLog
	currentType      multiparty.LocalType
	expectedSortType multiparty.Sort
	currentLabel     *string
	//TODO other stuff handy to have here?
}

func CreateChecker(id string, t multiparty.LocalType) Checker {
	ret := Checker{govec.Initialize(id, "TODOLogFile.txt"), t, multiparty.Sort("ERROR INITIAL SORT"), nil}
	//make sure we start with a type we can deal with
	ret.unfoldIfRecursive()
	ret.setInitialSort()
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
			return
		}
	}
}

func (checker *Checker) setInitialSort() {
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
		checker.expectedSortType = t.Value

	case multiparty.LocalReceiveType:
		checker.currentType = t.Next
		checker.expectedSortType = t.Value

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
//These are where we check to make sure that the correct type is the current type
//i.e. don't send on a receive, etc.
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
			panic(fmt.Sprintf("Invalid sort type in PrepareSend, is %s should be %s",
				sortType, interfaceType))
		}
	case multiparty.LocalBranchingType:
		//Make sure that what was sent was a label (string)
		//And that it is one of the labels of our current type
		switch unpackString := unpack.(type) {
		case *string:
			_, ok := t.Branches[*unpackString]
			if !ok {
				panic(fmt.Sprintf(
					"Received invalid label %s at branching point, should be one of TODO",
					*unpackString))
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
// Check that the send has a matching receive and that it contains the
// correct types
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
			panic(fmt.Sprintf("Invalid sort type in PrepareSend, is %s should be %s", sortType, interfaceType))
		}

	case multiparty.LocalSelectionType:
		// Make sure that what was sent was a label (string)
		// And that it is one of the labels of our current type
		switch unpackString := buf.(type) {
		case *string:
			_, ok := t.Branches[*unpackString]
			if !ok {
				panic(fmt.Sprintf("Sent invalid label %s at branching point, should be one of TODO", *unpackString))
			}

		case string:
			_, ok := t.Branches[unpackString]
			if !ok {
				panic(fmt.Sprintf("Sent invalid label %s at branching point, should be one of TODO", unpackString))
			}

		default:
			panic("Unpacking data of the wrong type at a Selection point. Should be a string")
		}
	case multiparty.LocalReceiveType:
		panic("Tried to do receive on send type")

	default:
		panic(fmt.Sprintf("Unknown type in PrepareSend %s", t))
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

//Same, but for sends
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

//Wrappers around GoVector functions
//Not much interesting happens, except that we have an extra parameter for the channel
//both in our wrapper, and in the function the user gives us

func (checker *Checker) Read(c multiparty.Channel, read func(multiparty.Channel, []byte) (int, error), b []byte) (int, error) {
	checker.checkRecvChannel(c)

	curriedRead := func([]byte) (int, error) { return read(c, b) }
	return capture.Read(curriedRead, b)
}

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

func (checker *Checker) ReadFrom(c multiparty.Channel, readFrom func(multiparty.Channel, []byte) (int, net.Addr, error), b []byte) (int, net.Addr, error) {
	checker.checkRecvChannel(c)

	curriedRead := func(b []byte) (int, net.Addr, error) { return readFrom(c, b) }
	return capture.ReadFrom(curriedRead, b)
}

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

func (checker *Checker) ReadFromUDP(c multiparty.Channel, readFrom func(multiparty.Channel, []byte) (int, *net.UDPAddr, error), b []byte) (int, *net.UDPAddr, error) {
	checker.checkRecvChannel(c)

	curriedRead := func(b []byte) (int, *net.UDPAddr, error) { return readFrom(c, b) }
	return capture.ReadFromUDP(curriedRead, b)
}

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
