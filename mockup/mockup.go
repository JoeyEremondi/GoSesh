// GoSesh Mockup is a set of wrapper structs and functions for the application
// developer to mockup a multiparty interaction.
// The developer uses the mockup functions to create a protocol for their programs,
// by calling the functions here. These functions create a stub Go file which convert the developer's mockup
// into Global Types that can be used by the dynamic Session Type checker.
//

package mockup

import (
	"fmt"
	"os"

	"github.com/JoeyEremondi/GoSesh/multiparty"
)

// A connection between two endpoints.
// The Name field should be the receiving ip:port string.
type Channel struct {
	Name        string
	Source      string
	Destination string
}

//Wraps the (string representing) the Go type of the message being sent.
type MessageType struct {
	// In the Session Type theory, this is called a Sort
	Type string
}

//An abstraction of an interaction that can occur between some parties.
//A mockup is made up of a number of events, run in sequence.
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

//Used to make the case for the given label inside of a Switch block.
func Case(label string, thenDo ...Event) SwitchCase {
	return SwitchCase{label: label, thenDo: thenDo}
}

/*
* Pass in the path to the file containing this mockup, the file name to generate,
* and a list of events forming a mockup.
* This will create a .go.stub file with the same contents (type definitions)
* as the input file, with the boilerplate code for a program performing the given events.
 */
func CreateStubProgram(infile string, outfile string, events ...Event) {
	root := Link(events...)

	outFile, err := os.Create(outfile + ".go.stub")
	if err != nil {
		fmt.Println("STUB GENERATION ERROR: ", err)
	}
	defer outFile.Close()

	initialProgram := modifiedFileString(infile,
		"github.com/JoeyEremondi/GoSesh/multiparty",
		"github.com/JoeyEremondi/GoSesh/dynamic",
		"github.com/JoeyEremondi/GoSesh/mockup",
		"os",
		"net",
		"fmt",
	)

	programLogic := generateProgram(root)

	outFile.WriteString(initialProgram + "\n" + programLogic)
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

//Send a value of the given type along the given channel.
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

//Create an event which performs sends a label on the given channel,
//then performs the events in the case for whichever label was sent.
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

//Create a named loop, that we can control using Continue() and Break().
//Note that all branches have an implicit Break() if Continue() is not specified.
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

//Create an event corresponding to the given events running in parallel.
//Because we don't know when a parallel block finished, no events may
//come after a parallel block.
func Parallel(events ...Event) Event {
	retFun := func(nextType multiparty.GlobalType) multiparty.GlobalType {
		switch nextType.(type) {
		case multiparty.EndType:
		default:
			panic("Can't have any events occur after a parallel block.")
		}
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

//We can model an empty event using the identity function
func internalDoNothing() Event {
	retFun := func(nextType multiparty.GlobalType) multiparty.GlobalType {
		return nextType
	}
	return Event{wrappedType: retFun}
}

//Empty event, useful for empty branches
//Just does whatever comes next
var DoNothing Event = internalDoNothing()

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

//Sequence the given events, and turn them into a global session type
func Link(events ...Event) multiparty.GlobalType {
	return linkWithType(events, multiparty.EndType{})
}

//Useful utility function
func setGlobalType(ptr *multiparty.GlobalType, infile string, outfile string, types ...Event) {
	*ptr = Link(types...)
	if infile == outfile {
	}
}
