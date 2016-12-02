package generation

import (
	"reflect"

	"github.com/JoeyEremondi/GoSesh/multiparty"
)

// Channel type here is a wrapper for the developer to use
type Channel struct {
	Name        string
	Source      string
	Destination string
}

// MessageType is a wrapper for the developer to use
type MessageType struct {
	// Call it a type instead of "value of a certain Sort"
	// which is non-intuitive to the developer
	Type string
}

// Event is a wrapper for the developer to use
type Event struct {
	// Wrap the global type into an event for easier comprehension
	Event string //ValueType
}

//Send : wrap this into a ValueType with a Participant for source & dest
func Send(channel Channel, messageType MessageType, nextEvent Event) string {

	participant1 := multiparty.Participant(channel.Source)
	participant2 := multiparty.Participant(channel.Destination)
	multipartyChannel := multiparty.Channel(channel.Name)

	prefix := multiparty.Prefix{
		P1:       participant1,
		P2:       participant2,
		PChannel: multipartyChannel}

	sort := multiparty.Sort(messageType.Type)

	// TODO how to parse a generic event from user into global type
	//next := nextEvent.Event

	send := multiparty.ValueType{
		ValuePrefix: prefix,
		Value:       sort,
		ValueNext:   nil}

	// TODO how to output the stub?
	return "/n" + reflect.ValueOf(send.Value).String() + " --> "
}

//Receive : wrap this into a ValueType with a Participant for dest & source
func Receive(channel Channel, messageType MessageType, nextEvent Event) string {
	return channel.Destination + " --> " + channel.Source + " : " + channel.Name
}

/*

//LocalBranchingType : perform unwrapping for the developer
func LocalBranchingType(channel Channel, branches map[string]LocalType) string {
	var buffer bytes.Buffer
	var i int
	// TODO
	for _, v := range branches {
		i++
		buffer.WriteString(v.Anything)
		if i != len(branches) {
			buffer.WriteString("\n + \n ")
		}
	}

	return buffer.String()
}

//LocalSelectionType : perform unwrapping for the developer
func LocalSelectionType(channel Channel, branches map[string]LocalType) string {
	var buffer bytes.Buffer
	// TODO
	for _, v := range branches {
		buffer.WriteString(v.Anything)
	}

	return buffer.String()
}

// TODO are we implementing the LocalRecursiveType type?
// TODO implicitly end the operations for the dev LocalEndType
*/
