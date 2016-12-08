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

type Case struct {
	Label  string
	ThenDo Event
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

//Send : wrap a Global Type into an Event for send channel
func Send(channel Channel, messageType MessageType) Event {

	participant1 := multiparty.Participant(channel.Source)
	participant2 := multiparty.Participant(channel.Destination)
	multipartyChannel := multiparty.Channel(channel.Name)

	prefix := multiparty.Prefix{
		P1:       participant1,
		P2:       participant2,
		PChannel: multipartyChannel}

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
func link(events []Event) multiparty.GlobalType {

	//Things get a little more complicated when we have Branching and Parallel types
	//We start at the back of our list of events, accumulating the current "doNext" value
	//When we see a Send type, we make our accum its next, and make it our accum
	//When we see a branching type, we make our accum the next of each branch
	//When we see a parallel type, TODO

	// Iterate through the list of events and place them into nested GlobalTypes
	var currentType multiparty.GlobalType = multiparty.EndType{}
	for i := len(events) - 1; i >= 0; i-- {
		currentType = events[i].wrappedType(currentType)
	}
	return currentType
}
