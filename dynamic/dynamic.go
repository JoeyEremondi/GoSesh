package dynamic

import (
	"fmt"
	"net"
	"reflect"

	"github.com/JoeyEremondi/GoSesh/multiparty"
	"github.com/arcaneiceman/GoVector/govec"
)

//Checker : Store the "current" type in the computation
//so that we can ensure each network operation preserves its
type Checker struct {
	gv               *govec.GoLog
	currentType      multiparty.LocalType
	expectedSortType multiparty.Sort
	currentLabel     *string
	channels         map[string]*net.Conn
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
		interfaceType := reflect.TypeOf(unpack).String()
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
		panic("Tried to do send on receive type")

	default:
		panic(fmt.Sprintf("Unknown type in UnpackReceive %s", t))
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

		default:
			panic("Unpacking data of the wrong type at a Selection point. Should be a string")
		}
	case multiparty.LocalReceiveType:
		panic("Tried to do receive on send type")

	default:
		panic(fmt.Sprintf("Unknown type in PrepareSend %s", t))
	}

	// Now that we're done, advance our type to whatever we do next
	err := checker.advanceType()
	if err == nil {
		return gvBuffer

	}
	panic(err)
}
