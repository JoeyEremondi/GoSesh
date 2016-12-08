package mockup

/* GoSesh Mockup is a set of wrapper structs and functions for the application
 * developer to mockup a method stub. The developer creates a method stub file
 * by calling the functions here. These functions convert the developer's code
 * into nested Global Types that are Session Type theory compatible.
 */

import (
	"fmt"
	"os"

	"github.com/JoeyEremondi/GoSesh/multiparty"
)

// Channel :  a connection between two endpoints
type Channel struct {
	Name        string
	Source      string
	Destination string
}

// MessageType : send a message of this type on a channel
// TODO custom struct message types for app dev
type MessageType struct {
	// In the Session Type theory, this is called a Sort
	Type string
}

// Event : Wraps a value type
type Event struct {
	// In Session Type theory, this is the "Value" Type
	// Representing a message being sent and received
	//We store this as a function waiting for whatever thing type "does next"
	wrappedType func(multiparty.GlobalType) multiparty.GlobalType
}

//Use these to make the branches of a Switch statement
type SwitchCase struct {
	label  string
	thenDo []Event
}

func Case(label string, thenDo ...Event) SwitchCase {
	return SwitchCase{label: label, thenDo: thenDo}
}

/* CreateStubProgram : pass in a list of events and file name to generate
 * a go stub file which links the events and converts them to Session Types
 */
func CreateStubProgram(events []Event, fileName string) {
	root := link(events)

	outFile, err := os.Create(fileName + ".go.stub")
	if err != nil {
		fmt.Println("STUB GENERATION ERROR: ", err)
	}
	defer outFile.Close()

	program := multiparty.GenerateProgram(root)

	outFile.WriteString(program)
}

//Helper function: make a prefix from a channel
func makePrefix(channel Channel) multiparty.Prefix {
	participant1 := multiparty.Participant(channel.Source)
	participant2 := multiparty.Participant(channel.Destination)
	multipartyChannel := multiparty.Channel(channel.Name)

	return multiparty.Prefix{
		P1:       participant1,
		P2:       participant2,
		PChannel: multipartyChannel}
}

//Send : wrap a Global Type into an Event for send channel
func Send(channel Channel, messageType MessageType) Event {
	prefix := makePrefix(channel)

	sort := multiparty.Sort(messageType.Type)

	valueType := func(endType multiparty.GlobalType) multiparty.GlobalType {
		return multiparty.ValueType{
			ValuePrefix: prefix,
			Value:       sort,
			ValueNext:   endType}
	}

	send := Event{wrappedType: valueType}

	return send
}

func Switch(channel Channel, branches ...SwitchCase) Event {
	//Each branch has a list of events to do in the branch
	//We make a function wating for the type of what we do after the branch
	//which we then use as next thing to do in each branch.
	//This lets you write natural-looking branches that are translated
	//into  session types
	retFun := func(nextType multiparty.GlobalType) multiparty.GlobalType {
		branchMap := make(map[string]multiparty.GlobalType)
		for _, someCase := range branches {
			branchMap[someCase.label] = linkWithType(someCase.thenDo, nextType)
		}
		return multiparty.BranchingType{
			BranchPrefix: makePrefix(channel),
			Branches:     branchMap}
	}
	return Event{wrappedType: retFun}
}

//Create a named loop, that we can control using Continue() and Break()
//Note that all branches have an implicit Break() if Continue() is not specified
func Loop(label string, bodyEvents ...Event) Event {
	//We pipe the recursive variable as the thing to do next
	//Looping explicitly until we see a break
	retFun := func(nextType multiparty.GlobalType) multiparty.GlobalType {
		return multiparty.RecursiveType{
			Bind: multiparty.NameType(label),
			Body: linkWithType(bodyEvents, nextType)}
	}
	return Event{wrappedType: retFun}
}

//Break from the current innermost loop
func Break() Event {
	//Does nothing, is just for nice syntax
	retFun := func(nextType multiparty.GlobalType) multiparty.GlobalType {
		return nextType
	}
	return Event{wrappedType: retFun}
}

//Jump back to the start of the loop with the given name
func Continue(label string) Event {
	//Since this is a GOTO, it ignores whatever the "next" thing is
	retFun := func(nextType multiparty.GlobalType) multiparty.GlobalType {
		if false { //UNused
			return nextType
		}
		return multiparty.NameType(label)
	}
	return Event{wrappedType: retFun}
}

//Run the given events in parallel
//TODO don't attach doNext to a branch?
func Parallel(events ...Event) Event {
	retFun := func(nextType multiparty.GlobalType) multiparty.GlobalType {
		if len(events) == 0 {
			return nextType
		} else {
			parSoFar := events[0].wrappedType(nextType)
			for _, event := range events[1:] {
				parSoFar = multiparty.MakeParallelType(parSoFar, event.wrappedType(multiparty.EndType{}))
			}
			return parSoFar
		}
	}
	return Event{wrappedType: retFun}
}

//Sequence a bunch of events into a single event
//This is mostly useful for putting them inside a Parallel block
func Sequence(events ...Event) Event {
	retFun := func(nextType multiparty.GlobalType) multiparty.GlobalType {
		return linkWithType(events, nextType)
	}
	return Event{wrappedType: retFun}
}

//Receive : wrap a Global Type into an Event for receive channel
/*
func Receive(channel Channel, messageType MessageType) Event {
	participant1 := multiparty.Participant(channel.Destination)
	participant2 := multiparty.Participant(channel.Source)
	multipartyChannel := multiparty.Channel(channel.Name)

	prefix := multiparty.Prefix{
		P1:       participant1,
		P2:       participant2,
		PChannel: multipartyChannel}

	sort := multiparty.Sort(messageType.Type)

	valueType := multiparty.ValueType{
		ValuePrefix: prefix,
		Value:       sort,
		ValueNext:   multiparty.EndType{}}

	receive := Event{wrappedValueType: valueType}

	return receive
} */

/* EventList : sequential list of events in a protocol stub
 * In Session Type theory, this is a nested linked list
 */
func linkWithType(events []Event, endType multiparty.GlobalType) multiparty.GlobalType {
	//Iterate from the back of our array, build up a single type
	//By passing our currentType to the next function waiting for its "doNext" type
	currentType := endType
	for i := len(events) - 1; i >= 0; i-- {
		currentType = events[i].wrappedType(currentType)
	}
	return currentType
}

func link(events []Event) multiparty.GlobalType {
	return linkWithType(events, multiparty.EndType{})
}
