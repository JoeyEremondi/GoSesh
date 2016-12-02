package generation

import (
	"testing"

	"github.com/JoeyEremondi/GoSesh/generation"
)

// Run a single test file with verbose comments like:
// go test -v generation_test.go
func TestLocalSendType(test *testing.T) {
	channel := generation.Channel{
		Name:        "testchannel",
		Source:      "A",
		Destination: "B"}

	valueType := generation.MessageType{Type: "int"}
	Event := generation.Event{Event: "nexttype"}

	localSendType := generation.Send(channel, valueType, Event)
	test.Log(localSendType)
}

func TestLocalReceiveType(test *testing.T) {
	channel := generation.Channel{
		Name:        "testchannel",
		Source:      "A",
		Destination: "B"}
	valueType := generation.MessageType{Type: "int"}
	Event := generation.Event{Event: "nexttype"}

	localReceiveType := generation.Receive(channel, valueType, Event)
	test.Log(localReceiveType)
}

/*
func TestLocalBranchingType(test *testing.T) {
	channel := generation.Channel{Name: "testchannel"}
	event1 := generation.Event{Event: "Event1"}
	event2 := generation.Event{Event: "Event2"}
	event3 := generation.Event{Event: "Event3"}
	// TODO what is this key supposed to be used for?

	EventMap := map[string]generation.Event{
		"1": event1,
		"2": event2,
		"3": event3}
	localBranchingType := generation.Branch(channel, EventMap)
	test.Log(localBranchingType)
}*/

// TODO test selection type
