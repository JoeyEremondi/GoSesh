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
	// In Session Type theory, this is the Value Type
	wrappedValueType multiparty.ValueType
}

/*
 * CreateStubProgram : pass in a list of events and file name to generate
 * a go stub file which links the events and converts them to Session Types
 */
func CreateStubProgram(events []Event, fileName string) {
	root := link(events)

	outFile, err := os.Create(fileName + ".go.stub")
	fmt.Println("STUB GENERATION ERROR: ", err)
	defer outFile.Close()

	multiparty.GenerateProgram(root)
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

	valueType := multiparty.ValueType{
		ValuePrefix: prefix,
		Value:       sort,
		ValueNext:   nil}

	send := Event{wrappedValueType: valueType}

	return send
}

//Receive : wrap a Global Type into an Event for receive channel
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
		ValueNext:   nil}

	receive := Event{wrappedValueType: valueType}

	return receive
}

/* EventList : sequential list of events in a protocol stub
 * In Session Type theory, this is a nested linked list
 */
func link(events []Event) multiparty.GlobalType {

	// Iterate through the list of events and place them into nested GlobalTypes
	for i := 0; i < len(events)-1; i++ {
		events[i].wrappedValueType.ValueNext = events[i+1].wrappedValueType
	}
	rootValueType := events[0].wrappedValueType

	return rootValueType
}
