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
	//TODO other stuff handy to have here?
}

func CreateChecker(id string, t multiparty.LocalType) Checker {
	ret := Checker{govec.Initialize(id, "TODOLogFile.txt"), t, nil}
	//make sure we start with a type we can deal with
	ret.unfoldIfRecurisve()
	return ret
}

//Unfold any top-level recursive types, if they're the current type
//Otherwise, do nothing
func (checker *Checker) unfoldIfRecurisve() {
	for {
		switch t := checker.currentType.(type) {
		case multiparty.LocalRecursiveType:
			//fmt.Printf("Current type before unfolding: %#v\n", checker.currentType)
			checker.currentType = t.UnfoldOneLevel()
			//fmt.Printf("Current type after unfolding: %#v\n", checker.currentType)
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
		//TODO check that it's the right Go type (the right sort)
	case multiparty.LocalBranchingType:
		//Make sure that what was sent was a label (string)
		//And that it is one of the labels of our current type
		switch unpackString := unpack.(type) {
		case *string:
			_, ok := t.Branches[*unpackString]
			if !ok {
				errString := fmt.Sprintf("Received invalid label %s at branching point, should be one of \n", *unpackString)
				for k, _ := range t.Branches {
					errString += k + ", "
				}
				panic(errString)
			}

		default:
			panic(fmt.Sprintf("Unpacking data of the wrong type at a Branching point. Should be a *string, but is %T", unpack))
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
				errString := fmt.Sprintf("Received invalid label %s at branching point, should be one of \n", *unpackString)
				for k, _ := range t.Branches {
					errString += k + ", "
				}
				panic(errString)
			}

		case string:
			_, ok := t.Branches[unpackString]
			if !ok {
				errString := fmt.Sprintf("Received invalid label %s at branching point, should be one of \n", unpackString)
				for k, _ := range t.Branches {
					errString += k + ", "
				}
				panic(errString)
			}

		default:
			panic(fmt.Sprintf("Unpacking data of the wrong type at a Selection point. Should be a string or *string, but is %T", unpackString))
		}
	default:
		panic(fmt.Sprintf("Tried to do send on receive type. "))

	}

	//Now that we're done, advance our type to whatever we do next
	err := checker.advanceType()
	if err == nil {
		return ret
	}
	panic(err)
}

func (checker *Checker) Read(c multiparty.Channel, read func(multiparty.Channel, []byte) (int, error), b []byte) (int, error) {
	curriedRead := func([]byte) (int, error) { return read(c, b) }
	return capture.Read(curriedRead, b)
}

func (checker *Checker) Write(c multiparty.Channel, write func(c multiparty.Channel, b []byte) (int, error), b []byte) (int, error) {
	curriedWrite := func(b []byte) (int, error) { return write(c, b) }
	return capture.Write(curriedWrite, b)
}

func (checker *Checker) ReadFrom(c multiparty.Channel, readFrom func(multiparty.Channel, []byte) (int, net.Addr, error), b []byte) (int, net.Addr, error) {
	curriedRead := func(b []byte) (int, net.Addr, error) { return readFrom(c, b) }
	return capture.ReadFrom(curriedRead, b)
}

func (checker *Checker) WriteTo(c multiparty.Channel, writeTo func(multiparty.Channel, []byte, net.Addr) (int, error), b []byte, addrMaker func(multiparty.Channel) net.Addr) (int, error) {
	curriedWrite := func(b []byte, a net.Addr) (int, error) { return writeTo(c, b, a) }
	return capture.WriteTo(curriedWrite, b, addrMaker(c))
}

func (checker *Checker) ReadFromUDP(c multiparty.Channel, readFrom func(multiparty.Channel, []byte) (int, *net.UDPAddr, error), b []byte) (int, *net.UDPAddr, error) {
	curriedRead := func(b []byte) (int, *net.UDPAddr, error) { return readFrom(c, b) }
	return capture.ReadFromUDP(curriedRead, b)
}

func (checker *Checker) WriteToUDP(c multiparty.Channel, writeTo func(multiparty.Channel, []byte, *net.UDPAddr) (int, error), b []byte, addrMaker func(multiparty.Channel) *net.UDPAddr) (int, error) {
	curriedWrite := func(b []byte, a *net.UDPAddr) (int, error) { return writeTo(c, b, a) }
	return capture.WriteToUDP(curriedWrite, b, addrMaker(c))
}
