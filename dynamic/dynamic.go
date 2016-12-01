package dynamic

import (
	"fmt"
	"net"

	"github.com/JoeyEremondi/GoSesh/multiparty"
	"github.com/arcaneiceman/GoVector/capture"
	"github.com/arcaneiceman/GoVector/govec"
)

//Store the "current" type in the computation
//so that we can ensure each network operation preserves its
type Checker struct {
	gv           *govec.GoLog
	currentType  multiparty.LocalType
	currentLabel *string
	Channels     map[string]*net.Conn
	//TODO other stuff handy to have here?
}

//Unfold any top-level recursive types, if they're the current type
//Otherwise, do nothing
func (checker *Checker) unfoldIfRecurisve() {
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
	checker.unfoldIfRecurisve()
	return nil
}

//Wrapper around GoVector's pack and unpack functions
//These are where we check to make sure that the correct type is the current type
//i.e. don't send on a receive, etc.
func (checker *Checker) UnpackReceive(mesg string, buf []byte, unpack interface{}) {

	//Do the GoVector unpack
	checker.gv.UnpackReceive(mesg, buf, unpack)

	//Make sure we're in a receive or a branch
	switch t := checker.currentType.(type) {
	case multiparty.LocalReceiveType:
		//Nothing to do here, we're good
	case multiparty.LocalBranchingType:
		//Make sure that what was sent was a label (string)
		//And that it is one of the labels of our current type
		switch unpackString := unpack.(type) {
		case *string:
			_, ok := t.Branches[*unpackString]
			if !ok {
				panic(fmt.Sprintf("Received invalid label %s at branching point, should be one of TODO", *unpackString))
			}

		default:
			panic("Unpacking data of the wrong type at a Branching point. Should be a string")
		}
	default:
		panic("Tried to do send on receive type")

	}

	//Now that we're done, advance our type to whatever we do next
	err := checker.advanceType()
	if err != nil {
		panic(err)
	}
}

//The sending version
func (checker *Checker) PrepareSend(mesg string, buf interface{}) []byte {

	//Do the GoVector unpack
	ret := checker.gv.PrepareSend(mesg, buf)

	//Make sure we're in a receive or a branch
	switch t := checker.currentType.(type) {
	case multiparty.LocalSendType:
		//Nothing to do here, we're good
	case multiparty.LocalSelectionType:
		//Make sure that what was sent was a label (string)
		//And that it is one of the labels of our current type
		switch unpackString := buf.(type) {
		case *string:
			_, ok := t.Branches[*unpackString]
			if !ok {
				panic(fmt.Sprintf("Sent invalid label %s at branching point, should be one of TODO", *unpackString))
			}

		default:
			panic("Unpacking data of the wrong type at a Selection point. Should be a string")
		}
	default:
		panic("Tried to do receive on send type")

	}

	//Now that we're done, advance our type to whatever we do next
	err := checker.advanceType()
	if err == nil {
		return ret
	}
	panic(err)
}

func (checker Checker) WriteTo(c multiparty.Channel, write func(c multiparty.Channel, b []byte) (int, error), b []byte) (int, error) {
	curriedWrite := func(b []byte) (int, error) { return write(c, b) }
	return capture.Write(curriedWrite, b)
}

func (checker Checker) ReadFrom(c multiparty.Channel, read func(multiparty.Channel, []byte) (int, error), b []byte) (int, error) {
	curriedRead := func([]byte) (int, error) { return read(c, b) }
	return capture.Read(curriedRead, b)
}
